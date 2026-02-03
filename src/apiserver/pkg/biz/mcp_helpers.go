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
	"reflect"
	"strings"

	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
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

	// Create ResourceCommonModel with config
	resource := model.ResourceCommonModel{
		ID:        resourceID,
		GatewayID: gatewayID,
		Config:    config,
		Status:    status,
		BaseModel: model.BaseModel{
			Updater: "mcp",
		},
	}

	// Convert to specific resource type (extracts association fields from config)
	resourceModel, exists := resourceModelMap[resourceType]
	if !exists {
		return fmt.Errorf("unsupported resource type: %v", resourceType)
	}

	newResourceModel := reflect.New(reflect.TypeOf(resourceModel).Elem()).Interface()
	resourceValue := reflect.ValueOf(resource.ToResourceModel(resourceType))
	if resourceValue.Kind() == reflect.Ptr {
		resourceValue = resourceValue.Elem()
	}
	reflect.ValueOf(newResourceModel).Elem().Set(resourceValue)

	// Set name on the typed model if provided
	if name != "" {
		setResourceName(newResourceModel, resourceType, name)
	}

	// Update with full model struct (includes association fields)
	return database.Client().WithContext(ctx).
		Table(resourceTableMap[resourceType]).
		Where("gateway_id = ? AND id = ?", gatewayID, resourceID).
		Updates(newResourceModel).Error
}

// setResourceName sets the name/username field on a resource model based on resource type
func setResourceName(resource any, resourceType constant.APISIXResource, name string) {
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

// GetPluginsList returns a list of available plugins for the given APISIX version and type
// Uses schema.GetPlugins to read from the version-specific plugin.json files
func GetPluginsList(
	ctx context.Context,
	apisixVersion constant.APISIXVersion,
	apisixType string,
) ([]string, error) {
	// Get plugins from schema based on version and apisix type
	plugins, err := schema.GetPlugins(apisixType, apisixVersion)
	if err != nil {
		return nil, err
	}

	// Filter and collect plugin names
	var pluginNames []string
	for _, plugin := range plugins {
		// Filter by apisixType
		if apisixType == constant.APISIXTypeAPISIX {
			// Standard apisix should not include tapisix or bk-apisix plugins
			if plugin.Type == constant.APISIXTypeTAPISIX || plugin.Type == constant.APISIXTypeBKAPISIX {
				continue
			}
		}
		if apisixType == constant.APISIXTypeTAPISIX {
			// tapisix should not include bk-apisix plugins
			if plugin.Type == constant.APISIXTypeBKAPISIX {
				continue
			}
		}

		pluginNames = append(pluginNames, plugin.Name)
	}

	return pluginNames, nil
}

// ResourceReference represents a resource that references another resource
type ResourceReference struct {
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	ResourceName string `json:"resource_name"`
}

// CheckResourceReferences checks if resources of the given type are referenced by other resources
// Returns a map of resource IDs to lists of resources that reference them
// If a resource is not referenced, it won't appear in the map
func CheckResourceReferences(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceIDs []string,
) (map[string][]ResourceReference, error) {
	if len(resourceIDs) == 0 {
		return nil, nil
	}

	result := make(map[string][]ResourceReference)

	switch resourceType {
	case constant.Service:
		// Services are referenced by routes and stream_routes via service_id
		routes, err := QueryRoutes(ctx, map[string]any{"service_id": resourceIDs})
		if err != nil {
			return nil, err
		}
		for _, route := range routes {
			if route.ServiceID != "" {
				result[route.ServiceID] = append(result[route.ServiceID], ResourceReference{
					ResourceType: constant.Route.String(),
					ResourceID:   route.ID,
					ResourceName: route.Name,
				})
			}
		}

		streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{"service_id": resourceIDs})
		if err != nil {
			return nil, err
		}
		for _, sr := range streamRoutes {
			if sr.ServiceID != "" {
				result[sr.ServiceID] = append(result[sr.ServiceID], ResourceReference{
					ResourceType: constant.StreamRoute.String(),
					ResourceID:   sr.ID,
					ResourceName: sr.Name,
				})
			}
		}

	case constant.Upstream:
		// Upstreams are referenced by services, routes, and stream_routes via upstream_id
		services, err := QueryServices(ctx, map[string]any{"upstream_id": resourceIDs})
		if err != nil {
			return nil, err
		}
		for _, svc := range services {
			if svc.UpstreamID != "" {
				result[svc.UpstreamID] = append(result[svc.UpstreamID], ResourceReference{
					ResourceType: constant.Service.String(),
					ResourceID:   svc.ID,
					ResourceName: svc.Name,
				})
			}
		}

		routes, err := QueryRoutes(ctx, map[string]any{"upstream_id": resourceIDs})
		if err != nil {
			return nil, err
		}
		for _, route := range routes {
			if route.UpstreamID != "" {
				result[route.UpstreamID] = append(result[route.UpstreamID], ResourceReference{
					ResourceType: constant.Route.String(),
					ResourceID:   route.ID,
					ResourceName: route.Name,
				})
			}
		}

		streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{"upstream_id": resourceIDs})
		if err != nil {
			return nil, err
		}
		for _, sr := range streamRoutes {
			if sr.UpstreamID != "" {
				result[sr.UpstreamID] = append(result[sr.UpstreamID], ResourceReference{
					ResourceType: constant.StreamRoute.String(),
					ResourceID:   sr.ID,
					ResourceName: sr.Name,
				})
			}
		}

	case constant.PluginConfig:
		// Plugin configs are referenced by routes via plugin_config_id
		routes, err := QueryRoutes(ctx, map[string]any{"plugin_config_id": resourceIDs})
		if err != nil {
			return nil, err
		}
		for _, route := range routes {
			if route.PluginConfigID != "" {
				result[route.PluginConfigID] = append(result[route.PluginConfigID], ResourceReference{
					ResourceType: constant.Route.String(),
					ResourceID:   route.ID,
					ResourceName: route.Name,
				})
			}
		}

	case constant.ConsumerGroup:
		// Consumer groups are referenced by consumers via group_id
		consumers, err := QueryConsumers(ctx, map[string]any{"group_id": resourceIDs})
		if err != nil {
			return nil, err
		}
		for _, consumer := range consumers {
			if consumer.GroupID != "" {
				result[consumer.GroupID] = append(result[consumer.GroupID], ResourceReference{
					ResourceType: constant.Consumer.String(),
					ResourceID:   consumer.ID,
					ResourceName: consumer.Username,
				})
			}
		}
	}

	return result, nil
}

// FormatResourceReferences formats resource references into a human-readable string
func FormatResourceReferences(refs []ResourceReference) string {
	if len(refs) == 0 {
		return ""
	}

	parts := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.ResourceName != "" {
			parts = append(
				parts,
				fmt.Sprintf("%s '%s' (id: %s)", ref.ResourceType, ref.ResourceName, ref.ResourceID),
			)
		} else {
			parts = append(parts, fmt.Sprintf("%s (id: %s)", ref.ResourceType, ref.ResourceID))
		}
	}
	return strings.Join(parts, ", ")
}
