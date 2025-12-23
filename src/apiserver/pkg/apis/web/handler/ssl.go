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

package handler

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
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

// SSLCheck  SSL 证书 check ...
//
//	@ID			ssl_check
//	@Summary	证书 check
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.ssl
//	@Param		gateway_id	path	int	true	"网关 ID"	@Param	request	body	serializer.SSLCheckRequest	true	"ssl
//
// check 请求参数"
//
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/ssls/check [post]
func SSLCheck(c *gin.Context) {
	var req serializer.SSLCheckRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	sslInfo, err := biz.ParseCert(c.Request.Context(), req.Name, req.Cert, req.Key)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, serializer.SSLCheckResponse{
		Name: req.Name,
		SSL:  *sslInfo,
	})
}

// SSLCreate  SSL 证书创建 ...
//
//	@ID			ssl_create
//	@Summary	证书 create
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.ssl
//	@Param		gateway_id	path	int	true	"网关 ID"	@Param	request	body	serializer.SSLInfo	true	"ssl
//
// create 请求参数"
//
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/ssls/ [post]
func SSLCreate(c *gin.Context) {
	var req serializer.SSLInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	// 再次 check
	sslEntity, err := req.ToEntity()
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	sslInfo, err := biz.ParseCert(c.Request.Context(), sslEntity.Name, sslEntity.Cert, sslEntity.Key)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	parseConfig, _ := json.Marshal(sslInfo)
	req.Config, err = jsonx.MergeJson(parseConfig, req.Config)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	err = biz.CreateSSL(c, &model.SSL{
		Name: req.Name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.SSL),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	})
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// SSLUpdate ...
//
//	@ID			ssl_update
//	@Summary	ssl 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.ssl
//	@Param		gateway_id	path	int					true	"网关 ID"	@Param	id	path	string	true	"SSL ID"
//	@Param		request		body	serializer.SSLInfo	true	"SSL 更新参数"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/ssls/{id}/ [put]
func SSLUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	req := serializer.SSLInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	// 再次 check
	sslEntity, err := req.ToEntity()
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	sslInfo, err := biz.ParseCert(c.Request.Context(), sslEntity.Name, sslEntity.Cert, sslEntity.Key)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	parseConfig, _ := json.Marshal(sslInfo)
	req.Config, err = jsonx.MergeJson(parseConfig, req.Config)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.SSL, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	sslModel := &model.SSL{
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
	if err := biz.UpdateSSL(c.Request.Context(), sslModel); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// SSLList ...
//
//	@ID			ssl_list
//	@Summary	ssl 列表
//	@Produce	json
//	@Tags		webapi.ssl
//	@Param		gateway_id	path		int							true	"网关 ID"
//	@Param		request		query		serializer.SSLListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.SSLListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/ssls/ [get]
func SSLList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.SSLListRequest
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
	queryParam["gateway_id"] = pathParam.GatewayID
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	ssls, total, err := biz.ListPagedSSL(
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
	var results serializer.SSLListResponse
	for _, ssl := range ssls {
		results = append(results, serializer.SSLOutputInfo{
			AutoID:    ssl.AutoID,
			GatewayID: ssl.GatewayID,
			ID:        ssl.ID,
			SSLInfo: serializer.SSLInfo{
				ID:     ssl.ID,
				Name:   ssl.Name,
				Config: json.RawMessage(ssl.Config),
			},
			Status:    ssl.Status,
			CreatedAt: ssl.CreatedAt.Unix(),
			UpdatedAt: ssl.UpdatedAt.Unix(),
			Creator:   ssl.Creator,
			Updater:   ssl.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// SSLGet ...
//
//	@ID			ssl_get
//	@Summary	ssl 详情
//	@Produce	json
//	@Tags		webapi.ssl
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.SSLOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/ssls/{id}/ [get]
func SSLGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	ssl, err := biz.GetSSL(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.SSLOutputInfo{
		GatewayID: ssl.GatewayID,
		AutoID:    ssl.AutoID,
		ID:        ssl.ID,
		SSLInfo: serializer.SSLInfo{
			ID:     ssl.ID,
			Name:   ssl.Name,
			Config: json.RawMessage(ssl.Config),
		},
		CreatedAt: ssl.CreatedAt.Unix(),
		UpdatedAt: ssl.UpdatedAt.Unix(),
		Creator:   ssl.Creator,
		Updater:   ssl.Updater,
		Status:    ssl.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// SSLDelete ...
//
//	@ID			ssl_delete
//	@Summary	ssl 删除
//	@Produce	json
//	@Tags		webapi.ssl
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/ssls/{id}/ [delete]
func SSLDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	ssl, err := biz.GetSSL(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if ssl.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteSSL(c.Request.Context(), []string{ssl.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.SSL, ssl.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// SSLDropDownList ...
//
//	@ID			ssl_dropdown_list
//	@Summary	ssl 下拉列表
//	@Produce	json
//	@Tags		webapi.ssl
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.SSLDropDownListResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/ssls-dropdown/ [get]
func SSLDropDownList(c *gin.Context) {
	ssls, err := biz.ListSSL(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	var output serializer.SSLDropDownListResponse
	for _, ssl := range ssls {
		output = append(output, serializer.SSLDropDownOutputInfo{
			AutoID: ssl.AutoID,
			ID:     ssl.ID,
			Name:   ssl.Name,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}
