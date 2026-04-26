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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// ============================================================================
// Input Types for Resource CRUD Tools
// ============================================================================

// ListResourceInput is the input for the list_resource tool
type ListResourceInput struct {
	ResourceType string `json:"resource_type" jsonschema:"Required. APISIX resource type to list."`
	Name         string `json:"name,omitempty" jsonschema:"Optional name substring filter (consumer uses username)."`
	//nolint:lll // Keep allowed status values explicit for MCP clients.
	Status   string `json:"status,omitempty" jsonschema:"Optional status filter: create_draft|update_draft|delete_draft|success."`
	Page     int    `json:"page,omitempty" jsonschema:"Optional page number. Default: 1."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Optional page size. Default: 20. Max: 100."`
}

// GetResourceInput is the input for the get_resource tool
type GetResourceInput struct {
	ResourceType string `json:"resource_type" jsonschema:"Required. APISIX resource type used to resolve resource_id."`
	ResourceID   string `json:"resource_id" jsonschema:"Required. Resource ID to retrieve."`
}

// CreateResourceInput is the input for the create_resource tool
type CreateResourceInput struct {
	ResourceType string         `json:"resource_type" jsonschema:"Required. APISIX resource type to create."`
	Name         string         `json:"name" jsonschema:"Required. Resource name (consumer maps this to username)."`
	Config       map[string]any `json:"config" jsonschema:"Required. Full resource config object following APISIX schema."`
}

// UpdateResourceInput is the input for the update_resource tool
type UpdateResourceInput struct {
	ResourceType string `json:"resource_type" jsonschema:"Required. APISIX resource type."`
	ResourceID   string `json:"resource_id" jsonschema:"Required. Resource ID to update."`
	Name         string `json:"name,omitempty" jsonschema:"Optional new resource name."`
	//nolint:lll // Clarify that patch-style updates are not supported.
	Config map[string]any `json:"config" jsonschema:"Required full config replacement; partial patch updates are not supported."`
}

// DeleteResourceInput is the input for the delete_resource tool
type DeleteResourceInput struct {
	ResourceType string   `json:"resource_type" jsonschema:"Required. APISIX resource type."`
	ResourceIDs  []string `json:"resource_ids" jsonschema:"Required. Resource IDs to delete."`
}

// RevertResourceInput is the input for the revert_resource tool
type RevertResourceInput struct {
	ResourceType string   `json:"resource_type" jsonschema:"Required. APISIX resource type."`
	ResourceIDs  []string `json:"resource_ids" jsonschema:"Required. Resource IDs to revert."`
}

type batchOperationFailure struct {
	ResourceID string `json:"resource_id"`
	Stage      string `json:"stage"`
	Error      string `json:"error"`
}

func buildBatchOperationResult(
	message string,
	resourceType string,
	totalRequested int,
	successCount int,
	extraCounts map[string]int,
	failedItems []batchOperationFailure,
) map[string]any {
	if failedItems == nil {
		failedItems = []batchOperationFailure{}
	}
	result := map[string]any{
		"message":         message,
		"total_requested": totalRequested,
		"success_count":   successCount,
		"failed_count":    len(failedItems),
		"partial_success": len(failedItems) > 0,
		"failed_items":    failedItems,
		"resource_type":   resourceType,
	}
	for key, value := range extraCounts {
		result[key] = value
	}
	return result
}

// RegisterResourceCRUDTools registers all resource CRUD tools
func RegisterResourceCRUDTools(server *mcp.Server) {
	// list_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_resource",
		Description: "List resources from the edit area with pagination and optional name/status filters. " +
			"Read-only operation.",
	}, listResourceHandler)

	// get_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_resource",
		Description: "Get a single resource by resource_type and resource_id, including full config. " +
			"Use this before update_resource to build a complete replacement config.",
	}, getResourceHandler)

	// create_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "create_resource",
		Description: "Create a new resource in the edit area. Result status is create_draft until publish via web UI/API. " +
			"Requires write scope. Name conflicts must be resolved by choosing a different name.",
	}, createResourceHandler)

	// update_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "update_resource",
		Description: "Replace an existing resource config in the edit area (full config required; no partial patch). " +
			"Status transitions to update_draft unless still create_draft. Requires write scope.",
	}, updateResourceHandler)

	// delete_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_resource",
		Description: "Delete resources by ID. create_draft resources are hard-deleted; " +
			"published resources are marked delete_draft. " +
			"Performs dependency checks and blocks unsafe deletes. Requires write scope.",
	}, deleteResourceHandler)

	// revert_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "revert_resource",
		Description: "Revert resources to the last synced snapshot state (gateway_sync_data), discarding local drafts. " +
			"Requires write scope.",
	}, revertResourceHandler)
}

// listResourceHandler handles the list_resource tool call
func listResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListResourceInput,
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
	page := max(input.Page, 1)
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Build status filter
	var statuses []string
	if input.Status != "" {
		statuses = []string{input.Status}
	}

	// Use biz layer to list resources
	resources, total, err := biz.ListResourcesWithPagination(
		ctx,
		resourceType,
		input.Name,
		statuses,
		offset,
		pageSize,
	)
	if err != nil {
		return errorResult(err), nil, nil
	}

	result := map[string]any{
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"resource_type": input.ResourceType,
		"resources":     resources,
	}

	return successResult(result), nil, nil
}

// getResourceHandler handles the get_resource tool call
func getResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetResourceInput,
) (*mcp.CallToolResult, any, error) {
	if input.ResourceID == "" {
		return errorResult(fmt.Errorf("resource_id is required")), nil, nil
	}

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

	resource, err := biz.GetResourceByID(ctx, resourceType, input.ResourceID)
	if err != nil {
		return errorResult(err), nil, nil
	}

	return successResult(resource), nil, nil
}

// createResourceHandler handles the create_resource tool call
func createResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input CreateResourceInput,
) (*mcp.CallToolResult, any, error) {
	if input.Name == "" {
		return errorResult(fmt.Errorf("name is required")), nil, nil
	}

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

	// Generate resource ID
	resourceID := idx.GenResourceID(resourceType)

	// Marshal config to JSON bytes
	config, err := json.Marshal(input.Config)
	if err != nil {
		return errorResult(fmt.Errorf("failed to marshal config: %w", err)), nil, nil
	}

	draft, err := resourcecodec.PrepareRequestDraft(resourcecodec.RequestInput{
		Source:       "mcp",
		Operation:    constant.OperationTypeCreate,
		GatewayID:    gateway.ID,
		ResourceType: resourceType,
		Version:      gateway.GetAPISIXVersionX(),
		PathID:       resourceID,
		OuterName:    input.Name,
		OuterFields:  mcpOuterFields(resourceType, input.Config),
		Config:       config,
	})
	if err == nil {
		config, err = resourcecodec.BuildStorageConfig(draft)
		if err != nil {
			return errorResult(fmt.Errorf("failed to build config: %w", err)), nil, nil
		}
	}

	// Create resource model
	resource := model.ResourceCommonModel{
		ID:        resourceID,
		GatewayID: gateway.ID,
		Config:    datatypes.JSON(config),
		Status:    constant.ResourceStatusCreateDraft,
		BaseModel: model.BaseModel{
			Creator: "mcp",
			Updater: "mcp",
		},
	}
	if err == nil {
		resource.NameValue = draft.Identity.NameValue
		resource.ServiceIDValue = draft.Identity.Associations["service_id"]
		resource.UpstreamIDValue = draft.Identity.Associations["upstream_id"]
		resource.PluginConfigIDValue = draft.Identity.Associations["plugin_config_id"]
		resource.GroupIDValue = draft.Identity.Associations["group_id"]
		resource.SSLIDValue = draft.Identity.Associations["tls.client_cert_id"]
	}

	// Convert to specific resource type and create
	specificResource := resource.ToResourceModel(resourceType)
	if specificResource == nil {
		return errorResult(fmt.Errorf("unsupported resource type: %s", input.ResourceType)), nil, nil
	}

	err = biz.CreateResource(ctx, resourceType, specificResource, input.Name)
	if err != nil {
		return errorResult(err), nil, nil
	}

	return successResult(map[string]any{
		"message":       "Resource created successfully",
		"resource_id":   resourceID,
		"resource_type": input.ResourceType,
		"status":        constant.ResourceStatusCreateDraft,
	}), nil, nil
}

// updateResourceHandler handles the update_resource tool call
func updateResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input UpdateResourceInput,
) (*mcp.CallToolResult, any, error) {
	if input.ResourceID == "" || input.Config == nil {
		return errorResult(fmt.Errorf("resource_id and config are required")), nil, nil
	}

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

	// Get the update status based on current status
	updateStatus, err := biz.GetResourceUpdateStatus(ctx, resourceType, input.ResourceID)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Marshal config to JSON bytes
	config, err := json.Marshal(input.Config)
	if err != nil {
		return errorResult(fmt.Errorf("failed to marshal config: %w", err)), nil, nil
	}

	draft, prepareErr := resourcecodec.PrepareRequestDraft(resourcecodec.RequestInput{
		Source:       "mcp",
		Operation:    constant.OperationTypeUpdate,
		GatewayID:    gateway.ID,
		ResourceType: resourceType,
		Version:      gateway.GetAPISIXVersionX(),
		PathID:       input.ResourceID,
		OuterName:    input.Name,
		OuterFields:  mcpOuterFields(resourceType, input.Config),
		Config:       config,
	})
	if prepareErr == nil {
		config, err = resourcecodec.BuildStorageConfig(draft)
		if err != nil {
			return errorResult(fmt.Errorf("failed to build config: %w", err)), nil, nil
		}
	} else if input.Name != "" {
		nameKey := model.GetResourceNameKey(resourceType)
		config, _ = sjson.SetBytes(config, nameKey, input.Name)
	}

	// Use UpdateResourceByTypeAndID which properly handles name updates
	err = biz.UpdateResourceByTypeAndID(
		ctx,
		resourceType,
		input.ResourceID,
		input.Name,
		datatypes.JSON(config),
		updateStatus,
	)
	if err != nil {
		return errorResult(err), nil, nil
	}

	return successResult(map[string]any{
		"message":       "Resource updated successfully",
		"resource_id":   input.ResourceID,
		"resource_type": input.ResourceType,
		"status":        updateStatus,
	}), nil, nil
}

func mcpOuterFields(resourceType constant.APISIXResource, config map[string]any) map[string]any {
	outer := map[string]any{}
	for _, key := range []string{"service_id", "upstream_id", "plugin_config_id", "group_id"} {
		if value, ok := config[key]; ok {
			outer[key] = value
		}
	}
	if resourceType == constant.Upstream {
		if tlsRaw, ok := config["tls"].(map[string]any); ok {
			if value, ok := tlsRaw["client_cert_id"]; ok {
				outer["tls.client_cert_id"] = value
			}
		}
	}
	return outer
}

// deleteResourceHandler handles the delete_resource tool call
func deleteResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input DeleteResourceInput,
) (*mcp.CallToolResult, any, error) {
	if len(input.ResourceIDs) == 0 {
		return errorResult(fmt.Errorf("resource_ids is required")), nil, nil
	}

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

	// Check if any of the resources are referenced by other resources
	references, err := biz.CheckResourceReferences(ctx, resourceType, input.ResourceIDs)
	if err != nil {
		return errorResult(fmt.Errorf("failed to check resource references: %w", err)), nil, nil
	}

	// If any resources are referenced, return an error with details
	if len(references) > 0 {
		// Build a detailed error message
		var blockedResources []string
		for resourceID, refs := range references {
			blockedResources = append(
				blockedResources,
				fmt.Sprintf(
					"resource '%s' is referenced by: %s",
					resourceID,
					biz.FormatResourceReferences(refs),
				),
			)
		}

		return errorResult(fmt.Errorf(
			"cannot delete %s resources because they are referenced by other resources. "+
				"You must delete the referencing resources first:\n%s",
			input.ResourceType, strings.Join(blockedResources, "\n"))), nil, nil
	}

	deletedCount := 0
	markedCount := 0
	failedItems := make([]batchOperationFailure, 0)

	for _, resourceID := range input.ResourceIDs {
		// Get current resource to check status
		resource, err := biz.GetResourceByID(ctx, resourceType, resourceID)
		if err != nil {
			failedItems = append(failedItems, batchOperationFailure{
				ResourceID: resourceID,
				Stage:      "get",
				Error:      err.Error(),
			})
			continue
		}

		// create_draft can be hard deleted
		if resource.Status == constant.ResourceStatusCreateDraft {
			err = biz.BatchDeleteResourceByIDs(ctx, resourceType, []string{resourceID})
			if err == nil {
				deletedCount++
			} else {
				failedItems = append(failedItems, batchOperationFailure{
					ResourceID: resourceID,
					Stage:      "delete",
					Error:      err.Error(),
				})
			}
		} else {
			// Others are marked as delete_draft
			err = biz.UpdateResourceStatusWithAuditLog(
				ctx,
				resourceType,
				resourceID,
				constant.ResourceStatusDeleteDraft,
			)
			if err == nil {
				markedCount++
			} else {
				failedItems = append(failedItems, batchOperationFailure{
					ResourceID: resourceID,
					Stage:      "mark_delete",
					Error:      err.Error(),
				})
			}
		}
	}

	successCount := deletedCount + markedCount
	return successResult(buildBatchOperationResult(
		"Delete operation completed",
		input.ResourceType,
		len(input.ResourceIDs),
		successCount,
		map[string]int{
			"hard_deleted_count":  deletedCount,
			"marked_delete_count": markedCount,
		},
		failedItems,
	)), nil, nil
}

// revertResourceHandler handles the revert_resource tool call
func revertResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input RevertResourceInput,
) (*mcp.CallToolResult, any, error) {
	if len(input.ResourceIDs) == 0 {
		return errorResult(fmt.Errorf("resource_ids is required")), nil, nil
	}

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

	revertedCount := 0
	failedItems := make([]batchOperationFailure, 0)
	for _, resourceID := range input.ResourceIDs {
		err := biz.RevertResource(ctx, resourceType, resourceID)
		if err == nil {
			revertedCount++
		} else {
			failedItems = append(failedItems, batchOperationFailure{
				ResourceID: resourceID,
				Stage:      "revert",
				Error:      err.Error(),
			})
		}
	}

	return successResult(buildBatchOperationResult(
		"Revert operation completed",
		input.ResourceType,
		len(input.ResourceIDs),
		revertedCount,
		map[string]int{
			"reverted_count": revertedCount,
		},
		failedItems,
	)), nil, nil
}
