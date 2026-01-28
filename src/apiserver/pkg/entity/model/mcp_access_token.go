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

package model

import (
	"time"
)

// MCPAccessScope MCP 访问范围类型
type MCPAccessScope string

const (
	// MCPAccessScopeRead 只读权限：允许 GET 请求和 sync 操作
	MCPAccessScopeRead MCPAccessScope = "read"
	// MCPAccessScopeWrite 读写权限：允许所有操作包括 POST/PUT/DELETE
	MCPAccessScopeWrite MCPAccessScope = "write"
)

// String 返回访问范围字符串
func (s MCPAccessScope) String() string {
	return string(s)
}

// IsValid 检查访问范围是否有效
func (s MCPAccessScope) IsValid() bool {
	return s == MCPAccessScopeRead || s == MCPAccessScopeWrite
}

// MCPAccessToken MCP 访问令牌表
type MCPAccessToken struct {
	ID          int            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GatewayID   int            `gorm:"column:gateway_id;type:int;not null;index:idx_gateway" json:"gateway_id"`
	Token       string         `gorm:"column:token;type:varchar(64);uniqueIndex:idx_token" json:"-"` // 不在 JSON 中返回完整 token
	Name        string         `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Description string         `gorm:"column:description;type:varchar(512)" json:"description"`
	AccessScope MCPAccessScope `gorm:"column:access_scope;type:varchar(16);not null" json:"access_scope"`
	ExpiredAt   time.Time      `gorm:"column:expired_at;type:datetime;not null" json:"expired_at"`
	LastUsedAt  *time.Time     `gorm:"column:last_used_at;type:datetime" json:"last_used_at"`
	BaseModel
}

// TableName 返回表名
func (MCPAccessToken) TableName() string {
	return "mcp_access_token"
}

// IsExpired 检查令牌是否已过期
func (t *MCPAccessToken) IsExpired() bool {
	return time.Now().After(t.ExpiredAt)
}

// CanRead 检查是否有读权限
func (t *MCPAccessToken) CanRead() bool {
	return t.AccessScope == MCPAccessScopeRead || t.AccessScope == MCPAccessScopeWrite
}

// CanWrite 检查是否有写权限
func (t *MCPAccessToken) CanWrite() bool {
	return t.AccessScope == MCPAccessScopeWrite
}

// MaskedToken 返回掩码后的令牌（仅显示前8位和后4位）
func (t *MCPAccessToken) MaskedToken() string {
	if len(t.Token) <= 12 {
		return "****"
	}
	return t.Token[:8] + "****" + t.Token[len(t.Token)-4:]
}

// UpdateLastUsed 更新最后使用时间
func (t *MCPAccessToken) UpdateLastUsed() {
	now := time.Now()
	t.LastUsedAt = &now
}
