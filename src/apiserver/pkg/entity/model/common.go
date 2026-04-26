/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
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
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
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
	AutoID    int            `gorm:"column:auto_id;type:int;primaryKey;autoIncrement"`                   // 自增 ID
	ID        string         `gorm:"column:id;type:varchar(255);uniqueIndex:idx_id"`                     // apisix ID
	GatewayID int            `gorm:"column:gateway_id;type:int;uniqueIndex:idx_name;uniqueIndex:idx_id"` // 网关 ID
	Config    datatypes.JSON `gorm:"column:config;type:json"`                                            // config
	// 发布状态：create-draft,update-draft,success,delete-draft
	Status              constant.ResourceStatus `gorm:"column:status;type:varchar(32)"`
	NameValue           string                  `gorm:"column:name_value;->" json:"-"`
	ServiceIDValue      string                  `gorm:"column:service_id_value;->" json:"-"`
	UpstreamIDValue     string                  `gorm:"column:upstream_id_value;->" json:"-"`
	PluginConfigIDValue string                  `gorm:"column:plugin_config_id_value;->" json:"-"`
	GroupIDValue        string                  `gorm:"column:group_id_value;->" json:"-"`
	SSLIDValue          string                  `gorm:"column:ssl_id_value;->" json:"-"`
}

var resourceStorageEchoFields = map[constant.APISIXResource][]string{
	constant.Route:          {"id", "name", "service_id", "upstream_id", "plugin_config_id"},
	constant.Service:        {"id", "name", "upstream_id"},
	constant.Upstream:       {"id", "name", "tls.client_cert_id"},
	constant.Consumer:       {"id", "username", "group_id"},
	constant.ConsumerGroup:  {"id", "name"},
	constant.PluginConfig:   {"id", "name"},
	constant.GlobalRule:     {"id", "name"},
	constant.PluginMetadata: {"id", "name"},
	constant.Proto:          {"id", "name"},
	constant.SSL:            {"id", "name"},
	constant.StreamRoute:    {"id", "name", "service_id", "upstream_id"},
}

func stripResourceConfigForStorage(
	resourceType constant.APISIXResource,
	config datatypes.JSON,
) (datatypes.JSON, error) {
	updated := datatypes.JSON(append([]byte(nil), config...))
	for _, fieldName := range resourceStorageEchoFields[resourceType] {
		if !gjson.GetBytes(updated, fieldName).Exists() {
			continue
		}
		next, err := sjson.DeleteBytes(updated, fieldName)
		if err != nil {
			return nil, err
		}
		updated = datatypes.JSON(next)
	}
	cleaned, err := jsonx.RemoveEmptyObjectsAndArrays(string(updated))
	if err == nil {
		updated = datatypes.JSON(cleaned)
	}
	return updated, nil
}

func restoreResourceConfigForRead(
	resourceType constant.APISIXResource,
	config datatypes.JSON,
	resourceID string,
	nameValue string,
	associations map[string]string,
) (datatypes.JSON, error) {
	updated := datatypes.JSON(append([]byte(nil), config...))
	var err error

	idValue := resourceID
	if resourceType == constant.PluginMetadata {
		idValue = nameValue
	}
	if idValue != "" {
		updated, err = sjson.SetBytes(updated, "id", idValue)
		if err != nil {
			return nil, err
		}
	}

	if nameValue != "" {
		updated, err = sjson.SetBytes(updated, GetResourceNameKey(resourceType), nameValue)
		if err != nil {
			return nil, err
		}
	}

	for fieldName, value := range associations {
		if value == "" {
			continue
		}
		updated, err = sjson.SetBytes(updated, fieldName, value)
		if err != nil {
			return nil, err
		}
	}

	return updated, nil
}

// RestoreConfigForRead reconstitutes the historical read-time config shape from authoritative columns.
func (r *ResourceCommonModel) RestoreConfigForRead(resourceType constant.APISIXResource) error {
	config, err := restoreResourceConfigForRead(
		resourceType,
		r.Config,
		r.ID,
		r.GetName(resourceType),
		map[string]string{
			"service_id":         r.GetServiceID(),
			"upstream_id":        r.GetUpstreamID(),
			"plugin_config_id":   r.GetPluginConfigID(),
			"group_id":           r.GetGroupID(),
			"tls.client_cert_id": r.GetSSLID(),
		},
	)
	if err != nil {
		return err
	}
	r.Config = config
	return nil
}

// GetResourceKey 获取资源 key
func (r ResourceCommonModel) GetResourceKey(resourceType constant.APISIXResource) string {
	// 插件元素数需要特殊处理，因为插件元素数没有真正 id
	if resourceType == constant.PluginMetadata {
		return fmt.Sprintf(constant.ResourceKeyFormat, resourceType, r.GetName(resourceType))
	}
	return fmt.Sprintf(constant.ResourceKeyFormat, resourceType, r.ID)
}

// GetResourceNameKey 获取资源名称 key
func GetResourceNameKey(resourceType constant.APISIXResource) string {
	if resourceType == constant.Consumer {
		return "username"
	}
	return "name"
}

// GetServiceID 获取 service id
func (r ResourceCommonModel) GetServiceID() string {
	if r.ServiceIDValue != "" {
		return r.ServiceIDValue
	}
	return gjson.GetBytes(r.Config, "service_id").String()
}

// GetUpstreamID 获取 upstream id
func (r ResourceCommonModel) GetUpstreamID() string {
	if r.UpstreamIDValue != "" {
		return r.UpstreamIDValue
	}
	return gjson.GetBytes(r.Config, "upstream_id").String()
}

// GetPluginConfigID 获取 plugin config id
func (r ResourceCommonModel) GetPluginConfigID() string {
	if r.PluginConfigIDValue != "" {
		return r.PluginConfigIDValue
	}
	return gjson.GetBytes(r.Config, "plugin_config_id").String()
}

// GetGroupID 获取 group id
func (r ResourceCommonModel) GetGroupID() string {
	if r.GroupIDValue != "" {
		return r.GroupIDValue
	}
	return gjson.GetBytes(r.Config, "group_id").String()
}

// GetSSLID 获取 ssl id
func (r ResourceCommonModel) GetSSLID() string {
	if r.SSLIDValue != "" {
		return r.SSLIDValue
	}
	return gjson.GetBytes(r.Config, "tls.client_cert_id").String()
}

// GetName 获取 name
func (r ResourceCommonModel) GetName(resourceType constant.APISIXResource) string {
	if r.NameValue != "" {
		return r.NameValue
	}
	return gjson.GetBytes(r.Config, GetResourceNameKey(resourceType)).String()
}

// ToResourceModel 转换为具体资源
func (r ResourceCommonModel) ToResourceModel(resourceType constant.APISIXResource) any {
	switch resourceType {
	case constant.Route:
		return &Route{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
			ServiceID:           r.GetServiceID(),
			PluginConfigID:      r.GetPluginConfigID(),
			UpstreamID:          r.GetUpstreamID(),
		}
	case constant.Service:
		return &Service{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
			UpstreamID:          r.GetUpstreamID(),
		}
	case constant.Upstream:
		return &Upstream{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
			SSLID:               r.GetSSLID(),
		}
	case constant.Consumer:
		return &Consumer{
			ResourceCommonModel: r,
			Username:            r.GetName(resourceType),
			GroupID:             r.GetGroupID(),
		}
	case constant.ConsumerGroup:
		return &ConsumerGroup{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.PluginConfig:
		return &PluginConfig{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.GlobalRule:
		return &GlobalRule{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.PluginMetadata:
		return &PluginMetadata{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.Proto:
		return &Proto{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.SSL:
		return &SSL{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
		}
	case constant.StreamRoute:
		return &StreamRoute{
			ResourceCommonModel: r,
			Name:                r.GetName(resourceType),
			ServiceID:           r.GetServiceID(),
			UpstreamID:          r.GetUpstreamID(),
		}
	}
	return nil
}
