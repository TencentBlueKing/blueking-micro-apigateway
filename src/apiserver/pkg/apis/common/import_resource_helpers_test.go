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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestApplyImportIgnoreFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		imported     string
		existing     string
		ignoreFields []string
		want         string
	}{
		{
			name:         "overlay top level field from existing config",
			imported:     `{"name":"route-a","desc":"new-desc","plugins":{}}`,
			existing:     `{"name":"route-a","desc":"old-desc","plugins":{"limit-count":{"count":1}}}`,
			ignoreFields: []string{"desc"},
			want:         `{"name":"route-a","desc":"old-desc","plugins":{}}`,
		},
		{
			name:         "overlay nested field from existing config",
			imported:     `{"plugins":{"limit-count":{"count":10,"time_window":60}}}`,
			existing:     `{"plugins":{"limit-count":{"count":1,"time_window":120}}}`,
			ignoreFields: []string{"plugins.limit-count.count"},
			want:         `{"plugins":{"limit-count":{"count":1,"time_window":60}}}`,
		},
		{
			name:         "ignore missing field keeps imported config",
			imported:     `{"plugins":{}}`,
			existing:     `{"name":"route-a"}`,
			ignoreFields: []string{"plugins.limit-count"},
			want:         `{"plugins":{}}`,
		},
		{
			name:         "partial missing fields only overlays existing fields",
			imported:     `{"desc":"new-desc","plugins":{"limit-count":{"count":10}}}`,
			existing:     `{"desc":"old-desc"}`,
			ignoreFields: []string{"desc", "plugins.limit-count.count"},
			want:         `{"desc":"old-desc","plugins":{"limit-count":{"count":10}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyImportIgnoreFields(
				json.RawMessage(tt.imported),
				datatypes.JSON([]byte(tt.existing)),
				tt.ignoreFields,
			)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestLoadExistingImportResources(t *testing.T) {
	util.InitEmbedDb()

	ctx := context.Background()
	gateway := &model.Gateway{
		Name:          "import-test-gateway",
		APISIXVersion: string(constant.APISIXVersion313),
	}
	assert.NoError(t, biz.CreateGateway(ctx, gateway))

	gatewayCtx := ginx.SetGatewayInfoToContext(ctx, gateway)
	assert.NoError(t, biz.CreatePluginConfig(gatewayCtx, model.PluginConfig{
		Name: "pc-demo",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "pc-1",
			GatewayID: gateway.ID,
			Config:    datatypes.JSON([]byte(`{"id":"pc-1","name":"pc-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`)),
			Status:    constant.ResourceStatusSuccess,
		},
	}))

	allResourceIDs := map[string]struct{}{}
	got, err := loadExistingImportResources(gatewayCtx, constant.PluginConfig, allResourceIDs)
	assert.NoError(t, err)
	assert.Contains(t, got, fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-1"))
	assert.Contains(t, allResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-1"))

	t.Run("empty DB returns empty map", func(t *testing.T) {
		empty := map[string]struct{}{}
		got, err := loadExistingImportResources(gatewayCtx, constant.Upstream, empty)
		assert.NoError(t, err)
		assert.Empty(t, got)
		assert.Empty(t, empty)
	})
}

func TestBuildImportSyncData(t *testing.T) {
	t.Parallel()

	ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ID: 23})
	info := &ResourceInfo{
		ResourceType: constant.Route,
		ResourceID:   "route-1",
		Name:         "route-demo",
		Config:       json.RawMessage(`{"id":"route-1","name":"route-demo","uri":"/demo"}`),
	}

	got := buildImportSyncData(ctx, constant.Route, info)
	assert.Equal(t, constant.Route, got.Type)
	assert.Equal(t, "route-1", got.ID)
	assert.Equal(t, 23, got.GatewayID)
	assert.JSONEq(t, `{"id":"route-1","name":"route-demo","uri":"/demo"}`, string(got.Config))
}

func TestPrepareImportResources(t *testing.T) {
	util.InitEmbedDb()

	ctx := context.Background()
	gateway := &model.Gateway{
		Name:          "prepare-import-gateway",
		APISIXVersion: string(constant.APISIXVersion313),
	}
	assert.NoError(t, biz.CreateGateway(ctx, gateway))
	gatewayCtx := ginx.SetGatewayInfoToContext(ctx, gateway)

	existing := model.PluginConfig{
		Name: "pc-demo",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "pc-1",
			GatewayID: gateway.ID,
			Config: datatypes.JSON([]byte(
				`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
			)),
			Status: constant.ResourceStatusSuccess,
		},
	}
	assert.NoError(t, biz.CreatePluginConfig(gatewayCtx, existing))

	resources, err := prepareImportResources(
		gatewayCtx,
		map[constant.APISIXResource][]*ResourceInfo{
			constant.PluginConfig: {
				{
					ResourceType: constant.PluginConfig,
					ResourceID:   "pc-1",
					Name:         "pc-demo",
					Config: json.RawMessage(
						`{"id":"pc-1","name":"pc-demo","desc":"new-desc","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
					),
				},
			},
		},
		map[string]struct{}{},
		map[constant.APISIXResource][]string{
			constant.PluginConfig: {"desc"},
		},
	)
	assert.NoError(t, err)
	if !assert.Len(t, resources[constant.PluginConfig], 1) {
		return
	}
	assert.JSONEq(
		t,
		`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		string(resources[constant.PluginConfig][0].Config),
	)

	t.Run("schema resources are skipped", func(t *testing.T) {
		got, err := prepareImportResources(
			gatewayCtx,
			map[constant.APISIXResource][]*ResourceInfo{
				constant.Schema: {
					{
						ResourceType: constant.Schema,
						ResourceID:   "schema-1",
						Name:         "demo-plugin",
						Config:       json.RawMessage(`{"name":"demo-plugin"}`),
					},
				},
				constant.Route: {
					{
						ResourceType: constant.Route,
						ResourceID:   "route-1",
						Name:         "route-demo",
						Config:       json.RawMessage(`{"id":"route-1","name":"route-demo","uris":["/demo"]}`),
					},
				},
			},
			map[string]struct{}{},
			nil,
		)
		assert.NoError(t, err)
		assert.NotContains(t, got, constant.Schema)
		assert.Len(t, got[constant.Route], 1)
	})
}

func TestPrepareImportValidationInput(t *testing.T) {
	t.Parallel()

	ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ID: 31})

	t.Run("add only", func(t *testing.T) {
		input, err := prepareImportValidationInput(
			ctx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-1",
							Name:         "route-demo",
							Config:       json.RawMessage(`{"id":"route-1","name":"route-demo","uri":"/demo"}`),
						},
					},
				},
				Update: map[constant.APISIXResource][]*ResourceInfo{},
			},
			nil,
		)
		assert.NoError(t, err)
		assert.Contains(t, input.AllResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-1"))
		assert.Len(t, input.Add, 1)
		assert.Len(t, input.Add[constant.Route], 1)
		assert.Empty(t, input.Update)
	})

	t.Run("add and update accumulate all resource ids", func(t *testing.T) {
		input, err := prepareImportValidationInput(
			ctx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-new",
							Name:         "route-new",
							Config:       json.RawMessage(`{"id":"route-new","uri":"/a"}`),
						},
					},
				},
				Update: map[constant.APISIXResource][]*ResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-upd",
							Name:         "route-upd",
							Config:       json.RawMessage(`{"id":"route-upd","uri":"/b"}`),
						},
					},
				},
			},
			nil,
		)
		assert.NoError(t, err)
		assert.Contains(t, input.AllResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-new"))
		assert.Contains(t, input.AllResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-upd"))
		assert.Len(t, input.Add[constant.Route], 1)
		assert.Len(t, input.Update[constant.Route], 1)
	})
}
