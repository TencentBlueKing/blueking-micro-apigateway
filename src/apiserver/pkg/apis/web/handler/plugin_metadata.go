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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// PluginMetadataCreate ...
//
//	@ID			plugin_metadata_create
//	@Summary	plugin_metadata 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.plugin_metadata
//	@Param		gateway_id	path	int								true	"网关 ID"
//	@Param		request		body	serializer.PluginMetadataInfo	true	"plugin_metadata 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/plugin_metadatas/ [post]
func PluginMetadataCreate(c *gin.Context) {
	var req serializer.PluginMetadataInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	pluginMetadata := model.PluginMetadata{
		Name: req.Name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginMetadata),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.CreatePluginMetadata(c.Request.Context(), pluginMetadata); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// PluginMetadataUpdate ...
//
//	@ID			plugin_metadata_update
//	@Summary	plugin_metadata 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.plugin_metadata
//	@Param		gateway_id	path	int								true	"网关ID"
//	@Param		id			path	string							true	"plugin_metadata ID"
//	@Param		request		body	serializer.PluginMetadataInfo	true	"plugin_metadata更新参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/plugin_metadatas/{id}/ [put]
func PluginMetadataUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	req := serializer.PluginMetadataInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.PluginMetadata, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	pluginMetadata := model.PluginMetadata{
		Name: req.Name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        pathParam.ID,
			GatewayID: pathParam.GatewayID,
			Config:    datatypes.JSON(req.Config),
			Status:    updateStatus,
			BaseModel: model.BaseModel{
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.UpdatePluginMetadata(c.Request.Context(), pluginMetadata); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
}

// PluginMetadataList ...
//
//	@ID			plugin_metadata_list
//	@Summary	plugin_metadata 列表
//	@Produce	json
//	@Tags		webapi.plugin_metadata
//	@Param		gateway_id	path		int										true	"网关 ID"
//	@Param		request		query		serializer.PluginMetadataListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.PluginMetadataListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/plugin_metadatas/ [get]
func PluginMetadataList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.PluginMetadataListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	queryParam := map[string]interface{}{}
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	pluginMetadataList, total, err := biz.ListPagedPluginMetadatas(
		c.Request.Context(),
		queryParam,
		strings.Split(req.Status, ","),
		req.Name,
		req.Updater,
		req.OrderBy,
		biz.PageParam{
			Offset: ginx.GetOffset(c),
			Limit:  ginx.GetLimit(c),
		},
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	var results serializer.PluginMetadataListResponse
	for _, pluginMetadata := range pluginMetadataList {
		results = append(results, serializer.PluginMetadataOutputInfo{
			ID:        pluginMetadata.ID,
			AutoID:    pluginMetadata.AutoID,
			GatewayID: pluginMetadata.GatewayID,
			PluginMetadataInfo: serializer.PluginMetadataInfo{
				ID:     pluginMetadata.ID,
				Name:   pluginMetadata.Name,
				Config: json.RawMessage(pluginMetadata.Config),
			},
			Status:    pluginMetadata.Status,
			CreatedAt: pluginMetadata.CreatedAt.Unix(),
			UpdatedAt: pluginMetadata.UpdatedAt.Unix(),
			Creator:   pluginMetadata.Creator,
			Updater:   pluginMetadata.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// PluginMetadataGet ...
//
//	@ID			plugin_metadata_get
//	@Summary	plugin_metadata 详情
//	@Produce	json
//	@Tags		webapi.plugin_metadata
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.PluginMetadataOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/plugin_metadatas/{id}/ [get]
func PluginMetadataGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	pluginMetadata, err := biz.GetPluginMetadata(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	pluginMetadata.Config = []byte(jsonx.RemoveJsonKey(string(pluginMetadata.Config), []string{"id", "name"}))
	output := serializer.PluginMetadataOutputInfo{
		AutoID:    pluginMetadata.AutoID,
		ID:        pluginMetadata.ID,
		GatewayID: pluginMetadata.GatewayID,
		PluginMetadataInfo: serializer.PluginMetadataInfo{
			ID:     pluginMetadata.ID,
			Name:   pluginMetadata.Name,
			Config: json.RawMessage(pluginMetadata.Config),
		},
		CreatedAt: pluginMetadata.CreatedAt.Unix(),
		UpdatedAt: pluginMetadata.UpdatedAt.Unix(),
		Creator:   pluginMetadata.Creator,
		Updater:   pluginMetadata.Updater,
		Status:    pluginMetadata.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// PluginMetadataDelete ...
//
//	@ID			plugin_metadata_delete
//	@Summary	plugin_metadata 删除
//	@Produce	json
//	@Tags		webapi.plugin_metadata
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/plugin_metadatas/{id}/ [delete]
func PluginMetadataDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	pluginMetadata, err := biz.GetPluginMetadata(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if pluginMetadata.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeletePluginMetadatas(c.Request.Context(), []string{pluginMetadata.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}

	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.PluginMetadata, pluginMetadata.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// PluginMetadataDropDownList ...
//
//	@ID			plugin_metadata_dropdown_list
//	@Summary	plugin_metadata 下拉列表
//	@Produce	json
//	@Tags		webapi.plugin_metadata
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.PluginMetadataDropDownResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/plugin_metadatas-dropdown/ [get]
func PluginMetadataDropDownList(c *gin.Context) {
	pluginMetadatas, err := biz.ListPluginMetadatas(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.PluginMetadataDropDownResponse
	for _, pluginMetadata := range pluginMetadatas {
		desc := gjson.ParseBytes(pluginMetadata.Config).Get("desc").String()
		output = append(output, serializer.PluginMetadataDropDownOutputInfo{
			AutoID: pluginMetadata.AutoID,
			ID:     pluginMetadata.ID,
			Name:   pluginMetadata.Name,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
