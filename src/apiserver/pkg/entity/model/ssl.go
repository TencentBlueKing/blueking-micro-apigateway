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

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/sslx"
)

// SSL ...
type SSL struct {
	Name string `gorm:"column:name;type:varchar(255);uniqueIndex:idx_name" json:"name"` // 证书名称
	// 资源通用model: 创建时间、更新时间、创建人、更新人、config、status等
	ResourceCommonModel
}

// TableName 设置表名
func (SSL) TableName() string {
	return "ssl"
}

// ToEntity 转换为实体
func (s *SSL) ToEntity() *entity.SSL {
	var ssl entity.SSL
	if err := json.Unmarshal(s.Config, &ssl); err != nil {
		return nil
	}
	return &ssl
}

// BeforeCreate 创建前钩子
func (s *SSL) BeforeCreate(tx *gorm.DB) (err error) {
	if err := s.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return s.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (s *SSL) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := s.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return s.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (s *SSL) BeforeDelete(tx *gorm.DB) (err error) {
	if s.ID == "" {
		return nil
	}
	if err := s.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return s.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (s *SSL) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	if s.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate {
		// 获取原始数据
		var origin SSL
		if err := tx.First(&origin, "id = ?", s.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		s.GatewayID, s.ID, s.Updater, s.Status, operation, constant.SSL, originConfig, s.Config)
}

// HandleConfig 处理配置
func (s *SSL) HandleConfig() (err error) {
	s.Config, err = sjson.SetBytes(s.Config, "id", s.ID)
	if err != nil {
		return err
	}
	if s.Name != "" {
		s.Config, err = sjson.SetBytes(s.Config, "name", s.Name)
		if err != nil {
			return err
		}
	}
	// Remove empty fields
	config, err := jsonx.RemoveEmptyObjectsAndArrays(string(s.Config))
	if err == nil {
		s.Config = []byte(config)
	}
	crt := gjson.GetBytes(s.Config, "cert").String()
	key := gjson.GetBytes(s.Config, "key").String()
	sins := gjson.GetBytes(s.Config, "snis").String()
	if sins == "" {
		snis, err := sslx.ParseCert(crt, key)
		if err != nil {
			return err
		}
		s.Config, _ = sjson.SetBytes(s.Config, "snis", snis)
	}
	return nil
}
