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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func TestLegacyRouteSaveNormalizesStoredConfig(t *testing.T) {
	util.InitEmbedDb()

	gateway := createMCPHelperTestGateway(t, "legacy-route-save-normalize")
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

	legacyRoute := model.Route{
		Name:           "route-column-name",
		ServiceID:      "svc-column",
		UpstreamID:     "ups-column",
		PluginConfigID: "pc-column",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "route-legacy-normalize",
			GatewayID: gateway.ID,
			Config: datatypes.JSON(
				`{"id":"route-legacy-normalize","name":"route-column-name","service_id":"svc-column","upstream_id":"ups-column","plugin_config_id":"pc-column","uris":["/legacy"]}`,
			),
			Status: constant.ResourceStatusSuccess,
		},
	}
	assert.NoError(t, CreateRoute(ctx, legacyRoute))

	routeForRead, err := GetRouteForRead(ctx, legacyRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, "route-column-name", gjson.GetBytes(routeForRead.Config, "name").String())
	assert.Equal(t, "svc-column", gjson.GetBytes(routeForRead.Config, "service_id").String())
	assert.Equal(t, "ups-column", gjson.GetBytes(routeForRead.Config, "upstream_id").String())
	assert.Equal(t, "pc-column", gjson.GetBytes(routeForRead.Config, "plugin_config_id").String())

	draft, err := resourcecodec.PrepareRequestDraft(resourcecodec.RequestInput{
		Source:       resourcecodec.SourceWeb,
		Operation:    constant.OperationTypeUpdate,
		GatewayID:    gateway.ID,
		ResourceType: constant.Route,
		Version:      gateway.GetAPISIXVersionX(),
		PathID:       routeForRead.ID,
		OuterName:    routeForRead.Name,
		OuterFields: map[string]any{
			"service_id":       routeForRead.ServiceID,
			"upstream_id":      routeForRead.UpstreamID,
			"plugin_config_id": routeForRead.PluginConfigID,
		},
		Config: json.RawMessage(routeForRead.Config),
	})
	assert.NoError(t, err)

	storageConfig, err := resourcecodec.BuildStorageConfig(draft)
	assert.NoError(t, err)

	routeForRead.Config = datatypes.JSON(storageConfig)
	routeForRead.Status = constant.ResourceStatusUpdateDraft
	assert.NoError(t, UpdateRoute(ctx, *routeForRead))

	storedRoute, err := GetRoute(ctx, legacyRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, "route-column-name", storedRoute.Name)
	assert.Equal(t, "svc-column", storedRoute.ServiceID)
	assert.Equal(t, "ups-column", storedRoute.UpstreamID)
	assert.Equal(t, "pc-column", storedRoute.PluginConfigID)
	assert.False(t, gjson.GetBytes(storedRoute.Config, "id").Exists())
	assert.False(t, gjson.GetBytes(storedRoute.Config, "name").Exists())
	assert.False(t, gjson.GetBytes(storedRoute.Config, "service_id").Exists())
	assert.False(t, gjson.GetBytes(storedRoute.Config, "upstream_id").Exists())
	assert.False(t, gjson.GetBytes(storedRoute.Config, "plugin_config_id").Exists())
	assert.Equal(t, []any{"/legacy"}, gjson.GetBytes(storedRoute.Config, "uris").Value())
}
