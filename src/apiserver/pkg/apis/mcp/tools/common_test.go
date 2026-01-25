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
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
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

// Note: TestParseArguments is not included because parseArguments depends on
// mcp.CallToolRequest internal structure. The core parsing logic is tested via
// the getXxxParamFromArgs helper function tests below, which is the pattern
// used throughout the codebase.

func TestGetIntParamFromArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       map[string]any
		paramName  string
		defaultVal int
		expected   int
	}{
		{
			name:       "float64 value",
			args:       map[string]any{"count": float64(10)},
			paramName:  "count",
			defaultVal: 0,
			expected:   10,
		},
		{
			name:       "int value",
			args:       map[string]any{"count": 20},
			paramName:  "count",
			defaultVal: 0,
			expected:   20,
		},
		{
			name:       "int64 value",
			args:       map[string]any{"count": int64(30)},
			paramName:  "count",
			defaultVal: 0,
			expected:   30,
		},
		{
			name:       "missing param uses default",
			args:       map[string]any{},
			paramName:  "count",
			defaultVal: 100,
			expected:   100,
		},
		{
			name:       "nil args uses default",
			args:       nil,
			paramName:  "count",
			defaultVal: 50,
			expected:   50,
		},
		{
			name:       "wrong type uses default",
			args:       map[string]any{"count": "not_a_number"},
			paramName:  "count",
			defaultVal: 25,
			expected:   25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := getIntParamFromArgs(tt.args, tt.paramName, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStringParamFromArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       map[string]any
		paramName  string
		defaultVal string
		expected   string
	}{
		{
			name:       "valid string value",
			args:       map[string]any{"name": "test-value"},
			paramName:  "name",
			defaultVal: "",
			expected:   "test-value",
		},
		{
			name:       "missing param uses default",
			args:       map[string]any{},
			paramName:  "name",
			defaultVal: "default",
			expected:   "default",
		},
		{
			name:       "nil args uses default",
			args:       nil,
			paramName:  "name",
			defaultVal: "default",
			expected:   "default",
		},
		{
			name:       "wrong type uses default",
			args:       map[string]any{"name": 123},
			paramName:  "name",
			defaultVal: "default",
			expected:   "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := getStringParamFromArgs(tt.args, tt.paramName, tt.defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStringArrayParamFromArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		args      map[string]any
		paramName string
		expected  []string
	}{
		{
			name:      "valid string array",
			args:      map[string]any{"ids": []any{"a", "b", "c"}},
			paramName: "ids",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "mixed array filters non-strings",
			args:      map[string]any{"ids": []any{"a", 123, "b"}},
			paramName: "ids",
			expected:  []string{"a", "b"},
		},
		{
			name:      "missing param returns nil",
			args:      map[string]any{},
			paramName: "ids",
			expected:  nil,
		},
		{
			name:      "nil args returns nil",
			args:      nil,
			paramName: "ids",
			expected:  nil,
		},
		{
			name:      "wrong type returns nil",
			args:      map[string]any{"ids": "not-an-array"},
			paramName: "ids",
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := getStringArrayParamFromArgs(tt.args, tt.paramName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetObjectParamFromArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          map[string]any
		paramName     string
		expectedNil   bool
		expectedError bool
	}{
		{
			name:          "valid object",
			args:          map[string]any{"config": map[string]any{"key": "value"}},
			paramName:     "config",
			expectedNil:   false,
			expectedError: false,
		},
		{
			name:          "missing param returns nil",
			args:          map[string]any{},
			paramName:     "config",
			expectedNil:   true,
			expectedError: false,
		},
		{
			name:          "nil args returns nil",
			args:          nil,
			paramName:     "config",
			expectedNil:   true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := getObjectParamFromArgs(tt.args, tt.paramName)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

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
