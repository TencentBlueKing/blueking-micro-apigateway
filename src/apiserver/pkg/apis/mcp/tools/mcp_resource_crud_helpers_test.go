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

package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestPrepareMCPCreateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		inputConfig  any
		nameValue    string
		assertConfig func(t *testing.T, config []byte)
	}{
		{
			name:         "route injects name",
			resourceType: constant.Route,
			inputConfig:  map[string]any{"uri": "/demo"},
			nameValue:    "route-demo",
			assertConfig: func(t *testing.T, config []byte) {
				assert.Equal(t, "route-demo", gjson.GetBytes(config, "name").String())
				assert.Equal(t, "/demo", gjson.GetBytes(config, "uri").String())
			},
		},
		{
			name:         "consumer injects username",
			resourceType: constant.Consumer,
			inputConfig:  map[string]any{"plugins": map[string]any{}},
			nameValue:    "consumer-demo",
			assertConfig: func(t *testing.T, config []byte) {
				assert.Equal(t, "consumer-demo", gjson.GetBytes(config, "username").String())
				assert.Empty(t, gjson.GetBytes(config, "name").String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := prepareMCPCreateConfig(tt.resourceType, tt.inputConfig, tt.nameValue)
			assert.NoError(t, err)
			tt.assertConfig(t, config)
		})
	}
}
