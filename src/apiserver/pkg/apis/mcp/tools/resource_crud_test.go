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
	"fmt"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	gatewaybiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/gateway"
	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestBuildBatchOperationResultPartialFailure(t *testing.T) {
	t.Parallel()

	failures := []batchOperationFailure{
		{
			ResourceID: "r1",
			Stage:      "delete",
			Error:      "failed to delete",
		},
	}

	result := buildBatchOperationResult(
		"Delete operation completed",
		"route",
		2,
		1,
		map[string]int{
			"hard_deleted_count":  1,
			"marked_delete_count": 0,
		},
		failures,
	)

	assert.Equal(t, 2, result["total_requested"])
	assert.Equal(t, 1, result["success_count"])
	assert.Equal(t, 1, result["failed_count"])
	assert.Equal(t, true, result["partial_success"])
	assert.Equal(t, "route", result["resource_type"])
	assert.Equal(t, 1, result["hard_deleted_count"])
	assert.Equal(t, 0, result["marked_delete_count"])
}

func TestBuildBatchOperationResultAllSuccess(t *testing.T) {
	t.Parallel()

	result := buildBatchOperationResult(
		"Revert operation completed",
		"service",
		3,
		3,
		map[string]int{
			"reverted_count": 3,
		},
		nil,
	)

	assert.Equal(t, 3, result["total_requested"])
	assert.Equal(t, 3, result["success_count"])
	assert.Equal(t, 0, result["failed_count"])
	assert.Equal(t, false, result["partial_success"])
	assert.Equal(t, 3, result["reverted_count"])
}

// Note: UpdateResourceInput.Config is currently typed as map[string]any, so the
// "invalid JSON reaches sjson.SetBytes and gets silently ignored" path described
// in the refactor plan is not reachable from this handler seam today.
func TestCreateResourceHandlerInjectsRouteNameIntoConfig(t *testing.T) {
	ctx, gateway := newMCPToolTestContext(t)
	route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
	inputName := fmt.Sprintf("mcp-route-create-%d", time.Now().UnixNano())

	result, _, err := createResourceHandler(ctx, nil, CreateResourceInput{
		ResourceType: constant.Route.String(),
		Name:         inputName,
		Config:       mustDecodeConfigMap(t, route.Config),
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)

	payload := mustDecodeResultPayload(t, result)
	resourceID, _ := payload["resource_id"].(string)
	assert.NotEmpty(t, resourceID)
	assert.Equal(t, string(constant.ResourceStatusCreateDraft), payload["status"])

	createdRoute, err := resourcebiz.GetRoute(ctx, resourceID)
	assert.NoError(t, err)
	assert.Equal(t, inputName, createdRoute.Name)
	assert.Equal(t, inputName, gjson.GetBytes(createdRoute.Config, "name").String())
	assert.Equal(t, "/get", gjson.GetBytes(createdRoute.Config, "uris.0").String())
	assert.Equal(t, constant.ResourceStatusCreateDraft, createdRoute.Status)
}

func TestCreateResourceHandlerInjectsConsumerUsernameIntoConfig(t *testing.T) {
	ctx, gateway := newMCPToolTestContext(t)
	consumer := data.Consumer1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
	inputName := fmt.Sprintf("mcp-consumer-create-%d", time.Now().UnixNano())

	result, _, err := createResourceHandler(ctx, nil, CreateResourceInput{
		ResourceType: constant.Consumer.String(),
		Name:         inputName,
		Config:       mustDecodeConfigMap(t, consumer.Config),
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)

	payload := mustDecodeResultPayload(t, result)
	resourceID, _ := payload["resource_id"].(string)
	assert.NotEmpty(t, resourceID)
	assert.Equal(t, string(constant.ResourceStatusCreateDraft), payload["status"])

	createdConsumer, err := resourcebiz.GetConsumer(ctx, resourceID)
	assert.NoError(t, err)
	assert.Equal(t, inputName, createdConsumer.Username)
	assert.Equal(t, inputName, gjson.GetBytes(createdConsumer.Config, "username").String())
	assert.False(t, gjson.GetBytes(createdConsumer.Config, "name").Exists())
	assert.Equal(t, constant.ResourceStatusCreateDraft, createdConsumer.Status)
}

func TestUpdateResourceHandlerSyncsTypedNameIntoConfig(t *testing.T) {
	ctx, gateway := newMCPToolTestContext(t)
	route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusSuccess)
	route.Name = fmt.Sprintf("mcp-route-before-update-%d", time.Now().UnixNano())
	assert.NoError(t, resourcebiz.CreateRoute(ctx, *route))

	storedRoute, err := resourcebiz.GetRoute(ctx, route.ID)
	assert.NoError(t, err)

	inputConfig := mustDecodeConfigMap(t, storedRoute.Config)
	inputConfig["name"] = "stale-route-name"
	inputConfig["uris"] = []string{"/updated"}
	inputName := fmt.Sprintf("mcp-route-after-update-%d", time.Now().UnixNano())

	result, _, err := updateResourceHandler(ctx, nil, UpdateResourceInput{
		ResourceType: constant.Route.String(),
		ResourceID:   route.ID,
		Name:         inputName,
		Config:       inputConfig,
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)

	payload := mustDecodeResultPayload(t, result)
	assert.Equal(t, string(constant.ResourceStatusUpdateDraft), payload["status"])

	updatedRoute, err := resourcebiz.GetRoute(ctx, route.ID)
	assert.NoError(t, err)
	assert.Equal(t, inputName, updatedRoute.Name)
	assert.Equal(t, inputName, gjson.GetBytes(updatedRoute.Config, "name").String())
	assert.Equal(t, "/updated", gjson.GetBytes(updatedRoute.Config, "uris.0").String())
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedRoute.Status)
}

func TestUpdateResourceHandlerWithoutNamePreservesConfigShape(t *testing.T) {
	ctx, gateway := newMCPToolTestContext(t)
	route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
	route.Name = fmt.Sprintf("mcp-route-no-name-%d", time.Now().UnixNano())
	assert.NoError(t, resourcebiz.CreateRoute(ctx, *route))

	storedRoute, err := resourcebiz.GetRoute(ctx, route.ID)
	assert.NoError(t, err)

	result, _, err := updateResourceHandler(ctx, nil, UpdateResourceInput{
		ResourceType: constant.Route.String(),
		ResourceID:   route.ID,
		Config:       mustDecodeConfigMap(t, storedRoute.Config),
	})

	assert.NoError(t, err)
	assert.False(t, result.IsError)

	payload := mustDecodeResultPayload(t, result)
	assert.Equal(t, string(constant.ResourceStatusCreateDraft), payload["status"])

	updatedRoute, err := resourcebiz.GetRoute(ctx, route.ID)
	assert.NoError(t, err)
	assert.Equal(t, storedRoute.Name, updatedRoute.Name)
	assert.JSONEq(t, string(storedRoute.Config), string(updatedRoute.Config))
	assert.Equal(t, constant.ResourceStatusCreateDraft, updatedRoute.Status)
}

func newMCPToolTestContext(t *testing.T) (context.Context, *model.Gateway) {
	t.Helper()

	util.InitEmbedDb()

	ctx := context.Background()
	gateway := &model.Gateway{
		Name:          fmt.Sprintf("mcp-tools-gateway-%d", time.Now().UnixNano()),
		APISIXVersion: string(constant.APISIXVersion313),
	}

	err := gatewaybiz.CreateGateway(ctx, gateway)
	assert.NoError(t, err)
	assert.Greater(t, gateway.ID, 0)

	return ginx.SetGatewayInfoToContext(ctx, gateway), gateway
}

func mustDecodeConfigMap(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	var config map[string]any
	err := json.Unmarshal(raw, &config)
	assert.NoError(t, err)

	return config
}

func mustDecodeResultPayload(t *testing.T, result *mcp.CallToolResult) map[string]any {
	t.Helper()

	if !assert.NotNil(t, result) {
		return nil
	}
	if !assert.Len(t, result.Content, 1) {
		return nil
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !assert.True(t, ok) {
		return nil
	}

	var payload map[string]any
	err := json.Unmarshal([]byte(textContent.Text), &payload)
	assert.NoError(t, err)

	return payload
}
