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

package constant_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestResourceSupportsNameField(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		expected     bool
		reason       string
	}{
		{
			name:         "route supports name field",
			resourceType: constant.Route,
			expected:     true,
			reason:       "route has name property in APISIX 3.11 and 3.13 schemas",
		},
		{
			name:         "service supports name field",
			resourceType: constant.Service,
			expected:     true,
			reason:       "service has name property in APISIX 3.11 and 3.13 schemas",
		},
		{
			name:         "upstream supports name field",
			resourceType: constant.Upstream,
			expected:     true,
			reason:       "upstream has name property in APISIX 3.11 and 3.13 schemas",
		},
		{
			name:         "plugin_config supports name field",
			resourceType: constant.PluginConfig,
			expected:     true,
			reason:       "plugin_config has name property in APISIX 3.11 and 3.13 schemas",
		},
		{
			name:         "consumer_group does NOT support name in 3.11",
			resourceType: constant.ConsumerGroup,
			expected:     false,
			reason:       "consumer_group doesn't have name in APISIX 3.11 (added in 3.13)",
		},
		{
			name:         "global_rule does NOT support name",
			resourceType: constant.GlobalRule,
			expected:     false,
			reason:       "global_rule doesn't have name property in any APISIX version",
		},
		{
			name:         "ssl does NOT support name",
			resourceType: constant.SSL,
			expected:     false,
			reason:       "ssl doesn't have name property in any APISIX version",
		},
		{
			name:         "stream_route does NOT support name in 3.11",
			resourceType: constant.StreamRoute,
			expected:     false,
			reason:       "stream_route doesn't have name in APISIX 3.11 (added in 3.13)",
		},
		{
			name:         "proto does NOT support name in 3.11",
			resourceType: constant.Proto,
			expected:     false,
			reason:       "proto doesn't have name in APISIX 3.11 (added in 3.13)",
		},
		{
			name:         "consumer does NOT support name",
			resourceType: constant.Consumer,
			expected:     false,
			reason:       "consumer uses username, not name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.ResourceSupportsNameField(tt.resourceType)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}

func TestResourceSupportsNameFieldForVersion(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		expected     bool
		reason       string
	}{
		// Resources that support name in all versions
		{
			name:         "route supports name in 3.11",
			resourceType: constant.Route,
			version:      constant.APISIXVersion311,
			expected:     true,
		},
		{
			name:         "route supports name in 3.13",
			resourceType: constant.Route,
			version:      constant.APISIXVersion313,
			expected:     true,
		},
		{
			name:         "plugin_config supports name in 3.11",
			resourceType: constant.PluginConfig,
			version:      constant.APISIXVersion311,
			expected:     true,
		},

		// Resources that added name in 3.13
		{
			name:         "consumer_group does NOT support name in 3.11",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion311,
			expected:     false,
			reason:       "consumer_group name not supported in 3.11",
		},
		{
			name:         "consumer_group DOES support name in 3.13",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			expected:     true,
			reason:       "consumer_group name added in 3.13",
		},
		{
			name:         "stream_route does NOT support name in 3.11",
			resourceType: constant.StreamRoute,
			version:      constant.APISIXVersion311,
			expected:     false,
		},
		{
			name:         "stream_route DOES support name in 3.13",
			resourceType: constant.StreamRoute,
			version:      constant.APISIXVersion313,
			expected:     true,
		},
		{
			name:         "proto does NOT support name in 3.11",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion311,
			expected:     false,
		},
		{
			name:         "proto DOES support name in 3.13",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion313,
			expected:     true,
		},

		// Resources that never support name
		{
			name:         "global_rule does NOT support name in 3.11",
			resourceType: constant.GlobalRule,
			version:      constant.APISIXVersion311,
			expected:     false,
		},
		{
			name:         "global_rule does NOT support name in 3.13",
			resourceType: constant.GlobalRule,
			version:      constant.APISIXVersion313,
			expected:     false,
		},
		{
			name:         "ssl does NOT support name in any version",
			resourceType: constant.SSL,
			version:      constant.APISIXVersion313,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.ResourceSupportsNameFieldForVersion(tt.resourceType, tt.version)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}

func TestShouldRemoveFieldBeforePublish(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		fieldName    string
		version      constant.APISIXVersion
		expected     bool
		reason       string
	}{
		// ID field removal
		{
			name:         "consumer id should be removed",
			resourceType: constant.Consumer,
			fieldName:    "id",
			version:      constant.APISIXVersion311,
			expected:     true,
			reason:       "consumer uses username as key, not id",
		},
		{
			name:         "consumer_group id should NOT be removed",
			resourceType: constant.ConsumerGroup,
			fieldName:    "id",
			version:      constant.APISIXVersion311,
			expected:     false,
			reason:       "consumer_group requires id in 3.11/3.13 schema",
		},
		{
			name:         "route id should NOT be removed",
			resourceType: constant.Route,
			fieldName:    "id",
			version:      constant.APISIXVersion311,
			expected:     false,
			reason:       "route includes id in config",
		},

		// Name field removal - version dependent
		{
			name:         "consumer_group name should be removed in 3.11",
			resourceType: constant.ConsumerGroup,
			fieldName:    "name",
			version:      constant.APISIXVersion311,
			expected:     true,
			reason:       "name not supported in 3.11",
		},
		{
			name:         "consumer_group name should NOT be removed in 3.13",
			resourceType: constant.ConsumerGroup,
			fieldName:    "name",
			version:      constant.APISIXVersion313,
			expected:     false,
			reason:       "name supported in 3.13",
		},
		{
			name:         "global_rule name should be removed in all versions",
			resourceType: constant.GlobalRule,
			fieldName:    "name",
			version:      constant.APISIXVersion313,
			expected:     true,
			reason:       "name never supported for global_rule",
		},
		{
			name:         "route name should NOT be removed",
			resourceType: constant.Route,
			fieldName:    "name",
			version:      constant.APISIXVersion311,
			expected:     false,
			reason:       "route supports name in all versions",
		},
		{
			name:         "plugin_config name should NOT be removed",
			resourceType: constant.PluginConfig,
			fieldName:    "name",
			version:      constant.APISIXVersion311,
			expected:     false,
			reason:       "plugin_config supports name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.ShouldRemoveFieldBeforeValidationOrPublish(
				tt.resourceType,
				tt.fieldName,
				tt.version,
			)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}

func TestResourceSupportsIDInConfig(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		expected     bool
	}{
		{
			name:         "consumer does NOT have id in config",
			resourceType: constant.Consumer,
			expected:     false,
		},
		{
			name:         "route has id in config",
			resourceType: constant.Route,
			expected:     true,
		},
		{
			name:         "consumer_group has id in config",
			resourceType: constant.ConsumerGroup,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.ResourceSupportsIDInConfig(tt.resourceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceRequiresIDInSchema(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		expected     bool
		reason       string
	}{
		{
			name:         "consumer_group requires id in schema",
			resourceType: constant.ConsumerGroup,
			expected:     true,
			reason:       "consumer_group.required includes 'id' in new schema",
		},
		{
			name:         "plugin_config requires id in schema",
			resourceType: constant.PluginConfig,
			expected:     true,
			reason:       "plugin_config.required includes 'id' in new schema",
		},
		{
			name:         "global_rule requires id in schema",
			resourceType: constant.GlobalRule,
			expected:     true,
			reason:       "global_rule.required includes 'id' in new schema",
		},
		{
			name:         "route does not require id in schema",
			resourceType: constant.Route,
			expected:     false,
			reason:       "route.required does not include 'id'",
		},
		{
			name:         "service does not require id in schema",
			resourceType: constant.Service,
			expected:     false,
			reason:       "service.required does not include 'id'",
		},
		{
			name:         "consumer does not require id in schema",
			resourceType: constant.Consumer,
			expected:     false,
			reason:       "consumer uses username as identifier, not id",
		},
		{
			name:         "ssl does not require id in schema",
			resourceType: constant.SSL,
			expected:     false,
			reason:       "ssl.required does not include 'id'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.ResourceRequiresIDInSchema(tt.resourceType)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}

func TestResourceUsesIDField(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		expected     bool
		reason       string
	}{
		{
			name:         "consumer uses username, not id",
			resourceType: constant.Consumer,
			expected:     false,
			reason:       "consumer is special - it uses 'username' as identifier",
		},
		{
			name:         "route uses id",
			resourceType: constant.Route,
			expected:     true,
			reason:       "route uses id as identifier",
		},
		{
			name:         "service uses id",
			resourceType: constant.Service,
			expected:     true,
			reason:       "service uses id as identifier",
		},
		{
			name:         "consumer_group uses id",
			resourceType: constant.ConsumerGroup,
			expected:     true,
			reason:       "consumer_group uses id as identifier",
		},
		{
			name:         "global_rule uses id",
			resourceType: constant.GlobalRule,
			expected:     true,
			reason:       "global_rule uses id as identifier",
		},
		{
			name:         "plugin_config uses id",
			resourceType: constant.PluginConfig,
			expected:     true,
			reason:       "plugin_config uses id as identifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.ResourceUsesIDField(tt.resourceType)
			assert.Equal(t, tt.expected, result, tt.reason)
		})
	}
}

func TestGetResourceSchemaCapability(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		expected     constant.ResourceSchemaCapability
	}{
		{
			name:         "route capabilities",
			resourceType: constant.Route,
			expected: constant.ResourceSchemaCapability{
				SupportsNameField:  true,
				RequiresIDInSchema: false,
				UsesIDField:        true,
			},
		},
		{
			name:         "consumer capabilities",
			resourceType: constant.Consumer,
			expected: constant.ResourceSchemaCapability{
				SupportsNameField:  false,
				RequiresIDInSchema: false,
				UsesIDField:        false, // Uses username instead
			},
		},
		{
			name:         "consumer_group capabilities",
			resourceType: constant.ConsumerGroup,
			expected: constant.ResourceSchemaCapability{
				SupportsNameField:  false, // Not in 3.11
				RequiresIDInSchema: true,
				UsesIDField:        true,
			},
		},
		{
			name:         "global_rule capabilities",
			resourceType: constant.GlobalRule,
			expected: constant.ResourceSchemaCapability{
				SupportsNameField:  false, // Never supported
				RequiresIDInSchema: true,
				UsesIDField:        true,
			},
		},
		{
			name:         "ssl capabilities",
			resourceType: constant.SSL,
			expected: constant.ResourceSchemaCapability{
				SupportsNameField:  false, // Never supported
				RequiresIDInSchema: false,
				UsesIDField:        true,
			},
		},
		{
			name:         "plugin_config capabilities",
			resourceType: constant.PluginConfig,
			expected: constant.ResourceSchemaCapability{
				SupportsNameField:  true,
				RequiresIDInSchema: true,
				UsesIDField:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constant.GetResourceSchemaCapability(tt.resourceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestResourceSchemaCapabilityConsistency verifies that all resources have consistent capability settings
func TestResourceSchemaCapabilityConsistency(t *testing.T) {
	allResources := []constant.APISIXResource{
		constant.Route,
		constant.Service,
		constant.Upstream,
		constant.PluginConfig,
		constant.PluginMetadata,
		constant.Consumer,
		constant.ConsumerGroup,
		constant.GlobalRule,
		constant.Proto,
		constant.SSL,
		constant.StreamRoute,
	}

	for _, resourceType := range allResources {
		t.Run(string(resourceType), func(t *testing.T) {
			cap := constant.GetResourceSchemaCapability(resourceType)

			// All resources should have at least one capability defined
			// Exception: Consumer uses username instead of id, so all three can be false
			hasAnyCapability := cap.SupportsNameField || cap.RequiresIDInSchema || cap.UsesIDField
			if resourceType != constant.Consumer {
				assert.True(t, hasAnyCapability,
					"Resource %s should have at least one capability defined", resourceType)
			}

			// Consumer is the only resource that doesn't use ID field
			if resourceType == constant.Consumer {
				assert.False(t, cap.UsesIDField,
					"Consumer should not use ID field (uses username)")
			} else {
				assert.True(t, cap.UsesIDField,
					"Resource %s should use ID field", resourceType)
			}

			// If a resource requires ID in schema, it should also use ID field
			if cap.RequiresIDInSchema {
				assert.True(t, cap.UsesIDField,
					"Resource %s requires ID in schema but doesn't use ID field", resourceType)
			}
		})
	}
}
