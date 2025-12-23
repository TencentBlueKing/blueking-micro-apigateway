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
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// SSLCheckRequest ...
type SSLCheckRequest struct {
	ID   string `json:"id"`
	Name string `json:"name" binding:"required" validate:"sslName"` // 证书名称
	Cert string `json:"cert" binding:"required"`                    // ca证书
	Key  string `json:"key" binding:"required"`                     // 证书私钥
}

// SSLInfo SSL 基本信息
type SSLInfo struct {
	AutoID int             `json:"-"`                                                       // 自增ID
	ID     string          `json:"-"`                                                       // 资源apisix资源id
	Name   string          `json:"name" binding:"required" validate:"sslName"`              // 证书名称
	Config json.RawMessage `json:"config" validate:"apisixConfig=ssl" swaggertype:"object"` // 配置数据(json格式)
}

// ToEntity This function takes an SSLInfo struct and returns an SSL entity struct
func (s *SSLInfo) ToEntity() (*entity.SSL, error) {
	// Create a new SSL entity struct
	var ssl entity.SSL
	// Unmarshal the Config field of the SSLInfo struct into the SSL entity struct
	err := json.Unmarshal(s.Config, &ssl)
	if err != nil {
		return nil, err
	}
	return &ssl, nil
}

// SSLCheckResponse ...
type SSLCheckResponse struct {
	Name string `json:"name" `
	entity.SSL
}

// SSLListRequest ...
type SSLListRequest struct {
	ID      string `json:"id,omitempty" form:"id"`
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	Label   string `json:"label" form:"label"`
	Status  string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// SSLDropDownListResponse ...
type SSLDropDownListResponse []SSLDropDownOutputInfo

// SSLDropDownOutputInfo ...
type SSLDropDownOutputInfo struct {
	AutoID int    `json:"auto_id"` // 自增 ID
	ID     string `json:"id"`      // 资源 apisix 资源 id
	Name   string `json:"name"`    // 证书名称
}

// SSLListResponse sls 列表
type SSLListResponse []SSLOutputInfo

// SSLOutputInfo ...
type SSLOutputInfo struct {
	AutoID    int    `json:"auto_id"`
	GatewayID int    `json:"gateway_id"` // 网关 ID
	ID        string `json:"id"`
	SSLInfo
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// ValidateSSLID 校验 证书ID
func ValidateSSLID(ctx context.Context, fl validator.FieldLevel) bool {
	sslID := fl.Field().String()
	if sslID == "" {
		return true
	}
	return biz.ExistsSSL(ctx, sslID)
}

// ValidationSSLName 校验证书名称
func ValidationSSLName(ctx context.Context, fl validator.FieldLevel) bool {
	routeName := fl.Field().String()
	if routeName == "" {
		return false
	}
	return !biz.DuplicatedResourceName(
		ctx,
		constant.SSL,
		fl.Parent().FieldByName("ID").String(),
		routeName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx("sslID", ValidateSSLID, "{0}:{1} 无效")
	validation.AddBizFieldTagValidatorWithCtx(
		"sslName",
		ValidationSSLName,
		"{0}: {1} 该资源名称已经被存在的 ssl 资源占用",
	)
}
