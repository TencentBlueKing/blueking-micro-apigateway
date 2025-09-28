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
)

// StreamRoute 表示数据库中的 stream_route 表
type StreamRoute struct {
	// stream_route 名称
	Name string `gorm:"column:name;type:varchar(255);uniqueIndex:idx_name"`
	// 关联 service_id 唯一标识
	ServiceID string `gorm:"column:service_id;type:varchar(255)"`
	// 关联 upstream_id 唯一标识
	UpstreamID          string                 `gorm:"column:upstream_id;type:varchar(255)"`
	ResourceCommonModel                        // 资源通用model: 创建时间、更新时间、创建人、更新人、config、status等
	OperationType       constant.OperationType `gorm:"-"` // 用于标识操作类型，不持久化到数据库
}

// GetLabels 获取 labels
func (s StreamRoute) GetLabels() string {
	return gjson.GetBytes(s.Config, "labels").String()
}

// TableName 设置表名
func (StreamRoute) TableName() string {
	return "stream_route"
}

// BeforeCreate 创建前钩子
func (s *StreamRoute) BeforeCreate(tx *gorm.DB) (err error) {
	if err := s.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return s.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (s *StreamRoute) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := s.HandleConfig(); err != nil {
		return err
	}
	// 如果更新的操作类型为撤销，则不触发审计
	if s.OperationType == constant.OperationTypeRevert {
		return nil
	}
	// 添加审计
	return s.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (s *StreamRoute) BeforeDelete(tx *gorm.DB) (err error) {
	if err := s.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return s.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (s *StreamRoute) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	// 排除批量删除，更新的情况
	if s.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate {
		// 获取原始数据
		var origin StreamRoute
		if err := tx.First(&origin, "id = ?", s.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		s.GatewayID, s.ID, s.Updater, s.Status, operation, constant.StreamRoute, originConfig, s.Config)
}

// HandleConfig 处理配置
func (s *StreamRoute) HandleConfig() (err error) {
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
	if s.ServiceID != "" {
		s.Config, err = sjson.SetBytes(s.Config, "service_id", s.ServiceID)
		if err != nil {
			return err
		}
	} else {
		s.Config, _ = sjson.DeleteBytes(s.Config, "service_id")
	}
	if s.UpstreamID != "" {
		s.Config, err = sjson.SetBytes(s.Config, "upstream_id", s.UpstreamID)
		if err != nil {
			return err
		}
	} else {
		s.Config, _ = sjson.DeleteBytes(s.Config, "upstream_id")
	}
	// Remove empty fields
	config, err := jsonx.RemoveEmptyObjectsAndArrays(string(s.Config))
	if err == nil {
		s.Config = []byte(config)
	}
	return nil
}
