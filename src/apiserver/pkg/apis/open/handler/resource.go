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

// Package handler ...
package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/common"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/open/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/filex"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ResourceBatchCreate ...
//
//	@ID			openapi_resource_batch_create
//	@Summary	资源批量创建
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header		string									true	"创建网关返回的 token"
//	@Param		gateway_name	path		string									true	"网关名称"
//	@Param		resource_type	path		constant.ResourcePath					true	"资源类型"
//	@Param		request			body		serializer.ResourceBatchCreateRequest	true	"资源创建参数"
//	@Success	200				{object}	[]serializer.ResourceCreateResponse
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/ [post]
func ResourceBatchCreate(c *gin.Context) {
	var req serializer.ResourceBatchCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	nameMap := make(map[string]struct{})
	var names []string
	for _, resource := range req {
		if _, ok := nameMap[resource.Name]; ok {
			ginx.BadRequestErrorJSONResponse(
				c,
				fmt.Errorf("resource name: %s is not unique", resource.Name),
			)
			return
		}
		nameMap[resource.Name] = struct{}{}
		names = append(names, resource.Name)
	}
	// 校验资源名称是否与已有的重复
	duplicated, err := biz.BatchCheckNameDuplication(c.Request.Context(), ginx.GetResourceType(c), names)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	if duplicated {
		ginx.BadRequestErrorJSONResponse(c,
			errors.New("resource name is duplicated with existing"))
		return
	}
	resources := req.ToCommonResource(ginx.GetGatewayInfo(c).ID, ginx.GetResourceType(c))
	// 批量创建资源
	err = biz.BatchCreateResources(c.Request.Context(), ginx.GetResourceType(c), resources)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	var res []serializer.ResourceCreateResponse
	for _, resource := range resources {
		res = append(res, serializer.ResourceCreateResponse{
			ID: resource.ID,
			Name: gjson.GetBytes(
				resource.Config,
				model.GetResourceNameKey(ginx.GetResourceType(c)),
			).String(),
		})
	}
	ginx.SuccessJSONResponse(c, res)
}

// ResourceBatchGet ...
//
//	@ID			openapi_resource_batch_get
//	@Summary	资源批量查询
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header		string								true	"创建网关返回的 token"
//	@Param		gateway_name	path		string								true	"网关名称"
//	@Param		resource_type	path		constant.ResourcePath				true	"资源类型"
//	@Param		request			query		serializer.ResourceBatchGetRequest	true	"资源查询参数"
//	@Success	200				{object}	[]serializer.ResourceBatchGetResponse
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/ [get]
func ResourceBatchGet(c *gin.Context) {
	var req serializer.ResourceBatchGetRequest
	if err := c.BindQuery(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	resources, err := biz.BatchGetResources(c.Request.Context(), ginx.GetResourceType(c), req.IDs)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	var res []serializer.ResourceBatchGetResponse
	for _, resource := range resources {
		configRaw, _ := resource.Config.MarshalJSON()
		res = append(res, serializer.ResourceBatchGetResponse{
			ID:         resource.ID,
			RawMessage: configRaw,
		})
	}
	ginx.SuccessJSONResponse(c, res)
}

// ResourceBatchDelete ...
//
//	@ID			openapi_resource_batch_delete
//	@Summary	资源批量删除
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header	string									true	"创建网关返回的 token"
//	@Param		gateway_name	path	string									true	"网关名称"
//	@Param		resource_type	path	constant.ResourcePath					true	"资源类型"
//	@Param		request			body	serializer.ResourceBatchDeleteRequest	true	"批量删除资源参数"
//	@Success	204
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/batch_delete [post]
func ResourceBatchDelete(c *gin.Context) {
	var req serializer.ResourceBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	// 状态机判断
	resources, err := biz.BatchGetResources(c.Request.Context(), ginx.GetResourceType(c), req.IDs)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	resourceIDMap := make(map[string]*model.ResourceCommonModel)
	for _, resource := range resources {
		resourceIDMap[resource.ID] = resource
		statusOp := status.NewResourceStatusOp(*resource)
		err = statusOp.CanDo(c.Request.Context(), constant.OperationTypeRevert)
		if err != nil {
			ginx.BadRequestErrorJSONResponse(c,
				fmt.Errorf("resource: %s can not do delete: %s", resource.ID, err.Error()))
			return
		}
	}
	err = biz.BatchUpdateResourceStatusWithAuditLog(c.Request.Context(),
		ginx.GetResourceType(c), req.IDs, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ResourceGet ...
//
//	@ID			openapi_resource_get
//	@Summary	资源详情
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header		string					true	"创建网关返回的 token"
//	@Param		gateway_name	path		string					true	"网关名称"
//	@Param		resource_type	path		constant.ResourcePath	true	"资源类型"
//	@Param		id				path		string					true	"资源 ID"
//	@Success	200				{object}	serializer.ResourceGetResponse
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/{id}/ [get]
func ResourceGet(c *gin.Context) {
	var pathParam serializer.ResourcePathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	resource, err := biz.GetResourceByID(c.Request.Context(), ginx.GetResourceType(c), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	configRaw, _ := resource.Config.MarshalJSON()
	res := serializer.ResourceGetResponse{
		ID:         resource.ID,
		RawMessage: configRaw,
	}
	ginx.SuccessJSONResponse(c, res)
}

// ResourceGetStatus ...
//
//	@ID			openapi_resource_get_status
//	@Summary	资源状态
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header		string					true	"创建网关返回的 token"
//	@Param		gateway_name	path		string					true	"网关名称"
//	@Param		resource_type	path		constant.ResourcePath	true	"资源类型"
//	@Param		id				path		string					true	"资源 ID"
//	@Success	200				{object}	serializer.ResourceGetStatusResponse
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/{id}/status/ [get]
func ResourceGetStatus(c *gin.Context) {
	var pathParam serializer.ResourcePathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	resource, err := biz.GetResourceByID(c.Request.Context(), ginx.GetResourceType(c), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	res := serializer.ResourceGetStatusResponse{
		ID:     resource.ID,
		Status: resource.Status,
	}
	ginx.SuccessJSONResponse(c, res)
}

// ResourceUpdate ...
//
//	@ID			openapi_resource_update
//	@Summary	资源更新
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header	string								true	"创建网关返回的 token"
//	@Param		gateway_name	path	string								true	"网关名称"
//	@Param		resource_type	path	constant.ResourcePath				true	"资源类型"
//	@Param		id				path	string								true	"资源 ID"
//	@Param		request			body	serializer.ResourceUpdateRequest	true	"资源更新参数"
//	@Success	201
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/{id}/ [put]
func ResourceUpdate(c *gin.Context) {
	var pathParam serializer.ResourcePathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.ResourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	duplicated := biz.DuplicatedResourceName(c.Request.Context(), ginx.GetResourceType(c), pathParam.ID, req.Name)
	if !duplicated {
		ginx.BadRequestErrorJSONResponse(c, errors.New(
			fmt.Sprintf("name: %s is duplicated with existing %s", req.Name, ginx.GetResourceType(c)),
		))
		return
	}

	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), ginx.GetResourceType(c), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	resource := req.ToCommonResource(c, pathParam.ID, updateStatus)
	// 更新资源
	err = biz.UpdateResource(c.Request.Context(), ginx.GetResourceType(c), pathParam.ID, resource)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
}

// ResourceDelete ...
//
//	@ID			openapi_resource_delete
//	@Summary	资源删除
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header	string					true	"创建网关返回的 token"
//	@Param		gateway_name	path	string					true	"网关名称"
//	@Param		resource_type	path	constant.ResourcePath	true	"资源类型"
//	@Param		id				path	string					true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/{id}/ [delete]
func ResourceDelete(c *gin.Context) {
	var pathParam serializer.ResourcePathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	err := biz.UpdateResourceStatusWithAuditLog(
		c.Request.Context(),
		ginx.GetResourceType(c),
		pathParam.ID,
		constant.ResourceStatusDeleteDraft,
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ResourcePublish ...
//
//	@ID			openapi_resource_publish
//	@Summary	资源发布
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header	string								true	"创建网关返回的 token"
//	@Param		gateway_name	path	string								true	"网关名称"
//	@Param		resource_type	path	constant.ResourcePath				true	"资源类型"
//	@Param		request			body	serializer.ResourcePublishRequest	true	"资源删除参数"
//	@Success	201
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/{resource_type}/publish/ [post]
func ResourcePublish(c *gin.Context) {
	var req serializer.ResourcePublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	err := biz.PublishResource(c.Request.Context(), ginx.GetResourceType(c), req.IDs)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// ResourceImport ...
//
//	@ID			openapi_resource_import
//	@Summary	资源导入
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.resource
//	@Param		X-BK-API-TOKEN	header	string	true	"创建网关返回的 token"
//	@Param		gateway_name	path	string	true	"网关名称"
//	@Accept		multipart/form-data
//	@Param		resource_file	formData	file	true	"资源配置文件 (json)"
//	@Success	200				{object}	common.ResourceUploadInfo
//	@Router		/api/v1/open/gateways/{gateway_name}/resources/-/import/ [post]
func ResourceImport(c *gin.Context) {
	fileHeader, err := c.FormFile("resource_file")
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var resourceImport serializer.ResourceImportRequest
	if err := filex.ReadFileToObject(fileHeader, &resourceImport); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	handlerResourceIndexResult, err := common.HandlerResourceIndexMap(c.Request.Context(),
		resourceImport.Data)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	uploadInfo, err := common.ClassifyImportResourceInfo(
		resourceImport.Data,
		handlerResourceIndexResult.ExistsResourceIdList,
		handlerResourceIndexResult.AddedSchemaMap,
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	handlerResult, err := common.HandleUploadResources(c.Request.Context(),
		uploadInfo, handlerResourceIndexResult.AllSchemaMap, resourceImport.Metadata.IgnoreFields)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// 插入数据
	err = biz.UploadResources(
		c.Request.Context(),
		handlerResult.AddResourceTypeMap,
		handlerResult.UpdateResourceTypeMap,
		handlerResourceIndexResult.AddedSchemaMap,
		handlerResourceIndexResult.UpdatedSchemaMap,
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, uploadInfo)
}
