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

package biz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
)

func TestBuildSyncedResourceFromKV(t *testing.T) {
	t.Parallel()

	normalizedPrefix := model.NormalizeEtcdPrefix("/apisix")

	t.Run("route strips timestamps and injects fallback name", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/routes/route-id",
			Value:       `{"uri":"/demo","create_time":111,"update_time":222}`,
			ModRevision: 9,
		})
		assert.True(t, ok)
		assert.Equal(t, "route-id", got.ID)
		assert.Equal(t, 17, got.GatewayID)
		assert.Equal(t, constant.Route, got.Type)
		assert.Equal(t, 9, got.ModRevision)
		assert.Equal(t, "routes_route-id", gjson.GetBytes(got.Config, "name").String())
		assert.False(t, gjson.GetBytes(got.Config, "create_time").Exists())
		assert.False(t, gjson.GetBytes(got.Config, "update_time").Exists())
	})

	t.Run("plugin metadata uses etcd key as snapshot name", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/plugin_metadata/clickhouse-logger",
			Value:       `{"value":{"disable":false}}`,
			ModRevision: 3,
		})
		assert.True(t, ok)
		assert.Equal(t, constant.PluginMetadata, got.Type)
		assert.Equal(t, "clickhouse-logger", got.ID)
		assert.Equal(t, "clickhouse-logger", got.GetName())
	})

	t.Run("invalid key is ignored", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/routes/too/many/parts",
			Value:       `{}`,
			ModRevision: 1,
		})
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}
