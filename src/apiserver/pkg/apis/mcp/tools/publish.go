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
)

// RegisterPublishTools registers all publish-related MCP tools
func RegisterPublishTools(server *mcp.Server) {
	// publish_preview
	server.AddTool(&mcp.Tool{
		Name:        "publish_preview",
		Description: "Preview pending changes before publishing. Shows what resources will be created, updated, or deleted in etcd/APISIX.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Filter by resource type (optional). " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"resource_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Filter by specific resource IDs (optional)",
				},
			},
			"required": []string{"gateway_id"},
		},
	}, publishPreviewHandler)

	// publish_resource
	server.AddTool(&mcp.Tool{
		Name:        "publish_resource",
		Description: "Publish specific resources to etcd/APISIX. This applies the changes from the edit area to the data plane.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type to publish. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"resource_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Array of resource IDs to publish (required)",
				},
			},
			"required": []string{"gateway_id", "resource_type", "resource_ids"},
		},
	}, publishResourceHandler)

	// publish_all
	server.AddTool(&mcp.Tool{
		Name:        "publish_all",
		Description: "Publish all pending changes (draft resources) to etcd/APISIX. This is a convenience tool for publishing all modified resources at once.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
			},
			"required": []string{"gateway_id"},
		},
	}, publishAllHandler)
}

// publishPreviewHandler handles the publish_preview tool call
func publishPreviewHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceIDs := getStringArrayParamFromArgs(args, "resource_ids")

	if gatewayID == 0 {
		return errorResult(fmt.Errorf("gateway_id is required")), nil
	}

	_, ctx, err := getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	// Build resource type list
	var resourceTypes []constant.APISIXResource
	if resourceTypeStr != "" {
		resourceType, err := parseResourceType(resourceTypeStr)
		if err != nil {
			return errorResult(err), nil
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
		"gateway_id":   gatewayID,
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

		resources, _, err := biz.ListResourcesWithPagination(ctx, rt, "", draftStatuses, 0, 1000)
		if err != nil {
			continue
		}

		for _, res := range resources {
			resMap, ok := res.(map[string]any)
			if !ok {
				continue
			}

			// Filter by resource IDs if specified
			if len(resourceIDs) > 0 {
				resID, _ := resMap["id"].(string)
				if !contains(resourceIDs, resID) {
					continue
				}
			}

			status, _ := resMap["status"].(string)
			info := map[string]any{
				"resource_type": rt.String(),
				"id":            resMap["id"],
				"name":          resMap["name"],
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

	return successResult(preview), nil
}

// publishResourceHandler handles the publish_resource tool call
func publishResourceHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceIDs := getStringArrayParamFromArgs(args, "resource_ids")

	if gatewayID == 0 || resourceTypeStr == "" || len(resourceIDs) == 0 {
		return errorResult(fmt.Errorf("gateway_id, resource_type, and resource_ids are required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	gateway, ctx, err := getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	// Publish resources
	err = biz.PublishResourcesByType(ctx, gateway, resourceType, resourceIDs)
	if err != nil {
		return errorResult(fmt.Errorf("publish failed: %w", err)), nil
	}

	return successResult(map[string]any{
		"message":         "Resources published successfully",
		"gateway_id":      gatewayID,
		"resource_type":   resourceTypeStr,
		"published_ids":   resourceIDs,
		"published_count": len(resourceIDs),
	}), nil
}

// publishAllHandler handles the publish_all tool call
func publishAllHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)

	if gatewayID == 0 {
		return errorResult(fmt.Errorf("gateway_id is required")), nil
	}

	gateway, ctx, err := getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

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
		"gateway_id":        gatewayID,
		"total_published":   totalPublished,
		"published_by_type": publishedCounts,
	}), nil
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
