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
	"strings"

	validator "github.com/go-playground/validator/v10"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/common"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// GatewayListRequest 网关列表请求
type GatewayListRequest struct {
	Mode   uint8 `json:"mode" form:"mode" binding:"omitempty,gatewayMode"` // 网关control模式：1-direct 2-indirect
	Offset int   `json:"offset" form:"offset"`
	Limit  int   `json:"limit" form:"limit"`
}

// GatewayListResponse 网关列表
type GatewayListResponse []GatewayOutputListInfo

// Count 统计信息
type Count struct {
	Route    int64 `json:"route"`
	Service  int64 `json:"service"`
	Upstream int64 `json:"upstream"`
}

// GatewayOutputListInfo 网关列表信息
type GatewayOutputListInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"` // 网关名称
	// 网关control模式：1-direct 2-indirect
	Mode        uint8         `json:"mode" binding:"required,gatewayMode" enums:"1,2,3"`
	Maintainers []string      `json:"maintainers"` // 网关维护者
	Description string        `json:"description"` // 网关描述
	APISIX      common.APISIX `json:"apisix"`
	ReadOnly    bool          `json:"read_only"` // 是否只读
	Etcd        common.Etcd   `json:"etcd"`
	Count       Count         `json:"count"`
	CreatedAt   int64         `json:"created_at"`
	UpdatedAt   int64         `json:"updated_at"`
	Creator     string        `json:"creator"`
	Updater     string        `json:"updater"`
}

// GatewayGetRequest 网关详情请求
type GatewayGetRequest struct {
	GatewayID int `json:"gateway_id" uri:"gateway_id"  binding:"required"`
}

// CheckGatewayNameRequest 校验网关名称请求
type CheckGatewayNameRequest struct {
	Name string `json:"name" form:"name" binding:"required"` // 网关名称
	ID   int    `json:"id" form:"id"`                        // 网关id
}

// EtcdTestConnectionRequest 探测etcd连接请求
type EtcdTestConnectionRequest struct {
	GatewayID int `json:"gateway_id"` // 编辑下探测需要传
	common.EtcdConfig
}

// EtcdTestConOutputInfo 探测etcd连接输出信息
type EtcdTestConOutputInfo struct {
	APISIXVersion string `json:"apisix_version"` // apisix版本信息
}

// CheckGatewayNameResponse 校验网关名称返回
type CheckGatewayNameResponse struct {
	Status string `json:"status"`
}

// GatewayTagListResponse 网关标签列表
type GatewayTagListResponse []map[string]string

// GatewayLabelRequest 网关标签请求
type GatewayLabelRequest struct {
	Label base.LabelMap `json:"label"`
}

// CheckResourceStatus 校验资源状态
func CheckResourceStatus(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}
	for _, v := range strings.Split(value, ",") {
		if v == "" {
			continue
		}
		_, ok := constant.ResourceStatusMap[constant.ResourceStatus(v)]
		if !ok {
			return false
		}
	}
	return true
}

func init() {
	validation.AddBizFieldTagValidator(
		"resourceStatus",
		CheckResourceStatus,
		"{0}:{1} 状态必须是：【新增待发布，更新待发布，删除待发布，冲突，已发布】",
	)
}
