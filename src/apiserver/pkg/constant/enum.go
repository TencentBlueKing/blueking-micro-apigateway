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

package constant

// GatewayControlModeDirect 网关纳管模式
const (
	GatewayControlModeDirect   uint8 = 1 // 直接管理
	GatewayControlModeInDirect uint8 = 2 // 纳管
)

// GatewayModeMap ...
var GatewayModeMap = map[uint8]string{
	GatewayControlModeDirect:   "direct",
	GatewayControlModeInDirect: "indirect",
}

// APISIX类型

// APISIXTypeAPISIX ...
const (
	APISIXTypeAPISIX   string = "apisix"    // 官方apisix
	APISIXTypeTAPISIX  string = "tapisix"   // tapisix
	APISIXTypeBKAPISIX string = "bk-apisix" // 蓝鲸apisix
)

// APISIXTypeMap ...
var APISIXTypeMap = map[string]string{
	APISIXTypeAPISIX:   "APISIX（开源社区版本）",
	APISIXTypeBKAPISIX: "BK-APISIX (蓝鲸定制版本）",
}

// SpecialPluginDocMap 特殊插件文档处理
var SpecialPluginDocMap = map[string]string{
	"serverless-pre-function":  "serverless",
	"serverless-post-function": "serverless",
}

// ResourceStatus ...
type ResourceStatus string

// String ...
func (r ResourceStatus) String() string {
	return string(r)
}

// ResourceStatusUpdateDraft 资源状态
const (
	ResourceStatusUpdateDraft ResourceStatus = "update_draft" // 更新待发布
	ResourceStatusCreateDraft ResourceStatus = "create_draft" // 创建待发布
	ResourceStatusSuccess     ResourceStatus = "success"      // 发布成功
	ResourceStatusConflict    ResourceStatus = "conflict"     // 配置冲突
	ResourceStatusDeleteDraft ResourceStatus = "delete_draft" // 删除等待发布
	ResourceStatusDeleted     ResourceStatus = "deleted"      // 已删除：用于操作日志记录
)

// ResourceStatusMap ...
var ResourceStatusMap = map[ResourceStatus]ResourceStatus{
	ResourceStatusUpdateDraft: ResourceStatusUpdateDraft,
	ResourceStatusCreateDraft: ResourceStatusCreateDraft,
	ResourceStatusSuccess:     ResourceStatusSuccess,
	ResourceStatusConflict:    ResourceStatusConflict,
	ResourceStatusDeleteDraft: ResourceStatusDeleteDraft,
}

// ResourcePath ...
type ResourcePath string

// String ...
func (r ResourcePath) String() string {
	return string(r)
}

// Routes 资源类型
const (
	Routes          ResourcePath = "routes"
	Upstreams       ResourcePath = "upstreams"
	Services        ResourcePath = "services"
	Consumers       ResourcePath = "consumers"
	GlobalRules     ResourcePath = "global_rules"
	ConsumerGroups  ResourcePath = "consumer_groups"
	PluginConfigs   ResourcePath = "plugin_configs"
	PluginMetadatas ResourcePath = "plugin_metadatas"
	Protos          ResourcePath = "protos"
	SSLs            ResourcePath = "ssls"
	StreamRoutes    ResourcePath = "stream_routes"
)

// ResourcePathToTypeMap ...
var ResourcePathToTypeMap = map[ResourcePath]APISIXResource{
	Routes:          Route,
	Upstreams:       Upstream,
	Services:        Service,
	Consumers:       Consumer,
	GlobalRules:     GlobalRule,
	ConsumerGroups:  ConsumerGroup,
	PluginConfigs:   PluginConfig,
	PluginMetadatas: PluginMetadata,
	Protos:          Proto,
	SSLs:            SSL,
	StreamRoutes:    StreamRoute,
}

// ResourceTypePrefixMap ...
var ResourceTypePrefixMap = map[APISIXResource]string{
	Route:          "routes",
	Upstream:       "upstreams",
	Service:        "services",
	Consumer:       "consumers",
	GlobalRule:     "global_rules",
	ConsumerGroup:  "consumer_groups",
	PluginConfig:   "plugin_configs",
	PluginMetadata: "plugin_metadata",
	Proto:          "protos",
	SSL:            "ssls",
	StreamRoute:    "stream_routes",
}

// ResourcePrefixTypeMap ...
var ResourcePrefixTypeMap = map[string]APISIXResource{
	"routes":          Route,
	"upstreams":       Upstream,
	"services":        Service,
	"consumers":       Consumer,
	"global_rules":    GlobalRule,
	"consumer_groups": ConsumerGroup,
	"plugin_configs":  PluginConfig,
	"plugin_metadata": PluginMetadata,
	"protos":          Proto,
	"ssls":            SSL,
	"stream_routes":   StreamRoute,
}

// SyncStatus 资源同步状态
type SyncStatus string

// SyncedResourceStatusSuccess ...
const (
	SyncedResourceStatusSuccess SyncStatus = "success" // 同步成功
	SyncedResourceStatusMiss    SyncStatus = "miss"    // 编辑区无此资源
)

// SyncedResourceStatusMap ...
var SyncedResourceStatusMap = map[SyncStatus]SyncStatus{
	SyncedResourceStatusSuccess: SyncedResourceStatusSuccess,
	SyncedResourceStatusMiss:    SyncedResourceStatusMiss,
}

// UploadStatus 资源上传状态
type UploadStatus string

const (
	UploadStatusAdd    UploadStatus = "add"    // 新增
	UploadStatusUpdate UploadStatus = "update" // 更新
)

// UploadResourceStatusMap ...
var UploadResourceStatusMap = map[UploadStatus]UploadStatus{
	UploadStatusAdd:    UploadStatusAdd,
	UploadStatusUpdate: UploadStatusUpdate,
}

// PublishByBkAPISIXControlPlane ...
const (
	PublishByBkAPISIXControlPlane = "bk-apisix-control-plane"
	PublishByOthers               = "others"
)

// OperationType 资源操作类型
type OperationType string

// String ...
func (o OperationType) String() string {
	return string(o)
}

// OperationTypeCreate ...
const (
	OperationTypeCreate      OperationType = "create"            // 创建
	OperationTypeUpdate      OperationType = "update"            // 更新
	OperationTypeDelete      OperationType = "delete"            // 删除
	OperationTypePublish     OperationType = "publish"           // 同步
	OperationTypeRevert      OperationType = "revert"            // 撤销
	OperationTypeFixConflict OperationType = "fix_conflict"      // 解决冲突
	OperationOneClickManaged OperationType = "one_click_managed" // 一键同步（数据量太大，不添加审计）
)

// OperationTypeMap ...
var OperationTypeMap = map[OperationType]string{
	OperationTypeCreate:      "新增",
	OperationTypeUpdate:      "更新",
	OperationTypeDelete:      "删除",
	OperationTypePublish:     "发布",
	OperationTypeRevert:      "撤销",
	OperationTypeFixConflict: "解决冲突",
}

// HTTP ...
const (
	HTTP  string = "http"
	HTTPS string = "https"
)

// SchemaTypeMap ...
var SchemaTypeMap = map[string]string{
	HTTP:  "http",
	HTTPS: "https",
}

// CustomizePlugin 自定义插件
const CustomizePlugin string = "customize plugin"

// Metadata 插件类别
const (
	Metadata string = "metadata"
	Stream   string = "stream"
)

// EmptyAssociationFilter 空关联关系的过滤标识
const EmptyAssociationFilter string = "--"

// ANYMethodFilter 空 methods 的过滤标识
const ANYMethodFilter string = "ANY"

// DataType 数据类型
type DataType string

const (
	DATABASE DataType = "db"
	ETCD     DataType = "etcd"
)
