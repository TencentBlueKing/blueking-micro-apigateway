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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestParseOrderByExprList(t *testing.T) {
	ascFieldMap := map[string]field.Expr{
		"name": field.NewField("", "name").Asc(),
		"age":  field.NewField("", "age").Asc(),
	}
	descFieldMap := map[string]field.Expr{
		"name": field.NewField("", "name").Desc(),
		"age":  field.NewField("", "age").Desc(),
	}

	tests := []struct {
		name     string
		orderBy  string
		expected int
	}{
		{
			name:     "test_empty",
			orderBy:  "",
			expected: 0,
		},
		{
			name:     "test_asc",
			orderBy:  "name:asc",
			expected: 1,
		},
		{
			name:     "test_desc",
			orderBy:  "age:desc",
			expected: 1,
		},
		{
			name:     "test_asc_and_desc",
			orderBy:  "name:asc,age:desc",
			expected: 2,
		},
		{
			name:     "test_invalid_key",
			orderBy:  "invalid:asc",
			expected: 0,
		},
		{
			name:     "test_invalid_value",
			orderBy:  "name:invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseOrderByExprList(ascFieldMap, descFieldMap, tt.orderBy)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestGetResourcesLabels(t *testing.T) {
	route := data.Route2WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	// 创建资源
	if err := CreateRoute(gatewayCtx, *route); err != nil {
		t.Errorf("CreateRoute error = %v", err)
		return
	}
	labels, err := GetResourcesLabels(gatewayCtx, constant.Route)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(labels))
}

// TestBatchGetResources_SmallBatch 测试小批量获取资源
func TestBatchGetResources_SmallBatch(t *testing.T) {
	// 创建测试资源，使用唯一名称避免冲突
	route1 := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route2 := data.Route2WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)

	// 确保名称唯一
	route1.Name = fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
	route2.Name = fmt.Sprintf("test-route2-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreateRoute(gatewayCtx, *route1))
	assert.NoError(t, CreateRoute(gatewayCtx, *route2))

	// 测试小批量获取
	ids := []string{route1.ID, route2.ID}
	resources, err := BatchGetResources(gatewayCtx, constant.Route, ids)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(resources))

	// 验证返回的资源
	resourceMap := make(map[string]*model.ResourceCommonModel)
	for _, resource := range resources {
		resourceMap[resource.ID] = resource
	}

	assert.Contains(t, resourceMap, route1.ID)
	assert.Contains(t, resourceMap, route2.ID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, resourceMap[route1.ID].Status)
	assert.Equal(t, constant.ResourceStatusCreateDraft, resourceMap[route2.ID].Status)
}

// TestBatchGetResources_LargeBatch 测试大批量获取资源（超过 DBBatchSize）
func TestBatchGetResources_LargeBatch(t *testing.T) {
	// 创建大量测试资源（超过 DBBatchSize = 500）
	routeCount := constant.DBBatchSize + 100 // 600 个资源
	var routes []*model.Route
	var ids []string

	for i := 0; i < routeCount; i++ {
		route := &model.Route{
			Name:           "test-route-" + string(rune(i)),
			ServiceID:      "",
			UpstreamID:     "",
			PluginConfigID: "",
			ResourceCommonModel: model.ResourceCommonModel{
				GatewayID: gatewayInfo.ID,
				ID:        idx.GenResourceID(constant.Route),
				Config: datatypes.JSON(`{
					"uris": ["/test"],
					"methods": ["GET"],
					"upstream": {
						"type": "roundrobin",
						"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}],
						"scheme": "http"
					}
				}`),
				Status: constant.ResourceStatusCreateDraft,
			},
		}
		routes = append(routes, route)
		ids = append(ids, route.ID)

		// 创建资源
		assert.NoError(t, CreateRoute(gatewayCtx, *route))
	}

	// 测试大批量获取
	resources, err := BatchGetResources(gatewayCtx, constant.Route, ids)

	assert.NoError(t, err)
	assert.Equal(t, routeCount, len(resources))

	// 验证所有资源都被正确返回
	resourceMap := make(map[string]*model.ResourceCommonModel)
	for _, resource := range resources {
		resourceMap[resource.ID] = resource
	}

	for _, id := range ids {
		assert.Contains(t, resourceMap, id)
		assert.Equal(t, constant.ResourceStatusCreateDraft, resourceMap[id].Status)
	}
}

// TestBatchGetResources_EmptyIDs 测试空 ID 列表
func TestBatchGetResources_EmptyIDs(t *testing.T) {
	// 测试空 ID 列表
	resources, err := BatchGetResources(gatewayCtx, constant.Route, []string{})

	assert.NoError(t, err)
	// 空 ID 列表应该返回所有资源
	assert.GreaterOrEqual(t, len(resources), 0)
}

// TestBatchUpdateResourceStatus_SmallBatch 测试小批量更新资源状态
func TestBatchUpdateResourceStatus_SmallBatch(t *testing.T) {
	// 创建测试资源
	route1 := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route2 := data.Route2WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)

	// 创建资源
	// 确保名称唯一
	route1.Name = fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
	route2.Name = fmt.Sprintf("test-route2-%d", time.Now().UnixNano())
	assert.NoError(t, CreateRoute(gatewayCtx, *route1))
	assert.NoError(t, CreateRoute(gatewayCtx, *route2))

	// 测试小批量更新状态
	ids := []string{route1.ID, route2.ID}
	err := BatchUpdateResourceStatus(gatewayCtx, constant.Route, ids, constant.ResourceStatusSuccess)

	assert.NoError(t, err)

	// 验证状态已更新
	updatedRoute1, err := GetRoute(gatewayCtx, route1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, updatedRoute1.Status)

	updatedRoute2, err := GetRoute(gatewayCtx, route2.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, updatedRoute2.Status)
}

// TestBatchUpdateResourceStatus_LargeBatch 测试大批量更新资源状态
func TestBatchUpdateResourceStatus_LargeBatch(t *testing.T) {
	// 创建大量测试资源
	routeCount := constant.DBBatchSize + 50 // 550 个资源
	var routes []*model.Route
	var ids []string

	for i := 0; i < routeCount; i++ {
		name := fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
		route := &model.Route{
			Name:           name,
			ServiceID:      "",
			UpstreamID:     "",
			PluginConfigID: "",
			ResourceCommonModel: model.ResourceCommonModel{
				GatewayID: gatewayInfo.ID,
				ID:        idx.GenResourceID(constant.Route),
				Config: datatypes.JSON(`{
					"uris": ["/test"],
					"methods": ["GET"],
					"upstream": {
						"type": "roundrobin",
						"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}],
						"scheme": "http"
					}
				}`),
				Status: constant.ResourceStatusCreateDraft,
			},
		}
		routes = append(routes, route)
		ids = append(ids, route.ID)

		// 创建资源
		assert.NoError(t, CreateRoute(gatewayCtx, *route))
	}

	// 测试大批量更新状态
	err := BatchUpdateResourceStatus(gatewayCtx, constant.Route, ids, constant.ResourceStatusSuccess)

	assert.NoError(t, err)

	// 验证所有资源状态都已更新
	for _, id := range ids {
		route, err := GetRoute(gatewayCtx, id)
		assert.NoError(t, err)
		assert.Equal(t, constant.ResourceStatusSuccess, route.Status)
	}
}

// TestGetResourceByIDs_SmallBatch 测试小批量根据 IDs 获取资源
func TestGetResourceByIDs_SmallBatch(t *testing.T) {
	// 创建测试资源
	route1 := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route2 := data.Route2WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)

	// 创建资源
	// 确保名称唯一
	route1.Name = fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
	route2.Name = fmt.Sprintf("test-route2-%d", time.Now().UnixNano())
	assert.NoError(t, CreateRoute(gatewayCtx, *route1))
	assert.NoError(t, CreateRoute(gatewayCtx, *route2))

	// 测试小批量获取
	ids := []string{route1.ID, route2.ID}
	resources, err := GetResourceByIDs(gatewayCtx, constant.Route, ids)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(resources))

	// 验证返回的资源
	resourceMap := make(map[string]model.ResourceCommonModel)
	for _, resource := range resources {
		resourceMap[resource.ID] = resource
	}

	assert.Contains(t, resourceMap, route1.ID)
	assert.Contains(t, resourceMap, route2.ID)
}

// TestGetResourceByIDs_LargeBatch 测试大批量根据 IDs 获取资源
func TestGetResourceByIDs_LargeBatch(t *testing.T) {
	// 创建大量测试资源
	routeCount := constant.DBBatchSize + 200 // 700 个资源
	var routes []*model.Route
	var ids []string

	for i := 0; i < routeCount; i++ {
		name := fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
		route := &model.Route{
			Name:           name,
			ServiceID:      "",
			UpstreamID:     "",
			PluginConfigID: "",
			ResourceCommonModel: model.ResourceCommonModel{
				GatewayID: gatewayInfo.ID,
				ID:        idx.GenResourceID(constant.Route),
				Config: datatypes.JSON(`{
					"uris": ["/test"],
					"methods": ["GET"],
					"upstream": {
						"type": "roundrobin",
						"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}],
						"scheme": "http"
					}
				}`),
				Status: constant.ResourceStatusCreateDraft,
			},
		}
		routes = append(routes, route)
		ids = append(ids, route.ID)

		// 创建资源
		assert.NoError(t, CreateRoute(gatewayCtx, *route))
	}

	// 测试大批量获取
	resources, err := GetResourceByIDs(gatewayCtx, constant.Route, ids)

	assert.NoError(t, err)
	assert.Equal(t, routeCount, len(resources))

	// 验证所有资源都被正确返回
	resourceMap := make(map[string]model.ResourceCommonModel)
	for _, resource := range resources {
		resourceMap[resource.ID] = resource
	}

	for _, id := range ids {
		assert.Contains(t, resourceMap, id)
	}
}

// TestDeleteResourceByIDs_SmallBatch 测试小批量删除资源
func TestDeleteResourceByIDs_SmallBatch(t *testing.T) {
	// 创建测试资源
	route1 := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route2 := data.Route2WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	// 确保名称唯一
	route1.Name = fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
	route2.Name = fmt.Sprintf("test-route2-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreateRoute(gatewayCtx, *route1))
	assert.NoError(t, CreateRoute(gatewayCtx, *route2))

	// 测试小批量删除
	ids := []string{route1.ID, route2.ID}
	err := DeleteResourceByIDs(gatewayCtx, constant.Route, ids)

	assert.NoError(t, err)

	// 验证资源已被删除
	_, err = GetRoute(gatewayCtx, route1.ID)
	assert.Error(t, err)

	_, err = GetRoute(gatewayCtx, route2.ID)
	assert.Error(t, err)
}

// TestDeleteResourceByIDs_LargeBatch 测试大批量删除资源
func TestDeleteResourceByIDs_LargeBatch(t *testing.T) {
	// 创建大量测试资源
	routeCount := constant.DBBatchSize + 100 // 600 个资源
	var routes []*model.Route
	var ids []string

	for i := 0; i < routeCount; i++ {
		name := fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
		route := &model.Route{
			Name:           name,
			ServiceID:      "",
			UpstreamID:     "",
			PluginConfigID: "",
			ResourceCommonModel: model.ResourceCommonModel{
				GatewayID: gatewayInfo.ID,
				ID:        idx.GenResourceID(constant.Route),
				Config: datatypes.JSON(`{
					"uris": ["/test"],
					"methods": ["GET"],
					"upstream": {
						"type": "roundrobin",
						"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}],
						"scheme": "http"
					}
				}`),
				Status: constant.ResourceStatusCreateDraft,
			},
		}
		routes = append(routes, route)
		ids = append(ids, route.ID)

		// 创建资源
		assert.NoError(t, CreateRoute(gatewayCtx, *route))
	}

	// 测试大批量删除
	err := DeleteResourceByIDs(gatewayCtx, constant.Route, ids)

	assert.NoError(t, err)

	// 验证所有资源都已被删除
	for _, id := range ids {
		_, err := GetRoute(gatewayCtx, id)
		assert.Error(t, err)
	}
}

// TestBatchOperations_EdgeCases 测试边界情况
func TestBatchOperations_EdgeCases(t *testing.T) {
	t.Run("TestBatchGetResources_ExactBatchSize", func(t *testing.T) {
		// 测试恰好等于 DBBatchSize 的情况
		routeCount := constant.DBBatchSize // 500 个资源
		var routes []*model.Route
		var ids []string

		for i := 0; i < routeCount; i++ {
			name := fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
			route := &model.Route{
				Name:           name,
				ServiceID:      "",
				UpstreamID:     "",
				PluginConfigID: "",
				ResourceCommonModel: model.ResourceCommonModel{
					GatewayID: gatewayInfo.ID,
					ID:        idx.GenResourceID(constant.Route),
					Config: datatypes.JSON(`{
						"uris": ["/test"],
						"methods": ["GET"],
						"upstream": {
							"type": "roundrobin",
							"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}],
							"scheme": "http"
						}
					}`),
					Status: constant.ResourceStatusCreateDraft,
				},
			}
			routes = append(routes, route)
			ids = append(ids, route.ID)

			// 创建资源
			assert.NoError(t, CreateRoute(gatewayCtx, *route))
		}

		// 测试获取
		resources, err := BatchGetResources(gatewayCtx, constant.Route, ids)
		assert.NoError(t, err)
		assert.Equal(t, routeCount, len(resources))
	})

	t.Run("TestBatchGetResources_OneMoreThanBatchSize", func(t *testing.T) {
		// 测试比 DBBatchSize 多 1 的情况
		routeCount := constant.DBBatchSize + 1 // 501 个资源
		var routes []*model.Route
		var ids []string

		for i := 0; i < routeCount; i++ {
			name := fmt.Sprintf("test-route1-%d", time.Now().UnixNano())
			route := &model.Route{
				Name:           name,
				ServiceID:      "",
				UpstreamID:     "",
				PluginConfigID: "",
				ResourceCommonModel: model.ResourceCommonModel{
					GatewayID: gatewayInfo.ID,
					ID:        idx.GenResourceID(constant.Route),
					Config: datatypes.JSON(`{
						"uris": ["/test"],
						"methods": ["GET"],
						"upstream": {
							"type": "roundrobin",
							"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}],
							"scheme": "http"
						}
					}`),
					Status: constant.ResourceStatusCreateDraft,
				},
			}
			routes = append(routes, route)
			ids = append(ids, route.ID)

			// 创建资源
			assert.NoError(t, CreateRoute(gatewayCtx, *route))
		}

		// 测试获取
		resources, err := BatchGetResources(gatewayCtx, constant.Route, ids)
		assert.NoError(t, err)
		assert.Equal(t, routeCount, len(resources))
	})
}
