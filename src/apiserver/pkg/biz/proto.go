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

// ListProtos 查询网关 Proto 列表
func ListProtos(ctx context.Context, gatewayID int) ([]*model.Proto, error) {
	u := repo.Proto
	return repo.Proto.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Order(u.UpdatedAt.Desc()).Find()
}

// GetProtoOrderExprList 获取 Proto 排序字段列表
func GetProtoOrderExprList(orderBy string) []field.Expr {
	u := repo.Proto
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

// ListPagedProtos 分页查询 Proto
func ListPagedProtos(
	ctx context.Context,
	param map[string]interface{},
	status []string,
	name string,
	updater string,
	orderBy string,
	page PageParam,
) ([]*model.Proto, int64, error) {
	u := repo.Proto
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
	orderByExprs := GetProtoOrderExprList(orderBy)
	return query.Where(field.Attrs(param)).
		Order(orderByExprs...).
		FindByPage(page.Offset, page.Limit)
}

// CreateProto 创建 Proto
func CreateProto(ctx context.Context, proto model.Proto) error {
	return repo.Proto.WithContext(ctx).Create(&proto)
}

// BatchCreateProtos 批量创建 Proto
func BatchCreateProtos(ctx context.Context, protos []*model.Proto) error {
	return repo.Proto.WithContext(ctx).Create(protos...)
}

// UpdateProto 更新 Proto
func UpdateProto(ctx context.Context, proto model.Proto) error {
	u := repo.Proto
	_, err := u.WithContext(ctx).Where(u.ID.Eq(proto.ID)).Select(
		u.Name,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(proto)
	return err
}

// GetProto 查询 Proto 详情
func GetProto(ctx context.Context, id string) (*model.Proto, error) {
	u := repo.Proto
	return u.WithContext(ctx).Where(u.ID.Eq(id)).First()
}

// QueryProtos 搜索 Proto
func QueryProtos(ctx context.Context, param map[string]interface{}) ([]*model.Proto, error) {
	u := repo.Proto
	return u.WithContext(ctx).Where(field.Attrs(param)).Find()
}

// ExistsProto 查询 Proto 是否存在
func ExistsProto(ctx context.Context, id string) bool {
	u := repo.Proto
	proto, err := u.WithContext(ctx).Where(
		u.ID.Eq(id),
		u.GatewayID.Eq(ginx.GetGatewayInfoFromContext(ctx).ID),
	).Find()
	if err != nil {
		return false
	}
	if len(proto) == 0 {
		return false
	}
	return true
}

// BatchDeleteProtos 批量删除 Proto 并添加审计日志
func BatchDeleteProtos(ctx context.Context, ids []string) error {
	u := repo.Proto
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		err := AddDeleteResourceByIDAuditLog(ctx, constant.Proto, ids)
		if err != nil {
			return err
		}
		_, err = u.WithContext(ctx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// BatchRevertProtos 批量回滚 Proto
func BatchRevertProtos(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	protos, err := QueryProtos(ctx, map[string]interface{}{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	for _, pb := range protos {
		if pb.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			pb.Status = constant.ResourceStatusSuccess
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[pb.ID]; ok {
			pb.Name = gjson.ParseBytes(syncData.Config).Get("name").String()
			pb.Config = syncData.Config
			pb.Status = constant.ResourceStatusSuccess
			continue
		} else {
			return errors.New("can not find sync data for Proto id:" + pb.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		for _, pb := range protos {
			err := repo.Proto.WithContext(ctx).Save(pb)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
