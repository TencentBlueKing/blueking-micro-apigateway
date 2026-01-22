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

// Package router 是项目 API 服务的主路由入口
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/basic"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/open"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// New create server router
func New(slogger *slog.Logger) *gin.Engine {
	gin.SetMode(config.G.Service.Server.GinRunMode)
	gin.DisableConsoleColor()

	// 注册校验器
	validation.RegisterValidator()

	router := gin.New()

	_ = router.SetTrustedProxies(nil)

	// middlewares: globally
	// -- recovery sentry
	router.Use(middleware.Recovery())
	// 注意：gin-contrib/cors 需要完整 URL 格式（如 https://example.com）
	corsAllowedOrigins := config.NormalizeOriginsForCORS(config.G.Service.AllowedOrigins)
	router.Use(middleware.CORS(corsAllowedOrigins))
	router.Use(middleware.RequestID())
	// -- trace
	if config.G.Tracing.GinAPIEnabled() {
		// set gin otel
		router.Use(otelgin.Middleware(config.G.Tracing.ServiceName))
	}

	// 基础 API
	basic.Register(router)

	// 用户认证组件
	apiRG := router.Group("/api")
	apiRG.Use(sloggin.NewWithConfig(slogger, sloggin.Config{
		DefaultLevel:      slog.LevelInfo,
		ClientErrorLevel:  slog.LevelWarn,
		ServerErrorLevel:  slog.LevelError,
		WithUserAgent:     false,
		WithRequestID:     true,
		WithRequestHeader: true,
		WithRequestBody:   true,
		WithResponseBody:  true,
		WithSpanID:        true,
		WithTraceID:       true,
	}))
	apiRG.Use(middleware.HandlerAccess())
	// 后端 API
	{
		// 注册web路由
		web.RegisterWebApi("/v1/web", apiRG)
		// 注册openapi路由
		open.RegisterOpenApi("/v1/open", apiRG)
	}

	return router
}
