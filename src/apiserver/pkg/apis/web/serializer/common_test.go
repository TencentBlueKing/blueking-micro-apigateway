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
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

var webSerializerTestOnce sync.Once

func initWebSerializerTestEnv() {
	webSerializerTestOnce.Do(func() {
		gin.SetMode(gin.TestMode)
		util.InitEmbedDb()
		validation.RegisterValidator()
	})
}

func newWebSerializerValidationContext(t *testing.T, gateway *model.Gateway) *gin.Context {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	ginx.SetGatewayInfo(c, gateway)
	ginx.SetValidateErrorInfo(c)
	return c
}

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
			got := injectGeneratedIDForValidation(tt.rawConfig, tt.resourceType, tt.version, tt.resourceID)

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

func TestCheckAPISIXConfigCurrentSeams(t *testing.T) {
	initWebSerializerTestEnv()

	t.Run("identity fallback uses outer name when config has no id", func(t *testing.T) {
		t.Parallel()

		type validationTarget struct {
			Name   string          `validate:"required"`
			Config json.RawMessage `validate:"apisixConfig=web_validation_identity_probe"`
		}

		ctx := newWebSerializerValidationContext(t, &model.Gateway{ID: 1101, APISIXVersion: "3.13.0"})
		req := validationTarget{
			Name:   "identity-probe",
			Config: json.RawMessage(`{"plugins":{}}`),
		}

		err := validation.ValidateStruct(ctx.Request.Context(), &req)
		assert.Error(t, err)
		assert.Contains(t, ginx.GetValidateErrorInfoFromContext(ctx.Request.Context()).Err.Error(), "resource:identity-probe validate failed")
	})

	t.Run("consumer group validation accepts generated id and outer name on 3.13", func(t *testing.T) {
		t.Parallel()

		ctx := newWebSerializerValidationContext(t, &model.Gateway{ID: 1102, APISIXVersion: "3.13.0"})
		req := ConsumerGroupInfo{
			ID:   "cg-generated-id",
			Name: "consumer-group-probe",
			Config: json.RawMessage(`{
				"plugins": {
					"limit-count": {
						"count": 100,
						"time_window": 60,
						"key": "remote_addr",
						"policy": "local"
					}
				}
			}`),
		}

		assert.NoError(t, validation.ValidateStruct(ctx.Request.Context(), &req))
	})

	t.Run("plugin metadata validation uses outer name as schema id", func(t *testing.T) {
		t.Parallel()

		ctx := newWebSerializerValidationContext(t, &model.Gateway{ID: 1103, APISIXVersion: "3.13.0"})
		req := PluginMetadataInfo{
			ID:   "existing-plugin-metadata-id",
			Name: "authz-casbin",
			Config: json.RawMessage(`{
				"model": "rbac_model.conf",
				"policy": "rbac_policy.csv"
			}`),
		}

		assert.NoError(t, validation.ValidateStruct(ctx.Request.Context(), &req))
	})

	t.Run("ssl validation succeeds without injecting synthetic name into payload", func(t *testing.T) {
		t.Parallel()

		ctx := newWebSerializerValidationContext(t, &model.Gateway{ID: 1104, APISIXVersion: "3.13.0"})
		sslFixture := data.SSL1(&model.Gateway{ID: 1104}, constant.ResourceStatusCreateDraft)
		req := SSLInfo{
			Name:   "ssl-validation-probe",
			Config: json.RawMessage(sslFixture.Config),
		}

		assert.NoError(t, validation.ValidateStruct(ctx.Request.Context(), &req))
	})
}

func TestResolveWebValidationIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            webValidationInput
		wantIdentity     string
		wantUsedFallback bool
	}{
		{
			name: "falls back to provided identity when config id is absent",
			input: webValidationInput{
				RawConfig:        json.RawMessage(`{"plugins":{}}`),
				FallbackIdentity: "route-a",
			},
			wantIdentity:     "route-a",
			wantUsedFallback: true,
		},
		{
			name: "existing config id wins",
			input: webValidationInput{
				RawConfig:        json.RawMessage(`{"id":"route-fixed","plugins":{}}`),
				FallbackIdentity: "route-a",
			},
			wantIdentity:     "route-fixed",
			wantUsedFallback: false,
		},
		{
			name: "empty fallback is preserved when no config id exists",
			input: webValidationInput{
				RawConfig: json.RawMessage(`{"plugins":{}}`),
			},
			wantIdentity:     "",
			wantUsedFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIdentity, gotUsedFallback := resolveWebValidationIdentity(tt.input)
			assert.Equal(t, tt.wantIdentity, gotIdentity)
			assert.Equal(t, tt.wantUsedFallback, gotUsedFallback)
		})
	}
}

func TestPrepareWebValidationPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        webValidationInput
		wantPayload  string
		wantIdentity string
	}{
		{
			name: "consumer group injects generated id and then uses that id as identity on 3.13",
			input: webValidationInput{
				ResourceType:     constant.ConsumerGroup,
				Version:          constant.APISIXVersion313,
				ResourceID:       "cg-generated-id",
				FallbackIdentity: "cg-demo",
				RawConfig:        json.RawMessage(`{"plugins":{}}`),
			},
			wantPayload:  `{"plugins":{},"id":"cg-generated-id"}`,
			wantIdentity: "cg-generated-id",
		},
		{
			name: "proto on 3.11 keeps name out of payload",
			input: webValidationInput{
				ResourceType:     constant.Proto,
				Version:          constant.APISIXVersion311,
				FallbackIdentity: "proto-demo",
				RawConfig:        json.RawMessage(`{"content":"syntax = \"proto3\";"}`),
			},
			wantPayload:  `{"content":"syntax = \"proto3\";"}`,
			wantIdentity: "proto-demo",
		},
		{
			name: "plugin metadata uses outer name as id on update-like input",
			input: webValidationInput{
				ResourceType:     constant.PluginMetadata,
				Version:          constant.APISIXVersion313,
				ResourceID:       "existing-plugin-metadata-id",
				Name:             "authz-casbin",
				FallbackIdentity: "authz-casbin",
				RawConfig: json.RawMessage(`{
					"model": "rbac_model.conf",
					"policy": "rbac_policy.csv"
				}`),
			},
			wantPayload: `{
				"model": "rbac_model.conf",
				"policy": "rbac_policy.csv",
				"id": "authz-casbin"
			}`,
			wantIdentity: "authz-casbin",
		},
		{
			name: "ssl never injects name",
			input: webValidationInput{
				ResourceType:     constant.SSL,
				Version:          constant.APISIXVersion313,
				FallbackIdentity: "ssl-demo",
				RawConfig:        json.RawMessage(`{"cert":"demo","key":"demo","snis":["demo.com"]}`),
			},
			wantPayload:  `{"cert":"demo","key":"demo","snis":["demo.com"]}`,
			wantIdentity: "ssl-demo",
		},
		{
			name: "existing config id stays authoritative when fallback is empty",
			input: webValidationInput{
				ResourceType: constant.ConsumerGroup,
				Version:      constant.APISIXVersion313,
				ResourceID:   "cg-generated-id",
				RawConfig:    json.RawMessage(`{"id":"cg-fixed","plugins":{}}`),
			},
			wantPayload:  `{"id":"cg-fixed","plugins":{}}`,
			wantIdentity: "cg-fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPayload, gotIdentity := prepareWebValidationPayload(tt.input)
			assert.JSONEq(t, tt.wantPayload, string(gotPayload))
			assert.Equal(t, tt.wantIdentity, gotIdentity)
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
			got := shouldInjectResourceNameForValidation(tt.resourceType, tt.version)
			if got != tt.want {
				t.Fatalf("unexpected result: got %v want %v", got, tt.want)
			}
		})
	}
}
