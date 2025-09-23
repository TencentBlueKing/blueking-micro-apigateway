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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

var noAllowHandlerMap = map[string]bool{
	// demo模式不允许进行网关更新操作
	"handler.GatewayUpdate": true,
	"handler.GatewayCreate": true,
}

// HandlerAccess  权限校验
func HandlerAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		fullHandlerName := c.HandlerName()
		lastSlashIndex := strings.LastIndex(fullHandlerName, "/")
		handlerName := fullHandlerName[lastSlashIndex+1:]
		if !config.IsDemoMode() {
			c.Next()
			return
		}
		// demo模式下，不允许进行网关更新操作
		if disabled, ok := noAllowHandlerMap[handlerName]; ok && disabled {
			ginx.BadRequestErrorJSONResponse(c, errors.New(config.G.Service.DemoModeWarnMsg))
			c.Abort()
			return
		}
		c.Next()
	}
}
