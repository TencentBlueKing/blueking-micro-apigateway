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
)

// ResourceInfo ...
type ResourceInfo struct {
	ResourceType constant.APISIXResource `json:"resource_type,omitempty"`               // 资源类型
	ResourceID   string                  `json:"resource_id,omitempty"`                 // 资源ID
	Name         string                  `json:"name,omitempty"`                        // 资源名称
	Config       json.RawMessage         `json:"config,omitempty" swaggertype:"object"` // 资源配置
	Status       constant.UploadStatus   `json:"status,omitempty"`                      // 资源导入状态(add/update)
}

// ResourceUploadInfo ...
type ResourceUploadInfo struct {
	Add    map[constant.APISIXResource][]ResourceInfo `json:"add,omitempty"`
	Update map[constant.APISIXResource][]ResourceInfo `json:"update,omitempty"`
}

// ClassifyImportResourceInfo 分类合并导入资源信息
func ClassifyImportResourceInfo(
	importDataList map[constant.APISIXResource][]ResourceInfo,
	existsResourceIdList map[string]struct{},
) (*ResourceUploadInfo, error) {
	resourceIDMap := make(map[constant.APISIXResource][]string) // resourceType:[]id
	for _, impList := range importDataList {
		for _, imp := range impList {
			if idList, ok := resourceIDMap[imp.ResourceType]; ok {
				resourceIDMap[imp.ResourceType] = append(idList, imp.ResourceID)
			} else {
				resourceIDMap[imp.ResourceType] = []string{imp.ResourceID}
			}
		}
	}
	uploadOutput := &ResourceUploadInfo{
		Add:    make(map[constant.APISIXResource][]ResourceInfo),
		Update: make(map[constant.APISIXResource][]ResourceInfo),
	}
	for _, impList := range importDataList {
		for _, imp := range impList {
			imp.Name = gjson.ParseBytes(imp.Config).Get(model.GetResourceNameKey(imp.ResourceType)).String()
			if _, ok := existsResourceIdList[imp.ResourceID]; !ok {
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

// HandleImportResources 处理导入资源
func HandleImportResources(
	ctx context.Context,
	resourcesImport *ResourceUploadInfo,
) (map[constant.APISIXResource][]*model.GatewaySyncData, map[constant.APISIXResource][]*model.GatewaySyncData, error) {
	// 分类聚合
	allResourceIdMap := make(map[string]struct{})
	resourceTypeAddMap, err := handleResources(ctx, resourcesImport.Add, allResourceIdMap)
	if err != nil {
		return nil, nil, err
	}
	resourceTypeUpdateMap, err := handleResources(ctx, resourcesImport.Update, allResourceIdMap)
	if err != nil {
		return nil, nil, err
	}
	err = biz.ValidateResource(ctx, resourceTypeAddMap, allResourceIdMap)
	if err != nil {
		return nil, nil, fmt.Errorf("add resources validate failed, err: %v", err)
	}
	err = biz.ValidateResource(ctx, resourceTypeUpdateMap, allResourceIdMap)
	if err != nil {
		return nil, nil, fmt.Errorf("updated resources validate failed, err: %v", err)
	}
	return resourceTypeAddMap, resourceTypeUpdateMap, nil
}

func handleResources(
	ctx context.Context,
	resourcesImport map[constant.APISIXResource][]ResourceInfo,
	allResourceIdMap map[string]struct{},
) (map[constant.APISIXResource][]*model.GatewaySyncData, error) {
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	for resourceType, resourceInfoList := range resourcesImport {
		allResourceList, err := biz.GetResourceByIDs(ctx, resourceType, []string{})
		if err != nil {
			return nil, fmt.Errorf("get exist resources failed, err: %v", err)
		}
		for _, imp := range resourceInfoList {
			allResourceIdMap[imp.ResourceID] = struct{}{}
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
			allResourceIdMap[resource.ID] = struct{}{}
		}
	}
	return resourceTypeMap, nil
}
