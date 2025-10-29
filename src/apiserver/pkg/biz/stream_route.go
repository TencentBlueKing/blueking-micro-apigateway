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

// buildStreamRouteQuery 获取 StreamRoute 查询对象
func buildStreamRouteQuery(ctx context.Context) repo.IStreamRouteDo {
	return repo.StreamRoute.WithContext(ctx).Where(field.Attrs(map[string]interface{}{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// GbuildStreamRouteQueryWithTx 获取 StreamRoute 查询对象
func buildStreamRouteQueryWithTx(ctx context.Context, tx *repo.Query) repo.IStreamRouteDo {
	return tx.WithContext(ctx).StreamRoute.Where(field.Attrs(map[string]interface{}{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// ListStreamRoutes 查询网关 StreamRoute 列表
func ListStreamRoutes(ctx context.Context) ([]*model.StreamRoute, error) {
	u := repo.StreamRoute
	return buildStreamRouteQuery(ctx).Order(u.UpdatedAt.Desc()).Find()
}

// GetStreamRouteOrderExprList 获取 StreamRoute 排序字段列表
func GetStreamRouteOrderExprList(orderBy string) []field.Expr {
	u := repo.StreamRoute
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

// ListPagedStreamRoutes 分页查询 StreamRoute
func ListPagedStreamRoutes(
	ctx context.Context,
	param map[string]interface{},
	label map[string][]string,
	status []string,
	name string,
	updater string,
	serviceID string,
	upstreamID string,
	orderBy string,
	page PageParam,
) ([]*model.StreamRoute, int64, error) {
	u := repo.StreamRoute
	query := buildStreamRouteQuery(ctx)
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	associationIDCond := u.WithContext(ctx).Clauses()
	if serviceID != "" {
		if serviceID == constant.EmptyAssociationFilter {
			associationIDCond.Where(u.ServiceID.Eq("")).Or(u.ServiceID.IsNull())
		} else {
			query = query.Where(u.ServiceID.Eq(serviceID))
		}
	}
	if upstreamID != "" {
		if upstreamID == constant.EmptyAssociationFilter {
			associationIDCond.Where(u.UpstreamID.Eq("")).Or(u.UpstreamID.IsNull())
		} else {
			query = query.Where(u.UpstreamID.Eq(upstreamID))
		}
	}
	orderByExprs := GetStreamRouteOrderExprList(orderBy)
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

// CreateStreamRoute 创建 StreamRoute
func CreateStreamRoute(ctx context.Context, streamRoute model.StreamRoute) error {
	if ginx.GetTx(ctx) != nil {
		return buildStreamRouteQueryWithTx(ctx, ginx.GetTx(ctx)).Create(&streamRoute)
	}
	return buildStreamRouteQuery(ctx).WithContext(ctx).Create(&streamRoute)
}

// BatchCreateStreamRoutes 批量创建 StreamRoute
func BatchCreateStreamRoutes(ctx context.Context, streamRoutes []*model.StreamRoute) error {
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).StreamRoute.WithContext(ctx).Create(streamRoutes...)
	}
	return repo.StreamRoute.WithContext(ctx).Create(streamRoutes...)
}

// UpdateStreamRoute 更新 StreamRoute
func UpdateStreamRoute(ctx context.Context, streamRoute model.StreamRoute) error {
	u := repo.StreamRoute
	_, err := buildStreamRouteQuery(ctx).Where(u.ID.Eq(streamRoute.ID)).Select(
		u.Name,
		u.ServiceID,
		u.UpstreamID,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(streamRoute)
	return err
}

// GetStreamRoute 查询 StreamRoute 详情
func GetStreamRoute(ctx context.Context, id string) (*model.StreamRoute, error) {
	u := repo.StreamRoute
	return buildStreamRouteQuery(ctx).Where(u.ID.Eq(id)).First()
}

// QueryStreamRoutes 搜索 StreamRoute
func QueryStreamRoutes(ctx context.Context, param map[string]interface{}) ([]*model.StreamRoute, error) {
	return buildStreamRouteQuery(ctx).Where(field.Attrs(param)).Find()
}

// BatchDeleteStreamRoutes 批量删除 StreamRoute 并添加审计日志
func BatchDeleteStreamRoutes(ctx context.Context, ids []string) error {
	u := repo.StreamRoute
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.StreamRoute, ids)
		if err != nil {
			return err
		}
		_, err = buildStreamRouteQueryWithTx(ctx, tx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// BatchRevertStreamRoutes 批量回滚 StreamRoute
func BatchRevertStreamRoutes(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	streamRoutes, err := QueryStreamRoutes(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(streamRoutes))
	for _, sr := range streamRoutes {
		// 标识此次更新的操作类型为撤销
		sr.OperationType = constant.OperationTypeRevert
		if sr.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			sr.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     sr.ID,
				Config: sr.Config,
				Status: sr.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[sr.ID]; ok {
			sr.Name = syncData.GetName()
			sr.Config = syncData.Config
			sr.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     sr.ID,
				Config: sr.Config,
				Status: sr.Status,
			})
			continue
		} else {
			return errors.New("can not find sync data for streamRoute id:" + sr.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.StreamRoute, ids, afterResources)
		if err != nil {
			return err
		}
		for _, sr := range streamRoutes {
			_, err := buildStreamRouteQueryWithTx(ctx, tx).Updates(sr)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
