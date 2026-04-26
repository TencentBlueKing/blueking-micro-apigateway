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
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// GlobalRule 表示数据库中的 global_rule 表
type GlobalRule struct {
	Name                string                 `gorm:"column:name;type:varchar(255);uniqueIndex:idx_name"` // 规则名称
	ResourceCommonModel                        // 资源通用 model: 创建时间、更新时间、创建人、更新人、config、status 等
	OperationType       constant.OperationType `gorm:"-"` // 用于标识操作类型，不持久化到数据库
}

// TableName 设置表名
func (GlobalRule) TableName() string {
	return "global_rule"
}

// BeforeCreate 创建前钩子
func (g *GlobalRule) BeforeCreate(tx *gorm.DB) (err error) {
	if err := g.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, g.GatewayID, g.ID, constant.GlobalRule, g.Config)
	if err != nil {
		return err
	}
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (g *GlobalRule) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := g.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, g.GatewayID, g.ID, constant.GlobalRule, g.Config)
	if err != nil {
		return err
	}
	// 如果更新的操作类型为撤销，则不触发审计
	if g.OperationType == constant.OperationTypeRevert {
		return nil
	}
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (g *GlobalRule) BeforeDelete(tx *gorm.DB) (err error) {
	if err := g.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (g *GlobalRule) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	// 排除批量删除，更新的情况
	if g.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate && g.ID != "" {
		// 获取原始数据
		var origin GlobalRule
		if err := tx.First(&origin, "id = ?", g.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		g.GatewayID, g.ID, g.Updater, g.Status, operation, constant.GlobalRule, originConfig, g.Config)
}

// HandleConfig 处理 config
func (g *GlobalRule) HandleConfig() (err error) {
	g.Config, err = stripResourceConfigForStorage(constant.GlobalRule, g.Config)
	return err
}

// AfterFind hydrates read-time config using authoritative global-rule columns.
func (g *GlobalRule) AfterFind(tx *gorm.DB) (err error) {
	g.Config, err = hydrateResourceConfigForRead(constant.GlobalRule, g.Config, g.ID, g.Name, nil)
	return err
}
