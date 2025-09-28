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
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/proto"
)

// Proto 表示数据库中的 proto 表
type Proto struct {
	// 文件名称
	Name                string                 `gorm:"column:name;type:varchar(255);uniqueIndex:idx_name" json:"name"`
	ResourceCommonModel                        // 资源通用model: 创建时间、更新时间、创建人、更新人、config、status等
	OperationType       constant.OperationType `gorm:"-"` // 用于标识操作类型，不持久化到数据库
}

// TableName 设置表名
func (Proto) TableName() string {
	return "proto"
}

// BeforeCreate 创建前钩子
func (p *Proto) BeforeCreate(tx *gorm.DB) (err error) {
	if err := p.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return p.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (p *Proto) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := p.HandleConfig(); err != nil {
		return err
	}
	// 如果更新的操作类型为撤销，则不触发审计
	if p.OperationType == constant.OperationTypeRevert {
		return nil
	}
	// 添加审计
	return p.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (p *Proto) BeforeDelete(tx *gorm.DB) (err error) {
	if err := p.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return p.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (p *Proto) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	// 排除批量删除，更新的情况
	if p.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate {
		// 获取原始数据
		var origin Proto
		if err := tx.First(&origin, "id = ?", p.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		p.GatewayID, p.ID, p.Updater, p.Status, operation, constant.Proto, originConfig, p.Config)
}

// HandleConfig 处理配置
func (p *Proto) HandleConfig() (err error) {
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
	// Remove empty fields
	config, err := jsonx.RemoveEmptyObjectsAndArrays(string(p.Config))
	if err == nil {
		p.Config = []byte(config)
	}
	content := gjson.GetBytes(p.Config, "content").String()
	err = proto.ParseContent(p.Name, content)
	if err != nil {
		return err
	}
	return nil
}
