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

// Package tools provides MCP tool implementations for the BK Micro APIGateway
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ValidResourceTypes lists all valid APISIX resource types
var ValidResourceTypes = []string{
	constant.Route.String(),
	constant.Service.String(),
	constant.Upstream.String(),
	constant.Consumer.String(),
	constant.ConsumerGroup.String(),
	constant.PluginConfig.String(),
	constant.GlobalRule.String(),
	constant.PluginMetadata.String(),
	constant.Proto.String(),
	constant.SSL.String(),
	constant.StreamRoute.String(),
}

// ValidResourceStatuses lists all valid resource statuses
var ValidResourceStatuses = []string{
	string(constant.ResourceStatusCreateDraft),
	string(constant.ResourceStatusUpdateDraft),
	string(constant.ResourceStatusDeleteDraft),
	string(constant.ResourceStatusSuccess),
}

// ValidAPISIXVersions lists APISIX versions for schema validation tools.
var ValidAPISIXVersions = []string{
	string(constant.APISIXVersion313),
}

// WriteToolNames defines MCP tools that require write access scope
var WriteToolNames = map[string]bool{
	"create_resource":                   true,
	"update_resource":                   true,
	"delete_resource":                   true,
	"revert_resource":                   true,
	"publish_resource":                  true,
	"publish_all":                       true,
	"add_synced_resources_to_edit_area": true,
}

// IsWriteTool checks if the given tool name requires write access
func IsWriteTool(toolName string) bool {
	return WriteToolNames[toolName]
}

// CheckWriteScope checks if the token has write scope for write tools
// Returns an error if write access is required but not granted
func CheckWriteScope(ctx context.Context) error {
	token := middleware.GetMCPAccessTokenFromContext(ctx)
	if token == nil {
		return fmt.Errorf("no access token found in context")
	}
	if !token.CanWrite() {
		return biz.ErrMCPInsufficientScope
	}
	return nil
}

// getGatewayFromContext retrieves the gateway info from context
// First tries to get from context directly (set by middleware for Gin handlers)
// If not found (MCP SDK uses its own context), fetches using the token's GatewayID
func getGatewayFromContext(ctx context.Context) (*model.Gateway, error) {
	// Try to get gateway from context first (works for Gin context)
	gateway := ginx.GetGatewayInfoFromContext(ctx)
	if gateway != nil {
		return gateway, nil
	}

	// Fallback: get gateway ID from token and fetch from database
	// This is needed because MCP SDK creates its own context that doesn't
	// inherit from the HTTP request context
	token := middleware.GetMCPAccessTokenFromContext(ctx)
	if token == nil {
		return nil, fmt.Errorf("gateway not found in context and no access token available")
	}

	// Fetch gateway from database using the token's gateway ID
	gateway, err := biz.GetGateway(ctx, token.GatewayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateway: %w", err)
	}

	return gateway, nil
}

// parseResourceType converts string to APISIXResource type
func parseResourceType(resourceType string) (constant.APISIXResource, error) {
	rt := constant.APISIXResource(resourceType)
	if _, ok := constant.ResourceTypeMap[rt]; !ok {
		return "", fmt.Errorf("invalid resource type: %s", resourceType)
	}
	return rt, nil
}

// parseAPISIXVersion converts string to APISIXVersion type
func parseAPISIXVersion(version string) (constant.APISIXVersion, error) {
	v := constant.APISIXVersion(version)
	switch v {
	case constant.APISIXVersion311, constant.APISIXVersion313:
		return v, nil
	default:
		return "", fmt.Errorf("invalid APISIX version: %s", version)
	}
}

// toJSON converts an object to JSON string
func toJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(data)
}

// successResult creates a successful tool result
func successResult(data any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: toJSON(data)},
		},
	}
}

// errorResult creates an error tool result
func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}
}

// parseArguments parses the raw arguments from the request (call once per handler)
// Returns an empty map if arguments are empty or if parsing fails.
func parseArguments(req *mcp.CallToolRequest) map[string]any {
	if len(req.Params.Arguments) == 0 {
		return make(map[string]any)
	}
	var args map[string]any
	// Intentionally ignore unmarshal errors and return empty map
	// This allows handlers to proceed with default values
	_ = json.Unmarshal(req.Params.Arguments, &args)
	if args == nil {
		return make(map[string]any)
	}
	return args
}

// getIntParamFromArgs extracts an integer parameter from pre-parsed arguments
func getIntParamFromArgs(args map[string]any, name string, defaultVal int) int {
	if args == nil {
		return defaultVal
	}
	if val, ok := args[name]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		}
	}
	return defaultVal
}

// getStringParamFromArgs extracts a string parameter from pre-parsed arguments
func getStringParamFromArgs(args map[string]any, name, defaultVal string) string {
	if args == nil {
		return defaultVal
	}
	if val, ok := args[name]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultVal
}

// getStringArrayParamFromArgs extracts a string array parameter from pre-parsed arguments
func getStringArrayParamFromArgs(args map[string]any, name string) []string {
	if args == nil {
		return nil
	}
	if val, ok := args[name]; ok {
		if arr, ok := val.([]any); ok {
			result := make([]string, 0, len(arr))
			for _, v := range arr {
				if s, ok := v.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}

// getObjectParamFromArgs extracts an object parameter from pre-parsed arguments
func getObjectParamFromArgs(args map[string]any, name string) (json.RawMessage, error) {
	if args == nil {
		return nil, nil
	}
	if val, ok := args[name]; ok {
		data, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %s: %w", name, err)
		}
		return data, nil
	}
	return nil, nil
}

// ResourceTypeDescription returns a description of valid resource types
func ResourceTypeDescription() string {
	return fmt.Sprintf("One of: %s", strings.Join(ValidResourceTypes, ", "))
}

// StatusDescription returns a description of valid resource statuses
func StatusDescription() string {
	return fmt.Sprintf("One of: %s", strings.Join(ValidResourceStatuses, ", "))
}

// APISIXVersionDescription returns a description of valid APISIX versions
func APISIXVersionDescription() string {
	return fmt.Sprintf("One of: %s", strings.Join(ValidAPISIXVersions, ", "))
}
