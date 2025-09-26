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

package handler

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// OperationAuditLogList ...
//
//	@ID			operation_audit_log_list
//	@Summary	审计日志 列表
//	@Produce	json
//	@Tags		webapi.operation_audit_log
//	@Param		gateway_id	path		int										true	"网关 ID"
//	@Param		request		query		serializer.OperationAuditLogListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=[]serializer.OperationAuditLogListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/audits/logs/ [get]
func OperationAuditLogList(c *gin.Context) {
	var req serializer.OperationAuditLogListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	queryParam := map[string]interface{}{
		"gateway_id": c.Param("gateway_id"),
	}
	if req.OperationType != "" {
		queryParam["operation_type"] = req.OperationType
	}
	if req.ResourceType != "" {
		queryParam["resource_type"] = req.ResourceType
	}
	var results []serializer.OperationAuditLogListResponse
	var total int64
	var err error
	if req.Name != "" {
		results, total, err = getOperationAuditLogResultsByName(
			c.Request.Context(),
			req,
			queryParam,
			ginx.GetOffset(c),
			ginx.GetLimit(c),
		)
	} else {
		results, total, err = getOperationAuditLogResults(
			c.Request.Context(),
			req,
			queryParam,
			ginx.GetOffset(c),
			ginx.GetLimit(c),
		)
	}
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// getOperationAuditLogResourceIDNames 获取审计日志资源名称
func getOperationAuditLogResourceIDNames(
	ctx context.Context,
	operationAuditLogs []*model.OperationAuditLog,
) (map[string]string, error) {
	// 按资源类型分类 IDs
	auditLogResourceTypeIDMap := map[constant.APISIXResource][]string{}
	deleteResourceIDNameMap := map[string]string{}
	for _, log := range operationAuditLogs {
		if log.ResourceType == "" || log.ResourceIDs == "" {
			continue
		}
		// 资源删除后，无法从数据库中查询到资源名称，获取删除前 config 中存储的资源名称
		if log.OperationType == constant.OperationTypeDelete {
			for _, data := range gjson.ParseBytes(log.DataBefore).Array() {
				resourceID := data.Get("id").String()
				resourceName := ""
				if log.ResourceType == constant.Schema {
					// schema 审计日志类型需要单独处理 Name
					resourceName = data.Get("config").Get("Name").String()
				} else {
					resourceName = data.Get("config").Get(model.GetResourceNameKey(log.ResourceType)).String()
				}
				if _, ok := deleteResourceIDNameMap[resourceID]; !ok && resourceName != "" {
					deleteResourceIDNameMap[resourceID] = resourceName
				}
			}
		}
		resourceIDs := strings.Split(log.ResourceIDs, ",")
		auditLogResourceTypeIDMap[log.ResourceType] = append(
			auditLogResourceTypeIDMap[log.ResourceType],
			resourceIDs...,
		)
	}
	// 根据资源类型+资源 IDs 获取对应资源名称
	resourceIDNameMap := map[string]string{}
	for resourceType, resourceIDs := range auditLogResourceTypeIDMap {
		switch resourceType {
		case constant.Gateway:
			for _, resourceID := range resourceIDs {
				resourceIDNameMap[resourceID] = ginx.GetGatewayInfoFromContext(ctx).Name
			}
		case constant.Schema:
			schemas, err := biz.GetSchemaByIDs(ctx, resourceIDs)
			if err != nil {
				return nil, err
			}
			for _, schema := range schemas {
				resourceIDNameMap[strconv.Itoa(schema.AutoID)] = schema.Name
			}
		default:
			resources, err := biz.GetResourceByIDs(ctx, resourceType, resourceIDs)
			if err != nil {
				return nil, err
			}
			for _, resource := range resources {
				resourceIDNameMap[resource.ID] = resource.GetName(resourceType)
			}
		}
	}
	// 可能存在删除后又撤销的操作，所以检查已删除的资源 ID 是否能够被匹配到，未匹配到需将删除前的资源名称补充进去
	for id, name := range deleteResourceIDNameMap {
		if _, ok := resourceIDNameMap[id]; !ok {
			resourceIDNameMap[id] = name
		}
	}
	return resourceIDNameMap, nil
}

// getOperationAuditLogResults 获取审计日志查询结果
func getOperationAuditLogResults(
	ctx context.Context,
	req serializer.OperationAuditLogListRequest,
	queryParam map[string]interface{},
	offset int,
	limit int,
) ([]serializer.OperationAuditLogListResponse, int64, error) {
	operationAuditLogs, total, err := biz.ListPagedOperationAuditLogs(
		ctx,
		queryParam,
		req.ResourceID,
		req.Operator,
		req.TimeStart,
		req.TimeEnd,
		biz.PageParam{
			Offset: offset,
			Limit:  limit,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	resourceIDNameMap, err := getOperationAuditLogResourceIDNames(ctx, operationAuditLogs)
	if err != nil {
		return nil, 0, err
	}
	var results []serializer.OperationAuditLogListResponse
	for _, log := range operationAuditLogs {
		resourceNames := []string{}
		resourceIDs := strings.Split(log.ResourceIDs, ",")
		for _, resourceID := range resourceIDs {
			// 根据 ID 获取对应资源名称
			if name, ok := resourceIDNameMap[resourceID]; ok {
				resourceNames = append(resourceNames, name)
			}
		}
		results = append(results, serializer.OperationAuditLogListResponse{
			ID:            log.ID,
			OperationType: log.OperationType,
			Names:         resourceNames,
			ResourceIDs:   resourceIDs,
			Operator:      log.Operator,
			DataBefore:    json.RawMessage(log.DataBefore),
			DataAfter:     json.RawMessage(log.DataAfter),
			ResourceType:  log.ResourceType,
			CreatedAt:     log.CreatedAt.Unix(),
		})
	}
	return results, total, nil
}

// getOperationAuditLogResultsByName 获取审计日志查询结果，按照 name 过滤
func getOperationAuditLogResultsByName(
	ctx context.Context,
	req serializer.OperationAuditLogListRequest,
	queryParam map[string]interface{},
	offset int,
	limit int,
) ([]serializer.OperationAuditLogListResponse, int64, error) {
	operationAuditLogs, err := biz.ListOperationAuditLogs(
		ctx,
		queryParam,
		req.ResourceID,
		req.Operator,
		req.TimeStart,
		req.TimeEnd,
	)
	if err != nil {
		return nil, 0, err
	}
	resourceIDNameMap, err := getOperationAuditLogResourceIDNames(ctx, operationAuditLogs)
	if err != nil {
		return nil, 0, err
	}
	var results []serializer.OperationAuditLogListResponse
	for _, log := range operationAuditLogs {
		resourceNames := []string{}
		resourceIDs := strings.Split(log.ResourceIDs, ",")
		shouldInclude := false
		for _, resourceID := range resourceIDs {
			// 根据 ID 获取对应资源名称
			if name, ok := resourceIDNameMap[resourceID]; ok {
				resourceNames = append(resourceNames, name)
				// 过滤 name
				if strings.Contains(name, req.Name) {
					shouldInclude = true
				}
			}
		}
		if !shouldInclude {
			continue
		}
		results = append(results, serializer.OperationAuditLogListResponse{
			ID:            log.ID,
			OperationType: log.OperationType,
			Names:         resourceNames,
			ResourceIDs:   resourceIDs,
			Operator:      log.Operator,
			DataBefore:    json.RawMessage(log.DataBefore),
			DataAfter:     json.RawMessage(log.DataAfter),
			ResourceType:  log.ResourceType,
			CreatedAt:     log.CreatedAt.Unix(),
		})
	}

	// name 搜索时，需要单独处理分页
	offset, end := serializer.PaginateResults(len(results), offset, limit)
	total := int64(len(results))
	results = results[offset:end]

	return results, total, nil
}
