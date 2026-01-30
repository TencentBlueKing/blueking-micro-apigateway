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

package mcp

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
)

var (
	mcpServer  *mcp.Server
	mcpHandler http.Handler
)

// RegisterMCPApi registers the MCP API routes
func RegisterMCPApi(path string, router *gin.RouterGroup) {
	// Initialize MCP server
	logger := GetMCPLogger()
	mcpServer = NewMCPServer(logger)

	// Create StreamableHTTP handler
	mcpHandler = mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return mcpServer
		},
		&mcp.StreamableHTTPOptions{
			Logger: logger,
		},
	)

	// Create SSE handler for backward compatibility
	sseHandler := mcp.NewSSEHandler(
		func(r *http.Request) *mcp.Server {
			return mcpServer
		},
		nil,
	)

	// Register routes with gateway_id in path
	// Format: /api/v1/mcp/gateways/:gateway_id/
	group := router.Group(path + "/gateways/:gateway_id")

	// MCP auth middleware (Bearer token + gateway_id validation)
	group.Use(middleware.MCPAuth())

	// StreamableHTTP endpoint (primary)
	group.Any("/", wrapHTTPHandler(mcpHandler))

	// SSE endpoint for backward compatibility
	group.Any("/sse", wrapHTTPHandler(sseHandler))
	group.Any("/sse/*path", wrapHTTPHandler(sseHandler))

	logging.Infof("MCP API registered at %s/gateways/:gateway_id", path)
}

// wrapHTTPHandler wraps an http.Handler to work with gin
func wrapHTTPHandler(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GetMCPLogger returns the MCP-specific logger
func GetMCPLogger() *slog.Logger {
	return logging.GetLogger("mcp")
}
