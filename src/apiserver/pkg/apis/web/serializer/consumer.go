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

package serializer

import (
	"context"
	"encoding/json"

	validator "github.com/go-playground/validator/v10"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// ConsumerInfo Consumer 基本信息
type ConsumerInfo struct {
	ID      string          `json:"-"`                                                            // 资源 apisix 资源 id
	Name    string          `json:"name" binding:"required" validate:"consumerName"`              // Consumer 名称
	GroupID string          `json:"group_id" validate:"groupID"`                                  // ConsumerGroupID
	Config  json.RawMessage `json:"config" validate:"apisixConfig=consumer" swaggertype:"object"` // 配置数据 (json 格式)
}

// ConsumerListRequest Consumer 列表请求参数
type ConsumerListRequest struct {
	ID      string `json:"id,omitempty" form:"id"`
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	GroupID string `json:"group_id,omitempty" form:"group_id"`
	Label   string `json:"label" form:"label"`
	Status  string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// ConsumerListResponse Consumer 列表
type ConsumerListResponse []ConsumerOutputInfo

// ConsumerOutputInfo Consumer 基本信息
type ConsumerOutputInfo struct {
	AutoID    int    `json:"auto_id"`
	ID        string `json:"id"`
	GatewayID int    `json:"gateway_id"` // 网关 ID
	ConsumerInfo
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// ConsumerDropDownListResponse Consumer 下拉列表
type ConsumerDropDownListResponse []ConsumerDropDownOutputInfo

// ConsumerDropDownOutputInfo Consumer 下拉列表
type ConsumerDropDownOutputInfo struct {
	AutoID int    `json:"auto_id"` // 自增 ID
	ID     string `json:"id"`      // 资源 apisix 资源 id
	Name   string `json:"name"`    // 路由名称
	Desc   string `json:"desc"`    // 路由描述
}

// ValidateConsumerName 校验 consumer 资源名称
func ValidateConsumerName(ctx context.Context, fl validator.FieldLevel) bool {
	consumerName := fl.Field().String()
	if consumerName == "" {
		return false
	}
	return !biz.DuplicatedResourceName(
		ctx,
		constant.Consumer,
		fl.Parent().FieldByName("ID").String(),
		consumerName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"consumerName",
		ValidateConsumerName,
		"{0}: {1} 该资源名称已经被存在的 consumer 资源占用",
	)
}
