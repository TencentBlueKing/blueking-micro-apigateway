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

// ProtoCreate ...
//
//	@ID			proto_create
//	@Summary	proto 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.proto
//	@Param		gateway_id	path	int						true	"网关 ID"
//	@Param		request		body	serializer.ProtoInfo	true	"proto 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/protos/ [post]
func ProtoCreate(c *gin.Context) {
	var req serializer.ProtoInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	proto := model.Proto{
		Name: req.Name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Proto),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}
	if err := biz.CreateProto(c.Request.Context(), proto); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// ProtoUpdate ...
//
//	@ID			proto_update
//	@Summary	proto 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.proto
//	@Param		gateway_id	path	int						true	"网关ID"
//	@Param		id			path	string					true	"proto ID"
//	@Param		request		body	serializer.ProtoInfo	true	"proto 更新参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/protos/{id}/ [put]
func ProtoUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	req := serializer.ProtoInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.Proto, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	proto := model.Proto{
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
	if err := biz.UpdateProto(c.Request.Context(), proto); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
}

// ProtoGet ...
//
//	@ID			proto_get
//	@Summary	proto 详情
//	@Produce	json
//	@Tags		webapi.proto
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.ProtoOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/protos/{id}/ [get]
func ProtoGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	proto, err := biz.GetProto(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.ProtoOutputInfo{
		AutoID:    proto.AutoID,
		ID:        proto.ID,
		GatewayID: proto.GatewayID,
		ProtoInfo: serializer.ProtoInfo{
			ID:     proto.ID,
			Name:   proto.Name,
			Config: json.RawMessage(proto.Config),
		},
		CreatedAt: proto.CreatedAt.Unix(),
		UpdatedAt: proto.UpdatedAt.Unix(),
		Creator:   proto.Creator,
		Updater:   proto.Updater,
		Status:    proto.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// ProtoDelete ...
//
//	@ID			proto_delete
//	@Summary	proto 删除
//	@Produce	json
//	@Tags		webapi.proto
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/protos/{id}/ [delete]
func ProtoDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	proto, err := biz.GetProto(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if proto.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteProtos(c.Request.Context(), []string{proto.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.Proto, proto.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ProtoList ...
//
//	@ID			proto_list
//	@Summary	proto 列表
//	@Produce	json
//	@Tags		webapi.proto
//	@Param		gateway_id	path		int							true	"网关 ID"
//	@Param		request		query		serializer.ProtoListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.ProtoListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/protos/ [get]
func ProtoList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.ProtoListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	queryParam := map[string]interface{}{}
	queryParam["gateway_id"] = pathParam.GatewayID
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	protoList, total, err := biz.ListPagedProtos(
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
	var results serializer.ProtoListResponse
	for _, pb := range protoList {
		results = append(results, serializer.ProtoOutputInfo{
			ID:        pb.ID,
			AutoID:    pb.AutoID,
			GatewayID: pb.GatewayID,
			ProtoInfo: serializer.ProtoInfo{
				ID:     pb.ID,
				Name:   pb.Name,
				Config: json.RawMessage(pb.Config),
			},
			Status:    pb.Status,
			CreatedAt: pb.CreatedAt.Unix(),
			UpdatedAt: pb.UpdatedAt.Unix(),
			Creator:   pb.Creator,
			Updater:   pb.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// ProtoDropDownList ...
//
//	@ID			proto_dropdown_list
//	@Summary	proto 下拉列表
//	@Produce	json
//	@Tags		webapi.proto
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.ProtoDropDownResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/protos-dropdown/ [get]
func ProtoDropDownList(c *gin.Context) {
	protos, err := biz.ListProtos(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.ProtoDropDownResponse
	for _, pb := range protos {
		desc := gjson.ParseBytes(pb.Config).Get("desc").String()
		output = append(output, serializer.ProtoDropDownOutputInfo{
			AutoID: pb.AutoID,
			ID:     pb.ID,
			Name:   pb.Name,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
