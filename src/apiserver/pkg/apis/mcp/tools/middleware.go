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

package tools

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// GatewayContextMiddleware injects the gateway info into the context for all tool calls.
// This middleware should be registered via Server.AddReceivingMiddleware().
func GatewayContextMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		// Only process tool calls - they need gateway context
		if strings.HasPrefix(method, "tools/call") {
			gateway, err := getGatewayFromContext(ctx)
			if err != nil {
				return nil, err
			}
			ctx = ginx.SetGatewayInfoToContext(ctx, gateway)
		}
		return next(ctx, method, req)
	}
}

// WriteAccessMiddleware checks write scope for write tools.
// This middleware should be registered via Server.AddReceivingMiddleware().
func WriteAccessMiddleware(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		// Only check for tool calls
		if strings.HasPrefix(method, "tools/call") {
			// Try to extract tool name from request
			if callReq, ok := req.(*mcp.CallToolRequest); ok {
				toolName := callReq.Params.Name
				if IsWriteTool(toolName) {
					if err := CheckWriteScope(ctx); err != nil {
						return nil, err
					}
				}
			}
		}
		return next(ctx, method, req)
	}
}
