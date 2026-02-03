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
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

var (
	gatewayInfo *model.Gateway
	gatewayCtx  context.Context
)

func TestMain(m *testing.M) {
	// init crypto
	err := cryptography.Init("jxi18GX5w2qgHwfZCFpn07q8FScXJOd3", "k2dbCGetyusW")
	if err != nil {
		panic(err)
	}

	// 初始化embed数据库
	util.InitEmbedDb()

	_, server, err := util.StartEmbedEtcdClient(context.Background())
	if err != nil {
		panic(err)
	}

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
	err := CreateGateway(context.Background(), gatewayInfo)
	if err != nil {
		panic(err)
	}
	gatewayCtx = ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
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
			if err := CreateRoute(tt.args.ctx, *tt.args.route); (err != nil) != tt.wantErr {
				t.Errorf("CreateRoute error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishRoutes(tt.args.ctx, []string{tt.args.route.ID}); (err != nil) != tt.wantErr {
				t.Errorf("PublishRoutes error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.Route)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Route], 1)

			// assert sync resource
			syncedRoute, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Route,
				tt.args.route.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedRoute.ID, tt.args.route.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.Route, []string{tt.args.route.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			route, err := GetRoute(tt.args.ctx, tt.args.route.ID)
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
			if err := CreateService(tt.args.ctx, *tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("CreateService error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishServices(
				tt.args.ctx,
				[]string{tt.args.service.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishService error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.Service)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Service], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Service,
				tt.args.service.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedResource.ID, tt.args.service.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.Service, []string{tt.args.service.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			service, err := GetService(tt.args.ctx, tt.args.service.ID)
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
			if err := CreateUpstream(tt.args.ctx, *tt.args.upstream); (err != nil) != tt.wantErr {
				t.Errorf("CreateUpstream error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishUpstreams(
				tt.args.ctx,
				[]string{tt.args.upstream.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishUpstream error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.Upstream)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Upstream], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Upstream,
				tt.args.upstream.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedResource.ID, tt.args.upstream.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.Upstream, []string{tt.args.upstream.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			upstream, err := GetUpstream(tt.args.ctx, tt.args.upstream.ID)
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
			if err := CreateConsumer(tt.args.ctx, *tt.args.consumer); (err != nil) != tt.wantErr {
				t.Errorf("CreateConsumer error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishConsumers(
				tt.args.ctx,
				[]string{tt.args.consumer.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishConsumer error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.Consumer)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Consumer], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Consumer,
				tt.args.consumer.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedResource.ID, tt.args.consumer.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.Consumer, []string{tt.args.consumer.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			consumer, err := GetConsumer(tt.args.ctx, tt.args.consumer.ID)
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
			if err := CreatePluginConfig(tt.args.ctx, *tt.args.pluginConfig); (err != nil) != tt.wantErr {
				t.Errorf("CreatePluginConfig error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishPluginConfigs(
				tt.args.ctx,
				[]string{tt.args.pluginConfig.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishPluginConfigs error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.PluginConfig)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.PluginConfig], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.PluginConfig,
				tt.args.pluginConfig.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.pluginConfig.ID)

			// assert diff resource
			resources, err := DiffResources(
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
			pluginConfig, err := GetPluginConfig(tt.args.ctx, tt.args.pluginConfig.ID)
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
			if err := CreateGlobalRule(tt.args.ctx, *tt.args.globalRule); (err != nil) != tt.wantErr {
				t.Errorf("CreateGlobalRule error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishGlobalRules(
				tt.args.ctx,
				[]string{tt.args.globalRule.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishGlobalRules error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.GlobalRule)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.GlobalRule], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.GlobalRule,
				tt.args.globalRule.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.globalRule.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.GlobalRule, []string{tt.args.globalRule.ID}, "", []constant.ResourceStatus{},
				false)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			globalRule, err := GetGlobalRule(tt.args.ctx, tt.args.globalRule.ID)
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
			if err := CreateProto(tt.args.ctx, *tt.args.proto); (err != nil) != tt.wantErr {
				t.Errorf("CreateProto error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishProtos(tt.args.ctx, []string{tt.args.proto.ID}); (err != nil) != tt.wantErr {
				t.Errorf("PublishProtos error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.Proto)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.Proto], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.Proto,
				tt.args.proto.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.proto.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.Proto, []string{tt.args.proto.ID}, "", []constant.ResourceStatus{},
				false)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			proto, err := GetProto(tt.args.ctx, tt.args.proto.ID)
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
			if err := CreatePluginMetadata(
				tt.args.ctx,
				*tt.args.pluginMetadata,
			); (err != nil) != tt.wantErr {
				t.Errorf("CreatePluginMetadata error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishPluginMetadatas(
				tt.args.ctx,
				[]string{tt.args.pluginMetadata.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishPluginMetadatas error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.PluginMetadata)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.PluginMetadata], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.PluginMetadata,
				tt.args.pluginMetadata.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.pluginMetadata.ID)

			// assert diff resource
			resources, err := DiffResources(
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
			metadata, err := GetPluginMetadata(tt.args.ctx, tt.args.pluginMetadata.ID)
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
			if err := CreateConsumerGroup(tt.args.ctx, *tt.args.consumerGroup); (err != nil) != tt.wantErr {
				t.Errorf("CreateConsumerGroup error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishConsumerGroups(
				tt.args.ctx,
				[]string{tt.args.consumerGroup.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishConsumerGroups error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.ConsumerGroup)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.ConsumerGroup], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.ConsumerGroup,
				tt.args.consumerGroup.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.consumerGroup.ID)

			// assert diff resource
			resources, err := DiffResources(
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
			group, err := GetConsumerGroup(tt.args.ctx, tt.args.consumerGroup.ID)
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
			if err := CreateSSL(tt.args.ctx, tt.args.ssl); (err != nil) != tt.wantErr {
				t.Errorf("CreateSSL error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishSSLs(tt.args.ctx, []string{tt.args.ssl.ID}); (err != nil) != tt.wantErr {
				t.Errorf("PublishSSLs error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.SSL)
			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.SSL], 1)

			// assert sync resource
			syncedResource, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.SSL,
				tt.args.ssl.ID,
			)
			assert.NoError(t, err)
			assert.Equal(t, syncedResource.ID, tt.args.ssl.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.SSL, []string{tt.args.ssl.ID}, "", []constant.ResourceStatus{},
				false)
			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			ssl, err := GetSSL(tt.args.ctx, tt.args.ssl.ID)
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
			if err := CreateStreamRoute(tt.args.ctx, *tt.args.streamRoute); (err != nil) != tt.wantErr {
				t.Errorf("CreateStreamRoute error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 发布资源
			if err := PublishStreamRoutes(
				tt.args.ctx,
				[]string{tt.args.streamRoute.ID},
			); (err != nil) != tt.wantErr {
				t.Errorf("PublishStreamRoutes error = %v, wantErr %v", err, tt.wantErr)
			}

			// sync resource
			syncedResourceTypeStats, err := SyncResources(tt.args.ctx, constant.StreamRoute)

			assert.NoError(t, err)

			// assert sync resource count
			assert.Equal(t, syncedResourceTypeStats[constant.StreamRoute], 1)

			// assert sync resource
			syncedStreamRoute, err := GetSyncedItemByResourceTypeAndID(
				tt.args.ctx,
				constant.StreamRoute,
				tt.args.streamRoute.ID,
			)
			assert.NoError(t, err)

			assert.Equal(t, syncedStreamRoute.ID, tt.args.streamRoute.ID)

			// assert diff resource
			resources, err := DiffResources(tt.args.ctx,
				constant.StreamRoute, []string{tt.args.streamRoute.ID}, "", []constant.ResourceStatus{},
				false)

			assert.NoError(t, err)
			assert.Equal(t, 0, len(resources))

			// assert resource status is published
			streamRoute, err := GetStreamRoute(tt.args.ctx, tt.args.streamRoute.ID)
			assert.NoError(t, err)

			assert.Equal(t, constant.ResourceStatusSuccess, streamRoute.Status)
		})
	}
}
