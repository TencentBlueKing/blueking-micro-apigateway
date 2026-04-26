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
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestMaterializeStoredDraftLegacyCompatibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      StoredRowInput
		wantETCD   string
		wantLegacy bool
	}{
		{
			name: "route stored row uses authoritative columns over legacy config duplicates",
			input: StoredRowInput{
				GatewayID:    1001,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				ResourceID:   "route-col",
				NameKey:      "name",
				NameValue:    "route-col-name",
				Associations: map[string]string{
					"service_id":       "svc-col",
					"plugin_config_id": "pc-col",
				},
				Config: json.RawMessage(
					`{"id":"route-legacy","name":"route-legacy-name","service_id":"svc-legacy","plugin_config_id":"pc-legacy","uris":["/test"]}`,
				),
				LegacyDetected: true,
			},
			wantETCD:   `{"id":"route-col","name":"route-col-name","service_id":"svc-col","plugin_config_id":"pc-col","uris":["/test"]}`,
			wantLegacy: true,
		},
		{
			name: "consumer stored row keeps username and strips id for etcd payload",
			input: StoredRowInput{
				GatewayID:    1001,
				ResourceType: constant.Consumer,
				Version:      constant.APISIXVersion313,
				ResourceID:   "consumer-col",
				NameKey:      "username",
				NameValue:    "consumer-name",
				Associations: map[string]string{"group_id": "group-col"},
				Config: json.RawMessage(
					`{"id":"consumer-legacy","username":"consumer-legacy-name","group_id":"group-legacy","plugins":{"key-auth":{"key":"demo"}}}`,
				),
				LegacyDetected: true,
			},
			wantETCD:   `{"username":"consumer-name","group_id":"group-col","plugins":{"key-auth":{"key":"demo"}}}`,
			wantLegacy: true,
		},
		{
			name: "plugin metadata publish key comes from authoritative name",
			input: StoredRowInput{
				GatewayID:    1001,
				ResourceType: constant.PluginMetadata,
				Version:      constant.APISIXVersion313,
				ResourceID:   "plugin-metadata-id",
				NameKey:      "name",
				NameValue:    "jwt-auth",
				Config: json.RawMessage(
					`{"id":"basic-auth","name":"basic-auth","key":"value"}`,
				),
				LegacyDetected: true,
			},
			wantETCD:   `{"id":"jwt-auth","key":"value"}`,
			wantLegacy: true,
		},
		{
			name: "stream route 3.11 removes unsupported name and labels",
			input: StoredRowInput{
				GatewayID:    1001,
				ResourceType: constant.StreamRoute,
				Version:      constant.APISIXVersion311,
				ResourceID:   "stream-route-id",
				NameKey:      "name",
				NameValue:    "stream-route-name",
				Associations: map[string]string{"upstream_id": "up-col"},
				Config: json.RawMessage(
					`{"name":"legacy-stream","upstream_id":"up-legacy","remote_addr":"127.0.0.1","server_port":9100,"labels":{"env":"test"}}`,
				),
				Labels: map[string]string{"env": "test"},
			},
			wantETCD:   `{"id":"stream-route-id","upstream_id":"up-col","remote_addr":"127.0.0.1","server_port":9100}`,
			wantLegacy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft := DraftFromStoredRow(tt.input)
			assert.Equal(t, tt.wantLegacy, draft.LegacyEchoes)
			assert.Equal(t, tt.wantLegacy, draft.Identity.LegacyDetected)
			for _, fieldName := range codecConfigFor(tt.input.ResourceType).stripFields {
				assert.False(t, gjson.GetBytes(draft.ConfigSpec, fieldName).Exists(), fieldName)
			}

			materialized, err := MaterializeStoredDraft(draft, constant.ETCD)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.wantETCD, string(materialized.Payload))
		})
	}
}
