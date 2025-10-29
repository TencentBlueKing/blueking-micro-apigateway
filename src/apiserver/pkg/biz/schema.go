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

	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// buildSchemaQuery 获取 GatewayCustomPluginSchema 查询对象
func buildSchemaQuery(ctx context.Context) repo.IGatewayCustomPluginSchemaDo {
	return repo.GatewayCustomPluginSchema.WithContext(ctx).Where(field.Attrs(map[string]interface{}{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}))
}

// buildSchemaQueryWithTx 获取 GatewayCustomPluginSchema 查询对象(带事务)
/**
 * buildSchemaQueryWithTx creates a query for GatewayCustomPluginSchema with transaction context
 * @param ctx context.Context - The context containing request information
 * @param tx *repo.Query - The transaction/query object
 * @return repo.IGatewayCustomPluginSchemaDo - Returns a query interface for GatewayCustomPluginSchema operations
 */
func buildSchemaQueryWithTx(ctx context.Context, tx *repo.Query) repo.IGatewayCustomPluginSchemaDo {
	// Create query with context and filter by gateway_id from context
	return tx.WithContext(ctx).GatewayCustomPluginSchema.Where(field.Attrs(map[string]interface{}{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID, // Get gateway ID from context and use as filter
	}))
}

// ListSchema 查询网关 schema 列表
func ListSchema(ctx context.Context) ([]*model.GatewayCustomPluginSchema, error) {
	u := repo.GatewayCustomPluginSchema
	return buildSchemaQuery(ctx).Order(u.UpdatedAt.Desc()).Find()
}

// GetSchemaExprList 获取 schema 排序字段列表
func GetSchemaExprList(orderBy string) []field.Expr {
	u := repo.GatewayCustomPluginSchema
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

// ListPagedSchema 分页查询 schema
func ListPagedSchema(
ctx context.Context,
name string,
updater string,
orderBy string,
page PageParam,
) ([]*model.GatewayCustomPluginSchema, int64, error) {
	u := repo.GatewayCustomPluginSchema
	query := buildSchemaQuery(ctx)
	if name != "" {
		query = query.Where(u.Name.Like("%" + name + "%"))
	}
	if updater != "" {
		query = query.Where(u.Updater.Like("%" + updater + "%"))
	}
	orderByExprs := GetSchemaExprList(orderBy)
	return query.Order(orderByExprs...).
		FindByPage(page.Offset, page.Limit)
}

// CreateSchema 创建 schema
func CreateSchema(ctx context.Context, schema *model.GatewayCustomPluginSchema) error {
	return repo.GatewayCustomPluginSchema.WithContext(ctx).Create(schema)
}

// BatchCreateSchema 批量创建 schema
func BatchCreateSchema(ctx context.Context, schemas []*model.GatewayCustomPluginSchema) error {
	if ginx.GetTx(ctx) != nil {
		return buildSchemaQueryWithTx(ctx, ginx.GetTx(ctx)).CreateInBatches(
			schemas, constant.DBBatchCreateSize)
	}
	return repo.GatewayCustomPluginSchema.WithContext(ctx).CreateInBatches(schemas, constant.DBBatchCreateSize)
}

// UpdateSchema 更新 schema
func UpdateSchema(ctx context.Context, schema model.GatewayCustomPluginSchema) error {
	u := repo.GatewayCustomPluginSchema
	_, err := buildSchemaQuery(ctx).Where(u.AutoID.Eq(schema.AutoID)).Select(
		u.Name,
		u.Schema,
		u.Example,
		u.Updater,
	).Updates(schema)
	return err
}

// GetSchemaByName 根据 name 查询 schema 详情
func GetSchemaByName(ctx context.Context, name string) (*model.GatewayCustomPluginSchema, error) {
	u := repo.GatewayCustomPluginSchema
	schemaInfo, err := buildSchemaQuery(ctx).
		Where(u.Name.Eq(name)).First()
	return schemaInfo, err
}

// GetSchemaByID 根据 id 查询 schema 详情
func GetSchemaByID(ctx context.Context, id int) (*model.GatewayCustomPluginSchema, error) {
	u := repo.GatewayCustomPluginSchema
	schemaInfo, err := buildSchemaQuery(ctx).
		Where(u.AutoID.Eq(id)).First()
	return schemaInfo, err
}

// DeleteSchemaByID 删除 schema
func DeleteSchemaByID(ctx context.Context, schemaID int) error {
	_, err := buildSchemaQuery(ctx).Delete(&model.GatewayCustomPluginSchema{AutoID: schemaID})
	return err
}

// DeleteSchemaByNames 根据 names 删除 schema
func DeleteSchemaByNames(ctx context.Context, names []string) error {
	u := repo.GatewayCustomPluginSchema
	if ginx.GetTx(ctx) != nil {
		_, err := buildSchemaQueryWithTx(ctx, ginx.GetTx(ctx)).Where(u.Name.In(names...)).Delete()
		if err != nil {
			return err
		}
		return nil
	}
	_, err := buildSchemaQuery(ctx).WithContext(ctx).Where(u.Name.In(names...)).Delete()
	return err
}

// DuplicatedSchemaName 查询插件名称是否重复
func DuplicatedSchemaName(
ctx context.Context,
id int,
name string,
) bool {
	u := repo.GatewayCustomPluginSchema
	query := buildSchemaQuery(ctx).Where(
		u.Name.Eq(name),
	)
	if id != 0 {
		query = query.Where(u.AutoID.Neq(id))
	}
	res, err := query.Find()
	if err != nil {
		return false
	}
	if len(res) == 0 {
		return true
	}
	return false
}

// GetCustomizePluginExampleList 查询自定义插件的示例列表
func GetCustomizePluginExampleList(ctx context.Context, gatewayID int) ([]*schema.Plugin, error) {
	var plugins []*schema.Plugin
	schemaList, err := ListSchema(ctx)
	if err != nil {
		return nil, err
	}
	for _, s := range schemaList {
		var plugin schema.Plugin
		var exampleMap map[string]interface{}
		err := json.Unmarshal(s.Example, &exampleMap)
		if err != nil {
			return nil, err
		}
		plugin.Name = s.Name
		plugin.Example = exampleMap
		plugin.Type = constant.CustomizePlugin
		plugins = append(plugins, &plugin)
	}
	return plugins, nil
}

// GetCustomizePluginSchemaMap 查询自定义插件 schema map
func GetCustomizePluginSchemaMap(ctx context.Context) (map[string]interface{}, error) {
	schemaList, err := ListSchema(ctx)
	if err != nil {
		return nil, err
	}
	pluginSchemaMap, err := GetCustomizePluginNameToSchemaMap(schemaList)
	if err != nil {
		return nil, err
	}
	return pluginSchemaMap, nil
}

// GetCustomizePluginNameToSchemaMap 查询自定义插件映射关系
func GetCustomizePluginNameToSchemaMap(schemaList []*model.GatewayCustomPluginSchema) (map[string]interface{}, error) {
	pluginSchemaMap := map[string]interface{}{}
	for _, s := range schemaList {
		var schemaInfo map[string]interface{}
		err := json.Unmarshal(s.Schema, &schemaInfo)
		if err != nil {
			return nil, err
		}
		pluginSchemaMap[s.Name] = schemaInfo
	}
	return pluginSchemaMap, nil
}

// GetCustomizePluginSchemaInfoMap 查询自定义插件 map
func GetCustomizePluginSchemaInfoMap(ctx context.Context) (map[string]*model.GatewayCustomPluginSchema, error) {
	schemaList, err := ListSchema(ctx)
	if err != nil {
		return nil, err
	}
	pluginSchemaMap := map[string]*model.GatewayCustomPluginSchema{}
	for _, s := range schemaList {
		pluginSchemaMap[s.Name] = s
	}
	return pluginSchemaMap, nil
}

// GetResourceSchemaAssociation 查询资源与自定义插件的关联记录
func GetResourceSchemaAssociation(
ctx context.Context,
schemaID int,
) ([]*model.GatewayResourceSchemaAssociation, error) {
	u := repo.GatewayResourceSchemaAssociation
	return u.WithContext(ctx).Where(u.SchemaID.Eq(schemaID)).Find()
}

// BatchDeleteResourceSchemaAssociation 批量删除资源与自定义插件的关联记录
func BatchDeleteResourceSchemaAssociation(
ctx context.Context,
resourceIDs []string,
resourceType constant.APISIXResource,
) error {
	u := repo.GatewayResourceSchemaAssociation
	if ginx.GetTx(ctx) != nil {
		_, err := ginx.GetTx(ctx).GatewayResourceSchemaAssociation.WithContext(ctx).Where(
			u.ResourceID.In(resourceIDs...),
			u.ResourceType.Eq(resourceType.String()),
		).Delete()
		return err
	}
	_, err := repo.GatewayResourceSchemaAssociation.WithContext(ctx).Where(
		u.ResourceID.In(resourceIDs...),
		u.ResourceType.Eq(resourceType.String()),
	).Delete()
	return err
}
