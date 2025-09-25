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

// Package open ...
package open

import (
	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/open/handler"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
)

// RegisterOpenApi  注册openapi路由
func RegisterOpenApi(path string, router *gin.RouterGroup) {
	group := router.Group(path)
	// gateway
	gatewayGroup := group.Group("/gateways/")
	gatewayGroup.Use(middleware.OpenAPIAccess())
	gatewayGroup.POST("/", handler.GatewayCreate)
	gatewayGroup.GET("/:gateway_name/", handler.GatewayGet)
	gatewayGroup.PUT("/:gateway_name/", handler.GatewayUpdate)
	gatewayGroup.DELETE("/:gateway_name/", handler.GatewayDelete)
	gatewayGroup.POST("/:gateway_name/publish/", handler.GatewayPublish)
	// resource import
	gatewayGroup.POST("/:gateway_name/resources/-/import/", handler.ResourceImport)

	// resource
	resourceGroup := gatewayGroup.Group("/:gateway_name/resources")
	resourceGroup.Use(middleware.OpenAPIResourceCheck())
	resourceGroup.POST("/:resource_type/", handler.ResourceBatchCreate)
	resourceGroup.POST("/:resource_type/batch_delete", handler.ResourceBatchDelete)
	resourceGroup.GET("/:resource_type/", handler.ResourceBatchGet)
	resourceGroup.GET("/:resource_type/:id/", handler.ResourceGet)
	resourceGroup.GET("/:resource_type/:id/status/", handler.ResourceGetStatus)
	resourceGroup.PUT("/:resource_type/:id/", handler.ResourceUpdate)
	resourceGroup.DELETE("/:resource_type/:id/", handler.ResourceDelete)

	// resource publish
	resourceGroup.POST("/:resource_type/publish/", handler.ResourcePublish)

}
