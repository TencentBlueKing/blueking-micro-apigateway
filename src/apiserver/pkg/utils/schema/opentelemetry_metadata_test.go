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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestOpentelemetryMetadataCatalog(t *testing.T) {
	for _, version := range []constant.APISIXVersion{
		constant.APISIXVersion313, constant.APISIXVersion317, constant.APISIXVersion318,
	} {
		for _, apisixType := range []string{constant.APISIXTypeAPISIX, constant.APISIXTypeBKAPISIX} {
			t.Run(fmt.Sprintf("%s/%s", version, apisixType), func(t *testing.T) {
				plugins, err := GetPlugins(apisixType, version)
				require.NoError(t, err)
				var found *Plugin
				for _, plugin := range plugins {
					if plugin.Name == "opentelemetry" {
						require.Nil(t, found, "opentelemetry must appear only once")
						found = plugin
					}
				}
				require.NotNil(t, found)
				require.NotEmpty(t, found.MetadataExample)
				for _, schemaType := range []string{"metadata", "metadata_schema"} {
					require.NotNil(t, GetPluginSchema(version, "opentelemetry", schemaType))
				}
				for scope, example := range map[string]map[string]any{
					"metadata": found.MetadataExample, "": found.Example,
				} {
					schema, err := gojsonschema.NewSchema(
						gojsonschema.NewGoLoader(
							GetPluginSchema(version, "opentelemetry", scope),
						),
					)
					require.NoError(t, err)
					result, err := schema.Validate(gojsonschema.NewGoLoader(example))
					require.NoError(t, err)
					assert.True(t, result.Valid(), "scope %s: %v", scope, result.Errors())
				}
			})
		}
	}
	for _, version := range []constant.APISIXVersion{
		constant.APISIXVersion32, constant.APISIXVersion33, constant.APISIXVersion311,
	} {
		t.Run(string(version)+"/legacy_plugin_attr", func(t *testing.T) {
			assert.Nil(t, GetPluginSchema(version, "opentelemetry", "metadata"))
			plugins, err := GetPlugins(constant.APISIXTypeAPISIX, version)
			require.NoError(t, err)
			for _, plugin := range plugins {
				if plugin.Name == "opentelemetry" {
					assert.Empty(t, plugin.MetadataExample)
				}
			}
		})
	}
}

func TestOpentelemetryMetadataValidation(t *testing.T) {
	cases := []struct {
		name   string
		config string
		valid  bool
	}{
		{"defaults", `{"id":"opentelemetry"}`, true},
		{"documented config", `{
			"id":"opentelemetry", "trace_id_source":"x-request-id",
			"resource":{"service.name":"APISIX", "enabled":true, "instance":1},
			"collector":{"address":"collector.example.com:4318", "request_timeout":3,
				"request_headers":{"Authorization":"example-token", "enabled":false, "instance":1}},
			"batch_span_processor":{"drop_on_queue_full":false, "max_queue_size":1024,
				"batch_timeout":2, "inactive_timeout":1, "max_export_batch_size":16},
			"set_ngx_var":false
		}`, true},
		{"invalid trace source", `{"id":"opentelemetry", "trace_id_source":"other"}`, false},
		{"resource object value", `{"id":"opentelemetry", "resource":{"service.name":{}}}`, false},
		{"resource array value", `{"id":"opentelemetry", "resource":{"service.name":[]}}`, false},
		{"header object value", `{"id":"opentelemetry", "collector":{"request_headers":{"header":{}}}}`, false},
		{"header array value", `{"id":"opentelemetry", "collector":{"request_headers":{"header":[]}}}`, false},
		{"fractional timeout", `{"id":"opentelemetry", "collector":{"request_timeout":1.5}}`, false},
		{
			"invalid batch size",
			`{"id":"opentelemetry", "batch_span_processor":{"max_queue_size":"1024"}}`,
			false,
		},
		{"invalid nginx variable flag", `{"id":"opentelemetry", "set_ngx_var":"false"}`, false},
	}
	for _, version := range []constant.APISIXVersion{
		constant.APISIXVersion313, constant.APISIXVersion317, constant.APISIXVersion318,
	} {
		for _, dataType := range []constant.DataType{constant.DATABASE, constant.ETCD} {
			for _, tc := range cases {
				t.Run(fmt.Sprintf("%s/%s/%s", version, dataType, tc.name), func(t *testing.T) {
					validator, err := NewAPISIXJsonSchemaValidator(
						version, constant.PluginMetadata, "main.plugin_metadata", nil, dataType,
					)
					require.NoError(t, err)
					err = validator.Validate(json.RawMessage(tc.config))
					if tc.valid {
						assert.NoError(t, err)
					} else {
						require.Error(t, err)
						assert.NotContains(t, err.Error(), "未找到 schema")
					}
				})
			}
		}
	}
}
