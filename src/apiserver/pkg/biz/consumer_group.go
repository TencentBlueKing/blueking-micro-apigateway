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

// buildConsumerGroupQuery 获取 ConsumerGroup 查询对象
func buildConsumerGroupQuery(ctx context.Context) repo.IConsumerGroupDo {
	return repo.ConsumerGroup.WithContext(ctx).Where(field.Attrs(map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// buildConsumerGroupQueryWithTx 获取 ConsumerGroup 查询对象
func buildConsumerGroupQueryWithTx(ctx context.Context, tx *repo.Query) repo.IConsumerGroupDo {
	return tx.ConsumerGroup.WithContext(ctx).Where(field.Attrs(map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// ListConsumerGroups 查询网关 ConsumerGroup 列表
func ListConsumerGroups(ctx context.Context) ([]*model.ConsumerGroup, error) {
	u := repo.ConsumerGroup
	return buildConsumerGroupQuery(ctx).Order(u.UpdatedAt.Desc()).Find()
}

// GetConsumerGroupOrderExprList 获取 ConsumerGroup 排序字段列表
func GetConsumerGroupOrderExprList(orderBy string) []field.Expr {
	u := repo.ConsumerGroup
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

// ListPagedConsumerGroups 分页查询 ConsumerGroup 列表
func ListPagedConsumerGroups(
	ctx context.Context,
	param map[string]any,
	label map[string][]string,
	status []string,
	name string,
	updater string,
	orderBy string,
	page PageParam,
) ([]*model.ConsumerGroup, int64, error) {
	u := repo.ConsumerGroup
	query := buildConsumerGroupQuery(ctx)
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	orderByExprs := GetConsumerGroupOrderExprList(orderBy)
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

// CreateConsumerGroup 创建 ConsumerGroup
func CreateConsumerGroup(ctx context.Context, consumerGroup model.ConsumerGroup) error {
	return repo.ConsumerGroup.WithContext(ctx).Create(&consumerGroup)
}

// BatchCreateConsumerGroups 批量创建 ConsumerGroup
func BatchCreateConsumerGroups(ctx context.Context, consumerGroups []*model.ConsumerGroup) error {
	if ginx.GetTx(ctx) != nil {
		return buildConsumerGroupQueryWithTx(ctx, ginx.GetTx(ctx)).Create(consumerGroups...)
	}
	return repo.ConsumerGroup.WithContext(ctx).Create(consumerGroups...)
}

// UpdateConsumerGroup 更新 ConsumerGroup
func UpdateConsumerGroup(ctx context.Context, consumerGroup model.ConsumerGroup) error {
	u := repo.ConsumerGroup
	_, err := buildConsumerGroupQuery(ctx).Where(u.ID.Eq(consumerGroup.ID)).Select(
		u.Name,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(consumerGroup)
	return err
}

// GetConsumerGroup 查询 ConsumerGroup 详情
func GetConsumerGroup(ctx context.Context, id string) (*model.ConsumerGroup, error) {
	u := repo.ConsumerGroup
	return buildConsumerGroupQuery(ctx).Where(u.ID.Eq(id)).First()
}

// QueryConsumerGroups 搜索 ConsumerGroup
func QueryConsumerGroups(ctx context.Context, param map[string]any) ([]*model.ConsumerGroup, error) {
	return buildConsumerGroupQuery(ctx).Where(field.Attrs(param)).Find()
}

// ExistsConsumerGroup 查询 ConsumerGroup 是否存在
func ExistsConsumerGroup(ctx context.Context, id string) bool {
	u := repo.ConsumerGroup
	groups, err := buildConsumerGroupQuery(ctx).Where(u.ID.Eq(id)).Find()
	if err != nil {
		return false
	}
	if len(groups) == 0 {
		return false
	}
	return true
}

// BatchDeleteConsumerGroups 批量删除 ConsumerGroup 并添加审计日志
func BatchDeleteConsumerGroups(ctx context.Context, ids []string) error {
	u := repo.ConsumerGroup
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		err := AddDeleteResourceByIDAuditLog(ctx, constant.ConsumerGroup, ids)
		if err != nil {
			return err
		}
		// 批量删除 ConsumerGroup 关联的自定义插件记录
		err = BatchDeleteResourceSchemaAssociation(ctx, ids, constant.ConsumerGroup)
		if err != nil {
			return err
		}
		_, err = buildConsumerGroupQueryWithTx(ctx, tx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// BatchRevertConsumerGroups 批量回滚 ConsumerGroup
func BatchRevertConsumerGroups(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	consumerGroups, err := QueryConsumerGroups(ctx, map[string]any{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(consumerGroups))
	for _, consumerGroup := range consumerGroups {
		// 标识此次更新的操作类型为撤销
		consumerGroup.OperationType = constant.OperationTypeRevert
		if consumerGroup.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			consumerGroup.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     consumerGroup.ID,
				Config: consumerGroup.Config,
				Status: consumerGroup.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[consumerGroup.ID]; ok {
			consumerGroup.Name = syncData.GetName()
			consumerGroup.Config = syncData.Config
			consumerGroup.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     consumerGroup.ID,
				Config: consumerGroup.Config,
				Status: consumerGroup.Status,
			})
			continue
		} else {
			return errors.New("can not find sync data for consumerGroup id:" + consumerGroup.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.ConsumerGroup, ids, afterResources)
		if err != nil {
			return err
		}
		for _, consumerGroup := range consumerGroups {
			_, err := buildConsumerGroupQueryWithTx(ctx, tx).Updates(consumerGroup)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
