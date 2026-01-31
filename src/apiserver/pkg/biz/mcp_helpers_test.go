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

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
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

// Note: TestRevertResource, TestUpdateResourceByTypeAndID, and TestCreateResource
// require ginx.SetGatewayInfoToContext which needs gin.Context setup.
// These functions are tested through integration tests and MCP tool handler tests
// that properly set up the context.
