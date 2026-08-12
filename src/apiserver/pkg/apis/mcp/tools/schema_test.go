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
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestGetResourceSchemaHandlerSupports317(t *testing.T) {
	result, _, err := getResourceSchemaHandler(
		gatewayContext("3.17.0"),
		nil,
		GetResourceSchemaInput{APISIXVersion: "3.17.X", ResourceType: "route"},
	)

	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, content.Text, `"apisix_version": "3.17.X"`)
	assert.Contains(t, content.Text, `"schema"`)
}

func TestGetPluginSchemaHandlerSupports317Plugin(t *testing.T) {
	result, _, err := getPluginSchemaHandler(
		gatewayContext("3.17.0"),
		nil,
		GetPluginSchemaInput{APISIXVersion: "3.17.X", PluginName: "oas-validator"},
	)

	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, content.Text, `"apisix_version": "3.17.X"`)
	assert.Contains(t, content.Text, `"plugin_name": "oas-validator"`)
}

func TestGetResourceSchemaHandlerDefaultsToGatewayVersion(t *testing.T) {
	result, _, err := getResourceSchemaHandler(
		gatewayContext("3.17.0"),
		nil,
		GetResourceSchemaInput{ResourceType: "route"},
	)

	require.NoError(t, err)
	require.False(t, result.IsError)
	content, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, content.Text, `"apisix_version": "3.17.X"`)
}

func TestValidateResourceConfigHandlerRejectsInvalid317PluginConfigs(t *testing.T) {
	util.InitEmbedDb()

	tests := []struct {
		name       string
		plugins    map[string]any
		wantErrMsg string
	}{
		{
			name: "missing required plugin field",
			plugins: map[string]any{
				"cas-auth": map[string]any{
					"idp_uri":          "https://cas.example.com",
					"cas_callback_uri": "/api/cas/callback",
					"logout_uri":       "https://cas.example.com/logout",
				},
			},
			wantErrMsg: "cookie",
		},
		{
			name: "unknown plugin",
			plugins: map[string]any{
				"unknown-plugin": map[string]any{},
			},
			wantErrMsg: "unknown-plugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := validateResourceConfigHandler(
				gatewayContext("3.17.0"),
				nil,
				ValidateResourceConfigInput{
					ResourceType: "route",
					Config: map[string]any{
						"uris":    []any{"/mcp-validation"},
						"plugins": tt.plugins,
					},
				},
			)

			require.NoError(t, err)
			require.False(t, result.IsError)
			content, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, content.Text, `"valid": false`)
			assert.Contains(t, content.Text, tt.wantErrMsg)
		})
	}
}

func TestSchemaHandlersRejectVersionDifferentFromGateway(t *testing.T) {
	ctx := gatewayContext("3.13.0")
	tests := []struct {
		name string
		call func() *mcp.CallToolResult
	}{
		{
			name: "resource schema",
			call: func() *mcp.CallToolResult {
				result, _, _ := getResourceSchemaHandler(
					ctx,
					nil,
					GetResourceSchemaInput{APISIXVersion: "3.17.X", ResourceType: "route"},
				)
				return result
			},
		},
		{
			name: "plugin schema",
			call: func() *mcp.CallToolResult {
				result, _, _ := getPluginSchemaHandler(
					ctx,
					nil,
					GetPluginSchemaInput{APISIXVersion: "3.17.X", PluginName: "oas-validator"},
				)
				return result
			},
		},
		{
			name: "resource validation",
			call: func() *mcp.CallToolResult {
				result, _, _ := validateResourceConfigHandler(
					ctx,
					nil,
					ValidateResourceConfigInput{
						APISIXVersion: "3.17.X",
						ResourceType:  "route",
						Config:        map[string]any{"uri": "/example"},
					},
				)
				return result
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.call()
			require.True(t, result.IsError)
			content, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, content.Text, "does not match current gateway version 3.13.X")
		})
	}
}

func gatewayContext(apisixVersion string) context.Context {
	return ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{
		APISIXVersion: apisixVersion,
		APISIXType:    constant.APISIXTypeAPISIX,
	})
}
