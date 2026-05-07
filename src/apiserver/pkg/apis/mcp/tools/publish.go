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
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	mcpbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/mcp"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ============================================================================
// Input Types for Publish Tools
// ============================================================================

// PublishPreviewInput is the input for the publish_preview tool
type PublishPreviewInput struct {
	ResourceType string   `json:"resource_type,omitempty" jsonschema:"Optional resource type filter for preview."`
	ResourceIDs  []string `json:"resource_ids,omitempty" jsonschema:"Optional resource ID filter for preview."`
}

// RegisterPublishTools registers all publish-related MCP tools
func RegisterPublishTools(server *mcp.Server) {
	// publish_preview
	mcp.AddTool(server, &mcp.Tool{
		Name: "publish_preview",
		Description: "Preview pending draft changes (create/update/delete) that would be published. " +
			"Read-only preview. Actual publish via MCP is disabled. Returns up to 2000 resources per query.",
	}, publishPreviewHandler)

	// NOTE: publish_resource and publish_all are commented out for safety.
	// Publishing directly via MCP is dangerous for production environments.
	// Users should use the web UI or API for publishing operations.
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
		resources, _, err := mcpbiz.ListResourcesWithPagination(ctx, rt, "", draftStatuses, 0, 2000)
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

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
