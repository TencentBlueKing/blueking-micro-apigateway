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
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// RegisterSchemaTools registers all schema-related MCP tools
func RegisterSchemaTools(server *mcp.Server) {
	// get_resource_schema
	server.AddTool(&mcp.Tool{
		Name: "get_resource_schema",
		Description: "Get the JSON Schema for a specific APISIX resource type. " +
			"Use this to understand the structure and validation rules.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"apisix_version": map[string]any{
					"type":        "string",
					"description": "APISIX version for schema. " + APISIXVersionDescription(),
					"enum":        ValidAPISIXVersions,
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type to get schema for. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
			},
			"required": []string{"apisix_version", "resource_type"},
		},
	}, getResourceSchemaHandler)

	// get_plugin_schema
	server.AddTool(&mcp.Tool{
		Name:        "get_plugin_schema",
		Description: "Get the JSON Schema for a specific APISIX plugin. Use this to understand plugin configuration options.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"apisix_version": map[string]any{
					"type":        "string",
					"description": "APISIX version for schema. " + APISIXVersionDescription(),
					"enum":        ValidAPISIXVersions,
				},
				"plugin_name": map[string]any{
					"type":        "string",
					"description": "Name of the plugin to get schema for (e.g., 'limit-req', 'proxy-rewrite', 'jwt-auth')",
				},
				"schema_type": map[string]any{
					"type":        "string",
					"description": "Type of schema to retrieve (default: 'main')",
					"enum":        []string{"main", "consumer", "metadata"},
					"default":     "main",
				},
			},
			"required": []string{"apisix_version", "plugin_name"},
		},
	}, getPluginSchemaHandler)

	// validate_resource_config
	server.AddTool(&mcp.Tool{
		Name: "validate_resource_config",
		Description: "Validate a resource configuration against the APISIX schema. " +
			"Use this to check if a configuration is valid before creating or updating.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"apisix_version": map[string]any{
					"type":        "string",
					"description": "APISIX version to validate against. " + APISIXVersionDescription(),
					"enum":        ValidAPISIXVersions,
				},
				"resource_type": map[string]any{
					"type":        "string",
					"description": "Resource type. " + ResourceTypeDescription(),
					"enum":        ValidResourceTypes,
				},
				"config": map[string]any{
					"type":        "object",
					"description": "The configuration object to validate",
				},
			},
			"required": []string{"apisix_version", "resource_type", "config"},
		},
	}, validateResourceConfigHandler)

	// list_plugins
	server.AddTool(&mcp.Tool{
		Name: "list_plugins",
		Description: "List available APISIX plugins for the current gateway. " +
			"The plugin list is determined by the gateway's APISIX version and type.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, listPluginsHandler)
}

// getResourceSchemaHandler handles the get_resource_schema tool call
func getResourceSchemaHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	apisixVersionStr := getStringParamFromArgs(args, "apisix_version", "")
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")

	if apisixVersionStr == "" || resourceTypeStr == "" {
		return errorResult(fmt.Errorf("apisix_version and resource_type are required")), nil
	}

	apisixVersion, err := parseAPISIXVersion(apisixVersionStr)
	if err != nil {
		return errorResult(err), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	// Get schema
	schemaData := schema.GetResourceSchema(apisixVersion, resourceType.String())
	if schemaData == nil {
		return errorResult(fmt.Errorf("schema not found for resource type: %s", resourceTypeStr)), nil
	}

	return successResult(map[string]any{
		"apisix_version": apisixVersionStr,
		"resource_type":  resourceTypeStr,
		"schema":         schemaData,
	}), nil
}

// getPluginSchemaHandler handles the get_plugin_schema tool call
func getPluginSchemaHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	apisixVersionStr := getStringParamFromArgs(args, "apisix_version", "")
	pluginName := getStringParamFromArgs(args, "plugin_name", "")
	schemaType := getStringParamFromArgs(args, "schema_type", "main")

	if apisixVersionStr == "" || pluginName == "" {
		return errorResult(fmt.Errorf("apisix_version and plugin_name are required")), nil
	}

	apisixVersion, err := parseAPISIXVersion(apisixVersionStr)
	if err != nil {
		return errorResult(err), nil
	}

	// Get plugin schema
	schemaData := schema.GetPluginSchema(apisixVersion, pluginName, schemaType)
	if schemaData == nil {
		return errorResult(fmt.Errorf("schema not found for plugin: %s", pluginName)), nil
	}

	return successResult(map[string]any{
		"apisix_version": apisixVersionStr,
		"plugin_name":    pluginName,
		"schema_type":    schemaType,
		"schema":         schemaData,
	}), nil
}

// validateResourceConfigHandler handles the validate_resource_config tool call
func validateResourceConfigHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := parseArguments(req)
	apisixVersionStr := getStringParamFromArgs(args, "apisix_version", "")
	resourceTypeStr := getStringParamFromArgs(args, "resource_type", "")
	config, err := getObjectParamFromArgs(args, "config")
	if err != nil {
		return errorResult(err), nil
	}

	if apisixVersionStr == "" || resourceTypeStr == "" || len(config) == 0 {
		return errorResult(fmt.Errorf("apisix_version, resource_type, and config are required")), nil
	}

	apisixVersion, err := parseAPISIXVersion(apisixVersionStr)
	if err != nil {
		return errorResult(err), nil
	}

	resourceType, err := parseResourceType(resourceTypeStr)
	if err != nil {
		return errorResult(err), nil
	}

	// Validate config
	validator, err := schema.NewAPISIXSchemaValidator(apisixVersion, "main."+resourceType.String())
	if err != nil {
		return errorResult(fmt.Errorf("failed to create validator: %w", err)), nil
	}

	validationErr := validator.Validate(config)
	if validationErr != nil {
		// Validation failure is not a handler error - return success with valid=false
		//nolint:nilerr // validation error is intentionally returned in success response
		return successResult(map[string]any{
			"valid":          false,
			"message":        validationErr.Error(),
			"apisix_version": apisixVersionStr,
			"resource_type":  resourceTypeStr,
		}), nil
	}

	return successResult(map[string]any{
		"valid":          true,
		"message":        "Configuration is valid",
		"apisix_version": apisixVersionStr,
		"resource_type":  resourceTypeStr,
	}), nil
}

// listPluginsHandler handles the list_plugins tool call
func listPluginsHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get gateway from context to determine version and type
	gateway, err := getGatewayFromContext(ctx)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get gateway: %w", err)), nil
	}

	// Use GetAPISIXVersionX() to properly convert version string to APISIXVersion type
	apisixVersion := gateway.GetAPISIXVersionX()
	apisixType := gateway.APISIXType

	// Get plugins list based on gateway's version and type
	plugins, err := biz.GetPluginsList(ctx, apisixVersion, apisixType)
	if err != nil {
		return errorResult(fmt.Errorf("failed to get plugins: %w", err)), nil
	}

	return successResult(map[string]any{
		"apisix_version": string(apisixVersion),
		"apisix_type":    apisixType,
		"plugins":        plugins,
		"count":          len(plugins),
	}), nil
}
