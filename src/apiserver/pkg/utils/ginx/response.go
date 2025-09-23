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

package ginx

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// BadRequestError ...
const (
	BadRequestError   = "BadRequest"
	UnauthorizedError = "Unauthorized"
	ForbiddenError    = "Forbidden"
	NotFoundError     = "NotFound"
	ConflictError     = "Conflict"
	TooManyRequests   = "TooManyRequests"

	SystemError = "InternalServerError"
)

// SuccessResponse ...
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// Error ...
type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	System  string      `json:"system"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse  ...
type ErrorResponse struct {
	Error Error `json:"error"`
}

// SuccessJSONResponse ...
func SuccessJSONResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Data: data,
	})
}

// SuccessCreateResponse ...
func SuccessCreateResponse(c *gin.Context) {
	c.JSON(http.StatusCreated, nil)
}

// SuccessNoContentResponse ...
func SuccessNoContentResponse(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// SuccessFileResponse ...
func SuccessFileResponse(c *gin.Context, contentType string, fileData []byte, fileName string) {
	c.Header(
		"Content-Disposition",
		`attachment; filename=`+fileName,
	)
	c.Data(200, contentType, fileData)
}

// BaseErrorJSONResponse ...
func BaseErrorJSONResponse(c *gin.Context, errorCode string, message string, statusCode int) {
	// BaseJSONResponse(c, statusCode, code, message, gin.H{})
	c.JSON(statusCode, ErrorResponse{Error: Error{
		Code:    errorCode,
		Message: message,
		System:  "bk-micro-apigateway",
	}})
}

// BaseErrorJSONResponseWithData ...
func BaseErrorJSONResponseWithData(
	c *gin.Context,
	errorCode string,
	message string,
	statusCode int,
	data interface{},
) {
	// BaseJSONResponse(c, statusCode, code, message, gin.H{})
	c.JSON(statusCode, ErrorResponse{Error: Error{
		Code:    errorCode,
		Message: message,
		System:  "bk-micro-apigateway",
		Data:    data,
	}})
}

// NewErrorJSONResponse ...
func NewErrorJSONResponse(
	errorCode string,
	statusCode int,
) func(c *gin.Context, err error) {
	return func(c *gin.Context, err error) {
		// 判断校验是否通过
		vErr := GetValidateErrorInfoFromContext(c.Request.Context())
		if vErr != nil && vErr.Err != nil {
			err = vErr.Err
		}
		validateErr, ok := err.(validator.ValidationErrors)
		if ok {
			BaseErrorJSONResponse(c, BadRequestError, validation.TranslateToString(validateErr), http.StatusBadRequest)
			return
		}
		BaseErrorJSONResponse(c, errorCode, err.Error(), statusCode)
	}
}

// BadRequestErrorJSONResponse ...
var (
	BadRequestErrorJSONResponse = NewErrorJSONResponse(BadRequestError, http.StatusBadRequest)
	ForbiddenJSONResponse       = NewErrorJSONResponse(ForbiddenError, http.StatusForbidden)
	UnauthorizedJSONResponse    = NewErrorJSONResponse(UnauthorizedError, http.StatusUnauthorized)
	NotFoundJSONResponse        = NewErrorJSONResponse(NotFoundError, http.StatusNotFound)
	ConflictJSONResponse        = NewErrorJSONResponse(ConflictError, http.StatusConflict)
	TooManyRequestsJSONResponse = NewErrorJSONResponse(TooManyRequests, http.StatusTooManyRequests)
)

// SystemErrorJSONResponse ...
func SystemErrorJSONResponse(c *gin.Context, err error) {
	// 判断校验是否通过
	vErr := GetValidateErrorInfoFromContext(c.Request.Context())
	if vErr != nil && vErr.Err != nil {
		err = vErr.Err
	}
	validateErr, ok := err.(validator.ValidationErrors)
	if ok {
		BaseErrorJSONResponse(c, BadRequestError, validation.TranslateToString(validateErr), http.StatusBadRequest)
		return
	}
	message := fmt.Sprintf("system error[request_id=%s]: %s", GetRequestID(c), err.Error())
	BaseErrorJSONResponse(c, SystemError, message, http.StatusInternalServerError)
}

// PaginatedResponse 分页响应数据体
type PaginatedResponse struct {
	Count   int64 `json:"count"`
	Results any   `json:"results"`
}

// NewPaginatedRespData 创建分页响应数据体
// 注意：results 类型应该是 Slice / Array
func NewPaginatedRespData(count int64, results any) PaginatedResponse {
	return PaginatedResponse{Count: count, Results: results}
}
