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

// StreamRouteCreate ...
//
//	@ID			stream_route_create
//	@Summary	stream_route 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.stream_route
//	@Param		gateway_id	path	int							true	"网关 ID"
//	@Param		request		body	serializer.StreamRouteInfo	true	"stream_route 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/stream_routes/ [post]
func StreamRouteCreate(c *gin.Context) {
	var req serializer.StreamRouteInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	streamRoute := model.StreamRoute{
		Name:       req.Name,
		ServiceID:  req.ServiceID,
		UpstreamID: req.UpstreamID,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.StreamRoute), // todo: generate
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.CreateStreamRoute(c.Request.Context(), streamRoute); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// StreamRouteUpdate ...
//
//	@ID			stream_route_update
//	@Summary	stream_route 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.stream_route
//	@Param		gateway_id	path	int							true	"网关ID"	@Param	id	path	string	true	"stream_route ID"
//	@Param		request		body	serializer.StreamRouteInfo	true	"stream_route 更新参数"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/stream_routes/{id}/ [put]
func StreamRouteUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	req := serializer.StreamRouteInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.StreamRoute, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	streamRoute := model.StreamRoute{
		Name:       req.Name,
		ServiceID:  req.ServiceID,
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
	if err := biz.UpdateStreamRoute(c.Request.Context(), streamRoute); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// StreamRouteGet ...
//
//	@ID			stream_route_get
//	@Summary	stream_route 详情
//	@Produce	json
//	@Tags		webapi.stream_route
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.StreamRouteOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/stream_routes/{id}/ [get]
func StreamRouteGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	streamRoute, err := biz.GetStreamRoute(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.StreamRouteOutputInfo{
		AutoID:    streamRoute.AutoID,
		ID:        streamRoute.ID,
		GatewayID: streamRoute.GatewayID,
		StreamRouteInfo: serializer.StreamRouteInfo{
			ID:         streamRoute.ID,
			Name:       streamRoute.Name,
			ServiceID:  streamRoute.ServiceID,
			UpstreamID: streamRoute.UpstreamID,
			Config:     json.RawMessage(streamRoute.Config),
		},
		CreatedAt: streamRoute.CreatedAt.Unix(),
		UpdatedAt: streamRoute.UpdatedAt.Unix(),
		Creator:   streamRoute.Creator,
		Updater:   streamRoute.Updater,
		Status:    streamRoute.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// StreamRouteDelete ...
//
//	@ID			stream_route_delete
//	@Summary	stream_route 删除
//	@Produce	json
//	@Tags		webapi.stream_route
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/stream_routes/{id}/ [delete]
func StreamRouteDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	streamRoute, err := biz.GetStreamRoute(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if streamRoute.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteStreamRoutes(c.Request.Context(), []string{streamRoute.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.StreamRoute, streamRoute.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// StreamRouteList ...
//
//	@ID			stream_route_list
//	@Summary	stream_route 列表
//	@Produce	json
//	@Tags		webapi.stream_route
//	@Param		gateway_id	path		int									true	"网关 ID"
//	@Param		request		query		serializer.StreamRouteListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.StreamRouteListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/stream_routes/ [get]
func StreamRouteList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.StreamRouteListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	labelMap, err := serializer.CheckLabel(req.Label)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	queryParam := map[string]any{}
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	streamRouteList, total, err := biz.ListPagedStreamRoutes(
		c.Request.Context(),
		queryParam,
		labelMap,
		strings.Split(req.Status, ","),
		req.Name,
		req.Updater,
		req.ServiceID,
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
	var results serializer.StreamRouteListResponse
	for _, sr := range streamRouteList {
		results = append(results, serializer.StreamRouteOutputInfo{
			ID:        sr.ID,
			AutoID:    sr.AutoID,
			GatewayID: sr.GatewayID,
			StreamRouteInfo: serializer.StreamRouteInfo{
				ID:         sr.ID,
				Name:       sr.Name,
				ServiceID:  sr.ServiceID,
				UpstreamID: sr.UpstreamID,
				Config:     json.RawMessage(sr.Config),
			},
			Status:    sr.Status,
			CreatedAt: sr.CreatedAt.Unix(),
			UpdatedAt: sr.UpdatedAt.Unix(),
			Creator:   sr.Creator,
			Updater:   sr.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// StreamRouteDropDownList ...
//
//	@ID			stream_route_dropdown_list
//	@Summary	stream_route 下拉列表
//	@Produce	json
//	@Tags		webapi.stream_route
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.StreamRouteDropDownResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/stream_routes-dropdown/ [get]
func StreamRouteDropDownList(c *gin.Context) {
	streamRouteList, err := biz.ListStreamRoutes(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.StreamRouteDropDownResponse
	for _, sr := range streamRouteList {
		desc := gjson.ParseBytes(sr.Config).Get("desc").String()
		output = append(output, serializer.StreamRouteDropDownOutputInfo{
			AutoID: sr.AutoID,
			ID:     sr.ID,
			Name:   sr.Name,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
