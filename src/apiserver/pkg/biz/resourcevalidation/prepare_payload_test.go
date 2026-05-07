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

package resourcevalidation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
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
			name:         "injects id when schema requires it",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			resourceID:   "cg-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{},"id":"cg-generated-id"}`,
		},
		{
			name:         "keeps existing id",
			resourceType: constant.GlobalRule,
			version:      constant.APISIXVersion313,
			resourceID:   "gr-generated-id",
			rawConfig:    json.RawMessage(`{"id":"client-id","plugins":{}}`),
			wantConfig:   `{"id":"client-id","plugins":{}}`,
		},
		{
			name:         "skips injection when resource id is empty",
			resourceType: constant.PluginConfig,
			version:      constant.APISIXVersion311,
			resourceID:   "",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{}}`,
		},
		{
			name:         "skips injection for versions that do not require id",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion33,
			resourceID:   "cg-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectGeneratedIDForValidation(tt.version, tt.resourceType, tt.rawConfig, tt.resourceID)
			assert.JSONEq(t, tt.wantConfig, string(got))
		})
	}
}

func TestBuildConfigRawForValidation(t *testing.T) {
	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		configRaw    string
		wantConfig   string
	}{
		{
			name:         "consumer strips id and name",
			resourceType: constant.Consumer,
			version:      constant.APISIXVersion313,
			configRaw:    `{"id":"consumer-id","name":"consumer-demo","username":"consumer-demo","plugins":{}}`,
			wantConfig:   `{"username":"consumer-demo","plugins":{}}`,
		},
		{
			name:         "proto 3.11 strips unsupported name",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion311,
			configRaw:    `{"name":"proto-demo","content":"syntax = \"proto3\";"}`,
			wantConfig:   `{"content":"syntax = \"proto3\";"}`,
		},
		{
			name:         "route keeps supported name",
			resourceType: constant.Route,
			version:      constant.APISIXVersion313,
			configRaw:    `{"id":"route-id","name":"route-demo","uri":"/demo"}`,
			wantConfig:   `{"id":"route-id","name":"route-demo","uri":"/demo"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildConfigRawForValidation(tt.version, tt.resourceType, tt.configRaw)

			assert.JSONEq(t, tt.wantConfig, string(got))
		})
	}
}

func TestPrepareOpenValidationPayload(t *testing.T) {
	tests := []struct {
		name          string
		resourceType  constant.APISIXResource
		version       constant.APISIXVersion
		configRaw     string
		assertPayload func(t *testing.T, payload string)
	}{
		{
			name:         "consumer group injects temporary id on 3.13",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			configRaw:    `{"plugins":{}}`,
			assertPayload: func(t *testing.T, payload string) {
				t.Helper()
				assert.NotEmpty(t, gjson.Get(payload, "id").String())
			},
		},
		{
			name:         "existing id is preserved during validation payload preparation",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			configRaw:    `{"id":"client-id","plugins":{}}`,
			assertPayload: func(t *testing.T, payload string) {
				t.Helper()
				assert.Equal(t, "client-id", gjson.Get(payload, "id").String())
			},
		},
		{
			name:         "proto on 3.11 strips unsupported name before validation",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion311,
			configRaw:    `{"name":"proto-demo","content":"syntax = \"proto3\";"}`,
			assertPayload: func(t *testing.T, payload string) {
				t.Helper()
				assert.False(t, gjson.Get(payload, "name").Exists())
				assert.Equal(t, `syntax = "proto3";`, gjson.Get(payload, "content").String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := PrepareOpenValidationPayload(tt.version, tt.resourceType, tt.configRaw)
			tt.assertPayload(t, string(payload))
		})
	}
}

func TestPrepareImportValidationPayload(t *testing.T) {
	t.Parallel()

	payload := PrepareImportValidationPayload(
		constant.APISIXVersion313,
		constant.Consumer,
		`{"id":"consumer-id","name":"consumer-demo","username":"consumer-demo","plugins":{}}`,
	)

	assert.JSONEq(t, `{"username":"consumer-demo","plugins":{}}`, string(payload))
}

func TestResolveWebValidationIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		configRaw        json.RawMessage
		fallbackIdentity string
		wantIdentity     string
		wantUsedFallback bool
	}{
		{
			name:             "falls back to provided identity when config id is absent",
			configRaw:        json.RawMessage(`{"plugins":{}}`),
			fallbackIdentity: "route-a",
			wantIdentity:     "route-a",
			wantUsedFallback: true,
		},
		{
			name:             "existing config id wins",
			configRaw:        json.RawMessage(`{"id":"route-fixed","plugins":{}}`),
			fallbackIdentity: "route-a",
			wantIdentity:     "route-fixed",
			wantUsedFallback: false,
		},
		{
			name:             "empty fallback is preserved when no config id exists",
			configRaw:        json.RawMessage(`{"plugins":{}}`),
			wantIdentity:     "",
			wantUsedFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIdentity, gotUsedFallback := resolveWebValidationIdentity(tt.configRaw, tt.fallbackIdentity)
			assert.Equal(t, tt.wantIdentity, gotIdentity)
			assert.Equal(t, tt.wantUsedFallback, gotUsedFallback)
		})
	}
}

func TestPrepareWebValidationPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		resourceType     constant.APISIXResource
		version          constant.APISIXVersion
		configRaw        string
		resourceID       string
		nameValue        string
		fallbackIdentity string
		wantPayload      string
		wantIdentity     string
	}{
		{
			name:             "consumer group injects generated id and then uses that id as identity on 3.13",
			resourceType:     constant.ConsumerGroup,
			version:          constant.APISIXVersion313,
			configRaw:        `{"plugins":{}}`,
			resourceID:       "cg-generated-id",
			fallbackIdentity: "cg-demo",
			wantPayload:      `{"plugins":{},"id":"cg-generated-id"}`,
			wantIdentity:     "cg-generated-id",
		},
		{
			name:             "proto on 3.11 keeps name out of payload",
			resourceType:     constant.Proto,
			version:          constant.APISIXVersion311,
			configRaw:        `{"content":"syntax = \"proto3\";"}`,
			fallbackIdentity: "proto-demo",
			wantPayload:      `{"content":"syntax = \"proto3\";"}`,
			wantIdentity:     "proto-demo",
		},
		{
			name:             "route uses fallback identity as validation name",
			resourceType:     constant.Route,
			version:          constant.APISIXVersion313,
			configRaw:        `{"uri":"/demo"}`,
			fallbackIdentity: "route-demo",
			wantPayload:      `{"uri":"/demo","name":"route-demo"}`,
			wantIdentity:     "route-demo",
		},
		{
			name:             "consumer uses fallback identity as validation username",
			resourceType:     constant.Consumer,
			version:          constant.APISIXVersion313,
			configRaw:        `{"plugins":{}}`,
			fallbackIdentity: "consumer-demo",
			wantPayload:      `{"plugins":{},"username":"consumer-demo"}`,
			wantIdentity:     "consumer-demo",
		},
		{
			name:             "plugin metadata uses outer name as id on update-like input",
			resourceType:     constant.PluginMetadata,
			version:          constant.APISIXVersion313,
			configRaw:        `{"model":"rbac_model.conf","policy":"rbac_policy.csv"}`,
			resourceID:       "existing-plugin-metadata-id",
			nameValue:        "authz-casbin",
			fallbackIdentity: "authz-casbin",
			wantPayload:      `{"model":"rbac_model.conf","policy":"rbac_policy.csv","id":"authz-casbin"}`,
			wantIdentity:     "authz-casbin",
		},
		{
			name:             "ssl never injects name",
			resourceType:     constant.SSL,
			version:          constant.APISIXVersion313,
			configRaw:        `{"cert":"demo","key":"demo","snis":["demo.com"]}`,
			fallbackIdentity: "ssl-demo",
			wantPayload:      `{"cert":"demo","key":"demo","snis":["demo.com"]}`,
			wantIdentity:     "ssl-demo",
		},
		{
			name:         "existing config id stays authoritative when fallback is empty",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			configRaw:    `{"id":"cg-fixed","plugins":{}}`,
			resourceID:   "cg-generated-id",
			wantPayload:  `{"id":"cg-fixed","plugins":{}}`,
			wantIdentity: "cg-fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPayload, gotIdentity := PrepareWebValidationPayload(
				tt.version,
				tt.resourceType,
				tt.configRaw,
				tt.resourceID,
				tt.nameValue,
				tt.fallbackIdentity,
			)
			assert.JSONEq(t, tt.wantPayload, string(gotPayload))
			assert.Equal(t, tt.wantIdentity, gotIdentity)
		})
	}
}

func TestPrepareMCPDatabaseValidationPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		resourceID   string
		nameValue    string
		configRaw    string
		assertConfig func(t *testing.T, config []byte)
	}{
		{
			name:         "consumer group injects generated id",
			resourceType: constant.ConsumerGroup,
			resourceID:   "consumer-group-id",
			nameValue:    "consumer-group-demo",
			configRaw:    `{"plugins":{}}`,
			assertConfig: func(t *testing.T, config []byte) {
				assert.Equal(t, "consumer-group-id", gjson.GetBytes(config, "id").String())
				assert.False(t, gjson.GetBytes(config, "name").Exists())
			},
		},
		{
			name:         "route injects name when config has no identity",
			resourceType: constant.Route,
			resourceID:   "route-id",
			nameValue:    "route-demo",
			configRaw:    `{"uri":"/demo"}`,
			assertConfig: func(t *testing.T, config []byte) {
				assert.Equal(t, "route-demo", gjson.GetBytes(config, "name").String())
			},
		},
		{
			name:         "global rule strips persisted name before validation",
			resourceType: constant.GlobalRule,
			resourceID:   "global-rule-id",
			nameValue:    "global-rule-demo",
			configRaw:    `{"name":"global-rule-demo","plugins":{}}`,
			assertConfig: func(t *testing.T, config []byte) {
				assert.Equal(t, "global-rule-id", gjson.GetBytes(config, "id").String())
				assert.False(t, gjson.GetBytes(config, "name").Exists())
			},
		},
		{
			name:         "consumer strips id and name before validation",
			resourceType: constant.Consumer,
			resourceID:   "consumer-id",
			nameValue:    "consumer-demo",
			configRaw:    `{"id":"consumer-id","name":"consumer-demo","username":"consumer-demo","plugins":{}}`,
			assertConfig: func(t *testing.T, config []byte) {
				assert.False(t, gjson.GetBytes(config, "id").Exists())
				assert.False(t, gjson.GetBytes(config, "name").Exists())
				assert.Equal(t, "consumer-demo", gjson.GetBytes(config, "username").String())
			},
		},
		{
			name:         "plugin metadata uses outer name as validation id",
			resourceType: constant.PluginMetadata,
			resourceID:   "plugin-metadata-id",
			nameValue:    "proxy-rewrite",
			configRaw:    `{"value":{"regex_uri":["^/old","/new"]}}`,
			assertConfig: func(t *testing.T, config []byte) {
				assert.Equal(t, "proxy-rewrite", gjson.GetBytes(config, "id").String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalConfig := tt.configRaw

			config, err := PrepareMCPDatabaseValidationPayload(
				constant.APISIXVersion313,
				tt.resourceType,
				tt.configRaw,
				tt.resourceID,
				tt.nameValue,
			)

			assert.NoError(t, err)
			tt.assertConfig(t, config)
			assert.JSONEq(t, originalConfig, tt.configRaw)
		})
	}
}

func TestPrepareMCPDatabaseValidationPayloadErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		configRaw    string
		resourceID   string
		nameValue    string
		wantErr      string
	}{
		{
			name:         "returns name injection error",
			resourceType: constant.Route,
			configRaw:    `[]`,
			resourceID:   "route-id",
			nameValue:    "route-demo",
			wantErr:      "failed to inject name into validation payload",
		},
		{
			name:         "returns plugin metadata id injection error",
			resourceType: constant.PluginMetadata,
			configRaw:    `[]`,
			resourceID:   "plugin-metadata-id",
			nameValue:    "proxy-rewrite",
			wantErr:      "failed to inject plugin metadata id into validation payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := PrepareMCPDatabaseValidationPayload(
				constant.APISIXVersion313,
				tt.resourceType,
				tt.configRaw,
				tt.resourceID,
				tt.nameValue,
			)

			assert.Nil(t, config)
			assert.ErrorContains(t, err, tt.wantErr)
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
			got := shouldInjectResourceNameForValidation(tt.version, tt.resourceType)
			assert.Equal(t, tt.want, got)
		})
	}
}
