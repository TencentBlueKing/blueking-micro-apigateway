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

// Package common ...
package common

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// ResourceInfo ...
type ResourceInfo struct {
	ResourceType constant.APISIXResource `json:"resource_type,omitempty"`               // 资源类型
	ResourceID   string                  `json:"resource_id,omitempty"`                 // 资源ID
	Name         string                  `json:"name,omitempty"`                        // 资源名称
	Config       json.RawMessage         `json:"config,omitempty" swaggertype:"object"` // 资源配置
	Status       constant.UploadStatus   `json:"status,omitempty"`                      // 资源导入状态(add/update)
}

func (r ResourceInfo) GetResourceKey() string {
	// 插件元素数需要特殊处理,因为插件元素数没有真正id
	if r.ResourceType == constant.PluginMetadata {
		return fmt.Sprintf(constant.ResourceKeyFormat, r.ResourceType, r.Name)
	}
	return fmt.Sprintf(constant.ResourceKeyFormat, r.ResourceType, r.ResourceID)
}

// ResourceUploadInfo ...
type ResourceUploadInfo struct {
	Add    map[constant.APISIXResource][]ResourceInfo `json:"add,omitempty"`
	Update map[constant.APISIXResource][]ResourceInfo `json:"update,omitempty"`
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
	AllSchemaMap         map[string]interface{}
	ResourceTypeMap      map[constant.APISIXResource][]*model.GatewaySyncData
}

// ClassifyImportResourceInfo 分类合并导入资源信息
func ClassifyImportResourceInfo(
	importDataList map[constant.APISIXResource][]ResourceInfo,
	existsResourceIdList map[string]struct{},
	addPluginSchemaMap map[string]*model.GatewayCustomPluginSchema,
) (*ResourceUploadInfo, error) {
	uploadOutput := &ResourceUploadInfo{
		Add:    make(map[constant.APISIXResource][]ResourceInfo),
		Update: make(map[constant.APISIXResource][]ResourceInfo),
	}
	for resourceType, impList := range importDataList {
		for _, imp := range impList {
			if resourceType == constant.Schema {
				if _, ok := addPluginSchemaMap[imp.Name]; ok {
					imp.Status = constant.UploadStatusAdd
					uploadOutput.Add[constant.Schema] = append(uploadOutput.Add[constant.Schema], imp)
				} else {
					imp.Status = constant.UploadStatusUpdate
					uploadOutput.Update[constant.Schema] = append(uploadOutput.Update[constant.Schema], imp)
				}
				continue
			}
			imp.Name = gjson.ParseBytes(imp.Config).Get(model.GetResourceNameKey(imp.ResourceType)).String()
			if _, ok := existsResourceIdList[imp.GetResourceKey()]; !ok {
				imp.Status = constant.UploadStatusAdd
				uploadOutput.Add[imp.ResourceType] = append(uploadOutput.Add[imp.ResourceType], imp)
			} else {
				imp.Status = constant.UploadStatusUpdate
				uploadOutput.Update[imp.ResourceType] = append(uploadOutput.Update[imp.ResourceType], imp)
			}
		}
	}
	return uploadOutput, nil
}

// HandleUploadResources 处理导入资源
func HandleUploadResources(
	ctx context.Context,
	resourcesImport *ResourceUploadInfo,
	allSchemaMap map[string]interface{},
) (*HandlerResourceResult, error) {
	// 分类聚合
	allResourceIdMap := make(map[string]struct{})
	resourceTypeAddMap, err := handleResources(ctx, resourcesImport.Add, allResourceIdMap)
	if err != nil {
		return nil, err
	}
	resourceTypeUpdateMap, err := handleResources(ctx, resourcesImport.Update, allResourceIdMap)
	if err != nil {
		return nil, err
	}
	err = biz.ValidateResource(ctx, resourceTypeAddMap, allResourceIdMap, allSchemaMap)
	if err != nil {
		return nil, fmt.Errorf("add resources validate failed, err: %v", err)
	}
	err = biz.ValidateResource(ctx, resourceTypeUpdateMap, allResourceIdMap, allSchemaMap)
	if err != nil {
		return nil, fmt.Errorf("updated resources validate failed, err: %v", err)
	}
	return &HandlerResourceResult{
		AddResourceTypeMap:    resourceTypeAddMap,
		UpdateResourceTypeMap: resourceTypeUpdateMap,
	}, nil
}

// HandlerCustomerPluginSchemaImport is a function that handles the import of customer plugin schemas.
func HandlerCustomerPluginSchemaImport(ctx context.Context, schemaInfoList []ResourceInfo) (
	allSchemaMap map[string]interface{}, addedSchemaMap,
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
		schemaRaw, _ := json.Marshal(schemaModel.Schema)
		var schemaMap map[string]interface{}
		_ = json.Unmarshal(schemaRaw, &schemaMap)
		existsPluginSchemaMap[schemaInfo.Name] = schemaMap
	}
	return existsPluginSchemaMap, addedSchemaMap, updatedSchemaMap, nil
}

// HandlerResourceIndexMap is a function that handles the indexing of resources.
func HandlerResourceIndexMap(ctx context.Context, resourceInfoTypeMap map[constant.APISIXResource][]ResourceInfo) (
	*HandlerResourceIndexResult, error,
) {
	existsResourceIdList := make(map[string]struct{})
	allResourceIdList := make(map[string]struct{})
	var addedSchemaMap map[string]*model.GatewayCustomPluginSchema
	var updatedSchemaMap map[string]*model.GatewayCustomPluginSchema
	var allSchemaMap map[string]interface{}
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
			// 生成ID,如果为空
			if resourceInfo.ResourceID == "" {
				resourceInfo.ResourceID = idx.GenResourceID(resourceType)
			}
			res := &model.GatewaySyncData{
				Type:   resourceInfo.ResourceType,
				ID:     resourceInfo.ResourceID,
				Config: datatypes.JSON(resourceInfo.Config),
			}
			allResourceIdList[res.GetResourceKey()] = struct{}{}
			if _, ok := resourceTypeMap[resourceInfo.ResourceType]; ok {
				resourceTypeMap[resourceInfo.ResourceType] = append(resourceTypeMap[resourceInfo.ResourceType], res)
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

func handleResources(
	ctx context.Context,
	resourcesImport map[constant.APISIXResource][]ResourceInfo,
	allResourceIdMap map[string]struct{},
) (map[constant.APISIXResource][]*model.GatewaySyncData, error) {
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	for resourceType, resourceInfoList := range resourcesImport {
		if resourceType == constant.Schema {
			continue
		}
		allResourceList, err := biz.GetResourceByIDs(ctx, resourceType, []string{})
		if err != nil {
			return nil, fmt.Errorf("get exist resources failed, err: %v", err)
		}
		for _, imp := range resourceInfoList {
			// 生成ID,如果为空
			if imp.ResourceID == "" {
				imp.ResourceID = idx.GenResourceID(resourceType)
			}
			allResourceIdMap[imp.GetResourceKey()] = struct{}{}
			resourceImp := &model.GatewaySyncData{
				Type:      resourceType,
				ID:        imp.ResourceID,
				Config:    datatypes.JSON(imp.Config),
				GatewayID: ginx.GetGatewayInfoFromContext(ctx).ID,
			}
			if _, ok := resourceTypeMap[imp.ResourceType]; !ok {
				resourceTypeMap[resourceType] = []*model.GatewaySyncData{resourceImp}
				continue
			}
			resourceTypeMap[resourceType] = append(resourceTypeMap[resourceType], resourceImp)
		}
		for _, resource := range allResourceList {
			allResourceIdMap[resource.GetResourceKey(resourceType)] = struct{}{}
		}
	}
	return resourceTypeMap, nil
}
