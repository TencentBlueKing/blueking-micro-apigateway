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

// Package ginx 提供一些 Gin 框架相关的工具
package ginx

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// GetRequestID ...
func GetRequestID(c *gin.Context) string {
	return c.GetString(constant.RequestIDCtxKey)
}

// SetRequestID ...
func SetRequestID(c *gin.Context, requestID string) {
	c.Set(constant.RequestIDCtxKey, requestID)
}

// SetSlogGinRequestID ...
func SetSlogGinRequestID(c *gin.Context, requestID string) {
	c.Request.Header.Set(constant.RequestIDHeaderKey, requestID)
}

// GetError ...
func GetError(c *gin.Context) (err any, ok bool) {
	return c.Get(constant.ErrorCtxKey)
}

// SetError ...
func SetError(c *gin.Context, err error) {
	c.Set(constant.ErrorCtxKey, err)
}

// GetUserID ...
func GetUserID(c *gin.Context) string {
	if c.Request == nil {
		return c.GetString(string(constant.UserIDKey))
	}
	userID, ok := c.Request.Context().Value(constant.UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserIDFromContext ...
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(constant.UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// SetTx sets a transaction for database operations
// This function takes a pointer to a repo.Query object which represents a database transaction
// It can be used to pass transaction context to various database operations
func SetTx(ctx context.Context, tx *repo.Query) context.Context {
	return context.WithValue(ctx, constant.DbTxKey, tx)
}

// GetTx get a transaction from context
func GetTx(ctx context.Context) *repo.Query {
	tx, ok := ctx.Value(constant.DbTxKey).(*repo.Query)
	if !ok {
		return nil
	}
	return tx
}

// SetUserID ...
func SetUserID(c *gin.Context, userID string) {
	c.Set(string(constant.UserIDKey), userID)
	if c.Request != nil {
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), constant.UserIDKey, userID))
	}
}

// GetGatewayInfo ...
func GetGatewayInfo(c *gin.Context) *model.Gateway {
	gatewayInfo, ok := c.Request.Context().Value(constant.GatewayInfoKey).(*model.Gateway)
	if !ok {
		return nil
	}
	return gatewayInfo
}

// GetGatewayInfoFromContext ...
func GetGatewayInfoFromContext(ctx context.Context) *model.Gateway {
	gatewayInfo, ok := ctx.Value(constant.GatewayInfoKey).(*model.Gateway)
	if !ok {
		return nil
	}
	return gatewayInfo
}

// SetGatewayInfo ...
func SetGatewayInfo(c *gin.Context, gatewayInfo *model.Gateway) {
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), constant.GatewayInfoKey, gatewayInfo))
}

// SetGatewayInfoToContext ...
func SetGatewayInfoToContext(c context.Context, gatewayInfo *model.Gateway) context.Context {
	return context.WithValue(c, constant.GatewayInfoKey, gatewayInfo)
}

// SetResourceType ...
func SetResourceType(c *gin.Context, resourceType constant.APISIXResource) {
	c.Request = c.Request.WithContext(
		context.WithValue(c.Request.Context(), constant.ResourceTypeKey, resourceType),
	)
}

// GetResourceType ...
func GetResourceType(c *gin.Context) constant.APISIXResource {
	resourceType, ok := c.Request.Context().Value(constant.ResourceTypeKey).(constant.APISIXResource)
	if !ok {
		return ""
	}
	return resourceType
}

// GetResourceTypeFromContext ...
func GetResourceTypeFromContext(ctx context.Context) constant.APISIXResource {
	resourceType, ok := ctx.Value(constant.ResourceTypeKey).(string)
	if !ok {
		return ""
	}
	return constant.APISIXResource(resourceType)
}

// GetValidateErrorInfoFromContext ...
func GetValidateErrorInfoFromContext(ctx context.Context) *schema.APISIXValidateError {
	validateError, ok := ctx.Value(constant.APISIXValidateErrKey).(*schema.APISIXValidateError)
	if !ok {
		return nil
	}
	return validateError
}

// SetValidateErrorInfo ...
func SetValidateErrorInfo(c *gin.Context) {
	validateError := &schema.APISIXValidateError{}
	c.Request = c.Request.WithContext(
		context.WithValue(c.Request.Context(), constant.APISIXValidateErrKey, validateError),
	)
}

// CloneCtx ...
func CloneCtx(ctx context.Context) context.Context {
	newCtx := context.Background()
	return context.WithValue(newCtx, constant.GatewayInfoKey, GetGatewayInfoFromContext(ctx))
}
