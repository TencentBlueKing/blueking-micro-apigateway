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

// Package basic ...
package basic

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggest/swgui"
	"github.com/swaggest/swgui/v5emb"

	_ "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/docs"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/basic/handler"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
)

// Register ...
func Register(router *gin.Engine) {
	// basic
	// router.GET("/ping", handler.Ping)
	router.GET("/version", handler.Version)

	// healthz
	healthzRouter := router.Group("/healthz")
	healthzRouter.Use(middleware.QueryTokenAuth(config.G.Service.HealthzToken))
	healthzRouter.GET("", handler.Healthz)
	// metrics
	metricRouter := router.Group("/metrics")
	metricRouter.Use(middleware.QueryTokenAuth(config.G.Service.MetricToken))
	metricRouter.GET("", gin.WrapH(promhttp.Handler()))

	if config.G.Service.EnableSwagger {
		router.StaticFile("/swagger.json", config.G.Service.DocFileBaseDir+"/swagger.json")
		// swagger docs（仅推荐开发环境使用）
		hd := v5emb.NewHandlerWithConfig(swgui.Config{
			Title:       "apiserver api doc",
			SwaggerJSON: "/swagger.json",
			BasePath:    "/swagger-ui",
			ShowTopBar:  true,
			HideCurl:    false,
			JsonEditor:  true,
		})
		router.GET("/swagger-ui/*any", gin.WrapH(hd))
	}
}
