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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeOriginsForCORS(t *testing.T) {
	tests := []struct {
		name     string
		origins  []string
		expected []string
	}{
		{
			name:     "wildcard",
			origins:  []string{"*"},
			expected: []string{"*"},
		},
		{
			name:     "full URLs with https",
			origins:  []string{"https://example.com", "https://api.example.org"},
			expected: []string{"https://example.com", "https://api.example.org"},
		},
		{
			name:     "full URLs with http",
			origins:  []string{"http://localhost:8080", "http://127.0.0.1:3000"},
			expected: []string{"http://localhost:8080", "http://127.0.0.1:3000"},
		},
		{
			name:     "host-only format (auto add https)",
			origins:  []string{"example.com", "api.example.org:8080"},
			expected: []string{"https://example.com", "https://api.example.org:8080"},
		},
		{
			name:     "mixed formats",
			origins:  []string{"https://example.com", "localhost:8080", "*"},
			expected: []string{"https://example.com", "https://localhost:8080", "*"},
		},
		{
			name:     "empty list",
			origins:  []string{},
			expected: []string{},
		},
		{
			name:     "with whitespace",
			origins:  []string{" https://example.com ", " api.example.org "},
			expected: []string{"https://example.com", "https://api.example.org"},
		},
		{
			name:     "nil input",
			origins:  nil,
			expected: []string{},
		},
		{
			name:     "empty string in list",
			origins:  []string{"https://example.com", "", "api.example.org"},
			expected: []string{"https://example.com", "https://api.example.org"},
		},
		{
			name:     "only whitespace string",
			origins:  []string{"   ", "https://example.com"},
			expected: []string{"https://example.com"},
		},
		{
			name:     "URL with path (should keep as is)",
			origins:  []string{"https://example.com/path"},
			expected: []string{"https://example.com/path"},
		},
		{
			name:     "IPv6 address",
			origins:  []string{"https://[::1]:8080", "[::1]:8080"},
			expected: []string{"https://[::1]:8080", "https://[::1]:8080"},
		},
		{
			name: "real world case - bktencent",
			origins: []string{
				"https://dev-t.paas3-dev.bktencent.com:8888",
				"https://bk-micro-web.paas3-dev.bktencent.com",
			},
			expected: []string{
				"https://dev-t.paas3-dev.bktencent.com:8888",
				"https://bk-micro-web.paas3-dev.bktencent.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeOriginsForCORS(tt.origins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractHostsForCSRF(t *testing.T) {
	tests := []struct {
		name     string
		origins  []string
		expected []string
	}{
		{
			name:     "wildcard",
			origins:  []string{"*"},
			expected: []string{"*"},
		},
		{
			name:     "full URLs with https",
			origins:  []string{"https://example.com", "https://api.example.org"},
			expected: []string{"example.com", "api.example.org"},
		},
		{
			name:     "full URLs with http",
			origins:  []string{"http://localhost:8080", "http://127.0.0.1:3000"},
			expected: []string{"localhost:8080", "127.0.0.1:3000"},
		},
		{
			name: "full URLs with port",
			origins: []string{
				"https://dev-t.paas3-dev.bktencent.com:8888",
				"https://bk-micro-web.paas3-dev.bktencent.com",
			},
			expected: []string{
				"dev-t.paas3-dev.bktencent.com:8888",
				"bk-micro-web.paas3-dev.bktencent.com",
			},
		},
		{
			name:     "host-only format (keep as is)",
			origins:  []string{"example.com", "api.example.org:8080"},
			expected: []string{"example.com", "api.example.org:8080"},
		},
		{
			name:     "mixed formats",
			origins:  []string{"https://example.com", "localhost:8080", "*"},
			expected: []string{"example.com", "localhost:8080", "*"},
		},
		{
			name:     "empty list",
			origins:  []string{},
			expected: []string{},
		},
		{
			name:     "with whitespace",
			origins:  []string{" https://example.com ", " api.example.org "},
			expected: []string{"example.com", "api.example.org"},
		},
		{
			name:     "nil input",
			origins:  nil,
			expected: []string{},
		},
		{
			name:     "empty string in list",
			origins:  []string{"https://example.com", "", "api.example.org"},
			expected: []string{"example.com", "api.example.org"},
		},
		{
			name:     "only whitespace string",
			origins:  []string{"   ", "https://example.com"},
			expected: []string{"example.com"},
		},
		{
			name:     "URL with path (should extract host only)",
			origins:  []string{"https://example.com/path"},
			expected: []string{"example.com"},
		},
		{
			name:     "IPv6 address with scheme",
			origins:  []string{"https://[::1]:8080"},
			expected: []string{"[::1]:8080"},
		},
		{
			name:     "IPv6 address without scheme",
			origins:  []string{"[::1]:8080"},
			expected: []string{"[::1]:8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractHostsForCSRF(tt.origins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOriginNormalizationConsistency 测试 CORS 和 CSRF 处理的一致性
// 确保同一个输入经过两个函数处理后，CSRF 的结果是 CORS 结果的主机名部分
func TestOriginNormalizationConsistency(t *testing.T) {
	testCases := []struct {
		name    string
		origins []string
	}{
		{
			name: "real world config",
			origins: []string{
				"https://dev-t.paas3-dev.bktencent.com:8888",
				"https://bk-micro-web.paas3-dev.bktencent.com",
			},
		},
		{
			name:    "mixed formats",
			origins: []string{"https://example.com", "api.example.org:8080", "*"},
		},
		{
			name:    "localhost development",
			origins: []string{"http://localhost:3000", "http://127.0.0.1:8080"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			corsOrigins := NormalizeOriginsForCORS(tc.origins)
			csrfOrigins := ExtractHostsForCSRF(tc.origins)

			// 两个结果应该长度相同
			assert.Equal(
				t,
				len(corsOrigins),
				len(csrfOrigins),
				"CORS and CSRF origins should have same length",
			)

			// 对于非 * 的情况，CORS origin 应该包含对应的 CSRF origin
			for i, corsOrigin := range corsOrigins {
				csrfOrigin := csrfOrigins[i]
				if corsOrigin == "*" {
					assert.Equal(t, "*", csrfOrigin, "wildcard should be preserved")
				} else {
					// CORS origin 应该以 http:// 或 https:// 开头
					assert.True(t,
						strings.HasPrefix(corsOrigin, "http://") || strings.HasPrefix(corsOrigin, "https://"),
						"CORS origin should have scheme: %s", corsOrigin)
					// CORS origin 应该包含 CSRF origin（主机名部分）
					assert.Contains(t, corsOrigin, csrfOrigin,
						"CORS origin %s should contain CSRF origin %s", corsOrigin, csrfOrigin)
				}
			}
		})
	}
}

// TestRealWorldScenario 测试真实场景：模拟从环境变量读取配置
func TestRealWorldScenario(t *testing.T) {
	// 模拟环境变量配置
	envValue := "https://dev-t.paas3-dev.bktencent.com:8888,https://bk-micro-web.paas3-dev.bktencent.com"
	origins := strings.Split(envValue, ",")

	// CORS 处理结果
	corsOrigins := NormalizeOriginsForCORS(origins)
	assert.Equal(t, []string{
		"https://dev-t.paas3-dev.bktencent.com:8888",
		"https://bk-micro-web.paas3-dev.bktencent.com",
	}, corsOrigins, "CORS origins should preserve full URLs")

	// CSRF 处理结果
	csrfOrigins := ExtractHostsForCSRF(corsOrigins)
	assert.Equal(t, []string{
		"dev-t.paas3-dev.bktencent.com:8888",
		"bk-micro-web.paas3-dev.bktencent.com",
	}, csrfOrigins, "CSRF origins should be host-only")

	// 验证 CSRF origin 可以匹配 Origin header
	// Origin header 通常是 "https://bk-micro-web.paas3-dev.bktencent.com" 格式
	// gorilla/csrf 会从 Origin header 中提取 host 进行匹配
	// 模拟请求 Origin: https://bk-micro-web.paas3-dev.bktencent.com
	expectedCSRFMatch := "bk-micro-web.paas3-dev.bktencent.com"
	assert.Contains(t, csrfOrigins, expectedCSRFMatch,
		"CSRF trusted origins should contain host from request Origin header")

	// 验证带端口的情况
	// 模拟请求 Origin: https://dev-t.paas3-dev.bktencent.com:8888
	expectedCSRFMatchWithPort := "dev-t.paas3-dev.bktencent.com:8888"
	assert.Contains(t, csrfOrigins, expectedCSRFMatchWithPort,
		"CSRF trusted origins should contain host:port from request Origin header")

	// 打印处理结果便于调试
	t.Logf("Environment variable: %s", envValue)
	t.Logf("CORS origins: %v", corsOrigins)
	t.Logf("CSRF trusted origins: %v", csrfOrigins)
}

// TestHostOnlyConfig 测试只配置主机名的场景
func TestHostOnlyConfig(t *testing.T) {
	// 用户可能只配置主机名（不带 scheme）
	envValue := "dev-t.paas3-dev.bktencent.com:8888,bk-micro-web.paas3-dev.bktencent.com"
	origins := strings.Split(envValue, ",")

	// CORS 处理结果 - 应该自动补全 https://
	corsOrigins := NormalizeOriginsForCORS(origins)
	assert.Equal(t, []string{
		"https://dev-t.paas3-dev.bktencent.com:8888",
		"https://bk-micro-web.paas3-dev.bktencent.com",
	}, corsOrigins, "CORS should auto-add https:// scheme")

	// CSRF 处理结果 - 应该保持主机名格式
	csrfOrigins := ExtractHostsForCSRF(corsOrigins)
	assert.Equal(t, []string{
		"dev-t.paas3-dev.bktencent.com:8888",
		"bk-micro-web.paas3-dev.bktencent.com",
	}, csrfOrigins, "CSRF origins should be host-only")
}
