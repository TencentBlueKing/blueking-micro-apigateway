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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestGenerateMCPToken(t *testing.T) {
	// 测试令牌生成
	token1, err := GenerateMCPToken()
	assert.NoError(t, err)
	assert.Len(t, token1, 64) // 32 bytes = 64 hex chars

	// 测试令牌唯一性
	token2, err := GenerateMCPToken()
	assert.NoError(t, err)
	assert.NotEqual(t, token1, token2)

	// 测试令牌只包含十六进制字符
	for _, c := range token1 {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"Token should only contain hex characters")
	}
}

func TestHashMCPToken(t *testing.T) {
	// 测试哈希一致性
	token := "test-token-12345"
	hash1 := HashMCPToken(token)
	hash2 := HashMCPToken(token)
	assert.Equal(t, hash1, hash2)

	// 测试哈希长度 (SHA-256 produces 32 bytes = 64 hex chars)
	assert.Len(t, hash1, 64)

	// 测试不同令牌产生不同哈希
	differentToken := "different-token-67890"
	hash3 := HashMCPToken(differentToken)
	assert.NotEqual(t, hash1, hash3)
}

func TestMCPAccessTokenCRUD(t *testing.T) {
	util.InitEmbedDb()
	ctx := context.Background()

	// 创建测试网关
	gateway := &model.Gateway{
		Name:          "test-mcp-gateway",
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := CreateGateway(ctx, gateway)
	assert.NoError(t, err)
	assert.Greater(t, gateway.ID, 0)

	// 创建令牌
	token := &model.MCPAccessToken{
		GatewayID:   gateway.ID,
		Name:        "test-token",
		Description: "Test MCP token",
		AccessScope: model.MCPAccessScopeRead,
		ExpiredAt:   time.Now().Add(24 * time.Hour),
		BaseModel: model.BaseModel{
			Creator: "tester",
			Updater: "tester",
		},
	}

	err = CreateMCPAccessToken(ctx, token)
	assert.NoError(t, err)
	assert.Greater(t, token.ID, 0)
	assert.Len(t, token.Token, 64) // Plain token returned

	// 保存原始令牌用于后续查询
	plainToken := token.Token

	// 按 ID 获取令牌
	retrieved, err := GetMCPAccessToken(ctx, token.ID)
	assert.NoError(t, err)
	assert.Equal(t, token.Name, retrieved.Name)
	assert.Equal(t, token.AccessScope, retrieved.AccessScope)

	// 按令牌字符串获取（验证哈希查询）
	retrievedByToken, err := GetMCPAccessTokenByToken(ctx, plainToken)
	assert.NoError(t, err)
	assert.Equal(t, token.ID, retrievedByToken.ID)

	// Note: UpdateMCPAccessToken was removed - tokens should be deleted and recreated

	// 列出令牌
	tokens, err := ListMCPAccessTokens(ctx, gateway.ID)
	assert.NoError(t, err)
	assert.Len(t, tokens, 1)

	// 删除令牌
	err = DeleteMCPAccessToken(ctx, token.ID)
	assert.NoError(t, err)

	// 验证删除
	_, err = GetMCPAccessToken(ctx, token.ID)
	assert.ErrorIs(t, err, ErrMCPTokenNotFound)
}

func TestMCPAccessTokenExpiration(t *testing.T) {
	// 测试未过期令牌
	token := &model.MCPAccessToken{
		ExpiredAt: time.Now().Add(1 * time.Hour),
	}
	assert.False(t, token.IsExpired())

	// 测试已过期令牌
	expiredToken := &model.MCPAccessToken{
		ExpiredAt: time.Now().Add(-1 * time.Hour),
	}
	assert.True(t, expiredToken.IsExpired())

	// 测试刚好过期
	justExpired := &model.MCPAccessToken{
		ExpiredAt: time.Now().Add(-1 * time.Second),
	}
	assert.True(t, justExpired.IsExpired())
}

func TestMCPAccessScope(t *testing.T) {
	// 测试读取权限
	readToken := &model.MCPAccessToken{
		AccessScope: model.MCPAccessScopeRead,
	}
	assert.True(t, readToken.CanRead())
	assert.False(t, readToken.CanWrite())

	// 测试写入权限（包含读取）
	writeToken := &model.MCPAccessToken{
		AccessScope: model.MCPAccessScopeReadWrite,
	}
	assert.True(t, writeToken.CanRead())
	assert.True(t, writeToken.CanWrite())

	// 测试权限检查
	err := CheckMCPAccessScope(readToken, false)
	assert.NoError(t, err)

	err = CheckMCPAccessScope(readToken, true)
	assert.ErrorIs(t, err, ErrMCPInsufficientScope)

	err = CheckMCPAccessScope(writeToken, true)
	assert.NoError(t, err)
}

func TestCheckGatewayMCPSupport(t *testing.T) {
	// 测试支持 MCP 的网关 (3.13)
	gateway313 := &model.Gateway{
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := CheckGatewayMCPSupport(gateway313)
	assert.NoError(t, err)

	// 测试不支持 MCP 的网关 (3.11)
	gateway311 := &model.Gateway{
		APISIXVersion: string(constant.APISIXVersion311),
	}
	err = CheckGatewayMCPSupport(gateway311)
	assert.ErrorIs(t, err, ErrMCPGatewayNotSupported)

	// 测试其他版本
	gatewayOther := &model.Gateway{
		APISIXVersion: "3.2.X",
	}
	err = CheckGatewayMCPSupport(gatewayOther)
	assert.ErrorIs(t, err, ErrMCPGatewayNotSupported)
}

func TestMCPAccessTokenNameExists(t *testing.T) {
	util.InitEmbedDb()
	ctx := context.Background()

	// 创建测试网关
	gateway := &model.Gateway{
		Name:          "test-gateway-name-exists",
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := CreateGateway(ctx, gateway)
	assert.NoError(t, err)

	// 创建令牌
	token := &model.MCPAccessToken{
		GatewayID:   gateway.ID,
		Name:        "unique-token-name",
		AccessScope: model.MCPAccessScopeRead,
		ExpiredAt:   time.Now().Add(24 * time.Hour),
		BaseModel: model.BaseModel{
			Creator: "tester",
			Updater: "tester",
		},
	}
	err = CreateMCPAccessToken(ctx, token)
	assert.NoError(t, err)

	// 测试名称存在
	exists, err := MCPAccessTokenNameExists(ctx, gateway.ID, "unique-token-name", 0)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试名称不存在
	exists, err = MCPAccessTokenNameExists(ctx, gateway.ID, "non-existent-name", 0)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试排除自身
	exists, err = MCPAccessTokenNameExists(ctx, gateway.ID, "unique-token-name", token.ID)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试创建重复名称失败
	duplicateToken := &model.MCPAccessToken{
		GatewayID:   gateway.ID,
		Name:        "unique-token-name",
		AccessScope: model.MCPAccessScopeRead,
		ExpiredAt:   time.Now().Add(24 * time.Hour),
		BaseModel: model.BaseModel{
			Creator: "tester",
			Updater: "tester",
		},
	}
	err = CreateMCPAccessToken(ctx, duplicateToken)
	assert.ErrorIs(t, err, ErrMCPTokenNameExists)
}

func TestMCPAccessTokenScopeValidation(t *testing.T) {
	util.InitEmbedDb()
	ctx := context.Background()

	// 创建测试网关
	gateway := &model.Gateway{
		Name:          "test-gateway-scope",
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := CreateGateway(ctx, gateway)
	assert.NoError(t, err)

	// 测试无效的访问范围
	token := &model.MCPAccessToken{
		GatewayID:   gateway.ID,
		Name:        "invalid-scope-token",
		AccessScope: model.MCPAccessScope("invalid"),
		ExpiredAt:   time.Now().Add(24 * time.Hour),
		BaseModel: model.BaseModel{
			Creator: "tester",
			Updater: "tester",
		},
	}
	err = CreateMCPAccessToken(ctx, token)
	assert.ErrorIs(t, err, ErrMCPTokenInvalidScope)
}

func TestMaskedToken(t *testing.T) {
	// 测试正常长度令牌
	token := &model.MCPAccessToken{
		Token: "abcdefgh12345678ijklmnop90qrstuv",
	}
	masked := token.MaskedToken()
	assert.Equal(t, "abcdefgh****stuv", masked)

	// 测试短令牌
	shortToken := &model.MCPAccessToken{
		Token: "short",
	}
	masked = shortToken.MaskedToken()
	assert.Equal(t, "****", masked)
}
