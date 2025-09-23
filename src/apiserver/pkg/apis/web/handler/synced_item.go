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

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - APIGateway) available.
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

package handler

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// SyncedItemList ...
//
//	@ID			sync_data_list
//	@Summary	sync data list
//	@Produce	json
//	@Tags		webapi.sync_data
//	@Param		gateway_id	path		int																true	"网关 ID"
//	@Param		request		query		serializer.SyncedItemListRequestRequest							false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.SyncDataListResponse}	"desc"
//	@Router		/api/v1/web/gateways/{gateway_id}/synced/items/ [get]
func SyncedItemList(c *gin.Context) {
	var req serializer.SyncedItemListRequestRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	queryParam := map[string]interface{}{}
	queryParam["gateway_id"] = ginx.GetGatewayInfo(c).ID
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	if req.ResourceType != "" {
		queryParam["type"] = req.ResourceType
	}
	var syncDataList []*model.GatewaySyncData
	var total int64
	var err error
	if req.Status != "" || req.Name != "" {
		syncDataList, err = biz.QuerySyncedItems(c.Request.Context(), queryParam)
	} else {
		syncDataList, total, err = biz.ListPagedSyncedItems(
			c.Request.Context(),
			queryParam,
			biz.PageParam{
				Offset: ginx.GetOffset(c),
				Limit:  ginx.GetLimit(c),
			},
		)
	}
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	results, err := enrichOutputInfo(c.Request.Context(), syncDataList, req.Status, req.Name)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// 处理排序
	results = sortResults(results, req.OrderBy)
	if req.Status != "" || req.Name != "" {
		// status/name 搜索时，需要单独处理分页
		offset, end := serializer.PaginateResults(len(results), ginx.GetOffset(c), ginx.GetLimit(c))
		total = int64(len(results))
		results = results[offset:end]
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// enrichOutputInfo 丰富输出信息，补充同步资源状态、来源等数据
func enrichOutputInfo(
	ctx context.Context,
	syncDataList []*model.GatewaySyncData,
	status constant.SyncStatus,
	name string,
) ([]*serializer.SyncDataOutputInfo, error) {
	resourceIDMap := make(map[constant.APISIXResource][]string)      // resourceType:[]id
	outputDataMap := make(map[string]*serializer.SyncDataOutputInfo) // id:sync
	var output []*serializer.SyncDataOutputInfo
	var filterOutput []*serializer.SyncDataOutputInfo
	for _, sync := range syncDataList {
		if idList, ok := resourceIDMap[sync.Type]; ok {
			resourceIDMap[sync.Type] = append(idList, sync.ID)
		} else {
			resourceIDMap[sync.Type] = []string{sync.ID}
		}
		syncData := &serializer.SyncDataOutputInfo{
			ID:           sync.ID,
			GatewayID:    sync.GatewayID,
			ResourceType: sync.Type,
			ModeRevision: sync.ModRevision,
			Config:       json.RawMessage(sync.Config),
			Status:       constant.SyncedResourceStatusSuccess,
			CreatedAt:    sync.CreatedAt.Unix(),
			UpdatedAt:    sync.UpdatedAt.Unix(),
		}
		outputDataMap[sync.ID] = syncData
		output = append(output, syncData)
	}

	dbResourceIDMap := make(map[string]*model.ResourceCommonModel)
	for resourceType, idList := range resourceIDMap {
		dbResources, err := biz.BatchGetResources(ctx, resourceType, idList)
		if err != nil {
			return nil, err
		}
		for _, dbResource := range dbResources {
			dbResourceIDMap[dbResource.ID] = dbResource
		}
	}
	for _, sync := range output {
		if _, ok := dbResourceIDMap[sync.ID]; !ok {
			sync.Status = constant.SyncedResourceStatusMiss
		}
		// 判断发布来源 todo: 根据标签来判断
		if strings.HasPrefix(sync.ID, "bk.") {
			sync.PublishSource = constant.PublishByBkAPISIXControlPlane
		} else {
			sync.PublishSource = constant.PublishByOthers
		}
		// 过滤同步状态
		if status != "" && sync.Status != status {
			continue
		}
		// 过滤名称
		resourceName := ""
		if sync.ResourceType == constant.Consumer {
			resourceName = gjson.ParseBytes(sync.Config).Get("username").String()
		} else {
			resourceName = gjson.ParseBytes(sync.Config).Get("name").String()
		}
		if name != "" && !strings.Contains(resourceName, name) {
			continue
		}
		// 记录资源名称，用于后面的排序
		sync.Name = resourceName
		filterOutput = append(filterOutput, sync)
	}

	return filterOutput, nil
}

// sortResults 对结果进行排序
func sortResults(results []*serializer.SyncDataOutputInfo, orderBy string) []*serializer.SyncDataOutputInfo {
	if orderBy == "" {
		return results
	}
	sortConditions := strings.Split(orderBy, ",")
	for _, condition := range sortConditions {
		parts := strings.Split(condition, ":")
		if len(parts) != 2 {
			continue
		}

		fieldName := parts[0]
		direction := strings.ToLower(parts[1])

		sort.Slice(results, func(i, j int) bool {
			var result bool
			if fieldName == "name" {
				result = results[i].Name < results[j].Name
			}

			// 如果是降序，反转结果
			if direction == "desc" {
				result = !result
			}
			return result
		})
	}
	return results
}

// SyncedLastTime ...
//
//	@ID			latest_sync_time
//	@Summary	最新同步时间
//	@Produce	json
//	@Tags		webapi.sync_data
//	@Param		gateway_id	path		int								true	"网关 ID"
//	@Success	200			{object}	serializer.SyncedTimeOutputInfo	"同步时间"
//	@Router		/api/v1/web/gateways/{gateway_id}/synced/last_time/ [get]
func SyncedLastTime(c *gin.Context) {
	var latestTime int64
	if !ginx.GetGatewayInfo(c).LastSyncedAt.IsZero() {
		latestTime = ginx.GetGatewayInfo(c).LastSyncedAt.Unix()
	}
	output := serializer.SyncedTimeOutputInfo{
		LatestTime: latestTime,
	}
	ginx.SuccessJSONResponse(c, output)
}
