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
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// MCPAccessTokenErrors 定义 MCP 访问令牌相关的错误
var (
	ErrMCPTokenNotFound       = errors.New("MCP access token not found")
	ErrMCPTokenExpired        = errors.New("MCP access token has expired")
	ErrMCPTokenInvalidScope   = errors.New("invalid MCP access token scope")
	ErrMCPGatewayNotSupported = errors.New("gateway does not support MCP (requires APISIX 3.13.X)")
	ErrMCPTokenNameExists     = errors.New("MCP access token name already exists")
	ErrMCPInsufficientScope   = errors.New("insufficient access scope for this operation")
	ErrMCPTokenLimitExceeded  = errors.New("maximum number of MCP access tokens per gateway exceeded (limit: 20)")
)

// MCPSupportedAPISIXVersion 支持 MCP 的 APISIX 版本
const MCPSupportedAPISIXVersion = constant.APISIXVersion313

// MaxMCPAccessTokensPerGateway 每个网关最大 MCP 访问令牌数量
const MaxMCPAccessTokensPerGateway = 20

// lastUsedUpdater handles batched updates for token last_used_at
var (
	lastUsedUpdateChan chan int
	lastUsedPending    map[int]time.Time
	lastUsedMu         sync.Mutex
	lastUsedOnce       sync.Once
)

// initLastUsedUpdater initializes the background goroutine for batched lastUsedAt updates
func initLastUsedUpdater() {
	lastUsedOnce.Do(func() {
		lastUsedUpdateChan = make(chan int, 1000)
		lastUsedPending = make(map[int]time.Time)
		go lastUsedUpdaterLoop()
	})
}

// lastUsedUpdaterLoop runs the background loop that flushes lastUsedAt updates every 5 minutes
func lastUsedUpdaterLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case tokenID := <-lastUsedUpdateChan:
			lastUsedMu.Lock()
			lastUsedPending[tokenID] = time.Now()
			lastUsedMu.Unlock()
		case <-ticker.C:
			flushLastUsedUpdates()
		}
	}
}

// flushLastUsedUpdates writes all pending lastUsedAt updates to the database
func flushLastUsedUpdates() {
	lastUsedMu.Lock()
	if len(lastUsedPending) == 0 {
		lastUsedMu.Unlock()
		return
	}
	// Copy and clear
	pending := lastUsedPending
	lastUsedPending = make(map[int]time.Time)
	lastUsedMu.Unlock()

	// Batch update
	for tokenID, lastUsed := range pending {
		err := database.Client().
			Model(&model.MCPAccessToken{}).
			Where("id = ?", tokenID).
			Update("last_used_at", lastUsed).Error
		if err != nil {
			logging.Errorf(
				"failed to batch update MCP token last_used_at for token ID %d: %v",
				tokenID,
				err,
			)
		}
	}
	logging.Infof("flushed %d MCP token last_used_at updates", len(pending))
}

// QueueLastUsedUpdate queues a token ID for batched lastUsedAt update
func QueueLastUsedUpdate(tokenID int) {
	initLastUsedUpdater()
	select {
	case lastUsedUpdateChan <- tokenID:
		// Successfully queued
	default:
		// Channel full, skip this update (not critical)
		logging.Warnf("lastUsedUpdateChan full, skipping update for token ID %d", tokenID)
	}
}

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

	// 检查是否超过最大令牌数量限制
	count, err := CountMCPAccessTokensByGateway(ctx, token.GatewayID)
	if err != nil {
		return err
	}
	if count >= MaxMCPAccessTokensPerGateway {
		return ErrMCPTokenLimitExceeded
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

	// Queue lastUsedAt update for batched processing (every 5 minutes)
	// This avoids frequent DB updates on every request
	QueueLastUsedUpdate(token.ID)

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

// AddMCPAccessTokenAuditLog adds audit log for MCP access token operations
func AddMCPAccessTokenAuditLog(
	ctx context.Context,
	operationType constant.OperationType,
	token *model.MCPAccessToken,
) error {
	if token == nil {
		return nil
	}
	gateway := ginx.GetGatewayInfoFromContext(ctx)
	if gateway == nil {
		return errors.New("gateway not found in context")
	}

	config, err := buildMCPAccessTokenAuditConfig(token)
	if err != nil {
		return err
	}

	var dataBefore []model.BatchOperationData
	var dataAfter []model.BatchOperationData
	tokenID := strconv.Itoa(token.ID)
	if operationType != constant.OperationTypeCreate {
		dataBefore = append(dataBefore, model.BatchOperationData{
			ID:     tokenID,
			Status: "",
			Config: config,
		})
	}
	if operationType != constant.OperationTypeDelete {
		dataAfter = append(dataAfter, model.BatchOperationData{
			ID:     tokenID,
			Status: "",
			Config: config,
		})
	}

	dataBeforeRaw, err := json.Marshal(dataBefore)
	if err != nil {
		return err
	}
	dataAfterRaw, err := json.Marshal(dataAfter)
	if err != nil {
		return err
	}

	operationAuditLog := &model.OperationAuditLog{
		GatewayID:     gateway.ID,
		ResourceType:  constant.Gateway,
		OperationType: operationType,
		ResourceIDs:   tokenID,
		DataBefore:    dataBeforeRaw,
		DataAfter:     dataAfterRaw,
		Operator:      ginx.GetUserIDFromContext(ctx),
	}
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).OperationAuditLog.WithContext(ctx).Create(operationAuditLog)
	}
	return repo.OperationAuditLog.WithContext(ctx).Create(operationAuditLog)
}

func buildMCPAccessTokenAuditConfig(token *model.MCPAccessToken) (json.RawMessage, error) {
	var lastUsedAt *int64
	if token.LastUsedAt != nil {
		ts := token.LastUsedAt.Unix()
		lastUsedAt = &ts
	}

	payload := map[string]any{
		"id":           token.ID,
		"gateway_id":   token.GatewayID,
		"name":         token.Name,
		"description":  token.Description,
		"access_scope": token.AccessScope,
		"expired_at":   token.ExpiredAt.Unix(),
		"last_used_at": lastUsedAt,
		"created_at":   token.CreatedAt.Unix(),
		"updated_at":   token.UpdatedAt.Unix(),
		"creator":      token.Creator,
		"updater":      token.Updater,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return data, nil
}
