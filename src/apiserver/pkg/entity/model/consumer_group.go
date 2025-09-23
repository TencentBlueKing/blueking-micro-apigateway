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

// ConsumerGroup 表示数据库中的 consumer_group 表
type ConsumerGroup struct {
	Name                string `gorm:"column:name;type:varchar(255);not null;uniqueIndex:idx_name"` // 消费group name
	ResourceCommonModel        // 资源通用model: 创建时间、更新时间、创建人、更新人、config、status等
}

// TableName 设置表名
func (ConsumerGroup) TableName() string {
	return "consumer_group"
}

// BeforeCreate 创建前钩子
func (c *ConsumerGroup) BeforeCreate(tx *gorm.DB) (err error) {
	if err := c.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, c.GatewayID, c.ID, constant.ConsumerGroup, c.Config)
	if err != nil {
		return err
	}
	// 添加审计
	return c.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (c *ConsumerGroup) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := c.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, c.GatewayID, c.ID, constant.ConsumerGroup, c.Config)
	if err != nil {
		return err
	}
	// 添加审计
	return c.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (c *ConsumerGroup) BeforeDelete(tx *gorm.DB) (err error) {
	if err := c.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return c.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (c *ConsumerGroup) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	// 排除批量删除，更新的情况
	if c.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate {
		// 获取原始数据
		var origin ConsumerGroup
		if err := tx.First(&origin, "id = ?", c.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		c.GatewayID, c.ID, c.Updater, c.Status, operation, constant.ConsumerGroup, originConfig, c.Config)
}

// HandleConfig 处理config
func (c *ConsumerGroup) HandleConfig() (err error) {
	c.Config, err = sjson.SetBytes(c.Config, "id", c.ID)
	if err != nil {
		return err
	}

	if c.Name != "" {
		c.Config, err = sjson.SetBytes(c.Config, "name", c.Name)
		if err != nil {
			return err
		}
	}
	// 去除空字段
	config, err := jsonx.RemoveEmptyObjectsAndArrays(string(c.Config))
	if err == nil {
		c.Config = []byte(config)
	}
	return nil
}
