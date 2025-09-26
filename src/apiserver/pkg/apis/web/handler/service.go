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
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// ServiceCreate ...
//
//	@ID			service_create
//	@Summary	service 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.service
//	@Param		gateway_id	path	int						true	"网关 ID"
//	@Param		request		body	serializer.ServiceInfo	true	"service 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/services/ [post]
func ServiceCreate(c *gin.Context) {
	var req serializer.ServiceInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	service := model.Service{
		Name:       req.Name,
		UpstreamID: req.UpstreamID,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Service),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.CreateService(c.Request.Context(), service); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// ServiceUpdate ...
//
//	@ID			service_update
//	@Summary	service 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.service
//	@Param		gateway_id	path	int						true	"网关ID"
//	@Param		id			path	string					true	"service ID"
//	@Param		request		body	serializer.ServiceInfo	true	"service更新参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/services/{id}/ [put]
func ServiceUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	req := serializer.ServiceInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.Service, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	service := model.Service{
		Name:       req.Name,
		UpstreamID: req.UpstreamID,
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

	if err := biz.UpdateService(c.Request.Context(), service); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
}

// ServiceList ...
//
//	@ID			service_list
//	@Summary	service 列表
//	@Produce	json
//	@Tags		webapi.service
//	@Param		gateway_id	path		int								true	"网关 ID"
//	@Param		request		query		serializer.ServiceListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.ServiceListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/services/ [get]
func ServiceList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.ServiceListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	labelMap, err := serializer.CheckLabel(req.Label)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	queryParam := map[string]interface{}{}
	queryParam["gateway_id"] = pathParam.GatewayID
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	services, total, err := biz.ListPagedServices(
		c.Request.Context(),
		queryParam,
		labelMap,
		strings.Split(req.Status, ","),
		req.Name,
		req.Updater,
		req.UpstreamID,
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
	var results serializer.ServiceListResponse
	for _, service := range services {
		results = append(results, serializer.ServiceOutputInfo{
			AutoID:    service.AutoID,
			GatewayID: service.GatewayID,
			ServiceInfo: serializer.ServiceInfo{
				ID:         service.ID,
				Name:       service.Name,
				UpstreamID: service.UpstreamID,
				Config:     json.RawMessage(service.Config),
			},
			Status:    service.Status,
			CreatedAt: service.CreatedAt.Unix(),
			UpdatedAt: service.UpdatedAt.Unix(),
			Creator:   service.Creator,
			Updater:   service.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// ServiceGet ...
//
//	@ID			service_get
//	@Summary	service 详情
//	@Produce	json
//	@Tags		webapi.service
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.ServiceOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/services/{id}/ [get]
func ServiceGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	service, err := biz.GetService(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	output := serializer.ServiceOutputInfo{
		GatewayID: service.GatewayID,
		AutoID:    service.AutoID,
		ServiceInfo: serializer.ServiceInfo{
			ID:         service.ID,
			Name:       service.Name,
			UpstreamID: service.UpstreamID,
			Config:     json.RawMessage(service.Config),
		},
		CreatedAt: service.CreatedAt.Unix(),
		UpdatedAt: service.UpdatedAt.Unix(),
		Creator:   service.Creator,
		Updater:   service.Updater,
		Status:    service.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// ServiceDelete ...
//
//	@ID			service_delete
//	@Summary	service 删除
//	@Produce	json
//	@Tags		webapi.service
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/services/{id}/ [delete]
func ServiceDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	service, err := biz.GetService(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if service.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteServices(c.Request.Context(), []string{service.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.Service, service.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ServiceDropDownList ...
//
//	@ID			service_dropdown_list
//	@Summary	service 下拉列表
//	@Produce	json
//	@Tags		webapi.service
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.ServiceDropDownListResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/services-dropdown/ [get]
func ServiceDropDownList(c *gin.Context) {
	services, err := biz.ListServices(c.Request.Context(), ginx.GetGatewayInfo(c).ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.ServiceDropDownListResponse
	for _, service := range services {
		desc := gjson.ParseBytes(service.Config).Get("desc").String()
		output = append(output, serializer.ServiceDropDownOutputInfo{
			AutoID: service.AutoID,
			ID:     service.ID,
			Name:   service.Name,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
