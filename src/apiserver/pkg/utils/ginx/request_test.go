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
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

func TestGetOffset(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{
			name: "empty",
			path: "/",
			want: 0,
		},
		{
			name: "offset=1",
			path: "/?offset=1",
			want: 1,
		},
		{
			name: "offset=5",
			path: "/?offset=5",
			want: 5,
		},
		{
			name: "invalid offset",
			path: "/?offset=-3",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", tt.path, nil)
			assert.Equal(t, tt.want, ginx.GetOffset(c))
		})
	}
}

func TestGetLimit(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{
			name: "empty",
			path: "/",
			want: 5,
		},
		{
			name: "limit=15",
			path: "/?limit=15",
			want: 15,
		},
		{
			name: "invalid limit",
			path: "/?limit=-1",
			want: 5,
		},
		{
			name: "too large limit",
			path: "/?limit=500",
			want: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("GET", tt.path, nil)
			assert.Equal(t, tt.want, ginx.GetLimit(c))
		})
	}
}
