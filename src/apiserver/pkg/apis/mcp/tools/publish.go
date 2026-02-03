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
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ============================================================================
// Input Types for Publish Tools
// ============================================================================

// PublishPreviewInput is the input for the publish_preview tool
type PublishPreviewInput struct {
	ResourceType string   `json:"resource_type,omitempty" jsonschema:"filter by resource type (optional)"`
	ResourceIDs  []string `json:"resource_ids,omitempty" jsonschema:"filter by specific resource IDs (optional)"`
}

// PublishResourceInput is the input for the publish_resource tool (currently disabled)
type PublishResourceInput struct {
	ResourceType string   `json:"resource_type" jsonschema:"resource type to publish"`
	ResourceIDs  []string `json:"resource_ids" jsonschema:"array of resource IDs to publish (required)"`
}

// PublishAllInput is the input for the publish_all tool (currently disabled)
// It has no required fields
type PublishAllInput struct{}

// RegisterPublishTools registers all publish-related MCP tools
func RegisterPublishTools(server *mcp.Server) {
	// publish_preview
	mcp.AddTool(server, &mcp.Tool{
		Name: "publish_preview",
		Description: "Preview pending changes before publishing. " +
			"Shows what resources will be created, updated, or deleted in etcd/APISIX." +
			"Currently only list 2000 resources at most",
	}, publishPreviewHandler)

	// NOTE: publish_resource and publish_all are commented out for safety.
	// Publishing directly via MCP is dangerous for production environments.
	// Users should use the web UI or API for publishing operations.

	// // publish_resource
	// mcp.AddTool(server, &mcp.Tool{
	// 	Name: "publish_resource",
	// 	Description: "Publish specific resources to etcd/APISIX. " +
	// 		"Applies the changes from the edit area to the data plane.",
	// }, publishResourceHandler)

	// // publish_all
	// mcp.AddTool(server, &mcp.Tool{
	// 	Name: "publish_all",
	// 	Description: "Publish all pending changes (draft resources) to etcd/APISIX. " +
	// 		"Convenience tool for publishing all modified resources at once.",
	// }, publishAllHandler)
}

// publishPreviewHandler handles the publish_preview tool call
func publishPreviewHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input PublishPreviewInput,
) (*mcp.CallToolResult, any, error) {
	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	// Build resource type list
	var resourceTypes []constant.APISIXResource
	if input.ResourceType != "" {
		resourceType, err := parseResourceType(input.ResourceType)
		if err != nil {
			return errorResult(err), nil, nil
		}
		resourceTypes = []constant.APISIXResource{resourceType}
	} else {
		// All resource types
		for rt := range constant.ResourceTypeMap {
			resourceTypes = append(resourceTypes, rt)
		}
	}

	// Get pending changes for each resource type
	preview := map[string]any{
		"gateway_id":   gateway.ID,
		"create_draft": []map[string]any{},
		"update_draft": []map[string]any{},
		"delete_draft": []map[string]any{},
		"summary": map[string]int{
			"create_count": 0,
			"update_count": 0,
			"delete_count": 0,
		},
	}

	createList := []map[string]any{}
	updateList := []map[string]any{}
	deleteList := []map[string]any{}

	for _, rt := range resourceTypes {
		// Get resources with draft status
		draftStatuses := []string{
			string(constant.ResourceStatusCreateDraft),
			string(constant.ResourceStatusUpdateDraft),
			string(constant.ResourceStatusDeleteDraft),
		}

		// TODO: currently the limit is 1000, if exceed the limit, we need to use a new solution to this tool.
		resources, _, err := biz.ListResourcesWithPagination(ctx, rt, "", draftStatuses, 0, 2000)
		if err != nil {
			continue
		}

		for _, res := range resources {
			resMap, ok := res.(map[string]any)
			if !ok {
				continue
			}

			// Filter by resource IDs if specified
			if len(input.ResourceIDs) > 0 {
				resID, _ := resMap["id"].(string)
				if !contains(input.ResourceIDs, resID) {
					continue
				}
			}

			// Determine the display name for the resource. Most resources use "name",
			// but Consumer uses "username" instead.
			var name any
			if rt == constant.Consumer {
				name = resMap["username"]
			} else {
				name = resMap["name"]
			}

			status, _ := resMap["status"].(string)
			info := map[string]any{
				"resource_type": rt.String(),
				"id":            resMap["id"],
				"name":          name,
			}

			switch constant.ResourceStatus(status) {
			case constant.ResourceStatusCreateDraft:
				createList = append(createList, info)
			case constant.ResourceStatusUpdateDraft:
				updateList = append(updateList, info)
			case constant.ResourceStatusDeleteDraft:
				deleteList = append(deleteList, info)
			}
		}
	}

	preview["create_draft"] = createList
	preview["update_draft"] = updateList
	preview["delete_draft"] = deleteList
	preview["summary"] = map[string]int{
		"create_count": len(createList),
		"update_count": len(updateList),
		"delete_count": len(deleteList),
		"total":        len(createList) + len(updateList) + len(deleteList),
	}

	return successResult(preview), nil, nil
}

//nolint:unused // Kept for future use when MCP publishing is enabled
func publishResourceHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input PublishResourceInput,
) (*mcp.CallToolResult, any, error) {
	if input.ResourceType == "" || len(input.ResourceIDs) == 0 {
		return errorResult(fmt.Errorf("resource_type and resource_ids are required")), nil, nil
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

	// Publish resources
	err = biz.PublishResourcesByType(ctx, gateway, resourceType, input.ResourceIDs)
	if err != nil {
		return errorResult(fmt.Errorf("publish failed: %w", err)), nil, nil
	}

	return successResult(map[string]any{
		"message":         "Resources published successfully",
		"gateway_id":      gateway.ID,
		"resource_type":   input.ResourceType,
		"published_ids":   input.ResourceIDs,
		"published_count": len(input.ResourceIDs),
	}), nil, nil
}

//nolint:unused // Kept for future use when MCP publishing is enabled
func publishAllHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input PublishAllInput,
) (*mcp.CallToolResult, any, error) {
	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	// Get all draft resources and publish them
	publishedCounts := map[string]int{}
	totalPublished := 0

	for rt := range constant.ResourceTypeMap {
		draftStatuses := []string{
			string(constant.ResourceStatusCreateDraft),
			string(constant.ResourceStatusUpdateDraft),
			string(constant.ResourceStatusDeleteDraft),
		}

		resources, _, err := biz.ListResourcesWithPagination(ctx, rt, "", draftStatuses, 0, 1000)
		if err != nil {
			continue
		}

		if len(resources) == 0 {
			continue
		}

		// Extract resource IDs
		var resourceIDs []string
		for _, res := range resources {
			if resMap, ok := res.(map[string]any); ok {
				if id, ok := resMap["id"].(string); ok {
					resourceIDs = append(resourceIDs, id)
				}
			}
		}

		if len(resourceIDs) == 0 {
			continue
		}

		// Publish
		err = biz.PublishResourcesByType(ctx, gateway, rt, resourceIDs)
		if err != nil {
			continue
		}

		publishedCounts[rt.String()] = len(resourceIDs)
		totalPublished += len(resourceIDs)
	}

	return successResult(map[string]any{
		"message":           "All draft resources published",
		"gateway_id":        gateway.ID,
		"total_published":   totalPublished,
		"published_by_type": publishedCounts,
	}), nil, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
