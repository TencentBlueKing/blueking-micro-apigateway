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

func Test318OfficialAssetsRegistered(t *testing.T) {
	assert.NotNil(t, GetResourceSchema(constant.APISIXVersion318, constant.Route.String()))
	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion318, "ai-cache", ""))
	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion318, "log-rotate", ""))
	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion318, "traffic-split", "stream"))
	assert.Nil(t, GetPluginSchema(constant.APISIXVersion318, "prometheus", "stream"))

	plugins, err := GetPlugins(constant.APISIXTypeAPISIX, constant.APISIXVersion318)
	require.NoError(t, err)
	names := pluginNames(plugins)
	assert.Contains(t, names, "ai-cache")
	assert.Contains(t, names, "ai-lakera-guard")
	assert.NotContains(t, names, "batch-requests")
	assert.NotContains(t, names, "mcp-bridge")
}

func Test318PluginCompositionPreservesCompatibilityCatalog(t *testing.T) {
	enableTAPISIXForTest(t)

	official, err := GetPlugins(constant.APISIXTypeAPISIX, constant.APISIXVersion318)
	require.NoError(t, err)
	tapisix, err := GetPlugins(constant.APISIXTypeTAPISIX, constant.APISIXVersion318)
	require.NoError(t, err)
	bk, err := GetPlugins(constant.APISIXTypeBKAPISIX, constant.APISIXVersion318)
	require.NoError(t, err)

	assert.Len(t, tapisix, len(official)+len(expectedTAPISIXPluginNames))
	for _, name := range expectedTAPISIXPluginNames {
		assert.Contains(t, pluginNames(tapisix), name)
	}
	assert.Len(t, bk, len(official)+len(expectedTAPISIXPluginNames)+7)
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
		assert.NotNil(t, GetPluginSchema(constant.APISIXVersion318, name, ""))
	}
}

func Test318CatalogSchemasAndExamples(t *testing.T) {
	catalogs := []struct {
		name string
		raw  []byte
	}{
		{name: "official", raw: rawPluginV318},
		{name: "bk", raw: rawBkAPISIXPluginV318},
		{name: "tapisix", raw: rawTAPISIXPluginV318},
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
			validate318PluginExample(t, plugin.Name, "main", mainSchemaType, plugin.Example)
			if plugin.ConsumerExample != nil {
				validate318PluginExample(t, plugin.Name, "consumer", "consumer", plugin.ConsumerExample)
			}
			if plugin.MetadataExample != nil {
				validate318PluginExample(t, plugin.Name, "metadata", "metadata", plugin.MetadataExample)
			}
		}
	}
}

func Test318SchemasCoverRuntimePluginsExcludedFromCatalog(t *testing.T) {
	official, err := GetPlugins(constant.APISIXTypeAPISIX, constant.APISIXVersion318)
	require.NoError(t, err)
	catalogNames := pluginNames(official)

	for _, name := range []string{
		"ai",
		"example-plugin",
		"inspect",
		"log-rotate",
		"mcp-bridge",
	} {
		assert.NotContains(t, catalogNames, name)
		assert.NotNil(t, GetPluginSchema(constant.APISIXVersion318, name, ""), name)
	}
	assert.Nil(t, GetPluginSchema(constant.APISIXVersion318, "node-status", ""))
}

func Test318AllSchemaNodesCompile(t *testing.T) {
	for _, resource := range constant.ResourceTypeList {
		compile318Schema(
			t,
			"main."+resource.String(),
			GetResourceSchema(constant.APISIXVersion318, resource.String()),
		)
	}

	sources := []struct {
		name string
		root any
	}{
		{name: "official", root: schemaVersionMap[constant.APISIXVersion318].Value()},
		{name: "bk", root: bkAPISIXPluginSchemaVersionMap[constant.APISIXVersion318].Value()},
		{name: "tapisix", root: tapisixPluginSchemaVersionMap[constant.APISIXVersion318].Value()},
	}
	for _, source := range sources {
		root := requireStringMap(t, source.root, source.name)
		plugins, _ := root["plugins"].(map[string]any)
		for pluginName, pluginValue := range plugins {
			plugin := requireStringMap(t, pluginValue, source.name+".plugins."+pluginName)
			for _, scope := range []string{"schema", "consumer_schema", "metadata_schema"} {
				if schemaValue, ok := plugin[scope]; ok {
					compile318Schema(
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
				compile318Schema(t, source.name+".stream_plugins."+pluginName+".schema", schemaValue)
			}
		}
	}
}

func Test318RepresentativeSchemaDifferences(t *testing.T) {
	assert.Nil(t, GetPluginSchema(constant.APISIXVersion317, "ai-cache", ""))
	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion318, "ai-cache", ""))
	assert.Nil(t, GetPluginSchema(constant.APISIXVersion317, "ai-lakera-guard", ""))
	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion318, "ai-lakera-guard", ""))

	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "batch-requests", ""))
	assert.Nil(t, GetPluginSchema(constant.APISIXVersion318, "batch-requests", ""))
	assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "prometheus", "stream"))
	assert.Nil(t, GetPluginSchema(constant.APISIXVersion318, "prometheus", "stream"))
}

func validate318PluginExample(
	t *testing.T,
	pluginName string,
	scope string,
	schemaType string,
	example map[string]any,
) {
	t.Helper()
	schemaValue := GetPluginSchema(constant.APISIXVersion318, pluginName, schemaType)
	require.NotNil(t, schemaValue, "%s/%s schema not found", pluginName, scope)
	schema := compile318Schema(t, pluginName+"/"+scope, schemaValue)
	result, err := schema.Validate(gojsonschema.NewGoLoader(example))
	require.NoError(t, err, "%s/%s", pluginName, scope)
	assert.Truef(t, result.Valid(), "%s/%s: %v", pluginName, scope, result.Errors())
}

func compile318Schema(t *testing.T, path string, schemaValue any) *gojsonschema.Schema {
	t.Helper()
	require.NotNil(t, schemaValue, "%s", path)
	schemaBytes, err := json.Marshal(schemaValue)
	require.NoError(t, err, "%s", path)
	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))
	require.NoError(t, err, "%s", path)
	return schema
}
