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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	unifyopbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/unifyop"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

func TestReadonlyBatchSyncAddsMissingAndOverwritesOnlyConfirmedDifferences(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 8877401, ReadOnly: true, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	sharedID := uniqueWebCreateName("shared-sync-id")
	draftID := uniqueWebCreateName("draft-sync")
	staleID := uniqueWebCreateName("stale-sync")
	equalID := uniqueWebCreateName("equal-sync")
	missingID := uniqueWebCreateName("missing-sync")
	for _, tc := range []struct {
		kind    constant.APISIXResource
		id      string
		state   constant.ResourceStatus
		missing bool
	}{
		{constant.Route, sharedID, constant.ResourceStatusSuccess, false},
		{constant.Service, sharedID, constant.ResourceStatusSuccess, false},
		{constant.Route, draftID, constant.ResourceStatusUpdateDraft, false},
		{constant.Route, staleID, constant.ResourceStatusSuccess, false},
		{constant.Route, equalID, constant.ResourceStatusSuccess, false},
		{constant.Route, missingID, constant.ResourceStatusSuccess, true},
	} {
		config := datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/snapshot","desc":"snapshot"}`, tc.id))
		if !tc.missing {
			before := datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/old","desc":"old"}`, tc.id))
			if tc.id == equalID {
				before = config
			}
			typed, err := resourcebiz.NewTypedResourceModel(tc.kind, model.ResourceCommonModel{
				ID: tc.id, GatewayID: gateway.ID, Config: before, Status: tc.state,
			})
			require.NoError(t, err)
			require.NoError(t, database.Client().WithContext(ctx).Create(typed).Error)
		}
		require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
			ID: tc.id, GatewayID: gateway.ID, Type: tc.kind, ModRevision: 1, Config: config,
		}))
	}
	c, w := newWebCreateTestContext(t, "", gateway, "bulk-sync-tester")
	ReadonlyEditorSyncPreview(c)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var preview struct {
		Data dto.ReadonlyEditorSyncPreview `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &preview))
	require.Equal(t, 1, preview.Data.Missing)
	require.Len(t, preview.Data.Resources, 4)
	for _, item := range preview.Data.Resources {
		require.Len(t, item.SyncToken, 64)
	}
	// A change after preview must fail independently; a draft still needs explicit consent.
	_, err := repo.GatewaySyncData.WithContext(ctx).Where(
		repo.GatewaySyncData.GatewayID.Eq(gateway.ID), repo.GatewaySyncData.ID.Eq(staleID),
	).Update(repo.GatewaySyncData.ModRevision, 2)
	require.NoError(t, err)
	// New differences appearing after preview are not silently added to the overwrite set.
	lateID := uniqueWebCreateName("late-sync")
	late := &model.Route{Name: lateID, ResourceCommonModel: model.ResourceCommonModel{
		ID: lateID, GatewayID: gateway.ID, Status: constant.ResourceStatusSuccess,
		Config: datatypes.JSON(`{"uri":"/keep"}`),
	}}
	require.NoError(t, repo.Route.WithContext(ctx).Create(late))
	require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID: lateID, GatewayID: gateway.ID, Type: constant.Route,
		Config: datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/late"}`, lateID)),
	}))
	c, w = newWebCreateTestContext(t, mustJSONBody(t, map[string]any{
		"resources": preview.Data.Resources, "overwrite_drafts": false,
	}), gateway, "bulk-sync-tester")
	ReadonlyEditorSync(c)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var result struct {
		Data dto.ReadonlyEditorSyncResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Equal(t, 1, result.Data.Added[constant.Route])
	require.Equal(t, 2, result.Data.Synced)
	require.Len(t, result.Data.Failed, 2)
	failures := map[string]string{}
	for _, item := range result.Data.Failed {
		failures[item.ID] = item.Error
	}
	require.Contains(t, failures[draftID], "草稿")
	require.Contains(t, failures[staleID], "资源已变化")
	for _, kind := range []constant.APISIXResource{constant.Route, constant.Service} {
		saved, err := resourcebiz.GetResourceByID(ctx, kind, sharedID)
		require.NoError(t, err)
		require.Equal(t, "snapshot", gjson.GetBytes(saved.Config, "desc").String())
		require.Equal(t, constant.ResourceStatusSuccess, saved.Status)
	}
	untouched, err := resourcebiz.GetResourceByID(ctx, constant.Route, lateID)
	require.NoError(t, err)
	require.JSONEq(t, string(late.Config), string(untouched.Config))
	fresh, err := unifyopbiz.PreviewReadonlyEditorSync(ctx)
	require.NoError(t, err)
	require.Zero(t, fresh.Missing)
	require.Len(t, fresh.Resources, 3)
	result.Data = dto.ReadonlyEditorSyncResult{}
	c, w = newWebCreateTestContext(t, mustJSONBody(t, map[string]any{
		"resources": fresh.Resources, "overwrite_drafts": true,
	}), gateway, "bulk-sync-tester")
	ReadonlyEditorSync(c)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Equal(t, 3, result.Data.Synced)
	require.Empty(t, result.Data.Failed)
	fresh, err = unifyopbiz.PreviewReadonlyEditorSync(ctx)
	require.NoError(t, err)
	require.Empty(t, fresh.Resources)
}

func TestReadonlyBatchSyncRejectsWritableAndInvalidRequestsBeforeAdding(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 8877402, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	id := uniqueWebCreateName("missing-guard")
	require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID: id, GatewayID: gateway.ID, Type: constant.Route,
		Config: datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/snapshot"}`, id)),
	}))
	item := dto.EditorSyncItem{ResourceType: constant.Route, ID: id, SyncToken: strings.Repeat("0", 64)}
	c, w := newWebCreateTestContext(t, "", gateway, "bulk-sync-tester")
	ReadonlyEditorSyncPreview(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
	c, w = newWebCreateTestContext(
		t,
		mustJSONBody(t, map[string]any{"resources": []dto.EditorSyncItem{item}}),
		gateway,
		"bulk-sync-tester",
	)
	ReadonlyEditorSync(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "仅只读网关")
	gateway.ReadOnly = true
	for _, body := range []map[string]any{
		{},
		{"resources": []dto.EditorSyncItem{item, item}},
		{"resources": []dto.EditorSyncItem{{ResourceType: constant.Route, ID: id, SyncToken: "invalid"}}},
		{"resources": []dto.EditorSyncItem{{ResourceType: "gateway", ID: id, SyncToken: item.SyncToken}}},
	} {
		c, w = newWebCreateTestContext(t, mustJSONBody(t, body), gateway, "bulk-sync-tester")
		ReadonlyEditorSync(c)
		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	}
	count, err := repo.Route.WithContext(ctx).Where(repo.Route.GatewayID.Eq(gateway.ID)).Count()
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestReadonlyBatchSyncRenamedResourceReleasesName(t *testing.T) {
	initWebCreateHandlerTestEnv()
	for index, tc := range []struct {
		name              string
		withDependency    bool
		renamedDependency bool
		blocked           string
	}{
		{name: "rename"},
		{name: "rename_with_missing_dependency", withDependency: true},
		{name: "rename_with_blocked_dependency", withDependency: true, renamedDependency: true},
		{name: "stale_rename", withDependency: true, blocked: "资源已变化"},
		{name: "unconfirmed_draft", withDependency: true, blocked: "草稿"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withDependency := tc.withDependency
			gateway := &model.Gateway{ID: 912340 + index, ReadOnly: true, APISIXVersion: "3.13.0"}
			ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
			a, b, upstream := uniqueWebCreateName("a"), uniqueWebCreateName("b"), uniqueWebCreateName("up")
			foo, bar := uniqueWebCreateName("foo"), uniqueWebCreateName("bar")
			editor := &model.Route{Name: foo, ResourceCommonModel: model.ResourceCommonModel{
				ID:        a,
				GatewayID: gateway.ID,
				Status:    constant.ResourceStatusSuccess,
				Config:    datatypes.JSON(`{"uri":"/old"}`),
			}}
			require.NoError(t, repo.Route.WithContext(ctx).Create(editor))
			configA := fmt.Sprintf(`{"name":%q,"uri":"/new"}`, bar)
			if withDependency {
				configA = fmt.Sprintf(`{"name":%q,"uri":"/new","upstream_id":%q}`, bar, upstream)
				require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
					ID:        upstream,
					GatewayID: gateway.ID,
					Type:      constant.Upstream,
					Config: datatypes.JSON(
						fmt.Sprintf(
							`{"name":%q,"type":"roundrobin","nodes":{"127.0.0.1:80":1}}`,
							upstream,
						),
					),
				}))
			}
			if tc.renamedDependency {
				// The missing upstream's name must also be released by a confirmed overwrite.
				oldID := uniqueWebCreateName("old-upstream")
				require.NoError(t, repo.Upstream.WithContext(ctx).Create(&model.Upstream{
					Name: upstream, ResourceCommonModel: model.ResourceCommonModel{
						ID:        oldID,
						GatewayID: gateway.ID,
						Status:    constant.ResourceStatusSuccess,
						Config: datatypes.JSON(
							`{"type":"roundrobin","nodes":{"127.0.0.1:80":1}}`,
						),
					},
				}))
				require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
					ID:        oldID,
					GatewayID: gateway.ID,
					Type:      constant.Upstream,
					Config: datatypes.JSON(
						fmt.Sprintf(
							`{"name":%q,"type":"roundrobin","nodes":{"127.0.0.1:80":1}}`,
							oldID,
						),
					),
				}))
			}
			for _, snapshot := range []*model.GatewaySyncData{
				{ID: a, GatewayID: gateway.ID, Type: constant.Route, Config: datatypes.JSON(configA)},
				{ID: b, GatewayID: gateway.ID, Type: constant.Route, Config: datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/b"}`, foo))},
			} {
				require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(snapshot))
			}
			if tc.blocked == "草稿" {
				_, err := repo.Route.WithContext(
					ctx,
				).Where(
					repo.Route.ID.Eq(a),
				).Update(
					repo.Route.Status,
					constant.ResourceStatusUpdateDraft,
				)
				require.NoError(t, err)
			}
			preview, err := unifyopbiz.PreviewReadonlyEditorSync(ctx)
			require.NoError(t, err)
			wantMissing := 1
			if withDependency {
				wantMissing++
			}
			require.Equal(t, wantMissing, preview.Missing)
			wantSynced := 1
			if tc.renamedDependency {
				wantSynced++
			}
			require.Len(t, preview.Resources, wantSynced)
			if tc.blocked == "资源已变化" {
				_, err := repo.GatewaySyncData.WithContext(
					ctx,
				).Where(
					repo.GatewaySyncData.GatewayID.Eq(gateway.ID),
					repo.GatewaySyncData.ID.Eq(a),
				).Update(
					repo.GatewaySyncData.ModRevision,
					2,
				)
				require.NoError(t, err)
			}
			c, w := newWebCreateTestContext(
				t,
				mustJSONBody(t, map[string]any{"resources": preview.Resources}),
				gateway,
				"sync-tester",
			)
			ReadonlyEditorSync(c)
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			var response struct {
				Data dto.ReadonlyEditorSyncResult `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			result := response.Data
			saved, err := repo.Route.WithContext(ctx).Where(repo.Route.ID.Eq(a)).Take()
			require.NoError(t, err)
			if withDependency {
				require.Equal(t, 1, result.Added[constant.Upstream])
			}
			if tc.blocked != "" {
				require.Equal(t, foo, saved.Name)
				require.Zero(t, result.Synced)
				require.Zero(t, result.Added[constant.Route])
				require.Len(t, result.Failed, 2)
				failures := map[string]string{}
				for _, item := range result.Failed {
					failures[item.ID] = item.Error
				}
				require.Contains(t, failures[a], tc.blocked)
				require.Contains(t, failures[b], unifyopbiz.ErrResourceNameConflict.Error())
				count, err := repo.Route.WithContext(ctx).Where(repo.Route.ID.Eq(b)).Count()
				require.NoError(t, err)
				require.Zero(t, count)
				return
			}
			require.Equal(t, wantSynced, result.Synced)
			require.Equal(t, 1, result.Added[constant.Route])
			require.Empty(t, result.Failed)
			require.Equal(t, bar, saved.Name)
			if withDependency {
				require.Equal(t, upstream, saved.UpstreamID)
			}
			saved, err = repo.Route.WithContext(ctx).Where(repo.Route.ID.Eq(b)).Take()
			require.NoError(t, err)
			require.Equal(t, foo, saved.Name)
			fresh, err := unifyopbiz.PreviewReadonlyEditorSync(ctx)
			require.NoError(t, err)
			require.Zero(t, fresh.Missing)
			require.Empty(t, fresh.Resources)
		})
	}
}

func TestReadonlyBatchSyncQueryErrorsReturnInternalServerError(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 912349, ReadOnly: true, APISIXVersion: "3.13.0"}
	for _, operation := range []string{"preview", "batch", "existing-add"} {
		t.Run(operation, func(t *testing.T) {
			c, w := newWebCreateTestContext(t, `{"resources":[]}`, gateway, "sync-tester")
			ctx, cancel := context.WithCancel(c.Request.Context())
			c.Request = c.Request.WithContext(ctx)
			cancel()
			switch operation {
			case "preview":
				ReadonlyEditorSyncPreview(c)
			case "batch":
				ReadonlyEditorSync(c)
			case "existing-add":
				SyncedResourceManaged(c)
			}
			want := http.StatusInternalServerError
			require.Equal(t, want, w.Code, w.Body.String())
		})
	}
}

func TestReadonlyBatchSyncUnresolvedNameConflictReturnsConflict(t *testing.T) {
	initWebCreateHandlerTestEnv()
	gateway := &model.Gateway{ID: 912350, ReadOnly: true, APISIXVersion: "3.13.0"}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	name := uniqueWebCreateName("duplicate")
	require.NoError(
		t,
		repo.Route.WithContext(
			ctx,
		).Create(
			&model.Route{Name: name, ResourceCommonModel: model.ResourceCommonModel{
				ID: uniqueWebCreateName(
					"existing",
				),
				GatewayID: gateway.ID,
				Status:    constant.ResourceStatusSuccess,
				Config:    datatypes.JSON(`{"uri":"/existing"}`),
			}},
		),
	)
	require.NoError(t, repo.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID: uniqueWebCreateName(
			"missing",
		),
		GatewayID: gateway.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/missing"}`, name)),
	}))
	c, w := newWebCreateTestContext(t, `{"resources":[]}`, gateway, "sync-tester")
	ReadonlyEditorSync(c)
	require.Equal(t, http.StatusConflict, w.Code, w.Body.String())
	count, err := repo.Route.WithContext(ctx).Where(repo.Route.GatewayID.Eq(gateway.ID)).Count()
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
}
