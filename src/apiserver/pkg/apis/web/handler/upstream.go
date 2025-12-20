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

// UpstreamCreate ...
//
//	@ID			upstream_create
//	@Summary	upstream 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.upstream
//	@Param		gateway_id	path	int	true	"网关 ID"	@Param	request	body	serializer.UpstreamInfo	true	"upstream
//
// 创建参数"
//
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/upstreams/ [post]
func UpstreamCreate(c *gin.Context) {
	var req serializer.UpstreamInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	upstream := model.Upstream{
		Name:  req.Name,
		SSLID: req.SSLID,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Upstream),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.CreateUpstream(c.Request.Context(), upstream); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// UpstreamUpdate ...
//
//	@ID			upstream_update
//	@Summary	upstream 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.upstream
//	@Param		gateway_id	path	int						true	"网关ID"	@Param	id	path	string	true	"upstream ID"
//	@Param		request		body	serializer.UpstreamInfo	true	"upstream更新参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/upstreams/{id}/ [put]
func UpstreamUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	req := serializer.UpstreamInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.Upstream, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	upstream := model.Upstream{
		Name:  req.Name,
		SSLID: req.SSLID,
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
	if err := biz.UpdateUpstream(c.Request.Context(), upstream); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// UpstreamList ...
//
//	@ID			upstream_list
//	@Summary	upstream 列表
//	@Produce	json
//	@Tags		webapi.upstream
//	@Param		gateway_id	path		int								true	"网关 ID"
//	@Param		request		query		serializer.UpstreamListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.UpstreamListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/upstreams/ [get]
func UpstreamList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.UpstreamListRequest
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
	upstreams, total, err := biz.ListPagedUpstreams(
		c.Request.Context(),
		queryParam,
		labelMap,
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
	var results serializer.UpstreamListResponse
	for _, upstream := range upstreams {
		results = append(results, serializer.UpstreamOutputInfo{
			AutoID:    upstream.AutoID,
			GatewayID: upstream.GatewayID,
			UpstreamInfo: serializer.UpstreamInfo{
				ID:     upstream.ID,
				Name:   upstream.Name,
				Config: json.RawMessage(upstream.Config),
				SSLID:  upstream.SSLID,
			},
			Status:    upstream.Status,
			CreatedAt: upstream.CreatedAt.Unix(),
			UpdatedAt: upstream.UpdatedAt.Unix(),
			Creator:   upstream.Creator,
			Updater:   upstream.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// UpstreamGet ...
//
//	@ID			upstream_get
//	@Summary	upstream 详情
//	@Produce	json
//	@Tags		webapi.upstream
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.UpstreamOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/upstreams/{id}/ [get]
func UpstreamGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	upstream, err := biz.GetUpstream(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	output := serializer.UpstreamOutputInfo{
		GatewayID: upstream.GatewayID,
		AutoID:    upstream.AutoID,
		UpstreamInfo: serializer.UpstreamInfo{
			ID:     upstream.ID,
			Name:   upstream.Name,
			Config: json.RawMessage(upstream.Config),
			SSLID:  upstream.SSLID,
		},
		CreatedAt: upstream.CreatedAt.Unix(),
		UpdatedAt: upstream.UpdatedAt.Unix(),
		Creator:   upstream.Creator,
		Updater:   upstream.Updater,
		Status:    upstream.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// UpstreamDelete ...
//
//	@ID			upstream_delete
//	@Summary	upstream 删除
//	@Produce	json
//	@Tags		webapi.upstream
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/upstreams/{id}/ [delete]
func UpstreamDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	upstream, err := biz.GetUpstream(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if upstream.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteUpstreams(c.Request.Context(), []string{upstream.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.Upstream, upstream.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// UpstreamDropDownList ...
//
//	@ID			upstream_dropdown_list
//	@Summary	upstream 下拉列表
//	@Produce	json
//	@Tags		webapi.upstream
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.UpstreamDropDownListResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/upstreams-dropdown/ [get]
func UpstreamDropDownList(c *gin.Context) {
	upstreams, err := biz.ListUpstreams(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	var output serializer.UpstreamDropDownListResponse
	for _, upstream := range upstreams {
		desc := gjson.ParseBytes(upstream.Config).Get("desc").String()
		output = append(output, serializer.UpstreamDropDownOutputInfo{
			AutoID: upstream.AutoID,
			ID:     upstream.ID,
			Name:   upstream.Name,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
