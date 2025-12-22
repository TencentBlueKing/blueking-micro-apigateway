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

	"github.com/pkg/errors"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/sslx"
)

// buildSSLQuery 获取 SSL 查询对象
func buildSSLQuery(ctx context.Context) repo.ISSLDo {
	return repo.SSL.WithContext(ctx).Where(field.Attrs(map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// buildSSLQueryWithTx 获取 SSL 查询对象（带事务）
func buildSSLQueryWithTx(ctx context.Context, tx *repo.Query) repo.ISSLDo {
	return tx.SSL.WithContext(ctx).Where(field.Attrs(map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// ListSSL 查询网关 ssl 列表
func ListSSL(ctx context.Context) ([]*model.SSL, error) {
	u := repo.SSL
	return buildSSLQuery(ctx).Order(u.UpdatedAt.Desc()).Find()
}

// GetSSLOrderExprList 获取 ssl 排序字段列表
func GetSSLOrderExprList(orderBy string) []field.Expr {
	u := repo.SSL
	ascFieldMap := map[string]field.Expr{
		"name":       u.Name.Asc(),
		"updated_at": u.UpdatedAt.Asc(),
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

// ListPagedSSL 分页查询 ssl
func ListPagedSSL(
	ctx context.Context,
	param map[string]any,
	label map[string][]string,
	status []string,
	name string,
	updater string,
	orderBy string,
	page PageParam,
) ([]*model.SSL, int64, error) {
	u := repo.SSL
	query := buildSSLQuery(ctx)
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	orderByExprs := GetSSLOrderExprList(orderBy)
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

// ParseCert 解析证书
func ParseCert(ctx context.Context, name, cert, key string) (*entity.SSL, error) {
	snis, err := sslx.ParseCert(cert, key)
	if err != nil {
		return nil, err
	}
	validity, err := sslx.X509CertValidity(cert)
	if err != nil {
		return nil, err
	}
	sslInfo := &entity.SSL{
		Cert:          cert,
		Key:           key,
		Snis:          snis,
		ValidityEnd:   validity.NotAfter,
		ValidityStart: validity.NotBefore,
		Status:        constant.SSLDefaultStatus,
		BaseInfo: entity.BaseInfo{
			Name: name, // 证书名称
			ID:   idx.GenResourceID(constant.SSL),
		},
	}
	return sslInfo, nil
}

// CreateSSL 创建 SSL
func CreateSSL(ctx context.Context, sslModel *model.SSL) error {
	return repo.SSL.WithContext(ctx).Create(sslModel)
}

// BatchCreateSSL 批量创建 SSL
func BatchCreateSSL(ctx context.Context, ssls []*model.SSL) error {
	if ginx.GetTx(ctx) != nil {
		return buildSSLQueryWithTx(ctx, ginx.GetTx(ctx)).Create(ssls...)
	}
	return repo.SSL.WithContext(ctx).Create(ssls...)
}

// UpdateSSL 更新 SSL
func UpdateSSL(ctx context.Context, ssl *model.SSL) error {
	u := repo.SSL
	_, err := buildSSLQuery(ctx).Where(u.ID.Eq(ssl.ID)).Updates(ssl)
	return err
}

// GetSSL 查询 SSL 详情
func GetSSL(ctx context.Context, id string) (*model.SSL, error) {
	u := repo.SSL
	return buildSSLQuery(ctx).Where(u.ID.Eq(id)).First()
}

// BatchRevertSSLs 批量回滚 ssl
func BatchRevertSSLs(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	ssls, err := QuerySSL(ctx, map[string]any{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(ssls))
	for _, ssl := range ssls {
		// 标识此次更新的操作类型为撤销
		ssl.OperationType = constant.OperationTypeRevert
		if ssl.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			ssl.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     ssl.ID,
				Config: ssl.Config,
				Status: ssl.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[ssl.ID]; ok {
			ssl.Name = syncData.GetName()
			ssl.Config = syncData.Config
			ssl.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     ssl.ID,
				Config: ssl.Config,
				Status: ssl.Status,
			})
			continue
		} else {
			return errors.New("can not find sync data for ssl id:" + ssl.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.SSL, ids, afterResources)
		if err != nil {
			return err
		}
		for _, sls := range ssls {
			_, err := buildSSLQueryWithTx(ctx, tx).Updates(sls)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// QuerySSL 搜索 SSL
func QuerySSL(ctx context.Context, param map[string]any) ([]*model.SSL, error) {
	return buildSSLQuery(ctx).Where(field.Attrs(param)).Find()
}

// ExistsSSL 查询 SSL 是否存在
func ExistsSSL(ctx context.Context, id string) bool {
	u := repo.SSL
	ssl, err := buildSSLQuery(ctx).Where(
		u.ID.Eq(id),
	).Find()
	if err != nil {
		return false
	}
	if len(ssl) == 0 {
		return false
	}
	return true
}

// BatchDeleteSSL 批量删除 SSL 并添加审计日志
func BatchDeleteSSL(ctx context.Context, ids []string) error {
	u := repo.SSL
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err := AddDeleteResourceByIDAuditLog(ctx, constant.SSL, ids)
		if err != nil {
			return err
		}
		_, err = buildSSLQueryWithTx(ctx, tx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}
