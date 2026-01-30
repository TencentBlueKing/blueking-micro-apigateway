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

package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
)

// RegisterSyncTools registers all sync-related MCP tools
func RegisterSyncTools(server *mcp.Server) {
	// sync_from_etcd
	server.AddTool(&mcp.Tool{
		Name: "sync_from_etcd",
		Description: "Synchronize resources from etcd to the sync area (gateway_sync_data). " +
			"Fetches the current state from the APISIX data plane.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"resource_type": map[string]any{
					"type": "string",
					"description": "Optional: Only sync specific resource type. " +
						"If omitted, syncs all. " + ResourceTypeDescription(),
					"enum": ValidResourceTypes,
				},
			},
		},
	}, syncFromEtcdHandler)

	// list_synced_resource
	server.AddTool(&mcp.Tool{
		Name: "list_synced_resource",
		Description: "List resources synced from etcd (from the sync area). " +
			"These are the resources currently deployed in APISIX.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type to list. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Filter by resource name (optional, supports fuzzy match)",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Filter by sync status: managed (already in edit area) or unmanaged (only in sync area)",
					"enum":        []string{"managed", "unmanaged"},
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Page number (default: 1)",
					"default":     1,
				},
				"page_size": map[string]any{
					"type":        "integer",
					"description": "Number of items per page (default: 20, max: 100)",
					"default":     20,
					"maximum":     100,
				},
			},
			"required": []string{"resource_type"},
		},
	}, listSyncedResourceHandler)

	// add_synced_resources_to_edit_area
	server.AddTool(&mcp.Tool{
		Name: "add_synced_resources_to_edit_area",
		Description: "Import synced resources from the sync area to the edit area. " +
			"Copies resources from gateway_sync_data to their respective tables.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"resource_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Array of resource IDs to import (required). Use IDs from list_synced_resource.",
				},
			},
			"required": []string{"resource_ids"},
		},
	}, addSyncedResourcesToEditAreaHandler)
}

// syncFromEtcdHandler handles the sync_from_etcd tool call
func syncFromEtcdHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")

	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	// Create UnifyOp for sync
	unifyOp, err := biz.NewUnifyOp(gateway, false)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create sync operator: %w", err)), nil
	}

	// Get prefix for sync
	var prefix string
	if resourceTypeStr != "" {
		resourceType, err := parseResourceType(resourceTypeStr)
		if err != nil {
			return errorResult(err), nil
		}
		prefix = gateway.GetEtcdResourcePrefix(resourceType)
	} else {
		prefix = gateway.GetEtcdPrefixForList()
	}

	// Perform sync
	counts, err := unifyOp.SyncWithPrefix(ctx, prefix)
	if err != nil {
		return errorResult(fmt.Errorf("sync failed: %w", err)), nil
	}

	// Build result
	result := map[string]any{
		"message":       "Sync completed successfully",
		"gateway_id":    gateway.ID,
		"gateway_name":  gateway.Name,
		"synced_counts": counts,
	}

	return successResult(result), nil
}

// listSyncedResourceHandler handles the list_synced_resource tool call
func listSyncedResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	name := getStringParamFromArgs(args, "name", "")
	status := getStringParamFromArgs(args, "status", "")
	page := getIntParamFromArgs(args, "page", 1)
	pageSize := getIntParamFromArgs(args, "page_size", 20)

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	// Calculate offset
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Query synced data
	var syncedData []*model.GatewaySyncData
	query := database.Client().WithContext(ctx).
		Where("gateway_id = ? AND type = ?", gateway.ID, resourceType.String())

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Filter by managed/unmanaged status
	if status == "managed" {
		query = query.Where("sync_status = ?", constant.SyncedResourceStatusSuccess)
	} else if status == "unmanaged" {
		query = query.Where("sync_status = ? OR sync_status IS NULL", constant.SyncedResourceStatusMiss)
	}

	var total int64
	err = query.Model(&model.GatewaySyncData{}).Count(&total).Error
	if err != nil {
		return errorResult(err), nil
	}

	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&syncedData).Error
	if err != nil {
		return errorResult(err), nil
	}

	result := map[string]any{
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"resource_type": resourceTypeStr,
		"resources":     syncedData,
	}

	return successResult(result), nil
}

// addSyncedResourcesToEditAreaHandler handles the add_synced_resources_to_edit_area tool call
func addSyncedResourcesToEditAreaHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check write scope
	if err := CheckWriteScope(ctx); err != nil {
		return errorResult(err), nil
	}

	args := parseArguments(req)
	resourceIDs := getStringArrayParamFromArgs(args, "resource_ids")

	if len(resourceIDs) == 0 {
		return errorResult(fmt.Errorf("resource_ids is required")), nil
	}

	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil
	}

	// Get synced resources by IDs
	var syncedResources []*model.GatewaySyncData
	err = database.Client().WithContext(ctx).
		Where("gateway_id = ? AND id IN ?", gateway.ID, resourceIDs).
		Find(&syncedResources).Error
	if err != nil {
		return errorResult(err), nil
	}

	if len(syncedResources) == 0 {
		return errorResult(fmt.Errorf("no synced resources found with the given IDs")), nil
	}

	// Collect all resource IDs
	var allResourceIDs []string
	for _, res := range syncedResources {
		allResourceIDs = append(allResourceIDs, res.ID)
	}

	// Import resources using AddSyncedResources
	stats, err := biz.AddSyncedResources(ctx, allResourceIDs)
	if err != nil {
		return errorResult(fmt.Errorf("failed to add resources to edit area: %w", err)), nil
	}

	addedCount := 0
	for _, count := range stats {
		addedCount += count
	}

	return successResult(map[string]any{
		"message":         "Resources imported to edit area",
		"added_count":     addedCount,
		"total_requested": len(resourceIDs),
	}), nil
}
