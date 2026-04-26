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
	"reflect"
	"testing"

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
