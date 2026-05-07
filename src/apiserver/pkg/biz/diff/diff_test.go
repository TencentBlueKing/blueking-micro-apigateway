/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
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

package diff

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func init() {
	if err := cryptography.Init("jxi18GX5w2qgHwfZCFpn07q8FScXJOd3", "k2dbCGetyusW"); err != nil {
		panic(err)
	}
	util.InitEmbedDb()
}

func mustSetJSONField(t *testing.T, config datatypes.JSON, path string, value any) datatypes.JSON {
	t.Helper()

	updated, err := sjson.SetBytes(config, path, value)
	assert.NoError(t, err)
	return datatypes.JSON(updated)
}

func diffResultByType(
	results []dto.ResourceChangeInfo,
	resourceType constant.APISIXResource,
) (dto.ResourceChangeInfo, bool) {
	for _, item := range results {
		if item.ResourceType == resourceType {
			return item, true
		}
	}
	return dto.ResourceChangeInfo{}, false
}

func TestDiffResources_BuildsChangeSummaryForDraftStatuses(t *testing.T) {
	gateway, ctx := newDiffGatewayContext(t, "3.11.0")

	route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
	route.Name = "diff-route-added"
	route.Config = mustSetJSONField(t, route.Config, "name", route.Name)
	assert.NoError(t, resourcebiz.CreateRoute(ctx, *route))

	service := data.Service1WithNoRelation(gateway, constant.ResourceStatusUpdateDraft)
	service.Name = "diff-service-updated"
	service.Config = mustSetJSONField(t, service.Config, "name", service.Name)
	assert.NoError(t, resourcebiz.CreateService(ctx, *service))

	upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusDeleteDraft)
	upstream.Name = "diff-upstream-deleted"
	upstream.Config = mustSetJSONField(t, upstream.Config, "name", upstream.Name)
	assert.NoError(t, resourcebiz.CreateUpstream(ctx, *upstream))

	results, err := DiffResources(ctx, "", nil, "", nil, true)
	assert.NoError(t, err)

	routeResult, ok := diffResultByType(results, constant.Route)
	assert.True(t, ok)
	assert.Equal(t, 1, routeResult.AddedCount)
	assert.Equal(t, 0, routeResult.DeletedCount)
	assert.Equal(t, 0, routeResult.UpdateCount)
	if assert.Len(t, routeResult.ChangeDetail, 1) {
		assert.Equal(t, route.ID, routeResult.ChangeDetail[0].ResourceID)
		assert.Equal(t, route.Name, routeResult.ChangeDetail[0].Name)
		assert.Equal(t, constant.ResourceStatusCreateDraft, routeResult.ChangeDetail[0].BeforeStatus)
		assert.Equal(t, constant.ResourceStatusSuccess, routeResult.ChangeDetail[0].AfterStatus)
		assert.Equal(t, constant.OperationTypeCreate, routeResult.ChangeDetail[0].PublishFrom)
	}

	serviceResult, ok := diffResultByType(results, constant.Service)
	assert.True(t, ok)
	assert.Equal(t, 0, serviceResult.AddedCount)
	assert.Equal(t, 0, serviceResult.DeletedCount)
	assert.Equal(t, 1, serviceResult.UpdateCount)
	if assert.Len(t, serviceResult.ChangeDetail, 1) {
		assert.Equal(t, service.ID, serviceResult.ChangeDetail[0].ResourceID)
		assert.Equal(t, service.Name, serviceResult.ChangeDetail[0].Name)
		assert.Equal(t, constant.ResourceStatusUpdateDraft, serviceResult.ChangeDetail[0].BeforeStatus)
		assert.Equal(t, constant.ResourceStatusSuccess, serviceResult.ChangeDetail[0].AfterStatus)
		assert.Equal(t, constant.OperationTypeUpdate, serviceResult.ChangeDetail[0].PublishFrom)
	}

	upstreamResult, ok := diffResultByType(results, constant.Upstream)
	assert.True(t, ok)
	assert.Equal(t, 0, upstreamResult.AddedCount)
	assert.Equal(t, 1, upstreamResult.DeletedCount)
	assert.Equal(t, 0, upstreamResult.UpdateCount)
	if assert.Len(t, upstreamResult.ChangeDetail, 1) {
		assert.Equal(t, upstream.ID, upstreamResult.ChangeDetail[0].ResourceID)
		assert.Equal(t, upstream.Name, upstreamResult.ChangeDetail[0].Name)
		assert.Equal(t, constant.ResourceStatusDeleteDraft, upstreamResult.ChangeDetail[0].BeforeStatus)
		assert.Equal(t, constant.ResourceStatusSuccess, upstreamResult.ChangeDetail[0].AfterStatus)
		assert.Equal(t, constant.OperationTypeDelete, upstreamResult.ChangeDetail[0].PublishFrom)
	}
}

func TestDiffResources_FiltersRequestedTypeByNameAndStatus(t *testing.T) {
	gateway, ctx := newDiffGatewayContext(t, "3.11.0")

	matchedRoute := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
	matchedRoute.Name = "diff-route-match"
	matchedRoute.Config = mustSetJSONField(t, matchedRoute.Config, "name", matchedRoute.Name)
	assert.NoError(t, resourcebiz.CreateRoute(ctx, *matchedRoute))

	unmatchedRoute := data.Route2WithNoRelationResource(gateway, constant.ResourceStatusUpdateDraft)
	unmatchedRoute.Name = "diff-route-other"
	unmatchedRoute.Config = mustSetJSONField(t, unmatchedRoute.Config, "name", unmatchedRoute.Name)
	assert.NoError(t, resourcebiz.CreateRoute(ctx, *unmatchedRoute))

	results, err := DiffResources(
		ctx,
		constant.Route,
		[]string{matchedRoute.ID, unmatchedRoute.ID},
		"match",
		[]constant.ResourceStatus{constant.ResourceStatusCreateDraft},
		false,
	)
	assert.NoError(t, err)
	if assert.Len(t, results, 1) {
		assert.Equal(t, constant.Route, results[0].ResourceType)
		assert.Equal(t, 1, results[0].AddedCount)
		assert.Equal(t, 0, results[0].UpdateCount)
		assert.Equal(t, 0, results[0].DeletedCount)
		if assert.Len(t, results[0].ChangeDetail, 1) {
			assert.Equal(t, matchedRoute.ID, results[0].ChangeDetail[0].ResourceID)
			assert.Equal(t, matchedRoute.Name, results[0].ChangeDetail[0].Name)
		}
	}
}

func TestDiffResources_ExpandsRelatedResourcesForDiffAll(t *testing.T) {
	gateway, ctx := newDiffGatewayContext(t, "3.11.0")

	service := data.Service1WithNoRelation(gateway, constant.ResourceStatusUpdateDraft)
	service.Name = "diff-related-service"
	service.Config = mustSetJSONField(t, service.Config, "name", service.Name)
	assert.NoError(t, resourcebiz.CreateService(ctx, *service))

	upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusDeleteDraft)
	upstream.Name = "diff-related-upstream"
	upstream.Config = mustSetJSONField(t, upstream.Config, "name", upstream.Name)
	assert.NoError(t, resourcebiz.CreateUpstream(ctx, *upstream))

	pluginConfig := data.PluginConfig1WithNoRelation(gateway, constant.ResourceStatusUpdateDraft)
	pluginConfig.Name = "diff-related-plugin-config"
	pluginConfig.Config = mustSetJSONField(t, pluginConfig.Config, "name", pluginConfig.Name)
	assert.NoError(t, resourcebiz.CreatePluginConfig(ctx, *pluginConfig))

	route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
	route.Name = "diff-root-route"
	route.Config = mustSetJSONField(t, route.Config, "name", route.Name)
	route.Config = mustSetJSONField(t, route.Config, "service_id", service.ID)
	route.Config = mustSetJSONField(t, route.Config, "upstream_id", upstream.ID)
	route.Config = mustSetJSONField(t, route.Config, "plugin_config_id", pluginConfig.ID)
	assert.NoError(t, resourcebiz.CreateRoute(ctx, *route))

	results, err := DiffResources(
		ctx,
		constant.Route,
		[]string{route.ID},
		"",
		nil,
		true,
	)
	assert.NoError(t, err)

	routeResult, ok := diffResultByType(results, constant.Route)
	assert.True(t, ok)
	assert.Equal(t, 1, routeResult.AddedCount)

	serviceResult, ok := diffResultByType(results, constant.Service)
	assert.True(t, ok)
	assert.Equal(t, 1, serviceResult.UpdateCount)
	if assert.Len(t, serviceResult.ChangeDetail, 1) {
		assert.Equal(t, service.ID, serviceResult.ChangeDetail[0].ResourceID)
	}

	upstreamResult, ok := diffResultByType(results, constant.Upstream)
	assert.True(t, ok)
	assert.Equal(t, 1, upstreamResult.DeletedCount)
	if assert.Len(t, upstreamResult.ChangeDetail, 1) {
		assert.Equal(t, upstream.ID, upstreamResult.ChangeDetail[0].ResourceID)
	}

	pluginConfigResult, ok := diffResultByType(results, constant.PluginConfig)
	assert.True(t, ok)
	assert.Equal(t, 1, pluginConfigResult.UpdateCount)
	if assert.Len(t, pluginConfigResult.ChangeDetail, 1) {
		assert.Equal(t, pluginConfig.ID, pluginConfigResult.ChangeDetail[0].ResourceID)
	}
}

func newDiffGatewayContext(t *testing.T, version string) (*model.Gateway, context.Context) {
	t.Helper()

	gateway := data.Gateway1WithBkAPISIX()
	gateway.Name = diffTestName(t, strings.ReplaceAll(version, ".", "-"))
	gateway.Desc = gateway.Name
	gateway.APISIXVersion = version
	gateway.EtcdConfig.Prefix = "/" + diffTestName(t, "etcd")
	if err := repo.Gateway.WithContext(context.Background()).Create(gateway); err != nil {
		t.Fatal(err)
	}
	return gateway, ginx.SetGatewayInfoToContext(context.Background(), gateway)
}

func diffTestName(t *testing.T, suffix string) string {
	t.Helper()
	name := strings.ToLower(t.Name() + "-" + suffix)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	if len(name) > 48 {
		name = name[:24] + "-" + name[len(name)-23:]
	}
	return name
}
