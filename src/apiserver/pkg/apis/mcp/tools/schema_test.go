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
)

func TestGetResourceSchemaHandlerSupports317(t *testing.T) {
	result, _, err := getResourceSchemaHandler(
		context.Background(),
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
		context.Background(),
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
