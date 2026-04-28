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

package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func decodeToolResultJSON(t *testing.T, result *mcp.CallToolResult, target any) {
	t.Helper()

	if !assert.NotNil(t, result) {
		return
	}
	if !assert.False(t, result.IsError) {
		return
	}
	if !assert.Len(t, result.Content, 1) {
		return
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !assert.True(t, ok) {
		return
	}
	assert.NoError(t, json.Unmarshal([]byte(textContent.Text), target))
}

func createMCPCRUDTestGateway(t *testing.T, name string) *model.Gateway {
	t.Helper()

	util.InitEmbedDb()

	gateway := &model.Gateway{
		Name:          name,
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := biz.CreateGateway(context.Background(), gateway)
	assert.NoError(t, err)
	assert.Greater(t, gateway.ID, 0)
	return gateway
}

func TestGetResourceHandlerReturnsRawStoredConfig(t *testing.T) {
	gateway := createMCPCRUDTestGateway(t, "mcp-get-raw-route")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	route := model.Route{
		Name:      "route-a",
		ServiceID: "svc-a",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "route-mcp-get-raw",
			GatewayID: gateway.ID,
			Config:    datatypes.JSON(`{"uris":["/test"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	assert.NoError(t, biz.CreateRoute(ctx, route))

	result, _, err := getResourceHandler(ctx, nil, GetResourceInput{
		ResourceType: constant.Route.String(),
		ResourceID:   route.ID,
	})
	assert.NoError(t, err)

	var payload struct {
		ID     string         `json:"id"`
		Config map[string]any `json:"config"`
	}
	decodeToolResultJSON(t, result, &payload)
	assert.Equal(t, route.ID, payload.ID)
	assert.Equal(t, []any{"/test"}, payload.Config["uris"])
	assert.NotContains(t, payload.Config, "name")
	assert.NotContains(t, payload.Config, "service_id")
}

func TestListResourceHandlerReturnsRawStoredConfig(t *testing.T) {
	gateway := createMCPCRUDTestGateway(t, "mcp-list-raw-route")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	route := model.Route{
		Name:      "route-b",
		ServiceID: "svc-b",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "route-mcp-list-raw",
			GatewayID: gateway.ID,
			Config:    datatypes.JSON(`{"uris":["/list-test"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	assert.NoError(t, biz.CreateRoute(ctx, route))

	result, _, err := listResourceHandler(ctx, nil, ListResourceInput{
		ResourceType: constant.Route.String(),
		Page:         1,
		PageSize:     10,
	})
	assert.NoError(t, err)

	var payload struct {
		Resources []struct {
			ID     string `json:"id"`
			Config string `json:"config"`
		} `json:"resources"`
	}
	decodeToolResultJSON(t, result, &payload)
	if assert.NotEmpty(t, payload.Resources) {
		var target *struct {
			ID     string `json:"id"`
			Config string `json:"config"`
		}
		for i := range payload.Resources {
			if payload.Resources[i].ID == route.ID {
				target = &payload.Resources[i]
				break
			}
		}
		if assert.NotNil(t, target) {
			var config map[string]any
			assert.NoError(t, json.Unmarshal([]byte(target.Config), &config))
			assert.Equal(t, []any{"/list-test"}, config["uris"])
			assert.NotContains(t, config, "name")
			assert.NotContains(t, config, "service_id")
		}
	}
}

func TestCreateResourceHandlerFallbackPersistsStrippedRouteConfig(t *testing.T) {
	gateway := createMCPCRUDTestGateway(t, "mcp-create-route-fallback")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	result, _, err := createResourceHandler(ctx, nil, CreateResourceInput{
		ResourceType: constant.Route.String(),
		Name:         "outer-route",
		Config: map[string]any{
			"name":       "inner-route",
			"service_id": "svc-a",
			"uris":       []any{"/test"},
		},
	})
	assert.NoError(t, err)

	var payload struct {
		ResourceID string `json:"resource_id"`
	}
	decodeToolResultJSON(t, result, &payload)
	assert.NotEmpty(t, payload.ResourceID)

	stored, err := biz.GetRoute(ctx, payload.ResourceID)
	assert.NoError(t, err)
	assert.Equal(t, "outer-route", stored.Name)
	assert.Equal(t, "svc-a", stored.ServiceID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, stored.Status)
	assert.JSONEq(t, `{"uris":["/test"]}`, string(stored.Config))
}

func TestCreateResourceHandlerPersistsConsumerResolvedFields(t *testing.T) {
	gateway := createMCPCRUDTestGateway(t, "mcp-create-consumer-success")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	result, _, err := createResourceHandler(ctx, nil, CreateResourceInput{
		ResourceType: constant.Consumer.String(),
		Name:         "consumer-a",
		Config: map[string]any{
			"username": "consumer-a",
			"group_id": "group-a",
			"plugins": map[string]any{
				"key-auth": map[string]any{
					"key": "token-a",
				},
			},
		},
	})
	assert.NoError(t, err)

	var payload struct {
		ResourceID string `json:"resource_id"`
	}
	decodeToolResultJSON(t, result, &payload)
	assert.NotEmpty(t, payload.ResourceID)

	stored, err := biz.GetConsumer(ctx, payload.ResourceID)
	assert.NoError(t, err)
	assert.Equal(t, "consumer-a", stored.Username)
	assert.Equal(t, "group-a", stored.GroupID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, stored.Status)
	assert.JSONEq(t, `{"plugins":{"key-auth":{"key":"token-a"}}}`, string(stored.Config))
}
