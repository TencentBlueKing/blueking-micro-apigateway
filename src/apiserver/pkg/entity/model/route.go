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

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - APIGateway) available.
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

// Route 路由资源表
type Route struct {
	Name string `gorm:"column:name;type:varchar(64);uniqueIndex:idx_name"` // route 名称
	// 关联 service 唯一标识
	ServiceID string `gorm:"column:service_id;type:varchar(255)"`
	// 关联 upstream_id 唯一标识
	UpstreamID string `gorm:"column:upstream_id;type:varchar(255)"`
	// 关联 plugin_config_id 唯一标识
	PluginConfigID      string                 `gorm:"column:plugin_config_id;type:varchar(255)"`
	ResourceCommonModel                        // 资源通用 model: 创建时间、更新时间、创建人、更新人、config、status 等
	OperationType       constant.OperationType `gorm:"-"` // 用于标识操作类型，不持久化到数据库
}

// TableName 设置表名
func (Route) TableName() string {
	return "route"
}

// BeforeCreate 创建前钩子
func (r *Route) BeforeCreate(tx *gorm.DB) (err error) {
	if err := r.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, r.GatewayID, r.ID, constant.Route, r.Config)
	if err != nil {
		return err
	}
	// 添加审计
	return r.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (r *Route) BeforeUpdate(tx *gorm.DB) (err error) {
	// 处理特殊字段
	if err := r.HandleConfig(); err != nil {
		return err
	}
	// 关联自定义插件
	err = ResourceSchemaCallback(tx, r.GatewayID, r.ID, constant.Route, r.Config)
	if err != nil {
		return err
	}
	// 如果更新的操作类型为撤销，则不触发审计
	if r.OperationType == constant.OperationTypeRevert {
		return nil
	}
	// 添加审计
	return r.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (r *Route) BeforeDelete(tx *gorm.DB) (err error) {
	// 处理特殊字段
	if err := r.HandleConfig(); err != nil {
		return err
	}
	// 添加审计
	return r.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AddAuditLog 添加审计
func (r *Route) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	// 排除批量删除，更新的情况
	if r.ID == "" {
		return nil
	}
	originConfig := datatypes.JSON{}
	if operation != constant.OperationTypeCreate {
		// 获取原始数据
		var origin Route
		if err := tx.First(&origin, "id = ?", r.ID).Error; err != nil {
			return err
		}
		originConfig = origin.Config
	}
	return auditCallback(tx,
		r.GatewayID, r.ID, r.Updater, r.Status, operation, constant.Route, originConfig, r.Config)
}

// HandleConfig 处理特殊字段
func (r *Route) HandleConfig() (err error) {
	r.Config, err = sjson.SetBytes(r.Config, "id", r.ID)
	if err != nil {
		return err
	}

	if r.Name != "" {
		r.Config, err = sjson.SetBytes(r.Config, "name", r.Name)
		if err != nil {
			return err
		}
	}

	if r.ServiceID != "" {
		r.Config, err = sjson.SetBytes(r.Config, "service_id", r.ServiceID)
		if err != nil {
			return err
		}
	} else {
		r.Config, _ = sjson.DeleteBytes(r.Config, "service_id")
	}

	if r.PluginConfigID != "" {
		r.Config, err = sjson.SetBytes(r.Config, "plugin_config_id", r.PluginConfigID)
		if err != nil {
			return err
		}
	} else {
		r.Config, _ = sjson.DeleteBytes(r.Config, "plugin_config_id")
	}

	if r.UpstreamID != "" {
		r.Config, err = sjson.SetBytes(r.Config, "upstream_id", r.UpstreamID)
		if err != nil {
			return err
		}
	} else {
		r.Config, _ = sjson.DeleteBytes(r.Config, "upstream_id")
	}
	// 去除空字段
	config, err := jsonx.RemoveEmptyObjectsAndArrays(string(r.Config))
	if err == nil {
		r.Config = []byte(config)
	}
	return nil
}
