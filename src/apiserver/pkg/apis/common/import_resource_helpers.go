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
