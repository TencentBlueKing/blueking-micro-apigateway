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

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// mcpContextKey is a custom type for context keys to avoid collisions
type mcpContextKey string

const (
	// mcpTokenCtxKey is the type-safe context key for standard Go context
	mcpTokenCtxKey mcpContextKey = "mcp_access_token"
	// MCPTokenContextKey is the string key for Gin context (c.Set/c.Get)
	MCPTokenContextKey = "mcp_access_token"
)

// MCPAuth middleware for MCP API authentication using Bearer token
func MCPAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.ErrorFWithContext(c.Request.Context(), "MCP auth: missing Authorization header")
			abortWithMCPError(c, http.StatusUnauthorized, "missing Authorization header")
			return
		}

		// Check for Bearer prefix
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			log.ErrorFWithContext(c.Request.Context(), "MCP auth: invalid Authorization header format")
			abortWithMCPError(
				c,
				http.StatusUnauthorized,
				"invalid Authorization header format, expected 'Bearer <token>'",
			)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, bearerPrefix)
		if tokenStr == "" {
			log.ErrorFWithContext(c.Request.Context(), "MCP auth: empty token")
			abortWithMCPError(c, http.StatusUnauthorized, "empty token")
			return
		}

		// Validate token and get gateway info
		token, gateway, err := biz.ValidateMCPAccessToken(c.Request.Context(), tokenStr)
		if err != nil {
			handleMCPAuthError(c, err)
			return
		}

		// Check access scope based on HTTP method
		requireWrite := isWriteOperation(c.Request.Method)
		if err := biz.CheckMCPAccessScope(token, requireWrite); err != nil {
			log.ErrorFWithContext(
				c.Request.Context(),
				"MCP auth: insufficient scope for %s",
				c.Request.Method,
			)
			abortWithMCPError(
				c,
				http.StatusForbidden,
				"insufficient access scope: write permission required",
			)
			return
		}

		// Set gateway info in context
		ginx.SetGatewayInfo(c, gateway)

		// Set MCP token in context
		c.Set(MCPTokenContextKey, token)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), mcpTokenCtxKey, token))

		// Set validation error info for downstream handlers
		ginx.SetValidateErrorInfo(c)

		c.Next()
	}
}

// handleMCPAuthError handles MCP authentication errors
func handleMCPAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, biz.ErrMCPTokenNotFound):
		log.ErrorFWithContext(c.Request.Context(), "MCP auth: token not found")
		abortWithMCPError(c, http.StatusUnauthorized, "invalid token")
	case errors.Is(err, biz.ErrMCPTokenExpired):
		log.ErrorFWithContext(c.Request.Context(), "MCP auth: token expired")
		abortWithMCPError(c, http.StatusForbidden, "token has expired")
	case errors.Is(err, biz.ErrMCPGatewayNotSupported):
		log.ErrorFWithContext(c.Request.Context(), "MCP auth: gateway does not support MCP")
		abortWithMCPError(c, http.StatusForbidden, "gateway does not support MCP (requires APISIX 3.13.X)")
	default:
		log.ErrorFWithContext(c.Request.Context(), "MCP auth: %v", err)
		abortWithMCPError(c, http.StatusInternalServerError, "authentication failed")
	}
}

// abortWithMCPError aborts the request with an MCP-compatible error response
func abortWithMCPError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, gin.H{
		"error": gin.H{
			"code":    statusCode,
			"message": message,
		},
	})
}

// isWriteOperation checks if the HTTP method is a write operation
func isWriteOperation(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// GetMCPAccessToken retrieves the MCP access token from the context
func GetMCPAccessToken(c *gin.Context) *model.MCPAccessToken {
	if token, exists := c.Get(MCPTokenContextKey); exists {
		if t, ok := token.(*model.MCPAccessToken); ok {
			return t
		}
	}
	return nil
}

// GetMCPAccessTokenFromContext retrieves the MCP access token from a standard context
func GetMCPAccessTokenFromContext(ctx context.Context) *model.MCPAccessToken {
	if token, ok := ctx.Value(mcpTokenCtxKey).(*model.MCPAccessToken); ok {
		return token
	}
	return nil
}

// SetMCPAccessTokenInContext sets the MCP access token in a standard context
// This is primarily used for testing purposes
func SetMCPAccessTokenInContext(ctx context.Context, token *model.MCPAccessToken) context.Context {
	return context.WithValue(ctx, mcpTokenCtxKey, token)
}
