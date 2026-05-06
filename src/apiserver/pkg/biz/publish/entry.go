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

package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/auditlog"
	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	unifyopbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/unifyop"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/publisher"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/goroutinex"
)

type publishResourceFunc func(ctx context.Context, resourceIDs []string) error

type publishResourceHandlers struct {
	delete publishResourceFunc
	put    publishResourceFunc
}

var publishResourceHandlerMap = map[constant.APISIXResource]publishResourceHandlers{
	constant.Route: {
		delete: deleteRoutes,
		put:    putRoutes,
	},
	constant.Service: {
		delete: deleteServices,
		put:    putServices,
	},
	constant.Upstream: {
		delete: deleteUpstreams,
		put:    putUpstreams,
	},
	constant.PluginConfig: {
		delete: deletePluginConfigs,
		put:    putPluginConfigs,
	},
	constant.Consumer: {
		delete: deleteConsumers,
		put:    putConsumers,
	},
	constant.ConsumerGroup: {
		delete: deleteConsumerGroups,
		put:    putConsumerGroups,
	},
	constant.GlobalRule: {
		delete: deleteGlobalRules,
		put:    putGlobalRules,
	},
	constant.PluginMetadata: {
		delete: deletePluginMetadatas,
		put:    putPluginMetadatas,
	},
	constant.Proto: {
		delete: deleteProtos,
		put:    putProtos,
	},
	constant.SSL: {
		delete: deleteSSLs,
		put:    putSSLs,
	},
	constant.StreamRoute: {
		delete: deleteStreamRoutes,
		put:    putStreamRoutes,
	},
}

// PublishResource 资源发布
func PublishResource(ctx context.Context, resourceType constant.APISIXResource, resourceIDs []string) error {
	handlers, ok := publishResourceHandlerMap[resourceType]
	if !ok {
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	err := WrapPublishResource(ctx, resourceType, resourceIDs, handlers)
	if err != nil {
		return err
	}
	// 主动同步一下资源
	goroutinex.GoroutineWithRecovery(ctx, func() {
		// 1s 后同步资源
		time.Sleep(time.Second * 1)
		_, err = unifyopbiz.SyncResources(ginx.CloneCtx(ctx), resourceType)
		if err != nil {
			logging.Errorf("sync resources failed, err: %v", err)
		}
	})
	return nil
}

// WrapPublishResource PublishResource 资源发布进行一些公共操作
func WrapPublishResource(ctx context.Context, resourceType constant.APISIXResource, resourceIDs []string,
	handlers publishResourceHandlers,
) error {
	// 状态机判断
	resourceList, err := resourcebiz.BatchGetResources(ctx, resourceType, resourceIDs)
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
	err = publishResourcesWithHandlers(ctx, resourceList, handlers)
	if err != nil {
		return err
	}
	err = auditlog.AddBatchAuditLog(
		ctx,
		constant.OperationTypePublish,
		resourceType,
		resourceList,
		resourceStatusMap,
	)
	if err != nil {
		logging.ErrorFWithContext(ctx, "%s add audit log err: %s", resourceType, err.Error())
		return err
	}
	return nil
}

func publishResourcesWithHandlers(
	ctx context.Context,
	resourceList []*model.ResourceCommonModel,
	handlers publishResourceHandlers,
) error {
	var deleteIDs []string
	var putIDs []string
	for _, resource := range resourceList {
		if resource.Status == constant.ResourceStatusDeleteDraft {
			deleteIDs = append(deleteIDs, resource.ID)
			continue
		}
		putIDs = append(putIDs, resource.ID)
	}
	if len(deleteIDs) > 0 {
		if err := handlers.delete(ctx, deleteIDs); err != nil {
			return err
		}
	}
	if len(putIDs) > 0 {
		if err := handlers.put(ctx, putIDs); err != nil {
			return err
		}
	}
	return nil
}

// PublishAllResource 资源一键发布
func PublishAllResource(ctx context.Context, gatewayID int) error {
	for _, resourceType := range constant.ResourceTypeList {
		resources, err := resourcebiz.QueryResource(ctx, resourceType,
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

// FormatResourceIDNameList 格式化资源 ID 和名称列表
func FormatResourceIDNameList(resources any, resourceType constant.APISIXResource) []string {
	switch resourceType {
	case constant.Route:
		routes := resources.([]*model.Route) //nolint:forcetypeassert
		routeDetails := make([]string, 0, len(routes))
		for _, route := range routes {
			routeDetails = append(
				routeDetails,
				fmt.Sprintf("%s(%s)", route.ID, route.GetName(resourceType)),
			)
		}
		return routeDetails
	case constant.Upstream:
		upstreams := resources.([]*model.Upstream) //nolint:forcetypeassert
		upstreamDetails := make([]string, 0, len(upstreams))
		for _, upstream := range upstreams {
			upstreamDetails = append(
				upstreamDetails,
				fmt.Sprintf("%s(%s)", upstream.ID, upstream.GetName(resourceType)),
			)
		}
		return upstreamDetails
	case constant.Consumer:
		consumers := resources.([]*model.Consumer) //nolint:forcetypeassert
		consumerDetails := make([]string, 0, len(consumers))
		for _, consumer := range consumers {
			consumerDetails = append(
				consumerDetails,
				fmt.Sprintf("%s(%s)", consumer.ID, consumer.GetName(resourceType)),
			)
		}
		return consumerDetails
	case constant.Service:
		services := resources.([]*model.Service) //nolint:forcetypeassert
		serviceDetails := make([]string, 0, len(services))
		for _, service := range services {
			serviceDetails = append(
				serviceDetails,
				fmt.Sprintf("%s(%s)", service.ID, service.GetName(resourceType)),
			)
		}
		return serviceDetails
	case constant.StreamRoute:
		streamRoutes := resources.([]*model.StreamRoute) //nolint:forcetypeassert
		streamRouteDetails := make([]string, 0, len(streamRoutes))
		for _, streamRoute := range streamRoutes {
			streamRouteDetails = append(
				streamRouteDetails,
				fmt.Sprintf("%s(%s)", streamRoute.ID, streamRoute.GetName(resourceType)),
			)
		}
		return streamRouteDetails
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
	return resourcebiz.BatchDeleteRoutes(ctx, routeIDs)
}

// deleteServices 删除 service
func deleteServices(ctx context.Context, serviceIDs []string) error {
	// 先判断 service 有没有关联的资源数据
	routes, err := resourcebiz.QueryRoutes(ctx, map[string]any{"service_id": serviceIDs})
	if err != nil {
		return err
	}
	if len(routes) > 0 {
		return fmt.Errorf("服务不可删除, 存在关联的路由资源 %v", FormatResourceIDNameList(routes, constant.Route))
	}
	// 判断 service 有没有关联的 streamRoute 数据
	streamRoutes, err := resourcebiz.QueryStreamRoutes(ctx, map[string]any{"service_id": serviceIDs})
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
	return resourcebiz.BatchDeleteServices(ctx, serviceIDs)
}

// deleteUpstreams 删除 upstream
func deleteUpstreams(ctx context.Context, upstreamIDs []string) error {
	// 判断 upstream 有没有关联的 service 数据
	services, err := resourcebiz.QueryServices(ctx, map[string]any{"upstream_id": upstreamIDs})
	if err != nil {
		return err
	}
	if len(services) > 0 {
		return fmt.Errorf("上游不可删除, 存在关联的服务资源 %v", FormatResourceIDNameList(services, constant.Service))
	}
	// 判断 upstream 有没有关联的 route 数据
	routes, err := resourcebiz.QueryRoutes(ctx, map[string]any{"upstream_id": upstreamIDs})
	if err != nil {
		return err
	}
	if len(routes) > 0 {
		return fmt.Errorf("上游不可删除, 存在关联的路由资源 %v", FormatResourceIDNameList(routes, constant.Route))
	}
	// 判断 upstream 有没有关联的 streamRoute 数据
	streamRoutes, err := resourcebiz.QueryStreamRoutes(ctx, map[string]any{"upstream_id": upstreamIDs})
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
	return resourcebiz.BatchDeleteUpstreams(ctx, upstreamIDs)
}

// deletePluginConfigs 删除 pluginConfig
func deletePluginConfigs(ctx context.Context, pluginConfigIDs []string) error {
	// 判断 plugin_config 有没有关联的 route 数据
	routes, err := resourcebiz.QueryRoutes(ctx, map[string]any{"plugin_config_id": pluginConfigIDs})
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
	return resourcebiz.BatchDeletePluginConfigs(ctx, pluginConfigIDs)
}

// deletePluginMetadatas 删除 pluginMetadata
func deletePluginMetadatas(ctx context.Context, pluginMetadataIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.PluginMetadata, pluginMetadataIDs)
	if err != nil {
		return err
	}

	// 删除数据库数据
	return resourcebiz.BatchDeletePluginMetadatas(ctx, pluginMetadataIDs)
}

// deleteConsumers 删除 consumer
func deleteConsumers(ctx context.Context, consumerIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.Consumer, consumerIDs)
	if err != nil {
		return err
	}

	// 删除数据库数据
	return resourcebiz.BatchDeleteConsumers(ctx, consumerIDs)
}

// deleteConsumerGroups 删除 consumerGroup
func deleteConsumerGroups(ctx context.Context, consumerGroupIDs []string) error {
	consumers, err := resourcebiz.QueryConsumers(ctx, map[string]any{"group_id": consumerGroupIDs})
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

	return resourcebiz.BatchDeleteConsumerGroups(ctx, consumerGroupIDs)
}

// deleteGlobalRules 删除 globalRule
func deleteGlobalRules(ctx context.Context, globalRuleIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.GlobalRule, globalRuleIDs)
	if err != nil {
		return err
	}
	// 删除数据库数据
	return resourcebiz.BatchDeleteGlobalRules(ctx, globalRuleIDs)
}

// deleteProtos 删除 Proto
func deleteProtos(ctx context.Context, protoIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.Proto, protoIDs)
	if err != nil {
		return err
	}
	return resourcebiz.BatchDeleteProtos(ctx, protoIDs)
}

// deleteSSLs 删除 SSL
func deleteSSLs(ctx context.Context, sslIDs []string) error {
	upstreams, err := resourcebiz.QueryUpstreams(ctx, map[string]any{"ssl_id": sslIDs})
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
	return resourcebiz.BatchDeleteSSL(ctx, sslIDs)
}

// deleteStreamRoutes 删除 StreamRoute
func deleteStreamRoutes(ctx context.Context, streamRouteIDs []string) error {
	// 先删除 etcd 的数据
	err := batchDeleteEtcdResource(ctx, constant.StreamRoute, streamRouteIDs)
	if err != nil {
		return err
	}
	return resourcebiz.BatchDeleteStreamRoutes(ctx, streamRouteIDs)
}

// putRoutes 发布路由
func putRoutes(ctx context.Context, routeIDs []string) error {
	routes, err := resourcebiz.QueryRoutes(ctx, map[string]any{"id": routeIDs})
	if err != nil {
		return err
	}
	if len(routes) == 0 {
		logging.ErrorFWithContext(ctx, "no routes found for the specified routeIDs %v", routeIDs)
		return fmt.Errorf("未找到指定的路由资源 IDs %v", routeIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	deps := collectRoutePublishDependencies(routes)
	var routeOps []publisher.ResourceOperation
	for _, route := range routes {
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.Route,
			ResourceKey:  route.ID,
			BaseInfo: entity.BaseInfo{
				ID:         route.ID,
				Name:       route.Name,
				CreateTime: route.CreatedAt.Unix(),
				UpdateTime: route.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(route.Config),
		})
		if err != nil {
			return err
		}
		routeOps = append(routeOps, op)
	}
	// 发布 upstream
	if len(deps.UpstreamIDs) > 0 {
		if err := putUpstreams(ctx, deps.UpstreamIDs); err != nil {
			return err
		}
	}

	// 发布 service
	if len(deps.ServiceIDs) > 0 {
		if err := putServices(ctx, deps.ServiceIDs); err != nil {
			return err
		}
	}

	// 发布 pluginConfig
	if len(deps.PluginConfigIDs) > 0 {
		if err := putPluginConfigs(ctx, deps.PluginConfigIDs); err != nil {
			return err
		}
	}
	return persistPublishedOperations(ctx, constant.Route, routeIDs, routeOps, "路由发布错误")
}

func putServices(ctx context.Context, serviceIDs []string) error {
	services, err := resourcebiz.QueryServices(ctx, map[string]any{"id": serviceIDs})
	if err != nil {
		return err
	}
	if len(services) == 0 {
		logging.ErrorFWithContext(ctx, "no services found for the specified serviceIDs %v", serviceIDs)
		return fmt.Errorf("未找到指定的服务资源 IDs %v", serviceIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	deps := collectServicePublishDependencies(services)
	var serviceOps []publisher.ResourceOperation
	for _, service := range services {
		if service.UpstreamID == "" {
			serviceOps = append(serviceOps, publisher.ResourceOperation{
				Key:    service.ID,
				Config: json.RawMessage(service.Config),
				Type:   constant.Service,
			})
			continue
		}

		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.Service,
			ResourceKey:  service.ID,
			BaseInfo: entity.BaseInfo{
				ID:         service.ID,
				CreateTime: service.CreatedAt.Unix(),
				UpdateTime: service.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(service.Config),
		})
		if err != nil {
			return err
		}
		serviceOps = append(serviceOps, op)
	}
	// 发布 upstream
	if len(deps.UpstreamIDs) > 0 {
		if err = putUpstreams(ctx, deps.UpstreamIDs); err != nil {
			return err
		}
	}
	return persistPublishedOperations(ctx, constant.Service, serviceIDs, serviceOps, "服务发布错误")
}

func putUpstreams(ctx context.Context, upstreamIDs []string) error {
	upstreams, err := resourcebiz.QueryUpstreams(ctx, map[string]any{"id": upstreamIDs})
	if err != nil {
		return err
	}
	if len(upstreams) == 0 {
		logging.ErrorFWithContext(ctx, "no upstreams found for the specified upstreamIDs %v", upstreamIDs)
		return fmt.Errorf("未找到指定的上游资源 IDs %v", upstreamIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	deps := collectUpstreamPublishDependencies(upstreams)
	var upstreamOps []publisher.ResourceOperation
	for _, upstream := range upstreams {
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.Upstream,
			ResourceKey:  upstream.ID,
			BaseInfo: entity.BaseInfo{
				ID:         upstream.ID,
				CreateTime: upstream.CreatedAt.Unix(),
				UpdateTime: upstream.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(upstream.Config),
		})
		if err != nil {
			return err
		}
		upstreamOps = append(upstreamOps, op)
	}
	if len(deps.SSLIDs) > 0 {
		if err = putSSLs(ctx, deps.SSLIDs); err != nil {
			return err
		}
	}
	return persistPublishedOperations(ctx, constant.Upstream, upstreamIDs, upstreamOps, "上游发布错误")
}

// putPluginConfigs ...
func putPluginConfigs(ctx context.Context, pluginConfigIDs []string) error {
	pluginConfigs, err := resourcebiz.QueryPluginConfigs(ctx, map[string]any{"id": pluginConfigIDs})
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
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.PluginConfig,
			ResourceKey:  pluginConfig.ID,
			BaseInfo: entity.BaseInfo{
				ID:         pluginConfig.ID,
				CreateTime: pluginConfig.CreatedAt.Unix(),
				UpdateTime: pluginConfig.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(pluginConfig.Config),
		})
		if err != nil {
			return err
		}
		pluginConfigOps = append(pluginConfigOps, op)
	}
	return persistPublishedOperations(ctx, constant.PluginConfig, pluginConfigIDs, pluginConfigOps, "插件组发布错误")
}

// putPluginMetadatas ...
func putPluginMetadatas(ctx context.Context, pluginMetadataIDs []string) error {
	pluginMetadatas, err := resourcebiz.QueryPluginMetadatas(ctx, map[string]any{"id": pluginMetadataIDs})
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

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	var pluginMetadataOps []publisher.ResourceOperation
	for _, pluginMetadata := range pluginMetadatas {
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.PluginMetadata,
			ResourceKey:  pluginMetadata.Name,
			BaseInfo: entity.BaseInfo{
				ID:         pluginMetadata.Name, // pluginMetadata.Name 必须是 pluginName
				CreateTime: pluginMetadata.CreatedAt.Unix(),
				UpdateTime: pluginMetadata.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(pluginMetadata.Config),
		})
		if err != nil {
			return err
		}
		pluginMetadataOps = append(pluginMetadataOps, op)
	}
	return persistPublishedOperations(
		ctx,
		constant.PluginMetadata,
		pluginMetadataIDs,
		pluginMetadataOps,
		"插件元数据发布错误",
	)
}

// putConsumers ...
func putConsumers(ctx context.Context, consumerIDs []string) error {
	consumers, err := resourcebiz.QueryConsumers(ctx, map[string]any{"id": consumerIDs})
	if err != nil {
		return err
	}
	if len(consumers) == 0 {
		logging.ErrorFWithContext(ctx, "no consumers found for the specified consumerIDs %v", consumerIDs)
		return fmt.Errorf("未找到指定的消费者资源 IDs %v", consumerIDs)
	}

	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	apisixVersion := gatewayInfo.GetAPISIXVersionX()

	deps := collectConsumerPublishDependencies(consumers)
	var consumerOps []publisher.ResourceOperation
	for _, consumer := range consumers {
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.Consumer,
			ResourceKey:  consumer.ID,
			BaseInfo: entity.BaseInfo{
				CreateTime: consumer.CreatedAt.Unix(),
				UpdateTime: consumer.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(consumer.Config),
		})
		if err != nil {
			return err
		}
		consumerOps = append(consumerOps, op)
	}

	if len(deps.ConsumerGroupIDs) > 0 {
		err = putConsumerGroups(ctx, deps.ConsumerGroupIDs)
		if err != nil {
			return err
		}
	}
	return persistPublishedOperations(ctx, constant.Consumer, consumerIDs, consumerOps, "消费者发布错误")
}

// putConsumerGroups ...
func putConsumerGroups(ctx context.Context, consumerGroupIDs []string) error {
	consumerGroups, err := resourcebiz.QueryConsumerGroups(ctx, map[string]any{"id": consumerGroupIDs})
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
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.ConsumerGroup,
			ResourceKey:  consumerGroup.ID,
			BaseInfo: entity.BaseInfo{
				ID:         consumerGroup.ID,
				CreateTime: consumerGroup.CreatedAt.Unix(),
				UpdateTime: consumerGroup.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(consumerGroup.Config),
		})
		if err != nil {
			return err
		}
		consumerGroupOps = append(consumerGroupOps, op)
	}
	return persistPublishedOperations(
		ctx,
		constant.ConsumerGroup,
		consumerGroupIDs,
		consumerGroupOps,
		"消费者组发布错误",
	)
}

// putGlobalRules ...
func putGlobalRules(ctx context.Context, globalRuleIDs []string) error {
	globalRules, err := resourcebiz.QueryGlobalRules(ctx, map[string]any{"id": globalRuleIDs})
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
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.GlobalRule,
			ResourceKey:  globalRule.ID,
			BaseInfo: entity.BaseInfo{
				ID:         globalRule.ID,
				CreateTime: globalRule.CreatedAt.Unix(),
				UpdateTime: globalRule.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(globalRule.Config),
		})
		if err != nil {
			return err
		}
		globalRuleOps = append(globalRuleOps, op)
	}
	return persistPublishedOperations(ctx, constant.GlobalRule, globalRuleIDs, globalRuleOps, "全局规则发布错误")
}

// putProtos ...
func putProtos(ctx context.Context, protoIDs []string) error {
	protos, err := resourcebiz.QueryProtos(ctx, map[string]any{"id": protoIDs})
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
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.Proto,
			ResourceKey:  pb.ID,
			BaseInfo: entity.BaseInfo{
				ID:         pb.ID,
				CreateTime: pb.CreatedAt.Unix(),
				UpdateTime: pb.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(pb.Config),
		})
		if err != nil {
			return err
		}
		protoOps = append(protoOps, op)
	}
	return persistPublishedOperations(ctx, constant.Proto, protoIDs, protoOps, "protos 发布错误")
}

// putSSLs ...
func putSSLs(ctx context.Context, sslIDs []string) error {
	ssls, err := resourcebiz.QuerySSL(ctx, map[string]any{"id": sslIDs})
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
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.SSL,
			ResourceKey:  ssl.ID,
			BaseInfo: entity.BaseInfo{
				ID:         ssl.ID,
				CreateTime: ssl.CreatedAt.Unix(),
				UpdateTime: ssl.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(ssl.Config),
		})
		if err != nil {
			return err
		}
		sslOps = append(sslOps, op)
	}
	return persistPublishedOperations(ctx, constant.SSL, sslIDs, sslOps, "ssls 发布错误")
}

// putStreamRoutes ...
func putStreamRoutes(ctx context.Context, streamRouteIDs []string) error {
	streamRoutes, err := resourcebiz.QueryStreamRoutes(ctx, map[string]any{"id": streamRouteIDs})
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

	deps := collectStreamRoutePublishDependencies(streamRoutes)
	var streamRouteOps []publisher.ResourceOperation
	for _, sr := range streamRoutes {
		op, err := buildPublishResourceOperation(publishResourceOperationInput{
			ResourceType: constant.StreamRoute,
			ResourceKey:  sr.ID,
			BaseInfo: entity.BaseInfo{
				ID:         sr.ID,
				CreateTime: sr.CreatedAt.Unix(),
				UpdateTime: sr.UpdatedAt.Unix(),
			},
			Version:   apisixVersion,
			RawConfig: json.RawMessage(sr.Config),
		})
		if err != nil {
			return err
		}
		streamRouteOps = append(streamRouteOps, op)
	}
	// 发布 upstream
	if len(deps.UpstreamIDs) > 0 {
		if err := putUpstreams(ctx, deps.UpstreamIDs); err != nil {
			return err
		}
	}
	// 发布 service
	if len(deps.ServiceIDs) > 0 {
		if err := putServices(ctx, deps.ServiceIDs); err != nil {
			return err
		}
	}
	return persistPublishedOperations(
		ctx,
		constant.StreamRoute,
		streamRouteIDs,
		streamRouteOps,
		"streamRoutes 发布错误",
	)
}
