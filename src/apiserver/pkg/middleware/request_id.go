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
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/uuidx"
)

// RequestID 中间件向 gin context 中注入 requestID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(constant.RequestIDHeaderKey)

		if len(requestID) == 0 {
			requestID = uuidx.New()
		}

		ginx.SetRequestID(c, requestID)

		// 设置日志上下文中的 requestID
		ginx.SetSlogGinRequestID(c, requestID)

		c.Request = c.Request.WithContext(
			logging.AppendCtx(c.Request.Context(), slog.String(constant.RequestIDHeaderKey, requestID)),
		)

		// 设置 response header 中的 requestID
		c.Writer.Header().Set(constant.RequestIDHeaderKey, requestID)

		c.Next()
	}
}
