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

func TestHandleUploadResources(t *testing.T) {
	t.Run("ignore_fields overlays existing config field", func(t *testing.T) {
		gatewayCtx, gateway := setupImportGatewayContext(t, "import-overlay")
		createPluginConfigForImportTest(
			t,
			gatewayCtx,
			gateway.ID,
			"pc-1",
			`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		)

		got, err := HandleUploadResources(
			gatewayCtx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{},
				Update: map[constant.APISIXResource][]*ResourceInfo{
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
			},
			map[string]any{},
			map[constant.APISIXResource][]string{
				constant.PluginConfig: {"desc"},
			},
		)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, got.UpdateResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		assert.JSONEq(
			t,
			`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
			string(got.UpdateResourceTypeMap[constant.PluginConfig][0].Config),
		)
	})

	t.Run("missing ignore_fields source keeps imported config", func(t *testing.T) {
		gatewayCtx, gateway := setupImportGatewayContext(t, "import-ignore-missing")
		createPluginConfigForImportTest(
			t,
			gatewayCtx,
			gateway.ID,
			"pc-2",
			`{"id":"pc-2","name":"pc-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		)

		got, err := HandleUploadResources(
			gatewayCtx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{},
				Update: map[constant.APISIXResource][]*ResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							ResourceID:   "pc-2",
							Name:         "pc-demo",
							Config: json.RawMessage(
								`{"id":"pc-2","name":"pc-demo","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","rejected_code":503,"policy":"local"}}}`,
							),
						},
					},
				},
			},
			map[string]any{},
			map[constant.APISIXResource][]string{
				constant.PluginConfig: {"plugins.limit-count.rejected_code"},
			},
		)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, got.UpdateResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		assert.JSONEq(
			t,
			`{"id":"pc-2","name":"pc-demo","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","rejected_code":503,"policy":"local"}}}`,
			string(got.UpdateResourceTypeMap[constant.PluginConfig][0].Config),
		)
	})

	t.Run("keeps provided resource ids and add update counts", func(t *testing.T) {
		gatewayCtx, gateway := setupImportGatewayContext(t, "import-id-preserve")
		createPluginConfigForImportTest(
			t,
			gatewayCtx,
			gateway.ID,
			"pc-existing",
			`{"id":"pc-existing","name":"pc-existing","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		)

		got, err := HandleUploadResources(
			gatewayCtx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							ResourceID:   "pc-new",
							Name:         "pc-new",
							Config: json.RawMessage(
								`{"id":"pc-new","name":"pc-new","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
							),
						},
					},
				},
				Update: map[constant.APISIXResource][]*ResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							ResourceID:   "pc-existing",
							Name:         "pc-existing",
							Config: json.RawMessage(
								`{"id":"pc-existing","name":"pc-existing-updated","plugins":{"limit-count":{"count":20,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
							),
						},
					},
				},
			},
			map[string]any{},
			nil,
		)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, got.AddResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		if !assert.Len(t, got.UpdateResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		assert.Equal(t, "pc-new", got.AddResourceTypeMap[constant.PluginConfig][0].ID)
		assert.Equal(t, "pc-existing", got.UpdateResourceTypeMap[constant.PluginConfig][0].ID)
	})

	t.Run("empty resource id fails before upload", func(t *testing.T) {
		gatewayCtx, _ := setupImportGatewayContext(t, "import-empty-id")

		got, err := HandleUploadResources(
			gatewayCtx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							Name:         "pc-empty-id",
							Config:       json.RawMessage(`{"name":"pc-empty-id"}`),
						},
					},
				},
				Update: map[constant.APISIXResource][]*ResourceInfo{},
			},
			map[string]any{},
			nil,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "resource id is empty")
	})

	t.Run("missing associated resource fails during upload handling", func(t *testing.T) {
		gatewayCtx, _ := setupImportGatewayContext(t, "import-missing-association")

		got, err := HandleUploadResources(
			gatewayCtx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-1",
							Name:         "route-demo",
							Config: json.RawMessage(
								`{"id":"route-1","name":"route-demo","uris":["/demo"],"upstream_id":"up-missing"}`,
							),
						},
					},
				},
				Update: map[constant.APISIXResource][]*ResourceInfo{},
			},
			map[string]any{},
			nil,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "associated upstream [id:up-missing] not found")
	})
}

func TestHandlerResourceIndexMap(t *testing.T) {
	gatewayCtx, gateway := setupImportGatewayContext(t, "import-index-map")
	createPluginConfigForImportTest(
		t,
		gatewayCtx,
		gateway.ID,
		"pc-existing",
		`{"id":"pc-existing","name":"pc-existing","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
	)

	got, err := HandlerResourceIndexMap(
		gatewayCtx,
		map[constant.APISIXResource][]*ResourceInfo{
			constant.PluginConfig: {
				{
					ResourceType: constant.PluginConfig,
					ResourceID:   "pc-new",
					Name:         "pc-new",
					Config: json.RawMessage(
						`{"id":"pc-new","name":"pc-new","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
					),
				},
			},
		},
	)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, got.ResourceTypeMap[constant.PluginConfig], 1) {
		return
	}
	assert.Contains(
		t,
		got.ExistsResourceIdList,
		fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-existing"),
	)
	assert.Contains(
		t,
		got.AllResourceIdList,
		fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-existing"),
	)
	assert.Contains(
		t,
		got.AllResourceIdList,
		fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-new"),
	)
	assert.Equal(t, "pc-new", got.ResourceTypeMap[constant.PluginConfig][0].ID)
}

func setupImportGatewayContext(t *testing.T, gatewayName string) (context.Context, *model.Gateway) {
	t.Helper()

	util.InitEmbedDb()

	ctx := context.Background()
	gateway := &model.Gateway{
		Name:          gatewayName,
		APISIXVersion: string(constant.APISIXVersion313),
	}
	err := biz.CreateGateway(ctx, gateway)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return ginx.SetGatewayInfoToContext(ctx, gateway), gateway
}

func createPluginConfigForImportTest(
	t *testing.T,
	ctx context.Context,
	gatewayID int,
	id string,
	config string,
) {
	t.Helper()

	err := biz.CreatePluginConfig(ctx, model.PluginConfig{
		Name: gjsonGetName(config),
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        id,
			GatewayID: gatewayID,
			Config:    datatypes.JSON([]byte(config)),
			Status:    constant.ResourceStatusSuccess,
		},
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func gjsonGetName(config string) string {
	var info struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal([]byte(config), &info)
	return info.Name
}
