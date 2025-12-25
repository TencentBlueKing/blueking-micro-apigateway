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

// ProtoInfo Proto 基本信息
type ProtoInfo struct {
	ID     string          `json:"-"`                                                         // 资源apisix资源id
	Name   string          `json:"name" binding:"required" validate:"ProtoName"`              // Proto名称
	Config json.RawMessage `json:"config" validate:"apisixConfig=proto" swaggertype:"object"` // 配置数据(json格式)
}

// ProtoListRequest ...
type ProtoListRequest struct {
	ID      string `json:"id,omitempty" form:"id"`
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	Status  string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// ProtoListResponse Proto 列表
type ProtoListResponse []ProtoOutputInfo

// ProtoOutputInfo ...
type ProtoOutputInfo struct {
	AutoID    int    `json:"auto_id"`
	ID        string `json:"id"`
	GatewayID int    `json:"gateway_id"` // 网关 ID
	ProtoInfo
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// ProtoDropDownResponse Proto 下拉列表
type ProtoDropDownResponse []ProtoDropDownOutputInfo

// ProtoDropDownOutputInfo ...
type ProtoDropDownOutputInfo struct {
	AutoID int    `json:"auto_id"` // 自增 ID
	ID     string `json:"id"`      // 资源 apisix 资源 id
	Name   string `json:"name"`    // 路由名称
	Desc   string `json:"desc"`    // 路由描述
}

// ValidateProtoName 校验 proto 名称
func ValidateProtoName(ctx context.Context, fl validator.FieldLevel) bool {
	ProtoName := fl.Field().String()
	if ProtoName == "" {
		return false
	}
	return !biz.DuplicatedResourceName(
		ctx,
		constant.Proto,
		fl.Parent().FieldByName("ID").String(),
		ProtoName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"ProtoName",
		ValidateProtoName,
		"{0}: {1} 该资源名称已经被存在的 proto 资源占用",
	)
}
