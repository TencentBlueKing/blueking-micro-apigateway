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

// Package constant 管理常量
package constant

// APISIXResource ...
type APISIXResource string

// Route ...
const (
	Route          APISIXResource = "route"
	Service        APISIXResource = "service"
	Upstream       APISIXResource = "upstream"
	PluginConfig   APISIXResource = "plugin_config"
	PluginMetadata APISIXResource = "plugin_metadata"
	Consumer       APISIXResource = "consumer"
	ConsumerGroup  APISIXResource = "consumer_group"
	GlobalRule     APISIXResource = "global_rule"
	Proto          APISIXResource = "proto"
	SSL            APISIXResource = "ssl"
	StreamRoute    APISIXResource = "stream_route"
	Schema         APISIXResource = "schema"  // 操作审计场景使用
	Gateway        APISIXResource = "gateway" // 操作审计场景使用
)

const ResourceKeyFormat = "%s-%s" // type-resource-id

// RelationIDFiledMap ...
var RelationIDFiledMap = map[APISIXResource]string{
	Service:       "service_id",
	Upstream:      "upstream_id",
	PluginConfig:  "plugin_config_id",
	ConsumerGroup: "group_id",
	SSL:           "ssl_id",
}

// String ...
func (r APISIXResource) String() string {
	return string(r)
}

// RelationIDFiled 获取关联ID字段
func (r APISIXResource) RelationIDFiled() string {
	return RelationIDFiledMap[r]
}

// Operation ...
type Operation string

// Put ...
const (
	Put    Operation = "put"
	Delete Operation = "delete"
)

// ResourceTypeMap 资源类型
var ResourceTypeMap = map[APISIXResource]string{
	Route:          "路由",
	StreamRoute:    "stream 路由",
	Service:        "服务",
	Upstream:       "上游",
	Proto:          "proto",
	SSL:            "证书",
	Consumer:       "消费者",
	ConsumerGroup:  "消费者组",
	PluginMetadata: "插件元数据",
	GlobalRule:     "全局规则",
	PluginConfig:   "插件组",
	Schema:         "自定义插件",
	Gateway:        "网关",
}

// ResourceTypeOrder 资源类型顺序
var ResourceTypeOrder = []APISIXResource{
	Route,
	StreamRoute,
	Service,
	Upstream,
	Proto,
	SSL,
	Consumer,
	ConsumerGroup,
	PluginMetadata,
	GlobalRule,
	PluginConfig,
	Schema,
	Gateway,
}

// ResourceTypeList ...
var ResourceTypeList = []APISIXResource{
	Route,
	Service,
	Upstream,
	PluginConfig,
	PluginMetadata,
	Consumer,
	ConsumerGroup,
	GlobalRule,
	Proto,
	SSL,
	StreamRoute,
}

// APISIXVersion ...
type APISIXVersion string

// APISIXVersion311 ...
const (
	APISIXVersion313 APISIXVersion = "3.13.X"
	APISIXVersion311 APISIXVersion = "3.11.X"
	APISIXVersion33  APISIXVersion = "3.3.X"
	APISIXVersion32  APISIXVersion = "3.2.X"
)

// SupportAPISIXVersionMap ...
var SupportAPISIXVersionMap = map[string]string{
	"3.13.X": string(APISIXVersion313),
	"3.11.X": string(APISIXVersion311),
	"3.3.X":  string(APISIXVersion33),
	"3.2.X":  string(APISIXVersion32),
}

// SSLDefaultStatus ...
const SSLDefaultStatus = 1

// ResourceRelationMap ...
var ResourceRelationMap = map[APISIXResource][]APISIXResource{
	Service:       {Route, StreamRoute},
	Upstream:      {Route, Service, StreamRoute},
	SSL:           {Upstream},
	ConsumerGroup: {Consumer},
}

// PluginsMustResourceMap 必须要配置插件的资源
var PluginsMustResourceMap = map[APISIXResource]bool{
	PluginConfig:   true,
	PluginMetadata: true,
	ConsumerGroup:  true,
	GlobalRule:     true,
}
