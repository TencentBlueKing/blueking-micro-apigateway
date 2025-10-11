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
	_ "embed"

	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

//go:embed 3.13/schema.json
var rawSchemaV313 []byte

// BK-APISIX plugin schema
//
//go:embed 3.11/bk_apisix_plugin_schema.json
var rawBkAPISIXPluginSchemaV311 []byte

//go:embed 3.13/bk_apisix_plugin_schema.json
var rawBkAPISIXPluginSchemaV313 []byte

// TAPISIX plugin schema
//
//go:embed 3.3/tapisix_plugin_schema.json
var rawTAPISIXPluginSchemaV33 []byte

//go:embed 3.11/tapisix_plugin_schema.json
var rawTAPISIXPluginSchemaV311 []byte

//go:embed 3.13/tapisix_plugin_schema.json
var rawTAPISIXPluginSchemaV313 []byte

//go:embed 3.11/schema.json
var rawSchemaV311 []byte

//go:embed 3.3/schema.json
var rawSchemaV33 []byte

//go:embed 3.2/schema.json
var rawSchemaV32 []byte

var schemaVersionMap = map[constant.APISIXVersion]gjson.Result{
	constant.APISIXVersion32:  gjson.ParseBytes(rawSchemaV32),
	constant.APISIXVersion33:  gjson.ParseBytes(rawSchemaV33),
	constant.APISIXVersion311: gjson.ParseBytes(rawSchemaV311),
	constant.APISIXVersion313: gjson.ParseBytes(rawSchemaV313),
}

var bkAPISIXPluginSchemaVersionMap = map[constant.APISIXVersion]gjson.Result{
	constant.APISIXVersion313: gjson.ParseBytes(rawBkAPISIXPluginSchemaV313),
	constant.APISIXVersion311: gjson.ParseBytes(rawBkAPISIXPluginSchemaV311),
}

var tapisixPluginSchemaVersionMap = map[constant.APISIXVersion]gjson.Result{
	constant.APISIXVersion33:  gjson.ParseBytes(rawTAPISIXPluginSchemaV33),
	constant.APISIXVersion311: gjson.ParseBytes(rawTAPISIXPluginSchemaV311),
	constant.APISIXVersion313: gjson.ParseBytes(rawTAPISIXPluginSchemaV313),
}

// GetResourceSchema 获取资源的schema
func GetResourceSchema(version constant.APISIXVersion, name string) interface{} {
	return schemaVersionMap[version].Get("main." + name).Value()
}

// GetMetadataPluginSchema 获取 metadata 插件类型的 schema
func GetMetadataPluginSchema(version constant.APISIXVersion, path string) interface{} {
	// 查找 apisix 插件
	ret := schemaVersionMap[version].Get(path).Value()
	if ret != nil {
		return ret
	}
	// 查找 bk-apisix 插件
	bkAPISIXPluginSchemaVersion, ok := bkAPISIXPluginSchemaVersionMap[version]
	if ok {
		ret = bkAPISIXPluginSchemaVersion.Get(path).Value()
	}
	if ret != nil {
		return ret
	}
	// 查找 tapisix 插件
	tapisixPluginSchemaVersion, ok := tapisixPluginSchemaVersionMap[version]
	if ok {
		ret = tapisixPluginSchemaVersion.Get(path).Value()
	}
	return ret
}

// GetPluginSchema 获取插件的schema
func GetPluginSchema(version constant.APISIXVersion, name string, schemaType string) interface{} {
	var ret interface{}
	if schemaType == "consumer" || schemaType == "consumer_schema" {
		// 需匹配常规插件和 consumer 插件，当未查询到时，继续匹配后面常规插件
		ret = schemaVersionMap[version].Get("plugins." + name + ".consumer_schema").Value()
	}
	if schemaType == "metadata" || schemaType == "metadata_schema" {
		// 只需匹配 metadata 类型的插件，根据 "plugins."+name+".metadata_schema" 路径查询 schema，可直接返回结果，无需再匹配常规插件
		return GetMetadataPluginSchema(version, "plugins."+name+".metadata_schema")
	}
	if schemaType == "stream" || schemaType == "stream_schema" {
		// 只需要匹配 stream 类型的插件，由于该类型所有插件已在 schema.json 中存在，可直接返回结果，无需再匹配常规插件
		return schemaVersionMap[version].Get("stream_plugins." + name + ".schema").Value()
	}
	// 常规插件匹配
	if ret == nil {
		ret = schemaVersionMap[version].Get("plugins." + name + ".schema").Value()
	}
	if ret != nil {
		return ret
	}
	// 如果apisix插件不存在，再去bk-apisix插件中查找
	bkAPISIXPluginSchemaVersion, ok := bkAPISIXPluginSchemaVersionMap[version]
	if ok {
		ret = bkAPISIXPluginSchemaVersion.Get("plugins." + name + ".schema").Value()
	}
	if ret != nil {
		return ret
	}
	// 如果bk-apisix插件也不存在，再去tapisix插件中查找
	tapisixPluginSchemaVersion, ok := tapisixPluginSchemaVersionMap[version]
	if ok {
		ret = tapisixPluginSchemaVersion.Get("plugins." + name + ".schema").Value()
	}

	return ret
}
