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
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

// MCPAccessTokenPathParam MCP 访问令牌路径参数
type MCPAccessTokenPathParam struct {
	GatewayID int `json:"gateway_id" uri:"gateway_id" binding:"required"`
	TokenID   int `json:"token_id" uri:"token_id"`
}

// MCPAccessTokenCreateRequest MCP 访问令牌创建请求
type MCPAccessTokenCreateRequest struct {
	Name        string               `json:"name" binding:"required,min=1,max=128"`
	Description string               `json:"description" binding:"max=512"`
	AccessScope model.MCPAccessScope `json:"access_scope" binding:"required,oneof=read write"`
	ExpiredAt   int64                `json:"expired_at" binding:"required"` // Unix timestamp
}

// MCPAccessTokenUpdateRequest MCP 访问令牌更新请求
type MCPAccessTokenUpdateRequest struct {
	Name        string               `json:"name" binding:"required,min=1,max=128"`
	Description string               `json:"description" binding:"max=512"`
	AccessScope model.MCPAccessScope `json:"access_scope" binding:"required,oneof=read write"`
	ExpiredAt   int64                `json:"expired_at" binding:"required"` // Unix timestamp
}

// MCPAccessTokenOutputInfo MCP 访问令牌输出信息
type MCPAccessTokenOutputInfo struct {
	ID          int                  `json:"id"`
	GatewayID   int                  `json:"gateway_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	AccessScope model.MCPAccessScope `json:"access_scope"`
	ExpiredAt   int64                `json:"expired_at"`   // Unix timestamp
	LastUsedAt  *int64               `json:"last_used_at"` // Unix timestamp, nullable
	CreatedAt   int64                `json:"created_at"`   // Unix timestamp
	UpdatedAt   int64                `json:"updated_at"`   // Unix timestamp
	Creator     string               `json:"creator"`
	Updater     string               `json:"updater"`
	MaskedToken string               `json:"masked_token"` // 掩码后的令牌
	IsExpired   bool                 `json:"is_expired"`   // 是否已过期
}

// MCPAccessTokenCreateOutputInfo MCP 访问令牌创建输出信息（包含完整令牌）
type MCPAccessTokenCreateOutputInfo struct {
	MCPAccessTokenOutputInfo
	Token string `json:"token"` // 完整令牌，仅在创建时返回
}

// MCPAccessTokenListResponse MCP 访问令牌列表响应
type MCPAccessTokenListResponse []MCPAccessTokenOutputInfo

// MCPAccessTokenToOutputInfo 将模型转换为输出信息
func MCPAccessTokenToOutputInfo(token *model.MCPAccessToken) MCPAccessTokenOutputInfo {
	var lastUsedAt *int64
	if token.LastUsedAt != nil {
		ts := token.LastUsedAt.Unix()
		lastUsedAt = &ts
	}

	return MCPAccessTokenOutputInfo{
		ID:          token.ID,
		GatewayID:   token.GatewayID,
		Name:        token.Name,
		Description: token.Description,
		AccessScope: token.AccessScope,
		ExpiredAt:   token.ExpiredAt.Unix(),
		LastUsedAt:  lastUsedAt,
		CreatedAt:   token.CreatedAt.Unix(),
		UpdatedAt:   token.UpdatedAt.Unix(),
		Creator:     token.Creator,
		Updater:     token.Updater,
		MaskedToken: token.MaskedToken(),
		IsExpired:   token.IsExpired(),
	}
}

// MCPAccessTokenToCreateOutputInfo 将模型转换为创建输出信息（包含完整令牌）
func MCPAccessTokenToCreateOutputInfo(token *model.MCPAccessToken) MCPAccessTokenCreateOutputInfo {
	return MCPAccessTokenCreateOutputInfo{
		MCPAccessTokenOutputInfo: MCPAccessTokenToOutputInfo(token),
		Token:                    token.Token,
	}
}
