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

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// buildGlobalRuleQuery 获取 GlobalRule 查询对象
func buildGlobalRuleQuery(ctx context.Context) repo.IGlobalRuleDo {
	return repo.GlobalRule.WithContext(ctx).Where(field.Attrs(map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// buildGlobalRuleQueryWithTx 获取 GlobalRule 查询对象(带事务)
func buildGlobalRuleQueryWithTx(ctx context.Context, tx *repo.Query) repo.IGlobalRuleDo {
	return tx.GlobalRule.WithContext(ctx).Where(field.Attrs(map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// ListGlobalRules 查询网关 GlobalRule 列表
func ListGlobalRules(ctx context.Context) ([]*model.GlobalRule, error) {
	u := repo.GlobalRule
	return buildGlobalRuleQuery(ctx).Order(u.UpdatedAt.Desc()).Find()
}

// GetGlobalRuleOrderExprList 获取 GlobalRule 排序字段列表
func GetGlobalRuleOrderExprList(orderBy string) []field.Expr {
	u := repo.GlobalRule
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

// ListPagedGlobalRules 分页查询 GlobalRule 列表
func ListPagedGlobalRules(
	ctx context.Context,
	param map[string]any,
	status []string,
	name string,
	updater string,
	orderBy string,
	page PageParam,
) ([]*model.GlobalRule, int64, error) {
	u := repo.GlobalRule
	query := buildGlobalRuleQuery(ctx)
	if len(status) > 1 || status[0] != "" {
		query = query.Where(u.Status.In(status...))
	}
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	orderByExprs := GetGlobalRuleOrderExprList(orderBy)
	return query.Where(field.Attrs(param)).Order(orderByExprs...).FindByPage(page.Offset, page.Limit)
}

// CreateGlobalRule 创建 GlobalRule
func CreateGlobalRule(ctx context.Context, globalRule model.GlobalRule) error {
	return repo.GlobalRule.WithContext(ctx).Create(&globalRule)
}

// BatchCreateGlobalRules 批量创建 GlobalRule
func BatchCreateGlobalRules(ctx context.Context, globalRules []*model.GlobalRule) error {
	if ginx.GetTx(ctx) != nil {
		return buildGlobalRuleQueryWithTx(ctx, ginx.GetTx(ctx)).Create(globalRules...)
	}
	return repo.GlobalRule.WithContext(ctx).Create(globalRules...)
}

// UpdateGlobalRule 更新 GlobalRule
func UpdateGlobalRule(ctx context.Context, globalRule model.GlobalRule) error {
	u := repo.GlobalRule
	_, err := buildGlobalRuleQuery(ctx).Where(u.ID.Eq(globalRule.ID)).Select(
		u.Name,
		u.Config,
		u.Status,
		u.Updater,
	).Updates(globalRule)
	return err
}

// GetGlobalRule 查询 GlobalRule 详情
func GetGlobalRule(ctx context.Context, id string) (*model.GlobalRule, error) {
	u := repo.GlobalRule
	return buildGlobalRuleQuery(ctx).Where(u.ID.Eq(id)).First()
}

// QueryGlobalRules 搜索 GlobalRule
func QueryGlobalRules(ctx context.Context, param map[string]any) ([]*model.GlobalRule, error) {
	return buildGlobalRuleQuery(ctx).Where(field.Attrs(param)).Find()
}

// BatchDeleteGlobalRules 批量删除 GlobalRule 并添加审计日志
func BatchDeleteGlobalRules(ctx context.Context, ids []string) error {
	u := repo.GlobalRule
	err := repo.Q.Transaction(func(tx *repo.Query) error {
		err := AddDeleteResourceByIDAuditLog(ctx, constant.GlobalRule, ids)
		if err != nil {
			return err
		}
		// 批量删除 GlobalRule 关联的自定义插件记录
		err = BatchDeleteResourceSchemaAssociation(ctx, ids, constant.GlobalRule)
		if err != nil {
			return err
		}
		_, err = buildGlobalRuleQueryWithTx(ctx, tx).Where(u.ID.In(ids...)).Delete()
		return err
	})
	return err
}

// BatchRevertGlobalRules 批量回滚 GlobalRule
func BatchRevertGlobalRules(ctx context.Context, syncDataList []*model.GatewaySyncData) error {
	var ids []string
	syncResourceMap := make(map[string]*model.GatewaySyncData)
	for _, syncData := range syncDataList {
		ids = append(ids, syncData.ID)
		syncResourceMap[syncData.ID] = syncData
	}
	// 查询原来的数据
	globalRules, err := QueryGlobalRules(ctx, map[string]any{
		"id": ids,
		"status": []constant.ResourceStatus{
			constant.ResourceStatusDeleteDraft,
			constant.ResourceStatusUpdateDraft,
		},
	})
	if err != nil {
		return err
	}
	afterResources := make([]*model.ResourceCommonModel, 0, len(globalRules))
	for _, globalRule := range globalRules {
		// 标识此次更新的操作类型为撤销
		globalRule.OperationType = constant.OperationTypeRevert
		if globalRule.Status == constant.ResourceStatusDeleteDraft {
			// 删除待发布回滚只需要更新状态即可
			globalRule.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     globalRule.ID,
				Config: globalRule.Config,
				Status: globalRule.Status,
			})
			continue
		}
		// 同步更新配置
		if syncData, ok := syncResourceMap[globalRule.ID]; ok {
			globalRule.Name = syncData.GetName()
			globalRule.Config = syncData.Config
			globalRule.Status = constant.ResourceStatusSuccess
			// 用于审计日志更新，只需要补充 ID, Config, Status 即可
			afterResources = append(afterResources, &model.ResourceCommonModel{
				ID:     globalRule.ID,
				Config: globalRule.Config,
				Status: globalRule.Status,
			})
		} else {
			return errors.New("can not find sync data for globalRule id:" + globalRule.ID)
		}
	}
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 添加撤销的审计日志
		err = WrapBatchRevertResourceAddAuditLog(ctx, constant.GlobalRule, ids, afterResources)
		if err != nil {
			return err
		}
		for _, globalRule := range globalRules {
			_, err := buildGlobalRuleQueryWithTx(ctx, tx).Updates(globalRule)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// GetGlobalRulePluginToID 获取 global rule 配置的插件映射
func GetGlobalRulePluginToID(ctx context.Context) (map[string]dto.GlobalRulePlugin, error) {
	globalRules, err := ListGlobalRules(ctx)
	if err != nil {
		return nil, err
	}
	pluginMap := make(map[string]dto.GlobalRulePlugin)
	for _, globalRule := range globalRules {
		plugins := gjson.ParseBytes(globalRule.Config).Get("plugins")
		globalRulePlugin := dto.GlobalRulePlugin{
			ID: globalRule.ID,
		}
		pluginConfigs := plugins.Array()
		for _, p := range pluginConfigs {
			globalRulePlugin.Config = json.RawMessage(p.Raw)
		}
		plugins.ForEach(func(key, value gjson.Result) bool {
			pluginMap[key.String()] = globalRulePlugin
			return true
		})
	}
	return pluginMap, nil
}
