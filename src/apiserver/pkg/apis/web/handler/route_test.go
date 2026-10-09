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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

func TestRouteListPluginConfigFilter(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 2201, APISIXVersion: "3.13.0"}
	otherGateway := &model.Gateway{ID: 2202, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	otherCtx := ginx.SetGatewayInfoToContext(context.Background(), otherGateway)

	fixtures := []struct {
		id             string
		pluginConfigID string
		serviceID      string
		gatewayID      int
	}{
		{"route-filter-a", "plugin-config-1", "service-1", gateway.ID},
		{"route-filter-b", "plugin-config-1", "service-2", gateway.ID},
		{"route-filter-c", "plugin-config-10", "service-1", gateway.ID},
		{"route-filter-d", "", "service-1", gateway.ID},
		{"route-filter-other", "plugin-config-1", "service-1", otherGateway.ID},
	}
	t.Cleanup(func() {
		assert.NoError(t, resourcebiz.BatchDeleteRoutes(ctx, []string{
			"route-filter-a", "route-filter-b", "route-filter-c", "route-filter-d",
		}))
		assert.NoError(t, resourcebiz.BatchDeleteRoutes(otherCtx, []string{"route-filter-other"}))
	})
	for _, fixture := range fixtures {
		require.NoError(t, resourcebiz.CreateRoute(ctx, model.Route{
			Name:           fixture.id,
			PluginConfigID: fixture.pluginConfigID,
			ServiceID:      fixture.serviceID,
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        fixture.id,
				GatewayID: fixture.gatewayID,
				Config:    datatypes.JSON(`{"uris":["/test"]}`),
				Status:    constant.ResourceStatusCreateDraft,
			},
		}))
	}

	router := gin.New()
	router.GET("/api/v1/web/gateways/:gateway_id/routes/", func(c *gin.Context) {
		ginx.SetGatewayInfo(c, gateway)
		RouteList(c)
	})
	tests := []struct {
		name      string
		query     string
		wantIDs   []string
		wantCount int64
	}{
		{
			"omitted filter",
			"",
			[]string{"route-filter-a", "route-filter-b", "route-filter-c", "route-filter-d"},
			4,
		},
		{
			"empty filter",
			"&plugin_config_id=",
			[]string{"route-filter-a", "route-filter-b", "route-filter-c", "route-filter-d"},
			4,
		},
		{
			"exact match and gateway isolation",
			"&plugin_config_id=plugin-config-1",
			[]string{"route-filter-a", "route-filter-b"},
			2,
		},
		{"distinct ID", "&plugin_config_id=plugin-config-10", []string{"route-filter-c"}, 1},
		{"unknown ID", "&plugin_config_id=missing", nil, 0},
		{
			"combined service filter",
			"&plugin_config_id=plugin-config-1&service_id=service-1",
			[]string{"route-filter-a"},
			1,
		},
		{"conflicting route ID", "&plugin_config_id=plugin-config-1&id=route-filter-c", nil, 0},
		{
			"combined route ID",
			"&plugin_config_id=plugin-config-1&id=route-filter-b",
			[]string{"route-filter-b"},
			1,
		},
		{
			"pagination keeps filtered count",
			"&plugin_config_id=plugin-config-1&offset=1",
			[]string{"route-filter-b"},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet,
				"/api/v1/web/gateways/2201/routes/?order_by=name:asc"+tt.query, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Equal(t, tt.wantCount, gjson.Get(recorder.Body.String(), "data.count").Int())
			var ids []string
			for _, result := range gjson.Get(recorder.Body.String(), "data.results").Array() {
				ids = append(ids, result.Get("id").String())
			}
			assert.Equal(t, tt.wantIDs, ids)
		})
	}
}
