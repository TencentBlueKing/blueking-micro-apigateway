/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestGetPlugins(t *testing.T) {
	tests := []struct {
		name       string
		apisixType string
		version    constant.APISIXVersion
		shouldFail bool
	}{
		{
			name:       "APISIX 3.13",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion313,
			shouldFail: false,
		},
		{
			name:       "APISIX 3.11",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion311,
			shouldFail: false,
		},
		{
			name:       "APISIX 3.2",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion32,
			shouldFail: false,
		},
		{
			name:       "APISIX 3.3",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion33,
			shouldFail: false,
		},
		{
			name:       "Invalid Version",
			apisixType: constant.APISIXTypeAPISIX,
			version:    "invalid_version",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugins, err := GetPlugins(tt.apisixType, tt.version)

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, plugins)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugins)

				// 验证插件的基本结构
				for _, v := range plugins {
					assert.NotEmpty(t, v.Name)
					assert.NotEmpty(t, v.Type)
					assert.NotNil(t, v.Example)
				}
			}
		})
	}
}

func TestPluginExamplesMatchSchema(t *testing.T) {
	versions := []constant.APISIXVersion{
		constant.APISIXVersion311,
		constant.APISIXVersion313,
	}

	type exampleCase struct {
		schemaType string
		scope      string
		example    map[string]any
	}

	for _, version := range versions {
		plugins, err := GetPlugins(constant.APISIXTypeAPISIX, version)
		if !assert.NoError(t, err) {
			continue
		}

		for _, plugin := range plugins {
			mainSchemaType := ""
			if plugin.ProxyType == "stream" {
				mainSchemaType = "stream"
			}

			cases := []exampleCase{
				{
					schemaType: mainSchemaType,
					scope:      "main",
					example:    plugin.Example,
				},
				{
					schemaType: "consumer",
					scope:      "consumer",
					example:    plugin.ConsumerExample,
				},
				{
					schemaType: "metadata",
					scope:      "metadata",
					example:    plugin.MetadataExample,
				},
			}

			for _, tc := range cases {
				if tc.example == nil {
					continue
				}

				testName := fmt.Sprintf("%s/%s/%s", version, plugin.Name, tc.scope)
				t.Run(testName, func(t *testing.T) {
					schemaValue := GetPluginSchema(version, plugin.Name, tc.schemaType)
					if schemaValue == nil {
						t.Fatalf(
							"schema not found for plugin %q scope %q in version %s",
							plugin.Name,
							tc.scope,
							version,
						)
						return
					}

					schemaBytes, err := json.Marshal(schemaValue)
					assert.NoError(t, err)

					s, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))
					assert.NoError(t, err)

					result, err := s.Validate(gojsonschema.NewGoLoader(tc.example))
					assert.NoError(t, err)
					assert.Truef(
						t,
						result.Valid(),
						"schema validation errors: %s",
						strings.Join(flattenSchemaErrors(result.Errors()), "; "),
					)
				})
			}
		}
	}
}

func TestOpenFunctionAuthorizationSchemaShape(t *testing.T) {
	schemaValue := GetPluginSchema(constant.APISIXVersion313, "openfunction", "")

	schemaMap, ok := schemaValue.(map[string]any)
	if !assert.True(t, ok) {
		return
	}

	properties, ok := schemaMap["properties"].(map[string]any)
	if !assert.True(t, ok) {
		return
	}

	authorization, ok := properties["authorization"].(map[string]any)
	if !assert.True(t, ok) {
		return
	}

	assert.Equal(t, "object", authorization["type"])
	assert.NotContains(t, authorization, "service_token")

	authProperties, ok := authorization["properties"].(map[string]any)
	if !assert.True(t, ok) {
		return
	}

	serviceToken, ok := authProperties["service_token"].(map[string]any)
	if !assert.True(t, ok) {
		return
	}

	assert.Equal(t, "string", serviceToken["type"])
}

func flattenSchemaErrors(errors []gojsonschema.ResultError) []string {
	if len(errors) == 0 {
		return nil
	}

	messages := make([]string, 0, len(errors))
	for _, err := range errors {
		messages = append(messages, err.String())
	}
	return messages
}
