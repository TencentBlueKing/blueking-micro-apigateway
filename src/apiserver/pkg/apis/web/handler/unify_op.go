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
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/common"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/filex"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// ResourceSync  资源同步 ...
//
//	@ID			resource_sync
//	@Summary	资源同步
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path		int						true	"网关 ID"
//	@Param		request		body		serializer.SyncRequest	true	"资源同步请求参数"
//	@Success	200			{object}	serializer.SyncResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/sync/ [post]
func ResourceSync(c *gin.Context) {
	var req serializer.SyncRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	syncedResourceTypeStats, err := biz.SyncResources(c.Request.Context(), req.ResourceType)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	resp := make(serializer.SyncResponse)
	for _, resourceType := range constant.ResourceTypeList {
		resp[resourceType] = syncedResourceTypeStats[resourceType]
	}
	ginx.SuccessJSONResponse(c, resp)
}

// ResourceRevert  资源撤销 ...
//
//	@ID			resource_revert
//	@Summary	资源撤销
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path	int							true	"网关 ID"
//	@Param		type		path	string						true	"资源类型:route/global_rule 等"
//	@Param		request		body	serializer.RevertRequest	true	"撤销资源修改请求参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/{type}/revert/ [post]
func ResourceRevert(c *gin.Context) {
	var req serializer.RevertRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	unifyOp, err := biz.NewUnifyOp(ginx.GetGatewayInfo(c), false)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	err = unifyOp.RevertConfigByIDList(c.Request.Context(), req.ResourceType, req.ResourceIDList)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// SyncedResourceManaged  同步的资源添加到编辑区 ...
//
//	@ID			synced_resource_add_edit
//	@Summary	同步的资源添加到编辑区
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path		int									true	"网关 ID"
//	@Param		request		body		serializer.ResourceManagedRequest	true	"添加资源到编辑区请求参数"
//	@Success	200			{object}	serializer.ResourceManagedResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/-/managed/ [post]
func SyncedResourceManaged(c *gin.Context) {
	var req serializer.ResourceManagedRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	syncedResourceTypeStats, err := biz.AddSyncedResources(c.Request.Context(), req.ResourceIDList)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	resp := make(serializer.ResourceManagedResponse)
	for _, resourceType := range constant.ResourceTypeList {
		resp[resourceType] = syncedResourceTypeStats[resourceType]
	}
	ginx.SuccessJSONResponse(c, resp)
}

// ResourcesDiffAll  一键资源对比 ...
//
//	@ID			resources_diff_all
//	@Summary	一键资源对比
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path		int									true	"网关 ID"
//	@Param		request		body		serializer.ResourceDiffAllRequest	false	"一键资源对比查询参数"
//	@Success	200			{object}	serializer.ResourceDiffResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/-/diff/ [post]
func ResourcesDiffAll(c *gin.Context) {
	var req serializer.ResourceDiffAllRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var idList []string
	if req.ID != "" {
		idList = append(idList, req.ID)
	}
	result, err := biz.DiffResources(c.Request.Context(),
		req.ResourceType,
		idList,
		req.Name,
		serializer.OperationTypeToResourceStatus(req.OperationType),
		true,
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, result)
}

// ResourcesDiff  资源对比 ...
//
//	@ID			resources_diff
//	@Summary	资源对比
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path		int								true	"网关 ID"
//	@Param		type		path		string							true	"资源类型:route/global_rule 等"
//	@Param		request		body		serializer.ResourceDiffRequest	false	"资源对比请求参数"
//	@Success	200			{object}	serializer.ResourceDiffResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/{type}/diff/ [post]
func ResourcesDiff(c *gin.Context) {
	var req serializer.ResourceDiffRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	result, err := biz.DiffResources(c.Request.Context(),
		constant.APISIXResource(c.Param("type")),
		req.ResourceIDList,
		req.Name,
		serializer.OperationTypeToResourceStatus(req.OperationType),
		false,
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, result)
}

// ResourceConfigDiffDetail  资源配置详情对比 ...
//
//	@ID			resource_config_diff_detail
//	@Summary	资源配置详情对比
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path		int		true	"网关 ID"
//	@Param		type		path		string	true	"资源类型:route/global_rule 等"
//	@Param		id			path		string	true	"resource ID"
//	@Success	200			{object}	dto.ResourceDiffDetailResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/{type}/diff/{id}/ [get]
func ResourceConfigDiffDetail(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	result, err := biz.GetResourceConfigDiffDetail(c.Request.Context(), pathParam.Type, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, result)
}

// ResourceDelete 资源删除 ...
//
//	@ID			resources_delete
//	@Summary	资源删除
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path	int							true	"网关 ID"
//	@Param		type		path	string						true	"资源类型:route/global_rule 等"
//	@Param		request		body	serializer.DeleteRequest	true	"删除资源请求参数"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/{type}/ [delete]
func ResourceDelete(c *gin.Context) {
	var req serializer.DeleteRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	err := biz.BatchDeleteResource(c.Request.Context(), req.ResourceType, req.ResourceIDList)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ResourceLabelsList 获取资源标签 ...
//
//	@ID			resources_labels_list
//	@Summary	获取资源标签列表
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path		int					true	"网关 ID"
//	@Param		type		path		string				true	"资源类型:route/global_rule 等"
//	@Success	200			{object}	[]map[string]string	"资源标签列表"
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/labels/{type}/ [get]
//
// ResourceLabelsList handles the deletion of resource labels
func ResourceLabelsList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	result, err := biz.GetResourcesLabels(c.Request.Context(),
		pathParam.Type)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, result)
}

// EtcdExport etcd 资源导出 ...
//
//	@ID			resources_export
//	@Summary	etcd 资源导出
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path	int	true	"网关 ID"
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/etcd/export/ [get]
//
// EtcdExport handles the export of etcd resources for a specified gateway.
// It exports all etcd resources and returns them as a downloadable JSON file.
func EtcdExport(c *gin.Context) {
	// pathParam holds the common path parameters for resource operations
	var pathParam serializer.ResourceCommonPathParam
	// Bind URI parameters to pathParam struct
	if err := c.ShouldBindUri(&pathParam); err != nil {
		// Return bad request response if URI binding fails
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	exporter, err := biz.NewUnifyOp(ginx.GetGatewayInfo(c), false)
	if err != nil {
		logging.ErrorFWithContext(c.Request.Context(), "new exporter error: %s", err.Error())
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	resources, err := exporter.ExportEtcdResources(c.Request.Context())
	if err != nil {
		logging.ErrorFWithContext(c.Request.Context(), "export etcd resources error: %s", err.Error())
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	outputs := handExportEtcdResources(resources)
	// response json
	fileData, _ := json.MarshalIndent(outputs, "", "    ")
	fileName := fmt.Sprintf("%s_export_etcd_resources.json", ginx.GetGatewayInfo(c).Name)
	ginx.SuccessFileResponse(c, "text/plain", fileData, fileName)
}

// handExportEtcdResources 处理导出etcd资源
func handExportEtcdResources(resources []*model.GatewaySyncData) serializer.EtcdExportOutput {
	outputs := make(serializer.EtcdExportOutput)
	for _, resource := range resources {
		if resource.ID == "" {
			resource.ID = idx.GenResourceID(resource.Type)
		}
		resourceOutput := serializer.ResourceInfo{
			ResourceType: resource.Type,
			ResourceID:   resource.ID,
			Name:         resource.GetName(),
			Config:       json.RawMessage(resource.Config),
		}
		if _, ok := outputs[resource.Type]; !ok {
			outputs[resource.Type] = []serializer.ResourceInfo{resourceOutput}
			continue
		}
		outputs[resource.Type] = append(outputs[resource.Type], resourceOutput)
	}
	return outputs
}

// ResourceUpload 资源上传 ...
//
//	@ID			resources_upload
//	@Summary	资源导入
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Accept		multipart/form-data
//	@Param		resource_file	formData	file					true	"资源配置文件(json)"
//	@Param		gateway_id		path		int						true	"网关 ID"
//	@Success	200				{object}	dto.ResourceUploadInfo	"导入资源列表"
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/upload/ [post]
//
// ResourceUpload handles the upload of resource configuration files for import.
// It processes the uploaded file, validates the resources, and returns the imported resource list.
func ResourceUpload(c *gin.Context) {
	// pathParam holds the common path parameters for resource operations
	var pathParam serializer.ResourceCommonPathParam
	// Bind URI parameters to pathParam struct
	if err := c.ShouldBindUri(&pathParam); err != nil {
		// Return bad request response if URI binding fails
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	fileHeader, err := c.FormFile("resource_file")
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var resourceInfoTypeMap map[constant.APISIXResource][]common.ResourceInfo
	if err := filex.ReadFileToObject(fileHeader, &resourceInfoTypeMap); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// check 配置
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	existsResourceIdList := make(map[string]struct{})
	for resourceType, resourceInfoList := range resourceInfoTypeMap {
		dbResources, err := biz.BatchGetResources(c.Request.Context(), resourceType, []string{})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		for _, dbResource := range dbResources {
			existsResourceIdList[dbResource.ID] = struct{}{}
		}
		for _, resourceInfo := range resourceInfoList {
			res := &model.GatewaySyncData{
				Type:   resourceInfo.ResourceType,
				ID:     resourceInfo.ResourceID,
				Config: datatypes.JSON(resourceInfo.Config),
			}
			if _, ok := resourceTypeMap[resourceInfo.ResourceType]; ok {
				resourceTypeMap[resourceInfo.ResourceType] = append(resourceTypeMap[resourceInfo.ResourceType], res)
				continue
			}
			resourceTypeMap[resourceInfo.ResourceType] = []*model.GatewaySyncData{res}
		}
	}
	err = biz.ValidateResource(c.Request.Context(), resourceTypeMap)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, fmt.Errorf("resource validate failed, err: %v", err))
		return
	}
	resources, err := common.ClassifyImportResourceInfo(resourceInfoTypeMap, existsResourceIdList)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, resources)
}

// ResourceImport 资源导入 ...
//
//	@ID			resources_import
//	@Summary	资源导入
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.unify_op
//	@Param		gateway_id	path	int						true	"网关 ID"
//	@Param		request		body	dto.ResourceUploadInfo	true	"待导入的资源列表"
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/import/ [post]
//
// ResourceImport handles importing resources from the request body,
// validates them, and inserts them into the system.
func ResourceImport(c *gin.Context) {
	// pathParam holds the common path parameters for resource operations
	var pathParam serializer.ResourceCommonPathParam
	// Bind URI parameters to pathParam struct
	if err := c.ShouldBindUri(&pathParam); err != nil {
		// Return bad request response if URI binding fails
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var resourcesImport common.ResourceUploadInfo
	if err := c.ShouldBindJSON(&resourcesImport); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	addResourcesMap, updateResourcesMap, err := common.HandleImportResources(c.Request.Context(), &resourcesImport)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// 插入数据
	err = biz.UploadResources(c.Request.Context(), addResourcesMap, updateResourcesMap)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}
