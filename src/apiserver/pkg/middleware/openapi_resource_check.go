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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/open/serializer"
	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	resourcevalidationbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resourcevalidation"
	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

var noneValidateSchemaHandlerSet = map[string]struct{}{
	// 发布接口不需要进行 schema 校验
	"handler.ResourcePublish": {},
}

// OpenAPIResourceCheck 资源操作校验
func OpenAPIResourceCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		pathPrefix := "/api/v1/open/gateways/:gateway_name/resources/"
		_, ok := strings.CutPrefix(c.FullPath(), pathPrefix)
		if !ok {
			c.Next()
			return
		}
		resourceTypeValue := c.Param("resource_type")
		resourceType, ok := constant.ResourcePathToTypeMap[constant.ResourcePath(resourceTypeValue)]
		if !ok {
			ginx.BadRequestErrorJSONResponse(c,
				fmt.Errorf("invalid resource type in path: %s", resourceTypeValue))
			c.Abort()
			return
		}
		ginx.SetResourceType(c, resourceType)

		method := strings.ToUpper(c.Request.Method)

		// 针对单个资源操作进行统一的状态机判断：
		if c.Param("id") != "" && (method == http.MethodPut || method == http.MethodDelete) {
			resourceInfo, err := resourcebiz.GetResourceByID(
				c.Request.Context(),
				resourceType,
				c.Param("id"),
			)
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
			}
			statusOp := status.NewResourceStatusOp(resourceInfo)
			// 校验资源操作变更
			err = statusOp.CanDo(c.Request.Context(), op)
			if err != nil {
				ginx.BadRequestErrorJSONResponse(c, fmt.Errorf(
					"status: %s can not do: %s,err: %s", resourceInfo.Status, op, err.Error()))
				c.Abort()
				return
			}
		}

		// 删除操作和查询操作不需要校验 schema
		if method == http.MethodDelete || method == http.MethodGet {
			c.Next()
			return
		}

		// 校验资源配置
		reqBody, err := c.GetRawData()
		if err != nil {
			ginx.BadRequestErrorJSONResponse(c, errors.Wrapf(err, "invalid config"))
			c.Abort()
			return
		}
		// other filter need it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		// 针对某些资源进行 schema 校验
		fullHandlerName := c.HandlerName()
		lastSlashIndex := strings.LastIndex(fullHandlerName, "/")
		handlerName := fullHandlerName[lastSlashIndex+1:]
		if _, ok := noneValidateSchemaHandlerSet[handlerName]; ok {
			c.Next()
			return
		}
		// validate config schema
		configs := gjson.ParseBytes(reqBody).Array()
		version := ginx.GetGatewayInfo(c).GetAPISIXVersionX()
		drafts := make([]serializer.OpenResolvedDraft, 0, len(configs))
		var databaseValidator resourcevalidationbiz.DatabasePayloadValidator
		if len(configs) > 0 {
			customizePluginSchemaMap, err := schemabiz.GetCustomizePluginSchemaMap(c.Request.Context())
			if err != nil {
				ginx.SystemErrorJSONResponse(c, err)
				c.Abort()
				return
			}
			databaseValidator, err = resourcevalidationbiz.NewDatabasePayloadValidator(
				version,
				resourceType,
				customizePluginSchemaMap,
			)
			if err != nil {
				ginx.BadRequestErrorJSONResponse(c, errors.Wrapf(err, "config validate failed"))
				c.Abort()
				return
			}
		}
		for _, config := range configs {
			configRaw := config.Get("config").Raw

			configRawForValidation := resourcevalidationbiz.PrepareOpenValidationPayload(
				version,
				resourceType,
				configRaw,
			)
			resolvedID := gjson.GetBytes(configRawForValidation, "id").String()
			drafts = append(drafts, serializer.BuildOpenResolvedDraft(
				resourceType,
				serializer.ResourceCreateRequest{
					Name:   config.Get("name").String(),
					Config: json.RawMessage(configRaw),
				},
				resolvedID,
			))

			if err = databaseValidator.Validate(configRawForValidation); err != nil {
				logging.Errorf("database payload validate failed, err: %v", err)
				ginx.BadRequestErrorJSONResponse(c, err)
				c.Abort()
				return
			}

			// 校验关联数据是否存在
			var resourceAssociateIDInfo serializer.ResourceAssociateID
			err = json.Unmarshal([]byte(configRaw), &resourceAssociateIDInfo)
			if err != nil {
				ginx.BadRequestErrorJSONResponse(c, errors.Wrapf(err, "invalid config"))
				c.Abort()
				return
			}
			err = validation.ValidateStruct(c.Request.Context(), &resourceAssociateIDInfo)
			if err != nil {
				ginx.BadRequestErrorJSONResponse(c, err)
				c.Abort()
				return
			}
		}
		serializer.SetOpenResolvedDrafts(c, drafts)
		c.Next()
	}
}
