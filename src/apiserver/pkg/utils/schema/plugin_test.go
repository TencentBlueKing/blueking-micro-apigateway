/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
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

package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestGetPlugins(t *testing.T) {
	tests := []struct {
		name       string
		apisixType string
		version    constant.APISIXVersion
		shouldFail bool
	}{
		{
			name:       "APISIX 3.13",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion313,
			shouldFail: false,
		},
		{
			name:       "APISIX 3.11",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion311,
			shouldFail: false,
		},
		{
			name:       "APISIX 3.2",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion32,
			shouldFail: false,
		},
		{
			name:       "APISIX 3.3",
			apisixType: constant.APISIXTypeAPISIX,
			version:    constant.APISIXVersion33,
			shouldFail: false,
		},
		{
			name:       "Invalid Version",
			apisixType: constant.APISIXTypeAPISIX,
			version:    "invalid_version",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugins, err := GetPlugins(tt.apisixType, tt.version)

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, plugins)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugins)

				// 验证插件的基本结构
				for _, v := range plugins {
					assert.NotEmpty(t, v.Name)
					assert.NotEmpty(t, v.Type)
					assert.NotNil(t, v.Example)
				}
			}
		})
	}
}
