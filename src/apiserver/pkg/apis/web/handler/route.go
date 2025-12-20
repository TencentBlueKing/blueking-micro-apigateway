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

// RouteCreate ...
//
//	@ID			route_create
//	@Summary	route 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.route
//	@Param		gateway_id	path	int						true	"网关 ID"
//	@Param		request		body	serializer.RouteInfo	true	"route 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/routes/ [post]
func RouteCreate(c *gin.Context) {
	var req serializer.RouteInfo

	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	route := model.Route{
		Name:           req.Name,
		ServiceID:      req.ServiceID,
		UpstreamID:     req.UpstreamID,
		PluginConfigID: req.PluginConfigID,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Route),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.CreateRoute(c.Request.Context(), route); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// RouteUpdate ...
//
//	@ID			route_update
//	@Summary	route 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.route
//	@Param		gateway_id	path	int						true	"网关ID"	@Param	id	path	string	true	"路由ID"
//	@Param		request		body	serializer.RouteInfo	true	"route更新参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/routes/{id}/ [put]
func RouteUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	req := serializer.RouteInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.Route, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	route := model.Route{
		Name:           req.Name,
		ServiceID:      req.ServiceID,
		UpstreamID:     req.UpstreamID,
		PluginConfigID: req.PluginConfigID,
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

	if err := biz.UpdateRoute(c.Request.Context(), route); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, route)
}

// RouteList ...
//
//	@ID			route_list
//	@Summary	route 列表
//	@Produce	json
//	@Tags		webapi.route
//	@Param		gateway_id	path		int							true	"网关 ID"
//	@Param		request		query		serializer.RouteListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.RouteListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/routes/ [get]
func RouteList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.RouteListRequest
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
	routes, total, err := biz.ListPagedRoutes(
		c.Request.Context(),
		queryParam,
		labelMap,
		strings.Split(req.Status, ","),
		req.Name,
		req.Updater,
		req.Path,
		req.Method,
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
	var results serializer.RouteListResponse
	for _, route := range routes {
		results = append(results, serializer.RouteOutputInfo{
			AutoID:    route.AutoID,
			GatewayID: route.GatewayID,
			RouteInfo: serializer.RouteInfo{
				Name:           route.Name,
				ServiceID:      route.ServiceID,
				UpstreamID:     route.UpstreamID,
				PluginConfigID: route.PluginConfigID,
				Config:         json.RawMessage(route.Config),
				ID:             route.ID,
			},
			Status:    route.Status,
			CreatedAt: route.CreatedAt.Unix(),
			UpdatedAt: route.UpdatedAt.Unix(),
			Creator:   route.Creator,
			Updater:   route.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// RouteGet ...
//
//	@ID			route_get
//	@Summary	route 详情
//	@Produce	json
//	@Tags		webapi.route
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"路由 ID"
//	@Success	200			{object}	serializer.RouteOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/routes/{id}/ [get]
func RouteGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	route, err := biz.GetRoute(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	output := serializer.RouteOutputInfo{
		AutoID:    route.AutoID,
		GatewayID: route.GatewayID,
		RouteInfo: serializer.RouteInfo{
			ID:             route.ID,
			Name:           route.Name,
			ServiceID:      route.ServiceID,
			UpstreamID:     route.UpstreamID,
			PluginConfigID: route.PluginConfigID,
			Config:         json.RawMessage(route.Config),
		},
		CreatedAt: route.CreatedAt.Unix(),
		UpdatedAt: route.UpdatedAt.Unix(),
		Creator:   route.Creator,
		Updater:   route.Updater,
		Status:    route.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// RouteDelete ...
//
//	@ID			route_delete
//	@Summary	路由删除
//	@Produce	json
//	@Tags		webapi.route
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"路由 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/routes/{id}/ [delete]
func RouteDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	route, err := biz.GetRoute(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if route.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteRoutes(c.Request.Context(), []string{route.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(
		c.Request.Context(),
		constant.Route,
		route.ID,
		constant.ResourceStatusDeleteDraft,
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// RouteDropDownList ...
//
//	@ID			route_dropdown_list
//	@Summary	route 下拉列表
//	@Produce	json
//	@Tags		webapi.route
//	@Param		gateway_id	path		int	true	"网关 id"
//	@Success	200			{object}	serializer.RouteDropDownOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/routes-dropdown/ [get]
func RouteDropDownList(c *gin.Context) {
	routes, err := biz.ListRoutes(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.RouteDropDownListResponse
	for _, route := range routes {
		desc := gjson.ParseBytes(route.Config).Get("desc").String()
		var uris []string
		for _, u := range gjson.ParseBytes(route.Config).Get("uris").Array() {
			uris = append(uris, u.Str)
		}
		output = append(output, serializer.RouteDropDownOutputInfo{
			ID:     route.ID,
			AutoID: route.AutoID,
			Name:   route.Name,
			Uris:   uris,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
