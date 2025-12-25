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

// RouteInfo route 基本信息
type RouteInfo struct {
	AutoID         int             `json:"-"`                                                         // 自增ID
	ID             string          `json:"id"`                                                        // 资源apisix资源id
	Name           string          `json:"name" binding:"required" validate:"routeName"`              // 路由名称
	ServiceID      string          `json:"service_id" validate:"serviceID"`                           // 服务ID
	UpstreamID     string          `json:"upstream_id" validate:"upstreamID"`                         // 上游服务地址ID
	PluginConfigID string          `json:"plugin_config_id" validate:"pluginConfigID"`                // 插件配置ID
	Config         json.RawMessage `json:"config" validate:"apisixConfig=route" swaggertype:"object"` // 路由配置(json格式)
}

// RouteListRequest ...
type RouteListRequest struct {
	ID         string `json:"id,omitempty" form:"id"`
	Name       string `json:"name,omitempty" form:"name"`
	Updater    string `json:"updater,omitempty" form:"updater"`
	ServiceID  string `json:"service_id" form:"service_id"`
	UpstreamID string `json:"upstream_id" form:"upstream_id"`
	Label      string `json:"label" form:"label"`
	Path       string `json:"path" form:"path"`
	Method     string `json:"method" form:"method"`
	Status     string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy    string `json:"order_by" form:"order_by"`
	Offset     int    `json:"offset" form:"offset"`
	Limit      int    `json:"limit" form:"limit"`
}

// RouteListResponse route 列表
type RouteListResponse []RouteOutputInfo

// RouteOutputInfo ...
type RouteOutputInfo struct {
	AutoID    int `json:"auto_id"`
	GatewayID int `json:"gateway_id"` // 网关 ID
	RouteInfo
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// RouteDropDownListResponse route 下拉列表
type RouteDropDownListResponse []RouteDropDownOutputInfo

// RouteDropDownOutputInfo ...
type RouteDropDownOutputInfo struct {
	ID     string   `json:"id"`      // 资源 apisix 资源 id
	AutoID int      `json:"auto_id"` // 自增 ID
	Name   string   `json:"name"`    // 路由名称
	Uris   []string `json:"uris"`    // 路由路径
	Desc   string   `json:"desc"`    // 路由描述
}

// ValidationRouteName ...
func ValidationRouteName(ctx context.Context, fl validator.FieldLevel) bool {
	routeName := fl.Field().String()
	if routeName == "" {
		return false
	}
	return !biz.DuplicatedResourceName(
		ctx,
		constant.Route,
		fl.Parent().FieldByName("ID").String(),
		routeName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"routeName",
		ValidationRouteName,
		"{0}: {1} 该资源名称已经被存在的 route 资源占用",
	)
}
