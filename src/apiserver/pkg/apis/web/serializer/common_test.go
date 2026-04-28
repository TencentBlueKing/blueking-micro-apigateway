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
