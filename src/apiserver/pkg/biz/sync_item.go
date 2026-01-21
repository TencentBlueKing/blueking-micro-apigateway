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

	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// buildSyncedItemQuery 获取查询同步资源列表的 query
func buildSyncedItemQuery(ctx context.Context) repo.IGatewaySyncDataDo {
	return repo.GatewaySyncData.WithContext(ctx).Where(field.Attrs(map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// ListPagedSyncedItems 分页查询同步资源列表
func ListPagedSyncedItems(
	ctx context.Context,
	param map[string]any,
	page PageParam,
) ([]*model.GatewaySyncData, int64, error) {
	u := repo.GatewaySyncData
	return buildSyncedItemQuery(ctx).
		Where(field.Attrs(param)).
		Order(u.CreatedAt.Desc()).
		FindByPage(page.Offset, page.Limit)
}

// QuerySyncedItems 查询同步资源
func QuerySyncedItems(ctx context.Context, param map[string]any) ([]*model.GatewaySyncData, error) {
	return buildSyncedItemQuery(ctx).Where(field.Attrs(param)).Find()
}

// GetSyncedItemByID 通过 ID 获取同步资源
func GetSyncedItemByID(ctx context.Context, id string) (*model.GatewaySyncData, error) {
	u := repo.GatewaySyncData
	return buildSyncedItemQuery(ctx).Where(u.ID.Eq(id)).Take()
}
