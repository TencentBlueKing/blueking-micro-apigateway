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

package unifyop

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	gatewaybiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/gateway"
	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	syncdatabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/syncdata"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

var (
	gatewayInfo  *model.Gateway
	gatewayCtx   context.Context
	etcdEndpoint string
)

func TestMain(m *testing.M) {
	err := cryptography.Init("jxi18GX5w2qgHwfZCFpn07q8FScXJOd3", "k2dbCGetyusW")
	if err != nil {
		panic(err)
	}

	util.InitEmbedDb()

	_, server, endpoint, err := util.StartEmbedEtcdClientRandom(context.Background())
	if err != nil {
		panic(err)
	}
	etcdEndpoint = endpoint

	gatewayInfo = data.Gateway1WithBkAPISIX()
	gatewayInfo.EtcdConfig.Endpoint = base.Endpoint(etcdEndpoint)
	err = gatewaybiz.CreateGateway(context.Background(), gatewayInfo)
	if err != nil {
		panic(err)
	}
	gatewayCtx = ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)

	code := m.Run()

	server.Close()
	os.Exit(code)
}

const (
	testSSLCert = `-----BEGIN CERTIFICATE-----\nMIIDJzCCAg+gAwIBAgIRAJvCZRh2nejK7+Ss3AgrEa0wDQYJKoZIhvcNAQELBQAw\ngYoxEjAQBgNVBAMMCWxkZGdvLm5ldDEMMAoGA1UECwwDZGV2MQ4wDAYDVQQKDAVs\nZGRnbzELMAkGA1UEBhMCQ04xIzAhBgkqhkiG9w0BCQEWFGxlY2hlbmdhZG1pbkAx\nMjYuY29tMREwDwYDVQQHDAhzaGFuZ2hhaTERMA8GA1UECAwIc2hhbmdoYWkwHhcN\nMjUwMjI2MDE0ODQ0WhcNMjcwMjI2MDE0ODQ0WjATMREwDwYDVQQDDAh0ZXN0LmNv\nbTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAIIJ82TMFlWOR7dDkJ0X\nLclmCUDlefEJY2laYPWxaCe3oaIndosUmgm5aovYUTWDRAByn56HPFub5fc2Kt9v\n5+HWVd149JuP43F5NXaUKbE6GuXUWR7WhorzIRbabvvkE4SdpkrGwthi6AxUnvKK\naHKn11hSk+MBUWxjhSJoQy/ds3fKSpq7j+LAMRmQo9a3uW/HBl7FdfWIH5ZTN3Q8\n+ZDMc2zrEqOXFBGFBwzsbcVGNppMkUBuYmxIp7O3slB7rH7oOkdpYReIwWQOOswO\nhbBu5UGqC8nMX0N0jhzMyxrvDOIFSjjKiXuu46qd+t/GxUB9+8ZJ/Fn3WsJ6iQf7\n+cMCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEARSufAXUin/eFxcpojYMZ6F3t6VYp\njiZ+3Sx+UjQ4mq3qq8eQ/r0haxGtw2GeMuyprfxj6YTX6erQlJKkDk8vJXpDbFR4\n4dj1g4VQDZshPH2j2HJ/4l/kAvbDy/Rj9eIdV0Ux+t8s7MYgP7yf35Nb1ejJyWhB\nPS56NWCyj43lJcwnUmH4EAvLiFdgGgiaPQdm2/XlyEd8UVZugihIgjlQ3XKwMwsb\nXFfjJdDgdhFO5jmtU+rdEQWuaJDCEEWQJfMFmWRGApri97T/14QOulTqCXfk8+Wq\nw4WMGMQt3zIALlf7Meknv2qfTxax3JAO8lf7KuN5A4S5SuqAHke9NfGzAA==\n-----END CERTIFICATE-----`
	testSSLKey  = `-----BEGIN RSA PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCCCfNkzBZVjke3\nQ5CdFy3JZglA5XnxCWNpWmD1sWgnt6GiJ3aLFJoJuWqL2FE1g0QAcp+ehzxbm+X3\nNirfb+fh1lXdePSbj+NxeTV2lCmxOhrl1Fke1oaK8yEW2m775BOEnaZKxsLYYugM\nVJ7yimhyp9dYUpPjAVFsY4UiaEMv3bN3ykqau4/iwDEZkKPWt7lvxwZexXX1iB+W\nUzd0PPmQzHNs6xKjlxQRhQcM7G3FRjaaTJFAbmJsSKezt7JQe6x+6DpHaWEXiMFk\nDjrMDoWwbuVBqgvJzF9DdI4czMsa7wziBUo4yol7ruOqnfrfxsVAffvGSfxZ91rC\neokH+/nDAgMBAAECggEACSzKj4IW0VKInNWXjn3kLSGV5Y5LXEZdTUGjNbKetq6u\nKNK/+nApriX27ocEs9HfKmjr+jNwfsYxI5Ae1kT/B2AoDshJ+e/dDFSRARzTFD4V\nR8IDx7k7JPKikwo2am9dMS4uXXhIpxvTY4tU66f4Vp6hAwpQhOPC6vLaoeLZWrcg\nAjjPTud/1N8D+CMsnsrfLh9XPLvUZIqYm5DCgE6fFle1/X/YrqzzMzflCG3Ns5Gv\nMY0i1xR7baAj8nT9iG+MCvCW8Ak2++pweX2Hli6l5aqk+esDU/zUAdddJdtpufGT\nkobCOKtqNXzEj6UGrsQU/27dc1tQKt4VgRvsgC+aAQKBgQC5zySFCpqtZY/naKnw\nGXf1Pl7r8aTuWVA+8ziRiyPlyI60oMHhu0bSIoRIh7lpa8km/cNsJOMTFWmHUANT\ndu53icmSCO++M1d+nrl3aWYyqbAlFvqMPtiW5/pYRnWJi4GSQTonGY32EhmN1qo5\nJbmj7NVxRnX0g9OTX4+f5MdCUQKBgQCzKXzwim/KxeOeVURVu/LQGK+Or2Ssyzjr\nz8MPQ2OE5DX528hLkE5h0EVhffSrsTfQiiMIhzU/Rywa7khNRqsTmhFEHM5JI+Rl\nGZgGgG4T5Q3idfrx3jXGqMylmoR0pA+4aGpSGg135vuIhJWCn8RI/mgMl0KP6Nax\nSSZkex4B0wKBgFr470FwIrEY068SEHnsjk31fpX4lq7X7bEUdjLUM/wyCKSpPKPf\nhFon6ip0wTO7QR4lCoQtPzw9tJA6fZZk2XaPcLBeTbsK+iCVZ+ruIMpXSFWwfXUi\n4/pmk6yaurtgIU1RQD6ahWXgEMDgRDF8pfp7Xzl5rRDNZk52cCRx55kxAoGAV4/p\nTi56oKHCszl9ImGvNGE8PAIgtArGkQmDjcwjsWlPsAPoinXGuStvHUzP7bG5U6SP\nprVeIsUIG0ll8M6fAf+EfMOPVlPCZl7x3AucwQBrnsiGkvtFUQhirHUuU0tzm278\nt4+gEX/EY15ZK/QlnH8qHy02DNuBQjg8GVPKwJ0CgYATHdUKjNJG0dMkJ8pjjsI1\nXOYqFo7bXeA5iw6gvmhGTt0Oc7QkOt/VWyvGvRn4UPXcaZixEsFj+rKVlCbZG9gJ\nDvC3nKL8jGXiVs0eJot2WHZJlM04YqzSlaqBNW5O+p/IMmJ1q1zehGm1oIHq0RlA\ncO+a+H4tgy7YSbgYm32XKQ==\n-----END RSA PRIVATE KEY-----`
)

type mockEtcdStore struct {
	storage.StorageInterface
	data   map[string]string
	kvList []storage.KeyValuePair
}

func (m *mockEtcdStore) List(ctx context.Context, prefix string) ([]storage.KeyValuePair, error) {
	if len(m.kvList) > 0 {
		return append([]storage.KeyValuePair(nil), m.kvList...), nil
	}

	var kvList []storage.KeyValuePair
	for key, value := range m.data {
		kvList = append(kvList, storage.KeyValuePair{
			Key:         key,
			Value:       value,
			ModRevision: 1,
		})
	}
	return kvList, nil
}

func clearGatewaySyncData(t *testing.T, ctx context.Context) {
	t.Helper()

	u := repo.GatewaySyncData
	_, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Delete()
	assert.NoError(t, err)
}

func TestGetSyncItemsAssociatedResources_ReturnsDependencies(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	clearGatewaySyncData(t, ctx)

	serviceID := idx.GenResourceID(constant.Service)
	upstreamID := idx.GenResourceID(constant.Upstream)
	pluginConfigID := idx.GenResourceID(constant.PluginConfig)
	groupID := idx.GenResourceID(constant.ConsumerGroup)
	sslID := idx.GenResourceID(constant.SSL)

	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).CreateInBatches([]*model.GatewaySyncData{
		{
			ID:        serviceID,
			GatewayID: gatewayInfo.ID,
			Type:      constant.Service,
			Config:    datatypes.JSON(`{"id":"` + serviceID + `","name":"service-associated"}`),
		},
		{
			ID:        upstreamID,
			GatewayID: gatewayInfo.ID,
			Type:      constant.Upstream,
			Config:    datatypes.JSON(`{"id":"` + upstreamID + `","name":"upstream-associated"}`),
		},
		{
			ID:        pluginConfigID,
			GatewayID: gatewayInfo.ID,
			Type:      constant.PluginConfig,
			Config:    datatypes.JSON(`{"id":"` + pluginConfigID + `","name":"plugin-config-associated"}`),
		},
		{
			ID:        groupID,
			GatewayID: gatewayInfo.ID,
			Type:      constant.ConsumerGroup,
			Config:    datatypes.JSON(`{"id":"` + groupID + `","name":"consumer-group-associated"}`),
		},
		{
			ID:        sslID,
			GatewayID: gatewayInfo.ID,
			Type:      constant.SSL,
			Config:    datatypes.JSON(`{"id":"` + sslID + `","snis":["example.com"]}`),
		},
	}, 100))

	items := []*model.GatewaySyncData{
		{
			ID:        idx.GenResourceID(constant.Route),
			GatewayID: gatewayInfo.ID,
			Type:      constant.Route,
			Config: datatypes.JSON(
				`{"name":"route-owner","uris":["/test"],"service_id":"` + serviceID +
					`","upstream_id":"` + upstreamID + `","plugin_config_id":"` + pluginConfigID + `"}`,
			),
		},
		{
			ID:        idx.GenResourceID(constant.Consumer),
			GatewayID: gatewayInfo.ID,
			Type:      constant.Consumer,
			Config:    datatypes.JSON(`{"username":"consumer-owner","group_id":"` + groupID + `"}`),
		},
		{
			ID:        idx.GenResourceID(constant.Upstream),
			GatewayID: gatewayInfo.ID,
			Type:      constant.Upstream,
			Config:    datatypes.JSON(`{"name":"upstream-owner","tls":{"client_cert_id":"` + sslID + `"}}`),
		},
	}

	associated, err := GetSyncItemsAssociatedResources(ctx, items)
	assert.NoError(t, err)
	if !assert.Len(t, associated, 5) {
		return
	}

	associatedByKey := make(map[string]*model.GatewaySyncData, len(associated))
	for _, item := range associated {
		associatedByKey[item.GetResourceKey()] = item
	}

	assert.Contains(t, associatedByKey, fmt.Sprintf(constant.ResourceKeyFormat, constant.Service, serviceID))
	assert.Contains(t, associatedByKey, fmt.Sprintf(constant.ResourceKeyFormat, constant.Upstream, upstreamID))
	assert.Contains(
		t,
		associatedByKey,
		fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, pluginConfigID),
	)
	assert.Contains(t, associatedByKey, fmt.Sprintf(constant.ResourceKeyFormat, constant.ConsumerGroup, groupID))
	assert.Contains(t, associatedByKey, fmt.Sprintf(constant.ResourceKeyFormat, constant.SSL, sslID))
}

func TestSyncWithPrefix_UpsertBehavior(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	clearGatewaySyncData(t, ctx)

	resource1ID := idx.GenResourceID(constant.Route)
	resource2ID := idx.GenResourceID(constant.Route)
	resource3ID := idx.GenResourceID(constant.Route)
	resource4ID := idx.GenResourceID(constant.Route)

	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:          resource1ID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + resource1ID + `","name":"route-1-old","uris":["/old"]}`),
		ModRevision: 1,
	}))
	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:        resource2ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"id":"` + resource2ID + `","name":"route-2-to-delete","uris":["/delete"]}`,
		),
		ModRevision: 1,
	}))
	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:        resource4ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"id":"` + resource4ID + `","name":"route-4-unchanged","uris":["/unchanged"]}`,
		),
		ModRevision: 5,
	}))

	before1, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resource1ID)
	assert.NoError(t, err)
	before4, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resource4ID)
	assert.NoError(t, err)

	prefix := gatewayInfo.GetEtcdPrefixForList()
	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			kvList: []storage.KeyValuePair{
				{
					Key:         prefix + "routes/" + resource1ID,
					Value:       `{"name":"route-1-updated","uris":["/updated"]}`,
					ModRevision: 2,
				},
				{
					Key:         prefix + "routes/" + resource3ID,
					Value:       `{"name":"route-3-new","uris":["/new"]}`,
					ModRevision: 1,
				},
				{
					Key:         prefix + "routes/" + resource4ID,
					Value:       `{"name":"route-4-unchanged","uris":["/unchanged"]}`,
					ModRevision: 5,
				},
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	counts, err := syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)
	assert.Equal(t, 1, counts[constant.Route])

	updated1, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resource1ID)
	assert.NoError(t, err)
	assert.Equal(t, before1.AutoID, updated1.AutoID)
	assert.Equal(t, 2, updated1.ModRevision)
	assert.Equal(t, "route-1-updated", gjson.GetBytes(updated1.Config, "name").String())
	assert.Equal(t, "/updated", gjson.GetBytes(updated1.Config, "uris.0").String())

	_, err = syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resource2ID)
	assert.Error(t, err)

	created3, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resource3ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, created3.ModRevision)
	assert.Equal(t, "route-3-new", gjson.GetBytes(created3.Config, "name").String())

	unchanged4, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resource4ID)
	assert.NoError(t, err)
	assert.Equal(t, before4.AutoID, unchanged4.AutoID)
	assert.Equal(t, 5, unchanged4.ModRevision)
	assert.Equal(t, "route-4-unchanged", gjson.GetBytes(unchanged4.Config, "name").String())

	u := repo.GatewaySyncData
	allSnapshots, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Find()
	assert.NoError(t, err)
	assert.Len(t, allSnapshots, 3)
}

func TestSyncWithPrefix_NoRaceCondition(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	clearGatewaySyncData(t, ctx)

	resourceID := idx.GenResourceID(constant.Route)
	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:          resourceID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + resourceID + `","name":"route-1","uris":["/test"]}`),
		ModRevision: 1,
	}))

	before, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resourceID)
	assert.NoError(t, err)

	prefix := gatewayInfo.GetEtcdPrefixForList()
	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			kvList: []storage.KeyValuePair{
				{
					Key:         prefix + "routes/" + resourceID,
					Value:       `{"name":"route-1-updated","uris":["/test-updated"]}`,
					ModRevision: 2,
				},
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	_, err = syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)

	after, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, resourceID)
	assert.NoError(t, err)
	assert.Equal(t, before.AutoID, after.AutoID)
	assert.Equal(t, 2, after.ModRevision)
	assert.Equal(t, "route-1-updated", gjson.GetBytes(after.Config, "name").String())
	assert.Equal(t, "/test-updated", gjson.GetBytes(after.Config, "uris.0").String())
}

func TestSyncWithPrefix_BatchProcessing(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	clearGatewaySyncData(t, ctx)

	const resourceCount = 510
	prefix := gatewayInfo.GetEtcdPrefixForList()
	kvList := make([]storage.KeyValuePair, 0, resourceCount)

	for i := 0; i < resourceCount; i++ {
		resourceID := idx.GenResourceID(constant.Route)
		kvList = append(kvList, storage.KeyValuePair{
			Key:         prefix + "routes/" + resourceID,
			Value:       fmt.Sprintf(`{"name":"route-%d","uris":["/test-%d"]}`, i, i),
			ModRevision: int64(i + 1),
		})
	}

	syncer := &UnifyOp{
		etcdStore:   &mockEtcdStore{kvList: kvList},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	counts, err := syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)
	assert.Equal(t, resourceCount, counts[constant.Route])

	u := repo.GatewaySyncData
	allSnapshots, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Find()
	assert.NoError(t, err)
	assert.Len(t, allSnapshots, resourceCount)
}

func TestSyncWithPrefix_SnapshotConfigShaping_CurrentSeam(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	clearGatewaySyncData(t, ctx)

	routeID := idx.GenResourceID(constant.Route)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	existingMetadata := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusSuccess)
	existingMetadata.Name = "limit-count-" + suffix
	assert.NoError(t, resourcebiz.CreatePluginMetadata(ctx, *existingMetadata))

	existingGroup := data.ConsumerGroup1WithNoRelation(gatewayInfo, constant.ResourceStatusSuccess)
	existingGroup.Name = "cg-from-db-" + suffix
	assert.NoError(t, resourcebiz.CreateConsumerGroup(ctx, *existingGroup))

	existingStreamRoute := data.StreamRoute1WithNoRelationResource(
		gatewayInfo, constant.ResourceStatusSuccess,
	)
	existingStreamRoute.Name = "sr-from-db-" + suffix
	existingStreamRoute.Config, _ = sjson.SetBytes(
		existingStreamRoute.Config, "labels", map[string]string{"env": "test"},
	)
	assert.NoError(t, resourcebiz.CreateStreamRoute(ctx, *existingStreamRoute))

	prefix := gatewayInfo.GetEtcdPrefixForList()
	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			data: map[string]string{
				prefix + "routes/" + routeID:                        `{"uri":"/from-etcd","create_time":111,"update_time":222}`,
				prefix + "plugin_metadata/" + existingMetadata.Name: `{"value":{"disable":false}}`,
				prefix + "consumer_groups/" + existingGroup.ID:      `{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
				prefix + "stream_routes/" + existingStreamRoute.ID:  `{"server_addr":"127.0.0.1","server_port":8080}`,
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	_, err := syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)

	routeSnapshot, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, routeID)
	assert.NoError(t, err)
	assert.Equal(t, "routes_"+routeID, gjson.GetBytes(routeSnapshot.Config, "name").String())
	assert.False(t, gjson.GetBytes(routeSnapshot.Config, "create_time").Exists())
	assert.False(t, gjson.GetBytes(routeSnapshot.Config, "update_time").Exists())

	metadataSnapshot, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
		ctx, constant.PluginMetadata, existingMetadata.ID,
	)
	assert.NoError(t, err)
	assert.Equal(t, existingMetadata.Name, metadataSnapshot.GetName())

	groupSnapshot, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
		ctx,
		constant.ConsumerGroup,
		existingGroup.ID,
	)
	assert.NoError(t, err)
	assert.Equal(t, existingGroup.ID, gjson.GetBytes(groupSnapshot.Config, "id").String())
	assert.Equal(t, existingGroup.Name, gjson.GetBytes(groupSnapshot.Config, "name").String())

	streamRouteSnapshot, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
		ctx, constant.StreamRoute, existingStreamRoute.ID,
	)
	assert.NoError(t, err)
	assert.Equal(t, existingStreamRoute.Name, gjson.GetBytes(streamRouteSnapshot.Config, "name").String())
	assert.Equal(t, "test", gjson.GetBytes(streamRouteSnapshot.Config, "labels.env").String())
}

func TestSyncWithPrefix_ReturnsOnlyNewSnapshotCounts(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	clearGatewaySyncData(t, ctx)

	prefix := gatewayInfo.GetEtcdPrefixForList()
	existingID := idx.GenResourceID(constant.Route)
	newID := idx.GenResourceID(constant.Route)

	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:          existingID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + existingID + `","name":"existing-route","uri":"/existing"}`),
		ModRevision: 1,
	}))

	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			data: map[string]string{
				prefix + "routes/" + existingID: `{"name":"existing-route-updated","uri":"/updated"}`,
				prefix + "routes/" + newID:      `{"name":"new-route","uri":"/new"}`,
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	counts, err := syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)
	assert.Equal(t, 1, counts[constant.Route])
}

func TestSyncWithPrefix_ReturnsErrorOnHelperLookupFailure(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	clearGatewaySyncData(t, ctx)

	prefix := gatewayInfo.GetEtcdPrefixForList()
	existingID := idx.GenResourceID(constant.Route)
	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:          existingID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + existingID + `","name":"existing-route","uri":"/existing"}`),
		ModRevision: 1,
	}))

	db := database.Client()
	assert.NoError(t, db.Migrator().DropTable(&model.PluginMetadata{}))
	t.Cleanup(func() {
		restoreErr := db.Exec(
			"CREATE TABLE `plugin_metadata` (`name` varchar(255),`creator` varchar(32),`updater` varchar(32),`created_at` datetime,`updated_at` datetime,`auto_id` integer PRIMARY KEY AUTOINCREMENT,`id` varchar(255),`gateway_id` integer,`config` JSON,`status` varchar(32))",
		).Error
		if restoreErr != nil {
			t.Fatalf("restore plugin_metadata table: %v", restoreErr)
		}
	})

	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			data: map[string]string{
				prefix + "plugin_metadata/failing-plugin": `{"value":{"disable":false}}`,
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	_, err := syncer.SyncWithPrefix(ctx, prefix)
	assert.Error(t, err)

	storedSnapshot, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, constant.Route, existingID)
	assert.NoError(t, err)
	assert.Equal(t, existingID, storedSnapshot.ID)
	assert.Equal(t, 1, storedSnapshot.ModRevision)

	u := repo.GatewaySyncData
	allSnapshots, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID)).
		Find()
	assert.NoError(t, err)
	assert.Len(t, allSnapshots, 1)
}

// TestInsertSyncedResources_RemoveDuplicated 验证 InsertSyncedResources 会移除与数据库已有资源 id/name 冲突的条目
func TestInsertSyncedResources_RemoveDuplicated(t *testing.T) {
	// 依赖包级 TestMain 初始化：gatewayInfo / gatewayCtx / embedDB
	// 1) 先创建一条已存在的 Route 资源（模拟数据库已有记录）
	existing := model.Route{
		Name: "dup-name",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "dup-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"dup-name"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	assert.NoError(t, resourcebiz.CreateRoute(gatewayCtx, existing))

	// 2) 构造三条同步资源：
	//   - 与数据库 ID 冲突（相同 id: dup-id）
	//   - 与数据库 Name 冲突（相同 name: dup-name）
	//   - 完全不冲突（应被成功插入）
	dupID := &model.GatewaySyncData{
		ID:        "dup-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"new-name-for-dup-id"}`),
	}
	dupName := &model.GatewaySyncData{
		ID:        "new-id-for-dup-name",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"dup-name"}`),
	}
	normal := &model.GatewaySyncData{
		ID:        "ok-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"ok-name"}`),
	}

	// 3) 调用 InsertSyncedResources（内部会调用 RemoveDuplicatedResource 做去重）
	typeSynced := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {dupID, dupName, normal},
	}
	err := InsertSyncedResources(gatewayCtx, typeSynced, constant.ResourceStatusSuccess)
	// 有冲突会报错
	assert.Error(t, err)

	// 4) 断言：数据库中不会新增与 existing 冲突的两条，只应新增 normal 这一条) 调用 InsertSyncedResources（内部会调用 RemoveDuplicatedResource 做去重）
	typeSynced = map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {dupID, normal},
	}
	err = InsertSyncedResources(gatewayCtx, typeSynced, constant.ResourceStatusSuccess)

	assert.NoError(t, err)

	if _, err := resourcebiz.GetRoute(gatewayCtx, "dup-id"); err == nil {
		// 依旧只能是 existing 这条，状态保持 success
		r, err := resourcebiz.GetRoute(gatewayCtx, "dup-id")
		assert.NoError(t, err)
		assert.Equal(t, "dup-name", r.Name)
		assert.Equal(t, constant.ResourceStatusSuccess, r.Status)
	}
	//    - 冲突 Name 的记录不应被创建（按 id 唯一，new-id-for-dup-name 不应落库为新资源）
	_, err = resourcebiz.GetRoute(gatewayCtx, "new-id-for-dup-name")
	assert.Error(t, err)

	//    - 正常的不冲突记录应被创建
	r, err := resourcebiz.GetRoute(gatewayCtx, "ok-id")
	assert.NoError(t, err)
	assert.Equal(t, "ok-name", r.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, r.Status)
}

// TestUploadResources_DifferentGatewaysSameID 测试不同网关存在相同资源ID的情况
func TestUploadResources_DifferentGatewaysSameID(t *testing.T) {
	// 创建第二个网关
	gateway2 := &model.Gateway{
		Name:          "gateway2",
		Mode:          1,
		Maintainers:   []string{"user1"},
		Desc:          "gateway2",
		APISIXType:    constant.APISIXTypeBKAPISIX,
		APISIXVersion: "3.11.0",
		EtcdConfig: model.EtcdConfig{
			InstanceID: "987654321",
			EtcdConfig: base.EtcdConfig{
				Endpoint: base.Endpoint(etcdEndpoint),
				Username: "test",
				Password: "test",
				Prefix:   "/apisix2",
			},
		},
	}
	err := gatewaybiz.CreateGateway(context.Background(), gateway2)
	assert.NoError(t, err)
	gateway2Ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway2)

	// 在第一个网关中创建资源
	route1 := &model.Route{
		Name: "same-id-route",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "same-resource-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"same-id-route","uris":["/gateway1"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err = resourcebiz.CreateRoute(gatewayCtx, *route1)
	assert.NoError(t, err)

	// 在第二个网关中创建相同ID的资源
	route2 := &model.Route{
		Name: "same-id-route-gateway2",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "same-resource-id",
			GatewayID: gateway2.ID,
			Config:    datatypes.JSON(`{"name":"same-id-route-gateway2","uris":["/gateway2"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err = resourcebiz.CreateRoute(gateway2Ctx, *route2)
	assert.NoError(t, err)

	// 验证两个网关的资源都存在且互不影响
	route1FromDB, err := resourcebiz.GetRoute(gatewayCtx, "same-resource-id")
	assert.NoError(t, err)
	assert.Equal(t, "same-id-route", route1FromDB.Name)
	assert.Equal(t, gatewayInfo.ID, route1FromDB.GatewayID)

	route2FromDB, err := resourcebiz.GetRoute(gateway2Ctx, "same-resource-id")
	assert.NoError(t, err)
	assert.Equal(t, "same-id-route-gateway2", route2FromDB.Name)
	assert.Equal(t, gateway2.ID, route2FromDB.GatewayID)

	// 清理第二个网关
	err = gatewaybiz.DeleteGateway(context.Background(), gateway2)
	assert.NoError(t, err)
}

// TestUploadResources_SameGatewayUpdateAndAdd 测试同一网关的更新覆盖和新增情况
func TestUploadResources_SameGatewayUpdateAndAdd(t *testing.T) {
	// 先创建一些现有资源
	existingRoute := &model.Route{
		Name: "existing-route",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "existing-route-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"existing-route","uris":["/existing"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateRoute(gatewayCtx, *existingRoute)
	assert.NoError(t, err)

	existingService := &model.Service{
		Name: "existing-service",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "existing-service-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"existing-service"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err = resourcebiz.CreateService(gatewayCtx, *existingService)
	assert.NoError(t, err)

	// 准备更新资源（相同ID，不同配置）
	updateRouteData := &model.GatewaySyncData{
		ID:        "existing-route-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"updated-route","uris":["/updated"]}`),
	}

	updateServiceData := &model.GatewaySyncData{
		ID:        "existing-service-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"updated-service"}`),
	}

	// 准备新增资源
	newRouteData := &model.GatewaySyncData{
		ID:        "new-route-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"new-route","uris":["/new"]}`),
	}

	newUpstreamData := &model.GatewaySyncData{
		ID:        "new-upstream-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"new-upstream","type":"roundrobin"}`),
	}

	// 构造上传资源参数
	addResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route:    {newRouteData},
		constant.Upstream: {newUpstreamData},
	}

	updateTypeResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route:   {updateRouteData},
		constant.Service: {updateServiceData},
	}

	// 执行上传
	err = UploadResources(
		gatewayCtx,
		addResourcesTypeMap,
		updateTypeResourcesTypeMap,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证更新后的资源
	updatedRoute, err := resourcebiz.GetRoute(gatewayCtx, "existing-route-id")
	assert.NoError(t, err)
	assert.Equal(t, "updated-route", updatedRoute.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedRoute.Status)

	updatedService, err := resourcebiz.GetService(gatewayCtx, "existing-service-id")
	assert.NoError(t, err)
	assert.Equal(t, "updated-service", updatedService.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedService.Status)

	// 验证新增的资源
	newRoute, err := resourcebiz.GetRoute(gatewayCtx, "new-route-id")
	assert.NoError(t, err)
	assert.Equal(t, "new-route", newRoute.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, newRoute.Status)

	newUpstream, err := resourcebiz.GetUpstream(gatewayCtx, "new-upstream-id")
	assert.NoError(t, err)
	assert.Equal(t, "new-upstream", newUpstream.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, newUpstream.Status)
}

// TestUploadResources_MixedResourceTypes 测试混合资源类型的上传
func TestUploadResources_MixedResourceTypes(t *testing.T) {
	// 准备多种资源类型的数据
	routeData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Route),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"mixed-route","uris":["/mixed"]}`),
	}

	serviceData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Service),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"mixed-service"}`),
	}

	upstreamData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Upstream),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"mixed-upstream","type":"roundrobin"}`),
	}

	consumerData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Consumer),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Consumer,
		Config:    datatypes.JSON(`{"username":"mixed-consumer"}`),
	}

	pluginConfigData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.PluginConfig),
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginConfig,
		Config:    datatypes.JSON(`{"name":"mixed-plugin-config"}`),
	}

	// 构造上传资源参数
	addResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route:        {routeData},
		constant.Service:      {serviceData},
		constant.Upstream:     {upstreamData},
		constant.Consumer:     {consumerData},
		constant.PluginConfig: {pluginConfigData},
	}

	// 执行上传
	err := UploadResources(
		gatewayCtx,
		addResourcesTypeMap,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证所有资源都被正确创建
	route, err := resourcebiz.GetRoute(gatewayCtx, routeData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-route", route.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, route.Status)

	service, err := resourcebiz.GetService(gatewayCtx, serviceData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-service", service.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, service.Status)

	upstream, err := resourcebiz.GetUpstream(gatewayCtx, upstreamData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-upstream", upstream.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, upstream.Status)

	consumer, err := resourcebiz.GetConsumer(gatewayCtx, consumerData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-consumer", consumer.Username)
	assert.Equal(t, constant.ResourceStatusCreateDraft, consumer.Status)

	pluginConfig, err := resourcebiz.GetPluginConfig(gatewayCtx, pluginConfigData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-plugin-config", pluginConfig.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, pluginConfig.Status)
}

// TestUploadResources_EmptyResources 测试空资源上传
func TestUploadResources_EmptyResources(t *testing.T) {
	// 测试空的新增资源
	err := UploadResources(
		gatewayCtx,
		nil,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 测试空的更新资源
	err = UploadResources(
		gatewayCtx,
		nil,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
}

// TestUploadResources_UpdateNonExistentResource 测试更新不存在的资源
func TestUploadResources_UpdateNonExistentResource(t *testing.T) {
	// 准备更新不存在的资源
	nonExistentRouteData := &model.GatewaySyncData{
		ID:        "non-existent-route-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"non-existent-route","uris":["/non-existent"]}`),
	}

	updateTypeResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {nonExistentRouteData},
	}

	// 执行上传（应该成功，因为会先删除再插入）
	err := UploadResources(
		gatewayCtx,
		nil,
		updateTypeResourcesTypeMap,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证资源被创建
	route, err := resourcebiz.GetRoute(gatewayCtx, "non-existent-route-id")
	assert.NoError(t, err)
	assert.Equal(t, "non-existent-route", route.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, route.Status)
}

// TestUploadResources_CrossGatewayIsolation 测试跨网关隔离
func TestUploadResources_CrossGatewayIsolation(t *testing.T) {
	// 创建第二个网关
	gateway2 := &model.Gateway{
		Name:          "gateway2-isolation",
		Mode:          1,
		Maintainers:   []string{"user1"},
		Desc:          "gateway2-isolation",
		APISIXType:    constant.APISIXTypeBKAPISIX,
		APISIXVersion: "3.11.0",
		EtcdConfig: model.EtcdConfig{
			InstanceID: "isolation-test",
			EtcdConfig: base.EtcdConfig{
				Endpoint: base.Endpoint(etcdEndpoint),
				Username: "test",
				Password: "test",
				Prefix:   "/apisix-isolation",
			},
		},
	}
	err := gatewaybiz.CreateGateway(context.Background(), gateway2)
	assert.NoError(t, err)
	gateway2Ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway2)

	// 在第一个网关中创建资源
	route1Data := &model.GatewaySyncData{
		ID:        "isolation-test-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"gateway1-route","uris":["/gateway1"]}`),
	}

	// 在第二个网关中创建相同ID的资源
	route2Data := &model.GatewaySyncData{
		ID:        "isolation-test-id",
		GatewayID: gateway2.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"gateway2-route","uris":["/gateway2"]}`),
	}

	// 分别上传到两个网关
	addResourcesTypeMap1 := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {route1Data},
	}
	err = UploadResources(
		gatewayCtx,
		addResourcesTypeMap1,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	addResourcesTypeMap2 := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {route2Data},
	}
	err = UploadResources(
		gateway2Ctx,
		addResourcesTypeMap2,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证两个网关的资源相互隔离
	route1, err := resourcebiz.GetRoute(gatewayCtx, "isolation-test-id")
	assert.NoError(t, err)
	assert.Equal(t, "gateway1-route", route1.Name)
	assert.Equal(t, gatewayInfo.ID, route1.GatewayID)

	route2, err := resourcebiz.GetRoute(gateway2Ctx, "isolation-test-id")
	assert.NoError(t, err)
	assert.Equal(t, "gateway2-route", route2.Name)
	assert.Equal(t, gateway2.ID, route2.GatewayID)

	// 验证在第一个网关中无法访问第二个网关的资源
	_, err = resourcebiz.GetRoute(gatewayCtx, "isolation-test-id")
	// 这里应该能访问到，因为ID相同但GatewayID不同，但GetRoute会通过GatewayID过滤
	// 所以实际上会返回第一个网关的资源
	assert.NoError(t, err)

	// 清理第二个网关
	err = gatewaybiz.DeleteGateway(context.Background(), gateway2)
	assert.NoError(t, err)
}

// TestBatchRevertRoutes_DeleteDraft 测试删除待发布状态的路由回滚
func TestBatchRevertRoutes_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的路由
	route := &model.Route{
		Name: "test-route-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Route),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-route-delete-draft","uris":["/test"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        route.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"test-route-delete-draft","uris":["/test"]}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新路由状态为删除待发布
	route.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateRoute(gatewayCtx, *route)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertRoutes(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证路由状态已恢复为成功状态
	revertedRoute, err := resourcebiz.GetRoute(gatewayCtx, route.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedRoute.Status)
	assert.Equal(t, "test-route-delete-draft", revertedRoute.Name)
}

// TestBatchRevertRoutes_UpdateDraft 测试更新待发布状态的路由回滚
func TestBatchRevertRoutes_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的路由
	originalRoute := &model.Route{
		Name: "original-route",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Route),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-route","uris":["/original"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateRoute(gatewayCtx, *originalRoute)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalRoute.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"original-route","uris":["/original"]}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新路由配置并设置为更新待发布状态，绑定关联ID
	originalRoute.Name = "updated-route-1"
	originalRoute.ServiceID = "service-1"
	originalRoute.UpstreamID = "upstream-1"
	originalRoute.Config = datatypes.JSON(
		`{"name":"updated-route-1","uris":["/updated"], "service_id":"service-1","upstream_id":"upstream-1"}`,
	)
	originalRoute.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateRoute(gatewayCtx, *originalRoute)
	assert.NoError(t, err)

	// 验证路由已被更新
	updatedRoute, err := resourcebiz.GetRoute(gatewayCtx, originalRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-route-1", updatedRoute.Name)
	assert.Equal(t, "service-1", updatedRoute.ServiceID)
	assert.Equal(t, "upstream-1", updatedRoute.UpstreamID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedRoute.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertRoutes(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证路由已回滚到原始配置，撤销后关联ID会为空
	revertedRoute, err := resourcebiz.GetRoute(gatewayCtx, originalRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-route", revertedRoute.Name)
	assert.Equal(t, "", revertedRoute.ServiceID)
	assert.Equal(t, "", revertedRoute.UpstreamID)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedRoute.Status)
}

// TestBatchRevertRoutes_MultipleRoutes 测试批量回滚多个路由
func TestBatchRevertRoutes_MultipleRoutes(t *testing.T) {
	// 创建第一个路由（删除待发布）
	route1 := &model.Route{
		Name: "route-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Route),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"route-1","uris":["/route1"]}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateRoute(gatewayCtx, *route1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        route1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"route-1","uris":["/route1"]}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个路由（更新待发布）
	route2 := &model.Route{
		Name: "route-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Route),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"route-2-updated","uris":["/route2-updated"]}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateRoute(gatewayCtx, *route2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        route2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"route-2-original","uris":["/route2"]}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertRoutes(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个路由状态恢复
	revertedRoute1, err := resourcebiz.GetRoute(gatewayCtx, route1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedRoute1.Status)

	// 验证第二个路由配置已回滚
	revertedRoute2, err := resourcebiz.GetRoute(gatewayCtx, route2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "route-2-original", revertedRoute2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedRoute2.Status)
}

// TestBatchRevertServices_DeleteDraft 测试删除待发布状态的服务回滚
func TestBatchRevertServices_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的服务
	service := &model.Service{
		Name: "test-service-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Service),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-service-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateService(gatewayCtx, *service)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        service.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"test-service-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新服务状态为删除待发布
	service.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateService(gatewayCtx, *service)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertServices(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证服务状态已恢复为成功状态
	revertedService, err := resourcebiz.GetService(gatewayCtx, service.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedService.Status)
	assert.Equal(t, "test-service-delete-draft", revertedService.Name)
}

// TestBatchRevertServices_UpdateDraft 测试更新待发布状态的服务回滚
func TestBatchRevertServices_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的服务
	originalService := &model.Service{
		Name: "original-service",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Service),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-service"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateService(gatewayCtx, *originalService)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalService.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"original-service"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新服务配置并设置为更新待发布状态，绑定关联ID
	originalService.Name = "updated-service-1"
	originalService.UpstreamID = "upstream-1"
	originalService.Config = datatypes.JSON(`{"name":"updated-service-1","upstream_id":"upstream-1"}`)
	originalService.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateService(gatewayCtx, *originalService)
	assert.NoError(t, err)

	// 验证服务已被更新
	updatedService, err := resourcebiz.GetService(gatewayCtx, originalService.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-service-1", updatedService.Name)
	assert.Equal(t, "upstream-1", updatedService.UpstreamID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedService.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertServices(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证服务已回滚到原始配置，撤销后关联ID会为空
	revertedService, err := resourcebiz.GetService(gatewayCtx, originalService.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-service", revertedService.Name)
	assert.Equal(t, "", revertedService.UpstreamID)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedService.Status)
}

// TestBatchRevertServices_MultipleServices 测试批量回滚多个服务
func TestBatchRevertServices_MultipleServices(t *testing.T) {
	// 创建第一个服务（删除待发布）
	service1 := &model.Service{
		Name: "service-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Service),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"service-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateService(gatewayCtx, *service1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        service1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"service-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个服务（更新待发布）
	service2 := &model.Service{
		Name: "service-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Service),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"service-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateService(gatewayCtx, *service2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        service2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"service-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertServices(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个服务状态恢复
	revertedService1, err := resourcebiz.GetService(gatewayCtx, service1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedService1.Status)

	// 验证第二个服务配置已回滚
	revertedService2, err := resourcebiz.GetService(gatewayCtx, service2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "service-2-original", revertedService2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedService2.Status)
}

// TestBatchRevertUpstreams_DeleteDraft 测试删除待发布状态的上游回滚
func TestBatchRevertUpstreams_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的上游
	upstream := &model.Upstream{
		Name: "test-upstream-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Upstream),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-upstream-delete-draft","type":"roundrobin"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateUpstream(gatewayCtx, *upstream)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        upstream.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"test-upstream-delete-draft","type":"roundrobin"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新上游状态为删除待发布
	upstream.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateUpstream(gatewayCtx, *upstream)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertUpstreams(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证上游状态已恢复为成功状态
	revertedUpstream, err := resourcebiz.GetUpstream(gatewayCtx, upstream.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedUpstream.Status)
	assert.Equal(t, "test-upstream-delete-draft", revertedUpstream.Name)
}

// TestBatchRevertUpstreams_UpdateDraft 测试更新待发布状态的上游回滚
func TestBatchRevertUpstreams_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的上游
	originalUpstream := &model.Upstream{
		Name: "original-upstream",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Upstream),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-upstream","type":"roundrobin"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateUpstream(gatewayCtx, *originalUpstream)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalUpstream.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"original-upstream","type":"roundrobin"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新上游配置并设置为更新待发布状态，绑定关联ID
	originalUpstream.Name = "updated-upstream-1"
	originalUpstream.SSLID = "ssl-1"
	originalUpstream.Config = datatypes.JSON(`{"name":"updated-upstream-1","type":"roundrobin","ssl_id":"ssl-1"}`)
	originalUpstream.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateUpstream(gatewayCtx, *originalUpstream)
	assert.NoError(t, err)

	// 验证上游已被更新
	updatedUpstream, err := resourcebiz.GetUpstream(gatewayCtx, originalUpstream.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-upstream-1", updatedUpstream.Name)
	assert.Equal(t, "ssl-1", updatedUpstream.SSLID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedUpstream.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertUpstreams(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证上游已回滚到原始配置，撤销后关联ID会为空
	revertedUpstream, err := resourcebiz.GetUpstream(gatewayCtx, originalUpstream.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-upstream", revertedUpstream.Name)
	assert.Equal(t, "", revertedUpstream.SSLID)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedUpstream.Status)
}

// TestBatchRevertUpstreams_MultipleUpstreams 测试批量回滚多个上游
func TestBatchRevertUpstreams_MultipleUpstreams(t *testing.T) {
	// 创建第一个上游（删除待发布）
	upstream1 := &model.Upstream{
		Name: "upstream-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Upstream),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"upstream-1","type":"roundrobin"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateUpstream(gatewayCtx, *upstream1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        upstream1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"upstream-1","type":"roundrobin"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个上游（更新待发布）
	upstream2 := &model.Upstream{
		Name: "upstream-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Upstream),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"upstream-2-updated","type":"roundrobin"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateUpstream(gatewayCtx, *upstream2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        upstream2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"upstream-2-original","type":"roundrobin"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertUpstreams(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个上游状态恢复
	revertedUpstream1, err := resourcebiz.GetUpstream(gatewayCtx, upstream1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedUpstream1.Status)

	// 验证第二个上游配置已回滚
	revertedUpstream2, err := resourcebiz.GetUpstream(gatewayCtx, upstream2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "upstream-2-original", revertedUpstream2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedUpstream2.Status)
}

// TestBatchRevertConsumerGroups_DeleteDraft 测试删除待发布状态的消费者组回滚
func TestBatchRevertConsumerGroups_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的消费者组
	consumerGroup := &model.ConsumerGroup{
		Name: "test-consumer-group-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.ConsumerGroup),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-consumer-group-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateConsumerGroup(gatewayCtx, *consumerGroup)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        consumerGroup.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.ConsumerGroup,
		Config:    datatypes.JSON(`{"name":"test-consumer-group-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新消费者组状态为删除待发布
	consumerGroup.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateConsumerGroup(gatewayCtx, *consumerGroup)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertConsumerGroups(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证消费者组状态已恢复为成功状态
	revertedConsumerGroup, err := resourcebiz.GetConsumerGroup(gatewayCtx, consumerGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumerGroup.Status)
	assert.Equal(t, "test-consumer-group-delete-draft", revertedConsumerGroup.Name)
}

// TestBatchRevertConsumerGroups_UpdateDraft 测试更新待发布状态的消费者组回滚
func TestBatchRevertConsumerGroups_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的消费者组
	originalConsumerGroup := &model.ConsumerGroup{
		Name: "original-consumer-group",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.ConsumerGroup),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-consumer-group"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateConsumerGroup(gatewayCtx, *originalConsumerGroup)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalConsumerGroup.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.ConsumerGroup,
		Config:    datatypes.JSON(`{"name":"original-consumer-group"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新消费者组配置并设置为更新待发布状态
	originalConsumerGroup.Name = "updated-consumer-group-1"
	originalConsumerGroup.Config = datatypes.JSON(`{"name":"updated-consumer-group-1"}`)
	originalConsumerGroup.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateConsumerGroup(gatewayCtx, *originalConsumerGroup)
	assert.NoError(t, err)

	// 验证消费者组已被更新
	updatedConsumerGroup, err := resourcebiz.GetConsumerGroup(gatewayCtx, originalConsumerGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-consumer-group-1", updatedConsumerGroup.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedConsumerGroup.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertConsumerGroups(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证消费者组已回滚到原始配置
	revertedConsumerGroup, err := resourcebiz.GetConsumerGroup(gatewayCtx, originalConsumerGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-consumer-group", revertedConsumerGroup.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumerGroup.Status)
}

// TestBatchRevertConsumerGroups_MultipleConsumerGroups 测试批量回滚多个消费者组
func TestBatchRevertConsumerGroups_MultipleConsumerGroups(t *testing.T) {
	// 创建第一个消费者组（删除待发布）
	consumerGroup1 := &model.ConsumerGroup{
		Name: "consumer-group-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.ConsumerGroup),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"consumer-group-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateConsumerGroup(gatewayCtx, *consumerGroup1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        consumerGroup1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.ConsumerGroup,
		Config:    datatypes.JSON(`{"name":"consumer-group-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个消费者组（更新待发布）
	consumerGroup2 := &model.ConsumerGroup{
		Name: "consumer-group-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.ConsumerGroup),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"consumer-group-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateConsumerGroup(gatewayCtx, *consumerGroup2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        consumerGroup2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.ConsumerGroup,
		Config:    datatypes.JSON(`{"name":"consumer-group-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertConsumerGroups(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个消费者组状态恢复
	revertedConsumerGroup1, err := resourcebiz.GetConsumerGroup(gatewayCtx, consumerGroup1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumerGroup1.Status)

	// 验证第二个消费者组配置已回滚
	revertedConsumerGroup2, err := resourcebiz.GetConsumerGroup(gatewayCtx, consumerGroup2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "consumer-group-2-original", revertedConsumerGroup2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumerGroup2.Status)
}

// TestBatchRevertPluginConfigs_DeleteDraft 测试删除待发布状态的插件配置回滚
func TestBatchRevertPluginConfigs_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的插件配置
	pluginConfig := &model.PluginConfig{
		Name: "test-plugin-config-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginConfig),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-plugin-config-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreatePluginConfig(gatewayCtx, *pluginConfig)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        pluginConfig.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginConfig,
		Config:    datatypes.JSON(`{"name":"test-plugin-config-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新插件配置状态为删除待发布
	pluginConfig.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdatePluginConfig(gatewayCtx, *pluginConfig)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertPluginConfigs(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证插件配置状态已恢复为成功状态
	revertedPluginConfig, err := resourcebiz.GetPluginConfig(gatewayCtx, pluginConfig.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginConfig.Status)
	assert.Equal(t, "test-plugin-config-delete-draft", revertedPluginConfig.Name)
}

// TestBatchRevertPluginConfigs_UpdateDraft 测试更新待发布状态的插件配置回滚
func TestBatchRevertPluginConfigs_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的插件配置
	originalPluginConfig := &model.PluginConfig{
		Name: "original-plugin-config",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginConfig),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-plugin-config"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreatePluginConfig(gatewayCtx, *originalPluginConfig)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalPluginConfig.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginConfig,
		Config:    datatypes.JSON(`{"name":"original-plugin-config"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新插件配置并设置为更新待发布状态
	originalPluginConfig.Name = "updated-plugin-config-1"
	originalPluginConfig.Config = datatypes.JSON(`{"name":"updated-plugin-config-1"}`)
	originalPluginConfig.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdatePluginConfig(gatewayCtx, *originalPluginConfig)
	assert.NoError(t, err)

	// 验证插件配置已被更新
	updatedPluginConfig, err := resourcebiz.GetPluginConfig(gatewayCtx, originalPluginConfig.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-plugin-config-1", updatedPluginConfig.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedPluginConfig.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertPluginConfigs(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证插件配置已回滚到原始配置
	revertedPluginConfig, err := resourcebiz.GetPluginConfig(gatewayCtx, originalPluginConfig.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-plugin-config", revertedPluginConfig.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginConfig.Status)
}

// TestBatchRevertPluginConfigs_MultiplePluginConfigs 测试批量回滚多个插件配置
func TestBatchRevertPluginConfigs_MultiplePluginConfigs(t *testing.T) {
	// 创建第一个插件配置（删除待发布）
	pluginConfig1 := &model.PluginConfig{
		Name: "plugin-config-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginConfig),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"plugin-config-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreatePluginConfig(gatewayCtx, *pluginConfig1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        pluginConfig1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginConfig,
		Config:    datatypes.JSON(`{"name":"plugin-config-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个插件配置（更新待发布）
	pluginConfig2 := &model.PluginConfig{
		Name: "plugin-config-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginConfig),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"plugin-config-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreatePluginConfig(gatewayCtx, *pluginConfig2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        pluginConfig2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginConfig,
		Config:    datatypes.JSON(`{"name":"plugin-config-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertPluginConfigs(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个插件配置状态恢复
	revertedPluginConfig1, err := resourcebiz.GetPluginConfig(gatewayCtx, pluginConfig1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginConfig1.Status)

	// 验证第二个插件配置已回滚
	revertedPluginConfig2, err := resourcebiz.GetPluginConfig(gatewayCtx, pluginConfig2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "plugin-config-2-original", revertedPluginConfig2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginConfig2.Status)
}

// TestBatchRevertPluginMetadatas_DeleteDraft 测试删除待发布状态的插件元数据回滚
func TestBatchRevertPluginMetadatas_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的插件元数据
	pluginMetadata := &model.PluginMetadata{
		Name: "test-plugin-metadata-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginMetadata),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-plugin-metadata-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreatePluginMetadata(gatewayCtx, *pluginMetadata)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        pluginMetadata.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginMetadata,
		Config:    datatypes.JSON(`{"name":"test-plugin-metadata-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新插件元数据状态为删除待发布
	pluginMetadata.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdatePluginMetadata(gatewayCtx, *pluginMetadata)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertPluginMetadatas(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证插件元数据状态已恢复为成功状态
	revertedPluginMetadata, err := resourcebiz.GetPluginMetadata(gatewayCtx, pluginMetadata.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginMetadata.Status)
	assert.Equal(t, "test-plugin-metadata-delete-draft", revertedPluginMetadata.Name)
}

// TestBatchRevertPluginMetadatas_UpdateDraft 测试更新待发布状态的插件元数据回滚
func TestBatchRevertPluginMetadatas_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的插件元数据
	originalPluginMetadata := &model.PluginMetadata{
		Name: "original-plugin-metadata",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginMetadata),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-plugin-metadata"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreatePluginMetadata(gatewayCtx, *originalPluginMetadata)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalPluginMetadata.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginMetadata,
		Config:    datatypes.JSON(`{"name":"original-plugin-metadata"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新插件元数据并设置为更新待发布状态
	originalPluginMetadata.Name = "updated-plugin-metadata-1"
	originalPluginMetadata.Config = datatypes.JSON(`{"name":"updated-plugin-metadata-1"}`)
	originalPluginMetadata.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdatePluginMetadata(gatewayCtx, *originalPluginMetadata)
	assert.NoError(t, err)

	// 验证插件元数据已被更新
	updatedPluginMetadata, err := resourcebiz.GetPluginMetadata(gatewayCtx, originalPluginMetadata.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-plugin-metadata-1", updatedPluginMetadata.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedPluginMetadata.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertPluginMetadatas(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证插件元数据已回滚到原始配置
	revertedPluginMetadata, err := resourcebiz.GetPluginMetadata(gatewayCtx, originalPluginMetadata.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-plugin-metadata", revertedPluginMetadata.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginMetadata.Status)
}

// TestBatchRevertPluginMetadatas_MultiplePluginMetadatas 测试批量回滚多个插件元数据
func TestBatchRevertPluginMetadatas_MultiplePluginMetadatas(t *testing.T) {
	// 创建第一个插件元数据（删除待发布）
	pluginMetadata1 := &model.PluginMetadata{
		Name: "plugin-metadata-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginMetadata),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"plugin-metadata-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreatePluginMetadata(gatewayCtx, *pluginMetadata1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        pluginMetadata1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginMetadata,
		Config:    datatypes.JSON(`{"name":"plugin-metadata-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个插件元数据（更新待发布）
	pluginMetadata2 := &model.PluginMetadata{
		Name: "plugin-metadata-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.PluginMetadata),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"plugin-metadata-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreatePluginMetadata(gatewayCtx, *pluginMetadata2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        pluginMetadata2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginMetadata,
		Config:    datatypes.JSON(`{"name":"plugin-metadata-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertPluginMetadatas(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个插件元数据状态恢复
	revertedPluginMetadata1, err := resourcebiz.GetPluginMetadata(gatewayCtx, pluginMetadata1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginMetadata1.Status)

	// 验证第二个插件元数据已回滚
	revertedPluginMetadata2, err := resourcebiz.GetPluginMetadata(gatewayCtx, pluginMetadata2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "plugin-metadata-2-original", revertedPluginMetadata2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedPluginMetadata2.Status)
}

// TestBatchRevertConsumers_DeleteDraft 测试删除待发布状态的消费者回滚
func TestBatchRevertConsumers_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的消费者
	consumer := &model.Consumer{
		Username: "test-consumer-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Consumer),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"username":"test-consumer-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateConsumer(gatewayCtx, *consumer)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        consumer.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Consumer,
		Config:    datatypes.JSON(`{"username":"test-consumer-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新消费者状态为删除待发布
	consumer.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateConsumer(gatewayCtx, *consumer)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertConsumers(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证消费者状态已恢复为成功状态
	revertedConsumer, err := resourcebiz.GetConsumer(gatewayCtx, consumer.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumer.Status)
	assert.Equal(t, "test-consumer-delete-draft", revertedConsumer.Username)
}

// TestBatchRevertConsumers_UpdateDraft 测试更新待发布状态的消费者回滚
func TestBatchRevertConsumers_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的消费者
	originalConsumer := &model.Consumer{
		Username: "original-consumer",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Consumer),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"username":"original-consumer"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateConsumer(gatewayCtx, *originalConsumer)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalConsumer.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Consumer,
		Config:    datatypes.JSON(`{"username":"original-consumer"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新消费者配置并设置为更新待发布状态，绑定关联ID
	originalConsumer.Username = "updated-consumer-1"
	originalConsumer.GroupID = "group-1"
	originalConsumer.Config = datatypes.JSON(`{"username":"updated-consumer-1","group_id":"group-1"}`)
	originalConsumer.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateConsumer(gatewayCtx, *originalConsumer)
	assert.NoError(t, err)

	// 验证消费者已被更新
	updatedConsumer, err := resourcebiz.GetConsumer(gatewayCtx, originalConsumer.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-consumer-1", updatedConsumer.Username)
	assert.Equal(t, "group-1", updatedConsumer.GroupID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedConsumer.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertConsumers(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证消费者已回滚到原始配置，撤销后关联ID会为空
	revertedConsumer, err := resourcebiz.GetConsumer(gatewayCtx, originalConsumer.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-consumer", revertedConsumer.Username)
	assert.Equal(t, "", revertedConsumer.GroupID)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumer.Status)
}

// TestBatchRevertConsumers_MultipleConsumers 测试批量回滚多个消费者
func TestBatchRevertConsumers_MultipleConsumers(t *testing.T) {
	// 创建第一个消费者（删除待发布）
	consumer1 := &model.Consumer{
		Username: "consumer-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Consumer),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"username":"consumer-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateConsumer(gatewayCtx, *consumer1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        consumer1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Consumer,
		Config:    datatypes.JSON(`{"username":"consumer-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个消费者（更新待发布）
	consumer2 := &model.Consumer{
		Username: "consumer-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Consumer),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"username":"consumer-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateConsumer(gatewayCtx, *consumer2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        consumer2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Consumer,
		Config:    datatypes.JSON(`{"username":"consumer-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertConsumers(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个消费者状态恢复
	revertedConsumer1, err := resourcebiz.GetConsumer(gatewayCtx, consumer1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumer1.Status)

	// 验证第二个消费者配置已回滚
	revertedConsumer2, err := resourcebiz.GetConsumer(gatewayCtx, consumer2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "consumer-2-original", revertedConsumer2.Username)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedConsumer2.Status)
}

// TestBatchRevertProtos_DeleteDraft 测试删除待发布状态的Proto回滚
func TestBatchRevertProtos_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的Proto
	proto := &model.Proto{
		Name: "test-proto-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Proto),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-proto-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateProto(gatewayCtx, *proto)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        proto.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Proto,
		Config:    datatypes.JSON(`{"name":"test-proto-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新Proto状态为删除待发布
	proto.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateProto(gatewayCtx, *proto)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertProtos(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证Proto状态已恢复为成功状态
	revertedProto, err := resourcebiz.GetProto(gatewayCtx, proto.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedProto.Status)
	assert.Equal(t, "test-proto-delete-draft", revertedProto.Name)
}

// TestBatchRevertProtos_UpdateDraft 测试更新待发布状态的Proto回滚
func TestBatchRevertProtos_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的Proto
	originalProto := &model.Proto{
		Name: "original-proto",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Proto),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-proto"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateProto(gatewayCtx, *originalProto)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalProto.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Proto,
		Config:    datatypes.JSON(`{"name":"original-proto"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新Proto配置并设置为更新待发布状态
	originalProto.Name = "updated-proto-1"
	originalProto.Config = datatypes.JSON(`{"name":"updated-proto-1"}`)
	originalProto.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateProto(gatewayCtx, *originalProto)
	assert.NoError(t, err)

	// 验证Proto已被更新
	updatedProto, err := resourcebiz.GetProto(gatewayCtx, originalProto.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-proto-1", updatedProto.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedProto.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertProtos(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证Proto已回滚到原始配置
	revertedProto, err := resourcebiz.GetProto(gatewayCtx, originalProto.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-proto", revertedProto.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedProto.Status)
}

// TestBatchRevertProtos_MultipleProtos 测试批量回滚多个Proto
func TestBatchRevertProtos_MultipleProtos(t *testing.T) {
	// 创建第一个Proto（删除待发布）
	proto1 := &model.Proto{
		Name: "proto-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Proto),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"proto-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateProto(gatewayCtx, *proto1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        proto1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Proto,
		Config:    datatypes.JSON(`{"name":"proto-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个Proto（更新待发布）
	proto2 := &model.Proto{
		Name: "proto-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.Proto),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"proto-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateProto(gatewayCtx, *proto2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        proto2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Proto,
		Config:    datatypes.JSON(`{"name":"proto-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertProtos(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个Proto状态恢复
	revertedProto1, err := resourcebiz.GetProto(gatewayCtx, proto1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedProto1.Status)

	// 验证第二个Proto配置已回滚
	revertedProto2, err := resourcebiz.GetProto(gatewayCtx, proto2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "proto-2-original", revertedProto2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedProto2.Status)
}

// 辅助函数：生成SSL配置JSON
func buildSSLConfig(name string) datatypes.JSON {
	return datatypes.JSON(fmt.Sprintf(`{"name":"%s","cert":"%s","key":"%s"}`,
		name, testSSLCert, testSSLKey))
}

// 辅助函数：创建SSL和同步数据
func createSSLWithSyncData(
	t *testing.T,
	ctx context.Context,
	name string,
	status constant.ResourceStatus,
) (*model.SSL, *model.GatewaySyncData) {
	ssl := &model.SSL{
		Name: name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.SSL),
			GatewayID: gatewayInfo.ID,
			Config:    buildSSLConfig(name),
			Status:    status,
		},
	}
	err := resourcebiz.CreateSSL(ctx, ssl)
	assert.NoError(t, err)

	syncData := &model.GatewaySyncData{
		ID:        ssl.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.SSL,
		Config:    buildSSLConfig(name),
	}
	err = repo.Q.GatewaySyncData.WithContext(ctx).Create(syncData)
	assert.NoError(t, err)

	return ssl, syncData
}

// TestBatchRevertSSLs_DeleteDraft 测试删除待发布状态的SSL回滚
func TestBatchRevertSSLs_DeleteDraft(t *testing.T) {
	// 创建SSL和同步数据
	ssl, syncData := createSSLWithSyncData(t, gatewayCtx, "test-ssl-delete-draft", constant.ResourceStatusSuccess)

	// 更新SSL状态为删除待发布
	ssl.Status = constant.ResourceStatusDeleteDraft
	err := resourcebiz.UpdateSSL(gatewayCtx, ssl)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertSSLs(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证SSL状态已恢复为成功状态
	revertedSSL, err := resourcebiz.GetSSL(gatewayCtx, ssl.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedSSL.Status)
	assert.Equal(t, "test-ssl-delete-draft", revertedSSL.Name)
}

// TestBatchRevertSSLs_UpdateDraft 测试更新待发布状态的SSL回滚
func TestBatchRevertSSLs_UpdateDraft(t *testing.T) {
	// 创建原始SSL和同步数据
	originalSSL, syncData := createSSLWithSyncData(t, gatewayCtx, "original-ssl", constant.ResourceStatusSuccess)

	// 更新SSL配置并设置为更新待发布状态
	originalSSL.Name = "updated-ssl-1"
	originalSSL.Config = buildSSLConfig("updated-ssl-1")
	originalSSL.Status = constant.ResourceStatusUpdateDraft
	err := resourcebiz.UpdateSSL(gatewayCtx, originalSSL)
	assert.NoError(t, err)

	// 验证SSL已被更新
	updatedSSL, err := resourcebiz.GetSSL(gatewayCtx, originalSSL.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-ssl-1", updatedSSL.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedSSL.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertSSLs(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证SSL已回滚到原始配置
	revertedSSL, err := resourcebiz.GetSSL(gatewayCtx, originalSSL.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-ssl", revertedSSL.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedSSL.Status)
}

// TestBatchRevertSSLs_MultipleSSLs 测试批量回滚多个SSL
func TestBatchRevertSSLs_MultipleSSLs(t *testing.T) {
	// 创建第一个SSL（删除待发布）
	ssl1, syncData1 := createSSLWithSyncData(t, gatewayCtx, "ssl-1", constant.ResourceStatusDeleteDraft)

	// 创建第二个SSL（更新待发布）
	ssl2, syncData2 := createSSLWithSyncData(t, gatewayCtx, "ssl-2-original", constant.ResourceStatusSuccess)

	// 更新第二个SSL
	ssl2.Name = "ssl-2-updated"
	ssl2.Config = buildSSLConfig("ssl-2-updated")
	ssl2.Status = constant.ResourceStatusUpdateDraft
	err := resourcebiz.UpdateSSL(gatewayCtx, ssl2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertSSLs(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个SSL状态恢复
	revertedSSL1, err := resourcebiz.GetSSL(gatewayCtx, ssl1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedSSL1.Status)

	// 验证第二个SSL配置已回滚
	revertedSSL2, err := resourcebiz.GetSSL(gatewayCtx, ssl2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "ssl-2-original", revertedSSL2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedSSL2.Status)
}

// TestBatchRevertGlobalRules_DeleteDraft 测试删除待发布状态的GlobalRule回滚
func TestBatchRevertGlobalRules_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的GlobalRule
	globalRule := &model.GlobalRule{
		Name: "test-global-rule-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.GlobalRule),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-global-rule-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateGlobalRule(gatewayCtx, *globalRule)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        globalRule.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.GlobalRule,
		Config:    datatypes.JSON(`{"name":"test-global-rule-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新GlobalRule状态为删除待发布
	globalRule.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateGlobalRule(gatewayCtx, *globalRule)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertGlobalRules(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证GlobalRule状态已恢复为成功状态
	revertedGlobalRule, err := resourcebiz.GetGlobalRule(gatewayCtx, globalRule.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedGlobalRule.Status)
	assert.Equal(t, "test-global-rule-delete-draft", revertedGlobalRule.Name)
}

// TestBatchRevertGlobalRules_UpdateDraft 测试更新待发布状态的GlobalRule回滚
func TestBatchRevertGlobalRules_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的GlobalRule
	originalGlobalRule := &model.GlobalRule{
		Name: "original-global-rule",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.GlobalRule),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-global-rule"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateGlobalRule(gatewayCtx, *originalGlobalRule)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalGlobalRule.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.GlobalRule,
		Config:    datatypes.JSON(`{"name":"original-global-rule"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新GlobalRule配置并设置为更新待发布状态
	originalGlobalRule.Name = "updated-global-rule-1"
	originalGlobalRule.Config = datatypes.JSON(`{"name":"updated-global-rule-1"}`)
	originalGlobalRule.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateGlobalRule(gatewayCtx, *originalGlobalRule)
	assert.NoError(t, err)

	// 验证GlobalRule已被更新
	updatedGlobalRule, err := resourcebiz.GetGlobalRule(gatewayCtx, originalGlobalRule.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-global-rule-1", updatedGlobalRule.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedGlobalRule.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertGlobalRules(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证GlobalRule已回滚到原始配置
	revertedGlobalRule, err := resourcebiz.GetGlobalRule(gatewayCtx, originalGlobalRule.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-global-rule", revertedGlobalRule.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedGlobalRule.Status)
}

// TestBatchRevertGlobalRules_MultipleGlobalRules 测试批量回滚多个GlobalRule
func TestBatchRevertGlobalRules_MultipleGlobalRules(t *testing.T) {
	// 创建第一个GlobalRule（删除待发布）
	globalRule1 := &model.GlobalRule{
		Name: "global-rule-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.GlobalRule),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"global-rule-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateGlobalRule(gatewayCtx, *globalRule1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        globalRule1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.GlobalRule,
		Config:    datatypes.JSON(`{"name":"global-rule-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个GlobalRule（更新待发布）
	globalRule2 := &model.GlobalRule{
		Name: "global-rule-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.GlobalRule),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"global-rule-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateGlobalRule(gatewayCtx, *globalRule2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        globalRule2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.GlobalRule,
		Config:    datatypes.JSON(`{"name":"global-rule-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertGlobalRules(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个GlobalRule状态恢复
	revertedGlobalRule1, err := resourcebiz.GetGlobalRule(gatewayCtx, globalRule1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedGlobalRule1.Status)

	// 验证第二个GlobalRule配置已回滚
	revertedGlobalRule2, err := resourcebiz.GetGlobalRule(gatewayCtx, globalRule2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "global-rule-2-original", revertedGlobalRule2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedGlobalRule2.Status)
}

// TestBatchRevertStreamRoutes_DeleteDraft 测试删除待发布状态的StreamRoute回滚
func TestBatchRevertStreamRoutes_DeleteDraft(t *testing.T) {
	// 创建一个成功状态的StreamRoute
	streamRoute := &model.StreamRoute{
		Name: "test-stream-route-delete-draft",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.StreamRoute),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"test-stream-route-delete-draft"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateStreamRoute(gatewayCtx, *streamRoute)
	assert.NoError(t, err)

	// 创建对应的同步数据
	syncData := &model.GatewaySyncData{
		ID:        streamRoute.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.StreamRoute,
		Config:    datatypes.JSON(`{"name":"test-stream-route-delete-draft"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新StreamRoute状态为删除待发布
	streamRoute.Status = constant.ResourceStatusDeleteDraft
	err = resourcebiz.UpdateStreamRoute(gatewayCtx, *streamRoute)
	assert.NoError(t, err)

	// 执行回滚
	err = resourcebiz.BatchRevertStreamRoutes(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证StreamRoute状态已恢复为成功状态
	revertedStreamRoute, err := resourcebiz.GetStreamRoute(gatewayCtx, streamRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedStreamRoute.Status)
	assert.Equal(t, "test-stream-route-delete-draft", revertedStreamRoute.Name)
}

// TestBatchRevertStreamRoutes_UpdateDraft 测试更新待发布状态的StreamRoute回滚
func TestBatchRevertStreamRoutes_UpdateDraft(t *testing.T) {
	// 创建一个成功状态的StreamRoute
	originalStreamRoute := &model.StreamRoute{
		Name: "original-stream-route",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.StreamRoute),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"original-stream-route"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := resourcebiz.CreateStreamRoute(gatewayCtx, *originalStreamRoute)
	assert.NoError(t, err)

	// 创建对应的同步数据（etcd中的原始数据）
	syncData := &model.GatewaySyncData{
		ID:        originalStreamRoute.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.StreamRoute,
		Config:    datatypes.JSON(`{"name":"original-stream-route"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData)
	assert.NoError(t, err)

	// 更新StreamRoute配置并设置为更新待发布状态，绑定关联ID
	originalStreamRoute.Name = "updated-stream-route-1"
	originalStreamRoute.ServiceID = "service-1"
	originalStreamRoute.UpstreamID = "upstream-1"
	originalStreamRoute.Config = datatypes.JSON(
		`{"name":"updated-stream-route-1","service_id":"service-1","upstream_id":"upstream-1"}`,
	)
	originalStreamRoute.Status = constant.ResourceStatusUpdateDraft
	err = resourcebiz.UpdateStreamRoute(gatewayCtx, *originalStreamRoute)
	assert.NoError(t, err)

	// 验证StreamRoute已被更新
	updatedStreamRoute, err := resourcebiz.GetStreamRoute(gatewayCtx, originalStreamRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-stream-route-1", updatedStreamRoute.Name)
	assert.Equal(t, "service-1", updatedStreamRoute.ServiceID)
	assert.Equal(t, "upstream-1", updatedStreamRoute.UpstreamID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedStreamRoute.Status)

	// 执行回滚
	err = resourcebiz.BatchRevertStreamRoutes(gatewayCtx, []*model.GatewaySyncData{syncData})
	assert.NoError(t, err)

	// 验证StreamRoute已回滚到原始配置，撤销后关联ID会为空
	revertedStreamRoute, err := resourcebiz.GetStreamRoute(gatewayCtx, originalStreamRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, "original-stream-route", revertedStreamRoute.Name)
	assert.Equal(t, "", revertedStreamRoute.ServiceID)
	assert.Equal(t, "", revertedStreamRoute.UpstreamID)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedStreamRoute.Status)
}

// TestBatchRevertStreamRoutes_MultipleStreamRoutes 测试批量回滚多个StreamRoute
func TestBatchRevertStreamRoutes_MultipleStreamRoutes(t *testing.T) {
	// 创建第一个StreamRoute（删除待发布）
	streamRoute1 := &model.StreamRoute{
		Name: "stream-route-1",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.StreamRoute),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"stream-route-1"}`),
			Status:    constant.ResourceStatusDeleteDraft,
		},
	}
	err := resourcebiz.CreateStreamRoute(gatewayCtx, *streamRoute1)
	assert.NoError(t, err)

	syncData1 := &model.GatewaySyncData{
		ID:        streamRoute1.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.StreamRoute,
		Config:    datatypes.JSON(`{"name":"stream-route-1"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData1)
	assert.NoError(t, err)

	// 创建第二个StreamRoute（更新待发布）
	streamRoute2 := &model.StreamRoute{
		Name: "stream-route-2-updated",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.StreamRoute),
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"stream-route-2-updated"}`),
			Status:    constant.ResourceStatusUpdateDraft,
		},
	}
	err = resourcebiz.CreateStreamRoute(gatewayCtx, *streamRoute2)
	assert.NoError(t, err)

	syncData2 := &model.GatewaySyncData{
		ID:        streamRoute2.ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.StreamRoute,
		Config:    datatypes.JSON(`{"name":"stream-route-2-original"}`),
	}
	err = repo.Q.GatewaySyncData.WithContext(gatewayCtx).Create(syncData2)
	assert.NoError(t, err)

	// 批量回滚
	err = resourcebiz.BatchRevertStreamRoutes(gatewayCtx, []*model.GatewaySyncData{syncData1, syncData2})
	assert.NoError(t, err)

	// 验证第一个StreamRoute状态恢复
	revertedStreamRoute1, err := resourcebiz.GetStreamRoute(gatewayCtx, streamRoute1.ID)
	assert.NoError(t, err)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedStreamRoute1.Status)

	// 验证第二个StreamRoute配置已回滚
	revertedStreamRoute2, err := resourcebiz.GetStreamRoute(gatewayCtx, streamRoute2.ID)
	assert.NoError(t, err)
	assert.Equal(t, "stream-route-2-original", revertedStreamRoute2.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, revertedStreamRoute2.Status)
}
