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
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
)

// RegisterDiffTools registers all diff-related MCP tools
func RegisterDiffTools(server *mcp.Server) {
	// diff_resources
	server.AddTool(&mcp.Tool{
		Name: "diff_resources",
		Description: "Compare resources between the edit area and the sync snapshot. " +
			"Shows what changes would be applied when publishing.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gateway_id": map[string]any{
					"type":        "integer",
					"description": "The gateway ID (required)",
				},
				"resource_type": map[string]any{
					"type": "string",
					"description": "Filter by resource type (optional). " +
						"If omitted, shows diff for all types. " + ResourceTypeDescription(),
					"enum": ValidResourceTypes,
				},
				"resource_id": map[string]any{
					"type":        "string",
					"description": "Filter by specific resource ID (optional)",
				},
			},
			"required": []string{"gateway_id"},
		},
	}, diffResourcesHandler)

	// diff_detail
	server.AddTool(&mcp.Tool{
		Name: "diff_detail",
		Description: "Get detailed JSON diff for a single resource. " +
			"Shows exact changes between edit area and sync snapshot.",
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
					"description": "The resource ID to get diff for (required)",
				},
			},
			"required": []string{"gateway_id", "resource_type", "resource_id"},
		},
	}, diffDetailHandler)
}

// diffResourcesHandler handles the diff_resources tool call
func diffResourcesHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceID := getStringParamFromArgs(args, "resource_id", "")

	if gatewayID == 0 {
		return errorResult(fmt.Errorf("gateway_id is required")), nil
	}

	_, ctx, err := getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	var resourceType constant.APISIXResource
	if resourceTypeStr != "" {
		resourceType, err = parseResourceType(resourceTypeStr)
		if err != nil {
			return errorResult(err), nil
		}
	}

	// Get diff results - DiffResources returns results for all resource types
	var idList []string
	if resourceID != "" {
		idList = []string{resourceID}
	}

	// Call DiffResources once - it returns results organized by resource type internally
	diffResults, err := biz.DiffResources(ctx, resourceType, idList, "", nil, resourceID == "")
	if err != nil {
		return errorResult(fmt.Errorf("failed to get diff: %w", err)), nil
	}

	// Organize results by resource type
	allDiffs := make(map[string]any)
	for _, diffResult := range diffResults {
		rtStr := diffResult.ResourceType.String()
		// Filter by requested resource type if specified
		if resourceType != "" && diffResult.ResourceType != resourceType {
			continue
		}
		allDiffs[rtStr] = []any{diffResult}
	}

	// Build summary
	summary := buildDiffSummary(allDiffs)

	result := map[string]any{
		"gateway_id": gatewayID,
		"summary":    summary,
		"details":    allDiffs,
	}

	return successResult(result), nil
}

// diffDetailHandler handles the diff_detail tool call
func diffDetailHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	gatewayID := getIntParamFromArgs(args, "gateway_id", 0)
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	resourceID := getStringParamFromArgs(args, "resource_id", "")

	if gatewayID == 0 || resourceTypeStr == "" || resourceID == "" {
		return errorResult(fmt.Errorf("gateway_id, resource_type, and resource_id are required")), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	_, ctx, err = getGatewayFromRequest(ctx, gatewayID)
	if err != nil {
		return errorResult(err), nil
	}

	// Get detailed config diff
	diffDetail, err := biz.GetResourceConfigDiffDetail(ctx, resourceType, resourceID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get diff detail: %w", err)), nil
	}

	result := map[string]any{
		"gateway_id":    gatewayID,
		"resource_type": resourceTypeStr,
		"resource_id":   resourceID,
		"diff":          diffDetail,
	}

	return successResult(result), nil
}

// buildDiffSummary builds a summary of diff results
func buildDiffSummary(diffs map[string]any) map[string]any {
	summary := map[string]any{
		"total_changes":  0,
		"create_count":   0,
		"update_count":   0,
		"delete_count":   0,
		"resource_types": []string{},
	}

	resourceTypes := []string{}
	totalChanges := 0
	createCount := 0
	updateCount := 0
	deleteCount := 0

	for rt, diffData := range diffs {
		resourceTypes = append(resourceTypes, rt)

		// Handle []any containing dto.ResourceChangeInfo
		if diffList, ok := diffData.([]any); ok {
			for _, item := range diffList {
				if changeInfo, ok := item.(dto.ResourceChangeInfo); ok {
					createCount += changeInfo.AddedCount
					updateCount += changeInfo.UpdateCount
					deleteCount += changeInfo.DeletedCount
					totalChanges += len(changeInfo.ChangeDetail)
				}
			}
		}
	}

	summary["total_changes"] = totalChanges
	summary["create_count"] = createCount
	summary["update_count"] = updateCount
	summary["delete_count"] = deleteCount
	summary["resource_types"] = resourceTypes

	return summary
}
