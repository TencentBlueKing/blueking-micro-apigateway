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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	diffbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/diff"
	syncdatabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/syncdata"
	unifyopbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/unifyop"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

func TestSyncToEditorSameConfigDraftRestoresPublishedState(t *testing.T) {
	initWebCreateHandlerTestEnv()
	for _, state := range []constant.ResourceStatus{
		constant.ResourceStatusCreateDraft, constant.ResourceStatusUpdateDraft,
	} {
		t.Run(string(state), func(t *testing.T) {
			gateway := &model.Gateway{ID: 9876547, APISIXVersion: "3.13.0"}
			ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
			id := uniqueWebCreateName("same-config-draft")
			route := &model.Route{Name: id, ResourceCommonModel: model.ResourceCommonModel{
				ID: id, GatewayID: gateway.ID, Status: state, Config: datatypes.JSON(`{"uri":"/demo"}`),
			}}
			require.NoError(t, repo.Route.WithContext(ctx).Create(route))
			snapshot := &model.GatewaySyncData{
				ID: id, GatewayID: gateway.ID, Type: constant.Route, Config: route.Config,
			}
			require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(snapshot))
			preview, err := unifyopbiz.GetResourceConfigDiffDetail(ctx, constant.Route, id)
			require.NoError(t, err)
			pending, err := diffbiz.DiffResources(ctx, constant.Route, []string{id}, "", nil, false)
			require.NoError(t, err)
			require.Len(t, pending, 1)
			require.Equal(t, 1, pending[0].AddedCount+pending[0].UpdateCount)
			call := func(token string, confirm bool) int {
				c, w := newWebCreateTestContext(t, mustJSONBody(t, map[string]any{
					"sync_token": token, "overwrite_draft": confirm,
				}), gateway, "sync-tester")
				c.Params = gin.Params{
					{Key: "gateway_id", Value: fmt.Sprint(gateway.ID)},
					{Key: "type", Value: "route"},
					{Key: "id", Value: id},
				}
				ResourceSyncToEditor(c)
				return w.Code
			}
			require.Equal(t, http.StatusConflict, call(preview.SyncToken, false))
			require.Equal(t, http.StatusOK, call(preview.SyncToken, true))
			saved, err := repo.Route.WithContext(ctx).Where(repo.Route.ID.Eq(id)).Take()
			require.NoError(t, err)
			require.Equal(t, constant.ResourceStatusSuccess, saved.Status)
			require.JSONEq(t, string(route.Config), string(saved.Config))
			require.Equal(t, "sync-tester", saved.Updater)
			pending, err = diffbiz.DiffResources(ctx, constant.Route, []string{id}, "", nil, false)
			require.NoError(t, err)
			require.Empty(t, pending)
			require.Equal(t, http.StatusConflict, call(preview.SyncToken, true))
			preview, err = unifyopbiz.GetResourceConfigDiffDetail(ctx, constant.Route, id)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, call(preview.SyncToken, true))
			logs, err := repo.OperationAuditLog.WithContext(ctx).
				Where(repo.OperationAuditLog.ResourceIDs.Eq(id)).Find()
			require.NoError(t, err)
			require.Len(t, logs, 2)
		})
	}
}

func TestSyncToEditorGuardsAndPersistence(t *testing.T) {
	initWebCreateHandlerTestEnv()
	for _, readOnly := range []bool{false, true} {
		for _, state := range []constant.ResourceStatus{
			constant.ResourceStatusSuccess,
			constant.ResourceStatusUpdateDraft, constant.ResourceStatusDeleteDraft, constant.ResourceStatusCreateDraft,
		} {
			t.Run(fmt.Sprintf("readonly=%t/%s", readOnly, state), func(t *testing.T) {
				gateway := &model.Gateway{ID: 9876544, ReadOnly: readOnly, APISIXVersion: "3.13.0"}
				ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
				id := uniqueWebCreateName("sync-route")
				route := &model.Route{
					Name:           id,
					ServiceID:      "old-service",
					UpstreamID:     "old-upstream",
					PluginConfigID: "old-config",
					ResourceCommonModel: model.ResourceCommonModel{
						ID: id, GatewayID: gateway.ID, Status: state,
						Config: datatypes.JSON(`{"uri":"/old","plugins":{"key-auth":{}}}`),
					},
				}
				require.NoError(t, repo.Route.WithContext(ctx).Create(route))
				snapshot := &model.GatewaySyncData{
					ID:          id,
					GatewayID:   gateway.ID,
					Type:        constant.Route,
					ModRevision: 2,
					Config: datatypes.JSON(
						fmt.Sprintf(
							`{"name":%q,"uri":"/new","plugins":{"ai-proxy-multi":{"instances":[{"auth":{}}]},"unknown-plugin":{},"key-auth":{"_meta":{"disable":true}}}}`,
							id,
						),
					),
				}
				require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(snapshot))
				preview, err := unifyopbiz.GetResourceConfigDiffDetail(ctx, constant.Route, id)
				require.NoError(t, err)
				require.NotEmpty(t, preview.SyncToken)
				call := func(token string, confirm bool) int {
					c, w := newWebCreateTestContext(
						t,
						mustJSONBody(
							t,
							map[string]any{"sync_token": token, "overwrite_draft": confirm},
						),
						gateway,
						"sync-tester",
					)
					c.Params = gin.Params{
						{Key: "gateway_id", Value: fmt.Sprint(gateway.ID)},
						{Key: "type", Value: "route"},
						{Key: "id", Value: id},
					}
					ResourceSyncToEditor(c)
					return w.Code
				}
				require.Equal(t, http.StatusConflict, call(fmt.Sprintf("%064d", 0), true))
				if state != constant.ResourceStatusSuccess {
					require.Equal(t, http.StatusConflict, call(preview.SyncToken, false))
				}
				unchanged, err := repo.Route.WithContext(ctx).Where(repo.Route.ID.Eq(id)).Take()
				require.NoError(t, err)
				require.Equal(t, state, unchanged.Status)
				require.Equal(t, "old-service", unchanged.ServiceID)
				require.Equal(t, http.StatusOK, call(preview.SyncToken, true))
				saved, err := repo.Route.WithContext(ctx).Where(repo.Route.ID.Eq(id)).Take()
				require.NoError(t, err)
				require.Equal(t, route.AutoID, saved.AutoID)
				require.Equal(t, constant.ResourceStatusSuccess, saved.Status)
				require.Equal(t, "sync-tester", saved.Updater)
				require.Empty(t, saved.ServiceID)
				require.Empty(t, saved.UpstreamID)
				require.Empty(t, saved.PluginConfigID)
				require.True(t, gjson.GetBytes(saved.Config, "plugins.unknown-plugin").Exists())
				require.True(
					t,
					gjson.GetBytes(
						saved.Config,
						"plugins.ai-proxy-multi.instances.0.auth",
					).Exists(),
				)
				require.Equal(t, "/new", gjson.GetBytes(saved.Config, "uri").String())
				result, err := syncdatabiz.ConfigStatus(
					snapshot,
					&saved.ResourceCommonModel,
					gateway.GetAPISIXVersionX(),
				)
				require.NoError(t, err)
				require.Equal(t, constant.SyncedResourceStatusSuccess, result)
				logs, err := repo.OperationAuditLog.WithContext(
					ctx,
				).Where(
					repo.OperationAuditLog.ResourceIDs.Eq(id),
				).Find()
				require.NoError(t, err)
				require.Len(t, logs, 2)
				require.Equal(t, "sync-tester", logs[1].Operator)
				// Old confirmations cannot overwrite changes made after the preview.
				require.Equal(t, http.StatusConflict, call(preview.SyncToken, true))
			})
		}
	}
}

func TestSyncedResourceManagedAllowsReadOnlyAddMissing(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 9876545, ReadOnly: true, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	id := uniqueWebCreateName("readonly-route")
	snapshot := &model.GatewaySyncData{
		ID: id, GatewayID: gateway.ID, Type: constant.Route,
		Config: datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/snapshot"}`, id)),
	}
	require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(snapshot))
	for range 2 {
		c, w := newWebCreateTestContext(t, mustJSONBody(t, map[string]any{
			"resource_type": "route", "resource_id_list": []string{id},
		}), gateway, "sync-tester")
		SyncedResourceManaged(c)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	}
	saved, err := repo.Route.WithContext(
		ctx,
	).Where(
		repo.Route.GatewayID.Eq(gateway.ID),
		repo.Route.ID.Eq(id),
	).Take()
	require.NoError(t, err)
	require.Equal(t, constant.ResourceStatusSuccess, saved.Status)
	require.Equal(t, "/snapshot", gjson.GetBytes(saved.Config, "uri").String())
}

func TestSyncToEditorMissingDependencyReturnsConflict(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 9876546, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	id := uniqueWebCreateName("missing-dependency-route")
	route := &model.Route{Name: id, ResourceCommonModel: model.ResourceCommonModel{
		ID: id, GatewayID: gateway.ID, Status: constant.ResourceStatusUpdateDraft,
		Config: datatypes.JSON(`{"uri":"/draft"}`),
	}}
	require.NoError(t, repo.Route.WithContext(ctx).Create(route))
	snapshot := &model.GatewaySyncData{
		ID: id, GatewayID: gateway.ID, Type: constant.Route,
		Config: datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/snapshot","upstream_id":"snapshot-only"}`, id)),
	}
	require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(snapshot))
	preview, err := unifyopbiz.GetResourceConfigDiffDetail(ctx, constant.Route, id)
	require.NoError(t, err)
	c, w := newWebCreateTestContext(t, mustJSONBody(t, map[string]any{
		"sync_token": preview.SyncToken, "overwrite_draft": true,
	}), gateway, "sync-tester")
	c.Params = gin.Params{
		{Key: "gateway_id", Value: fmt.Sprint(gateway.ID)},
		{Key: "type", Value: "route"},
		{Key: "id", Value: id},
	}
	ResourceSyncToEditor(c)
	require.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "snapshot-only")
	saved, err := repo.Route.WithContext(
		ctx,
	).Where(
		repo.Route.GatewayID.Eq(gateway.ID),
		repo.Route.ID.Eq(id),
	).Take()
	require.NoError(t, err)
	require.Equal(t, route.Config, saved.Config)
	require.Equal(t, constant.ResourceStatusUpdateDraft, saved.Status)
}
