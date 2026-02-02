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
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// MCPAccessTokenList 列出 MCP 访问令牌
//
//	@ID			mcp_access_token_list
//	@Summary	MCP 访问令牌列表
//	@Produce	json
//	@Tags		webapi.mcp_access_token
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	ginx.Response{data=serializer.MCPAccessTokenListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/mcp/tokens/ [get]
func MCPAccessTokenList(c *gin.Context) {
	var pathParam serializer.MCPAccessTokenPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	// 检查网关是否支持 MCP
	gateway := ginx.GetGatewayInfo(c)
	if err := biz.CheckGatewayMCPSupport(gateway); err != nil {
		ginx.NotImplementedJSONResponse(c, err)
		return
	}

	tokens, err := biz.ListMCPAccessTokens(c.Request.Context(), pathParam.GatewayID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	results := make(serializer.MCPAccessTokenListResponse, 0, len(tokens))
	for _, token := range tokens {
		results = append(results, serializer.MCPAccessTokenToOutputInfo(token))
	}

	ginx.SuccessJSONResponse(c, results)
}

// MCPAccessTokenCreate 创建 MCP 访问令牌
//
//	@ID			mcp_access_token_create
//	@Summary	创建 MCP 访问令牌
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.mcp_access_token
//	@Param		gateway_id	path		int										true	"网关 ID"
//	@Param		request		body		serializer.MCPAccessTokenCreateRequest	true	"创建参数"
//	@Success	201			{object}	ginx.Response{data=serializer.MCPAccessTokenCreateOutputInfo}
//	@Router		/api/v1/web/gateways/{gateway_id}/mcp/tokens/ [post]
func MCPAccessTokenCreate(c *gin.Context) {
	var pathParam serializer.MCPAccessTokenPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	var req serializer.MCPAccessTokenCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	// 检查网关是否支持 MCP
	gateway := ginx.GetGatewayInfo(c)
	if err := biz.CheckGatewayMCPSupport(gateway); err != nil {
		ginx.NotImplementedJSONResponse(c, err)
		return
	}

	// 验证过期时间
	expiredAt := time.Unix(req.ExpiredAt, 0)
	if expiredAt.Before(time.Now()) {
		ginx.BadRequestErrorJSONResponse(c, errors.New("expired_at must be in the future"))
		return
	}

	token := &model.MCPAccessToken{
		GatewayID:   pathParam.GatewayID,
		Name:        req.Name,
		Description: req.Description,
		AccessScope: req.AccessScope,
		ExpiredAt:   expiredAt,
		BaseModel: model.BaseModel{
			Creator: ginx.GetUserID(c),
			Updater: ginx.GetUserID(c),
		},
	}

	if err := biz.CreateMCPAccessToken(c.Request.Context(), token); err != nil {
		if errors.Is(err, biz.ErrMCPTokenNameExists) {
			ginx.ConflictJSONResponse(c, err)
			return
		}
		if errors.Is(err, biz.ErrMCPTokenLimitExceeded) {
			ginx.BadRequestErrorJSONResponse(c, err)
			return
		}
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	// Audit log failure should not fail the response
	_ = biz.AddMCPAccessTokenAuditLog(c.Request.Context(), constant.OperationTypeCreate, token)

	ginx.SuccessCreateJSONResponse(c, serializer.MCPAccessTokenToCreateOutputInfo(token))
}

// MCPAccessTokenGet 获取 MCP 访问令牌详情
//
//	@ID			mcp_access_token_get
//	@Summary	MCP 访问令牌详情
//	@Produce	json
//	@Tags		webapi.mcp_access_token
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Param		token_id	path		int	true	"令牌 ID"
//	@Success	200			{object}	ginx.Response{data=serializer.MCPAccessTokenOutputInfo}
//	@Router		/api/v1/web/gateways/{gateway_id}/mcp/tokens/{token_id}/ [get]
func MCPAccessTokenGet(c *gin.Context) {
	var pathParam serializer.MCPAccessTokenPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	// 检查网关是否支持 MCP
	gateway := ginx.GetGatewayInfo(c)
	if err := biz.CheckGatewayMCPSupport(gateway); err != nil {
		ginx.NotImplementedJSONResponse(c, err)
		return
	}

	token, err := biz.GetMCPAccessTokenByGatewayAndID(c.Request.Context(), pathParam.GatewayID, pathParam.TokenID)
	if err != nil {
		if errors.Is(err, biz.ErrMCPTokenNotFound) {
			ginx.NotFoundJSONResponse(c, err)
			return
		}
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	ginx.SuccessJSONResponse(c, serializer.MCPAccessTokenToOutputInfo(token))
}

// MCPAccessTokenDelete 删除 MCP 访问令牌
//
//	@ID			mcp_access_token_delete
//	@Summary	删除 MCP 访问令牌
//	@Produce	json
//	@Tags		webapi.mcp_access_token
//	@Param		gateway_id	path	int	true	"网关 ID"
//	@Param		token_id	path	int	true	"令牌 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/mcp/tokens/{token_id}/ [delete]
func MCPAccessTokenDelete(c *gin.Context) {
	var pathParam serializer.MCPAccessTokenPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	// 检查网关是否支持 MCP
	gateway := ginx.GetGatewayInfo(c)
	if err := biz.CheckGatewayMCPSupport(gateway); err != nil {
		ginx.NotImplementedJSONResponse(c, err)
		return
	}

	// 检查令牌是否属于该网关
	token, err := biz.GetMCPAccessTokenByGatewayAndID(c.Request.Context(), pathParam.GatewayID, pathParam.TokenID)
	if err != nil {
		if errors.Is(err, biz.ErrMCPTokenNotFound) {
			ginx.NotFoundJSONResponse(c, err)
			return
		}
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	if err := biz.DeleteMCPAccessToken(c.Request.Context(), pathParam.TokenID); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	// Audit log failure should not fail the response
	_ = biz.AddMCPAccessTokenAuditLog(c.Request.Context(), constant.OperationTypeDelete, token)

	ginx.SuccessNoContentResponse(c)
}
