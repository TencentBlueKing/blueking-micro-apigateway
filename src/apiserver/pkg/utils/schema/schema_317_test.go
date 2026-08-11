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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func pluginNames(plugins []*Plugin) []string {
	names := make([]string, 0, len(plugins))
	for _, plugin := range plugins {
		names = append(names, plugin.Name)
	}
	return names
}
