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

package dto

import (
	"encoding/json"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// ResourceChangeInfo ...
type ResourceChangeInfo struct {
	ResourceType constant.APISIXResource `json:"resource_type"`  // 资源类型：route/upstream/...
	AddedCount   int                     `json:"added_count"`    // 新增数量
	DeletedCount int                     `json:"deleted_count"`  // 删除数量
	UpdateCount  int                     `json:"modified_count"` // 更新数量
	ChangeDetail []ResourceChangeDetail  `json:"change_detail"`  // 变更详情
}

// ResourceChangeDetail ...
type ResourceChangeDetail struct {
	ResourceID   string                  `json:"resource_id"`
	Name         string                  `json:"name"`
	BeforeStatus constant.ResourceStatus `json:"before_status"`
	AfterStatus  constant.ResourceStatus `json:"after_status"`
	PublishFrom  constant.OperationType  `json:"operation_type"` // 发布变更来源操作： add/delete/update
	UpdatedAt    int64                   `json:"updated_at"`
}

// ResourceDiffDetailResponse ...
type ResourceDiffDetailResponse struct {
	EditorConfig json.RawMessage `json:"editor_config" swaggertype:"object"` // 编辑区配置
	EtcdConfig   json.RawMessage `json:"etcd_config" swaggertype:"object"`   // etcd生效配置
}

// ResourceAssociateID 资源关联ID
type ResourceAssociateID struct {
	ServiceID      string `json:"service_id" validate:"serviceID"`            // 服务ID
	UpstreamID     string `json:"upstream_id" validate:"upstreamID"`          // 上游服务地址ID
	PluginConfigID string `json:"plugin_config_id" validate:"pluginConfigID"` // 插件配置groupID
	GroupID        string `json:"group_id" validate:"groupID"`
}
