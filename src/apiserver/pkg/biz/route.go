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
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ListRoutes 查询网关路由列表
func ListRoutes(ctx context.Context, gatewayID int) ([]*model.Route, error) {
	u := repo.Route
	return repo.Route.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Order(u.UpdatedAt.Desc()).Find()
}

// GetRouteOrderExprList 获取路由排序字段列表
func GetRouteOrderExprList(orderBy string) []field.Expr {
	u := repo.Route
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

// ListPagedRoutes 分页查询网关路由列表
func ListPagedRoutes(
	ctx context.Context,
	param map[string]interface{},
	label map[string][]string,
	status []string,
	name string,
	updater string,
	path string,
	method string,
	serviceID string,
	upstreamID string,
	orderBy string,
	page PageParam,
) ([]*model.Route, int64, error) {
	u := repo.Route
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
	if path != "" {
		query = query.Where(gen.Cond(datatypes.JSONQuery("config").Likes("%"+path+"%", "uris"))...)
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
	methodCond := u.WithContext(ctx).Clauses()
	if method != "" {
		method = strings.ToUpper(method)

		// 过滤没有 methods 字段的（只要 method 有值，默认就查询 ANY）
		methodCond = methodCond.Not(gen.Cond(datatypes.JSONQuery("config").HasKey("methods"))...)

		// 查询非 ANY 时，过滤 methods 中包含查询条件的
		if method != constant.ANYMethodFilter {
			// 通过 methodCond 将两个查询条件合并，放在一个 Where 条件中
			methodCond = methodCond.Or(
				gen.Cond(datatypes.JSONQuery("config").Likes("%"+method+"%", "methods"))...,
			)
		}
	}
	orderByExprs := GetRouteOrderExprList(orderBy)
	cond := u.WithContext(ctx).Clauses()
	conditions := LabelConditionList(label)
	if len(conditions) > 0 {
		for _, condition := range conditions {
			cond = cond.Or(condition)
		}
	}
	return query.Where(cond).
		Where(methodCond).
		Where(associationIDCond).
		Where(field.Attrs(param)).
		Order(orderByExprs...).
		FindByPage(page.Offset, page.Limit)
}

// CreateRoute 创建路由
func CreateRoute(ctx context.Context, route model.Route) error {
	return repo.Route.WithContext(ctx).Create(&route)
}

// BatchCreateRoutes 批量创建路由
func BatchCreateRoutes(ctx context.Context, routes []*model.Route) error {
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).Route.WithContext(ctx).Create(routes...)
	}
	return repo.Route.WithContext(ctx).Create(routes...)
}

// UpdateRoute 更新路由
func UpdateRoute(ctx context.Context, route model.Route) error {
	u := repo.Route
	_, err := u.WithContext(ctx).Where(u.ID.Eq(route.ID)).Select(
		u.Name,
		u.PluginConfigID,
		u.ServiceID,
		u.UpstreamID,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(route)
	return err
}

// GetRoute 查询路由详情
func GetRoute(ctx context.Context, id string) (*model.Route, error) {
	u := repo.Route
	return u.WithContext(ctx).Where(u.ID.Eq(id)).First()
}

// QueryRoutes 搜索路由
func QueryRoutes(ctx context.Context, param map[string]interface{}) ([]*model.Route, error) {
	u := repo.Route
	return u.WithContext(ctx).Where(field.Attrs(param)).Find()
}

// BatchDeleteRoutes 批量删除路由 并记录审计日志
func BatchDeleteRoutes(ctx context.Context, ids []string) error {
	u := repo.Route
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.Route, ids)
		if err != nil {
			return err
		}
		// 批量删除路由关联的自定义插件记录
		err = BatchDeleteResourceSchemaAssociation(ctx, ids, constant.Route)
		if err != nil {
			return err
		}
		_, err = tx.Route.WithContext(ctx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// GetRouteCount 查询网关路由数量
func GetRouteCount(ctx context.Context, gatewayID int) (int64, error) {
	u := repo.Route
	return u.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Count()
}

// BatchRevertRoutes 批量回滚路由
func BatchRevertRoutes(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	routes, err := QueryRoutes(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(routes))
	for _, route := range routes {
		// 标识此次更新的类型为撤销
		route.OperationType = constant.OperationTypeRevert
		if route.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			route.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     route.ID,
				Config: route.Config,
				Status: route.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[route.ID]; ok {
			route.Name = gjson.ParseBytes(syncData.Config).Get("name").String()
			route.Config = syncData.Config
			route.Status = constant.ResourceStatusSuccess
			// 更新关联关系数据
			route.PluginConfigID = gjson.ParseBytes(syncData.Config).Get("plugin_config_id").String()
			route.UpstreamID = gjson.ParseBytes(syncData.Config).Get("upstream_id").String()
			route.ServiceID = gjson.ParseBytes(syncData.Config).Get("service_id").String()
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     route.ID,
				Config: route.Config,
				Status: route.Status,
			})
			continue
		} else {
			return errors.New("can not find sync data for route id:" + route.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.Route, ids, afterResources)
		if err != nil {
			return err
		}
		for _, route := range routes {
			_, err := tx.Route.WithContext(ctx).Updates(route)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
