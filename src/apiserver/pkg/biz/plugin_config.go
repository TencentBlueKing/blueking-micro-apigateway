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

	"github.com/pkg/errors"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ListPluginConfigs 查询网关 PluginConfig 列表
func ListPluginConfigs(ctx context.Context, gatewayID int) ([]*model.PluginConfig, error) {
	u := repo.PluginConfig
	return repo.PluginConfig.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Order(u.UpdatedAt.Desc()).Find()
}

// GetPluginConfigOrderExprList 获取 PluginConfig 排序字段列表
func GetPluginConfigOrderExprList(orderBy string) []field.Expr {
	u := repo.PluginConfig
	ascFieldMap := map[string]field.Expr{
		"name":       u.Name,
		"updated_at": u.UpdatedAt,
	}
	descFieldMap := map[string]field.Expr{
		"name":       u.Name.Desc(),
		"updated_at": u.UpdatedAt.Desc(),
	}
	orderByExprList := ParseOrderByExprList(ascFieldMap, descFieldMap, orderBy)
	if len(orderByExprList) == 0 {
		orderByExprList = append(orderByExprList, u.UpdatedAt.Desc())
	}
	return orderByExprList
}

// ListPagedPluginConfigs 分页查询 pluginConfig 表
func ListPagedPluginConfigs(
	ctx context.Context,
	param map[string]interface{},
	label map[string][]string,
	status []string,
	name string,
	updater string,
	orderBy string,
	page PageParam,
) ([]*model.PluginConfig, int64, error) {
	u := repo.PluginConfig
	query := u.WithContext(ctx)
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	orderByExprs := GetPluginConfigOrderExprList(orderBy)
	cond := u.WithContext(ctx).Clauses()
	conditions := LabelConditionList(label)
	if len(conditions) > 0 {
		for _, condition := range conditions {
			cond = cond.Or(condition)
		}
	}
	return query.Where(cond).
		Where(field.Attrs(param)).
		Order(orderByExprs...).
		FindByPage(page.Offset, page.Limit)
}

// CreatePluginConfig 创建 PluginConfig
func CreatePluginConfig(ctx context.Context, pluginConfig model.PluginConfig) error {
	return repo.PluginConfig.WithContext(ctx).Create(&pluginConfig)
}

// BatchCreatePluginConfigs 批量创建 PluginConfig
func BatchCreatePluginConfigs(ctx context.Context, pluginConfigs []*model.PluginConfig) error {
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).PluginConfig.WithContext(ctx).Create(pluginConfigs...)
	}
	return repo.PluginConfig.WithContext(ctx).Create(pluginConfigs...)
}

// UpdatePluginConfig 更新 PluginConfig
func UpdatePluginConfig(ctx context.Context, pluginConfig model.PluginConfig) error {
	u := repo.PluginConfig
	_, err := u.WithContext(ctx).Where(u.ID.Eq(pluginConfig.ID)).Select(
		u.Name,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(pluginConfig)
	return err
}

// GetPluginConfig 查询 PluginConfig 详情
func GetPluginConfig(ctx context.Context, id string) (*model.PluginConfig, error) {
	u := repo.PluginConfig
	return u.WithContext(ctx).Where(u.ID.Eq(id)).First()
}

// QueryPluginConfigs  搜索插件配置
func QueryPluginConfigs(ctx context.Context, param map[string]interface{}) ([]*model.PluginConfig, error) {
	u := repo.PluginConfig
	return u.WithContext(ctx).Where(field.Attrs(param)).Find()
}

// ExistsPluginConfig 查询 PluginConfig 是否存在
func ExistsPluginConfig(ctx context.Context, id string) bool {
	u := repo.PluginConfig
	pluginConfigs, err := u.WithContext(ctx).Where(
		u.ID.Eq(id),
		u.GatewayID.Eq(ginx.GetGatewayInfoFromContext(ctx).ID),
	).Find()
	if err != nil {
		return false
	}
	if len(pluginConfigs) == 0 {
		return false
	}
	return true
}

// BatchDeletePluginConfigs 批量删除 PluginConfig 并添加审计日志
func BatchDeletePluginConfigs(ctx context.Context, ids []string) error {
	u := repo.PluginConfig
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.PluginConfig, ids)
		if err != nil {
			return err
		}
		// 批量删除 PluginConfig 关联的自定义插件记录
		err = BatchDeleteResourceSchemaAssociation(ctx, ids, constant.PluginConfig)
		if err != nil {
			return err
		}
		_, err = tx.PluginConfig.WithContext(ctx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// BatchRevertPluginConfigs 批量回滚 PluginConfig
func BatchRevertPluginConfigs(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	pluginConfigs, err := QueryPluginConfigs(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(pluginConfigs))
	for _, pluginConfig := range pluginConfigs {
		// 标识此次更新的类型为撤销
		pluginConfig.OperationType = constant.OperationTypeRevert
		if pluginConfig.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			pluginConfig.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     pluginConfig.ID,
				Config: pluginConfig.Config,
				Status: pluginConfig.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[pluginConfig.ID]; ok {
			pluginConfig.Name = syncData.GetName()
			pluginConfig.Config = syncData.Config
			pluginConfig.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     pluginConfig.ID,
				Config: pluginConfig.Config,
				Status: pluginConfig.Status,
			})
			continue
		} else {
			return errors.New("can not find sync data for pluginConfig id:" + pluginConfig.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.PluginConfig, ids, afterResources)
		if err != nil {
			return err
		}
		for _, pluginConfig := range pluginConfigs {
			_, err := tx.PluginConfig.WithContext(ctx).Updates(pluginConfig)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
