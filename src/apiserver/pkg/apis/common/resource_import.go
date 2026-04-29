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

// Package common ...
package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ResourceInfo ...
type ResourceInfo struct {
	ResourceType constant.APISIXResource `json:"resource_type,omitempty"`               // 资源类型
	ResourceID   string                  `json:"resource_id,omitempty"`                 // 资源 ID
	Name         string                  `json:"name,omitempty"`                        // 资源名称
	Config       json.RawMessage         `json:"config,omitempty" swaggertype:"object"` // 资源配置
	Status       constant.UploadStatus   `json:"status,omitempty"`                      // 资源导入状态 (add/update)
}

// GetResourceKey 获取资源 key
func (r ResourceInfo) GetResourceKey() string {
	// 插件元数据需要特殊处理，因为插件元素数没有真正 id
	if r.ResourceType == constant.PluginMetadata {
		return fmt.Sprintf(constant.ResourceKeyFormat, r.ResourceType, r.Name)
	}
	return fmt.Sprintf(constant.ResourceKeyFormat, r.ResourceType, r.ResourceID)
}

// ResourceUploadInfo ...
type ResourceUploadInfo struct {
	Add    map[constant.APISIXResource][]*ResourceInfo `json:"add,omitempty"`
	Update map[constant.APISIXResource][]*ResourceInfo `json:"update,omitempty"`
}

// HandlerResourceResult ...
type HandlerResourceResult struct {
	AddResourceTypeMap    map[constant.APISIXResource][]*model.GatewaySyncData
	UpdateResourceTypeMap map[constant.APISIXResource][]*model.GatewaySyncData
}

// HandlerResourceIndexResult ...
type HandlerResourceIndexResult struct {
	ExistsResourceIdList map[string]struct{}
	AllResourceIdList    map[string]struct{}
	AddedSchemaMap       map[string]*model.GatewayCustomPluginSchema
	UpdatedSchemaMap     map[string]*model.GatewayCustomPluginSchema
	AllSchemaMap         map[string]any
	ResourceTypeMap      map[constant.APISIXResource][]*model.GatewaySyncData
}

// ClassifyImportResourceInfo 分类合并导入资源信息
func ClassifyImportResourceInfo(
	importDataList map[constant.APISIXResource][]*ResourceInfo,
	existsResourceIdList map[string]struct{},
	addPluginSchemaMap map[string]*model.GatewayCustomPluginSchema,
) (*ResourceUploadInfo, error) {
	uploadOutput := &ResourceUploadInfo{
		Add:    make(map[constant.APISIXResource][]*ResourceInfo),
		Update: make(map[constant.APISIXResource][]*ResourceInfo),
	}
	for resourceType, impList := range importDataList {
		for _, imp := range impList {
			if resourceType == constant.Schema {
				if _, ok := addPluginSchemaMap[imp.Name]; ok {
					imp.Status = constant.UploadStatusAdd
					uploadOutput.Add[constant.Schema] = append(
						uploadOutput.Add[constant.Schema],
						imp,
					)
				} else {
					imp.Status = constant.UploadStatusUpdate
					uploadOutput.Update[constant.Schema] = append(
						uploadOutput.Update[constant.Schema],
						imp,
					)
				}
				continue
			}
			imp.Name = gjson.ParseBytes(imp.Config).Get(model.GetResourceNameKey(imp.ResourceType)).String()
			if _, ok := existsResourceIdList[imp.GetResourceKey()]; !ok {
				imp.Status = constant.UploadStatusAdd
				uploadOutput.Add[imp.ResourceType] = append(uploadOutput.Add[imp.ResourceType], imp)
			} else {
				imp.Status = constant.UploadStatusUpdate
				uploadOutput.Update[imp.ResourceType] = append(
					uploadOutput.Update[imp.ResourceType],
					imp,
				)
			}
		}
	}
	return uploadOutput, nil
}

// applyImportIgnoreFields overlays the configured ignore-field paths from the
// existing stored config onto the imported config and returns the merged copy.
func applyImportIgnoreFields(
	importedConfig json.RawMessage,
	existingConfig datatypes.JSON,
	ignoreFields []string,
) (json.RawMessage, error) {
	merged := append(json.RawMessage(nil), importedConfig...)
	for _, field := range ignoreFields {
		value := gjson.GetBytes(existingConfig, field)
		if !value.Exists() {
			continue
		}
		var err error
		merged, err = sjson.SetBytes(merged, field, json.RawMessage(value.Raw))
		if err != nil {
			return nil, fmt.Errorf("set config failed, err: %w", err)
		}
	}
	return merged, nil
}

// loadExistingImportResources fetches all stored resources of the given type for
// the current gateway and returns them keyed by GetResourceKey(resourceType, id)
// together with the discovered resource key set.
func loadExistingImportResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
) (map[string]model.ResourceCommonModel, map[string]struct{}, error) {
	allResourceList, err := biz.GetResourceByIDs(ctx, resourceType, []string{})
	if err != nil {
		return nil, nil, fmt.Errorf("get exist resources failed, err: %w", err)
	}

	allResourceMap := make(map[string]model.ResourceCommonModel, len(allResourceList))
	existingResourceIDs := make(map[string]struct{}, len(allResourceList))
	for _, resource := range allResourceList {
		key := resource.GetResourceKey(resourceType)
		allResourceMap[key] = resource
		existingResourceIDs[key] = struct{}{}
	}
	return allResourceMap, existingResourceIDs, nil
}

func buildImportSyncData(
	ctx context.Context,
	resourceType constant.APISIXResource,
	imp *ResourceInfo,
) *model.GatewaySyncData {
	return &model.GatewaySyncData{
		Type:      resourceType,
		ID:        imp.ResourceID,
		Config:    datatypes.JSON(imp.Config),
		GatewayID: ginx.GetGatewayInfoFromContext(ctx).ID,
	}
}

// prepareImportResources transforms import payloads into sync-data groups for
// later validation and upload. It keeps the import-specific argument order
// because this orchestration helper accepts multiple resource maps together.
func prepareImportResources(
	ctx context.Context,
	resourcesImport map[constant.APISIXResource][]*ResourceInfo,
	ignoreFields map[constant.APISIXResource][]string,
) (map[constant.APISIXResource][]*model.GatewaySyncData, map[string]struct{}, error) {
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	allResourceIDs := make(map[string]struct{})
	for resourceType, resourceInfoList := range resourcesImport {
		if resourceType == constant.Schema {
			continue
		}

		existingMap, existingIDs, err := loadExistingImportResources(ctx, resourceType)
		if err != nil {
			return nil, nil, err
		}
		for key := range existingIDs {
			allResourceIDs[key] = struct{}{}
		}

		for _, imp := range resourceInfoList {
			// 如果 id 为空，直接报错
			if imp.ResourceID == "" {
				return nil, nil, fmt.Errorf("%s: resource id is empty: %s", resourceType, imp.Name)
			}
			// 如果已经存在，则需要判断是否有跳过规则
			if oldResource, ok := existingMap[imp.GetResourceKey()]; ok &&
				len(ignoreFields[resourceType]) > 0 {
				imp.Config, err = applyImportIgnoreFields(
					imp.Config,
					oldResource.Config,
					ignoreFields[resourceType],
				)
				if err != nil {
					return nil, nil, fmt.Errorf("set config failed, err: %w", err)
				}
			}

			allResourceIDs[imp.GetResourceKey()] = struct{}{}
			resourceTypeMap[resourceType] = append(
				resourceTypeMap[resourceType],
				buildImportSyncData(ctx, resourceType, imp),
			)
		}
	}
	return resourceTypeMap, allResourceIDs, nil
}

type importValidationInput struct {
	Add            map[constant.APISIXResource][]*model.GatewaySyncData
	Update         map[constant.APISIXResource][]*model.GatewaySyncData
	AllResourceIDs map[string]struct{}
}

// prepareImportValidationInput prepares add/update sync-data maps and builds a
// shared resource-id set for the validation phase across both import branches.
func prepareImportValidationInput(
	ctx context.Context,
	resourcesImport *ResourceUploadInfo,
	ignoreFields map[constant.APISIXResource][]string,
) (*importValidationInput, error) {
	addMap, addIDs, err := prepareImportResources(ctx, resourcesImport.Add, ignoreFields)
	if err != nil {
		return nil, err
	}
	updateMap, updateIDs, err := prepareImportResources(ctx, resourcesImport.Update, ignoreFields)
	if err != nil {
		return nil, err
	}
	allResourceIDs := make(map[string]struct{}, len(addIDs)+len(updateIDs))
	for key := range addIDs {
		allResourceIDs[key] = struct{}{}
	}
	for key := range updateIDs {
		allResourceIDs[key] = struct{}{}
	}
	return &importValidationInput{
		Add:            addMap,
		Update:         updateMap,
		AllResourceIDs: allResourceIDs,
	}, nil
}

// HandleUploadResources 处理导入资源
func HandleUploadResources(
	ctx context.Context,
	resourcesImport *ResourceUploadInfo,
	allSchemaMap map[string]any,
	ignoreFields map[constant.APISIXResource][]string,
) (*HandlerResourceResult, error) {
	validationInput, err := prepareImportValidationInput(ctx, resourcesImport, ignoreFields)
	if err != nil {
		return nil, err
	}
	err = biz.ValidateResource(ctx, validationInput.Add, validationInput.AllResourceIDs, allSchemaMap)
	if err != nil {
		return nil, fmt.Errorf("add resources validate failed, err: %w", err)
	}
	err = biz.ValidateResource(ctx, validationInput.Update, validationInput.AllResourceIDs, allSchemaMap)
	if err != nil {
		return nil, fmt.Errorf("updated resources validate failed, err: %w", err)
	}
	return &HandlerResourceResult{
		AddResourceTypeMap:    validationInput.Add,
		UpdateResourceTypeMap: validationInput.Update,
	}, nil
}

// HandlerCustomerPluginSchemaImport is a function that handles the import of customer plugin schemas.
func HandlerCustomerPluginSchemaImport(ctx context.Context, schemaInfoList []*ResourceInfo) (
	allSchemaMap map[string]any, addedSchemaMap,
	updatedSchemaMap map[string]*model.GatewayCustomPluginSchema, err error,
) {
	// Get the existing plugin schema map from the business logic layer
	existsPluginSchemaMap, err := biz.GetCustomizePluginSchemaMap(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	// Initialize a map to store the newly added schemas
	addedSchemaMap = make(map[string]*model.GatewayCustomPluginSchema)
	updatedSchemaMap = make(map[string]*model.GatewayCustomPluginSchema)
	// Iterate through each resource info in the schema info list
	for _, schemaInfo := range schemaInfoList {
		schemaModel := &model.GatewayCustomPluginSchema{
			GatewayID: ginx.GetGatewayInfoFromContext(ctx).ID,
			Name:      schemaInfo.Name,
			Schema:    datatypes.JSON(gjson.GetBytes(schemaInfo.Config, "schema").String()),
			Example:   datatypes.JSON(gjson.GetBytes(schemaInfo.Config, "example").String()),
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserIDFromContext(ctx),
			},
			OperationType: constant.OperationImport,
		}
		// Check if the schema already exists in the plugin schema map
		if _, ok := existsPluginSchemaMap[schemaInfo.Name]; !ok {
			addedSchemaMap[schemaInfo.Name] = schemaModel
		} else {
			updatedSchemaMap[schemaInfo.Name] = schemaModel
		}
		schemaRaw, err := json.Marshal(schemaModel.Schema)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("marshal schema failed: %w", err)
		}
		var schemaMap map[string]any
		_ = json.Unmarshal(schemaRaw, &schemaMap)
		existsPluginSchemaMap[schemaInfo.Name] = schemaMap
	}
	return existsPluginSchemaMap, addedSchemaMap, updatedSchemaMap, nil
}

// HandlerResourceIndexMap is a function that handles the indexing of resources.
func HandlerResourceIndexMap(ctx context.Context, resourceInfoTypeMap map[constant.APISIXResource][]*ResourceInfo) (
	*HandlerResourceIndexResult, error,
) {
	existsResourceIdList := make(map[string]struct{})
	allResourceIdList := make(map[string]struct{})
	var addedSchemaMap map[string]*model.GatewayCustomPluginSchema
	var updatedSchemaMap map[string]*model.GatewayCustomPluginSchema
	var allSchemaMap map[string]any
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	var err error
	for resourceType, resourceInfoList := range resourceInfoTypeMap {
		// 自定义插件特殊处理
		if resourceType == constant.Schema {
			allSchemaMap, addedSchemaMap, updatedSchemaMap, err = HandlerCustomerPluginSchemaImport(
				ctx, resourceInfoList)
			if err != nil {
				return nil, err
			}
			continue
		}
		dbResources, err := biz.BatchGetResources(ctx, resourceType, []string{})
		if err != nil {
			return nil, err
		}
		for _, dbResource := range dbResources {
			existsResourceIdList[dbResource.GetResourceKey(resourceType)] = struct{}{}
			allResourceIdList[dbResource.GetResourceKey(resourceType)] = struct{}{}
		}
		for _, resourceInfo := range resourceInfoList {
			// id 如果为空，直接报错
			if resourceInfo.ResourceID == "" {
				return nil, fmt.Errorf("%s: resource id is empty: %s",
					resourceInfo.ResourceType,
					resourceInfo.Name)
			}
			res := &model.GatewaySyncData{
				Type:   resourceInfo.ResourceType,
				ID:     resourceInfo.ResourceID,
				Config: datatypes.JSON(resourceInfo.Config),
			}
			allResourceIdList[res.GetResourceKey()] = struct{}{}
			if _, ok := resourceTypeMap[resourceInfo.ResourceType]; ok {
				resourceTypeMap[resourceInfo.ResourceType] = append(
					resourceTypeMap[resourceInfo.ResourceType],
					res,
				)
				continue
			}
			resourceTypeMap[resourceInfo.ResourceType] = []*model.GatewaySyncData{res}
		}
	}
	return &HandlerResourceIndexResult{
		ExistsResourceIdList: existsResourceIdList,
		AllResourceIdList:    allResourceIdList,
		AddedSchemaMap:       addedSchemaMap,
		UpdatedSchemaMap:     updatedSchemaMap,
		AllSchemaMap:         allSchemaMap,
		ResourceTypeMap:      resourceTypeMap,
	}, nil
}
