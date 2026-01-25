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

package biz

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

// MCPAccessTokenErrors 定义 MCP 访问令牌相关的错误
var (
	ErrMCPTokenNotFound       = errors.New("MCP access token not found")
	ErrMCPTokenExpired        = errors.New("MCP access token has expired")
	ErrMCPTokenInvalidScope   = errors.New("invalid MCP access token scope")
	ErrMCPGatewayNotSupported = errors.New("gateway does not support MCP (requires APISIX 3.13.X)")
	ErrMCPTokenNameExists     = errors.New("MCP access token name already exists")
	ErrMCPInsufficientScope   = errors.New("insufficient access scope for this operation")
)

// MCPSupportedAPISIXVersion 支持 MCP 的 APISIX 版本
const MCPSupportedAPISIXVersion = constant.APISIXVersion313

// GenerateMCPToken 生成随机的 MCP 访问令牌（32字节 = 64字符十六进制）
func GenerateMCPToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashMCPToken 使用 SHA-256 对令牌进行哈希
// 用于安全存储，避免明文令牌泄露
func HashMCPToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CheckGatewayMCPSupport 检查网关是否支持 MCP
func CheckGatewayMCPSupport(gateway *model.Gateway) error {
	if gateway.GetAPISIXVersionX() != MCPSupportedAPISIXVersion {
		return ErrMCPGatewayNotSupported
	}
	return nil
}

// ListMCPAccessTokens 列出网关的所有 MCP 访问令牌
func ListMCPAccessTokens(ctx context.Context, gatewayID int) ([]*model.MCPAccessToken, error) {
	var tokens []*model.MCPAccessToken
	err := database.Client().WithContext(ctx).
		Where("gateway_id = ?", gatewayID).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// GetMCPAccessToken 根据 ID 获取 MCP 访问令牌
func GetMCPAccessToken(ctx context.Context, id int) (*model.MCPAccessToken, error) {
	var token model.MCPAccessToken
	err := database.Client().WithContext(ctx).
		Where("id = ?", id).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMCPTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

// GetMCPAccessTokenByToken 根据令牌字符串获取 MCP 访问令牌
// 注意：输入的是原始令牌，函数会自动哈希后查询
func GetMCPAccessTokenByToken(ctx context.Context, token string) (*model.MCPAccessToken, error) {
	// 对输入令牌进行哈希，与数据库中存储的哈希值比较
	hashedToken := HashMCPToken(token)

	var accessToken model.MCPAccessToken
	err := database.Client().WithContext(ctx).
		Where("token = ?", hashedToken).
		First(&accessToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMCPTokenNotFound
		}
		return nil, err
	}
	return &accessToken, nil
}

// GetMCPAccessTokenByGatewayAndID 根据网关 ID 和令牌 ID 获取 MCP 访问令牌
func GetMCPAccessTokenByGatewayAndID(ctx context.Context, gatewayID, tokenID int) (*model.MCPAccessToken, error) {
	var token model.MCPAccessToken
	err := database.Client().WithContext(ctx).
		Where("gateway_id = ? AND id = ?", gatewayID, tokenID).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMCPTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

// CreateMCPAccessToken 创建新的 MCP 访问令牌
// 注意：创建成功后，token.Token 包含原始令牌（仅此一次可见），数据库中存储的是哈希值
func CreateMCPAccessToken(ctx context.Context, token *model.MCPAccessToken) error {
	// 验证访问范围
	if !token.AccessScope.IsValid() {
		return ErrMCPTokenInvalidScope
	}

	// 检查名称是否已存在
	exists, err := MCPAccessTokenNameExists(ctx, token.GatewayID, token.Name, 0)
	if err != nil {
		return err
	}
	if exists {
		return ErrMCPTokenNameExists
	}

	// 生成令牌
	plainToken, err := GenerateMCPToken()
	if err != nil {
		return err
	}

	// 存储哈希值到数据库
	hashedToken := HashMCPToken(plainToken)
	token.Token = hashedToken

	if err := database.Client().WithContext(ctx).Create(token).Error; err != nil {
		return err
	}

	// 创建成功后，将原始令牌设置回 token.Token，供调用方返回给用户
	// 这是用户唯一一次能看到原始令牌的机会
	token.Token = plainToken

	return nil
}

// UpdateMCPAccessToken 更新 MCP 访问令牌
func UpdateMCPAccessToken(ctx context.Context, token *model.MCPAccessToken) error {
	// 验证访问范围
	if !token.AccessScope.IsValid() {
		return ErrMCPTokenInvalidScope
	}

	// 检查名称是否已存在（排除自身）
	exists, err := MCPAccessTokenNameExists(ctx, token.GatewayID, token.Name, token.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrMCPTokenNameExists
	}

	return database.Client().WithContext(ctx).
		Model(token).
		Select("name", "description", "access_scope", "expired_at", "updater", "updated_at").
		Updates(token).Error
}

// DeleteMCPAccessToken 删除 MCP 访问令牌
func DeleteMCPAccessToken(ctx context.Context, id int) error {
	result := database.Client().WithContext(ctx).
		Delete(&model.MCPAccessToken{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMCPTokenNotFound
	}
	return nil
}

// DeleteMCPAccessTokenByGateway 删除网关下的所有 MCP 访问令牌
func DeleteMCPAccessTokenByGateway(ctx context.Context, gatewayID int) error {
	return database.Client().WithContext(ctx).
		Where("gateway_id = ?", gatewayID).
		Delete(&model.MCPAccessToken{}).Error
}

// MCPAccessTokenNameExists 检查 MCP 访问令牌名称是否已存在
func MCPAccessTokenNameExists(ctx context.Context, gatewayID int, name string, excludeID int) (bool, error) {
	var count int64
	query := database.Client().WithContext(ctx).
		Model(&model.MCPAccessToken{}).
		Where("gateway_id = ? AND name = ?", gatewayID, name)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ValidateMCPAccessToken 验证 MCP 访问令牌
func ValidateMCPAccessToken(ctx context.Context, tokenStr string) (*model.MCPAccessToken, *model.Gateway, error) {
	// 获取令牌
	token, err := GetMCPAccessTokenByToken(ctx, tokenStr)
	if err != nil {
		return nil, nil, err
	}

	// 检查是否过期
	if token.IsExpired() {
		return nil, nil, ErrMCPTokenExpired
	}

	// 获取网关信息
	gateway, err := GetGateway(ctx, token.GatewayID)
	if err != nil {
		return nil, nil, err
	}

	// 检查网关是否支持 MCP
	if err := CheckGatewayMCPSupport(gateway); err != nil {
		return nil, nil, err
	}

	// 更新最后使用时间（异步，不阻塞请求）
	go func() {
		if err := UpdateMCPAccessTokenLastUsed(context.Background(), token.ID); err != nil {
			logging.Errorf("failed to update MCP token last_used_at for token ID %d: %v", token.ID, err)
		}
	}()

	return token, gateway, nil
}

// UpdateMCPAccessTokenLastUsed 更新令牌最后使用时间
func UpdateMCPAccessTokenLastUsed(ctx context.Context, id int) error {
	now := time.Now()
	return database.Client().WithContext(ctx).
		Model(&model.MCPAccessToken{}).
		Where("id = ?", id).
		Update("last_used_at", now).Error
}

// CheckMCPAccessScope 检查令牌是否有足够的访问权限
func CheckMCPAccessScope(token *model.MCPAccessToken, requireWrite bool) error {
	if requireWrite && !token.CanWrite() {
		return ErrMCPInsufficientScope
	}
	return nil
}

// CountMCPAccessTokensByGateway 统计网关的 MCP 访问令牌数量
func CountMCPAccessTokensByGateway(ctx context.Context, gatewayID int) (int64, error) {
	var count int64
	err := database.Client().WithContext(ctx).
		Model(&model.MCPAccessToken{}).
		Where("gateway_id = ?", gatewayID).
		Count(&count).Error
	return count, err
}
