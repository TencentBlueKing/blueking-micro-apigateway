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
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
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

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

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

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	// Validate pagination params
	if page < 1 {
		page = 1
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Query all synced data for this resource type (no SQL filtering on name/status
	// since those columns don't exist in gateway_sync_data)
	var syncedData []*model.GatewaySyncData
	err = database.Client().WithContext(ctx).
		Where("gateway_id = ? AND type = ?", gateway.ID, resourceType.String()).
		Order("id DESC").
		Find(&syncedData).Error
	if err != nil {
		return errorResult(err), nil
	}

	// Enrich with managed/unmanaged status by checking edit-area tables
	var resourceIDs []string
	for _, sync := range syncedData {
		resourceIDs = append(resourceIDs, sync.ID)
	}

	// Get resources that exist in edit-area
	managedIDSet := make(map[string]bool)
	if len(resourceIDs) > 0 {
		dbResources, err := biz.BatchGetResources(ctx, resourceType, resourceIDs)
		if err == nil {
			for _, dbRes := range dbResources {
				managedIDSet[dbRes.ID] = true
			}
		}
	}

	// Build enriched output with filtering
	type enrichedResource struct {
		ID           string `json:"id"`
		GatewayID    int    `json:"gateway_id"`
		ResourceType string `json:"resource_type"`
		Name         string `json:"name"`
		Status       string `json:"status"` // managed or unmanaged
		Config       any    `json:"config"`
		ModRevision  int    `json:"mod_revision"`
	}

	var filteredResources []enrichedResource
	for _, sync := range syncedData {
		// Extract name from config JSON (username for Consumer, name for others)
		var resourceName string
		if resourceType == constant.Consumer {
			resourceName = gjson.GetBytes(sync.Config, "username").String()
		} else {
			resourceName = gjson.GetBytes(sync.Config, "name").String()
		}

		// Filter by name if specified
		if name != "" && !strings.Contains(resourceName, name) {
			continue
		}

		// Determine managed/unmanaged status
		resourceStatus := "unmanaged"
		if managedIDSet[sync.ID] {
			resourceStatus = "managed"
		}

		// Filter by status if specified
		if status != "" && status != resourceStatus {
			continue
		}

		filteredResources = append(filteredResources, enrichedResource{
			ID:           sync.ID,
			GatewayID:    sync.GatewayID,
			ResourceType: sync.Type.String(),
			Name:         resourceName,
			Status:       resourceStatus,
			Config:       sync.Config,
			ModRevision:  sync.ModRevision,
		})
	}

	// Apply pagination on filtered results
	total := len(filteredResources)
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + pageSize
	if end > total {
		end = total
	}

	pagedResources := filteredResources[offset:end]

	result := map[string]any{
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"resource_type": resourceTypeStr,
		"resources":     pagedResources,
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

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

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
