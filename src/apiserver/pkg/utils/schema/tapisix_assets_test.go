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

package schema

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

var expectedTAPISIXPluginNames = []string{
	"tof-auth",
	"zhiyan-log",
	"trpc-transcode",
	"trpc-canary",
	"tauth-proxy",
	"tauth-auth",
	"t-header-filter",
	"sign-auth",
	"response-wrapper",
	"polaris-limit",
	"polaris-circuit-breaker",
	"pangu-wolf-rbac",
	"pangu-authz",
	"pangu-authn",
	"ip-city",
	"log-replay",
	"galileo-sampler",
	"galileo-metrics",
	"downgrade-cache",
	"cls-logger",
}

func enableTAPISIXForTest(t *testing.T) {
	t.Helper()
	previousConfig := config.G
	config.G = &config.Config{Service: config.ServiceConfig{EnableTAPISIX: true}}
	t.Cleanup(func() {
		config.G = previousConfig
	})
}

func TestGetPluginsTAPISIXVisibility(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		wantTAPISIX bool
	}{
		{
			name:        "hidden from plugin catalogs by default",
			enabled:     false,
			wantTAPISIX: false,
		},
		{
			name:        "included in customized plugin catalogs when enabled",
			enabled:     true,
			wantTAPISIX: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previousConfig := config.G
			config.G = &config.Config{Service: config.ServiceConfig{EnableTAPISIX: tt.enabled}}
			t.Cleanup(func() {
				config.G = previousConfig
			})

			for _, apisixType := range []string{
				constant.APISIXTypeAPISIX,
				constant.APISIXTypeTAPISIX,
				constant.APISIXTypeBKAPISIX,
			} {
				plugins, err := GetPlugins(apisixType, constant.APISIXVersion318)
				require.NoError(t, err)
				names := pluginNames(plugins)
				for _, name := range expectedTAPISIXPluginNames {
					wantPlugin := tt.wantTAPISIX && apisixType != constant.APISIXTypeAPISIX
					if wantPlugin {
						assert.Contains(t, names, name)
						continue
					}
					assert.NotContains(t, names, name)
				}
			}
		})
	}
}

func TestTAPISIXAssetsAcrossCompositionVersions(t *testing.T) {
	versions := []constant.APISIXVersion{
		constant.APISIXVersion33,
		constant.APISIXVersion311,
		constant.APISIXVersion313,
		constant.APISIXVersion317,
		constant.APISIXVersion318,
	}

	for _, version := range versions {
		t.Run(string(version), func(t *testing.T) {
			rawCatalog, ok := versionTAPISIXPluginMap[version]
			require.True(t, ok)

			var plugins []*Plugin
			require.NoError(t, json.Unmarshal(rawCatalog, &plugins))
			assert.ElementsMatch(t, expectedTAPISIXPluginNames, pluginNames(plugins))

			schemaRoot, ok := tapisixPluginSchemaVersionMap[version]
			require.True(t, ok)
			schemaNames := make([]string, 0, len(schemaRoot.Get("plugins").Map()))
			for name := range schemaRoot.Get("plugins").Map() {
				schemaNames = append(schemaNames, name)
			}
			assert.ElementsMatch(t, expectedTAPISIXPluginNames, schemaNames)

			for _, plugin := range plugins {
				validateTAPISIXExample(t, version, plugin.Name, "main", "", plugin.Example)
				if plugin.ConsumerExample != nil {
					validateTAPISIXExample(
						t,
						version,
						plugin.Name,
						"consumer",
						"consumer",
						plugin.ConsumerExample,
					)
				}
				if plugin.MetadataExample != nil {
					validateTAPISIXExample(
						t,
						version,
						plugin.Name,
						"metadata",
						"metadata",
						plugin.MetadataExample,
					)
				}
			}
		})
	}
}

func TestTAPISIXConsumerSchemas(t *testing.T) {
	tests := []struct {
		version  constant.APISIXVersion
		plugin   string
		required []any
	}{
		{constant.APISIXVersion313, "tof-auth", []any{"paas_id", "token"}},
		{constant.APISIXVersion313, "tauth-auth", []any{"tauth_user"}},
		{constant.APISIXVersion317, "tof-auth", []any{"paas_id", "token"}},
		{constant.APISIXVersion317, "tauth-auth", []any{"tauth_user"}},
		{constant.APISIXVersion318, "tof-auth", []any{"paas_id", "token"}},
		{constant.APISIXVersion318, "tauth-auth", []any{"tauth_user"}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", tt.version, tt.plugin), func(t *testing.T) {
			schemaValue := GetPluginSchema(tt.version, tt.plugin, "consumer_schema")
			schemaMap, ok := schemaValue.(map[string]any)
			require.True(t, ok)
			required, ok := schemaMap["required"].([]any)
			require.True(t, ok, "consumer schema must declare required fields")
			assert.ElementsMatch(t, tt.required, required)

			schemaBytes, err := json.Marshal(schemaValue)
			require.NoError(t, err)
			compiled, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))
			require.NoError(t, err)
			result, err := compiled.Validate(gojsonschema.NewGoLoader(map[string]any{}))
			require.NoError(t, err)
			assert.False(t, result.Valid(), "empty consumer configuration must be rejected")
		})
	}
}

func validateTAPISIXExample(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	scope string,
	schemaType string,
	example map[string]any,
) {
	t.Helper()
	schemaValue := GetPluginSchema(version, pluginName, schemaType)
	require.NotNil(t, schemaValue, "%s/%s/%s schema not found", version, pluginName, scope)

	schemaBytes, err := json.Marshal(schemaValue)
	require.NoError(t, err)
	compiled, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))
	require.NoError(t, err, "%s/%s/%s", version, pluginName, scope)
	result, err := compiled.Validate(gojsonschema.NewGoLoader(example))
	require.NoError(t, err, "%s/%s/%s", version, pluginName, scope)
	assert.Truef(
		t,
		result.Valid(),
		"%s/%s/%s: %s",
		version,
		pluginName,
		scope,
		fmt.Sprint(result.Errors()),
	)
}
