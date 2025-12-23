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

package serializer

import (
	"context"
	"encoding/json"

	validator "github.com/go-playground/validator/v10"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// PluginMetadataInfo PluginMetadata 基本信息
type PluginMetadataInfo struct {
	ID     string          `json:"-"`                                                                   // 资源apisix资源id
	Name   string          `json:"name" binding:"required" validate:"pluginMetadataName"`               // PluginMetadata名称
	Config json.RawMessage `json:"config" validate:"apisixConfig=plugin_metadata" swaggertype:"object"` // 配置数据(json格式)
}

// PluginMetadataListRequest PluginMetadata 列表请求
type PluginMetadataListRequest struct {
	ID      string `json:"id,omitempty" form:"id"`
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	Status  string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// PluginMetadataListResponse PluginMetadata 列表
type PluginMetadataListResponse []PluginMetadataOutputInfo

// PluginMetadataOutputInfo PluginMetadata 详情
type PluginMetadataOutputInfo struct {
	AutoID    int    `json:"auto_id"`
	ID        string `json:"id"`
	GatewayID int    `json:"gateway_id"` // 网关 ID
	PluginMetadataInfo
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// PluginMetadataDropDownResponse PluginMetadata 下拉列表
type PluginMetadataDropDownResponse []PluginMetadataDropDownOutputInfo

// PluginMetadataDropDownOutputInfo ...
type PluginMetadataDropDownOutputInfo struct {
	AutoID int    `json:"auto_id"` // 自增 ID
	ID     string `json:"id"`      // 资源 apisix 资源 id
	Name   string `json:"name"`    // 路由名称
	Desc   string `json:"desc"`    // 路由描述
}

// ValidatePluginMetadataName 校验 plugin_metadata 资源名称
func ValidatePluginMetadataName(ctx context.Context, fl validator.FieldLevel) bool {
	pluginMetadataName := fl.Field().String()
	if pluginMetadataName == "" {
		return false
	}
	return !biz.DuplicatedResourceName(
		ctx,
		constant.PluginMetadata,
		fl.Parent().FieldByName("ID").String(),
		pluginMetadataName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"pluginMetadataName",
		ValidatePluginMetadataName,
		"{0}: {1} 该资源名称已经被存在的 plugin_metadata 资源占用",
	)
}
