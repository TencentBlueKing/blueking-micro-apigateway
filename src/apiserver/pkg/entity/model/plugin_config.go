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

package model

import (
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
)

// PluginConfig  plugin_config 表
type PluginConfig struct {
	Name                string `gorm:"column:name;type:varchar(255);uniqueIndex:idx_name"` // 配置策略名称
	ResourceCommonModel        // 资源通用model: 创建时间、更新时间、创建人、更新人、config、status等
}

// TableName 设置表名
func (PluginConfig) TableName() string {
	return "plugin_config"
}

// BeforeCreate 创建前钩子
func (p *PluginConfig) BeforeCreate(tx *gorm.DB) (err error) {
	if err := p.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, p.GatewayID, p.ID, constant.PluginConfig, p.Config)
	if err != nil {
		return err
	}
	// 添加审计
	return p.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (p *PluginConfig) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := p.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, p.GatewayID, p.ID, constant.PluginConfig, p.Config)
	if err != nil {
		return err
	}
	// 添加审计
	return p.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (p *PluginConfig) BeforeDelete(tx *gorm.DB) (err error) {
	if err := p.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return p.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (p *PluginConfig) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	// 排除批量删除，更新的情况
	if p.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate && p.ID != "" {
		// 获取原始数据
		var origin PluginConfig
		if err := tx.First(&origin, "id = ?", p.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		p.GatewayID, p.ID, p.Updater, p.Status, operation, constant.PluginConfig, originConfig, p.Config)
}

// HandleConfig 处理配置
func (p *PluginConfig) HandleConfig() (err error) {
	p.Config, err = sjson.SetBytes(p.Config, "id", p.ID)
	if err != nil {
		return err
	}

	if p.Name != "" {
		p.Config, err = sjson.SetBytes(p.Config, "name", p.Name)
		if err != nil {
			return err
		}
	}
	// 去除空字段
	config, err := jsonx.RemoveEmptyObjectsAndArrays(string(p.Config))
	if err == nil {
		p.Config = []byte(config)
	}
	return nil
}
