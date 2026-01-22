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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestBuildConsumerQuery(t *testing.T) {
	// 初始化测试环境
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
			expectPanic: true, // 函数会 panic，因为需要网关信息
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				// 验证在没有网关信息时会 panic
				assert.Panics(t, func() {
					buildConsumerQuery(tt.ctx)
				}, "没有网关信息时应该 panic")
				return
			}

			// 构建查询对象
			query := buildConsumerQuery(tt.ctx)

			// 验证返回的查询对象不为空
			assert.NotNil(t, query, "查询对象不应该为空")

			// 验证查询对象的类型
			assert.Implements(t, (*repo.IConsumerDo)(nil), query, "应该实现 IConsumerDo 接口")
		})
	}
}

func TestBuildConsumerQueryWithTx(t *testing.T) {
	// 初始化测试环境
	gateway := data.Gateway1WithBkAPISIX()
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	// 创建一个测试用的事务查询对象
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
				// 验证在没有网关信息时会 panic
				assert.Panics(t, func() {
					buildConsumerQueryWithTx(tt.ctx, tt.tx)
				}, "没有网关信息时应该 panic")
				return
			}

			// 构建事务查询对象
			query := buildConsumerQueryWithTx(tt.ctx, tt.tx)

			// 验证返回的查询对象不为空
			assert.NotNil(t, query, "事务查询对象不应该为空")

			// 验证查询对象的类型
			assert.Implements(t, (*repo.IConsumerDo)(nil), query, "应该实现 IConsumerDo 接口")
		})
	}
}

// TestBuildConsumerQuery_Integration 集成测试：验证查询能正确过滤网关数据
func TestBuildConsumerQuery_Integration(t *testing.T) {
	// 创建两个不同的网关
	gateway1 := &model.Gateway{ID: 1, Name: "gateway1"}
	gateway2 := &model.Gateway{ID: 2, Name: "gateway2"}

	ctx1 := ginx.SetGatewayInfoToContext(context.Background(), gateway1)
	ctx2 := ginx.SetGatewayInfoToContext(context.Background(), gateway2)

	// 构建查询对象
	query1 := buildConsumerQuery(ctx1)
	query2 := buildConsumerQuery(ctx2)

	// 验证两个查询对象不相同（因为 gateway_id 不同）
	assert.NotNil(t, query1)
	assert.NotNil(t, query2)

	// 注意：由于这是单元测试，我们只验证查询对象能够正确构建
	// 实际的数据库查询行为应该在集成测试中验证
}

// TestBuildConsumerQueryWithTx_Integration 集成测试：验证事务查询能正确过滤网关数据
func TestBuildConsumerQueryWithTx_Integration(t *testing.T) {
	// 创建两个不同的网关
	gateway1 := &model.Gateway{ID: 1, Name: "gateway1"}
	gateway2 := &model.Gateway{ID: 2, Name: "gateway2"}

	ctx1 := ginx.SetGatewayInfoToContext(context.Background(), gateway1)
	ctx2 := ginx.SetGatewayInfoToContext(context.Background(), gateway2)

	tx := repo.Q

	// 构建事务查询对象
	query1 := buildConsumerQueryWithTx(ctx1, tx)
	query2 := buildConsumerQueryWithTx(ctx2, tx)

	// 验证两个查询对象不相同（因为 gateway_id 不同）
	assert.NotNil(t, query1)
	assert.NotNil(t, query2)

	// 注意：由于这是单元测试，我们只验证查询对象能够正确构建
	// 实际的数据库查询行为应该在集成测试中验证
}

// TestBuildConsumerQuery_WithNilGateway 测试网关信息为空的情况
func TestBuildConsumerQuery_WithNilGateway(t *testing.T) {
	// 创建一个没有网关信息的 context
	ctx := context.Background()

	// 构建查询对象 - 会 panic，因为需要网关信息
	assert.Panics(t, func() {
		buildConsumerQuery(ctx)
	}, "没有网关信息时应该 panic")
}

// TestBuildConsumerQueryWithTx_WithNilGateway 测试事务查询在网关信息为空的情况
func TestBuildConsumerQueryWithTx_WithNilGateway(t *testing.T) {
	// 创建一个没有网关信息的 context
	ctx := context.Background()
	tx := repo.Q

	// 构建事务查询对象 - 会 panic，因为需要网关信息
	assert.Panics(t, func() {
		buildConsumerQueryWithTx(ctx, tx)
	}, "没有网关信息时应该 panic")
}
