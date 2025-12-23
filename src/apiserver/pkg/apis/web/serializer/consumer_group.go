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

// ConsumerGroupInfo   ConsumerGroup 基本信息
type ConsumerGroupInfo struct {
	ID     string          `json:"-"`                                                                   // 资源 apisix 资源 id
	Name   string          `json:"name" binding:"required" validate:"consumerGroupName"`                // ConsumerGroup 名称
	Config json.RawMessage `json:"config" validate:"apisixConfig=consumer_group"  swaggertype:"object"` // 配置数据 (json 格式)
}

// ConsumerGroupListRequest ConsumerGroup 表
type ConsumerGroupListRequest struct {
	ID      string `json:"id,omitempty" form:"id"`
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	Label   string `json:"label" form:"label"`
	Status  string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// ConsumerGroupListResponse ConsumerGroup 表
type ConsumerGroupListResponse []ConsumerGroupOutputInfo

// ConsumerGroupOutputInfo ConsumerGroup 表
type ConsumerGroupOutputInfo struct {
	AutoID    int    `json:"auto_id"`
	ID        string `json:"id"`
	GatewayID int    `json:"gateway_id"` // 网关 ID
	ConsumerGroupInfo
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// ConsumerGroupDropDownListResponse ConsumerGroup 下拉列表
type ConsumerGroupDropDownListResponse []ConsumerGroupDropDownOutputInfo

// ConsumerGroupDropDownOutputInfo ConsumerGroup 下拉列表
type ConsumerGroupDropDownOutputInfo struct {
	AutoID int    `json:"auto_id"` // 自增 ID
	ID     string `json:"id"`      // 资源 apisix 资源 id
	Name   string `json:"name"`    // 路由名称
	Desc   string `json:"desc"`    // 路由描述
}

// ValidateGroupID 校验 ConsumerGroupID
func ValidateGroupID(ctx context.Context, fl validator.FieldLevel) bool {
	groupID := fl.Field().String()
	if groupID == "" {
		return true
	}
	return biz.ExistsConsumerGroup(ctx, groupID)
}

// ValidateConsumerGroupName 校验 ConsumerGroupName
func ValidateConsumerGroupName(ctx context.Context, fl validator.FieldLevel) bool {
	consumerGroupName := fl.Field().String()
	if consumerGroupName == "" {
		return false
	}
	return !biz.DuplicatedResourceName(
		ctx,
		constant.ConsumerGroup,
		fl.Parent().FieldByName("ID").String(),
		consumerGroupName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"groupID",
		ValidateGroupID,
		"{0}:{1} 无效",
	)
	validation.AddBizFieldTagValidatorWithCtx(
		"consumerGroupName",
		ValidateConsumerGroupName,
		"{0}: {1} 该资源名称已经被存在的 consumer_group 资源占用",
	)
}
