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
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
)

func allowHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func notAllowHandler(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func TestHandlerAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	noAllowHandlerMap = map[string]bool{
		"middleware.allowHandler":    false,
		"middleware.notAllowHandler": true,
	}
	config.G = &config.Config{
		Service: config.ServiceConfig{
			DemoMode: true,
		},
	}
	router := gin.New()
	router.Use(HandlerAccess())
	router.GET("/allow", allowHandler)
	router.GET("/not_allow", notAllowHandler)
	tests := []struct {
		Name               string
		url                string
		expectedStatusCode int
	}{
		{
			"allow",
			"/allow",
			http.StatusOK,
		},
		{
			"not allow",
			"/not_allow",
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}
