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
	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// PublishResource ...
//
//	@ID			resource_publish
//	@Summary	资源发布
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.publish
//	@Param		gateway_id	path	int							true	"网关 ID"
//	@Param		request		body	serializer.PublishRequest	true	"发布资源请求参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/publish/ [post]
func PublishResource(c *gin.Context) {
	var req serializer.PublishRequest
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	err := biz.PublishResource(c.Request.Context(), req.ResourceType, req.ResourceIDList)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// PublishResourceAll ...
//
//	@ID			resource_publish_all
//	@Summary	资源一键发布
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.publish
//	@Param		gateway_id	path	int	true	"网关 ID"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/publish/all/ [post]
func PublishResourceAll(c *gin.Context) {
	err := biz.PublishAllResource(c.Request.Context(), ginx.GetGatewayInfo(c).ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}
