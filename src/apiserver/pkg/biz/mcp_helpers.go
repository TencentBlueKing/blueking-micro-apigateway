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

	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ListResourcesWithPagination lists resources with pagination support
func ListResourcesWithPagination(
	ctx context.Context,
	resourceType constant.APISIXResource,
	name string,
	statuses []string,
	offset int,
	limit int,
) ([]any, int64, error) {
	var results []map[string]any
	var total int64

	query := buildCommonDbQuery(ctx, resourceType)

	// Apply name filter
	if name != "" {
		nameKey := model.GetResourceNameKey(resourceType)
		query = query.Where(nameKey+" LIKE ?", "%"+name+"%")
	}

	// Apply status filter
	validStatuses := []string{}
	for _, s := range statuses {
		if s != "" {
			validStatuses = append(validStatuses, s)
		}
	}
	if len(validStatuses) > 0 {
		query = query.Where("status IN ?", validStatuses)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&results).Error; err != nil {
		return nil, 0, err
	}

	// Convert to []any
	anyResults := make([]any, len(results))
	for i, r := range results {
		anyResults[i] = r
	}

	return anyResults, total, nil
}

// CreateResource creates a new resource
func CreateResource(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resource any,
	name string,
) error {
	// Set the name on the resource based on resource type
	// The name field varies by type (e.g., "name" for most, "username" for consumer)
	switch r := resource.(type) {
	case *model.Route:
		r.Name = name
	case *model.Service:
		r.Name = name
	case *model.Upstream:
		r.Name = name
	case *model.Consumer:
		r.Username = name
	case *model.ConsumerGroup:
		r.Name = name
	case *model.PluginConfig:
		r.Name = name
	case *model.GlobalRule:
		r.Name = name
	case *model.PluginMetadata:
		r.Name = name
	case *model.Proto:
		r.Name = name
	case *model.SSL:
		r.Name = name
	case *model.StreamRoute:
		r.Name = name
	}

	return database.Client().WithContext(ctx).Create(resource).Error
}

// RevertResource reverts a resource to its synced snapshot state
func RevertResource(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceID string,
) error {
	gatewayID := ginx.GetGatewayInfoFromContext(ctx).ID

	// Get synced data
	var syncedData model.GatewaySyncData
	err := database.Client().WithContext(ctx).
		Where("gateway_id = ? AND type = ? AND id = ?", gatewayID, resourceType.String(), resourceID).
		First(&syncedData).Error
	if err != nil {
		return fmt.Errorf("synced data not found: %w", err)
	}

	// Update resource with synced config
	err = database.Client().WithContext(ctx).
		Table(resourceTableMap[resourceType]).
		Where("gateway_id = ? AND id = ?", gatewayID, resourceID).
		Updates(map[string]any{
			"config": syncedData.Config,
			"status": constant.ResourceStatusSuccess,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to revert resource: %w", err)
	}

	return nil
}

// UpdateResourceByTypeAndID updates a resource by type and ID
func UpdateResourceByTypeAndID(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceID string,
	name string,
	config datatypes.JSON,
	status constant.ResourceStatus,
) error {
	gatewayID := ginx.GetGatewayInfoFromContext(ctx).ID

	updates := map[string]any{
		"config": config,
		"status": status,
	}

	if name != "" {
		nameKey := model.GetResourceNameKey(resourceType)
		updates[nameKey] = name
	}

	return database.Client().WithContext(ctx).
		Table(resourceTableMap[resourceType]).
		Where("gateway_id = ? AND id = ?", gatewayID, resourceID).
		Updates(updates).Error
}

// PublishResourcesByType publishes resources to etcd using the existing PublishResource function
func PublishResourcesByType(
	ctx context.Context,
	gateway *model.Gateway,
	resourceType constant.APISIXResource,
	resourceIDs []string,
) error {
	if len(resourceIDs) == 0 {
		return fmt.Errorf("no resources to publish")
	}

	// Set gateway info in context
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	// Use existing PublishResource function
	return PublishResource(ctx, resourceType, resourceIDs)
}

// GetPluginsList returns a list of available plugins
func GetPluginsList(
	ctx context.Context,
	apisixVersion constant.APISIXVersion,
	apisixType string,
	pluginType string,
) ([]string, error) {
	// Get plugins from schema
	plugins, err := GetAvailablePlugins(apisixVersion, apisixType)
	if err != nil {
		return nil, err
	}

	// Filter by type if specified
	if pluginType != "" {
		filteredPlugins := []string{}
		for _, plugin := range plugins {
			// For now, return all plugins since we don't have type info
			// In a full implementation, this would filter by http/stream/metadata
			filteredPlugins = append(filteredPlugins, plugin)
		}
		return filteredPlugins, nil
	}

	return plugins, nil
}

// GetAvailablePlugins returns available plugins for a version and type
func GetAvailablePlugins(apisixVersion constant.APISIXVersion, apisixType string) ([]string, error) {
	// This is a simplified implementation
	// In production, this would read from the schema files
	commonPlugins := []string{
		"limit-req",
		"limit-count",
		"limit-conn",
		"proxy-rewrite",
		"response-rewrite",
		"redirect",
		"cors",
		"ip-restriction",
		"ua-restriction",
		"referer-restriction",
		"consumer-restriction",
		"key-auth",
		"jwt-auth",
		"basic-auth",
		"hmac-auth",
		"authz-keycloak",
		"authz-casdoor",
		"openid-connect",
		"prometheus",
		"zipkin",
		"skywalking",
		"http-logger",
		"file-logger",
		"syslog",
		"kafka-logger",
		"rocketmq-logger",
		"tcp-logger",
		"udp-logger",
		"clickhouse-logger",
		"sls-logger",
		"elasticsearch-logger",
		"request-id",
		"request-validation",
		"fault-injection",
		"traffic-split",
		"echo",
		"gzip",
		"real-ip",
		"serverless-pre-function",
		"serverless-post-function",
		"ext-plugin-pre-req",
		"ext-plugin-post-req",
		"ext-plugin-post-resp",
	}

	// Add bk-apisix specific plugins
	if apisixType == "bk-apisix" {
		commonPlugins = append(commonPlugins,
			"bk-auth-verify",
			"bk-permission",
			"bk-rate-limit",
			"bk-ip-restriction",
			"bk-user-restriction",
			"bk-header-rewrite",
			"bk-cors",
			"bk-mock",
			"bk-error-info",
			"bk-request-id",
			"bk-log",
		)
	}

	return commonPlugins, nil
}
