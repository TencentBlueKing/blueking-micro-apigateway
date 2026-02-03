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

// ============================================================================
// Input Types for Sync Tools
// ============================================================================

// SyncFromEtcdInput is the input for the sync_from_etcd tool
type SyncFromEtcdInput struct {
	ResourceType string `json:"resource_type,omitempty" jsonschema:"optional: only sync specific resource type. If omitted, syncs all."`
}

// ListSyncedResourceInput is the input for the list_synced_resource tool
type ListSyncedResourceInput struct {
	ResourceType string `json:"resource_type" jsonschema:"resource type to list"`
	Name         string `json:"name,omitempty" jsonschema:"filter by resource name (optional, supports fuzzy match)"`
	Status       string `json:"status,omitempty" jsonschema:"filter by sync status: managed (already in edit area) or unmanaged (only in sync area)"`
	Page         int    `json:"page,omitempty" jsonschema:"page number (default: 1)"`
	PageSize     int    `json:"page_size,omitempty" jsonschema:"number of items per page (default: 20, max: 100)"`
}

// AddSyncedResourcesToEditAreaInput is the input for the add_synced_resources_to_edit_area tool
type AddSyncedResourcesToEditAreaInput struct {
	ResourceIDs []string `json:"resource_ids" jsonschema:"array of resource IDs to import (required). Use IDs from list_synced_resource."`
}

// RegisterSyncTools registers all sync-related MCP tools
func RegisterSyncTools(server *mcp.Server) {
	// sync_from_etcd
	mcp.AddTool(server, &mcp.Tool{
		Name: "sync_from_etcd",
		Description: "Synchronize resources from etcd to the sync area (gateway_sync_data). " +
			"Fetches the current state from the APISIX data plane.",
	}, syncFromEtcdHandler)

	// list_synced_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_synced_resource",
		Description: "List resources synced from etcd (from the sync area). " +
			"These are the resources currently deployed in APISIX.",
	}, listSyncedResourceHandler)

	// add_synced_resources_to_edit_area
	mcp.AddTool(server, &mcp.Tool{
		Name: "add_synced_resources_to_edit_area",
		Description: "Import synced resources from the sync area to the edit area. " +
			"Copies resources from gateway_sync_data to their respective tables.",
	}, addSyncedResourcesToEditAreaHandler)
}

// syncFromEtcdHandler handles the sync_from_etcd tool call
func syncFromEtcdHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SyncFromEtcdInput,
) (*mcp.CallToolResult, any, error) {
	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	// Create UnifyOp for sync
	unifyOp, err := biz.NewUnifyOp(gateway, false)
	if err != nil {
		return errorResult(fmt.Errorf("failed to create sync operator: %w", err)), nil, nil
	}

	// Get prefix for sync
	var prefix string
	if input.ResourceType != "" {
		resourceType, err := parseResourceType(input.ResourceType)
		if err != nil {
			return errorResult(err), nil, nil
		}
		prefix = gateway.GetEtcdResourcePrefix(resourceType)
	} else {
		prefix = gateway.GetEtcdPrefixForList()
	}

	// Perform sync
	counts, err := unifyOp.SyncWithPrefix(ctx, prefix)
	if err != nil {
		return errorResult(fmt.Errorf("sync failed: %w", err)), nil, nil
	}

	// Build result
	result := map[string]any{
		"message":       "Sync completed successfully",
		"gateway_id":    gateway.ID,
		"gateway_name":  gateway.Name,
		"synced_counts": counts,
	}

	return successResult(result), nil, nil
}

// listSyncedResourceHandler handles the list_synced_resource tool call
func listSyncedResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListSyncedResourceInput,
) (*mcp.CallToolResult, any, error) {
	resourceType, err := parseResourceType(input.ResourceType)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	// Apply defaults
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
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
		return errorResult(err), nil, nil
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
		if err != nil {
			return errorResult(fmt.Errorf("failed to check managed status: %w", err)), nil, nil
		}
		for _, dbRes := range dbResources {
			managedIDSet[dbRes.ID] = true
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
		if input.Name != "" && !strings.Contains(resourceName, input.Name) {
			continue
		}

		// Determine managed/unmanaged status
		resourceStatus := "unmanaged"
		if managedIDSet[sync.ID] {
			resourceStatus = "managed"
		}

		// Filter by status if specified
		if input.Status != "" && input.Status != resourceStatus {
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
		"resource_type": input.ResourceType,
		"resources":     pagedResources,
	}

	return successResult(result), nil, nil
}

// addSyncedResourcesToEditAreaHandler handles the add_synced_resources_to_edit_area tool call
func addSyncedResourcesToEditAreaHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input AddSyncedResourcesToEditAreaInput,
) (*mcp.CallToolResult, any, error) {
	if len(input.ResourceIDs) == 0 {
		return errorResult(fmt.Errorf("resource_ids is required")), nil, nil
	}

	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	// Get synced resources by IDs
	var syncedResources []*model.GatewaySyncData
	err = database.Client().WithContext(ctx).
		Where("gateway_id = ? AND id IN ?", gateway.ID, input.ResourceIDs).
		Find(&syncedResources).Error
	if err != nil {
		return errorResult(err), nil, nil
	}

	if len(syncedResources) == 0 {
		return errorResult(fmt.Errorf("no synced resources found with the given IDs")), nil, nil
	}

	// Collect all resource IDs
	var allResourceIDs []string
	for _, res := range syncedResources {
		allResourceIDs = append(allResourceIDs, res.ID)
	}

	// Import resources using AddSyncedResources
	stats, err := biz.AddSyncedResources(ctx, allResourceIDs)
	if err != nil {
		return errorResult(fmt.Errorf("failed to add resources to edit area: %w", err)), nil, nil
	}

	addedCount := 0
	for _, count := range stats {
		addedCount += count
	}

	return successResult(map[string]any{
		"message":         "Resources imported to edit area",
		"added_count":     addedCount,
		"total_requested": len(input.ResourceIDs),
	}), nil, nil
}
