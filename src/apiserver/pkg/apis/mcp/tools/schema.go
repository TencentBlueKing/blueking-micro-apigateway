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

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// ============================================================================
// Input Types for Schema Tools
// ============================================================================

// GetResourceSchemaInput is the input for the get_resource_schema tool
type GetResourceSchemaInput struct {
	APISIXVersion string `json:"apisix_version" jsonschema:"APISIX version for schema (e.g., 3.11, 3.13)"`
	ResourceType  string `json:"resource_type" jsonschema:"resource type to get schema for"`
}

// GetPluginSchemaInput is the input for the get_plugin_schema tool
type GetPluginSchemaInput struct {
	APISIXVersion string `json:"apisix_version" jsonschema:"APISIX version for schema (e.g., 3.11, 3.13)"`
	PluginName    string `json:"plugin_name" jsonschema:"plugin name (e.g., 'limit-req', 'proxy-rewrite', 'jwt-auth')"`
	SchemaType    string `json:"schema_type,omitempty" jsonschema:"schema type: main, consumer, or metadata"`
}

// ValidateResourceConfigInput is the input for the validate_resource_config tool
type ValidateResourceConfigInput struct {
	APISIXVersion string         `json:"apisix_version" jsonschema:"APISIX version to validate against (e.g., 3.11, 3.13)"`
	ResourceType  string         `json:"resource_type" jsonschema:"resource type"`
	Config        map[string]any `json:"config" jsonschema:"the configuration object to validate"`
}

// ListPluginsInput is the input for the list_plugins tool
// It has no required fields - plugins list is determined by the gateway's APISIX version and type
type ListPluginsInput struct{}

// RegisterSchemaTools registers all schema-related MCP tools
func RegisterSchemaTools(server *mcp.Server) {
	// get_resource_schema
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_resource_schema",
		Description: "Get the JSON Schema for a specific APISIX resource type. " +
			"Use this to understand the structure and validation rules.",
	}, getResourceSchemaHandler)

	// get_plugin_schema
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_plugin_schema",
		Description: "Get the JSON Schema for a specific APISIX plugin. Use this to understand plugin configuration options.",
	}, getPluginSchemaHandler)

	// validate_resource_config
	mcp.AddTool(server, &mcp.Tool{
		Name: "validate_resource_config",
		Description: "Validate a resource configuration against the APISIX schema. " +
			"Use this to check if a configuration is valid before creating or updating.",
	}, validateResourceConfigHandler)

	// list_plugins
	mcp.AddTool(server, &mcp.Tool{
		Name: "list_plugins",
		Description: "List available APISIX plugins for the current gateway. " +
			"The plugin list is determined by the gateway's APISIX version and type.",
	}, listPluginsHandler)
}

// getResourceSchemaHandler handles the get_resource_schema tool call
func getResourceSchemaHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetResourceSchemaInput,
) (*mcp.CallToolResult, any, error) {
	if input.APISIXVersion == "" || input.ResourceType == "" {
		return errorResult(fmt.Errorf("apisix_version and resource_type are required")), nil, nil
	}

	apisixVersion, err := parseAPISIXVersion(input.APISIXVersion)
	if err != nil {
		return errorResult(err), nil, nil
	}

	resourceType, err := parseResourceType(input.ResourceType)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Get schema
	schemaData := schema.GetResourceSchema(apisixVersion, resourceType.String())
	if schemaData == nil {
		return errorResult(fmt.Errorf("schema not found for resource type: %s", input.ResourceType)), nil, nil
	}

	return successResult(map[string]any{
		"apisix_version": input.APISIXVersion,
		"resource_type":  input.ResourceType,
		"schema":         schemaData,
	}), nil, nil
}

// getPluginSchemaHandler handles the get_plugin_schema tool call
func getPluginSchemaHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetPluginSchemaInput,
) (*mcp.CallToolResult, any, error) {
	if input.APISIXVersion == "" || input.PluginName == "" {
		return errorResult(fmt.Errorf("apisix_version and plugin_name are required")), nil, nil
	}

	apisixVersion, err := parseAPISIXVersion(input.APISIXVersion)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Apply default schema type
	schemaType := input.SchemaType
	if schemaType == "" {
		schemaType = "main"
	}

	// Get plugin schema
	schemaData := schema.GetPluginSchema(apisixVersion, input.PluginName, schemaType)
	if schemaData == nil {
		return errorResult(fmt.Errorf("schema not found for plugin: %s", input.PluginName)), nil, nil
	}

	return successResult(map[string]any{
		"apisix_version": input.APISIXVersion,
		"plugin_name":    input.PluginName,
		"schema_type":    schemaType,
		"schema":         schemaData,
	}), nil, nil
}

// validateResourceConfigHandler handles the validate_resource_config tool call
func validateResourceConfigHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ValidateResourceConfigInput,
) (*mcp.CallToolResult, any, error) {
	if input.APISIXVersion == "" || input.ResourceType == "" || input.Config == nil {
		return errorResult(fmt.Errorf("apisix_version, resource_type, and config are required")), nil, nil
	}

	apisixVersion, err := parseAPISIXVersion(input.APISIXVersion)
	if err != nil {
		return errorResult(err), nil, nil
	}

	resourceType, err := parseResourceType(input.ResourceType)
	if err != nil {
		return errorResult(err), nil, nil
	}

	// Marshal config to JSON bytes for validation
	configBytes, err := json.Marshal(input.Config)
	if err != nil {
		return errorResult(fmt.Errorf("failed to marshal config: %w", err)), nil, nil
	}

	// Validate config
	validator, err := schema.NewAPISIXSchemaValidator(apisixVersion, "main."+resourceType.String())
	if err != nil {
		return errorResult(fmt.Errorf("failed to create validator: %w", err)), nil, nil
	}

	validationErr := validator.Validate(configBytes)
	if validationErr != nil {
		// Validation failure is not a handler error - return success with valid=false
		//nolint:nilerr // validation error is intentionally returned in success response
		return successResult(map[string]any{
			"valid":          false,
			"message":        validationErr.Error(),
			"apisix_version": input.APISIXVersion,
			"resource_type":  input.ResourceType,
		}), nil, nil
	}

	return successResult(map[string]any{
		"valid":          true,
		"message":        "Configuration is valid",
		"apisix_version": input.APISIXVersion,
		"resource_type":  input.ResourceType,
	}), nil, nil
}

// listPluginsHandler handles the list_plugins tool call
func listPluginsHandler(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input ListPluginsInput,
) (*mcp.CallToolResult, any, error) {
	// Get gateway from context to determine version and type
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get gateway: %w", err)), nil, nil
	}

	// Use GetAPISIXVersionX() to properly convert version string to APISIXVersion type
	apisixVersion := gateway.GetAPISIXVersionX()
	apisixType := gateway.APISIXType

	// Get plugins list based on gateway's version and type
	plugins, err := biz.GetPluginsList(ctx, apisixVersion, apisixType)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get plugins: %w", err)), nil, nil
	}

	return successResult(map[string]any{
		"apisix_version": string(apisixVersion),
		"apisix_type":    apisixType,
		"plugins":        plugins,
		"count":          len(plugins),
	}), nil, nil
}
