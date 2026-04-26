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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestInjectGeneratedIDForValidation(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		resourceID   string
		rawConfig    json.RawMessage
		wantConfig   string
	}{
		{
			name:         "inject generated id for consumer group",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			resourceID:   "cg-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{},"id":"cg-generated-id"}`,
		},
		{
			name:         "inject generated id for plugin config",
			resourceType: constant.PluginConfig,
			version:      constant.APISIXVersion311,
			resourceID:   "pc-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{},"id":"pc-generated-id"}`,
		},
		{
			name:         "inject generated id for global rule",
			resourceType: constant.GlobalRule,
			version:      constant.APISIXVersion313,
			resourceID:   "gr-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{"ip-restriction":{}}}`),
			wantConfig:   `{"plugins":{"ip-restriction":{}},"id":"gr-generated-id"}`,
		},
		{
			name:         "do not inject id for old consumer group schema",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion33,
			resourceID:   "cg-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{}}`,
		},
		{
			name:         "keep existing id",
			resourceType: constant.GlobalRule,
			version:      constant.APISIXVersion313,
			resourceID:   "gr-generated-id",
			rawConfig:    json.RawMessage(`{"id":"client-id","plugins":{}}`),
			wantConfig:   `{"id":"client-id","plugins":{}}`,
		},
		{
			name:         "do not inject for consumer",
			resourceType: constant.Consumer,
			version:      constant.APISIXVersion313,
			resourceID:   "consumer-generated-id",
			rawConfig:    json.RawMessage(`{"username":"demo"}`),
			wantConfig:   `{"username":"demo"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resourcecodec.PrepareValidationPayload(resourcecodec.ValidationInput{
				Source:       resourcecodec.SourceWeb,
				ResourceType: tt.resourceType,
				Version:      tt.version,
				Config:       tt.rawConfig,
				ResourceID:   tt.resourceID,
			})

			var gotObj any
			if err := json.Unmarshal(got, &gotObj); err != nil {
				t.Fatalf("unmarshal got config failed: %v", err)
			}

			var wantObj any
			if err := json.Unmarshal([]byte(tt.wantConfig), &wantObj); err != nil {
				t.Fatalf("unmarshal want config failed: %v", err)
			}

			if !reflect.DeepEqual(gotObj, wantObj) {
				t.Fatalf("unexpected config: got %s want %s", string(got), tt.wantConfig)
			}
		})
	}
}

func TestShouldInjectResourceNameForValidation(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		want         bool
	}{
		{
			name:         "inject consumer username",
			resourceType: constant.Consumer,
			version:      constant.APISIXVersion313,
			want:         true,
		},
		{
			name:         "inject route name",
			resourceType: constant.Route,
			version:      constant.APISIXVersion311,
			want:         true,
		},
		{
			name:         "do not inject ssl name",
			resourceType: constant.SSL,
			version:      constant.APISIXVersion313,
			want:         false,
		},
		{
			name:         "do not inject proto name on old schema",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion311,
			want:         false,
		},
		{
			name:         "inject proto name on 3.13",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion313,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resourcecodec.PrepareValidationPayload(resourcecodec.ValidationInput{
				Source:       resourcecodec.SourceWeb,
				ResourceType: tt.resourceType,
				Version:      tt.version,
				Config:       json.RawMessage(`{}`),
				OuterName:    "demo-name",
			})
			hasName := false
			var parsed map[string]any
			if err := json.Unmarshal(got, &parsed); err != nil {
				t.Fatalf("unmarshal got config failed: %v", err)
			}
			if tt.resourceType == constant.Consumer {
				_, hasName = parsed["username"]
			} else {
				_, hasName = parsed["name"]
			}
			gotResult := hasName
			if gotResult != tt.want {
				t.Fatalf("unexpected result: got %v want %v", gotResult, tt.want)
			}
		})
	}
}

func TestWebAndOpenAPIConflictParity(t *testing.T) {
	t.Parallel()
	validation.RegisterValidator()

	t.Run("web route validation rejects conflicting request and config names", func(t *testing.T) {
		type request struct {
			ID     string          `json:"id"`
			Name   string          `json:"name" validate:"required"`
			Config json.RawMessage `json:"config" validate:"apisixConfig=route"`
		}

		gateway := data.Gateway1WithBkAPISIX()
		gateway.ID = 1001
		router := gin.New()
		router.POST("/test", func(c *gin.Context) {
			ginx.SetGatewayInfo(c, gateway)
			ginx.SetValidateErrorInfo(c)

			var req request
			err := validation.BindAndValidate(c, &req)
			if err != nil {
				validateErr := ginx.GetValidateErrorInfoFromContext(c.Request.Context())
				if validateErr != nil && validateErr.Err != nil {
					c.String(http.StatusBadRequest, validateErr.Err.Error())
					return
				}
				c.String(http.StatusBadRequest, err.Error())
				return
			}
			c.Status(http.StatusNoContent)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"/test",
			bytes.NewBufferString(`{"name":"route-a","config":{"name":"route-b","uris":["/test"]}}`),
		)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "conflicts with request name")
	})
}

func TestWebValidationDraftParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        resourcecodec.RequestInput
		wantDatabase string
	}{
		{
			name: "consumer uses outer name as authoritative username",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceWeb,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.Consumer,
				Version:      constant.APISIXVersion313,
				OuterName:    "consumer-a",
				OuterFields:  map[string]any{"group_id": "group-a"},
				Config:       json.RawMessage(`{"plugins":{"key-auth":{"key":"token-a"}}}`),
			},
			wantDatabase: `{"username":"consumer-a","group_id":"group-a","plugins":{"key-auth":{"key":"token-a"}}}`,
		},
		{
			name: "consumer group 3.11 injects id but not name",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceWeb,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.ConsumerGroup,
				Version:      constant.APISIXVersion311,
				OuterName:    "group-a",
				Config:       json.RawMessage(`{"plugins":{}}`),
			},
			wantDatabase: `{"id":"@nonempty","plugins":{}}`,
		},
		{
			name: "plugin metadata derives id from outer name",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceWeb,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.PluginMetadata,
				Version:      constant.APISIXVersion313,
				OuterName:    "jwt-auth",
				Config:       json.RawMessage(`{"key":"value"}`),
			},
			wantDatabase: `{"id":"jwt-auth","key":"value"}`,
		},
		{
			name: "upstream uses ssl outer field as tls client cert id",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceWeb,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.Upstream,
				Version:      constant.APISIXVersion313,
				OuterName:    "upstream-a",
				OuterFields:  map[string]any{"tls.client_cert_id": "ssl-a"},
				Config: json.RawMessage(
					`{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}`,
				),
			},
			wantDatabase: `{"name":"upstream-a","tls":{"client_cert_id":"ssl-a"},"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft, err := resourcecodec.PrepareRequestDraft(tt.input)
			assert.NoError(t, err)

			builtPayload, err := resourcecodec.BuildRequestPayload(draft, constant.DATABASE)
			assert.NoError(t, err)
			assertJSONWithGeneratedID(t, tt.wantDatabase, string(builtPayload.Payload))
		})
	}
}

func assertJSONWithGeneratedID(t *testing.T, want, got string) {
	t.Helper()

	var gotObj map[string]any
	assert.NoError(t, json.Unmarshal([]byte(got), &gotObj))

	var wantObj map[string]any
	assert.NoError(t, json.Unmarshal([]byte(want), &wantObj))

	if wantID, ok := wantObj["id"].(string); ok && wantID == "@nonempty" {
		gotID, ok := gotObj["id"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, gotID)
		delete(wantObj, "id")
		delete(gotObj, "id")
	}

	gotNormalized, err := json.Marshal(gotObj)
	assert.NoError(t, err)
	wantNormalized, err := json.Marshal(wantObj)
	assert.NoError(t, err)
	assert.JSONEq(t, string(wantNormalized), string(gotNormalized))
}
