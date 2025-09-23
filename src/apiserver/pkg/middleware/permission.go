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

	"github.com/gin-gonic/gin"
	"github.com/gookit/goutil"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// Permission 权限校验
func Permission() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := ginx.GetUserID(c)
		if user != "" && config.G.Service.Standalone {
			users, err := biz.GetAllowUsers(c.Request.Context())
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				ginx.SystemErrorJSONResponse(c, err)
				c.Abort()
				return
			}
			if len(users) != 0 && !goutil.Contains(users, user) {
				ginx.ForbiddenJSONResponse(c,
					fmt.Errorf("user %s is not allowed to access the site. Please concat "+
						"the administrator to grant permission", user))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
