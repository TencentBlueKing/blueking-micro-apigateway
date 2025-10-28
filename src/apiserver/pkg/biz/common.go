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

// Package biz ...
package biz

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// PageParam 分页参数
type PageParam struct {
	Offset int
	Limit  int
}

var resourceTableMap = map[constant.APISIXResource]string{
	constant.Route:          model.Route{}.TableName(),
	constant.Upstream:       model.Upstream{}.TableName(),
	constant.Consumer:       model.Consumer{}.TableName(),
	constant.ConsumerGroup:  model.ConsumerGroup{}.TableName(),
	constant.PluginConfig:   model.PluginConfig{}.TableName(),
	constant.GlobalRule:     model.GlobalRule{}.TableName(),
	constant.PluginMetadata: model.PluginMetadata{}.TableName(),
	constant.Service:        model.Service{}.TableName(),
	constant.Proto:          model.Proto{}.TableName(),
	constant.SSL:            model.SSL{}.TableName(),
	constant.StreamRoute:    model.StreamRoute{}.TableName(),
}

var resourceModelSliceMap = map[constant.APISIXResource]interface{}{
	constant.Route:          &[]model.Route{},
	constant.Upstream:       &[]model.Upstream{},
	constant.Consumer:       &[]model.Consumer{},
	constant.ConsumerGroup:  &[]model.ConsumerGroup{},
	constant.PluginConfig:   &[]model.PluginConfig{},
	constant.GlobalRule:     &[]model.GlobalRule{},
	constant.PluginMetadata: &[]model.PluginMetadata{},
	constant.Service:        &[]model.Service{},
	constant.Proto:          &[]model.Proto{},
	constant.SSL:            &[]model.SSL{},
	constant.StreamRoute:    &[]model.StreamRoute{},
}

var resourceModelMap = map[constant.APISIXResource]interface{}{
	constant.Route:          &model.Route{},
	constant.Upstream:       &model.Upstream{},
	constant.Consumer:       &model.Consumer{},
	constant.ConsumerGroup:  &model.ConsumerGroup{},
	constant.PluginConfig:   &model.PluginConfig{},
	constant.GlobalRule:     &model.GlobalRule{},
	constant.PluginMetadata: &model.PluginMetadata{},
	constant.Service:        &model.Service{},
	constant.Proto:          &model.Proto{},
	constant.SSL:            &model.SSL{},
	constant.StreamRoute:    &model.StreamRoute{},
}

// Labels ...
type Labels map[string]string

// Scan 实现从数据库到结构体的转换
func (l *Labels) Scan(value interface{}) error {
	if value == nil {
		*l = nil
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("不支持的数据库类型: %T", v)
	}
	return json.Unmarshal(data, l)
}

// Value 实现从结构体到数据库的转换
func (l Labels) Value() (driver.Value, error) {
	if l == nil {
		return nil, nil
	}
	return json.Marshal(l)
}

// getCommonDbQuery 获取通用查询
// 该查询会根据网关 ID 进行过滤，确保只能查询当前网关下的资源
func getCommonDbQuery(
	ctx context.Context,
	resourceType constant.APISIXResource,
) *gorm.DB {
	return database.Client().WithContext(ctx).Table(resourceTableMap[resourceType]).Where(
		"gateway_id = ?", ginx.GetGatewayInfoFromContext(ctx).ID)
}

// BatchUpdateResourceStatusWithAuditLog 批量更新资源状态并添加审计日志
func BatchUpdateResourceStatusWithAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource, ids []string, status constant.ResourceStatus,
) error {
	return WrapBatchUpdateResourceStatusAddAuditLog(ctx, resourceType, ids, status, BatchUpdateResourceStatus)
}

// BatchDeleteResourceWithAuditLog 批量删除资源并添加审计日志
func BatchDeleteResourceWithAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource, ids []string,
) error {
	switch resourceType {
	case constant.Route:
		return BatchDeleteRoutes(ctx, ids)
	case constant.Service:
		return BatchDeleteServices(ctx, ids)
	case constant.Upstream:
		return BatchDeleteUpstreams(ctx, ids)
	case constant.Proto:
		return BatchDeleteProtos(ctx, ids)
	case constant.SSL:
		return BatchDeleteSSL(ctx, ids)
	case constant.Consumer:
		return BatchDeleteConsumers(ctx, ids)
	case constant.ConsumerGroup:
		return BatchDeleteConsumerGroups(ctx, ids)
	case constant.PluginMetadata:
		return BatchDeletePluginMetadatas(ctx, ids)
	case constant.GlobalRule:
		return BatchDeleteGlobalRules(ctx, ids)
	case constant.PluginConfig:
		return BatchDeletePluginConfigs(ctx, ids)
	case constant.StreamRoute:
		return BatchDeleteStreamRoutes(ctx, ids)
	}
	return nil
}

// BatchUpdateResourceStatus 批量更新资源状态
func BatchUpdateResourceStatus(
	ctx context.Context,
	resourceType constant.APISIXResource, ids []string, status constant.ResourceStatus,
) error {
	query := getCommonDbQuery(ctx, resourceType)
	// 如果 IDs 数量小于等于 DBConditionIDMaxLength，直接更新
	if len(ids) <= constant.DBConditionIDMaxLength {
		return query.Where("id IN (?)", ids).Updates(map[string]interface{}{
			"status": status,
		}).Error
	}

	// 分批处理大量 IDs
	for i := 0; i < len(ids); i += constant.DBConditionIDMaxLength {
		end := i + constant.DBConditionIDMaxLength
		if end > len(ids) {
			end = len(ids)
		}

		batchIDs := ids[i:end]
		err := query.Where("id IN (?)", batchIDs).Updates(map[string]interface{}{
			"status": status,
		}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateResourceStatus 单个更新状态
func UpdateResourceStatus(
	ctx context.Context,
	resourceType constant.APISIXResource, id string, status constant.ResourceStatus,
) error {
	query := getCommonDbQuery(ctx, resourceType)
	return query.Where("id = ?", id).Updates(map[string]interface{}{
		"status":  status,
		"updater": ginx.GetUserIDFromContext(ctx),
	}).Error
}

// UpdateResourceStatusWithAuditLog  更新资源状态并添加审计日志
func UpdateResourceStatusWithAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource, id string, status constant.ResourceStatus,
) error {
	return WrapUpdateResourceStatusByIDAddAuditLog(ctx, resourceType, id, status, UpdateResourceStatus)
}

// BatchGetResources 批量获取资源（修复分批次条件叠加问题）
func BatchGetResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
	ids []string,
) ([]*model.ResourceCommonModel, error) {
	var res []*model.ResourceCommonModel
	query := getCommonDbQuery(ctx, resourceType)

	// 空IDs直接返回
	if len(ids) == 0 {
		err := query.Find(&res).Error
		return res, err
	}

	// 直接查询短列表
	if len(ids) <= constant.DBConditionIDMaxLength {
		err := query.Where("id IN (?)", ids).Find(&res).Error
		return res, err
	}

	// 正确分批次逻辑：每个批次使用新的查询实例
	for i := 0; i < len(ids); i += constant.DBConditionIDMaxLength {
		end := i + constant.DBConditionIDMaxLength
		if end > len(ids) {
			end = len(ids)
		}
		batchIDs := ids[i:end]
		// 关键点：每个批次创建新查询，避免条件叠加
		batchQuery := getCommonDbQuery(ctx, resourceType)
		var batchRes []*model.ResourceCommonModel
		if err := batchQuery.Where("id IN (?)", batchIDs).Find(&batchRes).Error; err != nil {
			return nil, err
		}
		res = append(res, batchRes...)
	}

	return res, nil
}

// GetResourcesLabels 获取资源标签
func GetResourcesLabels(
	ctx context.Context,
	resourceType constant.APISIXResource,
) (map[string]string, error) {
	var labelsList []Labels
	err := getCommonDbQuery(ctx, resourceType).Select("JSON_EXTRACT(config, '$.labels') as labels").
		Scan(&labelsList).Error
	// 去重
	labelsMap := make(map[string]string)
	for _, labels := range labelsList {
		for k, v := range labels {
			key := fmt.Sprintf("%s:%s", k, v)
			if _, ok := labelsMap[key]; !ok {
				labelsMap[key] = v
			}
		}
	}
	return labelsMap, err
}

// BatchDeleteResource 批量删除资源
func BatchDeleteResource(ctx context.Context, resourceType constant.APISIXResource, ids []string) error {
	resourceList, err := BatchGetResources(ctx, resourceType, ids)
	if err != nil {
		return err
	}
	var deleteIDs []string
	var updateIDs []string
	for _, resource := range resourceList {
		// 新增待发布和success才能删除
		switch resource.Status {
		case constant.ResourceStatusCreateDraft:
			deleteIDs = append(deleteIDs, resource.ID)
		case constant.ResourceStatusSuccess:
			updateIDs = append(updateIDs, resource.ID)
		default:
			continue
		}
		statusOp := status.NewResourceStatusOp(*resource)
		err = statusOp.CanDo(ctx, constant.OperationTypeDelete)
		if err != nil {
			return fmt.Errorf("resource: %s can not do delete: %s", resource.ID, err.Error())
		}
	}
	err = BatchDeleteResourceWithAuditLog(ctx, resourceType, deleteIDs)
	if err != nil {
		return err
	}
	err = BatchUpdateResourceStatusWithAuditLog(ctx, resourceType, updateIDs, constant.ResourceStatusDeleteDraft)
	if err != nil {
		return err
	}
	return nil
}

// GetResourceByID 根据id获取资源
func GetResourceByID(
	ctx context.Context,
	resourceType constant.APISIXResource, id string,
) (model.ResourceCommonModel, error) {
	var res model.ResourceCommonModel
	query := getCommonDbQuery(ctx, resourceType)
	err := query.Where("id = ?", id).Take(&res).Error
	return res, err
}

// GetResourceByIDs 根据 ids 获取资源
func GetResourceByIDs(
	ctx context.Context,
	resourceType constant.APISIXResource,
	ids []string,
) ([]model.ResourceCommonModel, error) {
	var res []model.ResourceCommonModel
	query := getCommonDbQuery(ctx, resourceType)
	// 如果 IDs 数量小于等于 DBConditionIDMaxLength，直接查询
	if len(ids) <= constant.DBConditionIDMaxLength {
		err := query.Where("id IN ?", ids).Find(&res).Error
		return res, err
	}

	// 分批处理大量 IDs
	for i := 0; i < len(ids); i += constant.DBConditionIDMaxLength {
		end := i + constant.DBConditionIDMaxLength
		if end > len(ids) {
			end = len(ids)
		}

		batchIDs := ids[i:end]
		var batchRes []model.ResourceCommonModel
		err := query.Where("id IN ?", batchIDs).
			Find(&batchRes).Error
		if err != nil {
			return nil, err
		}

		res = append(res, batchRes...)
	}

	return res, nil
}

func DeleteResourceByIDs(
	ctx context.Context,
	resourceType constant.APISIXResource,
	ids []string,
) error {
	// 空 IDs 快速返回
	if len(ids) == 0 {
		return nil
	}

	// 单批次直接删除（优化小数据集性能）
	if len(ids) <= constant.DBConditionIDMaxLength {
		query := getCommonDbQuery(ctx, resourceType)
		err := query.Where("id IN (?)", ids).Delete(resourceModelMap[resourceType]).Error
		return err
	}

	// 分批次删除（避免条件叠加）
	for i := 0; i < len(ids); i += constant.DBConditionIDMaxLength {
		end := i + constant.DBConditionIDMaxLength
		if end > len(ids) {
			end = len(ids)
		}
		batchIDs := ids[i:end]
		// 关键点：每个批次创建新查询
		batchQuery := getCommonDbQuery(ctx, resourceType)
		if err := batchQuery.Where("id IN (?)", batchIDs).Delete(resourceModelMap[resourceType]).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetSchemaByIDs 根据 ids 获取 schema
func GetSchemaByIDs(
	ctx context.Context,
	ids []string,
) ([]model.GatewayCustomPluginSchema, error) {
	var res []model.GatewayCustomPluginSchema
	query := getCommonDbQuery(ctx, constant.PluginConfig)
	// 如果 IDs 数量小于等于 DBConditionIDMaxLength，直接查询
	if len(ids) <= constant.DBConditionIDMaxLength {
		err := query.Where("auto_id IN ?", ids).Find(&res).Error
		return res, err
	}

	// 分批处理大量 IDs
	for i := 0; i < len(ids); i += constant.DBConditionIDMaxLength {
		end := i + constant.DBConditionIDMaxLength
		if end > len(ids) {
			end = len(ids)
		}

		batchIDs := ids[i:end]
		var batchRes []model.GatewayCustomPluginSchema

		err := query.Where(
			"auto_id IN ?", batchIDs).Find(&batchRes).Error
		if err != nil {
			return nil, err
		}

		res = append(res, batchRes...)
	}

	return res, nil
}

// QueryResource ... 根据条件查询资源
func QueryResource(
	ctx context.Context,
	resourceType constant.APISIXResource,
	params map[string]interface{},
	name string,
) ([]*model.ResourceCommonModel, error) {
	var res []*model.ResourceCommonModel
	query := getCommonDbQuery(ctx, resourceType).Where(params)
	if name != "" {
		query = query.Where(model.GetResourceNameKey(resourceType)+" LIKE ?", "%"+name+"%")
	}
	err := query.Find(&res).Error
	return res, err
}

// LabelConditionList 标签查询条件列表
func LabelConditionList(
	labelList map[string][]string,
) []gen.Condition {
	var conditions []gen.Condition
	for k, values := range labelList {
		for _, v := range values {
			conditions = append(
				conditions,
				gen.Cond(datatypes.JSONQuery("config").Equals(v, "labels",
					fmt.Sprintf(`"%s"`, k)))...,
			)
		}
	}
	return conditions
}

// DuplicatedResourceName 查询资源名称是否重复
func DuplicatedResourceName(
	ctx context.Context,
	resourceType constant.APISIXResource,
	id string,
	name string,
) bool {
	var res []*model.ResourceCommonModel
	d := getCommonDbQuery(ctx, resourceType).Where(
		getQueryNameParams(ctx, resourceType, []string{name}))
	if id != "" {
		d = d.Not("id = ?", id)
	}
	err := d.Find(&res).Error
	if err != nil {
		logging.Errorf("query resource name: %s error: %s", name, err.Error())
		return false
	}
	if len(res) == 0 {
		return true
	}
	return false
}

func getQueryNameParams(
	ctx context.Context,
	resourceType constant.APISIXResource,
	name []string,
) map[string]interface{} {
	params := map[string]interface{}{}
	params[model.GetResourceNameKey(resourceType)] = name
	return params
}

// BatchCheckNameDuplication 批量校验名称是否重复
func BatchCheckNameDuplication(
	ctx context.Context,
	resourceType constant.APISIXResource,
	names []string,
) (bool, error) {
	var res []*model.ResourceCommonModel
	params := getQueryNameParams(ctx, resourceType, names)
	query := getCommonDbQuery(ctx, resourceType).Where(params)
	err := query.Find(&res).Error
	if err != nil {
		return false, err
	}
	if len(res) > 0 {
		return true, nil
	}
	return false, nil
}

// BatchCreateResources 批量创建资源
func BatchCreateResources(
	ctx context.Context,
	resourceType constant.APISIXResource, resources []*model.ResourceCommonModel,
) error {
	modelSlice, exists := resourceModelSliceMap[resourceType]
	if !exists {
		return fmt.Errorf("unsupported resource type: %v", resourceType)
	}
	newSlice := reflect.MakeSlice(reflect.TypeOf(modelSlice).Elem(), 0, len(resources))
	for _, resource := range resources {
		resourceModel := resource.ToResourceModel(resourceType)
		newSlice = reflect.Append(newSlice, reflect.ValueOf(resourceModel))
	}
	return database.Client().WithContext(ctx).Create(newSlice.Interface()).Error
}

// UpdateResource 更新单个资源
func UpdateResource(
	ctx context.Context,
	resourceType constant.APISIXResource, id string, resource *model.ResourceCommonModel,
) error {
	resourceModel, exists := resourceModelMap[resourceType]
	if !exists {
		return fmt.Errorf("unsupported resource type: %v", resourceType)
	}
	newResourceModel := reflect.New(reflect.TypeOf(resourceModel).Elem()).Interface()
	reflect.ValueOf(newResourceModel).Elem().Set(reflect.ValueOf(resource.ToResourceModel(resourceType)))
	return getCommonDbQuery(ctx, resourceType).Where("id = ?", id).Updates(newResourceModel).Error
}

// GetResourceUpdateStatus 获取资源更新状态
func GetResourceUpdateStatus(
	ctx context.Context,
	resourceType constant.APISIXResource, id string,
) (constant.ResourceStatus, error) {
	resource, err := GetResourceByID(ctx, resourceType, id)
	if err != nil {
		return "", err
	}
	updateStatus := constant.ResourceStatusUpdateDraft
	if resource.Status == constant.ResourceStatusCreateDraft {
		updateStatus = constant.ResourceStatusCreateDraft
	}
	return updateStatus, nil
}

// ParseOrderByExprList 解析排序字段
func ParseOrderByExprList(
	ascFieldMap map[string]field.Expr,
	descFieldMap map[string]field.Expr,
	orderBy string,
) []field.Expr {
	var orderByExprs []field.Expr

	sortConditions := strings.Split(orderBy, ",")
	for _, condition := range sortConditions {
		parts := strings.Split(condition, ":")
		if len(parts) != 2 {
			continue
		}

		fieldName := parts[0]
		direction := strings.ToLower(parts[1])

		switch direction {
		case "asc":
			if _, ok := ascFieldMap[fieldName]; ok {
				orderByExprs = append(orderByExprs, ascFieldMap[fieldName])
			}
		case "desc":
			if _, ok := descFieldMap[fieldName]; ok {
				orderByExprs = append(orderByExprs, descFieldMap[fieldName])
			}
		}
	}

	return orderByExprs
}

// ValidateResource 校验资源
// ValidateResource validates resources based on schema and association checks
// ctx: context containing request information
// resources: map of APISIX resources to their sync data
// allResourceIDMap: map of all valid resource IDs for association validation
// Returns: error if validation fails, nil otherwise
func ValidateResource(
	ctx context.Context,
	resources map[constant.APISIXResource][]*model.GatewaySyncData,
	allResourceIDMap map[string]struct{},
	allPluginSchemaMap map[string]interface{},
) error {
	// Extract gateway information from context
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	// Iterate through each resource type and its associated data
	for resourceType, resource := range resources {
		// Create schema validator for the resource type
		schemaValidator, err := schema.NewAPISIXSchemaValidator(gatewayInfo.GetAPISIXVersionX(),
			"main."+resourceType.String())
		if err != nil {
			return err
		}
		// Validate each resource instance
		for _, r := range resource {
			// Validate resource against schema
			if err = schemaValidator.Validate(json.RawMessage(r.Config)); err != nil {
				logging.Errorf("schema validate failed, err: %v", err)
				return err
			}
			jsonConfigValidator, err := schema.NewAPISIXJsonSchemaValidator(gatewayInfo.GetAPISIXVersionX(),
				resourceType, "main."+string(resourceType), allPluginSchemaMap, constant.DATABASE)
			if err != nil {
				return err
			}
			if err = jsonConfigValidator.Validate(json.RawMessage(r.Config)); err != nil { // 校验json schema
				return fmt.Errorf("resource config:%s validate failed, err: %v",
					r.Config, err)
			}

			// 校验关联数据是否存在
			var resourceAssociateIDInfo dto.ResourceAssociateID
			err = json.Unmarshal(r.Config, &resourceAssociateIDInfo)
			if err != nil {
				return err
			}
			if resourceAssociateIDInfo.ServiceID != "" {
				if _, ok := allResourceIDMap[resourceAssociateIDInfo.ServiceID]; !ok {
					return fmt.Errorf("associated service [id:%s] not found",
						resourceAssociateIDInfo.ServiceID)
				}
			}
			if resourceAssociateIDInfo.UpstreamID != "" {
				if _, ok := allResourceIDMap[resourceAssociateIDInfo.UpstreamID]; !ok {
					return fmt.Errorf("associated upstream [id:%s] not found",
						resourceAssociateIDInfo.UpstreamID)
				}

				if resourceAssociateIDInfo.PluginConfigID != "" {
					if _, ok := allResourceIDMap[resourceAssociateIDInfo.PluginConfigID]; !ok {
						return fmt.Errorf("associated plugin_config [id:%s] not found",
							resourceAssociateIDInfo.PluginConfigID)
					}
				}

				if resourceAssociateIDInfo.GroupID != "" {
					if _, ok := allResourceIDMap[resourceAssociateIDInfo.GroupID]; !ok {
						return fmt.Errorf("associated consumer_group [id:%s] not found",
							resourceAssociateIDInfo.GroupID)
					}
				}
			}
		}

		return nil
	}
	return nil
}
