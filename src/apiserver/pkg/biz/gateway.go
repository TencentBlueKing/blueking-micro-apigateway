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

package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/tidwall/gjson"
	"gorm.io/datatypes"
	"gorm.io/gen"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

var gatewayResourceModelNameList = []string{
	model.Route{}.TableName(),
	model.Consumer{}.TableName(),
	model.Service{}.TableName(),
	model.Upstream{}.TableName(),
	model.SSL{}.TableName(),
	model.Proto{}.TableName(),
	model.ConsumerGroup{}.TableName(),
	model.PluginConfig{}.TableName(),
	model.GlobalRule{}.TableName(),
	model.PluginMetadata{}.TableName(),
	model.GatewayResourceSchemaAssociation{}.TableName(),
	model.GatewayCustomPluginSchema{}.TableName(),
	model.StreamRoute{}.TableName(),
	model.GatewaySyncData{}.TableName(),
	model.GatewayReleaseVersion{}.TableName(),
}

// ListGateways 查询网关列表
func ListGateways(ctx context.Context, mode uint8) ([]*model.Gateway, error) {
	u := repo.Gateway
	if mode == 0 {
		return repo.Gateway.WithContext(ctx).Order(u.CreatedAt.Desc()).Find()
	}
	return repo.Gateway.WithContext(ctx).Where(u.Mode.Eq(mode)).Order(u.CreatedAt.Desc()).Find()
}

// CreateGateway 创建网关
func CreateGateway(ctx context.Context, gateway *model.Gateway) error {
	return repo.Gateway.WithContext(ctx).Create(gateway)
}

// UpdateGateway 更新网关
func UpdateGateway(ctx context.Context, gateway model.Gateway) error {
	u := repo.Gateway
	_, err := u.WithContext(ctx).Where(u.ID.Eq(gateway.ID)).Select(
		u.Name, u.Mode, u.Maintainers, u.Desc,
		u.EtcdConfig, u.Token, u.Updater, u.ReadOnly,
	).Updates(&gateway)
	return err
}

// SaveGateway save网关
func SaveGateway(ctx context.Context, gateway *model.Gateway) error {
	u := repo.Gateway
	return u.WithContext(ctx).Save(gateway)
}

// GetGateway 查询网关详情
func GetGateway(ctx context.Context, id int) (*model.Gateway, error) {
	u := repo.Gateway
	return u.WithContext(ctx).Where(u.ID.Eq(id)).First()
}

// GetGatewayByName 根据name查询网关详情
func GetGatewayByName(ctx context.Context, name string) (*model.Gateway, error) {
	u := repo.Gateway
	return u.WithContext(ctx).Where(u.Name.Eq(name)).First()
}

// ExistsGatewayName 查询网关是否存在 (不包括自己)
func ExistsGatewayName(ctx context.Context, name string, id int) bool {
	u := repo.Gateway
	conditions := []gen.Condition{u.Name.Eq(name)}
	if id != 0 {
		conditions = append(conditions, u.ID.Neq(id))
	}
	gateways, err := u.WithContext(ctx).Where(conditions...).Find()
	if err != nil {
		return false
	}
	if len(gateways) == 0 {
		return true
	}
	return false
}

// GetGatewayEtcdConfigList 查询 etcd_config 比对结果的网关列表
func GetGatewayEtcdConfigList(ctx context.Context, key string, val string) ([]*model.Gateway, error) {
	return repo.Gateway.WithContext(ctx).Where(
		gen.Cond(datatypes.JSONQuery("etcd_config").Equals(val, key))...,
	).Find()
}

// ListGatewayResourceLabels 查询网关对应资源的标签列表
func ListGatewayResourceLabels(
	ctx context.Context,
	resourceType constant.APISIXResource,
) ([]map[string]string, error) {
	var resourceList []*model.ResourceCommonModel
	err := database.Client().WithContext(ctx).Table(
		resourceTableMap[resourceType]).Where(
		"gateway_id = ?", ginx.GetGatewayInfoFromContext(ctx).ID).Find(&resourceList).Error
	if err != nil {
		return nil, err
	}
	var keys []string
	labelMap := make(map[string][]string)
	for _, resource := range resourceList {
		var labels map[string]interface{}
		err := json.Unmarshal([]byte(gjson.ParseBytes(resource.Config).Get("labels").String()), &labels)
		if err != nil {
			continue
		}
		for key, val := range labels {
			if _, ok := labelMap[key]; !ok {
				labelMap[key] = []string{}
				keys = append(keys, key)
			}
			labelMap[key] = append(labelMap[key], fmt.Sprint(val))
		}
	}
	sort.Strings(keys)
	var labels []map[string]string
	for _, key := range keys {
		for _, val := range labelMap[key] {
			label := make(map[string]string)
			label[key] = val
			labels = append(labels, label)
		}
	}
	return labels, nil
}

// DeleteGateway 删除网关
func DeleteGateway(ctx context.Context, gateway *model.Gateway) error {
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		// 删除网关下所有资源
		for _, v := range gatewayResourceModelNameList {
			err := database.Client().WithContext(ctx).Table(v).Where("gateway_id = ?", gateway.ID).Delete(v).Error
			if err != nil {
				return err
			}
		}
		// 删除网关本身
		u := repo.Gateway
		_, err := u.WithContext(ctx).Delete(gateway)
		return err
	})
	return err
}
