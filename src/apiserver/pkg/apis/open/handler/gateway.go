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
	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/common"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/open/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/stringx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// GatewayCreate ...
//
//	@ID			openapi_gateway_create
//	@Summary	网关创建
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.gateway
//	@Param		X-BK-API-TOKEN	header		string					false	"独立部署时 token 必须要传"
//	@Param		request			body		common.GatewayInputInfo	true	"网关创建参数"
//	@Success	200				{object}	serializer.GatewayCreateResponse
//	@Router		/api/v1/open/gateways/ [post]
func GatewayCreate(c *gin.Context) {
	var req common.GatewayInputInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		log.ErrorFWithContext(c.Request.Context(), "bind and validate failed: %s", err.Error())
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var gatewayID int
	var token string
	// 如果是独立部署模式
	if config.G.Service.Standalone {
		token = c.GetHeader(constant.OpenAPITokenHeaderKey)
	} else {
		token = stringx.RandString(constant.AccessTokenLength)
	}
	// check etcd and apisix instance
	_, instanceID, err := common.CheckEtcdConnAndAPISIXInstance(gatewayID, req.EtcdConfig)
	if err != nil {
		log.ErrorFWithContext(c.Request.Context(), "etcd check failed: %s", err.Error())
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	gatewayInfo := &model.Gateway{
		ID:            gatewayID,
		Name:          req.Name,
		Mode:          constant.GatewayControlModeDirect, // 注册的网关都是直接管理模式
		Maintainers:   req.Maintainers.Strip(),
		Desc:          req.Description,
		APISIXType:    req.APISIXType,
		APISIXVersion: req.APISIXVersion,
		EtcdConfig: model.EtcdConfig{
			InstanceID: instanceID,
			EtcdConfig: base.EtcdConfig{
				Endpoint: req.EtcdEndPoints.EndpointJoin(),
				Username: req.EtcdUsername,
				Password: req.EtcdPassword,
				Prefix:   req.EtcdPrefix,
				CACert:   req.EtcdCACert,
				CertCert: req.EtcdCertCert,
				CertKey:  req.EtcdCertKey,
			},
		},
		Token: token,
		BaseModel: model.BaseModel{
			Creator: ginx.GetUserID(c),
			Updater: ginx.GetUserID(c),
		},
	}
	if err = biz.CreateGateway(c.Request.Context(), gatewayInfo); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, serializer.GatewayCreateResponse{
		ID:    gatewayInfo.ID,
		Token: token,
	})
}

// GatewayGet ...
//
//	@ID			openapi_gateway_get
//	@Summary	网关查询
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.gateway
//	@Param		gateway_name	path		string	true	"网关名"
//	@Param		X-BK-API-TOKEN	header		string	true	"创建网关返回的 token"
//	@Success	200				{object}	common.GatewayOutputInfo
//	@Router		/api/v1/open/gateways/{gateway_name}/ [get]
func GatewayGet(c *gin.Context) {
	ginx.SuccessJSONResponse(c, common.GatewayToOutputInfo(ginx.GetGatewayInfo(c)))
}

// GatewayUpdate ...
//
//	@ID			openapi_gateway_update
//	@Summary	网关更新
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.gateway
//	@Param		gateway_name	path	string					true	"网关名称"
//	@Param		X-BK-API-TOKEN	header	string					true	"创建网关返回的 token"
//	@Param		request			body	common.GatewayInputInfo	true	"网关更新参数"
//	@Success	200
//	@Router		/api/v1/open/gateways/{gateway_name}/ [put]
func GatewayUpdate(c *gin.Context) {
	var req common.GatewayInputInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	if req.EtcdPassword == "" {
		req.EtcdPassword = ginx.GetGatewayInfo(c).EtcdConfig.Password
	}
	if req.EtcdCACert == "" {
		req.EtcdCACert = ginx.GetGatewayInfo(c).EtcdConfig.CACert
	}
	if req.EtcdCertCert == "" {
		req.EtcdCertCert = ginx.GetGatewayInfo(c).EtcdConfig.CertCert
	}
	if req.EtcdCertKey == "" {
		req.EtcdCertKey = ginx.GetGatewayInfo(c).EtcdConfig.CertKey
	}
	_, instanceID, err := common.CheckEtcdConnAndAPISIXInstance(ginx.GetGatewayInfo(c).ID, req.EtcdConfig)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	gatewayInfo := model.Gateway{
		ID:            ginx.GetGatewayInfo(c).ID,
		Name:          req.Name,
		Mode:          req.Mode,
		Maintainers:   req.Maintainers.Strip(),
		Desc:          req.Description,
		APISIXType:    ginx.GetGatewayInfo(c).APISIXType,
		APISIXVersion: ginx.GetGatewayInfo(c).APISIXVersion,
		EtcdConfig: model.EtcdConfig{
			EtcdConfig: base.EtcdConfig{
				Endpoint: req.EtcdEndPoints.EndpointJoin(),
				Username: req.EtcdUsername,
				Password: req.EtcdPassword,
				Prefix:   ginx.GetGatewayInfo(c).EtcdConfig.Prefix, // etcd 前缀保持不变
				CACert:   req.EtcdCACert,
				CertCert: req.EtcdCertCert,
				CertKey:  req.EtcdCertKey,
			},
			InstanceID: instanceID,
		},
		Token: ginx.GetGatewayInfo(c).Token,
		BaseModel: model.BaseModel{
			Updater: ginx.GetUserID(c),
		},
	}
	if err := biz.UpdateGateway(c.Request.Context(), gatewayInfo); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	gatewayInfo.RemoveSensitive()
	ginx.SuccessJSONResponse(c, gatewayInfo)
}

// GatewayDelete ...
//
//	@ID			openapi_gateway_delete
//	@Summary	网关删除
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.gateway
//	@Param		gateway_name	path	string	true	"网关名称"
//	@Param		X-BK-API-TOKEN	header	string	true	"创建网关返回的 token"
//	@Success	204
//	@Router		/api/v1/open/gateways/{gateway_name}/ [delete]
func GatewayDelete(c *gin.Context) {
	err := biz.DeleteGateway(c, ginx.GetGatewayInfo(c))
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// GatewayPublish ...
//
//	@ID			openapi_gateway_publish
//	@Summary	网关一键发布
//	@Accept		json
//	@Produce	json
//	@Tags		openapi.gateway
//	@Param		X-BK-API-TOKEN	header	string	true	"创建网关返回的 token"
//	@Param		gateway_name	path	string	true	"网关名称"
//	@Success	201
//	@Router		/api/v1/open/gateways/{gateway_name}/publish/ [post]
func GatewayPublish(c *gin.Context) {
	err := biz.PublishAllResource(c.Request.Context(), ginx.GetGatewayInfo(c).ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}
