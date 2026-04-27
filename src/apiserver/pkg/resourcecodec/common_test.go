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

package resourcecodec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestDetectRequestConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     ConflictInput
		wantError string
	}{
		{
			name: "route accepts matching outer and config name",
			input: ConflictInput{
				ResourceType: constant.Route,
				OuterName:    "route-a",
				Config:       json.RawMessage(`{"name":"route-a","uris":["/test"]}`),
			},
		},
		{
			name: "route rejects conflicting outer and config name",
			input: ConflictInput{
				ResourceType: constant.Route,
				OuterName:    "route-a",
				Config:       json.RawMessage(`{"name":"route-b","uris":["/test"]}`),
			},
			wantError: "conflicts with request name",
		},
		{
			name: "route rejects conflicting service association",
			input: ConflictInput{
				ResourceType: constant.Route,
				OuterFields:  map[string]string{"service_id": "svc-a"},
				Config:       json.RawMessage(`{"service_id":"svc-b","uris":["/test"]}`),
			},
			wantError: "service_id",
		},
		{
			name: "consumer rejects conflicting username",
			input: ConflictInput{
				ResourceType: constant.Consumer,
				OuterName:    "consumer-a",
				Config:       json.RawMessage(`{"username":"consumer-b"}`),
			},
			wantError: "username",
		},
		{
			name: "plugin metadata rejects config id conflict",
			input: ConflictInput{
				ResourceType: constant.PluginMetadata,
				OuterName:    "jwt-auth",
				Config:       json.RawMessage(`{"id":"basic-auth"}`),
			},
			wantError: "plugin_metadata.id",
		},
		{
			name: "update rejects config id mismatch",
			input: ConflictInput{
				ResourceType: constant.Route,
				ResourceID:   "route-a",
				Config:       json.RawMessage(`{"id":"route-b","uris":["/test"]}`),
			},
			wantError: "config.id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DetectRequestConflicts(tt.input)
			if tt.wantError == "" {
				assert.NoError(t, err)
				return
			}
			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), tt.wantError)
			}
		})
	}
}

func TestResolveRequestIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      RequestInput
		assertions func(t *testing.T, identity ResolvedIdentity)
	}{
		{
			name: "web create generates id once for plugin config",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.PluginConfig,
				Version:      constant.APISIXVersion311,
				OuterName:    "pc-a",
				Config:       json.RawMessage(`{"plugins":{}}`),
			},
			assertions: func(t *testing.T, identity ResolvedIdentity) {
				assert.NotEmpty(t, identity.ResourceID)
				assert.Equal(t, "generated", identity.ResolvedFrom)
				assert.True(t, identity.Generated)
				assert.Equal(t, "pc-a", identity.NameValue)
			},
		},
		{
			name: "openapi create generates id for route",
			input: RequestInput{
				Source:       SourceOpenAPI,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				OuterName:    "route-a",
				Config:       json.RawMessage(`{"uris":["/test"]}`),
			},
			assertions: func(t *testing.T, identity ResolvedIdentity) {
				assert.NotEmpty(t, identity.ResourceID)
				assert.Equal(t, "generated", identity.ResolvedFrom)
				assert.Equal(t, "route-a", identity.NameValue)
			},
		},
		{
			name: "update path id wins",
			input: RequestInput{
				Source:       SourceOpenAPI,
				Operation:    constant.OperationTypeUpdate,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				PathID:       "route-path-id",
				OuterName:    "route-a",
				Config: json.RawMessage(
					`{"id":"route-config-id","name":"route-a","uris":["/test"]}`,
				),
			},
			assertions: func(t *testing.T, identity ResolvedIdentity) {
				assert.Equal(t, "route-path-id", identity.ResourceID)
				assert.Equal(t, "path", identity.ResolvedFrom)
				assert.False(t, identity.Generated)
			},
		},
		{
			name: "consumer uses outer name as username",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.Consumer,
				Version:      constant.APISIXVersion313,
				OuterName:    "consumer-a",
				Config:       json.RawMessage(`{"plugins":{"key-auth":{"key":"demo"}}}`),
			},
			assertions: func(t *testing.T, identity ResolvedIdentity) {
				assert.Equal(t, "username", identity.NameKey)
				assert.Equal(t, "consumer-a", identity.NameValue)
			},
		},
		{
			name: "plugin metadata resolves name from config id during import",
			input: RequestInput{
				Source:       SourceImport,
				Operation:    constant.OperationImport,
				ResourceType: constant.PluginMetadata,
				Version:      constant.APISIXVersion313,
				PathID:       "stored-id",
				Config:       json.RawMessage(`{"id":"jwt-auth","config":{"key":"value"}}`),
			},
			assertions: func(t *testing.T, identity ResolvedIdentity) {
				assert.Equal(t, "stored-id", identity.ResourceID)
				assert.Equal(t, "jwt-auth", identity.NameValue)
			},
		},
		{
			name: "outer associations win resolution",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				OuterName:    "route-a",
				OuterFields:  map[string]any{"service_id": "svc-outer", "plugin_config_id": "pc-outer"},
				Config: json.RawMessage(
					`{"service_id":"svc-config","plugin_config_id":"pc-config","uris":["/test"]}`,
				),
			},
			assertions: func(t *testing.T, identity ResolvedIdentity) {
				assert.Equal(t, "svc-outer", identity.Associations["service_id"])
				assert.Equal(t, "pc-outer", identity.Associations["plugin_config_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, err := ResolveRequestIdentity(tt.input)
			assert.NoError(t, err)
			tt.assertions(t, identity)
		})
	}
}

func TestPrepareRequestDraftAndBuildRequestPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        RequestInput
		wantConfig   string
		wantSpec     string
		wantErrorSub string
	}{
		{
			name: "web route create builds outer name and service id into the payload",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				OuterName:    "route-a",
				OuterFields:  map[string]any{"service_id": "svc-a"},
				Config:       json.RawMessage(`{"uris":["/test"]}`),
			},
			wantSpec:   `{"uris":["/test"]}`,
			wantConfig: `{"name":"route-a","service_id":"svc-a","uris":["/test"]}`,
		},
		{
			name: "consumer payload build uses username and strips id for database validation",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.Consumer,
				Version:      constant.APISIXVersion313,
				OuterName:    "consumer-a",
				OuterFields:  map[string]any{"group_id": "group-a"},
				Config:       json.RawMessage(`{"plugins":{"key-auth":{"key":"demo"}}}`),
			},
			wantSpec:   `{"plugins":{"key-auth":{"key":"demo"}}}`,
			wantConfig: `{"username":"consumer-a","group_id":"group-a","plugins":{"key-auth":{"key":"demo"}}}`,
		},
		{
			name: "plugin metadata payload build keeps name and derives id from it",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.PluginMetadata,
				Version:      constant.APISIXVersion313,
				OuterName:    "jwt-auth",
				Config:       json.RawMessage(`{"key":"value"}`),
			},
			wantSpec:   `{"key":"value"}`,
			wantConfig: `{"id":"jwt-auth","key":"value"}`,
		},
		{
			name: "consumer group validation drops unsupported name on 3.11",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.ConsumerGroup,
				Version:      constant.APISIXVersion311,
				OuterName:    "group-a",
				Config:       json.RawMessage(`{"plugins":{}}`),
			},
			wantSpec:   `{"plugins":{}}`,
			wantConfig: `{"id":"@nonempty","plugins":{}}`,
		},
		{
			name: "conflicting duplicated association is rejected",
			input: RequestInput{
				Source:       SourceWeb,
				Operation:    constant.OperationTypeCreate,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				OuterName:    "route-a",
				OuterFields:  map[string]any{"service_id": "svc-outer"},
				Config:       json.RawMessage(`{"service_id":"svc-config","uris":["/test"]}`),
			},
			wantErrorSub: "service_id",
		},
		{
			name: "openapi update rejects path and config id mismatch",
			input: RequestInput{
				Source:       SourceOpenAPI,
				Operation:    constant.OperationTypeUpdate,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				PathID:       "route-a",
				OuterName:    "route-a",
				Config:       json.RawMessage(`{"id":"route-b","uris":["/test"]}`),
			},
			wantErrorSub: "config.id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft, err := PrepareRequestDraft(tt.input)
			if tt.wantErrorSub != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.wantErrorSub)
				}
				return
			}
			assert.NoError(t, err)
			assertJSONWithOptionalGeneratedID(t, tt.wantSpec, string(draft.ConfigSpec))

			built, err := BuildRequestPayload(draft, constant.DATABASE)
			assert.NoError(t, err)
			assertJSONWithOptionalGeneratedID(t, tt.wantConfig, string(built.Payload))
		})
	}
}

func TestValidateBuiltPayloadShape(t *testing.T) {
	t.Parallel()

	t.Run("accepts payload already in etcd built form", func(t *testing.T) {
		built, err := ValidateBuiltPayloadShape(ValidationPayloadInput{
			ResourceType: constant.Route,
			Version:      constant.APISIXVersion311,
			Profile:      constant.ETCD,
			Payload:      json.RawMessage(`{"id":"route-id","name":"route-a","uris":["/test"]}`),
		})
		assert.NoError(t, err)
		assert.JSONEq(t, `{"id":"route-id","name":"route-a","uris":["/test"]}`, string(built.Payload))
	})

	t.Run("rejects payload that still carries database-only fields", func(t *testing.T) {
		_, err := ValidateBuiltPayloadShape(ValidationPayloadInput{
			ResourceType: constant.Consumer,
			Version:      constant.APISIXVersion313,
			Profile:      constant.ETCD,
			Payload:      json.RawMessage(`{"id":"consumer-id","username":"consumer-a"}`),
		})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "schema 验证失败")
			assert.Contains(t, err.Error(), "built form")
		}
	})
}

func assertJSONWithOptionalGeneratedID(t *testing.T, want, got string) {
	t.Helper()

	if want == "" {
		assert.JSONEq(t, want, got)
		return
	}

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
