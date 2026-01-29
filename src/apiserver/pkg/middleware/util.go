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

package middleware

import (
	"net/url"
	"strings"
)

// NormalizeOriginsForCORS 规范化处理 origins 列表，用于 CORS 中间件
// gin-contrib/cors 需要完整 URL 格式（包含 http:// 或 https://）
// 如果输入是纯主机名，会自动补全 https://
func NormalizeOriginsForCORS(origins []string) []string {
	result := make([]string, 0, len(origins))

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}

		// 如果是 "*"，保持不变
		if origin == "*" {
			result = append(result, origin)
			continue
		}

		// 判断是否已经有 scheme
		hasScheme := strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://")

		if hasScheme {
			// 已经是完整 URL 格式
			result = append(result, origin)
		} else {
			// 纯主机名格式，需要补全 scheme
			result = append(result, "https://"+origin)
		}
	}

	return result
}

// ExtractHostsForCSRF 从 origins 列表中提取主机名，用于 CSRF 中间件
// gorilla/csrf 的 TrustedOrigins 只需要主机名（如 example.com:8080），不需要 scheme
func ExtractHostsForCSRF(origins []string) []string {
	result := make([]string, 0, len(origins))

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}

		// 如果是 "*"，保持不变
		if origin == "*" {
			result = append(result, origin)
			continue
		}

		// 判断是否已经有 scheme
		hasScheme := strings.HasPrefix(origin, "http://") || strings.HasPrefix(origin, "https://")

		if hasScheme {
			// 提取主机名
			parsedURL, err := url.Parse(origin)
			if err != nil || parsedURL.Host == "" {
				result = append(result, origin)
			} else {
				result = append(result, parsedURL.Host)
			}
		} else {
			// 已经是纯主机名格式
			result = append(result, origin)
		}
	}

	return result
}
