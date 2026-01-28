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

// ValidAPISIXVersions lists all supported APISIX versions for schema validation
var ValidAPISIXVersions = []string{
	string(constant.APISIXVersion311),
	string(constant.APISIXVersion313),
}

// getGatewayFromRequest retrieves the gateway info and sets it in context
func getGatewayFromRequest(ctx context.Context, gatewayID int) (*model.Gateway, context.Context, error) {
	if token := middleware.GetMCPAccessTokenFromContext(ctx); token != nil && token.GatewayID != gatewayID {
		return nil, ctx, fmt.Errorf("access token does not match gateway_id")
	}

	gateway, err := biz.GetGateway(ctx, gatewayID)
	if err != nil {
		return nil, ctx, fmt.Errorf("gateway not found: %w", err)
	}

	// Set gateway info in context for downstream functions
	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)

	return gateway, ctx, nil
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
