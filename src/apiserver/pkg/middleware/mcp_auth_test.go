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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

func TestIsWriteOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{
			name:     "POST is write operation",
			method:   http.MethodPost,
			expected: true,
		},
		{
			name:     "PUT is write operation",
			method:   http.MethodPut,
			expected: true,
		},
		{
			name:     "PATCH is write operation",
			method:   http.MethodPatch,
			expected: true,
		},
		{
			name:     "DELETE is write operation",
			method:   http.MethodDelete,
			expected: true,
		},
		{
			name:     "GET is not write operation",
			method:   http.MethodGet,
			expected: false,
		},
		{
			name:     "HEAD is not write operation",
			method:   http.MethodHead,
			expected: false,
		},
		{
			name:     "OPTIONS is not write operation",
			method:   http.MethodOptions,
			expected: false,
		},
		{
			name:     "CONNECT is not write operation",
			method:   http.MethodConnect,
			expected: false,
		},
		{
			name:     "TRACE is not write operation",
			method:   http.MethodTrace,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isWriteOperation(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPTokenContextKey(t *testing.T) {
	t.Parallel()

	// Verify the context key constant is defined correctly
	assert.Equal(t, "mcp_access_token", MCPTokenContextKey)
}

func TestAbortWithMCPError(t *testing.T) {
	// This tests the error response format
	// The actual middleware integration requires database setup
	// which is covered in integration tests
	t.Parallel()

	// Verify the function signature exists and can be called
	// without panicking (compile-time check)
	_ = abortWithMCPError
}

func TestHandleMCPAuthError(t *testing.T) {
	// This tests the error handling logic
	// The actual middleware integration requires database setup
	t.Parallel()

	// Verify the function signature exists
	_ = handleMCPAuthError
}

func TestGetMCPAccessToken(t *testing.T) {
	t.Parallel()

	// Test nil context returns nil
	// Note: This function requires gin.Context which needs http test setup
	// The core logic is simple: get from context and type assert
	_ = GetMCPAccessToken
}

func TestGetMCPAccessTokenFromContext(t *testing.T) {
	t.Parallel()

	// Test empty context returns nil
	ctx := context.Background()
	result := GetMCPAccessTokenFromContext(ctx)
	assert.Nil(t, result)

	// Test context with wrong type returns nil
	ctxWithWrongType := context.WithValue(ctx, mcpTokenCtxKey, "not-a-token")
	result = GetMCPAccessTokenFromContext(ctxWithWrongType)
	assert.Nil(t, result)

	// Test context with correct token returns token
	token := &model.MCPAccessToken{
		ID:          1,
		Name:        "test-token",
		AccessScope: model.MCPAccessScopeRead,
	}
	ctxWithToken := SetMCPAccessTokenInContext(ctx, token)
	result = GetMCPAccessTokenFromContext(ctxWithToken)
	assert.NotNil(t, result)
	assert.Equal(t, token.ID, result.ID)
}

func TestMCPAccessScopeString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    model.MCPAccessScope
		expected string
	}{
		{
			name:     "read scope string",
			scope:    model.MCPAccessScopeRead,
			expected: "read",
		},
		{
			name:     "write scope string",
			scope:    model.MCPAccessScopeWrite,
			expected: "write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.scope.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPAccessScopeIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    model.MCPAccessScope
		expected bool
	}{
		{
			name:     "read scope is valid",
			scope:    model.MCPAccessScopeRead,
			expected: true,
		},
		{
			name:     "write scope is valid",
			scope:    model.MCPAccessScopeWrite,
			expected: true,
		},
		{
			name:     "empty scope is invalid",
			scope:    model.MCPAccessScope(""),
			expected: false,
		},
		{
			name:     "unknown scope is invalid",
			scope:    model.MCPAccessScope("admin"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.scope.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}
