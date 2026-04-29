package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

type publishResourceEntryTestCase struct {
	name         string
	resourceType constant.APISIXResource
	create       func(context.Context, *model.Gateway, constant.ResourceStatus) (string, error)
}

func TestPublishResource_AllResourceTypes(t *testing.T) {
	testCases := []publishResourceEntryTestCase{
		{
			name:         "route",
			resourceType: constant.Route,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Route1WithNoRelationResource(gateway, status)
				return resource.ID, CreateRoute(ctx, *resource)
			},
		},
		{
			name:         "service",
			resourceType: constant.Service,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Service1WithNoRelation(gateway, status)
				return resource.ID, CreateService(ctx, *resource)
			},
		},
		{
			name:         "upstream",
			resourceType: constant.Upstream,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Upstream1WithNoRelation(gateway, status)
				return resource.ID, CreateUpstream(ctx, *resource)
			},
		},
		{
			name:         "consumer",
			resourceType: constant.Consumer,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Consumer1WithNoRelation(gateway, status)
				return resource.ID, CreateConsumer(ctx, *resource)
			},
		},
		{
			name:         "consumer_group",
			resourceType: constant.ConsumerGroup,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.ConsumerGroup1WithNoRelation(gateway, status)
				return resource.ID, CreateConsumerGroup(ctx, *resource)
			},
		},
		{
			name:         "plugin_config",
			resourceType: constant.PluginConfig,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.PluginConfig1WithNoRelation(gateway, status)
				return resource.ID, CreatePluginConfig(ctx, *resource)
			},
		},
		{
			name:         "global_rule",
			resourceType: constant.GlobalRule,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.GlobalRule1(gateway, status)
				return resource.ID, CreateGlobalRule(ctx, *resource)
			},
		},
		{
			name:         "plugin_metadata",
			resourceType: constant.PluginMetadata,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.PluginMetadata1(gateway, status)
				return resource.ID, CreatePluginMetadata(ctx, *resource)
			},
		},
		{
			name:         "proto",
			resourceType: constant.Proto,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.Proto1(gateway, status)
				return resource.ID, CreateProto(ctx, *resource)
			},
		},
		{
			name:         "ssl",
			resourceType: constant.SSL,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.SSL1(gateway, status)
				return resource.ID, CreateSSL(ctx, resource)
			},
		},
		{
			name:         "stream_route",
			resourceType: constant.StreamRoute,
			create: func(ctx context.Context, gateway *model.Gateway, status constant.ResourceStatus) (string, error) {
				resource := data.StreamRoute1WithNoRelationResource(gateway, status)
				return resource.ID, CreateStreamRoute(ctx, *resource)
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

			diffResources, err := DiffResources(
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

			storedResources, err := BatchGetResources(ctx, tc.resourceType, []string{resourceID})
			if err != nil {
				t.Fatal(err)
			}
			if assert.Len(t, storedResources, 1) {
				assert.Equal(t, constant.ResourceStatusSuccess, storedResources[0].Status)
			}
		})
	}
}
