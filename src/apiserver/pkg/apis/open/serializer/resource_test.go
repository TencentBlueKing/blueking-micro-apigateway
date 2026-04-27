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

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	testingutil "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/testing"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestResourceBatchCreateRequestToCommonResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		request      ResourceBatchCreateRequest
		assertions   func(t *testing.T, resources []*model.ResourceCommonModel)
	}{
		{
			name:         "inject route name and generate id when missing",
			resourceType: constant.Route,
			request: ResourceBatchCreateRequest{
				{
					Name: "route-a",
					Config: json.RawMessage(
						`{"uris":["/test"],"methods":["GET"],"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
					),
				},
			},
			assertions: func(t *testing.T, resources []*model.ResourceCommonModel) {
				assert.Len(t, resources, 1)
				assert.NotEmpty(t, resources[0].ID)
				assert.Empty(t, gjson.GetBytes(resources[0].Config, "name").String())
				assert.Equal(t, "route-a", resources[0].GetName(constant.Route))
			},
		},
		{
			name:         "keep config id while stripping echoed route name",
			resourceType: constant.Route,
			request: ResourceBatchCreateRequest{
				{
					Name: "inner-route",
					Config: json.RawMessage(
						`{"id":"route-id","name":"inner-route","uris":["/test"],"methods":["GET"],"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
					),
				},
			},
			assertions: func(t *testing.T, resources []*model.ResourceCommonModel) {
				assert.Len(t, resources, 1)
				assert.Equal(t, "route-id", resources[0].ID)
				assert.Empty(t, gjson.GetBytes(resources[0].Config, "name").String())
				assert.Equal(t, "inner-route", resources[0].GetName(constant.Route))
			},
		},
		{
			name:         "fallback path still stores raw config and keeps outer route name",
			resourceType: constant.Route,
			request: ResourceBatchCreateRequest{
				{
					Name: "outer-route",
					Config: json.RawMessage(
						`{"id":"route-id","name":"inner-route","uris":["/test"],"methods":["GET"],"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
					),
				},
			},
			assertions: func(t *testing.T, resources []*model.ResourceCommonModel) {
				assert.Len(t, resources, 1)
				assert.Equal(t, "route-id", resources[0].ID)
				assert.Empty(t, gjson.GetBytes(resources[0].Config, "name").String())
				assert.Equal(t, "outer-route", resources[0].GetName(constant.Route))
			},
		},
		{
			name:         "consumer uses username key injection",
			resourceType: constant.Consumer,
			request: ResourceBatchCreateRequest{{
				Name:   "consumer-a",
				Config: json.RawMessage(`{"plugins":{"key-auth":{"key":"token-a"}}}`),
			}},
			assertions: func(t *testing.T, resources []*model.ResourceCommonModel) {
				assert.Len(t, resources, 1)
				assert.NotEmpty(t, resources[0].ID)
				assert.Empty(t, gjson.GetBytes(resources[0].Config, "username").String())
				assert.Empty(t, gjson.GetBytes(resources[0].Config, "name").String())
				assert.Equal(t, "consumer-a", resources[0].GetName(constant.Consumer))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := tt.request.ToCommonResource(1001, tt.resourceType)
			tt.assertions(t, resources)
		})
	}
}

func TestResourceUpdateRequestToCommonResource(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	c := testingutil.CreateTestContextWithDefaultRequest(w)
	gateway := data.Gateway1WithBkAPISIX()
	gateway.ID = 1001
	ginx.SetGatewayInfo(c, gateway)
	ginx.SetUserID(c, "tester")

	req := ResourceUpdateRequest{
		Name:   "inner-route",
		Config: json.RawMessage(`{"name":"inner-route","uris":["/test"]}`),
	}

	resource := req.ToCommonResource(c, "route-id", constant.ResourceStatusUpdateDraft)
	assert.NotNil(t, resource)
	assert.Equal(t, "route-id", resource.ID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, resource.Status)
	assert.Equal(t, gateway.ID, resource.GatewayID)
	assert.Equal(t, "tester", resource.Updater)
	assert.JSONEq(t, `{"uris":["/test"]}`, string(resource.Config))
	assert.Equal(t, "inner-route", resource.GetName(constant.Route))
}
