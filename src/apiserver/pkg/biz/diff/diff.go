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

// Package diff contains snapshot-vs-edit-area diff helpers.
package diff

import (
	"context"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

var defaultDiffResourceStatuses = []constant.ResourceStatus{
	constant.ResourceStatusCreateDraft,
	constant.ResourceStatusDeleteDraft,
	constant.ResourceStatusUpdateDraft,
}

// DiffResources 对比资源数据
func DiffResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
	idList []string,
	name string,
	resourceStatus []constant.ResourceStatus,
	isDiffAll bool,
) ([]dto.ResourceChangeInfo, error) {
	diffResourceTypeMap := initDiffResourceTypeMap(resourceType, idList)
	result := make([]dto.ResourceChangeInfo, 0)

	for _, currentType := range constant.ResourceTypeList {
		queryParams, resourceName, shouldSkip := buildDiffQueryParams(
			ctx,
			currentType,
			resourceType,
			diffResourceTypeMap[currentType],
			name,
			resourceStatus,
			isDiffAll,
		)
		if shouldSkip {
			continue
		}

		resources, err := resourcebiz.QueryResource(ctx, currentType, queryParams, resourceName)
		if err != nil {
			return nil, err
		}
		if len(resources) == 0 {
			continue
		}

		diffResult, err := buildResourceTypeDiffResult(
			ctx,
			currentType,
			resources,
			diffResourceTypeMap,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, diffResult)
	}

	return result, nil
}

func initDiffResourceTypeMap(
	resourceType constant.APISIXResource,
	idList []string,
) map[constant.APISIXResource][]string {
	diffResourceTypeMap := make(map[constant.APISIXResource][]string, len(constant.ResourceTypeList))
	for _, resourceTypeItem := range constant.ResourceTypeList {
		diffResourceTypeMap[resourceTypeItem] = []string{}
	}
	if resourceType != "" && len(idList) != 0 {
		diffResourceTypeMap[resourceType] = idList
	}
	return diffResourceTypeMap
}

func buildDiffQueryParams(
	ctx context.Context,
	currentType constant.APISIXResource,
	requestedType constant.APISIXResource,
	resourceIDList []string,
	name string,
	resourceStatus []constant.ResourceStatus,
	isDiffAll bool,
) (map[string]any, string, bool) {
	queryParams := map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
		"status":     defaultDiffResourceStatuses,
	}

	if len(resourceIDList) != 0 && requestedType != "" {
		queryParams["id"] = resourceIDList
	}

	resourceName := ""
	if requestedType == currentType {
		resourceName = name
		if len(resourceStatus) > 0 {
			queryParams["status"] = resourceStatus
		}
	}

	if requestedType != "" && len(resourceIDList) == 0 && !isDiffAll {
		return nil, "", true
	}

	return queryParams, resourceName, false
}

func buildResourceTypeDiffResult(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resources []*model.ResourceCommonModel,
	diffResourceTypeMap map[constant.APISIXResource][]string,
) (dto.ResourceChangeInfo, error) {
	diffResult := dto.ResourceChangeInfo{
		ResourceType: resourceType,
		AddedCount:   0,
		DeletedCount: 0,
		UpdateCount:  0,
		ChangeDetail: make([]dto.ResourceChangeDetail, 0, len(resources)),
	}

	for _, resourceInfo := range resources {
		changeDetail, err := buildResourceChangeDetail(ctx, resourceType, resourceInfo)
		if err != nil {
			return dto.ResourceChangeInfo{}, err
		}

		applyResourceChangeStats(&diffResult, resourceInfo.Status)
		diffResult.ChangeDetail = append(diffResult.ChangeDetail, changeDetail)
		collectRelatedDiffResourceIDs(diffResourceTypeMap, resourceInfo)
	}

	return diffResult, nil
}

func buildResourceChangeDetail(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceInfo *model.ResourceCommonModel,
) (dto.ResourceChangeDetail, error) {
	statusOp := status.NewResourceStatusOp(*resourceInfo)
	afterStatus, err := statusOp.NextStatus(ctx, constant.OperationTypePublish)
	if err != nil {
		return dto.ResourceChangeDetail{}, err
	}

	changeDetail := dto.ResourceChangeDetail{
		ResourceID:   resourceInfo.ID,
		BeforeStatus: resourceInfo.Status,
		Name:         resourceInfo.GetName(resourceType),
		UpdatedAt:    resourceInfo.UpdatedAt.Unix(),
		AfterStatus:  afterStatus,
	}

	switch resourceInfo.Status {
	case constant.ResourceStatusCreateDraft:
		changeDetail.PublishFrom = constant.OperationTypeCreate
	case constant.ResourceStatusDeleteDraft:
		changeDetail.PublishFrom = constant.OperationTypeDelete
	case constant.ResourceStatusUpdateDraft:
		changeDetail.PublishFrom = constant.OperationTypeUpdate
	}

	return changeDetail, nil
}

func applyResourceChangeStats(
	diffResult *dto.ResourceChangeInfo,
	resourceStatus constant.ResourceStatus,
) {
	switch resourceStatus {
	case constant.ResourceStatusCreateDraft:
		diffResult.AddedCount++
	case constant.ResourceStatusDeleteDraft:
		diffResult.DeletedCount++
	case constant.ResourceStatusUpdateDraft:
		diffResult.UpdateCount++
	}
}

func collectRelatedDiffResourceIDs(
	diffResourceTypeMap map[constant.APISIXResource][]string,
	resourceInfo *model.ResourceCommonModel,
) {
	if serviceID := resourceInfo.GetServiceID(); serviceID != "" {
		diffResourceTypeMap[constant.Service] = append(
			diffResourceTypeMap[constant.Service],
			serviceID,
		)
	}
	if upstreamID := resourceInfo.GetUpstreamID(); upstreamID != "" {
		diffResourceTypeMap[constant.Upstream] = append(
			diffResourceTypeMap[constant.Upstream],
			upstreamID,
		)
	}
	if pluginConfigID := resourceInfo.GetPluginConfigID(); pluginConfigID != "" {
		diffResourceTypeMap[constant.PluginConfig] = append(
			diffResourceTypeMap[constant.PluginConfig],
			pluginConfigID,
		)
	}
	if consumerGroupID := resourceInfo.GetGroupID(); consumerGroupID != "" {
		diffResourceTypeMap[constant.ConsumerGroup] = append(
			diffResourceTypeMap[constant.ConsumerGroup],
			consumerGroupID,
		)
	}
}
