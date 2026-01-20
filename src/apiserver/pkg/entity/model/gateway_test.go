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

package model_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

const (
	testEncryptKey = "AES256Key-32Characters1234567890"
	nonceSize      = 12 // AES-GCM nonce size
)

// initTestEnvironment initializes the test database and cryptography
func initTestEnvironment() {
	// Initialize database
	util.InitEmbedDb()

	// Initialize cryptography with test key and nonce (12 bytes padded)
	timestamp := strconv.Itoa(int(time.Now().UTC().Unix()))
	// Pad timestamp to 12 bytes for AES-GCM nonce
	nonce := timestamp
	for len(nonce) < nonceSize {
		nonce += "0"
	}
	if len(nonce) > nonceSize {
		nonce = nonce[:nonceSize]
	}
	err := cryptography.Init(testEncryptKey, nonce)
	if err != nil {
		panic(err)
	}
}

// TestGatewayAuditSkipForSyncOperation 测试同步操作不记录审计日志
func TestGatewayAuditSkipForSyncOperation(t *testing.T) {
	// 初始化测试环境（数据库和加密）
	initTestEnvironment()

	// 创建测试网关
	gateway := data.Gateway1WithBkAPISIX()
	gateway.Name = "test-gateway-audit"
	gateway.Updater = "test-user"

	// 使用 GORM 创建网关（会触发 AfterCreate hook）
	db := database.Client()
	err := db.Create(&gateway).Error
	assert.NoError(t, err, "Failed to create gateway")
	assert.NotZero(t, gateway.ID, "Gateway ID should be set after creation")

	// 获取创建时的审计记录数量（应该有1条创建审计）
	var initialAuditCount int64
	err = db.Model(&model.OperationAuditLog{}).Where("gateway_id = ?", gateway.ID).Count(&initialAuditCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), initialAuditCount, "Should have exactly 1 audit log after creation")

	// 验证初始审计记录是创建操作
	var createAuditLog model.OperationAuditLog
	err = db.Where("gateway_id = ?", gateway.ID).First(&createAuditLog).Error
	assert.NoError(t, err)
	assert.Equal(
		t,
		constant.OperationTypeCreate,
		createAuditLog.OperationType,
		"First audit log should be CREATE operation",
	)

	// Test Case 1: 更新 last_synced_at 字段（模拟同步操作）- 不应该创建审计记录
	t.Run("Update only last_synced_at - should NOT create audit log", func(t *testing.T) {
		// 使用 db.Model().Select().Updates() 更新（模拟实际的同步操作）
		updateData := map[string]any{
			"last_synced_at": time.Now(),
		}

		err := db.Model(
			&model.Gateway{},
		).Where(
			"id = ?",
			gateway.ID,
		).Select(
			"last_synced_at",
		).Updates(
			updateData,
		).Error
		assert.NoError(t, err, "Failed to update last_synced_at")

		// 验证审计记录数量没有增加（仍然是1条创建审计）
		var auditCountAfterSync int64
		err = db.Model(
			&model.OperationAuditLog{},
		).Where(
			"gateway_id = ?",
			gateway.ID,
		).Count(
			&auditCountAfterSync,
		).Error
		assert.NoError(t, err)
		assert.Equal(
			t,
			initialAuditCount,
			auditCountAfterSync,
			"Audit log count should NOT increase after sync operation",
		)
	})

	// Test Case 2: 更新其他字段（正常更新操作）- 应该创建审计记录
	t.Run("Update name field - should create audit log", func(t *testing.T) {
		// 记录更新前的审计数量
		var auditCountBeforeUpdate int64
		err = db.Model(
			&model.OperationAuditLog{},
		).Where(
			"gateway_id = ?",
			gateway.ID,
		).Count(
			&auditCountBeforeUpdate,
		).Error
		assert.NoError(t, err)

		// 更新网关名称
		gateway.Name = "test-gateway-audit-updated"
		err = db.Model(&gateway).Updates(map[string]any{
			"name": gateway.Name,
		}).Error
		assert.NoError(t, err, "Failed to update gateway name")

		// 验证审计记录数量增加了1条
		var auditCountAfterUpdate int64
		err = db.Model(
			&model.OperationAuditLog{},
		).Where(
			"gateway_id = ?",
			gateway.ID,
		).Count(
			&auditCountAfterUpdate,
		).Error
		assert.NoError(t, err)
		assert.Equal(
			t,
			auditCountBeforeUpdate+1,
			auditCountAfterUpdate,
			"Audit log count should increase by 1 after normal update",
		)

		// 验证最新的审计记录是更新操作
		var latestAuditLog model.OperationAuditLog
		err = db.Where("gateway_id = ?", gateway.ID).Order("created_at DESC").First(&latestAuditLog).Error
		assert.NoError(t, err)
		assert.Equal(
			t,
			constant.OperationTypeUpdate,
			latestAuditLog.OperationType,
			"Latest audit log should be UPDATE operation",
		)
		assert.Equal(t, "test-user", latestAuditLog.Operator, "Operator should be set correctly")
	})

	// Test Case 3: 同时更新 last_synced_at 和其他字段 - 应该创建审计记录
	t.Run("Update both last_synced_at and name - should create audit log", func(t *testing.T) {
		// 记录更新前的审计数量
		var auditCountBeforeUpdate int64
		err = db.Model(
			&model.OperationAuditLog{},
		).Where(
			"gateway_id = ?",
			gateway.ID,
		).Count(
			&auditCountBeforeUpdate,
		).Error
		assert.NoError(t, err)

		// 同时更新多个字段
		gateway.Name = "test-gateway-audit-updated-again"
		gateway.LastSyncedAt = time.Now()
		err = db.Model(&gateway).Updates(map[string]any{
			"name":           gateway.Name,
			"last_synced_at": gateway.LastSyncedAt,
		}).Error
		assert.NoError(t, err, "Failed to update gateway")

		// 验证审计记录数量增加了1条（因为不只是更新 last_synced_at）
		var auditCountAfterUpdate int64
		err = db.Model(
			&model.OperationAuditLog{},
		).Where(
			"gateway_id = ?",
			gateway.ID,
		).Count(
			&auditCountAfterUpdate,
		).Error
		assert.NoError(t, err)
		assert.Equal(
			t,
			auditCountBeforeUpdate+1,
			auditCountAfterUpdate,
			"Audit log count should increase by 1 when updating multiple fields",
		)
	})

	// Test Case 4: 使用 Select 指定多个字段 - 应该创建审计记录
	t.Run("Update with Select multiple fields - should create audit log", func(t *testing.T) {
		// 记录更新前的审计数量
		var auditCountBeforeUpdate int64
		err = db.Model(
			&model.OperationAuditLog{},
		).Where(
			"gateway_id = ?",
			gateway.ID,
		).Count(
			&auditCountBeforeUpdate,
		).Error
		assert.NoError(t, err)

		// 先获取当前网关数据
		var currentGateway model.Gateway
		err := db.First(&currentGateway, gateway.ID).Error
		assert.NoError(t, err)

		// 更新多个字段
		currentGateway.Name = "test-gateway-audit-with-select"
		currentGateway.LastSyncedAt = time.Now()

		// 使用 Select 指定要更新的字段
		err = db.Model(&currentGateway).Select("name", "last_synced_at").Updates(currentGateway).Error
		assert.NoError(t, err, "Failed to update with Select")

		// 验证审计记录数量增加了1条
		var auditCountAfterUpdate int64
		err = db.Model(
			&model.OperationAuditLog{},
		).Where(
			"gateway_id = ?",
			gateway.ID,
		).Count(
			&auditCountAfterUpdate,
		).Error
		assert.NoError(t, err)
		assert.Equal(
			t,
			auditCountBeforeUpdate+1,
			auditCountAfterUpdate,
			"Audit log count should increase by 1 when selecting multiple fields",
		)
	})
}

// TestGatewayAuditForDeleteOperation 测试删除操作记录审计日志
func TestGatewayAuditForDeleteOperation(t *testing.T) {
	// 初始化测试环境（数据库和加密）
	initTestEnvironment()

	// 创建测试网关
	gateway := data.Gateway1WithBkAPISIX()
	gateway.Name = "test-gateway-delete"
	gateway.Updater = "test-user"

	db := database.Client()
	err := db.Create(&gateway).Error
	assert.NoError(t, err, "Failed to create gateway")

	// 获取创建时的审计记录数量
	var auditCountBeforeDelete int64
	err = db.Model(
		&model.OperationAuditLog{},
	).Where(
		"gateway_id = ?",
		gateway.ID,
	).Count(
		&auditCountBeforeDelete,
	).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), auditCountBeforeDelete, "Should have 1 audit log after creation")

	// 删除网关
	err = db.Delete(&gateway).Error
	assert.NoError(t, err, "Failed to delete gateway")

	// 验证审计记录数量增加了1条
	var auditCountAfterDelete int64
	err = db.Model(
		&model.OperationAuditLog{},
	).Where(
		"gateway_id = ?",
		gateway.ID,
	).Count(
		&auditCountAfterDelete,
	).Error
	assert.NoError(t, err)
	assert.Equal(
		t,
		auditCountBeforeDelete+1,
		auditCountAfterDelete,
		"Audit log count should increase by 1 after delete",
	)

	// 验证最新的审计记录是删除操作
	var deleteAuditLog model.OperationAuditLog
	err = db.Where("gateway_id = ?", gateway.ID).Order("created_at DESC").First(&deleteAuditLog).Error
	assert.NoError(t, err)
	assert.Equal(
		t,
		constant.OperationTypeDelete,
		deleteAuditLog.OperationType,
		"Latest audit log should be DELETE operation",
	)
}
