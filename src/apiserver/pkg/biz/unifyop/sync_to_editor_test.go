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
	"gorm.io/datatypes"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	syncdatabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/syncdata"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

func TestSyncToEditorAllResourceTypes(t *testing.T) {
	for _, kind := range constant.ResourceTypeList {
		t.Run(string(kind), func(t *testing.T) {
			id := idx.GenResourceID(kind)
			config := map[string]any{"name": id, "desc": "before"}
			if kind == constant.Consumer {
				config = map[string]any{"username": id, "desc": "before"}
			}
			if kind == constant.Upstream {
				config["tls"] = map[string]any{"client_cert_id": "client-cert"}
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
				require.Equal(t, "client-cert", saved.GetSSLID())
				upstream, queryErr := repo.Upstream.WithContext(
					gatewayCtx,
				).Where(
					repo.Upstream.ID.Eq(id),
				).Take()
				require.NoError(t, queryErr)
				require.Equal(t, "client-cert", upstream.SSLID)
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
