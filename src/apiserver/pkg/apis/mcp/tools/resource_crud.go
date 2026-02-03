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
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// ============================================================================
// Input Types for Resource CRUD Tools
// ============================================================================

// ListResourceInput is the input for the list_resource tool
type ListResourceInput struct {
	ResourceType string `json:"resource_type" jsonschema:"resource type to list"`
	Name         string `json:"name,omitempty" jsonschema:"filter by resource name (optional, supports fuzzy match)"`
	Status       string `json:"status,omitempty" jsonschema:"filter by resource status (optional)"`
	Page         int    `json:"page,omitempty" jsonschema:"page number (default: 1)"`
	PageSize     int    `json:"page_size,omitempty" jsonschema:"number of items per page (default: 20, max: 100)"`
}

// GetResourceInput is the input for the get_resource tool
type GetResourceInput struct {
	ResourceType string `json:"resource_type" jsonschema:"resource type (REQUIRED)"`
	ResourceID   string `json:"resource_id" jsonschema:"the resource ID to retrieve (required)"`
}

// CreateResourceInput is the input for the create_resource tool
type CreateResourceInput struct {
	ResourceType string         `json:"resource_type" jsonschema:"resource type to create"`
	Name         string         `json:"name" jsonschema:"resource name (required for most resource types, uses 'username' for consumer)"`
	Config       map[string]any `json:"config" jsonschema:"resource configuration object following APISIX schema (required)"`
}

// UpdateResourceInput is the input for the update_resource tool
type UpdateResourceInput struct {
	ResourceType string         `json:"resource_type" jsonschema:"resource type (REQUIRED)"`
	ResourceID   string         `json:"resource_id" jsonschema:"the resource ID to update (required)"`
	Name         string         `json:"name,omitempty" jsonschema:"new resource name (optional)"`
	Config       map[string]any `json:"config" jsonschema:"new resource configuration (required)"`
}

// DeleteResourceInput is the input for the delete_resource tool
type DeleteResourceInput struct {
	ResourceType string   `json:"resource_type" jsonschema:"resource type (REQUIRED)"`
	ResourceIDs  []string `json:"resource_ids" jsonschema:"array of resource IDs to delete (required)"`
}

// RevertResourceInput is the input for the revert_resource tool
type RevertResourceInput struct {
	ResourceType string   `json:"resource_type" jsonschema:"resource type (REQUIRED)"`
	ResourceIDs  []string `json:"resource_ids" jsonschema:"array of resource IDs to revert (required)"`
}

// RegisterResourceCRUDTools registers all resource CRUD tools
func RegisterResourceCRUDTools(server *mcp.Server) {
	// list_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_resource",
		Description: "List resources in the edit area with pagination and filtering. " +
			"Returns resources managed by the gateway.",
	}, listResourceHandler)

	// get_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_resource",
		Description: "Get detailed information about a specific resource including its full configuration. " +
			"IMPORTANT: Both resource_type and resource_id are required. If a user only provides an ID, " +
			"you MUST ask them to specify the resource_type (e.g., 'route', 'service', 'upstream', etc.) " +
			"because the same ID format could belong to different resource types.",
	}, getResourceHandler)

	// create_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "create_resource",
		Description: "Create a new resource in the edit area. The resource will be in 'create_draft' " +
			"status until published. If you found create failed because of name conflict " +
			"(Error 1062 (23000): Duplicate entry '1' for key 'route.idx_name'), " +
			"JUST TELL USER TO CHANGE THE NAME AND TRY AGAIN, DO NOT TRY TO FIX IT FOR USER.",
	}, createResourceHandler)

	// update_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "update_resource",
		Description: "Update an existing resource. The resource status will change to 'update_draft' " +
			"until published. If you update a resource, you should get the resource first, " +
			"and modify the fields then update. " +
			"DO NOT ONLY UPDATE PART OF FIELDS IN CONFIG, YOU SHOULD UPDATE THE WHOLE CONFIG. " +
			"IMPORTANT: Both resource_type and resource_id are required. If a user only provides an ID, " +
			"you MUST ask them to specify the resource_type.",
	}, updateResourceHandler)

	// delete_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "delete_resource",
		Description: "Mark resources for deletion. Resources in 'create_draft' status " +
			"will be hard-deleted; others will be marked as 'delete_draft' until published. " +
			"IMPORTANT: Both resource_type and resource_ids are required. If a user only provides IDs, " +
			"you MUST ask them to specify the resource_type. " +
			"NOTE: This tool checks if resources are referenced by other resources before deletion. " +
			"For example, a service referenced by routes cannot be deleted until those routes are deleted first.",
	}, deleteResourceHandler)

	// revert_resource
	mcp.AddTool(server, &mcp.Tool{
		Name: "revert_resource",
		Description: "Revert resources to their synced snapshot state. " +
			"Discards all local changes for the specified resources. " +
			"IMPORTANT: Both resource_type and resource_ids are required. If a user only provides IDs, " +
			"you MUST ask them to specify the resource_type.",
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

	// Inject name into config so ToResourceModel.GetName() picks it up
	nameKey := model.GetResourceNameKey(resourceType)
	config, err = sjson.SetBytes(config, nameKey, input.Name)
	if err != nil {
		return errorResult(fmt.Errorf("failed to inject name into config: %w", err)), nil, nil
	}

	// Verify name was successfully injected
	if !gjson.GetBytes(config, nameKey).Exists() {
		return errorResult(fmt.Errorf("name field not found in config after injection")), nil, nil
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

	// If name is provided, inject it into config using the correct key for the resource type
	if input.Name != "" {
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

	for _, resourceID := range input.ResourceIDs {
		// Get current resource to check status
		resource, err := biz.GetResourceByID(ctx, resourceType, resourceID)
		if err != nil {
			continue
		}

		// create_draft can be hard deleted
		if resource.Status == constant.ResourceStatusCreateDraft {
			err = biz.BatchDeleteResourceByIDs(ctx, resourceType, []string{resourceID})
			if err == nil {
				deletedCount++
			}
		} else {
			// Others are marked as delete_draft
			err = biz.UpdateResourceStatusWithAuditLog(ctx, resourceType, resourceID, constant.ResourceStatusDeleteDraft)
			if err == nil {
				markedCount++
			}
		}
	}

	return successResult(map[string]any{
		"message":             "Delete operation completed",
		"hard_deleted_count":  deletedCount,
		"marked_delete_count": markedCount,
		"resource_type":       input.ResourceType,
	}), nil, nil
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
	for _, resourceID := range input.ResourceIDs {
		err := biz.RevertResource(ctx, resourceType, resourceID)
		if err == nil {
			revertedCount++
		}
	}

	return successResult(map[string]any{
		"message":        "Revert operation completed",
		"reverted_count": revertedCount,
		"resource_type":  input.ResourceType,
	}), nil, nil
}
