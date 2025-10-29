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

// Package model 用于存放数据库模型
package model

import (
	"fmt"
	"time"

	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// BaseModel 基础模型
type BaseModel struct {
	Creator   string    `json:"creator" gorm:"type:varchar(32);null"`
	Updater   string    `json:"updater" gorm:"type:varchar(32);null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ResourceCommonModel  资源通用模型
type ResourceCommonModel struct {
	BaseModel
	AutoID    int            `gorm:"column:auto_id;type:int;primaryKey;autoIncrement"`                   // 自增ID
	ID        string         `gorm:"column:id;type:varchar(255);uniqueIndex:idx_id"`                     // apisix ID
	GatewayID int            `gorm:"column:gateway_id;type:int;uniqueIndex:idx_name;uniqueIndex:idx_id"` // 网关ID
	Config    datatypes.JSON `gorm:"column:config;type:json"`                                            // config
	// 发布状态: create-draft,update-draft,success,delete-draft
	Status constant.ResourceStatus `gorm:"column:status;type:varchar(32)"`
}

// GetResourceKey 获取资源key
func (r ResourceCommonModel) GetResourceKey(resourceType constant.APISIXResource) string {
	return fmt.Sprintf(constant.ResourceKeyFormat, resourceType, r.ID)
}

// GetResourceNameKey 获取资源名称key
func GetResourceNameKey(resourceType constant.APISIXResource) string {
	if resourceType == constant.Consumer {
		return "username"
	}
	return "name"
}

// GetServiceID 获取service id
func (r ResourceCommonModel) GetServiceID() string {
	return gjson.GetBytes(r.Config, "service_id").String()
}

// GetUpstreamID 获取upstream id
func (r ResourceCommonModel) GetUpstreamID() string {
	return gjson.GetBytes(r.Config, "upstream_id").String()
}

// GetPluginConfigID 获取plugin config id
func (r ResourceCommonModel) GetPluginConfigID() string {
	return gjson.GetBytes(r.Config, "plugin_config_id").String()
}

// GetGroupID 获取group id
func (r ResourceCommonModel) GetGroupID() string {
	return gjson.GetBytes(r.Config, "group_id").String()
}

// GetSSLID 获取ssl id
func (r ResourceCommonModel) GetSSLID() string {
	return gjson.GetBytes(r.Config, "tls.client_key").String()
}

// GetName 获取name
func (r ResourceCommonModel) GetName(resourceType constant.APISIXResource) string {
	return gjson.GetBytes(r.Config, GetResourceNameKey(resourceType)).String()
}

// ToResourceModel 转换为具体资源
func (r ResourceCommonModel) ToResourceModel(resourceType constant.APISIXResource) interface{} {
	switch resourceType {
	case constant.Route:
		return Route{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
			ServiceID:           r.GetServiceID(),
			PluginConfigID:      r.GetPluginConfigID(),
			UpstreamID:          r.GetUpstreamID(),
		}
	case constant.Service:
		return Service{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
			UpstreamID:          r.GetUpstreamID(),
		}
	case constant.Upstream:
		return Upstream{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.Consumer:
		return Consumer{
			ResourceCommonModel: r,
			Username:            r.GetName(resourceType),
			GroupID:             r.GetGroupID(),
		}
	case constant.ConsumerGroup:
		return ConsumerGroup{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.PluginConfig:
		return PluginConfig{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.GlobalRule:
		return GlobalRule{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.PluginMetadata:
		return PluginMetadata{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.Proto:
		return Proto{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.SSL:
		return SSL{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.StreamRoute:
		return StreamRoute{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
			ServiceID:           r.GetServiceID(),
			UpstreamID:          r.GetUpstreamID(),
		}
	}
	return nil
}
