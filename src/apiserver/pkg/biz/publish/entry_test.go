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

package publish

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	diffbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/diff"
	gatewaybiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/gateway"
	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	syncdatabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/syncdata"
	unifyopbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/unifyop"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

var (
	gatewayInfo  *model.Gateway
	gatewayCtx   context.Context
	etcdEndpoint string
)

func TestMain(m *testing.M) {
	// init crypto
	err := cryptography.Init("jxi18GX5w2qgHwfZCFpn07q8FScXJOd3", "k2dbCGetyusW")
	if err != nil {
		panic(err)
	}

	// 初始化embed数据库
	util.InitEmbedDb()

	_, server, endpoint, err := util.StartEmbedEtcdClientRandom(context.Background())
	if err != nil {
		panic(err)
	}
	etcdEndpoint = endpoint

	// 初始化网关数据
	CreatGateway()

	// 执行所有测试用例
	code := m.Run()

	// 关闭etcd server
	server.Close()

	// 退出时返回测试状态码
	os.Exit(code)
}

// 创建网关资源
func CreatGateway() {
	gatewayInfo = data.Gateway1WithBkAPISIX()
	gatewayInfo.EtcdConfig.Endpoint = base.Endpoint(etcdEndpoint)
	err := gatewaybiz.CreateGateway(context.Background(), gatewayInfo)
	if err != nil {
		panic(err)
	}
	gatewayCtx = ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
}

func publishTestName(t *testing.T, suffix string) string {
	t.Helper()
	name := strings.ToLower(t.Name() + "-" + suffix)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	if len(name) > 48 {
		name = name[:24] + "-" + name[len(name)-23:]
	}
	return name
}

func newPublishGatewayContext(t *testing.T, version string) (*model.Gateway, context.Context) {
	t.Helper()

	gateway := data.Gateway1WithBkAPISIX()
	gateway.Name = publishTestName(t, strings.ReplaceAll(version, ".", "-"))
	gateway.Desc = gateway.Name
	gateway.APISIXVersion = version
	gateway.EtcdConfig.Endpoint = base.Endpoint(etcdEndpoint)
	gateway.EtcdConfig.Prefix = "/" + publishTestName(t, "etcd")
	if err := gatewaybiz.CreateGateway(context.Background(), gateway); err != nil {
		t.Fatal(err)
	}
	return gateway, ginx.SetGatewayInfoToContext(context.Background(), gateway)
}

func mustSyncAndGetSyncedItem(
	t *testing.T,
	ctx context.Context,
	resourceType constant.APISIXResource,
	id string,
) *model.GatewaySyncData {
	t.Helper()

	if _, err := unifyopbiz.SyncResources(ctx, resourceType); err != nil {
		t.Fatal(err)
	}
	syncedItem, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(ctx, resourceType, id)
	if err != nil {
		t.Fatal(err)
	}
	return syncedItem
}

func mustPublishResource(
	t *testing.T,
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceID string,
) {
	t.Helper()

	if err := PublishResource(ctx, resourceType, []string{resourceID}); err != nil {
		t.Fatal(err)
	}
}

type publishResourceEntryTestCase struct {
	name         string
	resourceType constant.APISIXResource
	create       func(context.Context, *model.Gateway, constant.ResourceStatus) (string, error)
}

func TestPublishPayloadCharacterization_CurrentSeams(t *testing.T) {
	t.Run("route keeps id and name in final payload", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)

		if err := resourcebiz.CreateRoute(ctx, *route); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.Route, route.ID)

		syncedRoute := mustSyncAndGetSyncedItem(t, ctx, constant.Route, route.ID)
		assert.Equal(t, route.ID, gjson.GetBytes(syncedRoute.Config, "id").String())
		assert.Equal(t, route.Name, gjson.GetBytes(syncedRoute.Config, "name").String())
	})

	t.Run("service keeps id in final payload", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		service := data.Service1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)

		if err := resourcebiz.CreateService(ctx, *service); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.Service, service.ID)

		syncedService := mustSyncAndGetSyncedItem(t, ctx, constant.Service, service.ID)
		assert.Equal(t, service.ID, gjson.GetBytes(syncedService.Config, "id").String())
		assert.Equal(t, service.Name, gjson.GetBytes(syncedService.Config, "name").String())
	})

	t.Run("upstream keeps id in final payload", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)

		if err := resourcebiz.CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.Upstream, upstream.ID)

		syncedUpstream := mustSyncAndGetSyncedItem(t, ctx, constant.Upstream, upstream.ID)
		assert.Equal(t, upstream.ID, gjson.GetBytes(syncedUpstream.Config, "id").String())
		assert.Equal(t, upstream.Name, gjson.GetBytes(syncedUpstream.Config, "name").String())
	})

	t.Run("consumer removes id and keeps username", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		consumer := data.Consumer1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)

		if err := resourcebiz.CreateConsumer(ctx, *consumer); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.Consumer, consumer.ID)

		syncedConsumer := mustSyncAndGetSyncedItem(t, ctx, constant.Consumer, consumer.ID)
		assert.False(t, gjson.GetBytes(syncedConsumer.Config, "id").Exists())
		assert.Equal(t, consumer.Username, gjson.GetBytes(syncedConsumer.Config, "username").String())
	})

	t.Run("consumer group synced payload keeps name across versions", func(t *testing.T) {
		testCases := []struct {
			name     string
			version  string
			wantName bool
		}{
			{name: "3.11 keeps name", version: "3.11.0", wantName: true},
			{name: "3.13 keeps name", version: "3.13.0", wantName: true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				gateway, ctx := newPublishGatewayContext(t, tc.version)

				group := data.ConsumerGroup1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)

				if err := resourcebiz.CreateConsumerGroup(ctx, *group); err != nil {
					t.Fatal(err)
				}
				mustPublishResource(t, ctx, constant.ConsumerGroup, group.ID)

				syncedGroup := mustSyncAndGetSyncedItem(t, ctx, constant.ConsumerGroup, group.ID)
				assert.Equal(t, tc.wantName, gjson.GetBytes(syncedGroup.Config, "name").Exists())
			})
		}
	})

	t.Run("ssl removes internal validity fields", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		ssl := data.SSL1(gateway, constant.ResourceStatusCreateDraft)
		var err error
		ssl.Config, err = sjson.SetBytes(ssl.Config, "validity_start", 111)
		if err != nil {
			t.Fatal(err)
		}
		ssl.Config, err = sjson.SetBytes(ssl.Config, "validity_end", 222)
		if err != nil {
			t.Fatal(err)
		}

		if err := resourcebiz.CreateSSL(ctx, ssl); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.SSL, ssl.ID)

		syncedSSL := mustSyncAndGetSyncedItem(t, ctx, constant.SSL, ssl.ID)
		assert.False(t, gjson.GetBytes(syncedSSL.Config, "validity_start").Exists())
		assert.False(t, gjson.GetBytes(syncedSSL.Config, "validity_end").Exists())
	})

	t.Run("stream route synced payload keeps labels", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		streamRoute := data.StreamRoute1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
		var err error
		streamRoute.Config, err = sjson.SetBytes(streamRoute.Config, "labels", map[string]string{"env": "test"})
		if err != nil {
			t.Fatal(err)
		}

		if err := resourcebiz.CreateStreamRoute(ctx, *streamRoute); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.StreamRoute, streamRoute.ID)

		syncedStreamRoute := mustSyncAndGetSyncedItem(t, ctx, constant.StreamRoute, streamRoute.ID)
		assert.Equal(t, "test", gjson.GetBytes(syncedStreamRoute.Config, "labels.env").String())
	})
}

func TestPublishDependencyFanout_CurrentSeams(t *testing.T) {
	t.Run("route publishes upstream service and plugin config dependencies", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}

		service := data.Service1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		service.UpstreamID = upstream.ID
		if err := resourcebiz.CreateService(ctx, *service); err != nil {
			t.Fatal(err)
		}

		pluginConfig := data.PluginConfig1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreatePluginConfig(ctx, *pluginConfig); err != nil {
			t.Fatal(err)
		}

		route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
		route.ServiceID = service.ID
		route.UpstreamID = upstream.ID
		route.PluginConfigID = pluginConfig.ID
		route.Config = datatypes.JSON(`{"uris":["/route-dependency"],"methods":["GET"]}`)
		if err := resourcebiz.CreateRoute(ctx, *route); err != nil {
			t.Fatal(err)
		}

		mustPublishResource(t, ctx, constant.Route, route.ID)

		assert.Equal(t, route.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Route, route.ID).ID)
		assert.Equal(t, upstream.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Upstream, upstream.ID).ID)
		assert.Equal(t, service.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Service, service.ID).ID)
		assert.Equal(
			t,
			pluginConfig.ID,
			mustSyncAndGetSyncedItem(t, ctx, constant.PluginConfig, pluginConfig.ID).ID,
		)
	})

	t.Run("consumer publishes consumer group dependency", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		group := data.ConsumerGroup1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreateConsumerGroup(ctx, *group); err != nil {
			t.Fatal(err)
		}

		consumer := data.Consumer1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		consumer.GroupID = group.ID
		if err := resourcebiz.CreateConsumer(ctx, *consumer); err != nil {
			t.Fatal(err)
		}

		mustPublishResource(t, ctx, constant.Consumer, consumer.ID)

		assert.Equal(t, consumer.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Consumer, consumer.ID).ID)
		assert.Equal(t, group.ID, mustSyncAndGetSyncedItem(t, ctx, constant.ConsumerGroup, group.ID).ID)
	})

	t.Run("service publishes upstream dependency", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}

		service := data.Service1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		service.UpstreamID = upstream.ID
		if err := resourcebiz.CreateService(ctx, *service); err != nil {
			t.Fatal(err)
		}

		mustPublishResource(t, ctx, constant.Service, service.ID)

		assert.Equal(t, service.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Service, service.ID).ID)
		assert.Equal(t, upstream.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Upstream, upstream.ID).ID)
	})

	t.Run("upstream publishes ssl dependency", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		ssl := data.SSL1(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreateSSL(ctx, ssl); err != nil {
			t.Fatal(err)
		}

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		upstream.SSLID = ssl.ID
		if err := resourcebiz.CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}

		mustPublishResource(t, ctx, constant.Upstream, upstream.ID)

		assert.Equal(t, upstream.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Upstream, upstream.ID).ID)
		assert.Equal(t, ssl.ID, mustSyncAndGetSyncedItem(t, ctx, constant.SSL, ssl.ID).ID)
		storedSSL, err := resourcebiz.GetSSL(ctx, ssl.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, constant.ResourceStatusSuccess, storedSSL.Status)
	})

	t.Run("stream route publishes service and upstream dependencies", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}

		service := data.Service1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		service.UpstreamID = upstream.ID
		if err := resourcebiz.CreateService(ctx, *service); err != nil {
			t.Fatal(err)
		}

		streamRoute := data.StreamRoute1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
		streamRoute.ServiceID = service.ID
		streamRoute.UpstreamID = upstream.ID
		if err := resourcebiz.CreateStreamRoute(ctx, *streamRoute); err != nil {
			t.Fatal(err)
		}

		mustPublishResource(t, ctx, constant.StreamRoute, streamRoute.ID)

		assert.Equal(
			t,
			streamRoute.ID,
			mustSyncAndGetSyncedItem(t, ctx, constant.StreamRoute, streamRoute.ID).ID,
		)
		assert.Equal(t, upstream.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Upstream, upstream.ID).ID)
		assert.Equal(t, service.ID, mustSyncAndGetSyncedItem(t, ctx, constant.Service, service.ID).ID)
	})
}

func TestPublishPayloadFieldCleanup_CurrentSeams(t *testing.T) {
	t.Run("consumer synced payload removes id", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		consumer := data.Consumer1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		var err error
		consumer.Config, err = sjson.SetBytes(consumer.Config, "id", "should-disappear")
		if err != nil {
			t.Fatal(err)
		}

		if err := resourcebiz.CreateConsumer(ctx, *consumer); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.Consumer, consumer.ID)

		syncedConsumer := mustSyncAndGetSyncedItem(t, ctx, constant.Consumer, consumer.ID)
		assert.False(t, gjson.GetBytes(syncedConsumer.Config, "id").Exists())
		assert.Equal(t, consumer.Username, gjson.GetBytes(syncedConsumer.Config, "username").String())
	})

	t.Run("consumer group synced payload keeps name across versions", func(t *testing.T) {
		testCases := []struct {
			name    string
			version string
		}{
			{name: "3.11 keeps name", version: "3.11.0"},
			{name: "3.13 keeps name", version: "3.13.0"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				gateway, ctx := newPublishGatewayContext(t, tc.version)

				group := data.ConsumerGroup1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
				group.Name = "cg-demo"
				if err := resourcebiz.CreateConsumerGroup(ctx, *group); err != nil {
					t.Fatal(err)
				}
				mustPublishResource(t, ctx, constant.ConsumerGroup, group.ID)

				syncedGroup := mustSyncAndGetSyncedItem(t, ctx, constant.ConsumerGroup, group.ID)
				assert.Equal(t, "cg-demo", gjson.GetBytes(syncedGroup.Config, "name").String())
			})
		}
	})

	t.Run("ssl synced payload removes internal validity fields", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		ssl := data.SSL1(gateway, constant.ResourceStatusCreateDraft)
		var err error
		ssl.Config, err = sjson.SetBytes(ssl.Config, "validity_start", 1710000000)
		if err != nil {
			t.Fatal(err)
		}
		ssl.Config, err = sjson.SetBytes(ssl.Config, "validity_end", 1810000000)
		if err != nil {
			t.Fatal(err)
		}

		if err := resourcebiz.CreateSSL(ctx, ssl); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.SSL, ssl.ID)

		syncedSSL := mustSyncAndGetSyncedItem(t, ctx, constant.SSL, ssl.ID)
		assert.False(t, gjson.GetBytes(syncedSSL.Config, "validity_start").Exists())
		assert.False(t, gjson.GetBytes(syncedSSL.Config, "validity_end").Exists())
	})

	t.Run("stream route synced payload keeps labels", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		streamRoute := data.StreamRoute1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
		var err error
		streamRoute.Config, err = sjson.SetBytes(streamRoute.Config, "labels", map[string]string{"env": "prod"})
		if err != nil {
			t.Fatal(err)
		}

		if err := resourcebiz.CreateStreamRoute(ctx, *streamRoute); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.StreamRoute, streamRoute.ID)

		syncedStreamRoute := mustSyncAndGetSyncedItem(t, ctx, constant.StreamRoute, streamRoute.ID)
		assert.Equal(t, "prod", gjson.GetBytes(syncedStreamRoute.Config, "labels.env").String())
	})
}

func TestSimplePublishPayload_CurrentSeams(t *testing.T) {
	t.Run("plugin metadata uses plugin name as final payload id", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		pm := data.PluginMetadata1(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreatePluginMetadata(ctx, *pm); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.PluginMetadata, pm.ID)

		synced := mustSyncAndGetSyncedItem(t, ctx, constant.PluginMetadata, pm.ID)
		assert.Equal(t, pm.Name, gjson.GetBytes(synced.Config, "id").String())
		assert.Equal(t, pm.Name, gjson.GetBytes(synced.Config, "name").String())
	})

	t.Run("proto keeps name on 3.13", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.13.0")

		pb := data.Proto1(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreateProto(ctx, *pb); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.Proto, pb.ID)

		synced := mustSyncAndGetSyncedItem(t, ctx, constant.Proto, pb.ID)
		assert.Equal(t, pb.Name, gjson.GetBytes(synced.Config, "name").String())
		assert.Equal(t, pb.ID, gjson.GetBytes(synced.Config, "id").String())
	})
}

func TestPublishPersist_CurrentSeams(t *testing.T) {
	t.Run("plugin config publish still writes synced config and updates status", func(t *testing.T) {
		gateway, ctx := newPublishGatewayContext(t, "3.11.0")

		pluginConfig := data.PluginConfig1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := resourcebiz.CreatePluginConfig(ctx, *pluginConfig); err != nil {
			t.Fatal(err)
		}
		mustPublishResource(t, ctx, constant.PluginConfig, pluginConfig.ID)

		synced := mustSyncAndGetSyncedItem(t, ctx, constant.PluginConfig, pluginConfig.ID)
		assert.Equal(t, pluginConfig.ID, synced.ID)

		stored, err := resourcebiz.GetPluginConfig(ctx, pluginConfig.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, constant.ResourceStatusSuccess, stored.Status)
	})
}

func TestPublishResource_AllResourceTypes(t *testing.T) {
	testCases := []publishResourceEntryTestCase{
		{
			name:         "route",
			resourceType: constant.Route,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Route1WithNoRelationResource(gateway, status)
				return resource.ID, resourcebiz.CreateRoute(ctx, *resource)
			},
		},
		{
			name:         "service",
			resourceType: constant.Service,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Service1WithNoRelation(gateway, status)
				return resource.ID, resourcebiz.CreateService(ctx, *resource)
			},
		},
		{
			name:         "upstream",
			resourceType: constant.Upstream,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Upstream1WithNoRelation(gateway, status)
				return resource.ID, resourcebiz.CreateUpstream(ctx, *resource)
			},
		},
		{
			name:         "consumer",
			resourceType: constant.Consumer,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Consumer1WithNoRelation(gateway, status)
				return resource.ID, resourcebiz.CreateConsumer(ctx, *resource)
			},
		},
		{
			name:         "consumer_group",
			resourceType: constant.ConsumerGroup,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.ConsumerGroup1WithNoRelation(gateway, status)
				return resource.ID, resourcebiz.CreateConsumerGroup(ctx, *resource)
			},
		},
		{
			name:         "plugin_config",
			resourceType: constant.PluginConfig,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.PluginConfig1WithNoRelation(gateway, status)
				return resource.ID, resourcebiz.CreatePluginConfig(ctx, *resource)
			},
		},
		{
			name:         "global_rule",
			resourceType: constant.GlobalRule,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.GlobalRule1(gateway, status)
				return resource.ID, resourcebiz.CreateGlobalRule(ctx, *resource)
			},
		},
		{
			name:         "plugin_metadata",
			resourceType: constant.PluginMetadata,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.PluginMetadata1(gateway, status)
				return resource.ID, resourcebiz.CreatePluginMetadata(ctx, *resource)
			},
		},
		{
			name:         "proto",
			resourceType: constant.Proto,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Proto1(gateway, status)
				return resource.ID, resourcebiz.CreateProto(ctx, *resource)
			},
		},
		{
			name:         "ssl",
			resourceType: constant.SSL,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.SSL1(gateway, status)
				return resource.ID, resourcebiz.CreateSSL(ctx, resource)
			},
		},
		{
			name:         "stream_route",
			resourceType: constant.StreamRoute,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.StreamRoute1WithNoRelationResource(gateway, status)
				return resource.ID, resourcebiz.CreateStreamRoute(ctx, *resource)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gateway, ctx := newPublishGatewayContext(t, "3.11.0")

			resourceID, err := tc.create(ctx, gateway, constant.ResourceStatusCreateDraft)
			if err != nil {
				t.Fatal(err)
			}

			if err := PublishResource(ctx, tc.resourceType, []string{resourceID}); err != nil {
				t.Fatal(err)
			}

			synced := mustSyncAndGetSyncedItem(t, ctx, tc.resourceType, resourceID)
			assert.Equal(t, resourceID, synced.ID)

			diffResources, err := diffbiz.DiffResources(
				ctx,
				tc.resourceType,
				[]string{resourceID},
				"",
				[]constant.ResourceStatus{},
				false,
			)
			if err != nil {
				t.Fatal(err)
			}
			assert.Len(t, diffResources, 0)

			storedResources, err := resourcebiz.BatchGetResources(
				ctx,
				tc.resourceType,
				[]string{resourceID},
			)
			if err != nil {
				t.Fatal(err)
			}
			if assert.Len(t, storedResources, 1) {
				assert.Equal(t, constant.ResourceStatusSuccess, storedResources[0].Status)
			}
		})
	}
}

func TestPublishRoutes(t *testing.T) {
	type args struct {
		ctx   context.Context
		route *model.Route
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_route_without_related_resource",
			args: args{
				ctx: gatewayCtx,
				route: data.Route1WithNoRelationResource(
					gatewayInfo,
					constant.ResourceStatusCreateDraft,
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateRoute(tt.args.ctx, *tt.args.route); (err != nil) != tt.wantErr {
				t.Errorf("CreateRoute error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.Route, []string{tt.args.route.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.Route)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Route], 1)

			// assert sync resource
			syncedRoute, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Route,
				tt.args.route.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedRoute.ID, tt.args.route.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.Route, []string{tt.args.route.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			route, err := resourcebiz.GetRoute(tt.args.ctx, tt.args.route.ID)
			assert.NoError(t, err)

			assert.Equal(t, constant.ResourceStatusSuccess, route.Status)
		})
	}
}

func TestPublishService(t *testing.T) {
	type args struct {
		ctx     context.Context
		service *model.Service
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_service_without_related_resource",
			args: args{
				ctx:     gatewayCtx,
				service: data.Service1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateService(tt.args.ctx, *tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("CreateService error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.Service, []string{tt.args.service.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.Service)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Service], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Service,
				tt.args.service.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedResource.ID, tt.args.service.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.Service, []string{tt.args.service.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			service, err := resourcebiz.GetService(tt.args.ctx, tt.args.service.ID)
			assert.NoError(t, err)

			assert.Equal(t, constant.ResourceStatusSuccess, service.Status)
		})
	}
}

func TestPublishUpstreams(t *testing.T) {
	type args struct {
		ctx      context.Context
		upstream *model.Upstream
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_upstream_without_related_resource",
			args: args{
				ctx:      gatewayCtx,
				upstream: data.Upstream1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateUpstream(
				tt.args.ctx,
				*tt.args.upstream,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreateUpstream error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.Upstream, []string{tt.args.upstream.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.Upstream)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Upstream], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Upstream,
				tt.args.upstream.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedResource.ID, tt.args.upstream.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.Upstream, []string{tt.args.upstream.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			upstream, err := resourcebiz.GetUpstream(tt.args.ctx, tt.args.upstream.ID)
			assert.NoError(t, err)

			assert.Equal(t, constant.ResourceStatusSuccess, upstream.Status)
		})
	}
}

func TestPublishConsumer(t *testing.T) {
	type args struct {
		ctx      context.Context
		consumer *model.Consumer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_consumer_without_related_resource",
			args: args{
				ctx:      gatewayCtx,
				consumer: data.Consumer1WithNoRelation(gatewayInfo, constant.ResourceStatusCreateDraft),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateConsumer(
				tt.args.ctx,
				*tt.args.consumer,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreateConsumer error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.Consumer, []string{tt.args.consumer.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.Consumer)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Consumer], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Consumer,
				tt.args.consumer.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedResource.ID, tt.args.consumer.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.Consumer, []string{tt.args.consumer.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			consumer, err := resourcebiz.GetConsumer(tt.args.ctx, tt.args.consumer.ID)
			assert.NoError(t, err)

			assert.Equal(t, constant.ResourceStatusSuccess, consumer.Status)
		})
	}
}

func TestPublishPluginConfigs(t *testing.T) {
	type args struct {
		ctx          context.Context
		pluginConfig *model.PluginConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_plugin_config_without_related_resource",
			args: args{
				ctx: gatewayCtx,
				pluginConfig: data.PluginConfig1WithNoRelation(
					gatewayInfo,
					constant.ResourceStatusCreateDraft,
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreatePluginConfig(
				tt.args.ctx,
				*tt.args.pluginConfig,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreatePluginConfig error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.PluginConfig, []string{tt.args.pluginConfig.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.PluginConfig)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.PluginConfig], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.PluginConfig,
				tt.args.pluginConfig.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.pluginConfig.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(
				tt.args.ctx,
				constant.PluginConfig,
				[]string{tt.args.pluginConfig.ID},
				"",
				[]constant.ResourceStatus{},
				false,
			)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			pluginConfig, err := resourcebiz.GetPluginConfig(tt.args.ctx, tt.args.pluginConfig.ID)
			assert.NoError(t, err)
			assert.Equal(t, constant.ResourceStatusSuccess, pluginConfig.Status)
		})
	}
}

func TestPublishGlobalRules(t *testing.T) {
	type args struct {
		ctx        context.Context
		globalRule *model.GlobalRule
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_global_rule",
			args: args{
				ctx:        gatewayCtx,
				globalRule: data.GlobalRule1(gatewayInfo, constant.ResourceStatusCreateDraft),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateGlobalRule(
				tt.args.ctx,
				*tt.args.globalRule,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreateGlobalRule error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.GlobalRule, []string{tt.args.globalRule.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.GlobalRule)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.GlobalRule], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.GlobalRule,
				tt.args.globalRule.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.globalRule.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.GlobalRule, []string{tt.args.globalRule.ID}, "", []constant.ResourceStatus{},
				false)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			globalRule, err := resourcebiz.GetGlobalRule(tt.args.ctx, tt.args.globalRule.ID)
			assert.NoError(t, err)
			assert.Equal(t, constant.ResourceStatusSuccess, globalRule.Status)
		})
	}
}

func TestPublishProtos(t *testing.T) {
	type args struct {
		ctx   context.Context
		proto *model.Proto
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_proto",
			args: args{
				ctx:   gatewayCtx,
				proto: data.Proto1(gatewayInfo, constant.ResourceStatusCreateDraft),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateProto(tt.args.ctx, *tt.args.proto); (err != nil) != tt.wantErr {
				t.Errorf("CreateProto error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.Proto, []string{tt.args.proto.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.Proto)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Proto], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Proto,
				tt.args.proto.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.proto.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.Proto, []string{tt.args.proto.ID}, "", []constant.ResourceStatus{},
				false)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			proto, err := resourcebiz.GetProto(tt.args.ctx, tt.args.proto.ID)
			assert.NoError(t, err)
			assert.Equal(t, constant.ResourceStatusSuccess, proto.Status)
		})
	}
}

func TestPublishPluginMetadatas(t *testing.T) {
	type args struct {
		ctx            context.Context
		pluginMetadata *model.PluginMetadata
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_plugin_metadata",
			args: args{
				ctx:            gatewayCtx,
				pluginMetadata: data.PluginMetadata1(gatewayInfo, constant.ResourceStatusCreateDraft),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreatePluginMetadata(
				tt.args.ctx,
				*tt.args.pluginMetadata,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreatePluginMetadata error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(
				tt.args.ctx,
				constant.PluginMetadata,
				[]string{tt.args.pluginMetadata.ID},
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.PluginMetadata)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.PluginMetadata], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.PluginMetadata,
				tt.args.pluginMetadata.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.pluginMetadata.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(
				tt.args.ctx,
				constant.PluginMetadata,
				[]string{tt.args.pluginMetadata.ID},
				"",
				[]constant.ResourceStatus{},
				false,
			)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			metadata, err := resourcebiz.GetPluginMetadata(tt.args.ctx, tt.args.pluginMetadata.ID)
			assert.NoError(t, err)
			assert.Equal(t, constant.ResourceStatusSuccess, metadata.Status)
		})
	}
}

func TestPublishConsumerGroups(t *testing.T) {
	type args struct {
		ctx           context.Context
		consumerGroup *model.ConsumerGroup
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_consumer_group",
			args: args{
				ctx: gatewayCtx,
				consumerGroup: data.ConsumerGroup1WithNoRelation(
					gatewayInfo,
					constant.ResourceStatusCreateDraft,
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateConsumerGroup(
				tt.args.ctx,
				*tt.args.consumerGroup,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreateConsumerGroup error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.ConsumerGroup, []string{tt.args.consumerGroup.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.ConsumerGroup)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.ConsumerGroup], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.ConsumerGroup,
				tt.args.consumerGroup.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.consumerGroup.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(
				tt.args.ctx,
				constant.ConsumerGroup,
				[]string{tt.args.consumerGroup.ID},
				"",
				[]constant.ResourceStatus{},
				false,
			)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			group, err := resourcebiz.GetConsumerGroup(tt.args.ctx, tt.args.consumerGroup.ID)
			assert.NoError(t, err)
			assert.Equal(t, constant.ResourceStatusSuccess, group.Status)
		})
	}
}

func TestPublishSSLs(t *testing.T) {
	type args struct {
		ctx context.Context
		ssl *model.SSL
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_ssl",
			args: args{
				ctx: gatewayCtx,
				ssl: data.SSL1(gatewayInfo, constant.ResourceStatusCreateDraft),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateSSL(tt.args.ctx, tt.args.ssl); (err != nil) != tt.wantErr {
				t.Errorf("CreateSSL error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.SSL, []string{tt.args.ssl.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.SSL)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.SSL], 1)

			// assert sync resource
			syncedResource, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.SSL,
				tt.args.ssl.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.ssl.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.SSL, []string{tt.args.ssl.ID}, "", []constant.ResourceStatus{},
				false)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			ssl, err := resourcebiz.GetSSL(tt.args.ctx, tt.args.ssl.ID)
			assert.NoError(t, err)
			assert.Equal(t, constant.ResourceStatusSuccess, ssl.Status)
		})
	}
}

func TestPublishStreamRoutes(t *testing.T) {
	type args struct {
		ctx         context.Context
		streamRoute *model.StreamRoute
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_publish_stream_route_without_related_resource",
			args: args{
				ctx: gatewayCtx,
				streamRoute: data.StreamRoute1WithNoRelationResource(
					gatewayInfo,
					constant.ResourceStatusCreateDraft,
				),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建资源
			if err := resourcebiz.CreateStreamRoute(
				tt.args.ctx,
				*tt.args.streamRoute,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreateStreamRoute error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			err := PublishResource(tt.args.ctx, constant.StreamRoute, []string{tt.args.streamRoute.ID})
			if (err != nil) != tt.wantErr {
				t.Errorf("PublishResource error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := unifyopbiz.SyncResources(tt.args.ctx, constant.StreamRoute)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.StreamRoute], 1)

			// assert sync resource
			syncedStreamRoute, err := syncdatabiz.GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.StreamRoute,
				tt.args.streamRoute.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedStreamRoute.ID, tt.args.streamRoute.ID)

			// assert diff resource
			resources, err := diffbiz.DiffResources(tt.args.ctx,
				constant.StreamRoute, []string{tt.args.streamRoute.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			streamRoute, err := resourcebiz.GetStreamRoute(tt.args.ctx, tt.args.streamRoute.ID)
			assert.NoError(t, err)

			assert.Equal(t, constant.ResourceStatusSuccess, streamRoute.Status)
		})
	}
}
