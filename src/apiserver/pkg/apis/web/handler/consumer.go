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

// ConsumerCreate ...
//
//	@ID			consumer_create
//	@Summary	consumer 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.consumer
//	@Param		gateway_id	path	int	true	"网关 ID"	@Param	request	body	serializer.ConsumerInfo	true	"consumer
//
// 创建参数"
//
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/consumers/ [post]
func ConsumerCreate(c *gin.Context) {
	var req serializer.ConsumerInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	consumer := model.Consumer{
		Username: req.Name,
		GroupID:  req.GroupID,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Consumer), // todo: generate
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.CreateConsumer(c.Request.Context(), consumer); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// ConsumerUpdate ...
//
//	@ID			consumer_update
//	@Summary	consumer 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.consumer
//	@Param		gateway_id	path	int						true	"网关 ID"	@Param	id	path	string	true	"consumerID"
//	@Param		request		body	serializer.ConsumerInfo	true	"consumer 更新参数"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/consumers/{id}/ [put]
func ConsumerUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	req := serializer.ConsumerInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.Consumer, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	consumer := model.Consumer{
		Username: req.Name,
		GroupID:  req.GroupID,
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

	if err := biz.UpdateConsumer(c.Request.Context(), consumer); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ConsumerList ...
//
//	@ID			consumer_list
//	@Summary	consumer 列表
//	@Produce	json
//	@Tags		webapi.consumer
//	@Param		gateway_id	path		int								true	"网关 ID"
//	@Param		request		query		serializer.ConsumerListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.ConsumerListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/consumers/ [get]
func ConsumerList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.ConsumerListRequest
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
	consumers, total, err := biz.ListPagedConsumers(
		c.Request.Context(),
		queryParam,
		labelMap,
		strings.Split(req.Status, ","),
		req.Name,
		req.Updater,
		req.GroupID,
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
	var results serializer.ConsumerListResponse
	for _, consumer := range consumers {
		results = append(results, serializer.ConsumerOutputInfo{
			AutoID:    consumer.AutoID,
			ID:        consumer.ID,
			GatewayID: consumer.GatewayID,
			ConsumerInfo: serializer.ConsumerInfo{
				ID:      consumer.ID,
				Name:    consumer.Username,
				GroupID: consumer.GroupID,
				Config:  json.RawMessage(consumer.Config),
			},
			Status:    consumer.Status,
			CreatedAt: consumer.CreatedAt.Unix(),
			UpdatedAt: consumer.UpdatedAt.Unix(),
			Creator:   consumer.Creator,
			Updater:   consumer.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// ConsumerGet ...
//
//	@ID			consumer_get
//	@Summary	consumer 详情
//	@Produce	json
//	@Tags		webapi.consumer
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.ConsumerOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/consumers/{id}/ [get]
func ConsumerGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	consumer, err := biz.GetConsumer(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.ConsumerOutputInfo{
		GatewayID: consumer.GatewayID,
		ConsumerInfo: serializer.ConsumerInfo{
			ID:      consumer.ID,
			Name:    consumer.Username,
			GroupID: consumer.GroupID,
			Config:  json.RawMessage(consumer.Config),
		},
		CreatedAt: consumer.CreatedAt.Unix(),
		UpdatedAt: consumer.UpdatedAt.Unix(),
		Creator:   consumer.Creator,
		Updater:   consumer.Updater,
		Status:    consumer.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// ConsumerDelete ...
//
//	@ID			consumer_delete
//	@Summary	consumer 删除
//	@Produce	json
//	@Tags		webapi.consumer
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/consumers/{id}/ [delete]
func ConsumerDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	consumer, err := biz.GetConsumer(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if consumer.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteConsumers(c.Request.Context(), []string{consumer.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(
		c.Request.Context(), constant.Consumer, consumer.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ConsumerDropDownList ...
//
//	@ID			consumer_dropdown_list
//	@Summary	consumer 下拉列表
//	@Produce	json
//	@Tags		webapi.consumer
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.ConsumerDropDownListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/consumers-dropdown/ [get]
func ConsumerDropDownList(c *gin.Context) {
	consumers, err := biz.ListConsumers(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.ConsumerDropDownListResponse
	for _, consumer := range consumers {
		desc := gjson.ParseBytes(consumer.Config).Get("desc").String()
		output = append(output, serializer.ConsumerDropDownOutputInfo{
			AutoID: consumer.AutoID,
			ID:     consumer.ID,
			Name:   consumer.Username,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
