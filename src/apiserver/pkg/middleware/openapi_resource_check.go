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
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

var noneValidateSchemaHandlerMap = map[string]bool{
	// 发布接口不需要进行 schema 校验
	"handler.ResourcePublish": false,
}

const openAPIRequestDraftsKey = "openapi_request_drafts"

// SetOpenAPIRequestDrafts stores the canonical drafts resolved during OpenAPI request validation.
func SetOpenAPIRequestDrafts(c *gin.Context, drafts []resourcecodec.CanonicalDraft) {
	c.Set(openAPIRequestDraftsKey, drafts)
}

// GetOpenAPIRequestDrafts returns the canonical drafts resolved during OpenAPI request validation.
func GetOpenAPIRequestDrafts(c *gin.Context) ([]resourcecodec.CanonicalDraft, bool) {
	value, ok := c.Get(openAPIRequestDraftsKey)
	if !ok {
		return nil, false
	}
	drafts, ok := value.([]resourcecodec.CanonicalDraft)
	return drafts, ok
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
			resourceInfo, err := biz.GetResourceByID(c.Request.Context(), resourceType, c.Param("id"))
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
		if _, ok := noneValidateSchemaHandlerMap[handlerName]; ok {
			c.Next()
			return
		}
		// validate config schema
		configs := gjson.ParseBytes(reqBody).Array()
		resolvedDrafts := make([]resourcecodec.CanonicalDraft, 0, len(configs))
		for _, config := range configs {
			schemaValidator, err := schema.NewAPISIXSchemaValidator(
				ginx.GetGatewayInfo(c).GetAPISIXVersionX(),
				"main."+resourceType.String(),
			)
			if err != nil {
				ginx.BadRequestErrorJSONResponse(c, errors.Wrapf(err, "config validate failed"))
				c.Abort()
				return
			}
			configRaw := config.Get("config").Raw
			outerName := config.Get("name").String()
			draft, err := resourcecodec.CanonicalizeRequest(resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    openAPIOperation(method),
				GatewayID:    ginx.GetGatewayInfo(c).ID,
				ResourceType: resourceType,
				Version:      ginx.GetGatewayInfo(c).GetAPISIXVersionX(),
				PathID:       c.Param("id"),
				OuterName:    outerName,
				Config:       json.RawMessage(configRaw),
			})
			if err != nil {
				ginx.BadRequestErrorJSONResponse(c, err)
				c.Abort()
				return
			}
			resolvedDrafts = append(resolvedDrafts, draft)
			// OpenAPI request validation now runs against the canonical DATABASE projection too.
			materialized, err := resourcecodec.MaterializeRequestDraft(draft, constant.DATABASE)
			if err != nil {
				ginx.BadRequestErrorJSONResponse(c, err)
				c.Abort()
				return
			}
			configRawForValidation := materialized.Payload

			if err = schemaValidator.Validate(configRawForValidation); err != nil {
				logging.Errorf("schema validate failed, err: %v", err)
				ginx.BadRequestErrorJSONResponse(c, errors.Wrapf(err, "config validate failed"))
				c.Abort()
				return
			}
			// 配置校验
			customizePluginSchemaMap, err := biz.GetCustomizePluginSchemaMap(c.Request.Context())
			if err != nil {
				ginx.SystemErrorJSONResponse(c, err)
				c.Abort()
				return
			}
			jsonConfigValidator, err := schema.NewAPISIXJsonSchemaValidator(
				ginx.GetGatewayInfo(c).GetAPISIXVersionX(),
				resourceType,
				"main."+string(resourceType),
				customizePluginSchemaMap,
				constant.DATABASE,
			)
			if err != nil {
				ginx.BadRequestErrorJSONResponse(
					c,
					fmt.Errorf(
						"NewAPISIXJsonSchemaValidator failed, resource config:%s validate failed, err: %w",
						configRaw,
						err,
					),
				)
				c.Abort()
				return
			}
			if err = jsonConfigValidator.Validate(configRawForValidation); err != nil { // 校验 json schema
				ginx.BadRequestErrorJSONResponse(
					c,
					fmt.Errorf("resource config:%s validate failed, err: %w",
						configRaw, err),
				)
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
		if method == http.MethodPost && len(resolvedDrafts) == len(configs) {
			SetOpenAPIRequestDrafts(c, resolvedDrafts)
		}
		c.Next()
	}
}

func openAPIOperation(method string) constant.OperationType {
	if method == http.MethodPut {
		return constant.OperationTypeUpdate
	}
	return constant.OperationTypeCreate
}
