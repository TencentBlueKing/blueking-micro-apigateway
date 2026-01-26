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
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// RegisterResourceCRUDTools registers all resource CRUD tools
func RegisterResourceCRUDTools(server *mcp.Server) {
	// list_resource
	server.AddTool(&mcp.Tool{
		Name: "list_resource",
		Description: "List resources in the edit area with pagination and filtering. " +
			"Returns resources managed by the gateway.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID to list resources from (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Filter by resource name (optional, supports fuzzy match)",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Filter by resource status (optional). " + StatusDescription(),
					"enum":        ValidResourceStatuses,
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
			"required": []string{"gateway_id", "resource_type"},
		},
	}, listResourceHandler)

	// get_resource
	server.AddTool(&mcp.Tool{
		Name:        "get_resource",
		Description: "Get detailed information about a specific resource including its full configuration.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"resource_id": map[string]any{
					"type":        "string",
					"description": "The resource ID to retrieve (required)",
				},
			},
			"required": []string{"gateway_id", "resource_type", "resource_id"},
		},
	}, getResourceHandler)

	// create_resource
	server.AddTool(&mcp.Tool{
		Name:        "create_resource",
		Description: "Create a new resource in the edit area. The resource will be in 'create_draft' status until published.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type to create. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Resource name (required for most resource types, uses 'username' for consumer)",
				},
				"config": map[string]any{
					"type":        "object",
					"description": "Resource configuration object following APISIX schema (required)",
				},
			},
			"required": []string{"gateway_id", "resource_type", "name", "config"},
		},
	}, createResourceHandler)

	// update_resource
	server.AddTool(&mcp.Tool{
		Name:        "update_resource",
		Description: "Update an existing resource. The resource status will change to 'update_draft' until published.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"resource_id": map[string]any{
					"type":        "string",
					"description": "The resource ID to update (required)",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "New resource name (optional)",
				},
				"config": map[string]any{
					"type":        "object",
					"description": "New resource configuration (required)",
				},
			},
			"required": []string{"gateway_id", "resource_type", "resource_id", "config"},
		},
	}, updateResourceHandler)

	// delete_resource
	server.AddTool(&mcp.Tool{
		Name: "delete_resource",
		Description: "Mark resources for deletion. Resources in 'create_draft' status " +
			"will be hard-deleted; others will be marked as 'delete_draft' until published.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"resource_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Array of resource IDs to delete (required)",
				},
			},
			"required": []string{"gateway_id", "resource_type", "resource_ids"},
		},
	}, deleteResourceHandler)

	// revert_resource
	server.AddTool(&mcp.Tool{
		Name: "revert_resource",
		Description: "Revert resources to their synced snapshot state. " +
			"Discards all local changes for the specified resources.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"resource_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Array of resource IDs to revert (required)",
				},
			},
			"required": []string{"gateway_id", "resource_type", "resource_ids"},
		},
	}, revertResourceHandler)
}

// listResourceHandler handles the list_resource tool call
func listResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	name := getStringParamFromArgs(args, "name", "")
	status := getStringParamFromArgs(args, "status", "")
	page := getIntParamFromArgs(args, "page", 1)
	pageSize := getIntParamFromArgs(args, "page_size", 20)

	if gatewayID == 0 {
		return errorResult(fmt.Errorf("gateway_id is required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	_, ctx, err = getGatewayFromRequest(ctx, gatewayID)
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

	// Build status filter
	var statuses []string
	if status != "" {
		statuses = []string{status}
	}

	// Use biz layer to list resources
	resources, total, err := biz.ListResourcesWithPagination(ctx, resourceType, name, statuses, offset, pageSize)
	if err != nil {
		return errorResult(err), nil
	}

	result := map[string]any{
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"resource_type": resourceTypeStr,
		"resources":     resources,
	}

	return successResult(result), nil
}

// getResourceHandler handles the get_resource tool call
func getResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceID := getStringParamFromArgs(args, "resource_id", "")

	if gatewayID == 0 || resourceID == "" {
		return errorResult(fmt.Errorf("gateway_id and resource_id are required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	_, ctx, err = getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	resource, err := biz.GetResourceByID(ctx, resourceType, resourceID)
	if err != nil {
		return errorResult(err), nil
	}

	return successResult(resource), nil
}

// createResourceHandler handles the create_resource tool call
func createResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	name := getStringParamFromArgs(args, "name", "")
	config, err := getObjectParamFromArgs(args, "config")
	if err != nil {
		return errorResult(err), nil
	}

	if gatewayID == 0 || name == "" {
		return errorResult(fmt.Errorf("gateway_id and name are required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	gateway, ctx, err := getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	// Generate resource ID
	resourceID := idx.GenResourceID(resourceType)

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
		return errorResult(fmt.Errorf("unsupported resource type: %s", resourceTypeStr)), nil
	}

	err = biz.CreateResource(ctx, resourceType, specificResource, name)
	if err != nil {
		return errorResult(err), nil
	}

	return successResult(map[string]any{
		"message":       "Resource created successfully",
		"resource_id":   resourceID,
		"resource_type": resourceTypeStr,
		"status":        constant.ResourceStatusCreateDraft,
	}), nil
}

// updateResourceHandler handles the update_resource tool call
func updateResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceID := getStringParamFromArgs(args, "resource_id", "")
	config, err := getObjectParamFromArgs(args, "config")
	if err != nil {
		return errorResult(err), nil
	}

	if gatewayID == 0 || resourceID == "" || len(config) == 0 {
		return errorResult(fmt.Errorf("gateway_id, resource_id, and config are required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	_, ctx, err = getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	// Get the update status based on current status
	updateStatus, err := biz.GetResourceUpdateStatus(ctx, resourceType, resourceID)
	if err != nil {
		return errorResult(err), nil
	}

	// Build the resource model
	resource := &model.ResourceCommonModel{
		ID:     resourceID,
		Config: datatypes.JSON(config),
		Status: updateStatus,
	}

	err = biz.UpdateResource(ctx, resourceType, resourceID, resource)
	if err != nil {
		return errorResult(err), nil
	}

	return successResult(map[string]any{
		"message":       "Resource updated successfully",
		"resource_id":   resourceID,
		"resource_type": resourceTypeStr,
		"status":        updateStatus,
	}), nil
}

// deleteResourceHandler handles the delete_resource tool call
func deleteResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceIDs := getStringArrayParamFromArgs(args, "resource_ids")

	if gatewayID == 0 || len(resourceIDs) == 0 {
		return errorResult(fmt.Errorf("gateway_id and resource_ids are required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	_, ctx, err = getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	deletedCount := 0
	markedCount := 0

	for _, resourceID := range resourceIDs {
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
		"resource_type":       resourceTypeStr,
	}), nil
}

// revertResourceHandler handles the revert_resource tool call
func revertResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceIDs := getStringArrayParamFromArgs(args, "resource_ids")

	if gatewayID == 0 || len(resourceIDs) == 0 {
		return errorResult(fmt.Errorf("gateway_id and resource_ids are required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	_, ctx, err = getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	revertedCount := 0
	for _, resourceID := range resourceIDs {
		err := biz.RevertResource(ctx, resourceType, resourceID)
		if err == nil {
			revertedCount++
		}
	}

	return successResult(map[string]any{
		"message":        "Revert operation completed",
		"reverted_count": revertedCount,
		"resource_type":  resourceTypeStr,
	}), nil
}
