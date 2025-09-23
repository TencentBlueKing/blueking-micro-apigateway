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
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// FuncUpdateResourceStatusByID ...
type FuncUpdateResourceStatusByID func(ctx context.Context,
	resourceType constant.APISIXResource, id string, status constant.ResourceStatus) error

// FuncBatchUpdateResourceStatus ...
type FuncBatchUpdateResourceStatus func(ctx context.Context,
	resourceType constant.APISIXResource, ids []string, status constant.ResourceStatus) error

// FuncDeleteResourceByID ...
type FuncDeleteResourceByID func(ctx context.Context, ids []string) error

// AddBatchAuditLog ... 添加批量审计日志,适用于批量操作只改变状态的情况
func AddBatchAuditLog(ctx context.Context, operationType constant.OperationType, resourceType constant.APISIXResource,
	resources []*model.ResourceCommonModel,
	resourceIDStatusAfterMap map[string]constant.ResourceStatus,
) error {
	if len(resources) == 0 {
		return nil
	}
	var dataBefore []model.BatchOperationData
	var dataAfter []model.BatchOperationData
	var resourceIDs []string
	for _, resource := range resources {
		resourceIDs = append(resourceIDs, resource.ID)
		dataBefore = append(dataBefore, model.BatchOperationData{
			ID:     resource.ID,
			Status: resource.Status,
			Config: json.RawMessage(resource.Config),
		})
		if operationType != constant.OperationTypeDelete {
			dataAfter = append(dataAfter, model.BatchOperationData{
				ID:     resource.ID,
				Status: resourceIDStatusAfterMap[resource.ID],
				// 配置没有改变
				Config: json.RawMessage(resource.Config),
			})
		}
	}

	dataBeforeRaw, err := json.Marshal(dataBefore)
	if err != nil {
		return errors.Wrap(err, "marshal dataBefore failed")
	}

	dataAfterRaw, err := json.Marshal(dataAfter)
	if err != nil {
		return errors.Wrap(err, "marshal dataAfter failed")
	}

	operationAuditLog := &model.OperationAuditLog{
		GatewayID:     ginx.GetGatewayInfoFromContext(ctx).ID,
		ResourceType:  resourceType,
		OperationType: operationType,
		ResourceIDs:   strings.Join(resourceIDs, ","),
		DataBefore:    dataBeforeRaw,
		DataAfter:     dataAfterRaw,
		Operator:      ginx.GetUserIDFromContext(ctx),
	}
	return repo.OperationAuditLog.WithContext(ctx).Create(operationAuditLog)
}

// WrapUpdateResourceStatusByIDAddAuditLog ... 更新资源状态时添加审计日志
func WrapUpdateResourceStatusByIDAddAuditLog(ctx context.Context, resourceType constant.APISIXResource,
	id string, status constant.ResourceStatus, fn FuncUpdateResourceStatusByID,
) error {
	// 查询之前的配置
	resourceInfo, err := GetResourceByID(ctx, resourceType, id)
	if err != nil {
		return err
	}
	err = fn(ctx, resourceType, id, status)
	if err != nil {
		return err
	}
	// 添加审计日志
	err = AddBatchAuditLog(
		ctx,
		constant.OperationTypeUpdate,
		resourceType,
		[]*model.ResourceCommonModel{&resourceInfo},
		map[string]constant.ResourceStatus{id: status},
	)
	if err != nil {
		return err
	}
	return nil
}

// WrapBatchUpdateResourceStatusAddAuditLog ... 批量更新资源状态时添加审计日志
func WrapBatchUpdateResourceStatusAddAuditLog(ctx context.Context, resourceType constant.APISIXResource,
	resourceIDs []string, status constant.ResourceStatus, fn FuncBatchUpdateResourceStatus,
) error {
	// 查询之前的配置
	resourceList, err := BatchGetResources(ctx, resourceType, resourceIDs)
	if err != nil {
		return err
	}
	err = fn(ctx, resourceType, resourceIDs, status)
	if err != nil {
		return err
	}
	resourceStatusMap := make(map[string]constant.ResourceStatus)
	for _, resource := range resourceList {
		// 操作之后的状态映射
		resourceStatusMap[resource.ID] = status
	}
	// 添加审计日志
	err = AddBatchAuditLog(ctx, constant.OperationTypeUpdate, resourceType, resourceList,
		resourceStatusMap)
	if err != nil {
		return err
	}
	return fn(ctx, resourceType, resourceIDs, status)
}

// AddDeleteResourceByIDAuditLog ... 删除资源时添加审计日志
func AddDeleteResourceByIDAuditLog(ctx context.Context, resourceType constant.APISIXResource,
	resourceIDs []string,
) error {
	// 查询之前的配置
	resourceList, err := BatchGetResources(ctx, resourceType, resourceIDs)
	if err != nil {
		return err
	}
	resourceStatusMap := make(map[string]constant.ResourceStatus)
	for _, resource := range resourceList {
		// 操作之后的状态映射
		if resource.Status == constant.ResourceStatusCreateDraft {
			// 草稿状态直接删除
			resourceStatusMap[resource.ID] = constant.ResourceStatusDeleted
			continue
		}
		resourceStatusMap[resource.ID] = constant.ResourceStatusDeleteDraft
	}
	// 添加审计日志
	err = AddBatchAuditLog(ctx, constant.OperationTypeDelete, resourceType, resourceList, resourceStatusMap)
	if err != nil {
		return err
	}
	return nil
}

// ListOperationAuditLogs 查询操作审计列表
func ListOperationAuditLogs(
	ctx context.Context,
	param map[string]interface{},
	resourceID string,
	operator string,
	timeStart int,
	timeEnd int,
) ([]*model.OperationAuditLog, error) {
	u := repo.OperationAuditLog
	query := u.WithContext(ctx)
	if resourceID != "" {
		query = query.Where(u.ResourceIDs.Like("%" + resourceID + "%"))
	}
	if operator != "" {
		query = query.Where(u.Operator.Like("%" + operator + "%"))
	}
	if timeStart != 0 && timeEnd != 0 {
		query = query.Where(u.CreatedAt.Between(
			time.Unix(int64(timeStart), 0),
			time.Unix(int64(timeEnd), 0)),
		)
	}
	return query.Where(field.Attrs(param)).Order(u.CreatedAt.Desc()).Find()
}

// ListPagedOperationAuditLogs 分页查询 操作审计列表
func ListPagedOperationAuditLogs(
	ctx context.Context,
	param map[string]interface{},
	resourceID string,
	operator string,
	timeStart int,
	timeEnd int,
	page PageParam,
) ([]*model.OperationAuditLog, int64, error) {
	u := repo.OperationAuditLog
	query := u.WithContext(ctx)
	if resourceID != "" {
		query = query.Where(u.ResourceIDs.Like("%" + resourceID + "%"))
	}
	if operator != "" {
		query = query.Where(u.Operator.Like("%" + operator + "%"))
	}
	if timeStart != 0 && timeEnd != 0 {
		query = query.Where(u.CreatedAt.Between(
			time.Unix(int64(timeStart), 0),
			time.Unix(int64(timeEnd), 0)),
		)
	}
	return query.Where(field.Attrs(param)).Order(u.CreatedAt.Desc()).FindByPage(page.Offset, page.Limit)
}
