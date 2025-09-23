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
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/account"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// UserAuth 进行用户身份认证，并将用户信息注入到 context 中
func UserAuth(authBackend account.AuthBackend) gin.HandlerFunc {
	return func(c *gin.Context) {
		userToken, err := c.Request.Cookie(config.G.Service.UserTokenKey)
		// 重定向链接（当前访问的链接）
		scheme := lo.Ternary(c.Request.TLS != nil, "https", "http")
		referUrl := fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, c.Request.RequestURI)
		data := gin.H{
			"loginUrl":        authBackend.GetLoginUrl(),
			"login_plain_url": fmt.Sprintf("%s?c_url=%s", authBackend.GetLoginUrl(), referUrl),
			"width":           700,
			"height":          550,
		}
		// 没有获取到用户凭证信息 -> 401 -> 让用户通过页面登录
		if err != nil {
			ginx.BaseErrorJSONResponseWithData(c, ginx.UnauthorizedError,
				"用户未登录或登录态失效，请使用登录链接重新登录", http.StatusUnauthorized, data)
			c.Abort()
			return
		}

		session := sessions.Default(c)
		if userToken.Value == session.Get(config.G.Service.UserTokenKey) {
			// 从 session 获取用户信息并注入到 context
			ginx.SetUserID(c, session.Get(string(constant.UserIDKey)).(string))
			c.Next()
			return
		}
		userInfo, err := authBackend.GetUserInfo(userToken.Value)
		if err != nil {
			data["origin_error"] = err.Error()
			ginx.BaseErrorJSONResponseWithData(c, ginx.UnauthorizedError,
				"用户未登录或登录态失效，请使用登录链接重新登录", http.StatusUnauthorized, data)
			c.Abort()
			return
		}

		// 获取到用户凭证信息 -> 设置 context & session -> 通过
		ginx.SetUserID(c, userInfo.ID)
		session.Set(config.G.Service.UserTokenKey, userToken.Value)
		session.Set(string(constant.UserIDKey), userInfo.ID)
		_ = session.Save()
		c.Next()
	}
}
