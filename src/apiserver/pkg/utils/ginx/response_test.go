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
package ginx_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

func TestSuccessJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ginx.SuccessJSONResponse(c, "test data")

	assert.Equal(t, http.StatusOK, w.Code)
	var got ginx.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, "test data", got.Data)
}

func TestSuccessCreateResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ginx.SuccessCreateResponse(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSuccessNoContentResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ginx.SuccessNoContentResponse(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "", w.Body.String())
}

func TestBaseErrorJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ginx.BaseErrorJSONResponse(c, "TEST_ERROR", "test message", http.StatusBadRequest)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got ginx.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, "TEST_ERROR", got.Error.Code)
	assert.Equal(t, "test message", got.Error.Message)
}

func TestBaseErrorJSONResponseWithData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	extraData := map[string]interface{}{"field": "value"}
	ginx.BaseErrorJSONResponseWithData(c, "TEST_ERROR", "test message", http.StatusBadRequest, extraData)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var got ginx.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, "TEST_ERROR", got.Error.Code)
	assert.Equal(t, "test message", got.Error.Message)
	assert.Equal(t, extraData, got.Error.Data)
}

func TestNewErrorJSONResponse(t *testing.T) {
	t.Run("normal error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Request.Header.Set("X-Request-Id", "test-request-id")

		handler := ginx.NewErrorJSONResponse("TEST_ERROR", http.StatusBadRequest)
		handler(c, errors.New("test error"))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var got ginx.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, "TEST_ERROR", got.Error.Code)
		assert.Equal(t, "test error", got.Error.Message)
	})

	t.Run("validation error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{Header: make(http.Header)}
		c.Request.Header.Set("X-Request-Id", "test-request-id")

		// 设置上下文中的验证错误
		ctx := context.WithValue(c.Request.Context(), constant.APISIXValidateErrKey, &schema.APISIXValidateError{
			Err: errors.New("validation error from context"),
		})
		c.Request = c.Request.WithContext(ctx)

		// 模拟验证错误
		handler := ginx.NewErrorJSONResponse("TEST_ERROR", http.StatusBadRequest)
		handler(c, errors.New("original error"))

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var got ginx.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &got)
		assert.NoError(t, err)
		assert.Equal(t, "validation error from context", got.Error.Message)
	})
}

func TestSystemErrorJSONResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置Request对象
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	// 设置RequestID
	c.Request.Header.Set("X-Request-Id", "test-request-id")

	ginx.SystemErrorJSONResponse(c, errors.New("test error"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got ginx.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, ginx.SystemError, got.Error.Code)
	assert.Contains(t, got.Error.Message, "test error")
}

func TestNewPaginatedRespData(t *testing.T) {
	data := ginx.NewPaginatedRespData(100, []string{"alpha", "beta", "gamma"})
	assert.Equal(t, ginx.PaginatedResponse{Count: int64(100), Results: []string{"alpha", "beta", "gamma"}}, data)
}
