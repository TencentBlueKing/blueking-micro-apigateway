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

package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
)

func TestEnumTAPISIXVisibility(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		wantTAPISIX bool
	}{
		{
			name:        "hidden before config is loaded",
			config:      nil,
			wantTAPISIX: false,
		},
		{
			name: "hidden by default",
			config: &config.Config{Service: config.ServiceConfig{
				EnableTAPISIX: false,
			}},
			wantTAPISIX: false,
		},
		{
			name: "visible when enabled",
			config: &config.Config{Service: config.ServiceConfig{
				EnableTAPISIX: true,
			}},
			wantTAPISIX: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previousConfig := config.G
			config.G = tt.config
			t.Cleanup(func() {
				config.G = previousConfig
			})

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			Enum(c)

			response := gjson.ParseBytes(w.Body.Bytes())
			assert.Equal(t, tt.wantTAPISIX, response.Get("data.apisix_type.tapisix").Exists())
			assert.Equal(t, tt.wantTAPISIX, response.Get("data.support_apisix_version.tapisix").Exists())
		})
	}
}
