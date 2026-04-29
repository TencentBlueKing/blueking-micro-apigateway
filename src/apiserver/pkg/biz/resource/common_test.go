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

package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/gen/field"

	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	pkgutils "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

var (
	gatewayInfo *model.Gateway
	gatewayCtx  context.Context
)

func init() {
	if err := cryptography.Init("jxi18GX5w2qgHwfZCFpn07q8FScXJOd3", "k2dbCGetyusW"); err != nil {
		panic(err)
	}
	util.InitEmbedDb()

	gatewayInfo = data.Gateway1WithBkAPISIX()
	if err := repo.Gateway.WithContext(context.Background()).Create(gatewayInfo); err != nil {
		panic(err)
	}
	gatewayCtx = ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
}

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
			result := pkgutils.ParseOrderByExprList(ascFieldMap, descFieldMap, tt.orderBy)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

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
			name:         "injects id when schema requires it",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			resourceID:   "cg-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{},"id":"cg-generated-id"}`,
		},
		{
			name:         "keeps existing id",
			resourceType: constant.GlobalRule,
			version:      constant.APISIXVersion313,
			resourceID:   "gr-generated-id",
			rawConfig:    json.RawMessage(`{"id":"client-id","plugins":{}}`),
			wantConfig:   `{"id":"client-id","plugins":{}}`,
		},
		{
			name:         "skips injection when resource id is empty",
			resourceType: constant.PluginConfig,
			version:      constant.APISIXVersion311,
			resourceID:   "",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{}}`,
		},
		{
			name:         "skips injection for versions that do not require id",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion33,
			resourceID:   "cg-generated-id",
			rawConfig:    json.RawMessage(`{"plugins":{}}`),
			wantConfig:   `{"plugins":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InjectGeneratedIDForValidation(tt.rawConfig, tt.resourceType, tt.version, tt.resourceID)
			assert.JSONEq(t, tt.wantConfig, string(got))
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

// TestBatchGetResources_LargeBatch 测试大批量获取资源（超过 DBBatchCreateSize）
func TestBatchGetResources_LargeBatch(t *testing.T) {
	// 创建大量测试资源（超过 DBBatchCreateSize = 500）
	routeCount := constant.DBBatchCreateSize + 100 // 600 个资源
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
	routeCount := constant.DBConditionIDMaxLength + 50 // 550 个资源
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
	routeCount := constant.DBConditionIDMaxLength + 200 // 700 个资源
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

func TestGetSchemaByIDs_ReadsSchemaTable(t *testing.T) {
	pluginConfig := data.PluginConfig1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft)
	pluginConfig.Name = fmt.Sprintf("test-plugin-config-%d", time.Now().UnixNano())
	assert.NoError(t, CreatePluginConfig(gatewayCtx, *pluginConfig))

	schemaName := fmt.Sprintf("test-schema-%d", time.Now().UnixNano())
	assert.NoError(t, schemabiz.BatchCreateSchema(gatewayCtx, []*model.GatewayCustomPluginSchema{
		{
			GatewayID: gatewayInfo.ID,
			Name:      schemaName,
			Schema:    datatypes.JSON(`{"type":"object","properties":{"count":{"type":"integer"}}}`),
			Example:   datatypes.JSON(`{"count":1}`),
			BaseModel: model.BaseModel{
				Creator: "tester",
				Updater: "tester",
			},
			OperationType: constant.OperationImport,
		},
	}))

	schemaInfo, err := schemabiz.GetSchemaByName(gatewayCtx, schemaName)
	assert.NoError(t, err)

	schemas, err := GetSchemaByIDs(gatewayCtx, []string{strconv.Itoa(schemaInfo.AutoID)})
	assert.NoError(t, err)
	if !assert.Len(t, schemas, 1) {
		return
	}

	assert.Equal(t, schemaName, schemas[0].Name)
	assert.JSONEq(t, string(schemaInfo.Schema), string(schemas[0].Schema))
	assert.JSONEq(t, string(schemaInfo.Example), string(schemas[0].Example))
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
	routeCount := constant.DBConditionIDMaxLength + 100 //  300 个资源
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
		// 测试恰好等于 DBBatchCreateSize 的情况
		routeCount := constant.DBConditionIDMaxLength // 200 个资源
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
		// 测试比 DBBatchCreateSize 多 1 的情况
		routeCount := constant.DBConditionIDMaxLength + 1 // 201 个资源
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

// TestIsResourceConfigChanged_Route_SameConfig 测试 Route 资源配置相同的情况
func TestIsResourceConfigChanged_Route_SameConfig(t *testing.T) {
	// 创建测试路由
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-same-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreateRoute(gatewayCtx, *route))

	// 从数据库重新获取资源，使用存储的配置格式（经过 HandleConfig 处理）
	retrievedRoute, err := GetRoute(gatewayCtx, route.ID)
	assert.NoError(t, err)
	// 使用存储的配置，确保格式完全一致
	inputConfig := json.RawMessage(retrievedRoute.Config)

	// 测试：配置相同，应该返回 false（未变化）
	// 注意：由于函数使用字节排序比较，需要确保格式完全一致
	changed := IsResourceConfigChanged(gatewayCtx, constant.Route, route.ID, inputConfig)
	assert.False(t, changed, "相同配置应该返回 false")
}

// TestIsResourceConfigChanged_Route_DifferentConfig 测试 Route 资源配置不同的情况
func TestIsResourceConfigChanged_Route_DifferentConfig(t *testing.T) {
	// 创建测试路由
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-diff-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreateRoute(gatewayCtx, *route))

	// 使用不同的配置
	differentConfig := json.RawMessage(`{
		"uris": ["/different"],
		"methods": ["POST"],
		"upstream": {
			"type": "roundrobin",
			"nodes": [{"host": "example.com", "port": 80, "weight": 1}],
			"scheme": "http"
		}
	}`)

	// 测试：配置不同，应该返回 true（已变化）
	changed := IsResourceConfigChanged(gatewayCtx, constant.Route, route.ID, differentConfig)
	assert.True(t, changed, "不同配置应该返回 true")
}

// TestIsResourceConfigChanged_PluginMetadata_SameConfigDifferentName 测试 PluginMetadata 配置相同但名称不同的情况
func TestIsResourceConfigChanged_PluginMetadata_SameConfigDifferentName(t *testing.T) {
	// 创建测试 PluginMetadata
	metadata := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusCreateDraft)
	metadata.Name = fmt.Sprintf("test-metadata-same-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreatePluginMetadata(gatewayCtx, *metadata))

	// 从数据库重新获取资源，使用存储的配置格式
	retrievedMetadata, err := GetPluginMetadata(gatewayCtx, metadata.ID)
	assert.NoError(t, err)
	// 使用相同的配置（name 字段会被函数内部移除）
	inputConfig := json.RawMessage(retrievedMetadata.Config)

	// 测试：配置相同（name 被移除后），应该返回 false（未变化）
	changed := IsResourceConfigChanged(gatewayCtx, constant.PluginMetadata, metadata.ID, inputConfig)
	assert.False(t, changed, "PluginMetadata 相同配置（name 被移除后）应该返回 false")
}

// TestIsResourceConfigChanged_PluginMetadata_DifferentConfig 测试 PluginMetadata 配置不同的情况
func TestIsResourceConfigChanged_PluginMetadata_DifferentConfig(t *testing.T) {
	// 创建测试 PluginMetadata
	metadata := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusCreateDraft)
	metadata.Name = fmt.Sprintf("test-metadata-diff-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreatePluginMetadata(gatewayCtx, *metadata))

	// 使用不同的配置
	differentConfig := json.RawMessage(`{
		"config": {
			"log_format": {
				"@timestamp": "$time_iso8601",
				"client_ip": "$remote_addr",
				"host": "$host",
				"new_field": "new_value"
			},
			"name": "clickhouse-logger"
		}
	}`)

	// 测试：配置不同，应该返回 true（已变化）
	changed := IsResourceConfigChanged(gatewayCtx, constant.PluginMetadata, metadata.ID, differentConfig)
	assert.True(t, changed, "PluginMetadata 不同配置应该返回 true")
}

// TestIsResourceConfigChanged_ResourceNotFound 测试资源不存在的情况
func TestIsResourceConfigChanged_ResourceNotFound(t *testing.T) {
	// 使用不存在的资源 ID
	nonExistentID := idx.GenResourceID(constant.Route)
	inputConfig := json.RawMessage(`{"uris": ["/test"]}`)

	// 测试：资源不存在，应该返回 true（视为已变化）
	changed := IsResourceConfigChanged(gatewayCtx, constant.Route, nonExistentID, inputConfig)
	assert.True(t, changed, "资源不存在时应该返回 true（视为已变化）")
}

// TestIsResourceConfigChanged_EmptyConfig 测试空配置的情况
func TestIsResourceConfigChanged_EmptyConfig(t *testing.T) {
	// 创建测试路由
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-empty-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreateRoute(gatewayCtx, *route))

	// 使用空配置
	emptyConfig := json.RawMessage(`{}`)

	// 测试：空配置与现有配置不同，应该返回 true
	changed := IsResourceConfigChanged(gatewayCtx, constant.Route, route.ID, emptyConfig)
	assert.True(t, changed, "空配置应该被认为是不同的")
}

// TestIsResourceConfigChanged_Service_SameConfig 测试 Service 资源配置相同的情况
func TestIsResourceConfigChanged_Service_SameConfig(t *testing.T) {
	// 创建测试 Service
	service := data.Service1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft)
	service.Name = fmt.Sprintf("test-service-same-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreateService(gatewayCtx, *service))

	// 从数据库重新获取资源，使用存储的配置格式
	retrievedService, err := GetService(gatewayCtx, service.ID)
	assert.NoError(t, err)
	inputConfig := json.RawMessage(retrievedService.Config)

	// 测试：配置相同，应该返回 false（未变化）
	changed := IsResourceConfigChanged(gatewayCtx, constant.Service, service.ID, inputConfig)
	assert.False(t, changed, "Service 相同配置应该返回 false")
}

// TestIsResourceConfigChanged_Service_DifferentConfig 测试 Service 资源配置不同的情况
func TestIsResourceConfigChanged_Service_DifferentConfig(t *testing.T) {
	// 创建测试 Service
	service := data.Service1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft)
	service.Name = fmt.Sprintf("test-service-diff-%d", time.Now().UnixNano())

	// 创建资源
	assert.NoError(t, CreateService(gatewayCtx, *service))

	// 使用不同的配置
	differentConfig := json.RawMessage(`{
		"upstream": {
			"type": "roundrobin",
			"nodes": [
				{
					"host": "different.com",
					"port": 8080,
					"weight": 2
				}
			],
			"scheme": "https"
		}
	}`)

	// 测试：配置不同，应该返回 true（已变化）
	changed := IsResourceConfigChanged(gatewayCtx, constant.Service, service.ID, differentConfig)
	assert.True(t, changed, "Service 不同配置应该返回 true")
}

// TestIsResourceChanged_Route_NameChanged tests that route name changes are detected
func TestIsResourceChanged_Route_NameChanged(t *testing.T) {
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-name-%d", time.Now().UnixNano())
	route.Config = datatypes.JSON(`{"uri": "/test"}`)

	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	configJSON := json.RawMessage(`{"uri": "/test"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Route, route.ID, configJSON, map[string]any{
		"name":             "new-name",
		"service_id":       "",
		"upstream_id":      "",
		"plugin_config_id": "",
	})
	assert.True(t, changed, "Should detect name change even when config is same")
}

// TestIsResourceChanged_Route_ServiceIDChanged tests that service_id changes are detected
func TestIsResourceChanged_Route_ServiceIDChanged(t *testing.T) {
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-svc-%d", time.Now().UnixNano())
	route.ServiceID = "service-123"
	route.Config = datatypes.JSON(`{"uri": "/test"}`)

	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	configJSON := json.RawMessage(`{"uri": "/test"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Route, route.ID, configJSON, map[string]any{
		"name":             route.Name,
		"service_id":       "service-999",
		"upstream_id":      "",
		"plugin_config_id": "",
	})
	assert.True(t, changed, "Should detect service_id change even when config is same")
}

// TestIsResourceChanged_Route_NoChanges tests that no changes are correctly detected
func TestIsResourceChanged_Route_NoChanges(t *testing.T) {
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-nochange-%d", time.Now().UnixNano())
	route.Config = datatypes.JSON(`{"uri": "/test"}`)

	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	retrievedRoute, err := GetRoute(gatewayCtx, route.ID)
	assert.NoError(t, err)

	configJSON := json.RawMessage(retrievedRoute.Config)
	changed := IsResourceChanged(gatewayCtx, constant.Route, route.ID, configJSON, map[string]any{
		"name":             retrievedRoute.Name,
		"service_id":       retrievedRoute.ServiceID,
		"upstream_id":      retrievedRoute.UpstreamID,
		"plugin_config_id": retrievedRoute.PluginConfigID,
	})
	assert.False(t, changed, "Should not detect changes when nothing changed")
}

// TestIsResourceChanged_Consumer_UsernameChanged tests consumer username changes
func TestIsResourceChanged_Consumer_UsernameChanged(t *testing.T) {
	consumer := data.Consumer1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft)
	consumer.Username = fmt.Sprintf("test-user-%d", time.Now().UnixNano())
	consumer.GroupID = "group-123"
	consumer.Config = datatypes.JSON(`{"username": "` + consumer.Username + `"}`)

	err := CreateConsumer(gatewayCtx, *consumer)
	assert.NoError(t, err)

	configJSON := json.RawMessage(`{"username": "` + consumer.Username + `"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Consumer, consumer.ID, configJSON, map[string]any{
		"username": "new-user",
		"group_id": "group-123",
	})
	assert.True(t, changed, "Should detect username change even when config is same")
}

// TestIsResourceChanged_Service_UpstreamIDChanged tests service upstream_id changes
func TestIsResourceChanged_Service_UpstreamIDChanged(t *testing.T) {
	service := &model.Service{
		Name:       fmt.Sprintf("test-service-%d", time.Now().UnixNano()),
		UpstreamID: "upstream-123",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Service),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name": "test-service"}`),
			Status:    constant.ResourceStatusCreateDraft,
		},
	}

	err := CreateService(gatewayCtx, *service)
	assert.NoError(t, err)

	configJSON := json.RawMessage(`{"name": "test-service"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Service, service.ID, configJSON, map[string]any{
		"name":        service.Name,
		"upstream_id": "upstream-999",
	})
	assert.True(t, changed, "Should detect upstream_id change even when config is same")
}

// TestIsResourceChanged_Route_ConfigAndNameChanged tests both config and extra fields changed
func TestIsResourceChanged_Route_ConfigAndNameChanged(t *testing.T) {
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-both-%d", time.Now().UnixNano())
	route.Config = datatypes.JSON(`{"uri": "/old"}`)

	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	configJSON := json.RawMessage(`{"uri": "/new"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Route, route.ID, configJSON, map[string]any{
		"name":             "new-name",
		"service_id":       "",
		"upstream_id":      "",
		"plugin_config_id": "",
	})
	assert.True(t, changed, "Should detect changes when both config and name changed")
}

// TestIsResourceChanged_Upstream_NameAndSSLIDChanged tests upstream name and ssl_id changes
func TestIsResourceChanged_Upstream_NameAndSSLIDChanged(t *testing.T) {
	upstream := &model.Upstream{
		Name:  fmt.Sprintf("test-upstream-%d", time.Now().UnixNano()),
		SSLID: "ssl-123",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Upstream),
			GatewayID: gatewayInfo.ID,
			Config: datatypes.JSON(`{
				"type": "roundrobin",
				"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}]
			}`),
			Status: constant.ResourceStatusCreateDraft,
		},
	}

	err := CreateUpstream(gatewayCtx, *upstream)
	assert.NoError(t, err)

	retrievedUpstream, err := GetUpstream(gatewayCtx, upstream.ID)
	assert.NoError(t, err)
	configJSON := json.RawMessage(retrievedUpstream.Config)
	changed := IsResourceChanged(gatewayCtx, constant.Upstream, upstream.ID, configJSON, map[string]any{
		"name":   "new-upstream-name",
		"ssl_id": "ssl-123",
	})
	assert.True(t, changed, "Should detect name change for upstream")

	changed = IsResourceChanged(gatewayCtx, constant.Upstream, upstream.ID, configJSON, map[string]any{
		"name":   retrievedUpstream.Name,
		"ssl_id": "ssl-999",
	})
	assert.True(t, changed, "Should detect ssl_id change for upstream")
}

// TestIsResourceChanged_Proto_NameChanged tests proto name changes
func TestIsResourceChanged_Proto_NameChanged(t *testing.T) {
	proto := data.Proto1(gatewayInfo, constant.ResourceStatusCreateDraft)
	proto.Name = fmt.Sprintf("test-proto-%d", time.Now().UnixNano())

	err := CreateProto(gatewayCtx, *proto)
	assert.NoError(t, err)

	retrievedProto, err := GetProto(gatewayCtx, proto.ID)
	assert.NoError(t, err)

	configJSON := json.RawMessage(retrievedProto.Config)
	changed := IsResourceChanged(gatewayCtx, constant.Proto, proto.ID, configJSON, map[string]any{
		"name": "new-proto-name",
	})
	assert.True(t, changed, "Should detect name change for proto")
}
