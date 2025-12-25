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

package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestExistsGatewayName(t *testing.T) {
	// 初始化内存数据库
	util.InitEmbedDb()
	ctx := context.Background()

	// 创建测试网关
	gateway1 := data.Gateway1WithBkAPISIX()
	gateway1.Name = "test-gateway-1"
	err := CreateGateway(ctx, gateway1)
	assert.NoError(t, err)

	gateway2 := data.Gateway1WithBkAPISIX()
	gateway2.Name = "test-gateway-2"
	err = CreateGateway(ctx, gateway2)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		ctx         context.Context
		gatewayName string
		id          int
		expected    bool
	}{
		{
			name:        "网关名称不存在 - 应返回false",
			ctx:         ctx,
			gatewayName: "non-existent-gateway",
			id:          0,
			expected:    false,
		},
		{
			name:        "网关名称存在 - 应返回true",
			ctx:         ctx,
			gatewayName: "test-gateway-1",
			id:          0,
			expected:    true,
		},
		{
			name:        "网关名称存在但排除自己(id=gateway1.ID) - 应返回false",
			ctx:         ctx,
			gatewayName: "test-gateway-1",
			id:          gateway1.ID,
			expected:    false,
		},
		{
			name:        "网关名称存在但排除其他网关(id=gateway2.ID) - 应返回true",
			ctx:         ctx,
			gatewayName: "test-gateway-1",
			id:          gateway2.ID,
			expected:    true,
		},
		{
			name:        "网关名称存在但排除不存在的id - 应返回true",
			ctx:         ctx,
			gatewayName: "test-gateway-1",
			id:          99999,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExistsGatewayName(tt.ctx, tt.gatewayName, tt.id)
			assert.Equal(
				t,
				tt.expected,
				result,
				"ExistsGatewayName(%q, %d) = %v, want %v",
				tt.gatewayName,
				tt.id,
				result,
				tt.expected,
			)
		})
	}
}
