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
	"github.com/tidwall/gjson"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ListConsumers 查询网关 Consumer 列表
func ListConsumers(ctx context.Context, gatewayID int) ([]*model.Consumer, error) {
	u := repo.Consumer
	return repo.Consumer.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Order(u.UpdatedAt.Desc()).Find()
}

// GetConsumerOrderExprList 获取 Consumer 排序字段列表
func GetConsumerOrderExprList(orderBy string) []field.Expr {
	u := repo.Consumer
	ascFieldMap := map[string]field.Expr{
		"username":   u.Username,
		"updated_at": u.UpdatedAt,
	}
	descFieldMap := map[string]field.Expr{
		"username":   u.Username.Desc(),
		"updated_at": u.UpdatedAt.Desc(),
	}
	orderByExprList := ParseOrderByExprList(ascFieldMap, descFieldMap, orderBy)
	if len(orderByExprList) == 0 {
		orderByExprList = append(orderByExprList, u.UpdatedAt.Desc())
	}
	return orderByExprList
}

// ListPagedConsumers 分页查询 Consumer 列表
func ListPagedConsumers(
	ctx context.Context,
	param map[string]interface{},
	label map[string][]string,
	status []string,
	name string,
	updater string,
	groupID string,
	orderBy string,
	page PageParam,
) ([]*model.Consumer, int64, error) {
	u := repo.Consumer
	query := u.WithContext(ctx)
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	if name != "" {
		query = query.Where(u.Username.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	associationIDCond := u.WithContext(ctx).Clauses()
	if groupID != "" {
		if groupID == constant.EmptyAssociationFilter {
			associationIDCond.Where(u.GroupID.Eq("")).Or(u.GroupID.IsNull())
		} else {
			query = query.Where(u.GroupID.Eq(groupID))
		}
	}
	orderByExprs := GetConsumerOrderExprList(orderBy)
	cond := u.WithContext(ctx).Clauses()
	conditions := LabelConditionList(label)
	if len(conditions) > 0 {
		for _, condition := range conditions {
			cond = cond.Or(condition)
		}
	}
	return query.Where(cond).
		Where(associationIDCond).
		Where(field.Attrs(param)).
		Order(orderByExprs...).
		FindByPage(page.Offset, page.Limit)
}

// CreateConsumer 创建 Consumer
func CreateConsumer(ctx context.Context, consumer model.Consumer) error {
	return repo.Consumer.WithContext(ctx).Create(&consumer)
}

// BatchCreateConsumers 批量创建 Consumer
func BatchCreateConsumers(ctx context.Context, consumers []*model.Consumer) error {
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).Consumer.WithContext(ctx).Create(consumers...)
	}
	return repo.Consumer.WithContext(ctx).Create(consumers...)
}

// UpdateConsumer 更新 Consumer
func UpdateConsumer(ctx context.Context, consumer model.Consumer) error {
	u := repo.Consumer
	_, err := u.WithContext(ctx).Where(u.ID.Eq(consumer.ID)).Select(
		u.Username,
		u.Updater,
		u.GroupID,
		u.Status,
		u.Config,
	).Updates(consumer)
	return err
}

// GetConsumer 查询 Consumer 详情
func GetConsumer(ctx context.Context, id string) (*model.Consumer, error) {
	u := repo.Consumer
	return u.WithContext(ctx).Where(u.ID.Eq(id)).First()
}

// QueryConsumers 搜索 consumer
func QueryConsumers(ctx context.Context, param map[string]interface{}) ([]*model.Consumer, error) {
	u := repo.Consumer
	return u.WithContext(ctx).Where(field.Attrs(param)).Find()
}

// BatchDeleteConsumers 批量删除 consumer 并添加审计日志
func BatchDeleteConsumers(ctx context.Context, ids []string) error {
	u := repo.Consumer
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.Consumer, ids)
		if err != nil {
			return err
		}
		// 批量删除 consumer 关联的自定义插件记录
		err = BatchDeleteResourceSchemaAssociation(ctx, ids, constant.Consumer)
		if err != nil {
			return err
		}
		_, err = tx.Consumer.WithContext(ctx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// BatchRevertConsumers 批量回滚 Consumer
func BatchRevertConsumers(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	consumers, err := QueryConsumers(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	for _, consumer := range consumers {
		if consumer.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			consumer.Status = constant.ResourceStatusSuccess
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[consumer.ID]; ok {
			consumer.Username = gjson.ParseBytes(syncData.Config).Get("username").String()
			consumer.Config = syncData.Config
			consumer.Status = constant.ResourceStatusSuccess
			// 更新关联关系数据
			consumer.GroupID = gjson.ParseBytes(syncData.Config).Get("group_id").String()
			continue
		} else {
			return errors.New("can not find sync data for consumer id:" + consumer.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		for _, consumer := range consumers {
			err := tx.Consumer.WithContext(ctx).Save(consumer)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
