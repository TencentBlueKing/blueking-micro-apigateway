/*
 * TencentBlueKing is pleased to support the open source community by making
 * BlueKing - Micro APIGateway available.
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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

func TestApplyImportIdentityToSyncData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		syncData   *model.GatewaySyncData
		outerName  string
		assertions func(t *testing.T, syncData *model.GatewaySyncData)
	}{
		{
			name: "route import keeps authoritative name and relation fields",
			syncData: &model.GatewaySyncData{
				GatewayID: 1001,
				Type:      constant.Route,
				ID:        "route-id",
				Config: datatypes.JSON(
					`{"name":"legacy-route","service_id":"legacy-service","uris":["/test"]}`,
				),
			},
			outerName: "route-a",
			assertions: func(t *testing.T, syncData *model.GatewaySyncData) {
				assert.Equal(t, "route-a", syncData.NameValue)
				assert.Equal(t, "legacy-service", syncData.ServiceIDValue)
				assert.Equal(t, "route-a", syncData.GetName())
				assert.Equal(t, "legacy-service", syncData.GetServiceID())
			},
		},
		{
			name: "consumer import uses username semantics",
			syncData: &model.GatewaySyncData{
				GatewayID: 1001,
				Type:      constant.Consumer,
				ID:        "consumer-id",
				Config:    datatypes.JSON(`{"username":"legacy-consumer","group_id":"group-a"}`),
			},
			assertions: func(t *testing.T, syncData *model.GatewaySyncData) {
				assert.Equal(t, "legacy-consumer", syncData.NameValue)
				assert.Equal(t, "group-a", syncData.GroupIDValue)
				assert.Equal(t, "legacy-consumer", syncData.GetName())
				assert.Equal(t, "group-a", syncData.GetGroupID())
			},
		},
		{
			name: "plugin metadata import uses derived name from config id",
			syncData: &model.GatewaySyncData{
				GatewayID: 1001,
				Type:      constant.PluginMetadata,
				ID:        "plugin-metadata-row-id",
				Config:    datatypes.JSON(`{"id":"jwt-auth","name":"legacy-name","key":"value"}`),
			},
			assertions: func(t *testing.T, syncData *model.GatewaySyncData) {
				assert.Equal(t, "jwt-auth", syncData.NameValue)
				assert.Equal(t, "jwt-auth", syncData.GetName())
				assert.Equal(t, "jwt-auth", syncData.GetConfigID())
			},
		},
		{
			name: "upstream import extracts tls client cert relation",
			syncData: &model.GatewaySyncData{
				GatewayID: 1001,
				Type:      constant.Upstream,
				ID:        "upstream-id",
				Config: datatypes.JSON(
					`{"name":"upstream-a","tls":{"client_cert_id":"ssl-id"},"type":"roundrobin"}`,
				),
			},
			assertions: func(t *testing.T, syncData *model.GatewaySyncData) {
				assert.Equal(t, "ssl-id", syncData.SSLIDValue)
				assert.Equal(t, "ssl-id", syncData.GetSSLID())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyImportIdentityToSyncData(tt.syncData, tt.outerName)
			tt.assertions(t, tt.syncData)
		})
	}
}

func TestClassifyImportResourceInfoUsesResolvedNames(t *testing.T) {
	t.Parallel()

	importData := map[constant.APISIXResource][]*ResourceInfo{
		constant.PluginMetadata: {
			{
				ResourceType: constant.PluginMetadata,
				ResourceID:   "plugin-metadata-id",
				Config:       json.RawMessage(`{"id":"jwt-auth","name":"legacy-name","key":"value"}`),
			},
		},
	}

	uploadInfo, err := ClassifyImportResourceInfo(
		importData,
		map[string]struct{}{},
		map[string]*model.GatewayCustomPluginSchema{},
	)
	assert.NoError(t, err)
	if assert.Len(t, uploadInfo.Add[constant.PluginMetadata], 1) {
		assert.Equal(t, "jwt-auth", uploadInfo.Add[constant.PluginMetadata][0].Name)
	}
}
