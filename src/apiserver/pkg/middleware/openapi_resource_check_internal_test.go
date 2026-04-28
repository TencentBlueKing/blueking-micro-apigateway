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

package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

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
			payload := prepareOpenValidationPayload(tt.resourceType, tt.version, tt.configRaw)
			tt.assertPayload(t, string(payload))
		})
	}
}
