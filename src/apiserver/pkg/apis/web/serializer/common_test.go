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
		resourceID   string
		rawConfig    json.RawMessage
		wantConfig   string
	}{
		{
			name:         "inject generated id for consumer group",
			resourceType: constant.ConsumerGroup,
			resourceID:   "cg-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{},"id":"cg-generated-id"}`,
		},
		{
			name:         "inject generated id for plugin config",
			resourceType: constant.PluginConfig,
			resourceID:   "pc-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{},"id":"pc-generated-id"}`,
		},
		{
			name:         "keep existing id",
			resourceType: constant.GlobalRule,
			resourceID:   "gr-generated-id",
			rawConfig:    json.RawMessage(`{"id":"client-id","plugins":{}}`),
			wantConfig:   `{"id":"client-id","plugins":{}}`,
		},
		{
			name:         "do not inject for consumer",
			resourceType: constant.Consumer,
			resourceID:   "consumer-generated-id",
			rawConfig:    json.RawMessage(`{"username":"demo"}`),
			wantConfig:   `{"username":"demo"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectGeneratedIDForValidation(tt.rawConfig, tt.resourceType, tt.resourceID)

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
