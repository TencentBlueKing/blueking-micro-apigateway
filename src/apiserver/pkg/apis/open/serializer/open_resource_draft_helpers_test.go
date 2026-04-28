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

package serializer

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	testingutil "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/testing"
)

func TestBuildOpenCreateDraft(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		req          ResourceCreateRequest
		resolvedID   string
		assertDraft  func(t *testing.T, got OpenResolvedDraft)
	}{
		{
			name:         "route injects name and generates id",
			resourceType: constant.Route,
			req: ResourceCreateRequest{
				Name:   "route-demo",
				Config: json.RawMessage(`{"uri":"/demo"}`),
			},
			assertDraft: func(t *testing.T, got OpenResolvedDraft) {
				t.Helper()
				assert.NotEmpty(t, got.ID)
				assert.Equal(t, "route-demo", gjson.GetBytes(got.StorageConfig, "name").String())
				assert.Equal(t, "/demo", gjson.GetBytes(got.StorageConfig, "uri").String())
			},
		},
		{
			name:         "consumer writes username when config omits it",
			resourceType: constant.Consumer,
			req: ResourceCreateRequest{
				Name:   "consumer-demo",
				Config: json.RawMessage(`{"plugins":{}}`),
			},
			assertDraft: func(t *testing.T, got OpenResolvedDraft) {
				t.Helper()
				assert.Equal(t, "consumer-demo", gjson.GetBytes(got.StorageConfig, "username").String())
				assert.False(t, gjson.GetBytes(got.StorageConfig, "name").Exists())
			},
		},
		{
			name:         "consumer still overwrites username because create path checks literal name",
			resourceType: constant.Consumer,
			req: ResourceCreateRequest{
				Name:   "outer-name",
				Config: json.RawMessage(`{"username":"config-name"}`),
			},
			assertDraft: func(t *testing.T, got OpenResolvedDraft) {
				t.Helper()
				assert.Equal(t, "outer-name", gjson.GetBytes(got.StorageConfig, "username").String())
			},
		},
		{
			name:         "resolved id is reused without writing id into config",
			resourceType: constant.PluginConfig,
			req: ResourceCreateRequest{
				Name:   "pc-demo",
				Config: json.RawMessage(`{"plugins":{}}`),
			},
			resolvedID: "resolved-id",
			assertDraft: func(t *testing.T, got OpenResolvedDraft) {
				t.Helper()
				assert.Equal(t, "resolved-id", got.ID)
				assert.False(t, gjson.GetBytes(got.StorageConfig, "id").Exists())
				assert.Equal(t, "pc-demo", gjson.GetBytes(got.StorageConfig, "name").String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildOpenResolvedDraft(tt.resourceType, tt.req, tt.resolvedID)
			tt.assertDraft(t, got)
		})
	}
}

func TestBuildOpenResolvedDraft(t *testing.T) {
	got := BuildOpenResolvedDraft(
		constant.PluginConfig,
		ResourceCreateRequest{
			Name:   "pc-demo",
			Config: json.RawMessage(`{"plugins":{}}`),
		},
		"resolved-id",
	)

	assert.Equal(t, "resolved-id", got.ID)
	assert.Equal(t, "pc-demo", gjson.GetBytes(got.StorageConfig, "name").String())
	assert.False(t, gjson.GetBytes(got.StorageConfig, "id").Exists())
}

func TestBuildOpenUpdateDraft(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c := testingutil.CreateTestContextWithDefaultRequest(recorder)
	ginx.SetGatewayInfo(c, &model.Gateway{ID: 7})
	ginx.SetUserID(c, "openapi-user")

	got := buildOpenUpdateDraft(
		c,
		"route-id",
		constant.ResourceStatusUpdateDraft,
		json.RawMessage(`{"uri":"/demo","name":"route-demo"}`),
	)

	assert.Equal(t, "route-id", got.ID)
	assert.Equal(t, 7, got.GatewayID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, got.Status)
	assert.Equal(t, "openapi-user", got.Updater)
	assert.JSONEq(t, `{"uri":"/demo","name":"route-demo"}`, string(got.Config))
}
