/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
 * Copyright (C) 2025 Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *     http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// buildSyncedResourceFromKV normalizes one etcd KV into a GatewaySyncData snapshot.
//
// Returns (nil, false) when the KV key cannot be parsed. Caller MUST emit
// logging.Errorf("key is not validate: %s", kv.Key) on the false branch so that
// the existing operational observability is preserved.
//
// The internal branch layout is slightly different from the original inline
// code in kvToResource(...) (if-else-if -> explicit switch on resource type),
// but behavior is equivalent: PluginMetadata always has SetName(id); other
// types set fallback name only when GetName() == "".
func buildSyncedResourceFromKV(
	normalizedPrefix string,
	gatewayID int,
	kv storage.KeyValuePair,
) (*model.GatewaySyncData, bool) {
	resourceKeyWithoutPrefix := strings.TrimPrefix(kv.Key, normalizedPrefix)
	resourceKeyList := strings.Split(resourceKeyWithoutPrefix, "/")
	if len(resourceKeyList) != 2 {
		return nil, false
	}

	resourceTypeValue := resourceKeyList[0]
	id := resourceKeyList[1]
	resourceType := constant.ResourcePrefixTypeMap[resourceTypeValue]
	if resourceType == "" {
		return nil, false
	}

	resourceInfo := &model.GatewaySyncData{
		ID:          id,
		GatewayID:   gatewayID,
		Type:        resourceType,
		Config:      datatypes.JSON(kv.Value),
		ModRevision: int(kv.ModRevision),
	}
	resourceInfo.Config, _ = sjson.DeleteBytes(resourceInfo.Config, "update_time")
	resourceInfo.Config, _ = sjson.DeleteBytes(resourceInfo.Config, "create_time")

	if resourceType == constant.PluginMetadata {
		resourceInfo.SetName(id)
	} else if resourceInfo.GetName() == "" {
		resourceInfo.SetName(fmt.Sprintf("%s_%s", resourceTypeValue, id))
	}

	return resourceInfo, true
}

// backfillStoredSnapshotFields fills snapshot Config fields (name/id/labels)
// from the matching DB rows for global_rule / plugin_config / consumer_group /
// proto / stream_route. Resource types not in this list are passed through.
//
// TODO(perf): these 5 QueryXxx calls are serial today (preserving the original
// kvToResource(...) behavior). A follow-up can move them behind an errgroup.
func backfillStoredSnapshotFields(ctx context.Context, resources []*model.GatewaySyncData) error {
	globalRuleMap := make(map[string]*model.GatewaySyncData)
	pluginConfigMap := make(map[string]*model.GatewaySyncData)
	consumerGroupMap := make(map[string]*model.GatewaySyncData)
	protoMap := make(map[string]*model.GatewaySyncData)
	streamRouteMap := make(map[string]*model.GatewaySyncData)

	var globalRuleIDs []string
	var pluginConfigIDs []string
	var consumerGroupIDs []string
	var protoIDs []string
	var streamRouteIDs []string

	for _, resource := range resources {
		switch resource.Type {
		case constant.GlobalRule:
			globalRuleMap[resource.ID] = resource
			globalRuleIDs = append(globalRuleIDs, resource.ID)
		case constant.PluginConfig:
			pluginConfigMap[resource.ID] = resource
			pluginConfigIDs = append(pluginConfigIDs, resource.ID)
		case constant.ConsumerGroup:
			consumerGroupMap[resource.ID] = resource
			consumerGroupIDs = append(consumerGroupIDs, resource.ID)
		case constant.Proto:
			protoMap[resource.ID] = resource
			protoIDs = append(protoIDs, resource.ID)
		case constant.StreamRoute:
			streamRouteMap[resource.ID] = resource
			streamRouteIDs = append(streamRouteIDs, resource.ID)
		}
	}

	gatewayID := ginx.GetGatewayInfoFromContext(ctx).ID

	if len(globalRuleIDs) > 0 {
		globalRules, err := QueryGlobalRules(ctx, map[string]any{
			"gateway_id": gatewayID,
			"id":         globalRuleIDs,
		})
		if err != nil {
			return err
		}
		for _, globalRule := range globalRules {
			if resource, ok := globalRuleMap[globalRule.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", globalRule.Name)
			}
		}
	}

	if len(pluginConfigIDs) > 0 {
		pluginConfigs, err := QueryPluginConfigs(ctx, map[string]any{
			"gateway_id": gatewayID,
			"id":         pluginConfigIDs,
		})
		if err != nil {
			return err
		}
		for _, pluginConfig := range pluginConfigs {
			if resource, ok := pluginConfigMap[pluginConfig.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", pluginConfig.Name)
			}
		}
	}

	if len(consumerGroupIDs) > 0 {
		consumerGroups, err := QueryConsumerGroups(ctx, map[string]any{
			"gateway_id": gatewayID,
			"id":         consumerGroupIDs,
		})
		if err != nil {
			return err
		}
		for _, consumerGroup := range consumerGroups {
			if resource, ok := consumerGroupMap[consumerGroup.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "id", consumerGroup.ID)
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", consumerGroup.Name)
			}
		}
	}

	if len(protoIDs) > 0 {
		protos, err := QueryProtos(ctx, map[string]any{
			"gateway_id": gatewayID,
			"id":         protoIDs,
		})
		if err != nil {
			return err
		}
		for _, proto := range protos {
			if resource, ok := protoMap[proto.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", proto.Name)
			}
		}
	}

	if len(streamRouteIDs) > 0 {
		streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{
			"id": streamRouteIDs,
		})
		if err != nil {
			return err
		}
		for _, streamRoute := range streamRoutes {
			if resource, ok := streamRouteMap[streamRoute.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", streamRoute.Name)
				if labels := streamRoute.GetLabels(); labels != nil {
					resource.Config, _ = sjson.SetBytes(resource.Config, "labels", labels)
				}
			}
		}
	}

	return nil
}

// reconcilePluginMetadataSyncIDs aligns plugin_metadata snapshot IDs with the
// DB's authoritative ID when a row already exists. On entry, resource.ID is
// the etcd-key-derived name.
func reconcilePluginMetadataSyncIDs(ctx context.Context, resources []*model.GatewaySyncData) error {
	metadataByEtcdKey := make(map[string]*model.GatewaySyncData)
	var names []string

	for _, resource := range resources {
		if resource.Type != constant.PluginMetadata {
			continue
		}
		// resource.ID here is the etcd-key-derived name
		metadataByEtcdKey[resource.ID] = resource
		names = append(names, resource.ID)
	}
	if len(names) == 0 {
		return nil
	}

	metadatas, err := QueryPluginMetadatas(ctx, map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
		"name":       names,
	})
	if err != nil {
		return err
	}

	existingIDByName := make(map[string]string)
	for _, metadata := range metadatas {
		existingIDByName[metadata.Name] = metadata.ID
	}

	for name, resource := range metadataByEtcdKey {
		if existingID, ok := existingIDByName[name]; ok {
			resource.ID = existingID
			continue
		}
		resource.ID = idx.GenResourceID(constant.PluginMetadata)
	}
	return nil
}
