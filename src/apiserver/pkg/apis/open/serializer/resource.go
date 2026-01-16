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

// Package serializer ...
package serializer

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/common"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// ResourceCreateRequest 资源创建
type ResourceCreateRequest struct {
	Name   string          `json:"name" binding:"required"`
	Config json.RawMessage `json:"config"  swaggertype:"object"` // 配置数据 (json 格式)
}

// ResourceCreateResponse ...
type ResourceCreateResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// ResourceAssociateID 资源关联 ID
type ResourceAssociateID struct {
	ServiceID      string `json:"service_id" validate:"serviceID"`            // 服务 ID
	UpstreamID     string `json:"upstream_id" validate:"upstreamID"`          // 上游服务地址 ID
	PluginConfigID string `json:"plugin_config_id" validate:"pluginConfigID"` // 插件配置 groupID
	GroupID        string `json:"group_id" validate:"groupID"`
}

// ResourceBatchCreateRequest 资源批量创建
type ResourceBatchCreateRequest []ResourceCreateRequest

// ToCommonResource 转换为通用资源
func (rs ResourceBatchCreateRequest) ToCommonResource(gatewayID int,
	resourceType constant.APISIXResource,
) []*model.ResourceCommonModel {
	var resources []*model.ResourceCommonModel
	for _, r := range rs {
		if gjson.GetBytes(r.Config, "name").String() == "" {
			r.Config, _ = sjson.SetBytes(r.Config, model.GetResourceNameKey(resourceType), r.Name)
		}
		id := gjson.GetBytes(r.Config, "id").String()
		if id == "" {
			id = idx.GenResourceID(resourceType)
		}
		resource := &model.ResourceCommonModel{
			ID:        id,
			GatewayID: gatewayID,
			Config:    datatypes.JSON(r.Config),
			Status:    constant.ResourceStatusCreateDraft,
		}
		resources = append(resources, resource)
	}
	return resources
}

// ResourceBatchGetRequest 资源获取参数
type ResourceBatchGetRequest struct {
	IDs []string `form:"ids" `
}

// ResourceBatchDeleteRequest 资源批量删除请求参数
type ResourceBatchDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// ResourceBatchGetResponse ...
type ResourceBatchGetResponse struct {
	ID              string `json:"id"`
	json.RawMessage `json:"config" swaggertype:"object"`
}

// ResourcePublishRequest ...
type ResourcePublishRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// ResourceGetResponse 单个资源响应
type ResourceGetResponse struct {
	ID              string `json:"id"`
	json.RawMessage `json:"config" swaggertype:"object"`
}

// ResourceGetStatusResponse 单个资源状态响应
type ResourceGetStatusResponse struct {
	ID     string                  `json:"id"`
	Status constant.ResourceStatus `json:"status"`
}

// ResourcePathParam 单个资源获取
type ResourcePathParam struct {
	ID string `json:"id" uri:"id"`
}

// ResourceUpdateRequest 资源更新
type ResourceUpdateRequest struct {
	Name   string          `json:"name" binding:"required"`
	Config json.RawMessage `json:"config"  swaggertype:"object"` // 配置数据 (json 格式)
}

// ToCommonResource 转换为通用资源
func (r ResourceUpdateRequest) ToCommonResource(
	c *gin.Context,
	id string,
	status constant.ResourceStatus,
) *model.ResourceCommonModel {
	resource := &model.ResourceCommonModel{
		ID:        id,
		GatewayID: ginx.GetGatewayInfo(c).ID,
		Config:    datatypes.JSON(r.Config),
		Status:    status,
		BaseModel: model.BaseModel{
			Updater: ginx.GetUserID(c),
		},
	}
	return resource
}

// ResourceImportRequest 资源导入请求
type ResourceImportRequest struct {
	Data     map[constant.APISIXResource][]*common.ResourceInfo
	Metadata Metadata `json:"metadata"`
}

// UnmarshalJSON 自定义解析 JSON
func (w *ResourceImportRequest) UnmarshalJSON(data []byte) error {
	// 先解析整个 map
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// 提取 metadata
	if v, ok := raw["metadata"]; ok {
		if err := json.Unmarshal(v, &w.Metadata); err != nil {
			return err
		}
		delete(raw, "metadata")
	}
	// 剩余部分解析为资源数据
	w.Data = make(map[constant.APISIXResource][]*common.ResourceInfo)
	for key, val := range raw {
		var resources []*common.ResourceInfo
		if err := json.Unmarshal(val, &resources); err != nil {
			return err
		}
		w.Data[constant.APISIXResource(key)] = resources
	}
	return nil
}

type Metadata struct {
	// 跳过规则，用于设置针对某些资源不进行修改设置
	IgnoreFields map[constant.APISIXResource][]string `json:"ignore_fields"`
}
