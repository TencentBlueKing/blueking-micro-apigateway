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

package biz

import (
	"context"
	"testing"

	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestGetPluginsList_ByVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		version     constant.APISIXVersion
		apisixType  string
		minExpected int
	}{
		{
			name:        "standard APISIX plugins for 3.11",
			version:     constant.APISIXVersion311,
			apisixType:  "apisix",
			minExpected: 30, // At least 30 common plugins
		},
		{
			name:        "standard APISIX plugins for 3.13",
			version:     constant.APISIXVersion313,
			apisixType:  "apisix",
			minExpected: 30,
		},
		{
			name:        "bk-apisix includes additional plugins",
			version:     constant.APISIXVersion313,
			apisixType:  "bk-apisix",
			minExpected: 40, // Should have bk-* plugins too
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plugins, err := GetPluginsList(ctx, tt.version, tt.apisixType)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, len(plugins), tt.minExpected)
		})
	}
}

func TestGetPluginsList_ContainsExpectedPlugins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plugins, err := GetPluginsList(ctx, constant.APISIXVersion313, "apisix")
	assert.NoError(t, err)

	expectedPlugins := []string{
		"limit-req",
		"limit-count",
		"limit-conn",
		"proxy-rewrite",
		"cors",
		"ip-restriction",
		"key-auth",
		"jwt-auth",
		"prometheus",
	}

	for _, expected := range expectedPlugins {
		assert.Contains(t, plugins, expected)
	}
}

func TestGetPluginsList_BKApisixHasCustomPlugins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plugins, err := GetPluginsList(ctx, constant.APISIXVersion313, "bk-apisix")
	assert.NoError(t, err)

	// Check for bk-* plugins that are actually in the bk_apisix_plugin.json file
	bkPlugins := []string{
		"bk-traffic-label",
		"bk-break-recursive-call",
		"bk-delete-cookie",
		"bk-echo",
		"bk-header-rewrite",
		"bk-jwt",
		"bk-login-required",
	}

	for _, expected := range bkPlugins {
		assert.Contains(t, plugins, expected)
	}
}

func TestGetPluginsList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		version     constant.APISIXVersion
		apisixType  string
		expectError bool
	}{
		{
			name:        "list apisix plugins",
			version:     constant.APISIXVersion313,
			apisixType:  "apisix",
			expectError: false,
		},
		{
			name:        "list bk-apisix plugins",
			version:     constant.APISIXVersion313,
			apisixType:  "bk-apisix",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			plugins, err := GetPluginsList(ctx, tt.version, tt.apisixType)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, plugins)
			}
		})
	}
}

// Note: TestListResourcesWithPagination requires ginx.SetGatewayInfoToContext
// which needs gin.Context. These functions are tested through integration tests
// and the MCP tool handlers that properly set up the context.

func TestPublishResourcesByType_EmptyIDs(t *testing.T) {
	ctx := context.Background()

	gateway := &model.Gateway{
		ID:            1,
		Name:          "test-publish-gateway",
		APISIXVersion: string(constant.APISIXVersion313),
	}

	// Test with empty resource IDs
	err := PublishResourcesByType(ctx, gateway, constant.Route, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no resources to publish")
}

// Note: TestRevertResource and TestUpdateResourceByTypeAndID
// require ginx.SetGatewayInfoToContext which needs gin.Context setup.
// These functions are tested through integration tests and MCP tool handler tests
// that properly set up the context.

func TestCreateTypedResource_PersistsTypedModelFields(t *testing.T) {
	util.InitEmbedDb()

	gateway := createMCPHelperTestGateway(t, "mcp-create-typed-route")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	route := &model.Route{
		Name:      "route-a",
		ServiceID: "svc-a",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "route-mcp-create-typed",
			GatewayID: gateway.ID,
			Config:    datatypes.JSON(`{"uris":["/test"]}`),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: "mcp",
				Updater: "mcp",
			},
		},
	}

	assert.NoError(t, CreateTypedResource(ctx, route))

	var stored model.Route
	err := database.Client().Where("gateway_id = ? AND id = ?", gateway.ID, route.ID).First(&stored).Error
	assert.NoError(t, err)
	assert.Equal(t, "route-a", stored.Name)
	assert.Equal(t, "svc-a", stored.ServiceID)
	assert.Equal(t, datatypes.JSON(`{"uris":["/test"]}`), stored.Config)
}

func TestUpdateResourceByTypeAndID_RoutePersistsResolvedColumns(t *testing.T) {
	util.InitEmbedDb()

	gateway := createMCPHelperTestGateway(t, "mcp-update-route")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	resourceID := "route-mcp-update"
	rawConfig := datatypes.JSON(`{"uri":"/new-route"}`)

	err := database.Client().Table("route").Create(map[string]any{
		"id":               resourceID,
		"gateway_id":       gateway.ID,
		"name":             "old-route",
		"service_id":       "svc-old",
		"upstream_id":      "ups-old",
		"plugin_config_id": "pc-old",
		"config":           datatypes.JSON(`{"uri":"/old-route"}`),
		"status":           constant.ResourceStatusSuccess,
		"creator":          "test",
		"updater":          "test",
	}).Error
	assert.NoError(t, err)

	err = UpdateResourceByTypeAndID(
		ctx,
		constant.Route,
		resourceID,
		rawConfig,
		constant.ResourceStatusUpdateDraft,
		ResourceResolvedValues{
			NameValue:           "new-route",
			ServiceIDValue:      "svc-new",
			UpstreamIDValue:     "ups-new",
			PluginConfigIDValue: "pc-new",
		},
	)
	assert.NoError(t, err)

	var route model.Route
	err = database.Client().Where("gateway_id = ? AND id = ?", gateway.ID, resourceID).First(&route).Error
	assert.NoError(t, err)
	assert.Equal(t, "new-route", route.Name)
	assert.Equal(t, "svc-new", route.ServiceID)
	assert.Equal(t, "ups-new", route.UpstreamID)
	assert.Equal(t, "pc-new", route.PluginConfigID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, route.Status)
	assert.Equal(t, rawConfig, route.Config)
	assert.False(t, gjson.GetBytes(route.Config, "service_id").Exists())
	assert.False(t, gjson.GetBytes(route.Config, "upstream_id").Exists())
	assert.False(t, gjson.GetBytes(route.Config, "plugin_config_id").Exists())
	assert.False(t, gjson.GetBytes(route.Config, "name").Exists())
}

func TestUpdateResourceByTypeAndID_ConsumerPersistsGroupID(t *testing.T) {
	util.InitEmbedDb()

	gateway := createMCPHelperTestGateway(t, "mcp-update-consumer")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	resourceID := "consumer-mcp-update"
	rawConfig := datatypes.JSON(`{"plugins":{"key-auth":{}}}`)

	err := database.Client().Table("consumer").Create(map[string]any{
		"id":         resourceID,
		"gateway_id": gateway.ID,
		"username":   "old-consumer",
		"group_id":   "group-old",
		"config":     datatypes.JSON(`{"plugins":{"key-auth":{}}}`),
		"status":     constant.ResourceStatusSuccess,
		"creator":    "test",
		"updater":    "test",
	}).Error
	assert.NoError(t, err)

	err = UpdateResourceByTypeAndID(
		ctx,
		constant.Consumer,
		resourceID,
		rawConfig,
		constant.ResourceStatusUpdateDraft,
		ResourceResolvedValues{
			NameValue:    "new-consumer",
			GroupIDValue: "group-new",
		},
	)
	assert.NoError(t, err)

	var consumer model.Consumer
	err = database.Client().Where("gateway_id = ? AND id = ?", gateway.ID, resourceID).First(&consumer).Error
	assert.NoError(t, err)
	assert.Equal(t, "new-consumer", consumer.Username)
	assert.Equal(t, "group-new", consumer.GroupID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, consumer.Status)
	assert.Equal(t, rawConfig, consumer.Config)
	assert.False(t, gjson.GetBytes(consumer.Config, "group_id").Exists())
	assert.False(t, gjson.GetBytes(consumer.Config, "username").Exists())
}

func TestUpdateResourceByTypeAndID_NameOnlyFallbackUpdatesName(t *testing.T) {
	util.InitEmbedDb()

	gateway := createMCPHelperTestGateway(t, "mcp-update-name-only")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	resourceID := "service-mcp-update"
	rawConfig := datatypes.JSON(`{"hosts":["example.com"]}`)

	err := database.Client().Table("service").Create(map[string]any{
		"id":          resourceID,
		"gateway_id":  gateway.ID,
		"name":        "old-service",
		"upstream_id": "ups-old",
		"config":      datatypes.JSON(`{"hosts":["old.example.com"]}`),
		"status":      constant.ResourceStatusSuccess,
		"creator":     "test",
		"updater":     "test",
	}).Error
	assert.NoError(t, err)

	err = UpdateResourceByTypeAndID(
		ctx,
		constant.Service,
		resourceID,
		rawConfig,
		constant.ResourceStatusUpdateDraft,
		ResourceResolvedValues{NameValue: "new-service"},
	)
	assert.NoError(t, err)

	var service model.Service
	err = database.Client().Where("gateway_id = ? AND id = ?", gateway.ID, resourceID).First(&service).Error
	assert.NoError(t, err)
	assert.Equal(t, "new-service", service.Name)
	assert.Equal(t, "ups-old", service.UpstreamID)
	assert.Equal(t, rawConfig, service.Config)
	assert.False(t, gjson.GetBytes(service.Config, "name").Exists())
}

func TestRevertResource_RestoresSyncedResolvedColumns(t *testing.T) {
	util.InitEmbedDb()

	gateway := createMCPHelperTestGateway(t, "mcp-revert-route")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	resourceID := "route-mcp-revert"
	expectedStoredConfig := datatypes.JSON(`{"uri":"/synced-route"}`)
	syncedConfig := datatypes.JSON(`{
		"name":"synced-route",
		"service_id":"svc-synced",
		"upstream_id":"ups-synced",
		"plugin_config_id":"pc-synced",
		"uri":"/synced-route"
	}`)

	err := database.Client().Table("route").Create(map[string]any{
		"id":               resourceID,
		"gateway_id":       gateway.ID,
		"name":             "draft-route",
		"service_id":       "svc-draft",
		"upstream_id":      "ups-draft",
		"plugin_config_id": "pc-draft",
		"config":           datatypes.JSON(`{"uri":"/draft-route"}`),
		"status":           constant.ResourceStatusUpdateDraft,
		"creator":          "test",
		"updater":          "test",
	}).Error
	assert.NoError(t, err)

	err = database.Client().Table("gateway_sync_data").Create(map[string]any{
		"id":           resourceID,
		"gateway_id":   gateway.ID,
		"type":         constant.Route.String(),
		"config":       syncedConfig,
		"mod_revision": 1,
	}).Error
	assert.NoError(t, err)

	err = RevertResource(ctx, constant.Route, resourceID)
	assert.NoError(t, err)

	var route model.Route
	err = database.Client().Where("gateway_id = ? AND id = ?", gateway.ID, resourceID).First(&route).Error
	assert.NoError(t, err)
	assert.Equal(t, "synced-route", route.Name)
	assert.Equal(t, "svc-synced", route.ServiceID)
	assert.Equal(t, "ups-synced", route.UpstreamID)
	assert.Equal(t, "pc-synced", route.PluginConfigID)
	assert.Equal(t, constant.ResourceStatusSuccess, route.Status)
	assert.JSONEq(t, string(expectedStoredConfig), string(route.Config))
	assert.False(t, gjson.GetBytes(route.Config, "name").Exists())
	assert.False(t, gjson.GetBytes(route.Config, "service_id").Exists())
	assert.False(t, gjson.GetBytes(route.Config, "upstream_id").Exists())
	assert.False(t, gjson.GetBytes(route.Config, "plugin_config_id").Exists())
}

func createMCPHelperTestGateway(t *testing.T, name string) *model.Gateway {
	t.Helper()

	gateway := &model.Gateway{
		Name:          name,
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := CreateGateway(context.Background(), gateway)
	assert.NoError(t, err)
	assert.Greater(t, gateway.ID, 0)
	return gateway
}
