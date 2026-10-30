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

package handler

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

func TestSyncedItemDetectsPluginDriftAndScopesResourceType(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 9876543, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	id := uniqueWebCreateName("infer-dep")
	route := &model.Route{Name: id, ResourceCommonModel: model.ResourceCommonModel{
		ID: id, GatewayID: gateway.ID, Status: constant.ResourceStatusSuccess,
		Config: datatypes.JSON(`{"uri":"/infer/*","plugins":{"key-auth":{}}}`),
	}}
	require.NoError(t, repo.Route.WithContext(ctx).Create(route))
	snapshot := &model.GatewaySyncData{
		ID:        id,
		GatewayID: gateway.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"name":"` + id + `","uri":"/infer/*","plugins":{"key-auth":{},"ai-proxy-multi":{"timeout":600000}}}`,
		),
	}
	items, err := enrichOutputInfo(ctx, []*model.GatewaySyncData{snapshot}, "", "")
	require.NoError(t, err)
	require.Equal(t, constant.SyncStatus("different"), items[0].Status)
	require.Equal(t, constant.PublishByOthers, items[0].PublishSource)
	snapshot.Type = constant.Service
	items, err = enrichOutputInfo(ctx, []*model.GatewaySyncData{
		{ID: id, GatewayID: gateway.ID, Type: constant.Route, Config: route.Config}, snapshot,
	}, "", "")
	require.NoError(t, err)
	require.Equal(t, constant.SyncedResourceStatusSuccess, items[0].Status)
	require.Equal(t, constant.SyncedResourceStatusMiss, items[1].Status)
}

func TestSyncedItemSummaryAndFilterUseConfigStatus(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 9876550, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	for _, prefix := range []string{"bk.", "external-"} {
		id := uniqueWebCreateName(prefix)
		route := &model.Route{Name: id, ResourceCommonModel: model.ResourceCommonModel{
			ID: id, GatewayID: gateway.ID,
			Status: constant.ResourceStatusSuccess, Config: datatypes.JSON(`{"uri":"/old"}`),
		}}
		require.NoError(t, repo.Route.WithContext(ctx).Create(route))
		snapshot := &model.GatewaySyncData{
			ID:        id,
			GatewayID: gateway.ID,
			Type:      constant.Route,
			Config:    datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/new"}`, id)),
		}
		require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(snapshot))
	}
	c, w := newWebCreateTestContext(t, ``, gateway, "tester")
	SyncedItemSummary(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int64(2), gjson.Get(w.Body.String(), "data.different").Int())
	require.Equal(t, int64(0), gjson.Get(w.Body.String(), "data.success").Int())
	c, w = newWebCreateTestContext(t, ``, gateway, "tester")
	c.Request.URL.RawQuery = "status=different&limit=10&offset=0"
	SyncedItemList(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int64(2), gjson.Get(w.Body.String(), "data.count").Int())
	require.Len(t, gjson.Get(w.Body.String(), "data.results").Array(), 2)
}
