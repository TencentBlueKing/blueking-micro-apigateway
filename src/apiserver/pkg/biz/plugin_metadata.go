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

// ListPluginMetadatas 查询网关 PluginMetadata 列表
func ListPluginMetadatas(ctx context.Context, gatewayID int) ([]*model.PluginMetadata, error) {
	u := repo.PluginMetadata
	return repo.PluginMetadata.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Order(u.UpdatedAt.Desc()).Find()
}

// GetPluginMetadataOrderExprList 获取 PluginMetadata 排序字段列表
func GetPluginMetadataOrderExprList(orderBy string) []field.Expr {
	u := repo.PluginMetadata
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

// ListPagedPluginMetadatas 分页查询 PluginMetadata 列表
func ListPagedPluginMetadatas(
	ctx context.Context,
	param map[string]interface{},
	status []string,
	name string,
	updater string,
	orderBy string,
	page PageParam,
) ([]*model.PluginMetadata, int64, error) {
	u := repo.PluginMetadata
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
	orderByExprs := GetPluginMetadataOrderExprList(orderBy)
	return query.Where(field.Attrs(param)).Order(orderByExprs...).FindByPage(page.Offset, page.Limit)
}

// CreatePluginMetadata 创建 PluginMetadata
func CreatePluginMetadata(ctx context.Context, pluginMetadata model.PluginMetadata) error {
	return repo.PluginMetadata.WithContext(ctx).Create(&pluginMetadata)
}

// batchCreatePluginMetadatas 批量创建 PluginMetadata
func batchCreatePluginMetadatas(ctx context.Context, pluginMetadataList []*model.PluginMetadata) error {
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).PluginMetadata.WithContext(ctx).Create(pluginMetadataList...)
	}
	return repo.PluginMetadata.WithContext(ctx).Create(pluginMetadataList...)
}

// UpdatePluginMetadata 更新 PluginMetadata
func UpdatePluginMetadata(ctx context.Context, pluginMetadata model.PluginMetadata) error {
	u := repo.PluginMetadata
	_, err := u.WithContext(ctx).Where(u.ID.Eq(pluginMetadata.ID)).Select(
		u.Name,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(pluginMetadata)
	return err
}

// GetPluginMetadata 查询 PluginMetadata 详情
func GetPluginMetadata(ctx context.Context, id string) (*model.PluginMetadata, error) {
	u := repo.PluginMetadata
	return u.WithContext(ctx).Where(u.ID.Eq(id)).First()
}

// QueryPluginMetadatas 搜索 PluginMetadata
func QueryPluginMetadatas(ctx context.Context, param map[string]interface{}) ([]*model.PluginMetadata, error) {
	u := repo.PluginMetadata
	return u.WithContext(ctx).Where(field.Attrs(param)).Find()
}

// BatchDeletePluginMetadatas 批量删除 PluginMetadata 并记录审计日志
func BatchDeletePluginMetadatas(ctx context.Context, ids []string) error {
	u := repo.PluginMetadata
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.PluginMetadata, ids)
		if err != nil {
			return err
		}
		// 批量删除 PluginMetadata 关联的自定义插件记录
		err = BatchDeleteResourceSchemaAssociation(ctx, ids, constant.PluginMetadata)
		if err != nil {
			return err
		}
		_, err = tx.PluginMetadata.WithContext(ctx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// BatchRevertPluginMetadatas 批量回滚 PluginMetadata
func BatchRevertPluginMetadatas(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	pluginMetadatas, err := QueryPluginMetadatas(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(pluginMetadatas))
	for _, pluginMetadata := range pluginMetadatas {
		// 标识此次更新的操作类型为撤销
		pluginMetadata.OperationType = constant.OperationTypeRevert
		if pluginMetadata.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			pluginMetadata.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     pluginMetadata.ID,
				Config: pluginMetadata.Config,
				Status: pluginMetadata.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[pluginMetadata.ID]; ok {
			pluginMetadata.Name = syncData.GetName()
			pluginMetadata.Config = syncData.Config
			pluginMetadata.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     pluginMetadata.ID,
				Config: pluginMetadata.Config,
				Status: pluginMetadata.Status,
			})
			continue
		} else {
			return errors.New("未找到插件元数据 id 的同步数据:" + pluginMetadata.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.PluginMetadata, ids, afterResources)
		if err != nil {
			return err
		}
		for _, pluginMetadata := range pluginMetadatas {
			_, err := tx.PluginMetadata.WithContext(ctx).Updates(pluginMetadata)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
