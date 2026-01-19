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
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/publisher"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/goroutinex"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
)

// FuncPublishResource ...
type FuncPublishResource func(ctx context.Context, resourceIDs []string) error

// PublishResource 资源发布
func PublishResource(ctx context.Context, resourceType constant.APISIXResource, resourceIDs []string) error {
	var err error
	// 发布资源
	switch resourceType {
	case constant.Route:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishRoutes)
	case constant.Upstream:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishUpstreams)
	case constant.GlobalRule:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishGlobalRules)
	case constant.Consumer:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishConsumers)
	case constant.ConsumerGroup:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishConsumerGroups)
	case constant.PluginConfig:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishPluginConfigs)
	case constant.Service:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishServices)
	case constant.PluginMetadata:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishPluginMetadatas)
	case constant.Proto:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishProtos)
	case constant.SSL:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishSSLs)
	case constant.StreamRoute:
		err = WrapPublishResource(ctx, resourceType, resourceIDs, PublishStreamRoutes)
	}
	if err != nil {
		return err
	}
	// 主动同步一下资源
	goroutinex.GoroutineWithRecovery(ctx, func() {
		// 1s 后同步资源
		time.Sleep(time.Second * 1)
		_, err = SyncResources(ginx.CloneCtx(ctx), resourceType)
		if err != nil {
			logging.Errorf("sync resources failed, err: %v", err)
		}
	})
	return nil
}

// WrapPublishResource PublishResource 资源发布进行一些公共操作
func WrapPublishResource(ctx context.Context, resourceType constant.APISIXResource, resourceIDs []string,
	publishFunc FuncPublishResource,
) error {
	// 状态机判断
	resourceList, err := BatchGetResources(ctx, resourceType, resourceIDs)
	if err != nil {
		logging.ErrorFWithContext(ctx, "%s query err: %s", resourceType, err.Error())
		return fmt.Errorf("%s 查询错误: %w", constant.ResourceTypeMap[resourceType], err)
	}
	if len(resourceList) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no %s found for the specified resourceIDs %v",
			resourceType,
			resourceIDs,
		)
		return fmt.Errorf("未找到指定的 %s 资源 IDs %v", constant.ResourceTypeMap[resourceType], resourceIDs)
	}
	resourceStatusMap := make(map[string]constant.ResourceStatus)
	for _, resource := range resourceList {
		statusOp := status.NewResourceStatusOp(*resource)
		nextStatus, err := statusOp.NextStatus(ctx, constant.OperationTypePublish)
		if err != nil {
			logging.ErrorFWithContext(ctx,
				"resource: %s can not be publish: %s", resource.GetName(resourceType), err.Error())
			return fmt.Errorf("资源: %s 不能发布: %w", resource.GetName(resourceType), err)
		}
		// 发布之后的状态映射
		resourceStatusMap[resource.ID] = nextStatus
	}
	err = publishFunc(ctx, resourceIDs)
	if err != nil {
		return err
	}
	err = AddBatchAuditLog(ctx, constant.OperationTypePublish, resourceType, resourceList, resourceStatusMap)
	if err != nil {
		logging.ErrorFWithContext(ctx, "%s add audit log err: %s", resourceType, err.Error())
		return err
	}
	return nil
}

// PublishAllResource 资源一键发布
func PublishAllResource(ctx context.Context, gatewayID int) error {
	for _, resourceType := range constant.ResourceTypeList {
		resources, err := QueryResource(ctx, resourceType,
			map[string]any{
				"gateway_id": gatewayID,
				"status": []constant.ResourceStatus{
					constant.ResourceStatusCreateDraft,
					constant.ResourceStatusUpdateDraft,
					constant.ResourceStatusDeleteDraft,
				},
			}, "")
		if err != nil {
			logging.ErrorFWithContext(ctx, "%s query err: %s", resourceType, err.Error())
			return fmt.Errorf("%s 查询错误: %w", constant.ResourceTypeMap[resourceType], err)
		}
		if len(resources) == 0 {
			continue
		}
		resourceIDs := make([]string, 0)
		for _, resource := range resources {
			resourceIDs = append(resourceIDs, resource.ID)
		}
		err = PublishResource(ctx, resourceType, resourceIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishRoutes 路由发布
func PublishRoutes(ctx context.Context, routeIDs []string) error {
	routes, err := QueryRoutes(ctx, map[string]any{"id": routeIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "routes query err: %s", err.Error())
		return fmt.Errorf("路由查询错误：%w", err)
	}
	if len(routes) == 0 {
		logging.ErrorFWithContext(ctx, "no routes found for the specified routeIDs %v", routeIDs)
		return fmt.Errorf("未找到指定的路由资源 IDs %v", routeIDs)
	}
	var deleteRouteIDs []string
	var addRouteIDs []string
	for _, route := range routes {
		if route.Status == constant.ResourceStatusDeleteDraft {
			deleteRouteIDs = append(deleteRouteIDs, route.ID)
			continue
		}
		addRouteIDs = append(addRouteIDs, route.ID)
	}
	if len(deleteRouteIDs) > 0 {
		err = deleteRoutes(ctx, deleteRouteIDs)
		if err != nil {
			return err
		}
	}
	if len(addRouteIDs) > 0 {
		err = putRoutes(ctx, addRouteIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishServices 发布 service
func PublishServices(ctx context.Context, serviceIDs []string) error {
	services, err := QueryServices(ctx, map[string]any{"id": serviceIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "services query err: %s", err.Error())
		return fmt.Errorf("服务查询错误：%w", err)
	}
	if len(services) == 0 {
		logging.ErrorFWithContext(ctx, "no services found for the specified serviceIDs %v", serviceIDs)
		return fmt.Errorf("未找到指定的服务资源 IDs %v", serviceIDs)
	}
	var deleteServiceIDs []string
	var addServiceIDs []string
	for _, service := range services {
		if service.Status == constant.ResourceStatusDeleteDraft {
			deleteServiceIDs = append(deleteServiceIDs, service.ID)
			continue
		}
		addServiceIDs = append(addServiceIDs, service.ID)
	}
	if len(deleteServiceIDs) > 0 {
		err = deleteServices(ctx, deleteServiceIDs)
		if err != nil {
			return err
		}
	}
	if len(addServiceIDs) > 0 {
		err = putServices(ctx, addServiceIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishUpstreams 发布 upstream
func PublishUpstreams(ctx context.Context, upstreamIDs []string) error {
	upstreams, err := QueryUpstreams(ctx, map[string]any{"id": upstreamIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "upstreams query err: %s", err.Error())
		return fmt.Errorf("上游查询错误：%w", err)
	}
	if len(upstreams) == 0 {
		logging.ErrorFWithContext(ctx, "no upstreams found for the specified upstreamIDs %v", upstreamIDs)
		return fmt.Errorf("未找到指定的上游资源 IDs %v", upstreamIDs)
	}
	var deleteUpstreamIDs []string
	var addUpstreamIDs []string
	for _, upstream := range upstreams {
		if upstream.Status == constant.ResourceStatusDeleteDraft {
			deleteUpstreamIDs = append(deleteUpstreamIDs, upstream.ID)
			continue
		}
		addUpstreamIDs = append(addUpstreamIDs, upstream.ID)
	}
	if len(deleteUpstreamIDs) > 0 {
		err = deleteUpstreams(ctx, deleteUpstreamIDs)
		if err != nil {
			return err
		}
	}
	if len(addUpstreamIDs) > 0 {
		err = putUpstreams(ctx, addUpstreamIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishPluginConfigs 发布 pluginConfig
func PublishPluginConfigs(ctx context.Context, pluginConfigIDs []string) error {
	pluginConfigs, err := QueryPluginConfigs(ctx, map[string]any{"id": pluginConfigIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "pluginConfigs query err: %s", err.Error())
		return fmt.Errorf("插件组查询错误：%w", err)
	}
	if len(pluginConfigs) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no pluginConfigs found for the specified pluginConfigIDs %v",
			pluginConfigIDs,
		)
		return fmt.Errorf("未找到指定的插件组资源 IDs %v", pluginConfigIDs)
	}
	var deletePluginConfigIDs []string
	var addPluginConfigIDs []string
	for _, pluginConfig := range pluginConfigs {
		if pluginConfig.Status == constant.ResourceStatusDeleteDraft {
			deletePluginConfigIDs = append(deletePluginConfigIDs, pluginConfig.ID)
			continue
		}
		addPluginConfigIDs = append(addPluginConfigIDs, pluginConfig.ID)
	}
	if len(deletePluginConfigIDs) > 0 {
		err = deletePluginConfigs(ctx, deletePluginConfigIDs)
		if err != nil {
			return err
		}
	}
	if len(addPluginConfigIDs) > 0 {
		err = putPluginConfigs(ctx, addPluginConfigIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishConsumers 发布 consumer
func PublishConsumers(ctx context.Context, consumerIDs []string) error {
	consumers, err := QueryConsumers(ctx, map[string]any{"id": consumerIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "consumers query err: %s", err.Error())
		return fmt.Errorf("消费者查询错误：%w", err)
	}
	if len(consumers) == 0 {
		logging.ErrorFWithContext(ctx, "no consumers found for the specified consumerIDs %v", consumerIDs)
		return fmt.Errorf("未找到指定的消费者资源 IDs %v", consumerIDs)
	}
	var deleteConsumerIDs []string
	var addConsumerIDs []string
	for _, consumer := range consumers {
		if consumer.Status == constant.ResourceStatusDeleteDraft {
			deleteConsumerIDs = append(deleteConsumerIDs, consumer.ID)
			continue
		}
		addConsumerIDs = append(addConsumerIDs, consumer.ID)
	}
	if len(deleteConsumerIDs) > 0 {
		err = deleteConsumers(ctx, deleteConsumerIDs)
		if err != nil {
			return err
		}
	}
	if len(addConsumerIDs) > 0 {
		err = putConsumers(ctx, addConsumerIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishConsumerGroups 发布 consumerGroup
func PublishConsumerGroups(ctx context.Context, consumerGroupIDs []string) error {
	consumerGroups, err := QueryConsumerGroups(ctx, map[string]any{"id": consumerGroupIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "consumerGroups query err: %s", err.Error())
		return fmt.Errorf("消费者组查询错误：%w", err)
	}
	if len(consumerGroups) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no consumerGroups found for the specified consumerGroupIDs %v",
			consumerGroupIDs,
		)
		return fmt.Errorf("未找到指定的消费者组资源 IDs %v", consumerGroupIDs)
	}
	var deleteConsumerGroupIDs []string
	var addConsumerGroupIDs []string
	for _, consumerGroup := range consumerGroups {
		if consumerGroup.Status == constant.ResourceStatusDeleteDraft {
			deleteConsumerGroupIDs = append(deleteConsumerGroupIDs, consumerGroup.ID)
			continue
		}
		addConsumerGroupIDs = append(addConsumerGroupIDs, consumerGroup.ID)
	}
	if len(deleteConsumerGroupIDs) > 0 {
		err = deleteConsumerGroups(ctx, deleteConsumerGroupIDs)
		if err != nil {
			return err
		}
	}
	if len(addConsumerGroupIDs) > 0 {
		err = putConsumerGroups(ctx, addConsumerGroupIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishGlobalRules 发布 globalRule
func PublishGlobalRules(ctx context.Context, globalRuleIDs []string) error {
	globalRules, err := QueryGlobalRules(ctx, map[string]any{"id": globalRuleIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "globalRules query err: %s", err.Error())
		return fmt.Errorf("全局规则查询错误：%w", err)
	}
	if len(globalRules) == 0 {
		logging.ErrorFWithContext(ctx, "no globalRules found for the specified globalRuleIDs %v", globalRuleIDs)
		return fmt.Errorf("未找到指定的全局规则资源 IDs %v", globalRuleIDs)
	}
	var deleteGlobalRuleIDs []string
	var addGlobalRuleIDs []string
	for _, globalRule := range globalRules {
		if globalRule.Status == constant.ResourceStatusDeleteDraft {
			deleteGlobalRuleIDs = append(deleteGlobalRuleIDs, globalRule.ID)
			continue
		}
		addGlobalRuleIDs = append(addGlobalRuleIDs, globalRule.ID)
	}
	if len(deleteGlobalRuleIDs) > 0 {
		err = deleteGlobalRules(ctx, deleteGlobalRuleIDs)
		if err != nil {
			return err
		}
	}
	if len(addGlobalRuleIDs) > 0 {
		err = putGlobalRules(ctx, addGlobalRuleIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishPluginMetadatas 发布 pluginMetadata
func PublishPluginMetadatas(ctx context.Context, pluginMetadataIDs []string) error {
	pluginMetadatas, err := QueryPluginMetadatas(ctx, map[string]any{"id": pluginMetadataIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "pluginMetadatas query err: %s", err.Error())
		return fmt.Errorf("插件元数据查询错误：%w", err)
	}
	if len(pluginMetadatas) == 0 {
		logging.ErrorFWithContext(ctx,
			"no pluginMetadatas found for the specified pluginMetadataIDs %v", pluginMetadataIDs)
		return fmt.Errorf("未找到指定的插件元数据资源 IDs %v", pluginMetadataIDs)
	}
	var deletePluginMetadataIDs []string
	var addPluginMetadataIDs []string
	for _, pluginMetadata := range pluginMetadatas {
		if pluginMetadata.Status == constant.ResourceStatusDeleteDraft {
			deletePluginMetadataIDs = append(deletePluginMetadataIDs, pluginMetadata.ID)
			continue
		}
		addPluginMetadataIDs = append(addPluginMetadataIDs, pluginMetadata.ID)
	}
	if len(deletePluginMetadataIDs) > 0 {
		err = deletePluginMetadatas(ctx, deletePluginMetadataIDs)
		if err != nil {
			return err
		}
	}
	if len(addPluginMetadataIDs) > 0 {
		err = putPluginMetadatas(ctx, addPluginMetadataIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishProtos 发布 Proto
func PublishProtos(ctx context.Context, protoIDs []string) error {
	protos, err := QueryProtos(ctx, map[string]any{"id": protoIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "protos query err: %s", err.Error())
		return fmt.Errorf("protos 查询错误：%w", err)
	}
	if len(protos) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no protos found for the specified protoIDs %v",
			protoIDs,
		)
		return fmt.Errorf("未找到指定的 protos 资源 IDs %v", protoIDs)
	}
	var deleteProtoIDs []string
	var addProtoIDs []string
	for _, pb := range protos {
		if pb.Status == constant.ResourceStatusDeleteDraft {
			deleteProtoIDs = append(deleteProtoIDs, pb.ID)
			continue
		}
		addProtoIDs = append(addProtoIDs, pb.ID)
	}
	if len(deleteProtoIDs) > 0 {
		err = deleteProtos(ctx, deleteProtoIDs)
		if err != nil {
			return err
		}
	}
	if len(addProtoIDs) > 0 {
		err = PutProtos(ctx, addProtoIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishSSLs 发布 ssls
func PublishSSLs(ctx context.Context, sslIDs []string) error {
	ssls, err := QuerySSL(ctx, map[string]any{"id": sslIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "ssls query err: %s", err.Error())
		return fmt.Errorf("ssls 查询错误：%w", err)
	}
	if len(ssls) == 0 {
		logging.ErrorFWithContext(ctx, "no ssls found for the specified sslIDs %v", sslIDs)
		return fmt.Errorf("未找到指定的 ssls 资源 IDs %v", sslIDs)
	}
	var deleteSSLIDs []string
	var addSSLIDs []string
	for _, ssl := range ssls {
		if ssl.Status == constant.ResourceStatusDeleteDraft {
			deleteSSLIDs = append(deleteSSLIDs, ssl.ID)
			continue
		}
		addSSLIDs = append(addSSLIDs, ssl.ID)
	}
	if len(deleteSSLIDs) > 0 {
		err = deleteSSLs(ctx, deleteSSLIDs)
		if err != nil {
			return err
		}
	}
	if len(addSSLIDs) > 0 {
		err = PutSSLs(ctx, addSSLIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// PublishStreamRoutes 发布 StreamRoute
func PublishStreamRoutes(ctx context.Context, streamRouteIDs []string) error {
	streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{"id": streamRouteIDs})
	if err != nil {
		logging.ErrorFWithContext(ctx, "streamRoutes query err: %s", err.Error())
		return fmt.Errorf("streamRoutes 查询错误：%w", err)
	}
	if len(streamRoutes) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no streamRoutes found for the specified streamRouteIDs %v",
			streamRouteIDs,
		)
		return fmt.Errorf("未找到指定的 streamRoutes 资源 IDs %v", streamRouteIDs)
	}
	var deleteStreamRouteIDs []string
	var addStreamRouteIDs []string
	for _, sr := range streamRoutes {
		if sr.Status == constant.ResourceStatusDeleteDraft {
			deleteStreamRouteIDs = append(deleteStreamRouteIDs, sr.ID)
			continue
		}
		addStreamRouteIDs = append(addStreamRouteIDs, sr.ID)
	}
	if len(deleteStreamRouteIDs) > 0 {
		err = deleteStreamRoutes(ctx, deleteStreamRouteIDs)
		if err != nil {
			return err
		}
	}
	if len(addStreamRouteIDs) > 0 {
		err = PutStreamRoutes(ctx, addStreamRouteIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

// getEtcdPublisher 获取 publisher
func getEtcdPublisher(ctx context.Context) (*publisher.EtcdPublisher, error) {
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	pub, err := publisher.NewEtcdPublisher(ctx, gatewayInfo)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

func batchCreateEtcdResource(ctx context.Context, ops []publisher.ResourceOperation) error {
	etcdPublisher, err := getEtcdPublisher(ctx)
	if err != nil {
		return err
	}
	return etcdPublisher.BatchCreate(ctx, ops)
}

func batchDeleteEtcdResource(ctx context.Context, resourceType constant.APISIXResource, ids []string) error {
	pub, err := getEtcdPublisher(ctx)
	if err != nil {
		return err
	}
	var ops []publisher.ResourceOperation
	for _, id := range ids {
		ops = append(ops, publisher.ResourceOperation{
			Type: resourceType,
			Key:  id,
		})
	}
	err = pub.BatchDelete(ctx, ops)
	if err != nil {
		logging.ErrorFWithContext(ctx, "etcd deletes associated data err: %s", err.Error())
		return fmt.Errorf("etcd 删除关联数据错误：%w", err)
	}
	return nil
}

// deleteRoutes 删除 route
func deleteRoutes(ctx context.Context, routeIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.Route, routeIDs)
	if err != nil {
		return err
	}
	// 删除数据库数据
	return BatchDeleteRoutes(ctx, routeIDs)
}

// deleteServices 删除 service
func deleteServices(ctx context.Context, serviceIDs []string) error {
	// 先判断 service 有没有关联的资源数据
	routes, err := QueryRoutes(ctx, map[string]any{"service_id": serviceIDs})
	if err != nil {
		return err
	}
	if len(routes) > 0 {
		return fmt.Errorf("服务不可删除, 存在关联的路由资源 %v", FormatResourceIDNameList(routes, constant.Route))
	}
	// 判断 service 有没有关联的 streamRoute 数据
	streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{"service_id": serviceIDs})
	if err != nil {
		return err
	}
	if len(streamRoutes) > 0 {
		return fmt.Errorf(
			"服务不可删除, 存在关联的 streamRoute 资源 %v",
			FormatResourceIDNameList(streamRoutes, constant.StreamRoute),
		)
	}
	// 先删除 etcd 的数据
	err = batchDeleteEtcdResource(ctx, constant.Service, serviceIDs)
	if err != nil {
		return err
	}
	// 删除数据库数据
	return BatchDeleteServices(ctx, serviceIDs)
}

// deleteUpstreams 删除 upstream
func deleteUpstreams(ctx context.Context, upstreamIDs []string) error {
	// 判断 upstream 有没有关联的 service 数据
	services, err := QueryServices(ctx, map[string]any{"upstream_id": upstreamIDs})
	if err != nil {
		return err
	}
	if len(services) > 0 {
		return fmt.Errorf("上游不可删除, 存在关联的服务资源 %v", FormatResourceIDNameList(services, constant.Service))
	}
	// 判断 upstream 有没有关联的 route 数据
	routes, err := QueryRoutes(ctx, map[string]any{"upstream_id": upstreamIDs})
	if err != nil {
		return err
	}
	if len(routes) > 0 {
		return fmt.Errorf("上游不可删除, 存在关联的路由资源 %v", FormatResourceIDNameList(routes, constant.Route))
	}
	// 判断 upstream 有没有关联的 streamRoute 数据
	streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{"upstream_id": upstreamIDs})
	if err != nil {
		return err
	}
	if len(streamRoutes) > 0 {
		return fmt.Errorf(
			"上游不可删除, 存在关联的 streamRoute 资源 %v",
			FormatResourceIDNameList(streamRoutes, constant.StreamRoute),
		)
	}

	// 先删除 etcd 的数据
	err = batchDeleteEtcdResource(ctx, constant.Upstream, upstreamIDs)
	if err != nil {
		return err
	}

	// 删除数据库数据
	return BatchDeleteUpstreams(ctx, upstreamIDs)
}

// deletePluginConfigs 删除 pluginConfig
func deletePluginConfigs(ctx context.Context, pluginConfigIDs []string) error {
	// 判断 plugin_config 有没有关联的 route 数据
	routes, err := QueryRoutes(ctx, map[string]any{"plugin_config_id": pluginConfigIDs})
	if err != nil {
		return err
	}
	if len(routes) > 0 {
		return fmt.Errorf("插件组不可删除, 存在关联的路由资源 %v", FormatResourceIDNameList(routes, constant.Route))
	}

	// 先删除 etcd 的数据
	err = batchDeleteEtcdResource(ctx, constant.PluginConfig, pluginConfigIDs)
	if err != nil {
		return err
	}

	// 删除数据库数据
	return BatchDeletePluginConfigs(ctx, pluginConfigIDs)
}

// deletePluginMetadatas 删除 pluginMetadata
func deletePluginMetadatas(ctx context.Context, pluginMetadataIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.PluginMetadata, pluginMetadataIDs)
	if err != nil {
		return err
	}

	// 删除数据库数据
	return BatchDeletePluginMetadatas(ctx, pluginMetadataIDs)
}

// deleteConsumers 删除 consumer
func deleteConsumers(ctx context.Context, consumerIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.Consumer, consumerIDs)
	if err != nil {
		return err
	}

	// 删除数据库数据
	return BatchDeleteConsumers(ctx, consumerIDs)
}

// deleteConsumerGroups 删除 consumerGroup
func deleteConsumerGroups(ctx context.Context, consumerGroupIDs []string) error {
	consumers, err := QueryConsumers(ctx, map[string]any{"group_id": consumerGroupIDs})
	if err != nil {
		return err
	}
	if len(consumers) > 0 {
		return fmt.Errorf("消费者组不可删除, 存在关联的消费者资源 %v", FormatResourceIDNameList(consumers, constant.Consumer))
	}
	// 先删除 etcd 的数据
	err = batchDeleteEtcdResource(ctx, constant.ConsumerGroup, consumerGroupIDs)
	if err != nil {
		return err
	}

	return BatchDeleteConsumerGroups(ctx, consumerGroupIDs)
}

// deleteGlobalRules 删除 globalRule
func deleteGlobalRules(ctx context.Context, globalRuleIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.GlobalRule, globalRuleIDs)
	if err != nil {
		return err
	}
	// 删除数据库数据
	return BatchDeleteGlobalRules(ctx, globalRuleIDs)
}

// deleteProtos 删除 Proto
func deleteProtos(ctx context.Context, protoIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.Proto, protoIDs)
	if err != nil {
		return err
	}
	return BatchDeleteProtos(ctx, protoIDs)
}

// deleteSSLs 删除 SSL
func deleteSSLs(ctx context.Context, sslIDs []string) error {
	upstreams, err := QueryUpstreams(ctx, map[string]any{"ssl_id": sslIDs})
	if err != nil {
		return err
	}
	if len(upstreams) > 0 {
		return fmt.Errorf("ssl 不可删除, 存在关联的上游资源 %v", FormatResourceIDNameList(upstreams, constant.Upstream))
	}
	// 先删除 etcd 的数据
	err = batchDeleteEtcdResource(ctx, constant.SSL, sslIDs)
	if err != nil {
		return err
	}
	return BatchDeleteSSL(ctx, sslIDs)
}

// deleteStreamRoutes 删除 StreamRoute
func deleteStreamRoutes(ctx context.Context, streamRouteIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.StreamRoute, streamRouteIDs)
	if err != nil {
		return err
	}
	return BatchDeleteStreamRoutes(ctx, streamRouteIDs)
}

// putRoutes 发布路由
func putRoutes(ctx context.Context, routeIDs []string) error {
	routes, err := QueryRoutes(ctx, map[string]any{"id": routeIDs})
	if err != nil {
		return err
	}
	if len(routes) == 0 {
		logging.ErrorFWithContext(ctx, "no routes found for the specified routeIDs %v", routeIDs)
		return fmt.Errorf("未找到指定的路由资源 IDs %v", routeIDs)
	}
	var serviceIDs []string
	var upstreamIDs []string
	var pluginConfigIDs []string
	var routeOps []publisher.ResourceOperation
	for _, route := range routes {
		baseInfo := entity.BaseInfo{
			ID:         route.ID,
			Name:       route.Name,
			CreateTime: route.CreatedAt.Unix(),
			UpdateTime: route.UpdatedAt.Unix(),
		}
		if route.ServiceID != "" {
			serviceIDs = append(serviceIDs, route.ServiceID)
		}
		if route.UpstreamID != "" {
			upstreamIDs = append(upstreamIDs, route.UpstreamID)
		}
		if route.PluginConfigID != "" {
			pluginConfigIDs = append(pluginConfigIDs, route.PluginConfigID)
		}
		baseConfig, _ := json.Marshal(baseInfo)
		route.Config, err = jsonx.MergeJson(route.Config, baseConfig)
		if err != nil {
			return err
		}
		routeOps = append(routeOps, publisher.ResourceOperation{
			Key:    route.ID,
			Config: json.RawMessage(route.Config),
			Type:   constant.Route,
		})
	}
	// 发布 upstream
	if len(upstreamIDs) > 0 {
		if err := putUpstreams(ctx, upstreamIDs); err != nil {
			return err
		}
	}

	// 发布 service
	if len(serviceIDs) > 0 {
		if err := putServices(ctx, serviceIDs); err != nil {
			return err
		}
	}

	// 发布 pluginConfig
	if len(pluginConfigIDs) > 0 {
		if err := putPluginConfigs(ctx, pluginConfigIDs); err != nil {
			return err
		}
	}

	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, routeOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.Route, routeIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "routes status change err: %s", err.Error())
		return fmt.Errorf("路由发布错误：%w", err)
	}
	return nil
}

func putServices(ctx context.Context, serviceIDs []string) error {
	services, err := QueryServices(ctx, map[string]any{"id": serviceIDs})
	if err != nil {
		return err
	}
	if len(services) == 0 {
		logging.ErrorFWithContext(ctx, "no services found for the specified serviceIDs %v", serviceIDs)
		return fmt.Errorf("未找到指定的服务资源 IDs %v", serviceIDs)
	}
	var upstreamIDs []string
	var serviceOps []publisher.ResourceOperation
	for _, service := range services {
		baseInfo := entity.BaseInfo{
			ID:         service.ID,
			CreateTime: service.CreatedAt.Unix(),
			UpdateTime: service.UpdatedAt.Unix(),
		}
		if service.UpstreamID != "" {
			upstreamIDs = append(upstreamIDs, service.UpstreamID)
			baseConfig, _ := json.Marshal(baseInfo)
			service.Config, err = jsonx.MergeJson(service.Config, baseConfig)
			if err != nil {
				return err
			}
		}
		serviceOps = append(serviceOps, publisher.ResourceOperation{
			Key:    service.ID,
			Config: json.RawMessage(service.Config),
			Type:   constant.Service,
		})
	}
	// 发布 upstream
	if len(upstreamIDs) > 0 {
		if err = putUpstreams(ctx, upstreamIDs); err != nil {
			return err
		}
	}

	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, serviceOps)
	if err != nil {
		return err
	}

	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.Service, serviceIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "services status change err: %s", err.Error())
		return fmt.Errorf("服务发布错误：%w", err)
	}
	return nil
}

func putUpstreams(ctx context.Context, upstreamIDs []string) error {
	upstreams, err := QueryUpstreams(ctx, map[string]any{"id": upstreamIDs})
	if err != nil {
		return err
	}
	if len(upstreams) == 0 {
		logging.ErrorFWithContext(ctx, "no upstreams found for the specified upstreamIDs %v", upstreamIDs)
		return fmt.Errorf("未找到指定的上游资源 IDs %v", upstreamIDs)
	}
	var upstreamOps []publisher.ResourceOperation
	var sslIDs []string
	for _, upstream := range upstreams {
		baseInfo := entity.BaseInfo{
			ID:         upstream.ID,
			CreateTime: upstream.CreatedAt.Unix(),
			UpdateTime: upstream.UpdatedAt.Unix(),
		}
		baseConfig, _ := json.Marshal(baseInfo)
		upstream.Config, err = jsonx.MergeJson(upstream.Config, baseConfig)
		if err != nil {
			return err
		}
		if upstream.GetSSLID() != "" {
			sslIDs = append(sslIDs, upstream.GetSSLID())
		}
		upstreamOps = append(upstreamOps, publisher.ResourceOperation{
			Key:    upstream.ID,
			Config: json.RawMessage(upstream.Config),
			Type:   constant.Upstream,
		})
	}
	if len(sslIDs) > 0 {
		if err = PutSSLs(ctx, sslIDs); err != nil {
			return err
		}
	}
	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, upstreamOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.Upstream, upstreamIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "upstreams status change err: %s", err.Error())
		return fmt.Errorf("上游发布错误：%w", err)
	}
	return nil
}

// putPluginConfigs ...
func putPluginConfigs(ctx context.Context, pluginConfigIDs []string) error {
	pluginConfigs, err := QueryPluginConfigs(ctx, map[string]any{"id": pluginConfigIDs})
	if err != nil {
		return err
	}
	if len(pluginConfigs) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no pluginConfigs found for the specified pluginConfigIDs %v",
			pluginConfigIDs,
		)
		return fmt.Errorf("未找到指定的插件组资源 IDs %v", pluginConfigIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var pluginConfigOps []publisher.ResourceOperation
	for _, pluginConfig := range pluginConfigs {
		baseInfo := entity.BaseInfo{
			ID:         pluginConfig.ID,
			CreateTime: pluginConfig.CreatedAt.Unix(),
			UpdateTime: pluginConfig.UpdatedAt.Unix(),
		}
		baseConfig, _ := json.Marshal(baseInfo)
		pluginConfig.Config, err = jsonx.MergeJson(pluginConfig.Config, baseConfig)

		// Version-aware field cleanup: only remove fields that are invalid for this APISIX version
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.PluginConfig, "id", apisixVersion) {
			pluginConfig.Config, _ = sjson.DeleteBytes(pluginConfig.Config, "id")
		}
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.PluginConfig, "name", apisixVersion) {
			pluginConfig.Config, _ = sjson.DeleteBytes(pluginConfig.Config, "name")
		}

		if err != nil {
			return err
		}
		pluginConfigOps = append(pluginConfigOps, publisher.ResourceOperation{
			Key:    pluginConfig.ID,
			Config: json.RawMessage(pluginConfig.Config),
			Type:   constant.PluginConfig,
		})
	}

	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, pluginConfigOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.PluginConfig, pluginConfigIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "pluginConfigs status change err: %s", err.Error())
		return fmt.Errorf("插件组发布错误：%w", err)
	}
	return nil
}

// putPluginMetadatas ...
func putPluginMetadatas(ctx context.Context, pluginMetadataIDs []string) error {
	pluginMetadatas, err := QueryPluginMetadatas(ctx, map[string]any{"id": pluginMetadataIDs})
	if err != nil {
		return err
	}
	if len(pluginMetadatas) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no pluginMetadatas found for the specified pluginMetadataIDs %v",
			pluginMetadataIDs,
		)
		return fmt.Errorf("未找到指定的插件元数据资源 IDs %v", pluginMetadataIDs)
	}
	var pluginMetadataOps []publisher.ResourceOperation
	for _, pluginMetadata := range pluginMetadatas {
		baseInfo := entity.BaseInfo{
			ID:         pluginMetadata.Name, // pluginMetadata.Name 必须是 pluginName
			CreateTime: pluginMetadata.CreatedAt.Unix(),
			UpdateTime: pluginMetadata.UpdatedAt.Unix(),
		}
		baseConfig, _ := json.Marshal(baseInfo)
		pluginMetadata.Config, err = jsonx.MergeJson(pluginMetadata.Config, baseConfig)
		if err != nil {
			return err
		}
		pluginMetadataOps = append(pluginMetadataOps, publisher.ResourceOperation{
			Key:    pluginMetadata.Name,
			Config: json.RawMessage(pluginMetadata.Config),
			Type:   constant.PluginMetadata,
		})
	}
	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, pluginMetadataOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.PluginMetadata, pluginMetadataIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "pluginMetadatas status change err: %s", err.Error())
		return fmt.Errorf("插件元数据发布错误：%w", err)
	}
	return nil
}

// putConsumers ...
func putConsumers(ctx context.Context, consumerIDs []string) error {
	consumers, err := QueryConsumers(ctx, map[string]any{"id": consumerIDs})
	if err != nil {
		return err
	}
	if len(consumers) == 0 {
		logging.ErrorFWithContext(ctx, "no consumers found for the specified consumerIDs %v", consumerIDs)
		return fmt.Errorf("未找到指定的消费者资源 IDs %v", consumerIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var consumerOps []publisher.ResourceOperation
	var consumerGroupIDs []string
	for _, consumer := range consumers {
		baseInfo := entity.BaseInfo{
			CreateTime: consumer.CreatedAt.Unix(),
			UpdateTime: consumer.UpdatedAt.Unix(),
		}
		if consumer.GroupID != "" {
			consumerGroupIDs = append(consumerGroupIDs, consumer.GroupID)
		}
		baseConfig, _ := json.Marshal(baseInfo)
		consumer.Config, err = jsonx.MergeJson(consumer.Config, baseConfig)

		// Version-aware field cleanup: consumer uses username as identifier, id should always be removed
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.Consumer, "id", apisixVersion) {
			consumer.Config, _ = sjson.DeleteBytes(consumer.Config, "id")
		}

		if err != nil {
			return err
		}
		consumerOps = append(consumerOps, publisher.ResourceOperation{
			Key:    consumer.ID,
			Config: json.RawMessage(consumer.Config),
			Type:   constant.Consumer,
		})
	}

	if len(consumerGroupIDs) > 0 {
		err = putConsumerGroups(ctx, consumerGroupIDs)
		if err != nil {
			return err
		}
	}

	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, consumerOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.Consumer, consumerIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "consumers status change err: %s", err.Error())
		return fmt.Errorf("消费者发布错误：%w", err)
	}
	return nil
}

// putConsumerGroups ...
func putConsumerGroups(ctx context.Context, consumerGroupIDs []string) error {
	consumerGroups, err := QueryConsumerGroups(ctx, map[string]any{"id": consumerGroupIDs})
	if err != nil {
		return err
	}
	if len(consumerGroups) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no consumerGroups found for the specified consumerGroupIDs %v",
			consumerGroupIDs,
		)
		return fmt.Errorf("未找到指定的消费者组资源 IDs %v", consumerGroupIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var consumerGroupOps []publisher.ResourceOperation
	for _, consumerGroup := range consumerGroups {
		baseInfo := entity.BaseInfo{
			ID:         consumerGroup.ID,
			CreateTime: consumerGroup.CreatedAt.Unix(),
			UpdateTime: consumerGroup.UpdatedAt.Unix(),
		}
		baseConfig, _ := json.Marshal(baseInfo)
		consumerGroup.Config, err = jsonx.MergeJson(consumerGroup.Config, baseConfig)

		// Version-aware field cleanup: consumer_group requires id in schema
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.ConsumerGroup, "id", apisixVersion) {
			consumerGroup.Config, _ = sjson.DeleteBytes(consumerGroup.Config, "id")
		}
		// name is only valid in 3.13+
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.ConsumerGroup, "name", apisixVersion) {
			consumerGroup.Config, _ = sjson.DeleteBytes(consumerGroup.Config, "name")
		}

		if err != nil {
			return err
		}
		consumerGroupOps = append(consumerGroupOps, publisher.ResourceOperation{
			Key:    consumerGroup.ID,
			Config: json.RawMessage(consumerGroup.Config),
			Type:   constant.ConsumerGroup,
		})
	}

	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, consumerGroupOps)
	if err != nil {
		return err
	}

	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.ConsumerGroup, consumerGroupIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "consumerGroups status change err: %s", err.Error())
		return fmt.Errorf("消费者组发布错误：%w", err)
	}
	return nil
}

// putGlobalRules ...
func putGlobalRules(ctx context.Context, globalRuleIDs []string) error {
	globalRules, err := QueryGlobalRules(ctx, map[string]any{"id": globalRuleIDs})
	if err != nil {
		return err
	}
	if len(globalRules) == 0 {
		logging.ErrorFWithContext(ctx, "no globalRules found for the specified globalRuleIDs %v", globalRuleIDs)
		return fmt.Errorf("未找到指定的全局规则资源 IDs %v", globalRuleIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var globalRuleOps []publisher.ResourceOperation
	for _, globalRule := range globalRules {
		baseInfo := entity.BaseInfo{
			ID:         globalRule.ID,
			CreateTime: globalRule.CreatedAt.Unix(),
			UpdateTime: globalRule.UpdatedAt.Unix(),
		}
		baseConfig, _ := json.Marshal(baseInfo)
		globalRule.Config, err = jsonx.MergeJson(globalRule.Config, baseConfig)

		// Version-aware field cleanup: global_rule never supports name in any version
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.GlobalRule, "name", apisixVersion) {
			globalRule.Config, _ = sjson.DeleteBytes(globalRule.Config, "name")
		}

		if err != nil {
			return err
		}
		globalRuleOps = append(globalRuleOps, publisher.ResourceOperation{
			Key:    globalRule.ID,
			Config: json.RawMessage(globalRule.Config),
			Type:   constant.GlobalRule,
		})
	}
	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, globalRuleOps)
	if err != nil {
		return err
	}

	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.GlobalRule, globalRuleIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "globalRules status change err: %s", err.Error())
		return fmt.Errorf("全局规则发布错误：%w", err)
	}
	return nil
}

// PutProtos  ...
func PutProtos(ctx context.Context, protoIDs []string) error {
	protos, err := QueryProtos(ctx, map[string]any{"id": protoIDs})
	if err != nil {
		return err
	}
	if len(protos) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no Protos found for the specified protoIDs %v",
			protoIDs,
		)
		return fmt.Errorf("未找到指定的 protos 资源 IDs %v", protoIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var protoOps []publisher.ResourceOperation
	for _, pb := range protos {
		baseInfo := entity.BaseInfo{
			ID:         pb.ID,
			CreateTime: pb.CreatedAt.Unix(),
			UpdateTime: pb.UpdatedAt.Unix(),
		}
		baseConfig, _ := json.Marshal(baseInfo)
		pb.Config, err = jsonx.MergeJson(pb.Config, baseConfig)
		if err != nil {
			return err
		}

		// Version-aware field cleanup: proto name is only supported in 3.13+
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.Proto, "name", apisixVersion) {
			pb.Config, _ = sjson.DeleteBytes(pb.Config, "name")
		}

		protoOps = append(protoOps, publisher.ResourceOperation{
			Key:    pb.ID,
			Config: json.RawMessage(pb.Config),
			Type:   constant.Proto,
		})
	}

	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, protoOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.Proto, protoIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "Protos status change err: %s", err.Error())
		return fmt.Errorf("protos 发布错误：%w", err)
	}
	return nil
}

// PutSSLs ...
func PutSSLs(ctx context.Context, sslIDs []string) error {
	ssls, err := QuerySSL(ctx, map[string]any{"id": sslIDs})
	if err != nil {
		return err
	}
	if len(ssls) == 0 {
		logging.ErrorFWithContext(ctx, "no ssls found for the specified sslIDs %v", sslIDs)
		return fmt.Errorf("未找到指定的 ssls 资源 IDs %v", sslIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var sslOps []publisher.ResourceOperation
	for _, ssl := range ssls {
		baseInfo := entity.BaseInfo{
			ID:         ssl.ID,
			CreateTime: ssl.CreatedAt.Unix(),
			UpdateTime: ssl.UpdatedAt.Unix(),
		}
		baseConfig, _ := json.Marshal(baseInfo)
		ssl.Config, err = jsonx.MergeJson(ssl.Config, baseConfig)
		if err != nil {
			return err
		}

		// Version-aware field cleanup: ssl never supports name in any version
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.SSL, "name", apisixVersion) {
			ssl.Config, _ = sjson.DeleteBytes(ssl.Config, "name")
		}
		// Remove internal fields that are not part of APISIX schema
		ssl.Config, _ = sjson.DeleteBytes(ssl.Config, "validity_start")
		ssl.Config, _ = sjson.DeleteBytes(ssl.Config, "validity_end")

		sslOps = append(sslOps, publisher.ResourceOperation{
			Key:    ssl.ID,
			Config: json.RawMessage(ssl.Config),
			Type:   constant.SSL,
		})
	}

	// 先创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, sslOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.SSL, sslIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "ssls status change err: %s", err.Error())
		return fmt.Errorf("ssls 发布错误：%w", err)
	}
	return nil
}

// PutStreamRoutes ...
func PutStreamRoutes(ctx context.Context, streamRouteIDs []string) error {
	streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{"id": streamRouteIDs})
	if err != nil {
		return err
	}
	if len(streamRoutes) == 0 {
		logging.ErrorFWithContext(
			ctx,
			"no streamRoutes found for the specified streamRouteIDs %v",
			streamRouteIDs,
		)
		return fmt.Errorf("未找到指定的 streamRoutes 资源 IDs %v", streamRouteIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var upstreamIDs []string
	var serviceIDs []string
	var streamRouteOps []publisher.ResourceOperation
	for _, sr := range streamRoutes {
		baseInfo := entity.BaseInfo{
			ID:         sr.ID,
			CreateTime: sr.CreatedAt.Unix(),
			UpdateTime: sr.UpdatedAt.Unix(),
		}
		if sr.UpstreamID != "" {
			upstreamIDs = append(upstreamIDs, sr.UpstreamID)
		}
		if sr.ServiceID != "" {
			serviceIDs = append(serviceIDs, sr.ServiceID)
		}
		baseConfig, _ := json.Marshal(baseInfo)
		sr.Config, err = jsonx.MergeJson(sr.Config, baseConfig)

		// Version-aware field cleanup: stream_route name is only supported in 3.13+
		if constant.ShouldRemoveFieldBeforeValidationOrPublish(constant.StreamRoute, "name", apisixVersion) {
			sr.Config, _ = sjson.DeleteBytes(sr.Config, "name")
		}
		// Remove internal fields that are not part of APISIX schema
		sr.Config, _ = sjson.DeleteBytes(sr.Config, "labels")

		if err != nil {
			return err
		}
		streamRouteOps = append(streamRouteOps, publisher.ResourceOperation{
			Key:    sr.ID,
			Config: json.RawMessage(sr.Config),
			Type:   constant.StreamRoute,
		})
	}
	// 发布 upstream
	if len(upstreamIDs) > 0 {
		if err := putUpstreams(ctx, upstreamIDs); err != nil {
			return err
		}
	}
	// 发布 service
	if len(serviceIDs) > 0 {
		if err := putServices(ctx, serviceIDs); err != nil {
			return err
		}
	}

	// 创建 etcd 的数据
	err = batchCreateEtcdResource(ctx, streamRouteOps)
	if err != nil {
		return err
	}
	// 变更资源状态为发布成功
	if err = BatchUpdateResourceStatus(
		ctx, constant.StreamRoute, streamRouteIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "streamRoutes status change err: %s", err.Error())
		return fmt.Errorf("streamRoutes 发布错误：%w", err)
	}
	return nil
}
