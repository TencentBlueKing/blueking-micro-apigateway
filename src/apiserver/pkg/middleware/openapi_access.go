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

// Package middleware ...
package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// OpenAPIAccess  openapi 权限校验
func OpenAPIAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		gatewayName := c.Param("gateway_name")
		queryToken := c.GetHeader(constant.OpenAPITokenHeaderKey)
		if gatewayName != "" {
			gatewayInfo, err := biz.GetGatewayByName(c.Request.Context(), gatewayName)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				ginx.SystemErrorJSONResponse(c, err)
				c.Abort()
				return
			}
			if gatewayInfo == nil {
				ginx.BadRequestErrorJSONResponse(c, fmt.Errorf("网关 [%s] 不存在", gatewayName))
				c.Abort()
				return
			}
			// 非第一次注册，则校验 token
			if !config.G.Service.Standalone && queryToken != gatewayInfo.Token {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			ginx.SetGatewayInfo(c, gatewayInfo)
		}

		// 两种情况：
		// 独立部署，校验 token 是否是在白名单里面。
		// 非独立部署，校验 token 是否是自动生成的。
		if config.G.Service.Standalone &&
			(queryToken == "" || !config.G.Biz.OpenApiTokenWhitelist[queryToken]) {
			log.ErrorFWithContext(c.Request.Context(), "openapi token [%s] is not valid", queryToken)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ginx.SetValidateErrorInfo(c)
		c.Next()
	}
}
