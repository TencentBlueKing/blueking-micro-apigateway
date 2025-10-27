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

// ListUpstreams 查询网关 upstream 列表
func ListUpstreams(ctx context.Context, gatewayID int) ([]*model.Upstream, error) {
	u := repo.Upstream
	return repo.Upstream.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Order(u.UpdatedAt.Desc()).Find()
}

// GetUpstreamOrderExprList 获取 upstream 排序字段列表
func GetUpstreamOrderExprList(orderBy string) []field.Expr {
	u := repo.Upstream
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

// ListPagedUpstreams 分页查询upstream列表
func ListPagedUpstreams(
	ctx context.Context,
	param map[string]interface{},
	label map[string][]string,
	status []string,
	name string,
	updater string,
	orderBy string,
	page PageParam,
) ([]*model.Upstream, int64, error) {
	u := repo.Upstream
	query := u.WithContext(ctx)
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	orderByExprs := GetUpstreamOrderExprList(orderBy)
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

// CreateUpstream 创建 upstream
func CreateUpstream(ctx context.Context, upstream model.Upstream) error {
	return repo.Upstream.WithContext(ctx).Create(&upstream)
}

// BatchCreateUpstreams 批量创建 upstream
func BatchCreateUpstreams(ctx context.Context, upstreams []*model.Upstream) error {
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).Upstream.WithContext(ctx).Create(upstreams...)
	}
	return repo.Upstream.WithContext(ctx).Create(upstreams...)
}

// UpdateUpstream 更新 upstream
func UpdateUpstream(ctx context.Context, upstream model.Upstream) error {
	u := repo.Upstream
	gatewayID := ginx.GetGatewayInfoFromContext(ctx).ID
	_, err := u.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID), u.ID.Eq(upstream.ID)).Select(
		u.Name,
		u.Config,
		u.Status,
		u.SSLID,
		u.Updater,
	).Updates(upstream)
	return err
}

// GetUpstream 查询 upstream 详情
func GetUpstream(ctx context.Context, id string) (*model.Upstream, error) {
	u := repo.Upstream
	gatewayID := ginx.GetGatewayInfoFromContext(ctx).ID
	return u.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID), u.ID.Eq(id)).First()
}

// QueryUpstreams 搜索 upstream
func QueryUpstreams(ctx context.Context, param map[string]interface{}) ([]*model.Upstream, error) {
	u := repo.Upstream
	return u.WithContext(ctx).Where(field.Attrs(param)).Find()
}

// ExistsUpstream 查询 upstream 是否存在
func ExistsUpstream(ctx context.Context, id string) bool {
	u := repo.Upstream
	upstreams, err := u.WithContext(ctx).Where(
		u.ID.Eq(id),
		u.GatewayID.Eq(ginx.GetGatewayInfoFromContext(ctx).ID),
	).Find()
	if err != nil {
		return false
	}
	if len(upstreams) == 0 {
		return false
	}
	return true
}

// BatchDeleteUpstreams 批量删除 upstream 并添加审计日志
func BatchDeleteUpstreams(ctx context.Context, ids []string) error {
	u := repo.Upstream
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.Upstream, ids)
		if err != nil {
			return err
		}
		gatewayID := ginx.GetGatewayInfoFromContext(ctx).ID
		_, err = tx.Upstream.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID), u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// GetUpstreamCount 查询网关 upstream 数量
func GetUpstreamCount(ctx context.Context, gatewayID int) (int64, error) {
	u := repo.Upstream
	return u.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Count()
}

// BatchRevertUpstreams 批量回滚 upstream
func BatchRevertUpstreams(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	upstreams, err := QueryUpstreams(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(upstreams))
	for _, upstream := range upstreams {
		// 标识此次更新的操作类型为撤销
		upstream.OperationType = constant.OperationTypeRevert
		if upstream.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			upstream.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     upstream.ID,
				Config: upstream.Config,
				Status: upstream.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[upstream.ID]; ok {
			upstream.Name = syncData.GetName()
			upstream.Config = syncData.Config
			upstream.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     upstream.ID,
				Config: upstream.Config,
				Status: upstream.Status,
			})
			continue
		} else {
			return errors.New("can not find sync data for upstream id:" + upstream.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.Upstream, ids, afterResources)
		if err != nil {
			return err
		}
		for _, upstream := range upstreams {
			_, err := tx.Upstream.WithContext(ctx).Updates(upstream)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
