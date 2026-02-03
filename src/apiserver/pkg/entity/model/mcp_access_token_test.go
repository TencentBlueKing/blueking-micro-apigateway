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

package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMCPAccessToken_TableName(t *testing.T) {
	t.Parallel()

	token := MCPAccessToken{}
	assert.Equal(t, "mcp_access_token", token.TableName())
}

func TestMCPAccessToken_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expiredAt time.Time
		expected  bool
	}{
		{
			name:      "not expired - future time",
			expiredAt: time.Now().Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "expired - past time",
			expiredAt: time.Now().Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "expired - just past",
			expiredAt: time.Now().Add(-1 * time.Second),
			expected:  true,
		},
		{
			name:      "not expired - far future",
			expiredAt: time.Now().Add(365 * 24 * time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token := &MCPAccessToken{ExpiredAt: tt.expiredAt}
			assert.Equal(t, tt.expected, token.IsExpired())
		})
	}
}

func TestMCPAccessToken_CanRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    MCPAccessScope
		expected bool
	}{
		{
			name:     "read scope can read",
			scope:    MCPAccessScopeRead,
			expected: true,
		},
		{
			name:     "write scope can read",
			scope:    MCPAccessScopeReadWrite,
			expected: true,
		},
		{
			name:     "empty scope cannot read",
			scope:    MCPAccessScope(""),
			expected: false,
		},
		{
			name:     "invalid scope cannot read",
			scope:    MCPAccessScope("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token := &MCPAccessToken{AccessScope: tt.scope}
			assert.Equal(t, tt.expected, token.CanRead())
		})
	}
}

func TestMCPAccessToken_CanWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    MCPAccessScope
		expected bool
	}{
		{
			name:     "read scope cannot write",
			scope:    MCPAccessScopeRead,
			expected: false,
		},
		{
			name:     "write scope can write",
			scope:    MCPAccessScopeReadWrite,
			expected: true,
		},
		{
			name:     "empty scope cannot write",
			scope:    MCPAccessScope(""),
			expected: false,
		},
		{
			name:     "invalid scope cannot write",
			scope:    MCPAccessScope("invalid"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token := &MCPAccessToken{AccessScope: tt.scope}
			assert.Equal(t, tt.expected, token.CanWrite())
		})
	}
}

func TestMCPAccessToken_MaskedToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "normal length token",
			token:    "abcdefgh12345678ijklmnop90qrstuv",
			expected: "abcdefgh****stuv",
		},
		{
			name:     "64 char token (standard)",
			token:    "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2",
			expected: "a1b2c3d4****e1f2",
		},
		{
			name:     "short token",
			token:    "short",
			expected: "****",
		},
		{
			name:     "exactly 12 chars",
			token:    "123456789012",
			expected: "****",
		},
		{
			name:     "13 chars (just over threshold)",
			token:    "1234567890123",
			expected: "12345678****0123",
		},
		{
			name:     "empty token",
			token:    "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token := &MCPAccessToken{Token: tt.token}
			assert.Equal(t, tt.expected, token.MaskedToken())
		})
	}
}

func TestMCPAccessToken_UpdateLastUsed(t *testing.T) {
	t.Parallel()

	token := &MCPAccessToken{}
	assert.Nil(t, token.LastUsedAt)

	beforeUpdate := time.Now()
	token.UpdateLastUsed()
	afterUpdate := time.Now()

	assert.NotNil(t, token.LastUsedAt)
	assert.True(t, token.LastUsedAt.After(beforeUpdate) || token.LastUsedAt.Equal(beforeUpdate))
	assert.True(t, token.LastUsedAt.Before(afterUpdate) || token.LastUsedAt.Equal(afterUpdate))
}

func TestMCPAccessScope_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    MCPAccessScope
		expected string
	}{
		{
			name:     "read scope",
			scope:    MCPAccessScopeRead,
			expected: "read",
		},
		{
			name:     "readwrite scope",
			scope:    MCPAccessScopeReadWrite,
			expected: "readwrite",
		},
		{
			name:     "custom scope",
			scope:    MCPAccessScope("custom"),
			expected: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.scope.String())
		})
	}
}

func TestMCPAccessScope_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    MCPAccessScope
		expected bool
	}{
		{
			name:     "read is valid",
			scope:    MCPAccessScopeRead,
			expected: true,
		},
		{
			name:     "write is valid",
			scope:    MCPAccessScopeReadWrite,
			expected: true,
		},
		{
			name:     "empty is invalid",
			scope:    MCPAccessScope(""),
			expected: false,
		},
		{
			name:     "admin is invalid",
			scope:    MCPAccessScope("admin"),
			expected: false,
		},
		{
			name:     "READ (uppercase) is invalid",
			scope:    MCPAccessScope("READ"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.scope.IsValid())
		})
	}
}

func TestMCPAccessScopeConstants(t *testing.T) {
	t.Parallel()

	// Verify the scope constants are defined correctly
	assert.Equal(t, MCPAccessScope("read"), MCPAccessScopeRead)
	assert.Equal(t, MCPAccessScope("readwrite"), MCPAccessScopeReadWrite)
}
