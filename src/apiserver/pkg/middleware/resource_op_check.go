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
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ResourceOperationCheck 资源操作变更校验中间件
func ResourceOperationCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := strings.ToUpper(c.Request.Method)
		if method != http.MethodPut && method != http.MethodDelete {
			c.Next()
			return
		}
		pathPrefix := "/api/v1/web/gateways/:gateway_id/"
		resourcePath, ok := strings.CutPrefix(c.FullPath(), pathPrefix)
		if !ok {
			c.Next()
			return
		}
		resourcePaths := strings.Split(resourcePath, "/")
		if len(resourcePaths) != 3 {
			c.Next()
			return
		}
		resourceType, ok := constant.ResourcePrefixTypeMap[resourcePaths[0]]
		if !ok {
			c.Next()
			return
		}
		resourceId := c.Param("id")
		resourceInfo, err := biz.GetResourceByID(c.Request.Context(), resourceType, resourceId)
		if err != nil {
			ginx.BadRequestErrorJSONResponse(c, err)
			c.Abort()
			return
		}
		var op constant.OperationType
		switch method {
		case http.MethodPut:
			op = constant.OperationTypeUpdate
		case http.MethodDelete:
			op = constant.OperationTypeDelete
			// 校验资源是否被引用
			relationResourceTypes, ok := constant.ResourceRelationMap[resourceType]
			if ok {
				for _, relationResourceType := range relationResourceTypes {
					resources, err := biz.QueryResource(
						c.Request.Context(),
						relationResourceType,
						map[string]any{
							resourceType.RelationIDFiled(): resourceId,
						},
						"",
					)
					if err != nil {
						ginx.BadRequestErrorJSONResponse(c, err)
						c.Abort()
						return
					}
					if len(resources) > 0 {
						ginx.BadRequestErrorJSONResponse(
							c,
							fmt.Errorf(
								"该资源不能删除，被: %s %s 引用",
								relationResourceType,
								resources[0].ID,
							),
						)
						c.Abort()
						return
					}
				}
			}
		}
		statusOp := status.NewResourceStatusOp(resourceInfo)
		// 校验资源操作变更
		err = statusOp.CanDo(c.Request.Context(), op)
		if err != nil {
			ginx.BadRequestErrorJSONResponse(c, err)
			c.Abort()
			return
		}
		c.Next()
	}
}
