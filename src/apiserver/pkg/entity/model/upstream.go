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

// Upstream upstream 表
type Upstream struct {
	Name                string `gorm:"column:name;type:varchar(255);uniqueIndex:idx_name"` // upstream名称
	SSLID               string `gorm:"column:ssl_id;type:varchar(255)"`                    // ssl证书id
	ResourceCommonModel        // 资源通用model: 创建时间、更新时间、创建人、更新人、config、status等
}

// TableName 设置表名
func (Upstream) TableName() string {
	return "upstream"
}

// BeforeCreate 创建前钩子
func (u *Upstream) BeforeCreate(tx *gorm.DB) (err error) {
	if err := u.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return u.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (u *Upstream) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := u.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return u.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (u *Upstream) BeforeDelete(tx *gorm.DB) (err error) {
	if err := u.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return u.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (u *Upstream) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	if u.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate && u.ID != "" {
		// 获取原始数据
		var origin Upstream
		if err := tx.First(&origin, "id = ?", u.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		u.GatewayID, u.ID, u.Updater, u.Status, operation, constant.Upstream, originConfig, u.Config)
}

// HandleConfig 处理配置
func (u *Upstream) HandleConfig() (err error) {
	u.Config, err = sjson.SetBytes(u.Config, "id", u.ID)
	if err != nil {
		return err
	}

	if u.Name != "" {
		u.Config, err = sjson.SetBytes(u.Config, "name", u.Name)
		if err != nil {
			return err
		}
	}
	if u.SSLID != "" {
		u.Config, err = sjson.SetBytes(u.Config, "tls.client_cert_id", u.SSLID)
		if err != nil {
			return err
		}
	} else {
		u.Config, _ = sjson.DeleteBytes(u.Config, "tls.client_cert_id")
	}

	// 去除空字段
	config, err := jsonx.RemoveEmptyObjectsAndArrays(string(u.Config))
	if err == nil {
		u.Config = []byte(config)
	}
	return nil
}
