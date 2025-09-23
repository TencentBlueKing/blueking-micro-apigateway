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

// PluginConfigInfo PluginConf 基本信息
type PluginConfigInfo struct {
	ID     string          `json:"id"`                                                                // 资源apisix资源id
	Name   string          `json:"name" binding:"required" validate:"pluginConfigName"`               // PluginConf名称
	Config json.RawMessage `json:"config" validate:"apisixConfig=plugin_config" swaggertype:"object"` // 配置数据(json格式)
}

// PluginConfigListRequest PluginConf 列表请求参数
type PluginConfigListRequest struct {
	ID      string `json:"id,omitempty" form:"id"`
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	Label   string `json:"label" form:"label"`
	Status  string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// PluginConfListResponse PluginConf 列表
type PluginConfListResponse []PluginConfigOutputInfo

// PluginConfigOutputInfo PluginConf 详情
type PluginConfigOutputInfo struct {
	AutoID int    `json:"auto_id"`
	ID     string `json:"id"`
	PluginConfigInfo
	GatewayID int                     `json:"gateway_id"` // 网关 ID
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// PluginConfigDropDownResponse PluginConf 下拉列表
type PluginConfigDropDownResponse []PluginConfigDropDownInfo

// PluginConfigDropDownInfo PluginConf 下拉列表
type PluginConfigDropDownInfo struct {
	AutoID int    `json:"auto_id"` // 自增 ID
	ID     string `json:"id"`      // 资源 apisix 资源 id
	Name   string `json:"name"`    // 路由名称
	Desc   string `json:"desc"`    // 路由描述
}

// ValidatePluginConfigID 校验 PluginConfigID
func ValidatePluginConfigID(ctx context.Context, fl validator.FieldLevel) bool {
	pluginConfigID := fl.Field().String()
	if pluginConfigID == "" {
		return true
	}
	return biz.ExistsPluginConfig(ctx, pluginConfigID)
}

// ValidatePluginConfigName 校验 PluginConfigName
func ValidatePluginConfigName(ctx context.Context, fl validator.FieldLevel) bool {
	pluginConfigName := fl.Field().String()
	if pluginConfigName == "" {
		return false
	}
	return biz.DuplicatedResourceName(
		ctx,
		constant.PluginConfig,
		fl.Parent().FieldByName("ID").String(),
		pluginConfigName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"pluginConfigID",
		ValidatePluginConfigID,
		"{0}:{1} 无效",
	)
	validation.AddBizFieldTagValidatorWithCtx(
		"pluginConfigName",
		ValidatePluginConfigName,
		"{0}: {1} 该资源名称已经被存在的 plugin_config 资源占用",
	)
}
