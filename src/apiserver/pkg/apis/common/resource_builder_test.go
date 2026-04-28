/*
 * TencentBlueKing is pleased to support the open source community by making
 * BlueKing - Micro APIGateway available.
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

package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
)

func TestPrepareStoredResource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      resourcecodec.RequestInput
		wantID     string
		wantConfig string
	}{
		{
			name: "route strips echoed id name and associations",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    constant.OperationTypeUpdate,
				GatewayID:    1001,
				ResourceType: constant.Route,
				PathID:       "route-id",
				OuterName:    "route-a",
				OuterFields: map[string]any{
					"service_id": "svc-a",
				},
				Config: json.RawMessage(
					`{"id":"route-id","name":"route-a","service_id":"svc-a","uris":["/test"]}`,
				),
			},
			wantID:     "route-id",
			wantConfig: `{"uris":["/test"]}`,
			wantValues: model.ResourceResolvedValues{
				NameValue:      "route-a",
				ServiceIDValue: "svc-a",
			},
		},
		{
			name: "consumer keeps username in resolved fields only",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    constant.OperationTypeUpdate,
				GatewayID:    1001,
				ResourceType: constant.Consumer,
				PathID:       "consumer-id",
				OuterName:    "consumer-a",
				OuterFields: map[string]any{
					"group_id": "group-a",
				},
				Config: json.RawMessage(
					`{"id":"consumer-id","username":"consumer-a","group_id":"group-a","plugins":{"key-auth":{"key":"token-a"}}}`,
				),
			},
			wantID:     "consumer-id",
			wantConfig: `{"plugins":{"key-auth":{"key":"token-a"}}}`,
			wantValues: model.ResourceResolvedValues{
				NameValue:    "consumer-a",
				GroupIDValue: "group-a",
			},
		},
		{
			name: "plugin metadata uses config id as resolved name",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.PluginMetadata,
				Config:       json.RawMessage(`{"id":"jwt-auth","name":"legacy-name","key":"value"}`),
			},
			wantID:     "jwt-auth",
			wantConfig: `{"key":"value"}`,
			wantValues: model.ResourceResolvedValues{
				NameValue: "jwt-auth",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prepared, err := PrepareStoredResource(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, prepared.ResourceID)
			assert.JSONEq(t, tt.wantConfig, string(prepared.StorageConfig))
			assert.Equal(t, tt.wantValues, prepared.ResolvedValues)
		})
	}
}

func TestBuildFallbackStoredResource(t *testing.T) {
	t.Parallel()

	input := resourcecodec.RequestInput{
		Source:       resourcecodec.SourceOpenAPI,
		Operation:    constant.OperationTypeUpdate,
		GatewayID:    1001,
		ResourceType: constant.Route,
		PathID:       "route-id",
		OuterName:    "outer-route",
		Config: json.RawMessage(
			`{"id":"route-id","name":"inner-route","service_id":"svc-a","uris":["/test"]}`,
		),
	}

	prepared := BuildFallbackStoredResource(input)
	assert.Equal(t, "route-id", prepared.ResourceID)
	assert.JSONEq(t, `{"uris":["/test"]}`, string(prepared.StorageConfig))
	assert.Equal(t, model.ResourceResolvedValues{
		NameValue:      "outer-route",
		ServiceIDValue: "svc-a",
	}, prepared.ResolvedValues)
}

func TestBuildResourceCommonModel(t *testing.T) {
	t.Parallel()

	resource := BuildResourceCommonModel(
		PreparedStoredResource{
			GatewayID:     1001,
			ResourceID:    "route-id",
			StorageConfig: []byte(`{"uris":["/test"]}`),
			ResolvedValues: model.ResourceResolvedValues{
				NameValue:      "route-a",
				ServiceIDValue: "svc-a",
			},
		},
		constant.ResourceStatusUpdateDraft,
		"creator",
		"updater",
	)

	assert.NotNil(t, resource)
	assert.Equal(t, "route-id", resource.ID)
	assert.Equal(t, 1001, resource.GatewayID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, resource.Status)
	assert.Equal(t, "creator", resource.Creator)
	assert.Equal(t, "updater", resource.Updater)
	assert.JSONEq(t, `{"uris":["/test"]}`, string(resource.Config))
	assert.Equal(t, "route-a", resource.NameValue)
	assert.Equal(t, "svc-a", resource.ServiceIDValue)
}
