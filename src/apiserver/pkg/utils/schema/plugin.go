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

// Package schema ...
package schema

import (
	_ "embed"
	"encoding/json"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

//go:embed 3.13/plugin.json
var rawPluginV313 []byte

//go:embed 3.11/plugin.json
var rawPluginV311 []byte

//go:embed 3.3/plugin.json
var rawPluginV33 []byte

//go:embed 3.2/plugin.json
var rawPluginV32 []byte

// bk-apisix plugin
//
//go:embed 3.13/bk_apisix_plugin.json
var rawBkAPISIXPluginV313 []byte

//go:embed 3.11/bk_apisix_plugin.json
var rawBkAPISIXPluginV311 []byte

// tapisix plugin
//
//go:embed 3.13/tapisix_plugin.json
var rawTAPISIXPluginV313 []byte

//go:embed 3.11/tapisix_plugin.json
var rawTAPISIXPluginV311 []byte

//go:embed 3.3/tapisix_plugin.json
var rawTAPISIXPluginV33 []byte

var versionPluginMap = map[constant.APISIXVersion][]byte{
	constant.APISIXVersion32:  rawPluginV32,
	constant.APISIXVersion33:  rawPluginV33,
	constant.APISIXVersion311: rawPluginV311,
	constant.APISIXVersion313: rawPluginV313,
}

var versionBkAPISIXPluginMap = map[constant.APISIXVersion][]byte{
	constant.APISIXVersion313: rawBkAPISIXPluginV313,
	constant.APISIXVersion311: rawBkAPISIXPluginV311,
}

var versionTAPISIXPluginMap = map[constant.APISIXVersion][]byte{
	constant.APISIXVersion33:  rawTAPISIXPluginV33,
	constant.APISIXVersion311: rawTAPISIXPluginV311,
	constant.APISIXVersion313: rawTAPISIXPluginV313,
}

// VersionDocUrlMap ...
var VersionDocUrlMap = map[constant.APISIXVersion]string{
	constant.APISIXVersion32:  "https://apache-apisix.netlify.app/zh/docs/apisix/3.2/plugins/%s/",
	constant.APISIXVersion33:  "https://apache-apisix.netlify.app/zh/docs/apisix/3.3/plugins/%s/",
	constant.APISIXVersion311: "https://apache-apisix.netlify.app/zh/docs/apisix/3.11/plugins/%s/",
	constant.APISIXVersion313: "https://apisix.apache.org/zh/docs/apisix/plugins/%s/",
}

// Plugin ...
type Plugin struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	ProxyType       string         `json:"proxy_type"` // only for stream plugin
	Example         map[string]any `json:"example"`
	MetadataExample map[string]any `json:"metadata_example,omitempty"`
	ConsumerExample map[string]any `json:"consumer_example,omitempty"`
	DocUrl          string         `json:"doc_url"`
}

// StreamRoutePluginMap ...
var StreamRoutePluginMap = map[string]string{
	"ip-restriction": "ip-restriction",
	"limit-conn":     "limit-conn",
	"mqtt-proxy":     "mqtt-proxy",
	"prometheus":     "prometheus",
	"syslog":         "syslog",
}

// GetPlugins 获取插件
func GetPlugins(apisixType string, version constant.APISIXVersion) ([]*Plugin, error) {
	var plugins []*Plugin
	err := json.Unmarshal(versionPluginMap[version], &plugins)
	if err != nil {
		return nil, err
	}
	// 如果是apisix类型，直接返回
	if apisixType == constant.APISIXTypeAPISIX {
		return plugins, nil
	}

	if tapisixPluginInfo, ok := versionTAPISIXPluginMap[version]; ok {
		var tapisixPlugins []*Plugin
		err = json.Unmarshal(tapisixPluginInfo, &tapisixPlugins)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, tapisixPlugins...)
	}

	// 如果是tapisix类型，直接返回 apisix插件+tapisix插件
	if apisixType == constant.APISIXTypeTAPISIX {
		return plugins, nil
	}

	// 如果是蓝鲸类型，直接返回 apisix插件+tapisix插件+bk插件
	if bkAPISIXPluginInfo, ok := versionBkAPISIXPluginMap[version]; ok {
		var bkPlugins []*Plugin
		err = json.Unmarshal(bkAPISIXPluginInfo, &bkPlugins)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, bkPlugins...)
		return plugins, nil
	}
	return plugins, nil
}
