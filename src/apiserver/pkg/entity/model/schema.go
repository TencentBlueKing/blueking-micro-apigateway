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
	"encoding/json"
	"strconv"

	"github.com/tidwall/gjson"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// GatewayCustomPluginSchema 表示数据库中的 gateway_custom_plugin_schema 表
type GatewayCustomPluginSchema struct {
	AutoID    int            `gorm:"column:auto_id;type:int;primaryKey;autoIncrement"`   // 自增ID
	GatewayID int            `gorm:"column:gateway_id;type:int;uniqueIndex:idx_name"`    // 网关ID
	Name      string         `gorm:"column:name;type:varchar(255);uniqueIndex:idx_name"` // 插件名称
	Schema    datatypes.JSON `gorm:"column:schema;type:json"`                            // schema
	Example   datatypes.JSON `gorm:"column:example;type:json"`                           // example
	BaseModel
}

// TableName 设置表名
func (GatewayCustomPluginSchema) TableName() string {
	return "gateway_custom_plugin_schema"
}

// GatewayResourceSchemaAssociation 表示数据库中的 gateway_resource_schema_association 表
type GatewayResourceSchemaAssociation struct {
	AutoID     int    `gorm:"column:auto_id;type:int;primaryKey;autoIncrement"`            // 自增ID
	GatewayID  int    `gorm:"column:gateway_id;type:int"`                                  // 网关ID
	SchemaID   int    `gorm:"column:schema_id;type:int;uniqueIndex:idx_schema"`            // 自定义插件ID
	ResourceID string `gorm:"column:resource_id;type:varchar(255);uniqueIndex:idx_schema"` // 资源id
	// route/service/upstream
	ResourceType constant.APISIXResource `gorm:"column:resource_type"`
}

// TableName 设置表名
func (GatewayResourceSchemaAssociation) TableName() string {
	return "gateway_resource_schema_association"
}

// AfterCreate 创建后钩子
func (g *GatewayCustomPluginSchema) AfterCreate(tx *gorm.DB) (err error) {
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (g *GatewayCustomPluginSchema) BeforeUpdate(tx *gorm.DB) (err error) {
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (g *GatewayCustomPluginSchema) BeforeDelete(tx *gorm.DB) (err error) {
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeDelete)
}

// CopyCustomPluginSchema 复制自定义插件的结构体
func (g *GatewayCustomPluginSchema) CopyCustomPluginSchema() GatewayCustomPluginSchema {
	schema := GatewayCustomPluginSchema{
		AutoID:  g.AutoID,
		Name:    g.Name,
		Schema:  g.Schema,
		Example: g.Example,
	}
	return schema
}

// AddAuditLog 添加审计
func (g *GatewayCustomPluginSchema) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	updater := g.Updater
	dataAfter := datatypes.JSON{}
	if operation != constant.OperationTypeDelete {
		b, err := json.Marshal(g.CopyCustomPluginSchema())
		if err != nil {
			return err
		}
		dataAfter = b
	}

	dataBefore := datatypes.JSON{}
	if operation != constant.OperationTypeCreate {
		// 获取原始数据
		var origin GatewayCustomPluginSchema
		if err := tx.First(&origin, "auto_id = ?", g.AutoID).Error; err != nil {
			return err
		}
		if updater == "" {
			updater = origin.Updater
		}
		b, err := json.Marshal(origin.CopyCustomPluginSchema())
		if err != nil {
			return err
		}
		dataBefore = b
	}
	return auditCallback(tx,
		g.GatewayID, strconv.Itoa(g.AutoID), updater, "", operation, constant.Schema, dataBefore, dataAfter)
}

// ResourceSchemaCallback 资源与自定义插件的关联回调
func ResourceSchemaCallback(tx *gorm.DB, gatewayID int,
	resourceID string, resourceType constant.APISIXResource, resourceConfig datatypes.JSON,
) error {
	var plugins map[string]interface{}
	if resourceType == constant.PluginMetadata {
		pluginName := gjson.GetBytes(resourceConfig, "name").String()
		plugins = map[string]interface{}{
			pluginName: resourceConfig,
		}
	} else {
		pluginsRaw := gjson.GetBytes(resourceConfig, "plugins").Raw
		if pluginsRaw != "" {
			err := json.Unmarshal([]byte(pluginsRaw), &plugins)
			if err != nil {
				return err
			}
		}
	}
	// 先删除当前资源与自定义插件的关联记录
	err := tx.Where("resource_id = ? AND resource_type = ?", resourceID, resourceType).
		Delete(&GatewayResourceSchemaAssociation{}).Error
	if err != nil {
		return err
	}
	if plugins == nil {
		return nil
	}
	var pluginNames []string
	for name := range plugins {
		pluginNames = append(pluginNames, name)
	}
	// 查询当前资源绑定的所有自定义插件
	var schemaList []*GatewayCustomPluginSchema
	if err := tx.Find(&schemaList, "gateway_id = ? AND name IN (?)", gatewayID, pluginNames).Error; err != nil {
		return err
	}
	if len(schemaList) == 0 {
		return nil
	}
	var resourceSchemaList []*GatewayResourceSchemaAssociation
	for _, s := range schemaList {
		resourceSchemaList = append(resourceSchemaList, &GatewayResourceSchemaAssociation{
			GatewayID:    gatewayID,
			SchemaID:     s.AutoID,
			ResourceID:   resourceID,
			ResourceType: resourceType,
		})
	}
	// 再创建所有的关联记录
	err = tx.Create(&resourceSchemaList).Error
	if err != nil {
		return err
	}
	return nil
}
