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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

// TestIsResourceChanged_Route_NameChanged tests that route name changes are detected
func TestIsResourceChanged_Route_NameChanged(t *testing.T) {
	// Create a test route
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-name-%d", time.Now().UnixNano())
	route.Config = datatypes.JSON(`{"uri": "/test"}`)

	// Create the resource
	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	// Test: Name changed, config same
	configJSON := json.RawMessage(`{"uri": "/test"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Route, route.ID, configJSON, map[string]any{
		"name":             "new-name", // Changed
		"service_id":       "",
		"upstream_id":      "",
		"plugin_config_id": "",
	})
	assert.True(t, changed, "Should detect name change even when config is same")
}

// TestIsResourceChanged_Route_ServiceIDChanged tests that service_id changes are detected
func TestIsResourceChanged_Route_ServiceIDChanged(t *testing.T) {
	// Create a test route
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-svc-%d", time.Now().UnixNano())
	route.ServiceID = "service-123"
	route.Config = datatypes.JSON(`{"uri": "/test"}`)

	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	// Test: ServiceID changed, config same
	configJSON := json.RawMessage(`{"uri": "/test"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Route, route.ID, configJSON, map[string]any{
		"name":             route.Name,
		"service_id":       "service-999", // Changed
		"upstream_id":      "",
		"plugin_config_id": "",
	})
	assert.True(t, changed, "Should detect service_id change even when config is same")
}

// TestIsResourceChanged_Route_NoChanges tests that no changes are correctly detected
func TestIsResourceChanged_Route_NoChanges(t *testing.T) {
	// Create a test route
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-nochange-%d", time.Now().UnixNano())
	route.Config = datatypes.JSON(`{"uri": "/test"}`)

	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	// Get the route from DB to get exact config format
	retrievedRoute, err := GetRoute(gatewayCtx, route.ID)
	assert.NoError(t, err)

	// Test: Nothing changed
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
	// Create a test consumer
	consumer := data.Consumer1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft)
	consumer.Username = fmt.Sprintf("test-user-%d", time.Now().UnixNano())
	consumer.GroupID = "group-123"
	consumer.Config = datatypes.JSON(`{"username": "` + consumer.Username + `"}`)

	err := CreateConsumer(gatewayCtx, *consumer)
	assert.NoError(t, err)

	// Test: Username changed, config same
	configJSON := json.RawMessage(`{"username": "` + consumer.Username + `"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Consumer, consumer.ID, configJSON, map[string]any{
		"username": "new-user", // Changed
		"group_id": "group-123",
	})
	assert.True(t, changed, "Should detect username change even when config is same")
}

// TestIsResourceChanged_Service_UpstreamIDChanged tests service upstream_id changes
func TestIsResourceChanged_Service_UpstreamIDChanged(t *testing.T) {
	// Create a test service
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

	// Test: UpstreamID changed, config same
	configJSON := json.RawMessage(`{"name": "test-service"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Service, service.ID, configJSON, map[string]any{
		"name":        service.Name,
		"upstream_id": "upstream-999", // Changed
	})
	assert.True(t, changed, "Should detect upstream_id change even when config is same")
}

// TestIsResourceChanged_Route_ConfigAndNameChanged tests both config and extra fields changed
func TestIsResourceChanged_Route_ConfigAndNameChanged(t *testing.T) {
	// Create a test route
	route := data.Route1WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("test-route-both-%d", time.Now().UnixNano())
	route.Config = datatypes.JSON(`{"uri": "/old"}`)

	err := CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	// Test: Both config and name changed
	configJSON := json.RawMessage(`{"uri": "/new"}`)
	changed := IsResourceChanged(gatewayCtx, constant.Route, route.ID, configJSON, map[string]any{
		"name":             "new-name", // Changed
		"service_id":       "",
		"upstream_id":      "",
		"plugin_config_id": "",
	})
	assert.True(t, changed, "Should detect changes when both config and name changed")
}

// TestIsResourceChanged_Upstream_NameAndSSLIDChanged tests upstream name and ssl_id changes
func TestIsResourceChanged_Upstream_NameAndSSLIDChanged(t *testing.T) {
	// Create a test upstream
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

	// Test: Name changed
	retrievedUpstream, err := GetUpstream(gatewayCtx, upstream.ID)
	assert.NoError(t, err)
	configJSON := json.RawMessage(retrievedUpstream.Config)
	changed := IsResourceChanged(gatewayCtx, constant.Upstream, upstream.ID, configJSON, map[string]any{
		"name":   "new-upstream-name", // Changed
		"ssl_id": "ssl-123",
	})
	assert.True(t, changed, "Should detect name change for upstream")

	// Test: SSLID changed
	changed = IsResourceChanged(gatewayCtx, constant.Upstream, upstream.ID, configJSON, map[string]any{
		"name":   retrievedUpstream.Name,
		"ssl_id": "ssl-999", // Changed
	})
	assert.True(t, changed, "Should detect ssl_id change for upstream")
}

// TestIsResourceChanged_Proto_NameChanged tests proto name changes
func TestIsResourceChanged_Proto_NameChanged(t *testing.T) {
	// Create a test proto
	proto := data.Proto1(gatewayInfo, constant.ResourceStatusCreateDraft)
	proto.Name = fmt.Sprintf("test-proto-%d", time.Now().UnixNano())

	err := CreateProto(gatewayCtx, *proto)
	assert.NoError(t, err)

	// Get the proto from DB to get exact config format
	retrievedProto, err := GetProto(gatewayCtx, proto.ID)
	assert.NoError(t, err)

	// Test: Name changed, config same
	configJSON := json.RawMessage(retrievedProto.Config)
	changed := IsResourceChanged(gatewayCtx, constant.Proto, proto.ID, configJSON, map[string]any{
		"name": "new-proto-name", // Changed
	})
	assert.True(t, changed, "Should detect name change for proto")
}
