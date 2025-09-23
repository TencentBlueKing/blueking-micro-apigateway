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
	"github.com/gookit/goutil/arrutil"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/common"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// GatewayCreate ...
//
//	@ID			gateway_create
//	@Summary	网关创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.gateway
//	@Param		request	body	common.GatewayInputInfo	true	"网关创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/ [post]
func GatewayCreate(c *gin.Context) {
	var req common.GatewayInputInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	_, instanceID, err := common.CheckEtcdConnAndAPISIXInstance(0, req.EtcdConfig)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	// 处理 maintainer
	if !arrutil.Contains(req.Maintainers, ginx.GetUserID(c)) {
		req.Maintainers = append(req.Maintainers, ginx.GetUserID(c))
	}
	// FIXME:  common.GatewayInputInfo -> model.Gateway
	gateway := model.Gateway{
		Name:          req.Name,
		Mode:          req.Mode,
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
		ReadOnly: req.ReadOnly,
		BaseModel: model.BaseModel{
			Creator: ginx.GetUserID(c),
			Updater: ginx.GetUserID(c),
		},
	}

	if err := biz.CreateGateway(c.Request.Context(), &gateway); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// GatewayUpdate ...
//
//	@ID			gateway_update
//	@Summary	网关更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.gateway
//	@Param		gateway_id	path	int						true	"网关 ID"
//	@Param		request		body	common.GatewayInputInfo	true	"网关更新参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/ [put]
func GatewayUpdate(c *gin.Context) {
	var req common.GatewayInputInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	// FIXME: req.FillDefault() => for unittest
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
	// 处理 maintainer
	if !arrutil.Contains(req.Maintainers, ginx.GetUserID(c)) {
		req.Maintainers = append(req.Maintainers, ginx.GetUserID(c))
	}

	// FIXME:  serializer.GatewayInfo -> model.Gateway
	gateway := model.Gateway{
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
		ReadOnly: req.ReadOnly,
		BaseModel: model.BaseModel{
			Updater: ginx.GetUserID(c),
		},
	}
	if err := biz.UpdateGateway(c.Request.Context(), gateway); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	gateway.RemoveSensitive()
	ginx.SuccessJSONResponse(c, gateway)
}

// GatewayList ...
//
//	@ID			gateway_list
//	@Summary	网关列表
//	@Produce	json
//	@Tags		webapi.gateway
//	@Param		request	query		serializer.GatewayListRequest	false	"查询参数"
//	@Success	200		{object}	serializer.GatewayListResponse
//	@Router		/api/v1/web/gateways/ [get]
func GatewayList(c *gin.Context) {
	var req serializer.GatewayListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	gateways, err := biz.ListGateways(c.Request.Context(), req.Mode)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	var output serializer.GatewayListResponse
	for _, gateway := range gateways {
		// 校验权限 todo: 需要优化直接在数据库层过滤
		if !gateway.HasPermission(ginx.GetUserID(c)) {
			continue
		}
		routeCount, err := biz.GetRouteCount(c, gateway.ID)
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		serviceCount, err := biz.GetServiceCount(c, gateway.ID)
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		upstreamCount, err := biz.GetUpstreamCount(c, gateway.ID)
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		output = append(output, serializer.GatewayOutputListInfo{
			ID:          gateway.ID,
			Name:        gateway.Name,
			Mode:        gateway.Mode,
			Maintainers: gateway.Maintainers,
			Description: gateway.Desc,
			APISIX: common.APISIX{
				Version: gateway.APISIXVersion,
				Type:    gateway.APISIXType,
			},
			ReadOnly: gateway.ReadOnly,
			Etcd: common.Etcd{
				InstanceID: gateway.EtcdConfig.InstanceID,
				EndPoints:  gateway.EtcdConfig.Endpoint.Endpoints(),
				Prefix:     gateway.EtcdConfig.Prefix,
			},
			Count: serializer.Count{
				Route:    routeCount,
				Service:  serviceCount,
				Upstream: upstreamCount,
			},
			CreatedAt: gateway.CreatedAt.Unix(),
			UpdatedAt: gateway.UpdatedAt.Unix(),
			Creator:   gateway.Creator,
			Updater:   gateway.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, output)
}

// GatewayGet ...
//
//	@ID			gateway_get
//	@Summary	网关详情
//	@Produce	json
//	@Tags		webapi.gateway
//	@Param		gateway_id	path		int	true	"网关 id"
//	@Success	200			{object}	common.GatewayOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/ [get]
func GatewayGet(c *gin.Context) {
	var req serializer.GatewayGetRequest
	if err := c.ShouldBindUri(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	gateway, err := biz.GetGateway(c.Request.Context(), req.GatewayID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, common.GatewayToOutputInfo(gateway))
}

// GatewayDelete ...
//
//	@ID			gateway_delete
//	@Summary	网关删除
//	@Produce	json
//	@Tags		webapi.gateway
//	@Param		gateway_id	path	int	true	"网关 id"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/ [delete]
func GatewayDelete(c *gin.Context) {
	var req serializer.GatewayGetRequest
	if err := c.ShouldBindUri(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	gateway, err := biz.GetGateway(c.Request.Context(), req.GatewayID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	err = biz.DeleteGateway(c, gateway)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// GatewayCheckName ...
//
//	@ID			gateway_check_name
//	@Summary	网关重名检测
//	@Param		request	body	serializer.CheckGatewayNameRequest	true	"查询参数"
//	@Tags		webapi.gateway
//	@Success	200	{object}	serializer.CheckGatewayNameResponse
//	@Router		/api/v1/web/gateways/check_name/ [post]
func GatewayCheckName(c *gin.Context) {
	var req serializer.CheckGatewayNameRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	if !biz.ExistsGatewayName(c.Request.Context(), req.Name, req.ID) {
		output := serializer.CheckGatewayNameResponse{
			Status: "error",
		}
		ginx.SuccessJSONResponse(c, output)
		return
	}
	output := serializer.CheckGatewayNameResponse{
		Status: "ok",
	}
	ginx.SuccessJSONResponse(c, output)
}

// EtcdTestConnection ...
//
//	@ID			etcd_test_connection
//	@Summary	etcd 连通性测试
//	@Param		request	body	serializer.EtcdTestConnectionRequest	true	"etcd 配置连通性测试"
//	@Tags		webapi.gateway
//	@Success	200	{object}	serializer.EtcdTestConOutputInfo	"连通性测试结果信息"
//	@Router		/api/v1/web/gateways/etcd/test_connection/ [post]
func EtcdTestConnection(c *gin.Context) {
	var req serializer.EtcdTestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	if req.GatewayID != 0 {
		gateway, err := biz.GetGateway(c.Request.Context(), req.GatewayID)
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
		}
		// 如果输入密码为脱敏信息，则替换为已保存的密码进行连通性测试
		if req.EtcdSchemaType == constant.HTTP {
			if req.EtcdPassword == constant.SensitiveInfoFiledDisplay {
				req.EtcdPassword = gateway.EtcdConfig.Password
			}
		} else {
			if req.EtcdCACert == gateway.EtcdConfig.GetMaskCaCert() {
				req.EtcdCACert = gateway.EtcdConfig.CACert
			}
			if req.EtcdCertCert == gateway.EtcdConfig.GetMaskCertCert() {
				req.EtcdCertCert = gateway.EtcdConfig.CertCert
			}
			if req.EtcdCertKey == gateway.EtcdConfig.GetMaskCertKey() {
				req.EtcdCertKey = gateway.EtcdConfig.CertKey
			}
		}
	}
	apisixVersion, _, err := common.CheckEtcdConnAndAPISIXInstance(req.GatewayID, req.EtcdConfig)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	output := serializer.EtcdTestConOutputInfo{
		APISIXVersion: apisixVersion,
	}
	ginx.SuccessJSONResponse(c, output)
}
