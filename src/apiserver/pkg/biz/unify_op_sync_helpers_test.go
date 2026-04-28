/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestBuildSyncedResourceFromKV(t *testing.T) {
	t.Parallel()

	normalizedPrefix := model.NormalizeEtcdPrefix("/apisix")

	t.Run("route strips timestamps and injects fallback name", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/routes/route-id",
			Value:       `{"uri":"/demo","create_time":111,"update_time":222}`,
			ModRevision: 9,
		})
		assert.True(t, ok)
		assert.Equal(t, "route-id", got.ID)
		assert.Equal(t, 17, got.GatewayID)
		assert.Equal(t, constant.Route, got.Type)
		assert.Equal(t, 9, got.ModRevision)
		assert.Equal(t, "routes_route-id", gjson.GetBytes(got.Config, "name").String())
		assert.False(t, gjson.GetBytes(got.Config, "create_time").Exists())
		assert.False(t, gjson.GetBytes(got.Config, "update_time").Exists())
	})

	t.Run("plugin metadata uses etcd key as snapshot name", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/plugin_metadata/clickhouse-logger",
			Value:       `{"value":{"disable":false}}`,
			ModRevision: 3,
		})
		assert.True(t, ok)
		assert.Equal(t, constant.PluginMetadata, got.Type)
		assert.Equal(t, "clickhouse-logger", got.ID)
		assert.Equal(t, "clickhouse-logger", got.GetName())
	})

	t.Run("invalid key is ignored", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/routes/too/many/parts",
			Value:       `{}`,
			ModRevision: 1,
		})
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestBackfillStoredSnapshotFields(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	pluginConfig := data.PluginConfig1WithNoRelation(gatewayInfo, constant.ResourceStatusSuccess)
	pluginConfig.Name = "pc-from-db-" + suffix
	assert.NoError(t, CreatePluginConfig(ctx, *pluginConfig))

	consumerGroup := data.ConsumerGroup1WithNoRelation(gatewayInfo, constant.ResourceStatusSuccess)
	consumerGroup.Name = "cg-from-db-" + suffix
	assert.NoError(t, CreateConsumerGroup(ctx, *consumerGroup))

	proto := data.Proto1(gatewayInfo, constant.ResourceStatusSuccess)
	proto.Name = "proto-from-db-" + suffix
	assert.NoError(t, CreateProto(ctx, *proto))

	streamRoute := data.StreamRoute1WithNoRelationResource(gatewayInfo, constant.ResourceStatusSuccess)
	streamRoute.Name = "sr-from-db-" + suffix
	streamRoute.Config, _ = sjson.SetBytes(streamRoute.Config, "labels", map[string]string{"env": "test"})
	assert.NoError(t, CreateStreamRoute(ctx, *streamRoute))

	globalRule := data.GlobalRule1(gatewayInfo, constant.ResourceStatusSuccess)
	globalRule.Name = "gr-from-db-" + suffix
	assert.NoError(t, CreateGlobalRule(ctx, *globalRule))

	resources := []*model.GatewaySyncData{
		{ID: globalRule.ID, GatewayID: gatewayInfo.ID, Type: constant.GlobalRule, Config: datatypes.JSON(`{"plugins":{}}`)},
		{ID: pluginConfig.ID, GatewayID: gatewayInfo.ID, Type: constant.PluginConfig, Config: datatypes.JSON(`{"plugins":{}}`)},
		{ID: consumerGroup.ID, GatewayID: gatewayInfo.ID, Type: constant.ConsumerGroup, Config: datatypes.JSON(`{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`)},
		{ID: proto.ID, GatewayID: gatewayInfo.ID, Type: constant.Proto, Config: datatypes.JSON(`{"content":"syntax = \"proto3\";"}`)},
		{ID: streamRoute.ID, GatewayID: gatewayInfo.ID, Type: constant.StreamRoute, Config: datatypes.JSON(`{"server_addr":"127.0.0.1","server_port":8080}`)},
	}

	err := backfillStoredSnapshotFields(ctx, resources)
	assert.NoError(t, err)
	assert.Equal(t, globalRule.Name, gjson.GetBytes(resources[0].Config, "name").String())
	assert.Equal(t, pluginConfig.Name, gjson.GetBytes(resources[1].Config, "name").String())
	assert.Equal(t, consumerGroup.ID, gjson.GetBytes(resources[2].Config, "id").String())
	assert.Equal(t, consumerGroup.Name, gjson.GetBytes(resources[2].Config, "name").String())
	assert.Equal(t, proto.Name, gjson.GetBytes(resources[3].Config, "name").String())
	assert.Equal(t, streamRoute.Name, gjson.GetBytes(resources[4].Config, "name").String())
	assert.Equal(t, "test", gjson.GetBytes(resources[4].Config, "labels.env").String())

	t.Run("missing DB row keeps original config", func(t *testing.T) {
		missingID := idx.GenResourceID(constant.PluginConfig)
		missingResources := []*model.GatewaySyncData{
			{
				ID:        missingID,
				GatewayID: gatewayInfo.ID,
				Type:      constant.PluginConfig,
				Config:    datatypes.JSON(`{"name":"keep-me","plugins":{"example":{}}}`),
			},
		}

		err := backfillStoredSnapshotFields(ctx, missingResources)
		assert.NoError(t, err)
		assert.Equal(t, "keep-me", gjson.GetBytes(missingResources[0].Config, "name").String())
		assert.True(t, gjson.GetBytes(missingResources[0].Config, "plugins.example").Exists())
	})
}

func TestReconcilePluginMetadataSyncIDs(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	existing := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusSuccess)
	existing.Name = "limit-count-" + suffix
	existing.ID = idx.GenResourceID(constant.PluginMetadata)
	assert.NoError(t, CreatePluginMetadata(ctx, *existing))

	resources := []*model.GatewaySyncData{
		{
			ID:        existing.Name,
			GatewayID: gatewayInfo.ID,
			Type:      constant.PluginMetadata,
			Config:    datatypes.JSON(`{"value":{"disable":false}}`),
		},
		{
			ID:        "new-plugin",
			GatewayID: gatewayInfo.ID,
			Type:      constant.PluginMetadata,
			Config:    datatypes.JSON(`{"value":{"disable":true}}`),
		},
	}
	resources[0].SetName(existing.Name)
	resources[1].SetName("new-plugin")

	err := reconcilePluginMetadataSyncIDs(ctx, resources)
	assert.NoError(t, err)
	assert.Equal(t, existing.ID, resources[0].ID)
	assert.NotEmpty(t, resources[1].ID)
	assert.NotEqual(t, "new-plugin", resources[1].ID)
	assert.Equal(t, "new-plugin", resources[1].GetName())
}

func TestReconcilePluginMetadataSyncIDs_AlreadyDBID(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	existing := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusSuccess)
	existing.Name = "limit-count-" + suffix
	existing.ID = idx.GenResourceID(constant.PluginMetadata)
	assert.NoError(t, CreatePluginMetadata(ctx, *existing))

	resources := []*model.GatewaySyncData{
		{
			ID:        existing.Name,
			GatewayID: gatewayInfo.ID,
			Type:      constant.PluginMetadata,
			Config:    datatypes.JSON(`{"value":{"disable":false}}`),
		},
	}
	resources[0].SetName(existing.Name)

	assert.NoError(t, reconcilePluginMetadataSyncIDs(ctx, resources))
	assert.Equal(t, existing.ID, resources[0].ID)
	assert.Equal(t, existing.Name, resources[0].GetName())
}
