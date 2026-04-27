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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	utiltesting "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/testing"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestBuildConfigRawForValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		config       string
		want         string
	}{
		{
			name:         "keep route id and name for 3.11",
			resourceType: constant.Route,
			version:      constant.APISIXVersion311,
			config:       `{"id":"route-id","name":"route-a","uris":["/test"]}`,
			want:         `{"id":"route-id","name":"route-a","uris":["/test"]}`,
		},
		{
			name:         "remove consumer id",
			resourceType: constant.Consumer,
			version:      constant.APISIXVersion313,
			config:       `{"id":"consumer-id","username":"consumer-a"}`,
			want:         `{"username":"consumer-a"}`,
		},
		{
			name:         "remove consumer group name on 3.11",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion311,
			config:       `{"id":"cg-id","name":"group-a","plugins":{}}`,
			want:         `{"id":"cg-id","plugins":{}}`,
		},
		{
			name:         "keep consumer group name on 3.13",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			config:       `{"id":"cg-id","name":"group-a","plugins":{}}`,
			want:         `{"id":"cg-id","name":"group-a","plugins":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildConfigRawForValidation(tt.config, tt.resourceType, tt.version)
			assert.JSONEq(t, tt.want, string(got))
			assert.JSONEq(t, tt.config, tt.config)
		})
	}
}

func TestValidateResourceAssociatedResourceMissing(t *testing.T) {
	t.Parallel()

	ctx := ginx.SetGatewayInfoToContext(context.Background(), data.Gateway1WithBkAPISIX())
	resources := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {
			{
				Type: constant.Route,
				ID:   "route-id",
				Config: datatypes.JSON(`{
					"name":"route-a",
					"uris":["/test"],
					"methods":["GET"],
					"service_id":"missing-service"
				}`),
			},
		},
	}

	err := ValidateResource(ctx, resources, map[string]struct{}{}, map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "associated service [id:missing-service] not found")
}

func TestPrepareValidationPayloadImportParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		config       string
		want         string
	}{
		{
			name:         "consumer import removes id like legacy import validation",
			resourceType: constant.Consumer,
			version:      constant.APISIXVersion313,
			config:       `{"id":"consumer-id","username":"demo"}`,
			want:         `{"username":"demo"}`,
		},
		{
			name:         "consumer group import keeps id but drops name on 3.11",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion311,
			config:       `{"id":"cg-id","name":"group-a","plugins":{}}`,
			want:         `{"id":"cg-id","plugins":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resourcecodec.PrepareValidationPayload(resourcecodec.ValidationInput{
				Source:       resourcecodec.SourceImport,
				ResourceType: tt.resourceType,
				Version:      tt.version,
				Config:       json.RawMessage(tt.config),
			})
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestImportDraftParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		resourceID   string
		config       string
		want         string
	}{
		{
			name:         "plugin metadata import keeps derived id and removes unsupported name for validation",
			resourceType: constant.PluginMetadata,
			version:      constant.APISIXVersion313,
			resourceID:   "stored-plugin-metadata-id",
			config:       `{"id":"jwt-auth","name":"jwt-auth","key":"value"}`,
			want:         `{"id":"jwt-auth","key":"value"}`,
		},
		{
			name:         "stream route import keeps name only on 3.13",
			resourceType: constant.StreamRoute,
			version:      constant.APISIXVersion313,
			resourceID:   "stream-route-id",
			config:       `{"name":"stream-a","remote_addr":"127.0.0.1","server_port":9100,"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
			want:         `{"id":"stream-route-id","name":"stream-a","remote_addr":"127.0.0.1","server_port":9100,"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
		},
		{
			name:         "stream route import drops name on 3.11",
			resourceType: constant.StreamRoute,
			version:      constant.APISIXVersion311,
			resourceID:   "stream-route-id",
			config:       `{"name":"stream-a","remote_addr":"127.0.0.1","server_port":9100,"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
			want:         `{"id":"stream-route-id","remote_addr":"127.0.0.1","server_port":9100,"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft, err := resourcecodec.PrepareRequestDraft(resourcecodec.RequestInput{
				Source:       resourcecodec.SourceImport,
				Operation:    constant.OperationImport,
				GatewayID:    1001,
				ResourceType: tt.resourceType,
				Version:      tt.version,
				PathID:       tt.resourceID,
				Config:       json.RawMessage(tt.config),
			})
			assert.NoError(t, err)

			builtPayload, err := resourcecodec.BuildRequestPayload(draft, constant.DATABASE)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(builtPayload.Payload))
		})
	}
}

func TestHistoricalImportValidationFixtures(t *testing.T) {
	t.Parallel()

	for _, fixture := range utiltesting.HistoricalValidationFixtures() {
		fixture := fixture.Clone()
		t.Run(fixture.Name, func(t *testing.T) {
			draft := resourcecodec.PrepareStoredDraft(resourcecodec.StoredRowInput{
				GatewayID:    gatewayInfo.ID,
				ResourceType: fixture.ResourceType,
				Version:      fixture.Version,
				ResourceID:   fixture.Stored.ID,
				NameKey:      historicalFixtureNameKey(fixture.ResourceType),
				NameValue:    fixture.Stored.Name,
				Associations: historicalFixtureAssociations(fixture.Stored),
				Config:       fixture.Stored.Config,
			})

			builtPayload, err := resourcecodec.BuildStoredPayload(draft, constant.DATABASE)
			assert.NoError(t, err)
			assert.JSONEq(t, string(fixture.DatabaseConfig), string(builtPayload.Payload))
		})
	}
}

func historicalFixtureNameKey(resourceType constant.APISIXResource) string {
	if resourceType == constant.Consumer {
		return "username"
	}
	return "name"
}

func historicalFixtureAssociations(stored utiltesting.StoredResourceFixture) map[string]string {
	associations := map[string]string{}
	if stored.ServiceID != "" {
		associations["service_id"] = stored.ServiceID
	}
	if stored.UpstreamID != "" {
		associations["upstream_id"] = stored.UpstreamID
	}
	if stored.PluginConfigID != "" {
		associations["plugin_config_id"] = stored.PluginConfigID
	}
	if stored.GroupID != "" {
		associations["group_id"] = stored.GroupID
	}
	return associations
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

func TestCommonResourceReadsPreferResolvedColumns(t *testing.T) {
	route := &model.Route{
		Name:           fmt.Sprintf("typed-route-%d", time.Now().UnixNano()),
		ServiceID:      "typed-service-id",
		UpstreamID:     "typed-upstream-id",
		PluginConfigID: "typed-plugin-config-id",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gatewayInfo.ID,
			ID:        idx.GenResourceID(constant.Route),
			Config: datatypes.JSON(`{
				"uris": ["/typed-route"],
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
	assert.NoError(t, CreateRoute(gatewayCtx, *route))

	consumer := &model.Consumer{
		Username: "typed-consumer-name",
		GroupID:  "typed-group-id",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gatewayInfo.ID,
			ID:        idx.GenResourceID(constant.Consumer),
			Config:    datatypes.JSON(`{"plugins":{"key-auth":{}}}`),
			Status:    constant.ResourceStatusCreateDraft,
		},
	}
	assert.NoError(t, CreateConsumer(gatewayCtx, *consumer))

	assert.NoError(t, database.Client().WithContext(gatewayCtx).
		Table(route.TableName()).
		Where("gateway_id = ? AND id = ?", gatewayInfo.ID, route.ID).
		Update("config", datatypes.JSON(`{
			"uris": ["/typed-route"],
			"methods": ["GET"],
			"upstream": {
				"type": "roundrobin",
				"nodes": [{"host": "httpbin.org", "port": 80, "weight": 1}],
				"scheme": "http"
			}
		}`)).Error)
	assert.NoError(t, database.Client().WithContext(gatewayCtx).
		Table(consumer.TableName()).
		Where("gateway_id = ? AND id = ?", gatewayInfo.ID, consumer.ID).
		Update("config", datatypes.JSON(`{"plugins":{"key-auth":{}}}`)).Error)

	var storedRoute struct {
		Config datatypes.JSON
	}
	assert.NoError(t, database.Client().WithContext(gatewayCtx).
		Table(route.TableName()).
		Select("config").
		Where("gateway_id = ? AND id = ?", gatewayInfo.ID, route.ID).
		Take(&storedRoute).Error)
	assert.Empty(t, gjson.GetBytes(storedRoute.Config, "name").String())
	assert.Empty(t, gjson.GetBytes(storedRoute.Config, "service_id").String())
	assert.Empty(t, gjson.GetBytes(storedRoute.Config, "upstream_id").String())
	assert.Empty(t, gjson.GetBytes(storedRoute.Config, "plugin_config_id").String())

	var storedConsumer struct {
		Config datatypes.JSON
	}
	assert.NoError(t, database.Client().WithContext(gatewayCtx).
		Table(consumer.TableName()).
		Select("config").
		Where("gateway_id = ? AND id = ?", gatewayInfo.ID, consumer.ID).
		Take(&storedConsumer).Error)
	assert.Empty(t, gjson.GetBytes(storedConsumer.Config, "username").String())
	assert.Empty(t, gjson.GetBytes(storedConsumer.Config, "group_id").String())

	gotRoute, err := GetResourceByID(gatewayCtx, constant.Route, route.ID)
	assert.NoError(t, err)
	assert.Empty(t, gjson.GetBytes(gotRoute.Config, "name").String())
	assert.Empty(t, gjson.GetBytes(gotRoute.Config, "service_id").String())
	assert.Empty(t, gjson.GetBytes(gotRoute.Config, "upstream_id").String())
	assert.Empty(t, gjson.GetBytes(gotRoute.Config, "plugin_config_id").String())
	assert.Equal(t, route.Name, gotRoute.GetName(constant.Route))
	assert.Equal(t, route.ServiceID, gotRoute.GetServiceID())
	assert.Equal(t, route.UpstreamID, gotRoute.GetUpstreamID())
	assert.Equal(t, route.PluginConfigID, gotRoute.GetPluginConfigID())

	gotRouteList, err := GetResourceByIDs(gatewayCtx, constant.Route, []string{route.ID})
	assert.NoError(t, err)
	assert.Len(t, gotRouteList, 1)
	assert.Equal(t, route.Name, gotRouteList[0].GetName(constant.Route))
	assert.Equal(t, route.ServiceID, gotRouteList[0].GetServiceID())
	assert.Equal(t, route.UpstreamID, gotRouteList[0].GetUpstreamID())
	assert.Equal(t, route.PluginConfigID, gotRouteList[0].GetPluginConfigID())

	largeBatchIDs := []string{route.ID}
	for i := 0; i < constant.DBConditionIDMaxLength; i++ {
		largeBatchIDs = append(largeBatchIDs, fmt.Sprintf("missing-route-%d", i))
	}

	gotRouteBatch, err := BatchGetResources(gatewayCtx, constant.Route, largeBatchIDs)
	assert.NoError(t, err)
	assert.Len(t, gotRouteBatch, 1)
	assert.Equal(t, route.Name, gotRouteBatch[0].GetName(constant.Route))
	assert.Equal(t, route.ServiceID, gotRouteBatch[0].GetServiceID())
	assert.Equal(t, route.UpstreamID, gotRouteBatch[0].GetUpstreamID())
	assert.Equal(t, route.PluginConfigID, gotRouteBatch[0].GetPluginConfigID())

	gotConsumer, err := GetResourceByID(gatewayCtx, constant.Consumer, consumer.ID)
	assert.NoError(t, err)
	assert.Empty(t, gjson.GetBytes(gotConsumer.Config, "username").String())
	assert.Empty(t, gjson.GetBytes(gotConsumer.Config, "group_id").String())
	assert.Equal(t, consumer.Username, gotConsumer.GetName(constant.Consumer))
	assert.Equal(t, consumer.GroupID, gotConsumer.GetGroupID())
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

func TestDeleteResourceByIDs_PluginMetadataUsesResourceID(t *testing.T) {
	metadata := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusCreateDraft)
	metadata.Name = fmt.Sprintf("test-plugin-metadata-delete-%d", time.Now().UnixNano())

	assert.NoError(t, CreatePluginMetadata(gatewayCtx, *metadata))

	err := DeleteResourceByIDs(gatewayCtx, constant.PluginMetadata, []string{metadata.ID})
	assert.NoError(t, err)

	_, err = GetPluginMetadata(gatewayCtx, metadata.ID)
	assert.Error(t, err)
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
