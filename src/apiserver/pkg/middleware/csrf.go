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

// Package middleware ...
package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	adapter "github.com/gwatts/gin-adapter"
)

// CSRF 中间件用于防止跨站请求伪造
func CSRF(appID, secret string) gin.HandlerFunc {
	csrfMiddleware := csrf.Protect(
		[]byte(secret),
		csrf.Secure(false),
		csrf.Path("/"),
		csrf.CookieName(appID+"-csrf"),
	)

	return func(c *gin.Context) {
		// 对于非 HTTPS 环境，使用 PlaintextHTTPRequest 来标记请求
		// 这会跳过 Referer 检查，避免 v1.7.3 的默认 Referer 检查导致的问题
		c.Request = csrf.PlaintextHTTPRequest(c.Request)

		// 使用适配器包装 CSRF 中间件
		adapter.Wrap(csrfMiddleware)(c)
	}
}

// CSRFToken 中间件用于在 cookie 中设置 csrf token
func CSRFToken(appID, domain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetSameSite(http.SameSiteLaxMode)
		// 参数依次为：cookie 名称，值，有效期（0 表示会话 cookie）
		// 路径（根），域名（ "" 表示当前域），是否只通过 https 访问，httpOnly 属性
		c.SetCookie(appID+"-csrf-token", csrf.Token(c.Request),
			int(time.Hour.Seconds())*24, "/", domain, false, false)
	}
}
