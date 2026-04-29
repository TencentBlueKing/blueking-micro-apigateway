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
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// PreparedImportResources contains the validated sync-data groups ready for the
// upload transaction flow.
type PreparedImportResources struct {
	AddResourceTypeMap    map[constant.APISIXResource][]*model.GatewaySyncData
	UpdateResourceTypeMap map[constant.APISIXResource][]*model.GatewaySyncData
}

// ImportIndexResult contains the discovered DB/index state needed by preview
// and open import flows before classification or validation.
type ImportIndexResult struct {
	ExistingResourceIDs map[string]struct{}
	AllResourceIDs      map[string]struct{}
	AddedSchemaMap      map[string]*model.GatewayCustomPluginSchema
	UpdatedSchemaMap    map[string]*model.GatewayCustomPluginSchema
	AllSchemaMap        map[string]any
	ResourceTypeMap     map[constant.APISIXResource][]*model.GatewaySyncData
}

// ClassifyImportResources splits imported resources into add/update buckets
// while preserving the current import status semantics.
func ClassifyImportResources(
	importDataList map[constant.APISIXResource][]*dto.ImportResourceInfo,
	existingResourceIDs map[string]struct{},
	addPluginSchemaMap map[string]*model.GatewayCustomPluginSchema,
) (*dto.ImportUploadInfo, error) {
	uploadOutput := &dto.ImportUploadInfo{
		Add:    make(map[constant.APISIXResource][]*dto.ImportResourceInfo),
		Update: make(map[constant.APISIXResource][]*dto.ImportResourceInfo),
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
			if _, ok := existingResourceIDs[imp.GetResourceKey()]; !ok {
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

// BuildImportIndex collects existing resource ids, schema state, and raw
// sync-data previews for upload/import requests.
func BuildImportIndex(
	ctx context.Context,
	resourceInfoTypeMap map[constant.APISIXResource][]*dto.ImportResourceInfo,
) (*ImportIndexResult, error) {
	existingResourceIDs := make(map[string]struct{})
	allResourceIDs := make(map[string]struct{})
	var addedSchemaMap map[string]*model.GatewayCustomPluginSchema
	var updatedSchemaMap map[string]*model.GatewayCustomPluginSchema
	var allSchemaMap map[string]any
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)

	for resourceType, resourceInfoList := range resourceInfoTypeMap {
		if resourceType == constant.Schema {
			var err error
			allSchemaMap, addedSchemaMap, updatedSchemaMap, err = buildImportSchemaIndex(
				ctx,
				resourceInfoList,
			)
			if err != nil {
				return nil, err
			}
			continue
		}

		dbResources, err := BatchGetResources(ctx, resourceType, []string{})
		if err != nil {
			return nil, err
		}
		for _, dbResource := range dbResources {
			key := dbResource.GetResourceKey(resourceType)
			existingResourceIDs[key] = struct{}{}
			allResourceIDs[key] = struct{}{}
		}
		for _, resourceInfo := range resourceInfoList {
			if resourceInfo.ResourceID == "" {
				return nil, fmt.Errorf(
					"%s: resource id is empty: %s",
					resourceInfo.ResourceType,
					resourceInfo.Name,
				)
			}
			res := &model.GatewaySyncData{
				Type:   resourceInfo.ResourceType,
				ID:     resourceInfo.ResourceID,
				Config: datatypes.JSON(resourceInfo.Config),
			}
			allResourceIDs[res.GetResourceKey()] = struct{}{}
			resourceTypeMap[resourceInfo.ResourceType] = append(
				resourceTypeMap[resourceInfo.ResourceType],
				res,
			)
		}
	}
	return &ImportIndexResult{
		ExistingResourceIDs: existingResourceIDs,
		AllResourceIDs:      allResourceIDs,
		AddedSchemaMap:      addedSchemaMap,
		UpdatedSchemaMap:    updatedSchemaMap,
		AllSchemaMap:        allSchemaMap,
		ResourceTypeMap:     resourceTypeMap,
	}, nil
}

// BuildImportUploadSchemaModels converts classified schema resources into DB
// models for the web import path and mutates allSchemaMap for validation.
func BuildImportUploadSchemaModels(
	ctx context.Context,
	resources map[constant.APISIXResource][]*dto.ImportResourceInfo,
	allSchemaMap map[string]any,
) (map[string]*model.GatewayCustomPluginSchema, error) {
	schemaMap := make(map[string]*model.GatewayCustomPluginSchema)
	for _, resourceList := range resources {
		for _, resource := range resourceList {
			if resource.ResourceType != constant.Schema {
				continue
			}
			schemaModel := buildImportSchemaModel(ctx, resource, true)
			schemaMap[resource.Name] = schemaModel
			if err := mergeSchemaIntoMap(allSchemaMap, schemaModel); err != nil {
				return nil, err
			}
		}
	}
	return schemaMap, nil
}

// PrepareImportUpload prepares add/update sync-data groups and validates them
// before the upload transaction flow begins.
func PrepareImportUpload(
	ctx context.Context,
	resourcesImport *dto.ImportUploadInfo,
	allSchemaMap map[string]any,
	ignoreFields map[constant.APISIXResource][]string,
) (*PreparedImportResources, error) {
	validationInput, err := prepareImportValidationInput(ctx, resourcesImport, ignoreFields)
	if err != nil {
		return nil, err
	}
	err = ValidateImportedResources(ctx, validationInput.Add, validationInput.AllResourceIDs, allSchemaMap)
	if err != nil {
		return nil, fmt.Errorf("add resources validate failed, err: %w", err)
	}
	err = ValidateImportedResources(ctx, validationInput.Update, validationInput.AllResourceIDs, allSchemaMap)
	if err != nil {
		return nil, fmt.Errorf("updated resources validate failed, err: %w", err)
	}
	return &PreparedImportResources{
		AddResourceTypeMap:    validationInput.Add,
		UpdateResourceTypeMap: validationInput.Update,
	}, nil
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

// loadExistingImportResources fetches all stored resources of the given type
// for the current gateway and returns them keyed by resource key together with
// the discovered resource key set.
func loadExistingImportResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
) (map[string]model.ResourceCommonModel, map[string]struct{}, error) {
	allResourceList, err := GetResourceByIDs(ctx, resourceType, []string{})
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
	imp *dto.ImportResourceInfo,
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
	resourcesImport map[constant.APISIXResource][]*dto.ImportResourceInfo,
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
			if imp.ResourceID == "" {
				return nil, nil, fmt.Errorf("%s: resource id is empty: %s", resourceType, imp.Name)
			}
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
	resourcesImport *dto.ImportUploadInfo,
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

func buildImportSchemaIndex(
	ctx context.Context,
	schemaInfoList []*dto.ImportResourceInfo,
) (
	allSchemaMap map[string]any,
	addedSchemaMap map[string]*model.GatewayCustomPluginSchema,
	updatedSchemaMap map[string]*model.GatewayCustomPluginSchema,
	err error,
) {
	existsPluginSchemaMap, err := GetCustomizePluginSchemaMap(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	addedSchemaMap = make(map[string]*model.GatewayCustomPluginSchema)
	updatedSchemaMap = make(map[string]*model.GatewayCustomPluginSchema)
	for _, schemaInfo := range schemaInfoList {
		schemaModel := buildImportSchemaModel(ctx, schemaInfo, false)
		if _, ok := existsPluginSchemaMap[schemaInfo.Name]; !ok {
			addedSchemaMap[schemaInfo.Name] = schemaModel
		} else {
			updatedSchemaMap[schemaInfo.Name] = schemaModel
		}
		if err := mergeSchemaIntoMap(existsPluginSchemaMap, schemaModel); err != nil {
			return nil, nil, nil, err
		}
	}
	return existsPluginSchemaMap, addedSchemaMap, updatedSchemaMap, nil
}

func buildImportSchemaModel(
	ctx context.Context,
	schemaInfo *dto.ImportResourceInfo,
	setUpdater bool,
) *model.GatewayCustomPluginSchema {
	baseModel := model.BaseModel{
		Creator: ginx.GetUserIDFromContext(ctx),
	}
	if setUpdater {
		baseModel.Updater = ginx.GetUserIDFromContext(ctx)
	}
	return &model.GatewayCustomPluginSchema{
		GatewayID:     ginx.GetGatewayInfoFromContext(ctx).ID,
		Name:          schemaInfo.Name,
		Schema:        datatypes.JSON(gjson.GetBytes(schemaInfo.Config, "schema").String()),
		Example:       datatypes.JSON(gjson.GetBytes(schemaInfo.Config, "example").String()),
		BaseModel:     baseModel,
		OperationType: constant.OperationImport,
	}
}

func mergeSchemaIntoMap(allSchemaMap map[string]any, schemaModel *model.GatewayCustomPluginSchema) error {
	schemaRaw, err := json.Marshal(schemaModel.Schema)
	if err != nil {
		return fmt.Errorf("marshal schema failed: %w", err)
	}
	var schemaMap map[string]any
	_ = json.Unmarshal(schemaRaw, &schemaMap)
	allSchemaMap[schemaModel.Name] = schemaMap
	return nil
}
