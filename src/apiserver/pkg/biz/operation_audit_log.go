/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
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
	"time"

	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils"
)

// ListOperationAuditLogs 查询操作审计列表
func ListOperationAuditLogs(
	ctx context.Context,
	param map[string]any,
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
	param map[string]any,
	resourceID string,
	operator string,
	timeStart int,
	timeEnd int,
	page utils.PageParam,
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
