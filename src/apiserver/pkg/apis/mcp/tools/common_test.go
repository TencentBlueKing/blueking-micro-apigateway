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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestParseResourceType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		resourceType  string
		expectedType  constant.APISIXResource
		expectedError bool
	}{
		{
			name:          "valid route type",
			resourceType:  "route",
			expectedType:  constant.Route,
			expectedError: false,
		},
		{
			name:          "valid service type",
			resourceType:  "service",
			expectedType:  constant.Service,
			expectedError: false,
		},
		{
			name:          "valid upstream type",
			resourceType:  "upstream",
			expectedType:  constant.Upstream,
			expectedError: false,
		},
		{
			name:          "valid consumer type",
			resourceType:  "consumer",
			expectedType:  constant.Consumer,
			expectedError: false,
		},
		{
			name:          "valid consumer_group type",
			resourceType:  "consumer_group",
			expectedType:  constant.ConsumerGroup,
			expectedError: false,
		},
		{
			name:          "valid plugin_config type",
			resourceType:  "plugin_config",
			expectedType:  constant.PluginConfig,
			expectedError: false,
		},
		{
			name:          "valid global_rule type",
			resourceType:  "global_rule",
			expectedType:  constant.GlobalRule,
			expectedError: false,
		},
		{
			name:          "valid plugin_metadata type",
			resourceType:  "plugin_metadata",
			expectedType:  constant.PluginMetadata,
			expectedError: false,
		},
		{
			name:          "valid proto type",
			resourceType:  "proto",
			expectedType:  constant.Proto,
			expectedError: false,
		},
		{
			name:          "valid ssl type",
			resourceType:  "ssl",
			expectedType:  constant.SSL,
			expectedError: false,
		},
		{
			name:          "valid stream_route type",
			resourceType:  "stream_route",
			expectedType:  constant.StreamRoute,
			expectedError: false,
		},
		{
			name:          "invalid resource type",
			resourceType:  "invalid_type",
			expectedType:  "",
			expectedError: true,
		},
		{
			name:          "empty resource type",
			resourceType:  "",
			expectedType:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := parseResourceType(tt.resourceType)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid resource type")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, result)
			}
		})
	}
}

func TestParseAPISIXVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		version         string
		expectedVersion constant.APISIXVersion
		expectedError   bool
	}{
		{
			name:            "valid version 3.11",
			version:         "3.11.X",
			expectedVersion: constant.APISIXVersion311,
			expectedError:   false,
		},
		{
			name:            "valid version 3.13",
			version:         "3.13.X",
			expectedVersion: constant.APISIXVersion313,
			expectedError:   false,
		},
		{
			name:            "invalid version",
			version:         "3.0.X",
			expectedVersion: "",
			expectedError:   true,
		},
		{
			name:            "empty version",
			version:         "",
			expectedVersion: "",
			expectedError:   true,
		},
		{
			name:            "invalid version format",
			version:         "invalid",
			expectedVersion: "",
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := parseAPISIXVersion(tt.version)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid APISIX version")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, result)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple map",
			input:    map[string]string{"key": "value"},
			expected: "{\n  \"key\": \"value\"\n}",
		},
		{
			name:     "simple struct",
			input:    struct{ Name string }{Name: "test"},
			expected: "{\n  \"Name\": \"test\"\n}",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "null",
		},
		{
			name:     "slice",
			input:    []string{"a", "b"},
			expected: "[\n  \"a\",\n  \"b\"\n]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: TestGetGatewayFromRequestRejectsMismatchedToken was removed because
// gateway_id validation is now done in the MCP auth middleware (mcp_auth.go).
// The middleware validates that path gateway_id matches the token's gateway_id.

func TestGetGatewayFromContextReturnsGateway(t *testing.T) {
	util.InitEmbedDb()

	ctx := context.Background()
	gateway := &model.Gateway{
		Name:          "mcp-tools-gateway",
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := biz.CreateGateway(ctx, gateway)
	assert.NoError(t, err)
	assert.Greater(t, gateway.ID, 0)

	// Set gateway info in context (as middleware would do)
	ctxWithGateway := ginx.SetGatewayInfoToContext(ctx, gateway)

	gotGateway, err := getGatewayFromContext(ctxWithGateway)
	assert.NoError(t, err)
	assert.NotNil(t, gotGateway)
	assert.Equal(t, gateway.ID, gotGateway.ID)
}

func TestGetGatewayFromContextReturnsErrorWhenMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, err := getGatewayFromContext(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gateway not found in context")
}

func TestSuccessResult(t *testing.T) {
	t.Parallel()

	data := map[string]string{"status": "ok"}
	result := successResult(data)

	assert.False(t, result.IsError)
	assert.Len(t, result.Content, 1)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok)
	assert.Contains(t, textContent.Text, "status")
	assert.Contains(t, textContent.Text, "ok")
}

func TestErrorResult(t *testing.T) {
	t.Parallel()

	err := assert.AnError
	result := errorResult(err)

	assert.True(t, result.IsError)
	assert.Len(t, result.Content, 1)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	assert.True(t, ok)
	assert.Equal(t, err.Error(), textContent.Text)
}

// Note: getXxxParamFromArgs helper functions were removed in favor of typed inputs
// with the MCP SDK's generic AddTool[In, Out] function, which auto-parses inputs.

func TestResourceTypeDescription(t *testing.T) {
	t.Parallel()

	desc := ResourceTypeDescription()
	assert.Contains(t, desc, "One of:")
	assert.Contains(t, desc, "route")
	assert.Contains(t, desc, "service")
	assert.Contains(t, desc, "upstream")
}

func TestStatusDescription(t *testing.T) {
	t.Parallel()

	desc := StatusDescription()
	assert.Contains(t, desc, "One of:")
	assert.Contains(t, desc, "create_draft")
	assert.Contains(t, desc, "update_draft")
	assert.Contains(t, desc, "delete_draft")
	assert.Contains(t, desc, "success")
}

func TestAPISIXVersionDescription(t *testing.T) {
	t.Parallel()

	desc := APISIXVersionDescription()
	assert.Contains(t, desc, "One of:")
	assert.Contains(t, desc, "3.11.X")
	assert.Contains(t, desc, "3.13.X")
}

func TestValidResourceTypes(t *testing.T) {
	t.Parallel()

	// Verify all 11 resource types are present
	assert.Len(t, ValidResourceTypes, 11)
	assert.Contains(t, ValidResourceTypes, "route")
	assert.Contains(t, ValidResourceTypes, "service")
	assert.Contains(t, ValidResourceTypes, "upstream")
	assert.Contains(t, ValidResourceTypes, "consumer")
	assert.Contains(t, ValidResourceTypes, "consumer_group")
	assert.Contains(t, ValidResourceTypes, "plugin_config")
	assert.Contains(t, ValidResourceTypes, "global_rule")
	assert.Contains(t, ValidResourceTypes, "plugin_metadata")
	assert.Contains(t, ValidResourceTypes, "proto")
	assert.Contains(t, ValidResourceTypes, "ssl")
	assert.Contains(t, ValidResourceTypes, "stream_route")
}

func TestValidResourceStatuses(t *testing.T) {
	t.Parallel()

	// Verify all 4 statuses are present
	assert.Len(t, ValidResourceStatuses, 4)
	assert.Contains(t, ValidResourceStatuses, "create_draft")
	assert.Contains(t, ValidResourceStatuses, "update_draft")
	assert.Contains(t, ValidResourceStatuses, "delete_draft")
	assert.Contains(t, ValidResourceStatuses, "success")
}

func TestValidAPISIXVersions(t *testing.T) {
	t.Parallel()

	// Verify all supported versions are present
	assert.Len(t, ValidAPISIXVersions, 2)
	assert.Contains(t, ValidAPISIXVersions, "3.11.X")
	assert.Contains(t, ValidAPISIXVersions, "3.13.X")
}
