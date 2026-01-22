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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// Syncer ...
type Syncer struct {
	SystemItemChannel chan []*model.GatewaySyncData
	ctx               context.Context
}

// NewSyncer 创建 Syncer 实例
func NewSyncer(ctx context.Context) *Syncer {
	return &Syncer{
		SystemItemChannel: make(chan []*model.GatewaySyncData, 100),
		ctx:               ctx,
	}
}

// Run 启动同步器
func (s *Syncer) Run() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case resourceList := <-s.SystemItemChannel:
			ctx := context.Background()
			u := repo.GatewaySyncData
			err := repo.Q.Transaction(func(tx *repo.Query) error {
				if len(resourceList) == 0 {
					return nil
				}
				// 先删除后插入
				_, err := tx.GatewaySyncData.WithContext(
					ctx,
				).Where(
					u.GatewayID.Eq(resourceList[0].GatewayID),
				).
					Delete()
				if err != nil {
					return err
				}
				return tx.GatewaySyncData.WithContext(ctx).CreateInBatches(resourceList, 500)
			})
			if err != nil {
				logging.Errorf(
					"sync gateway:%d resource error: %s",
					resourceList[0].GatewayID,
					err.Error(),
				)
			}
		}
	}
}

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

// GetSyncedItemByResourceTypeAndID 通过 ResourceType 和 ID 获取同步资源
func GetSyncedItemByResourceTypeAndID(
	ctx context.Context,
	resourceType constant.APISIXResource,
	id string,
) (*model.GatewaySyncData, error) {
	u := repo.GatewaySyncData
	return buildSyncedItemQuery(ctx).Where(u.Type.Eq(string(resourceType)), u.ID.Eq(id)).Take()
}
