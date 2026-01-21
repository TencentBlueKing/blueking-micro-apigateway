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

package common

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func init() {
	// 初始化加密组件，用于测试
	_ = cryptography.Init("jxi18GX5w2qgHwfZCFpn07q8FScXJOd3", "k2dbCGetyusW")
}

// createTestGateway 创建测试网关
func createTestGateway(ctx context.Context, name, endpoint, prefix string) (*model.Gateway, error) {
	gateway := &model.Gateway{
		Name:          name,
		Mode:          1,
		Maintainers:   []string{"admin"},
		Desc:          "test gateway",
		APISIXType:    constant.APISIXTypeBKAPISIX,
		APISIXVersion: "3.11.0",
		EtcdConfig: model.EtcdConfig{
			InstanceID: "",
			EtcdConfig: base.EtcdConfig{
				Endpoint: base.Endpoint(endpoint),
				Username: "test",
				Password: "test",
				Prefix:   prefix,
			},
		},
	}
	err := biz.CreateGateway(ctx, gateway)
	return gateway, err
}

// checkPrefixConflictLogic 模拟 CheckEtcdConnAndAPISIXInstance 中的 prefix 冲突检测逻辑
// 这个函数不依赖真实的 etcd 连接，只测试 prefix 冲突检测逻辑
// 使用优化后的查询方式：只查询使用相同 etcd 集群的网关
func checkPrefixConflictLogic(
	ctx context.Context,
	gatewayID int,
	newEndpoints []string,
	newPrefix string,
) error {
	for _, endpoint := range newEndpoints {
		// 去除协议前缀
		cleanEndpoint := model.RemoveEndpointProtocol(endpoint)
		if cleanEndpoint == "" {
			continue
		}

		// 查询 endpoint 包含当前地址的网关
		sameClusterGateways, err := biz.GetGatewaysByEndpointLike(ctx, cleanEndpoint, gatewayID)
		if err != nil {
			return fmt.Errorf("查询相同 etcd 集群的网关失败: %w", err)
		}

		// 检查这些网关是否有 prefix 冲突
		for _, gateway := range sameClusterGateways {
			if model.CheckEtcdPrefixConflict(gateway.EtcdConfig.Prefix, newPrefix) {
				return &prefixConflictError{
					newPrefix:      newPrefix,
					existingPrefix: gateway.EtcdConfig.Prefix,
					gatewayName:    gateway.Name,
				}
			}
		}
	}
	return nil
}

type prefixConflictError struct {
	newPrefix      string
	existingPrefix string
	gatewayName    string
}

func (e *prefixConflictError) Error() string {
	return "etcd 前缀 [" + e.newPrefix + "] 与网关 [" + e.gatewayName +
		"] 的前缀 [" + e.existingPrefix + "] 在同一 etcd 集群中存在层级冲突"
}

func TestCheckEtcdPrefixConflictLogic(t *testing.T) {
	// 初始化内存数据库
	util.InitEmbedDb()
	ctx := context.Background()

	tests := []struct {
		name             string
		existingEndpoint string
		existingPrefix   string
		newEndpoints     []string
		newPrefix        string
		expectConflict   bool
	}{
		// 同一 etcd 集群，prefix 层级冲突
		{
			name:             "同一集群-prefix层级冲突-父子关系",
			existingEndpoint: "http://etcd1:2379",
			existingPrefix:   "/apisix",
			newEndpoints:     []string{"http://etcd1:2379"},
			newPrefix:        "/apisix/gateway1",
			expectConflict:   true,
		},
		{
			name:             "同一集群-prefix层级冲突-子父关系",
			existingEndpoint: "http://etcd2:2379",
			existingPrefix:   "/apisix/gateway2",
			newEndpoints:     []string{"http://etcd2:2379"},
			newPrefix:        "/apisix",
			expectConflict:   true,
		},
		{
			name:             "同一集群-prefix完全相同",
			existingEndpoint: "http://etcd3:2379",
			existingPrefix:   "/same-prefix",
			newEndpoints:     []string{"http://etcd3:2379"},
			newPrefix:        "/same-prefix",
			expectConflict:   true,
		},
		// 同一 etcd 集群，prefix 不冲突
		{
			name:             "同一集群-名称相似但不冲突(a-b和a-b-test)",
			existingEndpoint: "http://etcd4:2379",
			existingPrefix:   "/a-b",
			newEndpoints:     []string{"http://etcd4:2379"},
			newPrefix:        "/a-b-test",
			expectConflict:   false,
		},
		{
			name:             "同一集群-同级不同名称",
			existingEndpoint: "http://etcd5:2379",
			existingPrefix:   "/apisix/gw1",
			newEndpoints:     []string{"http://etcd5:2379"},
			newPrefix:        "/apisix/gw2",
			expectConflict:   false,
		},
		{
			name:             "同一集群-完全不同prefix",
			existingEndpoint: "http://etcd6:2379",
			existingPrefix:   "/gateway-a",
			newEndpoints:     []string{"http://etcd6:2379"},
			newPrefix:        "/gateway-b",
			expectConflict:   false,
		},
		// 不同 etcd 集群，即使 prefix 有层级关系也不冲突
		{
			name:             "不同集群-即使prefix相同也允许",
			existingEndpoint: "http://etcd7:2379",
			existingPrefix:   "/apisix",
			newEndpoints:     []string{"http://etcd8:2379"},
			newPrefix:        "/apisix",
			expectConflict:   false,
		},
		{
			name:             "不同集群-即使prefix有层级关系也允许",
			existingEndpoint: "http://etcd9:2379",
			existingPrefix:   "/apisix",
			newEndpoints:     []string{"http://etcd10:2379"},
			newPrefix:        "/apisix/gateway",
			expectConflict:   false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建已存在的网关
			existingGateway, err := createTestGateway(
				ctx,
				"existing-gateway-"+string(rune('a'+i)),
				tt.existingEndpoint,
				tt.existingPrefix,
			)
			assert.NoError(t, err, "创建已存在网关失败")

			// 调用检查函数（gatewayID=0 表示新建网关）
			err = checkPrefixConflictLogic(ctx, 0, tt.newEndpoints, tt.newPrefix)

			if tt.expectConflict {
				assert.Error(t, err, "期望返回冲突错误但没有")
				assert.True(t, strings.Contains(err.Error(), "层级冲突"),
					"错误信息应包含 '层级冲突'，实际: %s", err.Error())
			} else {
				assert.NoError(t, err, "不期望返回错误: %v", err)
			}

			// 清理：删除测试网关
			_ = biz.DeleteGateway(ctx, existingGateway)
		})
	}
}

func TestCheckEtcdPrefixConflictLogic_EditGateway(t *testing.T) {
	// 初始化内存数据库
	util.InitEmbedDb()
	ctx := context.Background()

	// 创建一个网关
	existingGateway, err := createTestGateway(
		ctx,
		"edit-test-gateway",
		"http://etcd-edit:2379",
		"/apisix/edit",
	)
	assert.NoError(t, err)

	// 编辑自己时，即使 prefix 相同也不应该报冲突
	t.Run("编辑自己-prefix不变不应报冲突", func(t *testing.T) {
		// 传入自己的 ID，应该排除自己
		err = checkPrefixConflictLogic(
			ctx,
			existingGateway.ID,
			[]string{"http://etcd-edit:2379"},
			"/apisix/edit",
		)
		assert.NoError(t, err, "编辑自己时不应该报 prefix 冲突")
	})

	// 清理
	_ = biz.DeleteGateway(ctx, existingGateway)
}
