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

// ListServices 查询网关 Service 列表
func ListServices(ctx context.Context, gatewayID int) ([]*model.Service, error) {
	u := repo.Service
	return repo.Service.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Order(u.UpdatedAt.Desc()).Find()
}

// GetServiceOrderExprList 获取 Service 排序字段列表
func GetServiceOrderExprList(orderBy string) []field.Expr {
	u := repo.Service
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

// ListPagedServices 分页查询网 service 列表
func ListPagedServices(
	ctx context.Context,
	param map[string]interface{},
	label map[string][]string,
	status []string,
	name string,
	updater string,
	upstreamID string,
	orderBy string,
	page PageParam,
) ([]*model.Service, int64, error) {
	u := repo.Service
	query := u.WithContext(ctx)
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	associationIDCond := u.WithContext(ctx).Clauses()
	if upstreamID != "" {
		if upstreamID == constant.EmptyAssociationFilter {
			associationIDCond.Where(u.UpstreamID.Eq("")).Or(u.UpstreamID.IsNull())
		} else {
			query = query.Where(u.UpstreamID.Eq(upstreamID))
		}
	}
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	orderByExprs := GetServiceOrderExprList(orderBy)
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

// CreateService 创建 service
func CreateService(ctx context.Context, service model.Service) error {
	return repo.Service.WithContext(ctx).Create(&service)
}

// BatchCreateServices 批量创建 service
func BatchCreateServices(ctx context.Context, services []*model.Service) error {
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).Service.WithContext(ctx).Create(services...)
	}
	return repo.Service.WithContext(ctx).Create(services...)
}

// UpdateService 更新 Service
func UpdateService(ctx context.Context, service model.Service) error {
	u := repo.Service
	_, err := u.WithContext(ctx).Where(u.ID.Eq(service.ID)).Select(
		u.Name,
		u.UpstreamID,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(service)
	return err
}

// GetService 查询 Service 详情
func GetService(ctx context.Context, id string) (*model.Service, error) {
	u := repo.Service
	return u.WithContext(ctx).Where(u.ID.Eq(id)).First()
}

// QueryServices 搜索 service
func QueryServices(ctx context.Context, param map[string]interface{}) ([]*model.Service, error) {
	u := repo.Service
	return u.WithContext(ctx).Where(field.Attrs(param)).Find()
}

// ExistsService 查询 Service 是否存在
func ExistsService(ctx context.Context, id string) bool {
	u := repo.Service
	services, err := u.WithContext(ctx).Where(
		u.ID.Eq(id),
		u.GatewayID.Eq(ginx.GetGatewayInfoFromContext(ctx).ID),
	).Find()
	if err != nil {
		return false
	}
	if len(services) == 0 {
		return false
	}
	return true
}

// BatchDeleteServices 批量删除 service 并记录审计日志
func BatchDeleteServices(ctx context.Context, ids []string) error {
	u := repo.Service
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.Service, ids)
		if err != nil {
			return err
		}
		// 批量删除 service 关联的自定义插件记录
		err = BatchDeleteResourceSchemaAssociation(ctx, ids, constant.Service)
		if err != nil {
			return err
		}
		_, err = tx.Service.WithContext(ctx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// GetServiceCount 查询网关 Service 数量
func GetServiceCount(ctx context.Context, gatewayID int) (int64, error) {
	u := repo.Service
	return u.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Count()
}

// BatchRevertServices 批量回滚 Service
func BatchRevertServices(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	services, err := QueryServices(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	for _, service := range services {
		if service.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			service.Status = constant.ResourceStatusSuccess
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[service.ID]; ok {
			service.Name = gjson.ParseBytes(syncData.Config).Get("name").String()
			service.Config = syncData.Config
			service.Status = constant.ResourceStatusSuccess
			// 更新关联关系数据
			service.UpstreamID = gjson.ParseBytes(syncData.Config).Get("upstream_id").String()
			continue
		} else {
			return errors.New("can not find sync data for service id:" + service.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		for _, service := range services {
			err := tx.Service.WithContext(ctx).Save(service)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
