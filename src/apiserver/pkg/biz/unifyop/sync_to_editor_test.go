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

package unifyop

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	syncdatabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/syncdata"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

func TestSyncToEditorConvergesAfterSaveNormalization(t *testing.T) {
	for _, test := range []struct {
		name   string
		kind   constant.APISIXResource
		config string
	}{
		{"route vars", constant.Route, `{"uri":"/demo","vars":[],"plugins":{"key-auth":{},"custom":{"items":[],"options":{}}}}`},
		{"route labels", constant.Route, `{"uri":"/demo","labels":{}}`},
		{"upstream nodes", constant.Upstream, `{"type":"roundrobin","nodes":[]}`},
		{"inline upstream", constant.Route, `{"uri":"/demo","upstream":{"type":"roundrobin","nodes":{}}}`},
		{"SSL inferred snis", constant.SSL, `{"sni":"test.com"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			id := idx.GenResourceID(test.kind)
			config, err := sjson.Set(test.config, "name", id)
			require.NoError(t, err)
			if test.kind == constant.SSL {
				config, err = sjson.Set(config, "cert", strings.ReplaceAll(testSSLCert, `\n`, "\n"))
				require.NoError(t, err)
				config, err = sjson.Set(config, "key", strings.ReplaceAll(testSSLKey, `\n`, "\n"))
				require.NoError(t, err)
			}
			before, err := sjson.Set(config, "desc", "before")
			require.NoError(t, err)
			typed, err := resourcebiz.NewTypedResourceModel(test.kind, model.ResourceCommonModel{
				ID:        id,
				GatewayID: gatewayInfo.ID,
				Status:    constant.ResourceStatusSuccess,
				Config:    []byte(before),
			})
			require.NoError(t, err)
			require.NoError(t, database.Client().WithContext(gatewayCtx).Create(typed).Error)
			after, err := sjson.Set(config, "desc", "after")
			require.NoError(t, err)
			snapshot := &model.GatewaySyncData{
				ID:          id,
				GatewayID:   gatewayInfo.ID,
				Type:        test.kind,
				Config:      []byte(after),
				ModRevision: 1,
			}
			require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
			for range 2 {
				preview, err := GetResourceConfigDiffDetail(gatewayCtx, test.kind, id)
				require.NoError(t, err)
				require.NoError(
					t,
					SyncResourceToEditor(gatewayCtx, test.kind, id, preview.SyncToken, false),
				)
				saved, err := resourcebiz.GetResourceByID(gatewayCtx, test.kind, id)
				require.NoError(t, err)
				require.Equal(t, "after", gjson.GetBytes(saved.Config, "desc").String())
				state, err := syncdatabiz.ConfigStatus(
					snapshot,
					&saved,
					gatewayInfo.GetAPISIXVersionX(),
				)
				require.NoError(t, err)
				require.Equal(t, constant.SyncedResourceStatusSuccess, state)
				if test.name == "route vars" {
					require.JSONEq(
						t,
						`{"key-auth":{},"custom":{"items":[],"options":{}}}`,
						gjson.GetBytes(saved.Config, "plugins").Raw,
					)
				}
			}
			logs, err := repo.OperationAuditLog.WithContext(gatewayCtx).
				Where(repo.OperationAuditLog.ResourceIDs.Eq(id)).Find()
			require.NoError(t, err)
			require.Len(t, logs, 2)
		})
	}
}

func TestAddSyncedResourcesConvergesAfterSaveNormalization(t *testing.T) {
	id := idx.GenResourceID(constant.Route)
	snapshot := &model.GatewaySyncData{
		ID:        id,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			fmt.Sprintf(`{"name":%q,"uri":"/demo","vars":[],"labels":{},"plugins":{"key-auth":{}}}`, id),
		),
	}
	require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
	_, err := AddSyncedResources(gatewayCtx, []string{id})
	require.NoError(t, err)
	saved, err := resourcebiz.GetResourceByID(gatewayCtx, constant.Route, id)
	require.NoError(t, err)
	state, err := syncdatabiz.ConfigStatus(snapshot, &saved, gatewayInfo.GetAPISIXVersionX())
	require.NoError(t, err)
	require.Equal(t, constant.SyncedResourceStatusSuccess, state)
	require.True(t, gjson.GetBytes(saved.Config, "plugins.key-auth").Exists())
}

func TestAddSyncedResourcesReadOnly(t *testing.T) {
	gateway := *gatewayInfo
	gateway.ReadOnly = true
	ctx := ginx.SetGatewayInfoToContext(gatewayCtx, &gateway)
	existingID := idx.GenResourceID(constant.Route)
	newID := idx.GenResourceID(constant.Route)
	upstreamID := idx.GenResourceID(constant.Upstream)
	existing := &model.Route{Name: existingID, ResourceCommonModel: model.ResourceCommonModel{
		ID: existingID, GatewayID: gateway.ID, Status: constant.ResourceStatusUpdateDraft,
		Config: datatypes.JSON(`{"uri":"/draft"}`),
	}}
	require.NoError(t, repo.Route.WithContext(ctx).Create(existing))
	require.NoError(t, repo.GatewaySyncData.WithContext(ctx).CreateInBatches([]*model.GatewaySyncData{
		{
			ID: existingID, GatewayID: gateway.ID, Type: constant.Route,
			Config: datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/deployed"}`, existingID)),
		},
		{
			ID:        newID,
			GatewayID: gateway.ID,
			Type:      constant.Route,
			Config: datatypes.JSON(
				fmt.Sprintf(`{"name":%q,"uri":"/new","upstream_id":%q}`, newID, upstreamID),
			),
		},
		{
			ID:        upstreamID,
			GatewayID: gateway.ID,
			Type:      constant.Upstream,
			Config: datatypes.JSON(
				fmt.Sprintf(`{"name":%q,"type":"roundrobin","nodes":{"127.0.0.1:80":1}}`, upstreamID),
			),
		},
	}, 3))
	counts, err := AddSyncedResources(ctx, []string{existingID, newID})
	require.NoError(t, err)
	require.Equal(t, 1, counts[constant.Route])
	require.Equal(t, 1, counts[constant.Upstream])
	saved, err := resourcebiz.GetResourceByID(ctx, constant.Route, existingID)
	require.NoError(t, err)
	require.Equal(t, existing.Config, saved.Config)
	require.Equal(t, constant.ResourceStatusUpdateDraft, saved.Status)
	added, err := resourcebiz.GetResourceByID(ctx, constant.Route, newID)
	require.NoError(t, err)
	require.Equal(t, upstreamID, added.GetUpstreamID())
	require.Equal(t, constant.ResourceStatusSuccess, added.Status)
	_, err = resourcebiz.GetResourceByID(ctx, constant.Upstream, upstreamID)
	require.NoError(t, err)
	preview, err := GetResourceConfigDiffDetail(ctx, constant.Route, existingID)
	require.NoError(t, err)
	require.ErrorIs(
		t,
		SyncResourceToEditor(ctx, constant.Route, existingID, preview.SyncToken, true),
		ErrSyncReadOnly,
	)
}

func TestSyncToEditorRequiresEditorDependencies(t *testing.T) {
	for _, test := range []struct {
		kind       constant.APISIXResource
		field      string
		dependency constant.APISIXResource
	}{
		{constant.Route, "service_id", constant.Service},
		{constant.Route, "upstream_id", constant.Upstream},
		{constant.Route, "plugin_config_id", constant.PluginConfig},
		{constant.Service, "upstream_id", constant.Upstream},
		{constant.StreamRoute, "service_id", constant.Service},
		{constant.StreamRoute, "upstream_id", constant.Upstream},
		{constant.Consumer, "group_id", constant.ConsumerGroup},
		{constant.Upstream, "tls.client_cert_id", constant.SSL},
	} {
		t.Run(string(test.kind)+"/"+test.field, func(t *testing.T) {
			id := idx.GenResourceID(test.kind)
			dependencyID := idx.GenResourceID(test.dependency)
			nameField := model.GetResourceNameKey(test.kind)
			before := datatypes.JSON(fmt.Sprintf(`{%q:%q,"desc":"draft"}`, nameField, id))
			common := model.ResourceCommonModel{
				ID:        id,
				GatewayID: gatewayInfo.ID,
				Config:    before,
				Status:    constant.ResourceStatusUpdateDraft,
			}
			typed, err := resourcebiz.NewTypedResourceModel(test.kind, common)
			require.NoError(t, err)
			require.NoError(t, database.Client().WithContext(gatewayCtx).Create(typed).Error)
			original, err := resourcebiz.GetResourceByID(gatewayCtx, test.kind, id)
			require.NoError(t, err)
			incoming, err := sjson.SetBytes(before, test.field, dependencyID)
			require.NoError(t, err)
			snapshot := &model.GatewaySyncData{
				ID: id, GatewayID: gatewayInfo.ID, Type: test.kind, Config: incoming, ModRevision: 1,
			}
			require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
			dependencyConfig := datatypes.JSON(fmt.Sprintf(`{"name":%q}`, dependencyID))
			if test.dependency == constant.SSL {
				dependencyConfig, err = sjson.SetBytes(
					dependencyConfig,
					"cert",
					strings.ReplaceAll(testSSLCert, `\n`, "\n"),
				)
				require.NoError(t, err)
				dependencyConfig, err = sjson.SetBytes(
					dependencyConfig,
					"key",
					strings.ReplaceAll(testSSLKey, `\n`, "\n"),
				)
				require.NoError(t, err)
			}
			require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(&model.GatewaySyncData{
				ID:        dependencyID,
				GatewayID: gatewayInfo.ID,
				Type:      test.dependency,
				Config:    dependencyConfig,
			}))
			// A snapshot, another type with this ID, or another gateway's row cannot satisfy the reference.
			wrongType := constant.PluginConfig
			if test.dependency == wrongType {
				wrongType = constant.ConsumerGroup
			}
			for _, candidate := range []struct {
				kind    constant.APISIXResource
				gateway int
			}{{wrongType, gatewayInfo.ID}, {test.dependency, gatewayInfo.ID + 100}} {
				dependency, err := resourcebiz.NewTypedResourceModel(
					candidate.kind,
					model.ResourceCommonModel{
						ID:        dependencyID,
						GatewayID: candidate.gateway,
						Config:    dependencyConfig,
						Status:    constant.ResourceStatusSuccess,
					},
				)
				require.NoError(t, err)
				require.NoError(t, database.Client().WithContext(gatewayCtx).Create(dependency).Error)
			}
			preview, err := GetResourceConfigDiffDetail(gatewayCtx, test.kind, id)
			require.NoError(t, err)
			err = SyncResourceToEditor(gatewayCtx, test.kind, id, preview.SyncToken, true)
			require.Error(t, err)
			require.Contains(t, err.Error(), dependencyID)
			unchanged, err := resourcebiz.GetResourceByID(gatewayCtx, test.kind, id)
			require.NoError(t, err)
			require.Equal(t, original, unchanged)
			logs, err := repo.OperationAuditLog.WithContext(gatewayCtx).
				Where(repo.OperationAuditLog.ResourceIDs.Eq(id)).Find()
			require.NoError(t, err)
			require.Len(t, logs, 1)

			dependency, err := resourcebiz.NewTypedResourceModel(test.dependency, model.ResourceCommonModel{
				ID: dependencyID, GatewayID: gatewayInfo.ID, Config: dependencyConfig,
				Status: constant.ResourceStatusSuccess,
			})
			require.NoError(t, err)
			require.NoError(t, database.Client().WithContext(gatewayCtx).Create(dependency).Error)
			require.NoError(t, SyncResourceToEditor(gatewayCtx, test.kind, id, preview.SyncToken, true))
			saved, err := resourcebiz.GetResourceByID(gatewayCtx, test.kind, id)
			require.NoError(t, err)
			require.Equal(t, dependencyID, gjson.GetBytes(saved.Config, test.field).String())
			require.Equal(t, constant.ResourceStatusSuccess, saved.Status)
		})
	}
}

func TestSyncToEditorAllResourceTypes(t *testing.T) {
	clientCertID := idx.GenResourceID(constant.SSL)
	require.NoError(t, repo.SSL.WithContext(gatewayCtx).Create(&model.SSL{
		Name: clientCertID,
		ResourceCommonModel: model.ResourceCommonModel{
			ID: clientCertID, GatewayID: gatewayInfo.ID, Status: constant.ResourceStatusSuccess,
			Config: datatypes.JSON(`{"snis":["client.example.com"]}`),
		},
	}))
	for _, kind := range constant.ResourceTypeList {
		t.Run(string(kind), func(t *testing.T) {
			id := idx.GenResourceID(kind)
			config := map[string]any{"name": id, "desc": "before"}
			if kind == constant.Consumer {
				config = map[string]any{"username": id, "desc": "before"}
			}
			if kind == constant.Upstream {
				config["tls"] = map[string]any{"client_cert_id": clientCertID}
			}
			if kind == constant.SSL {
				config["snis"] = []string{"example.com"}
			}
			before, err := json.Marshal(config)
			require.NoError(t, err)
			common := model.ResourceCommonModel{
				ID:        id,
				GatewayID: gatewayInfo.ID,
				Status:    constant.ResourceStatusSuccess,
				Config:    datatypes.JSON(before),
			}
			typed, err := resourcebiz.NewTypedResourceModel(kind, common)
			require.NoError(t, err)
			require.NoError(t, database.Client().WithContext(gatewayCtx).Create(typed).Error)
			config["desc"] = "after"
			after, err := json.Marshal(config)
			require.NoError(t, err)
			snapshot := &model.GatewaySyncData{
				ID:          id,
				GatewayID:   gatewayInfo.ID,
				Type:        kind,
				Config:      datatypes.JSON(after),
				ModRevision: 1,
			}
			require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
			preview, err := GetResourceConfigDiffDetail(gatewayCtx, kind, id)
			require.NoError(t, err)
			require.NoError(t, SyncResourceToEditor(gatewayCtx, kind, id, preview.SyncToken, false))
			saved, err := resourcebiz.GetResourceByID(gatewayCtx, kind, id)
			require.NoError(t, err)
			require.Equal(t, "after", gjson.GetBytes(saved.Config, "desc").String())
			if kind == constant.Upstream {
				require.Equal(t, clientCertID, saved.GetSSLID())
				upstream, queryErr := repo.Upstream.WithContext(
					gatewayCtx,
				).Where(
					repo.Upstream.ID.Eq(id),
				).Take()
				require.NoError(t, queryErr)
				require.Equal(t, clientCertID, upstream.SSLID)
			}
			got, err := syncdatabiz.ConfigStatus(snapshot, &saved, gatewayInfo.GetAPISIXVersionX())
			require.NoError(t, err)
			require.Equal(t, constant.SyncedResourceStatusSuccess, got)
		})
	}
}

func TestSyncToEditorSnapshotChangeInvalidatesConfirmation(t *testing.T) {
	id := idx.GenResourceID(constant.Route)
	route := &model.Route{Name: id, ResourceCommonModel: model.ResourceCommonModel{
		ID: id, GatewayID: gatewayInfo.ID,
		Status: constant.ResourceStatusSuccess, Config: datatypes.JSON(`{"uri":"/old"}`),
	}}
	require.NoError(t, repo.Route.WithContext(gatewayCtx).Create(route))
	snapshot := &model.GatewaySyncData{
		ID: id, GatewayID: gatewayInfo.ID, Type: constant.Route, ModRevision: 1,
		Config: datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/new"}`, id)),
	}
	require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
	preview, err := GetResourceConfigDiffDetail(gatewayCtx, constant.Route, id)
	require.NoError(t, err)
	_, err = repo.GatewaySyncData.WithContext(
		gatewayCtx,
	).Where(
		repo.GatewaySyncData.AutoID.Eq(snapshot.AutoID),
	).Update(
		repo.GatewaySyncData.ModRevision,
		2,
	)
	require.NoError(t, err)
	require.ErrorIs(
		t,
		SyncResourceToEditor(gatewayCtx, constant.Route, id, preview.SyncToken, true),
		ErrSyncConflict,
	)
	saved, err := resourcebiz.GetResourceByID(gatewayCtx, constant.Route, id)
	require.NoError(t, err)
	require.Equal(t, "/old", gjson.GetBytes(saved.Config, "uri").String())
}

func TestSyncToEditorPreservesSSLMetadata(t *testing.T) {
	id := idx.GenResourceID(constant.SSL)
	name := "local-" + id
	cert := strings.ReplaceAll(testSSLCert, `\n`, "\n")
	key := strings.ReplaceAll(testSSLKey, `\n`, "\n")
	config := map[string]any{
		"name":           name,
		"snis":           []string{"test.com"},
		"cert":           cert,
		"key":            key,
		"validity_start": 1,
		"validity_end":   2,
	}
	raw, err := json.Marshal(config)
	require.NoError(t, err)
	ssl := &model.SSL{Name: name, ResourceCommonModel: model.ResourceCommonModel{
		ID: id, GatewayID: gatewayInfo.ID,
		Status: constant.ResourceStatusSuccess, Config: datatypes.JSON(raw),
	}}
	require.NoError(t, repo.SSL.WithContext(gatewayCtx).Create(ssl))
	delete(config, "validity_start")
	delete(config, "validity_end")
	config["name"] = "ssls_" + id
	config["snis"] = []string{"new.example.com"}
	raw, err = json.Marshal(config)
	require.NoError(t, err)
	snapshot := &model.GatewaySyncData{
		ID:          id,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.SSL,
		Config:      datatypes.JSON(raw),
		ModRevision: 1,
	}
	require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
	preview, err := GetResourceConfigDiffDetail(gatewayCtx, constant.SSL, id)
	require.NoError(t, err)
	require.NoError(t, SyncResourceToEditor(gatewayCtx, constant.SSL, id, preview.SyncToken, false))
	saved, err := repo.SSL.WithContext(gatewayCtx).Where(repo.SSL.ID.Eq(id)).Take()
	require.NoError(t, err)
	require.Equal(t, name, saved.Name)
	require.Greater(t, gjson.GetBytes(saved.Config, "validity_start").Int(), int64(1))
	require.Greater(t, gjson.GetBytes(saved.Config, "validity_end").Int(), int64(2))
	require.Equal(t, "new.example.com", gjson.GetBytes(saved.Config, "snis.0").String())
	got, err := syncdatabiz.ConfigStatus(snapshot, &saved.ResourceCommonModel, gatewayInfo.GetAPISIXVersionX())
	require.NoError(t, err)
	require.Equal(t, constant.SyncedResourceStatusSuccess, got)
}

func TestSyncToEditorNumericReferenceRoundTrip(t *testing.T) {
	for _, dependency := range []struct {
		kind constant.APISIXResource
		id   string
	}{{constant.Upstream, "123"}, {constant.Service, "456"}, {constant.PluginConfig, "789"}} {
		typed, err := resourcebiz.NewTypedResourceModel(dependency.kind, model.ResourceCommonModel{
			ID: dependency.id, GatewayID: gatewayInfo.ID, Status: constant.ResourceStatusSuccess,
			Config: datatypes.JSON(fmt.Sprintf(`{"name":%q}`, dependency.id)),
		})
		require.NoError(t, err)
		require.NoError(t, database.Client().WithContext(gatewayCtx).Create(typed).Error)
	}
	id := idx.GenResourceID(constant.Route)
	route := &model.Route{Name: id, ResourceCommonModel: model.ResourceCommonModel{
		ID: id, GatewayID: gatewayInfo.ID,
		Status: constant.ResourceStatusSuccess, Config: datatypes.JSON(`{"uri":"/old"}`),
	}}
	require.NoError(t, repo.Route.WithContext(gatewayCtx).Create(route))
	snapshot := &model.GatewaySyncData{
		ID:          id,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		ModRevision: 1,
		Config: datatypes.JSON(
			fmt.Sprintf(
				`{"name":%q,"uri":"/new","upstream_id":123,"service_id":456,"plugin_config_id":789}`,
				id,
			),
		),
	}
	require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
	preview, err := GetResourceConfigDiffDetail(gatewayCtx, constant.Route, id)
	require.NoError(t, err)
	require.NoError(t, SyncResourceToEditor(gatewayCtx, constant.Route, id, preview.SyncToken, false))
	saved, err := resourcebiz.GetResourceByID(gatewayCtx, constant.Route, id)
	require.NoError(t, err)
	got, err := syncdatabiz.ConfigStatus(snapshot, &saved, gatewayInfo.GetAPISIXVersionX())
	require.NoError(t, err)
	require.Equal(t, constant.SyncedResourceStatusSuccess, got)
}

func TestSyncToEditorPreservesCertificateReferences(t *testing.T) {
	for _, prefix := range []string{"$secret://vault/", "$env://", "$ENV://"} {
		t.Run(prefix, func(t *testing.T) {
			id := idx.GenResourceID(constant.SSL)
			config := map[string]any{
				"name": id, "snis": []string{"before.example.com"},
				"cert": prefix + "SERVER_CERT", "key": prefix + "SERVER_KEY", "validity_end": 123,
			}
			raw, err := json.Marshal(config)
			require.NoError(t, err)
			ssl := &model.SSL{Name: id, ResourceCommonModel: model.ResourceCommonModel{
				ID: id, GatewayID: gatewayInfo.ID,
				Status: constant.ResourceStatusSuccess, Config: datatypes.JSON(raw),
			}}
			require.NoError(t, repo.SSL.WithContext(gatewayCtx).Create(ssl))
			config["snis"] = []string{"after.example.com"}
			delete(config, "validity_end")
			raw, err = json.Marshal(config)
			require.NoError(t, err)
			snapshot := &model.GatewaySyncData{
				ID: id, GatewayID: gatewayInfo.ID, Type: constant.SSL,
				Config: datatypes.JSON(raw), ModRevision: 1,
			}
			require.NoError(t, repo.GatewaySyncData.WithContext(gatewayCtx).Create(snapshot))
			preview, err := GetResourceConfigDiffDetail(gatewayCtx, constant.SSL, id)
			require.NoError(t, err)
			require.NoError(t, SyncResourceToEditor(gatewayCtx, constant.SSL, id, preview.SyncToken, false))
			saved, err := resourcebiz.GetResourceByID(gatewayCtx, constant.SSL, id)
			require.NoError(t, err)
			require.Equal(t, prefix+"SERVER_CERT", gjson.GetBytes(saved.Config, "cert").String())
			require.Equal(t, prefix+"SERVER_KEY", gjson.GetBytes(saved.Config, "key").String())
			require.Equal(t, "after.example.com", gjson.GetBytes(saved.Config, "snis.0").String())
			require.False(t, gjson.GetBytes(saved.Config, "validity_end").Exists())
			got, err := syncdatabiz.ConfigStatus(snapshot, &saved, gatewayInfo.GetAPISIXVersionX())
			require.NoError(t, err)
			require.Equal(t, constant.SyncedResourceStatusSuccess, got)
		})
	}
}
