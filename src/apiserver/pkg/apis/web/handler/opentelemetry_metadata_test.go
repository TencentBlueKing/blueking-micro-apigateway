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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

func TestOpentelemetryMetadataWebAPI(t *testing.T) {
	initWebCreateHandlerTestEnv()
	for i, version := range []string{"3.13.0", "3.17.0", "3.18.0"} {
		for j, apisixType := range []string{constant.APISIXTypeAPISIX, constant.APISIXTypeBKAPISIX} {
			t.Run(fmt.Sprintf("%s/%s", version, apisixType), func(t *testing.T) {
				gateway := &model.Gateway{
					ID:            15400 + i*2 + j,
					APISIXVersion: version,
					APISIXType:    apisixType,
				}
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodGet, "/plugins/?kind=metadata", nil)
				ginx.SetGatewayInfo(c, gateway)
				PluginsGet(c)
				require.Equal(t, http.StatusOK, w.Code, w.Body.String())
				var response struct {
					Data serializer.PluginListResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
				var found *schema.Plugin
				for _, group := range response.Data {
					for _, plugin := range group.Plugins {
						if plugin.Name == "opentelemetry" {
							require.Nil(t, found)
							found = plugin
						}
					}
				}
				require.NotNil(t, found, "metadata dropdown must include opentelemetry")
				require.NotEmpty(t, found.MetadataExample)

				w = httptest.NewRecorder()
				c, _ = gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(
					http.MethodGet,
					"/schemas/plugins/opentelemetry/?schema_type=metadata",
					nil,
				)
				c.Params = gin.Params{{Key: "name", Value: "opentelemetry"}}
				ginx.SetGatewayInfo(c, gateway)
				PluginSchemaGet(c)
				require.Equal(t, http.StatusOK, w.Code, w.Body.String())
				var schemaResponse struct {
					Data map[string]any `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &schemaResponse))
				assert.Equal(t, "object", schemaResponse.Data["type"])
				assert.Contains(t, schemaResponse.Data["properties"], "collector")

				body := `{"name":"opentelemetry", "config":{"collector":{"request_timeout":"3"}}}`
				c, w = newWebCreateTestContext(t, body, gateway, "metadata-tester")
				PluginMetadataCreate(c)
				assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

				body = mustJSONBody(
					t,
					map[string]any{"name": "opentelemetry", "config": found.MetadataExample},
				)
				c, w = newWebCreateTestContext(t, body, gateway, "metadata-tester")
				PluginMetadataCreate(c)
				require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
				stored, err := resourcebiz.ListPluginMetadatas(c.Request.Context())
				require.NoError(t, err)
				require.Len(t, stored, 1)
				var config map[string]any
				require.NoError(t, json.Unmarshal(stored[0].Config, &config))
				assert.Equal(t, "opentelemetry", config["id"])
				assert.Equal(t, found.MetadataExample["resource"], config["resource"])
				validator, err := schema.NewAPISIXJsonSchemaValidator(
					gateway.GetAPISIXVersionX(),
					constant.PluginMetadata,
					"main.plugin_metadata",
					nil,
					constant.ETCD,
				)
				require.NoError(t, err)
				assert.NoError(t, validator.Validate(json.RawMessage(stored[0].Config)))
			})
		}
	}
}
