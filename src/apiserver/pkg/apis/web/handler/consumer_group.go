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

// ConsumerGroupCreate ...
//
//	@ID			consumer_group_create
//	@Summary	consumer_group 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.consumer_group
//	@Param		gateway_id	path	int								true	"网关 ID"
//	@Param		request		body	serializer.ConsumerGroupInfo	true	"consumer_group 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/consumer_groups/ [post]
func ConsumerGroupCreate(c *gin.Context) {
	var req serializer.ConsumerGroupInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	consumerGroup := model.ConsumerGroup{
		Name: req.Name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.ConsumerGroup),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.CreateConsumerGroup(c.Request.Context(), consumerGroup); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// ConsumerGroupUpdate ...
//
//	@ID			consumer_group_update
//	@Summary	consumer_group 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.consumer_group
//	@Param		gateway_id	path	int								true	"网关ID"
//	@Param		id			path	string							true	"consumer_group ID"
//	@Param		request		body	serializer.ConsumerGroupInfo	true	"consumer_group更新参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/consumer_groups/{id}/ [put]
func ConsumerGroupUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	req := serializer.ConsumerGroupInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.ConsumerGroup, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	consumerGroup := model.ConsumerGroup{
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

	if err := biz.UpdateConsumerGroup(c.Request.Context(), consumerGroup); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
}

// ConsumerGroupList ...
//
//	@ID			consumer_group_list
//	@Summary	consumer_group 列表
//	@Produce	json
//	@Tags		webapi.consumer_group
//	@Param		gateway_id	path		int									true	"网关 ID"
//	@Param		request		query		serializer.ConsumerGroupListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.ConsumerGroupListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/consumer_groups/ [get]
func ConsumerGroupList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.ConsumerGroupListRequest
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
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	consumerGroups, total, err := biz.ListPagedConsumerGroups(
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
	var results serializer.ConsumerGroupListResponse
	for _, consumerGroup := range consumerGroups {
		results = append(results, serializer.ConsumerGroupOutputInfo{
			AutoID:    consumerGroup.AutoID,
			ID:        consumerGroup.ID,
			GatewayID: consumerGroup.GatewayID,
			ConsumerGroupInfo: serializer.ConsumerGroupInfo{
				ID:     consumerGroup.ID,
				Name:   consumerGroup.Name,
				Config: json.RawMessage(consumerGroup.Config),
			},
			Status:    consumerGroup.Status,
			CreatedAt: consumerGroup.CreatedAt.Unix(),
			UpdatedAt: consumerGroup.UpdatedAt.Unix(),
			Creator:   consumerGroup.Creator,
			Updater:   consumerGroup.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// ConsumerGroupGet ...
//
//	@ID			consumer_group_get
//	@Summary	consumer_group 详情
//	@Produce	json
//	@Tags		webapi.consumer_group
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.ConsumerGroupOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/consumer_groups/{id}/ [get]
func ConsumerGroupGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	consumerGroup, err := biz.GetConsumerGroup(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.ConsumerGroupOutputInfo{
		GatewayID: consumerGroup.GatewayID,
		ConsumerGroupInfo: serializer.ConsumerGroupInfo{
			ID:     consumerGroup.ID,
			Name:   consumerGroup.Name,
			Config: json.RawMessage(consumerGroup.Config),
		},
		CreatedAt: consumerGroup.CreatedAt.Unix(),
		UpdatedAt: consumerGroup.UpdatedAt.Unix(),
		Creator:   consumerGroup.Creator,
		Updater:   consumerGroup.Updater,
		Status:    consumerGroup.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// ConsumerGroupDelete ...
//
//	@ID			consumer_group_delete
//	@Summary	consumer_group 删除
//	@Produce	json
//	@Tags		webapi.consumer_group
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/consumer_groups/{id}/ [delete]
func ConsumerGroupDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	consumerGroup, err := biz.GetConsumerGroup(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if consumerGroup.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteConsumerGroups(c.Request.Context(), []string{consumerGroup.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.ConsumerGroup, consumerGroup.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// ConsumerGroupDropDownList ...
//
//	@ID			consumer_group_dropdown_list
//	@Summary	consumer_group 下拉列表
//	@Produce	json
//	@Tags		webapi.consumer_group
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.ConsumerGroupDropDownListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/consumer_groups-dropdown/ [get]
func ConsumerGroupDropDownList(c *gin.Context) {
	consumerGroups, err := biz.ListConsumerGroups(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.ConsumerGroupDropDownListResponse
	for _, consumerGroup := range consumerGroups {
		desc := gjson.ParseBytes(consumerGroup.Config).Get("desc").String()
		output = append(output, serializer.ConsumerGroupDropDownOutputInfo{
			AutoID: consumerGroup.AutoID,
			ID:     consumerGroup.ID,
			Name:   consumerGroup.Name,
			Desc:   desc,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
