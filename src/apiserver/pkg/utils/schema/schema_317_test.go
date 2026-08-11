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

func Test317PluginComposition(t *testing.T) {
	official, err := GetPlugins(constant.APISIXTypeAPISIX, constant.APISIXVersion317)
	require.NoError(t, err)
	tapisix, err := GetPlugins(constant.APISIXTypeTAPISIX, constant.APISIXVersion317)
	require.NoError(t, err)
	bk, err := GetPlugins(constant.APISIXTypeBKAPISIX, constant.APISIXVersion317)
	require.NoError(t, err)

	assert.Equal(t, pluginNames(official), pluginNames(tapisix))
	for _, name := range []string{
		"proxy-buffering",
		"saml-auth",
		"dingtalk-auth",
		"feishu-auth",
		"acl",
		"data-mask",
		"ai-aliyun-content-moderation",
		"graphql-proxy-cache",
		"graphql-limit-count",
		"traffic-label",
		"oas-validator",
	} {
		assert.Contains(t, pluginNames(official), name)
	}
	assert.Len(t, bk, len(official)+7)
	for _, name := range []string{
		"bk-break-recursive-call",
		"bk-delete-cookie",
		"bk-echo",
		"bk-header-rewrite",
		"bk-jwt",
		"bk-login-required",
		"bk-traffic-label",
	} {
		assert.Contains(t, pluginNames(bk), name)
		assert.NotContains(t, pluginNames(official), name)
	}
}

func Test317EmptyTAPISIXAssets(t *testing.T) {
	catalog, ok := versionTAPISIXPluginMap[constant.APISIXVersion317]
	require.True(t, ok)
	var plugins []Plugin
	require.NoError(t, json.Unmarshal(catalog, &plugins))
	assert.Empty(t, plugins)

	schemaValue, ok := tapisixPluginSchemaVersionMap[constant.APISIXVersion317]
	require.True(t, ok)
	assert.True(t, schemaValue.Get("plugins").IsObject())
	assert.Empty(t, schemaValue.Get("plugins").Map())

	_, ok = versionBkAPISIXPluginMap[constant.APISIXVersion317]
	assert.True(t, ok)
	_, ok = bkAPISIXPluginSchemaVersionMap[constant.APISIXVersion317]
	assert.True(t, ok)
}

func Test317CatalogSchemasAndExamples(t *testing.T) {
	catalogs := []struct {
		name string
		raw  []byte
	}{
		{name: "official", raw: rawPluginV317},
		{name: "bk", raw: rawBkAPISIXPluginV317},
		{name: "tapisix", raw: rawTAPISIXPluginV317},
	}
	seen := make(map[string]string)

	for _, catalog := range catalogs {
		var plugins []*Plugin
		require.NoError(t, json.Unmarshal(catalog.raw, &plugins), catalog.name)

		for _, plugin := range plugins {
			if previousSource, exists := seen[plugin.Name]; exists {
				t.Errorf(
					"plugin %q exists in both %s and %s catalogs",
					plugin.Name,
					previousSource,
					catalog.name,
				)
			}
			seen[plugin.Name] = catalog.name

			mainSchemaType := ""
			if plugin.ProxyType == "stream" {
				mainSchemaType = "stream"
			}
			validate317PluginExample(t, plugin.Name, "main", mainSchemaType, plugin.Example)
			if plugin.ConsumerExample != nil {
				validate317PluginExample(t, plugin.Name, "consumer", "consumer", plugin.ConsumerExample)
			}
			if plugin.MetadataExample != nil {
				validate317PluginExample(t, plugin.Name, "metadata", "metadata", plugin.MetadataExample)
			}
		}
	}
}

func Test317AllSchemaNodesCompile(t *testing.T) {
	for _, resource := range constant.ResourceTypeList {
		compile317Schema(
			t,
			"main."+resource.String(),
			GetResourceSchema(constant.APISIXVersion317, resource.String()),
		)
	}

	sources := []struct {
		name string
		root any
	}{
		{name: "official", root: schemaVersionMap[constant.APISIXVersion317].Value()},
		{name: "bk", root: bkAPISIXPluginSchemaVersionMap[constant.APISIXVersion317].Value()},
		{name: "tapisix", root: tapisixPluginSchemaVersionMap[constant.APISIXVersion317].Value()},
	}
	for _, source := range sources {
		root := requireStringMap(t, source.root, source.name)
		plugins, _ := root["plugins"].(map[string]any)
		for pluginName, pluginValue := range plugins {
			plugin := requireStringMap(t, pluginValue, source.name+".plugins."+pluginName)
			for _, scope := range []string{"schema", "consumer_schema", "metadata_schema"} {
				if schemaValue, ok := plugin[scope]; ok {
					compile317Schema(
						t,
						fmt.Sprintf("%s.plugins.%s.%s", source.name, pluginName, scope),
						normalizePluginSchema(schemaValue),
					)
				}
			}
		}

		streamPlugins, _ := root["stream_plugins"].(map[string]any)
		for pluginName, pluginValue := range streamPlugins {
			plugin := requireStringMap(t, pluginValue, source.name+".stream_plugins."+pluginName)
			if schemaValue, ok := plugin["schema"]; ok {
				compile317Schema(t, source.name+".stream_plugins."+pluginName+".schema", schemaValue)
			}
		}
	}
}

func Test317RepresentativeSchemaDifferences(t *testing.T) {
	assert.Nil(t, schemaProperty(t, constant.APISIXVersion313, "jwt-auth", "", "claims_to_verify"))
	assert.NotNil(t, schemaProperty(t, constant.APISIXVersion317, "jwt-auth", "", "claims_to_verify"))
	assert.Nil(t, schemaProperty(t, constant.APISIXVersion313, "jwt-auth", "", "realm"))
	assert.NotNil(t, schemaProperty(t, constant.APISIXVersion317, "jwt-auth", "", "realm"))

	assert.NotContains(t, jwtAlgorithms(t, constant.APISIXVersion313), "EdDSA")
	assert.Contains(t, jwtAlgorithms(t, constant.APISIXVersion317), "EdDSA")

	assert.Nil(t, schemaProperty(t, constant.APISIXVersion313, "hmac-auth", "", "max_req_body_size"))
	assert.Equal(
		t,
		float64(1),
		schemaNumberKeyword(t, constant.APISIXVersion317, "hmac-auth", "", "max_req_body_size", "minimum"),
	)
	assert.NotNil(t, schemaProperty(t, constant.APISIXVersion317, "hmac-auth", "", "realm"))

	assert.NotContains(t, schemaRequired(t, constant.APISIXVersion313, "cas-auth", ""), "cookie")
	assert.Contains(t, schemaRequired(t, constant.APISIXVersion317, "cas-auth", ""), "cookie")
	assert.Equal(t, float64(32), schemaNestedNumberKeyword(
		t,
		constant.APISIXVersion317,
		"cas-auth",
		"",
		[]string{"properties", "cookie", "properties", "secret", "minLength"},
	))

	assert.Nil(t, schemaProperty(t, constant.APISIXVersion313, "batch-requests", "metadata", "max_pipeline_items"))
	assert.NotNil(
		t,
		schemaProperty(t, constant.APISIXVersion317, "batch-requests", "metadata", "max_pipeline_items"),
	)

	assert.NotContains(
		t,
		schemaEnum(t, constant.APISIXVersion313, "ai-rate-limiting", "", "limit_strategy"),
		"expression",
	)
	assert.Contains(
		t,
		schemaEnum(t, constant.APISIXVersion317, "ai-rate-limiting", "", "limit_strategy"),
		"expression",
	)

	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "traffic-split", ""))
	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "traffic-split", "stream"))
}

func Test317ChangedPluginConfigsValidateByVersion(t *testing.T) {
	legacyCASConfig := map[string]any{
		"idp_uri":          "https://cas.example.com",
		"cas_callback_uri": "/api/cas/callback",
		"logout_uri":       "https://cas.example.com/logout",
	}
	assertPluginConfigValidity(t, constant.APISIXVersion313, "cas-auth", "", legacyCASConfig, true)
	assertPluginConfigValidity(t, constant.APISIXVersion317, "cas-auth", "", legacyCASConfig, false)

	assertPluginConfigValidity(
		t,
		constant.APISIXVersion317,
		"hmac-auth",
		"",
		map[string]any{"max_req_body_size": 0},
		false,
	)
	assertPluginConfigValidity(
		t,
		constant.APISIXVersion317,
		"hmac-auth",
		"",
		map[string]any{"max_req_body_size": 1},
		true,
	)

	edDSAConsumer := map[string]any{
		"key":        "consumer-key",
		"algorithm":  "EdDSA",
		"public_key": "public-key",
	}
	assertPluginConfigValidity(t, constant.APISIXVersion313, "jwt-auth", "consumer", edDSAConsumer, false)
	assertPluginConfigValidity(t, constant.APISIXVersion317, "jwt-auth", "consumer", edDSAConsumer, true)

	assertPluginConfigValidity(
		t,
		constant.APISIXVersion317,
		"batch-requests",
		"metadata",
		map[string]any{"max_pipeline_items": 0},
		false,
	)
	assertPluginConfigValidity(
		t,
		constant.APISIXVersion317,
		"batch-requests",
		"metadata",
		map[string]any{"max_pipeline_items": 1},
		true,
	)
}

func validate317PluginExample(
	t *testing.T,
	pluginName string,
	scope string,
	schemaType string,
	example map[string]any,
) {
	t.Helper()
	schemaValue := GetPluginSchema(constant.APISIXVersion317, pluginName, schemaType)
	require.NotNil(t, schemaValue, "%s/%s schema not found", pluginName, scope)
	schema := compile317Schema(t, pluginName+"/"+scope, schemaValue)
	result, err := schema.Validate(gojsonschema.NewGoLoader(example))
	require.NoError(t, err, "%s/%s", pluginName, scope)
	assert.Truef(t, result.Valid(), "%s/%s: %v", pluginName, scope, result.Errors())
}

func compile317Schema(t *testing.T, path string, schemaValue any) *gojsonschema.Schema {
	t.Helper()
	require.NotNil(t, schemaValue, "%s", path)
	schemaBytes, err := json.Marshal(schemaValue)
	require.NoError(t, err, "%s", path)
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))
	require.NoError(t, err, "%s", path)
	return schema
}

func requireStringMap(t *testing.T, value any, path string) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	require.True(t, ok, "%s should be an object, got %T", path, value)
	return result
}

func pluginSchemaMap(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	schemaType string,
) map[string]any {
	t.Helper()
	path := fmt.Sprintf("%s/%s/%s", version, pluginName, schemaType)
	return requireStringMap(t, GetPluginSchema(version, pluginName, schemaType), path)
}

func schemaProperty(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	schemaType string,
	property string,
) any {
	t.Helper()
	properties := requireStringMap(
		t,
		pluginSchemaMap(t, version, pluginName, schemaType)["properties"],
		fmt.Sprintf("%s/%s/%s/properties", version, pluginName, schemaType),
	)
	return properties[property]
}

func jwtAlgorithms(t *testing.T, version constant.APISIXVersion) []string {
	t.Helper()
	return schemaEnum(t, version, "jwt-auth", "consumer", "algorithm")
}

func schemaNumberKeyword(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	schemaType string,
	property string,
	keyword string,
) float64 {
	t.Helper()
	propertySchema := requireStringMap(
		t,
		schemaProperty(t, version, pluginName, schemaType, property),
		fmt.Sprintf("%s/%s/%s/properties/%s", version, pluginName, schemaType, property),
	)
	value, ok := propertySchema[keyword].(float64)
	require.True(t, ok, "%s should be numeric, got %T", keyword, propertySchema[keyword])
	return value
}

func schemaNestedNumberKeyword(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	schemaType string,
	path []string,
) float64 {
	t.Helper()
	var value any = pluginSchemaMap(t, version, pluginName, schemaType)
	for _, segment := range path {
		value = requireStringMap(t, value, fmt.Sprintf("%s/%s/%s", version, pluginName, segment))[segment]
	}
	result, ok := value.(float64)
	require.True(t, ok, "%v should be numeric, got %T", path, value)
	return result
}

func schemaRequired(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	schemaType string,
) []string {
	t.Helper()
	values, ok := pluginSchemaMap(t, version, pluginName, schemaType)["required"].([]any)
	if !ok {
		return nil
	}
	return requireStringSlice(t, values, fmt.Sprintf("%s/%s/%s/required", version, pluginName, schemaType))
}

func schemaEnum(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	schemaType string,
	property string,
) []string {
	t.Helper()
	propertySchema := requireStringMap(
		t,
		schemaProperty(t, version, pluginName, schemaType, property),
		fmt.Sprintf("%s/%s/%s/properties/%s", version, pluginName, schemaType, property),
	)
	values, ok := propertySchema["enum"].([]any)
	require.True(t, ok, "%s enum should be an array, got %T", property, propertySchema["enum"])
	return requireStringSlice(t, values, fmt.Sprintf("%s/%s/%s/%s/enum", version, pluginName, schemaType, property))
}

func requireStringSlice(t *testing.T, values []any, path string) []string {
	t.Helper()
	result := make([]string, 0, len(values))
	for index, value := range values {
		item, ok := value.(string)
		require.True(t, ok, "%s/%d should be a string, got %T", path, index, value)
		result = append(result, item)
	}
	return result
}

func assertPluginConfigValidity(
	t *testing.T,
	version constant.APISIXVersion,
	pluginName string,
	schemaType string,
	config map[string]any,
	expectedValid bool,
) {
	t.Helper()
	schema := compile317Schema(
		t,
		fmt.Sprintf("%s/%s/%s", version, pluginName, schemaType),
		GetPluginSchema(version, pluginName, schemaType),
	)
	result, err := schema.Validate(gojsonschema.NewGoLoader(config))
	require.NoError(t, err)
	assert.Equalf(t, expectedValid, result.Valid(), "%s/%s: %v", version, pluginName, result.Errors())
}

func pluginNames(plugins []*Plugin) []string {
	names := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		names = append(names, plugin.Name)
	}
	return names
}
