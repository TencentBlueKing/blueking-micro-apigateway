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

func applyImportIgnoreFields(
	importedConfig json.RawMessage,
	existingConfig datatypes.JSON,
	ignoreFields []string,
) (json.RawMessage, error) {
	merged := append(json.RawMessage(nil), importedConfig...)
	for _, field := range ignoreFields {
		result := gjson.GetBytes(existingConfig, field)
		if !result.Exists() {
			continue
		}
		var err error
		merged, err = sjson.SetBytes(merged, field, json.RawMessage(result.Raw))
		if err != nil {
			return nil, err
		}
	}
	return merged, nil
}

// loadExistingImportResources fetches all stored resources of the given type for
// the current gateway and returns them keyed by GetResourceKey(resourceType, id).
//
// mutates allResourceIDs: every loaded resource key is appended to the passed-in
// set so that the import orchestration can build a global id set across all
// resource types without re-querying the DB.
func loadExistingImportResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
	allResourceIDs map[string]struct{},
) (map[string]model.ResourceCommonModel, error) {
	allResourceList, err := biz.GetResourceByIDs(ctx, resourceType, []string{})
	if err != nil {
		return nil, fmt.Errorf("get exist resources failed, err: %w", err)
	}

	allResourceMap := make(map[string]model.ResourceCommonModel, len(allResourceList))
	for _, resource := range allResourceList {
		key := resource.GetResourceKey(resourceType)
		allResourceMap[key] = resource
		allResourceIDs[key] = struct{}{}
	}
	return allResourceMap, nil
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
//
// mutates allResourceIDs: appends every encountered resource key from both
// stored resources and imported resources while building the returned map.
func prepareImportResources(
	ctx context.Context,
	resourcesImport map[constant.APISIXResource][]*ResourceInfo,
	allResourceIDs map[string]struct{},
	ignoreFields map[constant.APISIXResource][]string,
) (map[constant.APISIXResource][]*model.GatewaySyncData, error) {
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	for resourceType, resourceInfoList := range resourcesImport {
		if resourceType == constant.Schema {
			continue
		}

		existingMap, err := loadExistingImportResources(ctx, resourceType, allResourceIDs)
		if err != nil {
			return nil, err
		}

		for _, imp := range resourceInfoList {
			if imp.ResourceID == "" {
				return nil, fmt.Errorf("%s: resource id is empty: %s", resourceType, imp.Name)
			}
			if oldResource, ok := existingMap[imp.GetResourceKey()]; ok &&
				len(ignoreFields[resourceType]) > 0 {
				imp.Config, err = applyImportIgnoreFields(
					imp.Config,
					oldResource.Config,
					ignoreFields[resourceType],
				)
				if err != nil {
					return nil, fmt.Errorf("set config failed, err: %w", err)
				}
			}

			allResourceIDs[imp.GetResourceKey()] = struct{}{}
			resourceTypeMap[resourceType] = append(
				resourceTypeMap[resourceType],
				buildImportSyncData(ctx, resourceType, imp),
			)
		}
	}
	return resourceTypeMap, nil
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
	allResourceIDs := make(map[string]struct{})
	addMap, err := prepareImportResources(ctx, resourcesImport.Add, allResourceIDs, ignoreFields)
	if err != nil {
		return nil, err
	}
	updateMap, err := prepareImportResources(ctx, resourcesImport.Update, allResourceIDs, ignoreFields)
	if err != nil {
		return nil, err
	}
	return &importValidationInput{
		Add:            addMap,
		Update:         updateMap,
		AllResourceIDs: allResourceIDs,
	}, nil
}
