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

	diffbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/diff"
	unifyopbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/unifyop"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ============================================================================
// Input Types for Diff Tools
// ============================================================================

// DiffResourcesInput is the input for the diff_resources tool
type DiffResourcesInput struct {
	ResourceType string `json:"resource_type,omitempty" jsonschema:"Optional resource type filter."`
	ResourceID   string `json:"resource_id,omitempty" jsonschema:"Optional resource ID filter for scoped diff."`
}

// DiffDetailInput is the input for the diff_detail tool
type DiffDetailInput struct {
	ResourceType string `json:"resource_type" jsonschema:"Required. APISIX resource type."`
	ResourceID   string `json:"resource_id" jsonschema:"Required. Resource ID to get detailed diff for."`
}

// RegisterDiffTools registers all diff-related MCP tools
func RegisterDiffTools(server *mcp.Server) {
	// diff_resources
	mcp.AddTool(server, &mcp.Tool{
		Name: "diff_resources",
		Description: "Compare edit area vs sync snapshot and return pending publish impact by resource type. " +
			"Includes before_status/after_status transition hints for each change.",
	}, diffResourcesHandler)

	// diff_detail
	mcp.AddTool(server, &mcp.Tool{
		Name:        "diff_detail",
		Description: "Return field-level JSON diff for one resource between edit area and sync snapshot.",
	}, diffDetailHandler)
}

// diffResourcesHandler handles the diff_resources tool call
func diffResourcesHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input DiffResourcesInput,
) (*mcp.CallToolResult, any, error) {
	// Gateway is already set in context by middleware
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Set gateway info in context for downstream biz functions that use ginx.GetGatewayInfoFromContext
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	var resourceType constant.APISIXResource
	if input.ResourceType != "" {
		resourceType, err = parseResourceType(input.ResourceType)
		if err != nil {
			return errorResult(err), nil, nil
		}
	}

	// Get diff results - DiffResources returns results for all resource types
	var idList []string
	if input.ResourceID != "" {
		idList = []string{input.ResourceID}
	}

	// Call DiffResources once - it returns results organized by resource type internally
	diffResults, err := diffbiz.DiffResources(ctx, resourceType, idList, "", nil, input.ResourceID == "")
	if err != nil {
		return errorResult(fmt.Errorf("failed to get diff: %w", err)), nil, nil
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
		"gateway_id": gateway.ID,
		"summary":    summary,
		"details":    allDiffs,
	}

	return successResult(result), nil, nil
}

// diffDetailHandler handles the diff_detail tool call
func diffDetailHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input DiffDetailInput,
) (*mcp.CallToolResult, any, error) {
	if input.ResourceType == "" || input.ResourceID == "" {
		return errorResult(fmt.Errorf("resource_type and resource_id are required")), nil, nil
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

	// Get detailed config diff
	diffDetail, err := unifyopbiz.GetResourceConfigDiffDetail(ctx, resourceType, input.ResourceID)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get diff detail: %w", err)), nil, nil
	}

	result := map[string]any{
		"gateway_id":    gateway.ID,
		"resource_type": input.ResourceType,
		"resource_id":   input.ResourceID,
		"diff":          diffDetail,
	}

	return successResult(result), nil, nil
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
