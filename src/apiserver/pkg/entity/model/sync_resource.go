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
	"fmt"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// GatewaySyncData  gateway_sync_data 表
type GatewaySyncData struct {
	AutoID    int    `gorm:"column:auto_id;primaryKey;autoIncrement"`                     // 自增ID
	ID        string `gorm:"column:id;type:varchar(255);uniqueIndex:idx_resource_unique"` // apisix资源ID
	GatewayID int    `gorm:"column:gateway_id;uniqueIndex:idx_resource_unique"`           // 对应网关ID
	// apisix资源类型: route/service/upstream
	Type        constant.APISIXResource `gorm:"column:type;type:varchar(32);uniqueIndex:idx_resource_unique"`
	Config      datatypes.JSON          `gorm:"column:config;type:json"` // etcd raw config
	ModRevision int                     `gorm:"column:mod_revision"`     // 更新版本
	CreatedAt   time.Time               `json:"createdAt"`               // 创建时间
	UpdatedAt   time.Time               `json:"updatedAt"`               // 更新时间
}

func (g GatewaySyncData) GetResourceKey() string {
	// 插件元素数需要特殊处理,因为插件元素数没有真正id
	if g.Type == constant.PluginMetadata {
		return fmt.Sprintf(constant.ResourceKeyFormat, g.Type, g.GetName())
	}
	return fmt.Sprintf(constant.ResourceKeyFormat, g.Type, g.ID)
}

// GetServiceID 获取service id
func (g GatewaySyncData) GetServiceID() string {
	return gjson.GetBytes(g.Config, "service_id").String()
}

// GetUpstreamID 获取upstream id
func (g GatewaySyncData) GetUpstreamID() string {
	return gjson.GetBytes(g.Config, "upstream_id").String()
}

// GetPluginConfigID 获取plugin config id
func (g GatewaySyncData) GetPluginConfigID() string {
	return gjson.GetBytes(g.Config, "plugin_config_id").String()
}

// GetGroupID 获取group id
func (g GatewaySyncData) GetGroupID() string {
	return gjson.GetBytes(g.Config, "group_id").String()
}

// GetName 获取name
func (g GatewaySyncData) GetName() string {
	return gjson.GetBytes(g.Config, GetResourceNameKey(g.Type)).String()
}

// SetName 设置name
func (g *GatewaySyncData) SetName(name string) {
	g.Config, _ = sjson.SetBytes(g.Config, GetResourceNameKey(g.Type), name)
}

// GetConfigID 获取config id
func (g GatewaySyncData) GetConfigID() string {
	return gjson.GetBytes(g.Config, "id").String()
}

// GetSSLID 获取ssl id
func (g GatewaySyncData) GetSSLID() string {
	return gjson.GetBytes(g.Config, "tls.client_cert_id").String()
}

// TableName 设置表名
func (GatewaySyncData) TableName() string {
	return "gateway_sync_data"
}

// GetConfigCreatedAt 获取更新时间
func (g GatewaySyncData) GetConfigCreatedAt() int64 {
	return gjson.ParseBytes(g.Config).Get("update_time").Int()
}
