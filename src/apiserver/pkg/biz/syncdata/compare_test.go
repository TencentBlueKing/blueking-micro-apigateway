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

package syncdata

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

func TestConfigStatus(t *testing.T) {
	for _, test := range []struct {
		name          string
		kind          constant.APISIXResource
		version       constant.APISIXVersion
		before, after string
		want          constant.SyncStatus
	}{
		{"order and top-level timestamps", constant.Route, constant.APISIXVersion313, `{"uri":"/","id":1,"update_time":2,"plugins":{"key-auth":{}}}`, `{"plugins":{"key-auth":{}},"uri":"/","id":"1"}`, constant.SyncedResourceStatusSuccess},
		{"empty plugin is enabled", constant.Route, constant.APISIXVersion313, `{"plugins":{"key-auth":{}}}`, `{"plugins":{}}`, constant.SyncedResourceStatusDifferent},
		{"plugin timestamp is meaningful", constant.Route, constant.APISIXVersion313, `{"plugins":{"custom":{"update_time":1}}}`, `{"plugins":{"custom":{"update_time":2}}}`, constant.SyncedResourceStatusDifferent},
		{"plugin meta is meaningful", constant.Route, constant.APISIXVersion313, `{"plugins":{"key-auth":{"_meta":{"disable":true}}}}`, `{"plugins":{"key-auth":{}}}`, constant.SyncedResourceStatusDifferent},
		{"large integers retain precision", constant.Route, constant.APISIXVersion313, `{"plugins":{"custom":{"value":9007199254740992}}}`, `{"plugins":{"custom":{"value":9007199254740993}}}`, constant.SyncedResourceStatusDifferent},
		{"3.18 plugin config name", constant.PluginConfig, constant.APISIXVersion318, `{"plugins":{"key-auth":{}}}`, `{"name":"local-name","plugins":{"key-auth":{}}}`, constant.SyncedResourceStatusSuccess},
		{"3.13 plugin config name", constant.PluginConfig, constant.APISIXVersion313, `{"name":"old"}`, `{"name":"new"}`, constant.SyncedResourceStatusDifferent},
		{"SSL internal validity", constant.SSL, constant.APISIXVersion313, `{"cert":"cert"}`, `{"cert":"cert","name":"local","validity_start":1,"validity_end":2}`, constant.SyncedResourceStatusSuccess},
		{"stream labels", constant.StreamRoute, constant.APISIXVersion313, `{"server_port":123}`, `{"server_port":123,"labels":{"internal":"1"}}`, constant.SyncedResourceStatusSuccess},
		{"array order matters", constant.Route, constant.APISIXVersion313, `{"plugins":{"custom":{"items":[1,2]}}}`, `{"plugins":{"custom":{"items":[2,1]}}}`, constant.SyncedResourceStatusDifferent},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := []byte(test.before)
			after := []byte(test.after)
			snapshot := &model.GatewaySyncData{Type: test.kind, Config: datatypes.JSON(before)}
			editor := &model.ResourceCommonModel{
				Config: datatypes.JSON(after),
				Status: constant.ResourceStatusSuccess,
			}
			got, err := ConfigStatus(snapshot, editor, test.version)
			require.NoError(t, err)
			require.Equal(t, test.want, got)
			require.Equal(t, test.before, string(snapshot.Config))
			require.Equal(t, test.after, string(editor.Config))
		})
	}
	for _, invalid := range []string{`null`, `[]`, `invalid`} {
		_, err := ComparableConfig(json.RawMessage(invalid), constant.Route, constant.APISIXVersion313)
		require.Error(t, err)
	}
}

func TestConfigStatusNormalizesOnlyTypedReferences(t *testing.T) {
	for _, test := range []struct {
		kind          constant.APISIXResource
		before, after string
	}{
		{constant.Route, `{"upstream_id":123,"service_id":456,"plugin_config_id":789}`, `{"upstream_id":"123","service_id":"456","plugin_config_id":"789"}`},
		{constant.Service, `{"upstream_id":123}`, `{"upstream_id":"123"}`},
		{constant.StreamRoute, `{"upstream_id":123,"service_id":456}`, `{"upstream_id":"123","service_id":"456"}`},
		{constant.Consumer, `{"group_id":123}`, `{"group_id":"123"}`},
		{constant.Upstream, `{"tls":{"client_cert_id":123}}`, `{"tls":{"client_cert_id":"123"}}`},
	} {
		snapshot := &model.GatewaySyncData{Type: test.kind, Config: datatypes.JSON(test.before)}
		editor := &model.ResourceCommonModel{
			Config: datatypes.JSON(test.after),
			Status: constant.ResourceStatusSuccess,
		}
		got, err := ConfigStatus(snapshot, editor, constant.APISIXVersion313)
		require.NoError(t, err)
		require.Equal(t, constant.SyncedResourceStatusSuccess, got)
	}
	snapshot := &model.GatewaySyncData{
		Type:   constant.Route,
		Config: datatypes.JSON(`{"plugins":{"custom":{"upstream_id":123}}}`),
	}
	editor := &model.ResourceCommonModel{
		Config: datatypes.JSON(`{"plugins":{"custom":{"upstream_id":"123"}}}`),
		Status: constant.ResourceStatusSuccess,
	}
	got, err := ConfigStatus(snapshot, editor, constant.APISIXVersion313)
	require.NoError(t, err)
	require.Equal(t, constant.SyncedResourceStatusDifferent, got)
}
