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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestBuildConsumerGroupQuery(t *testing.T) {
	gateway := data.Gateway1WithBkAPISIX()
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	tests := []struct {
		name        string
		ctx         context.Context
		expectPanic bool
	}{
		{
			name:        "正常构建查询 - 带网关信息",
			ctx:         ctx,
			expectPanic: false,
		},
		{
			name:        "构建查询失败 - 无网关信息会panic",
			ctx:         context.Background(),
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					buildConsumerGroupQuery(tt.ctx)
				}, "没有网关信息时应该 panic")
				return
			}

			query := buildConsumerGroupQuery(tt.ctx)
			assert.NotNil(t, query, "查询对象不应该为空")
			assert.Implements(t, (*repo.IConsumerGroupDo)(nil), query, "应该实现 IConsumerGroupDo 接口")
		})
	}
}

func TestBuildConsumerGroupQueryWithTx(t *testing.T) {
	gateway := data.Gateway1WithBkAPISIX()
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	tx := repo.Q

	tests := []struct {
		name        string
		ctx         context.Context
		tx          *repo.Query
		expectPanic bool
	}{
		{
			name:        "正常构建事务查询 - 带网关信息",
			ctx:         ctx,
			tx:          tx,
			expectPanic: false,
		},
		{
			name:        "构建事务查询失败 - 无网关信息会panic",
			ctx:         context.Background(),
			tx:          tx,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					buildConsumerGroupQueryWithTx(tt.ctx, tt.tx)
				}, "没有网关信息时应该 panic")
				return
			}

			query := buildConsumerGroupQueryWithTx(tt.ctx, tt.tx)
			assert.NotNil(t, query, "事务查询对象不应该为空")
			assert.Implements(t, (*repo.IConsumerGroupDo)(nil), query, "应该实现 IConsumerGroupDo 接口")
		})
	}
}
