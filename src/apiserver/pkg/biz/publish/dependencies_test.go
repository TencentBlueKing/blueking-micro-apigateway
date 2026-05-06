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
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

func TestCollectRoutePublishDependencies(t *testing.T) {
	t.Parallel()

	routes := []*model.Route{
		{
			ServiceID:      "service-1",
			UpstreamID:     "upstream-1",
			PluginConfigID: "plugin-config-1",
		},
		{
			ServiceID:      "service-1",
			UpstreamID:     "upstream-1",
			PluginConfigID: "plugin-config-1",
		},
		{
			ServiceID:      "service-2",
			PluginConfigID: "plugin-config-2",
		},
	}

	deps := collectRoutePublishDependencies(routes)
	assert.Equal(t, []string{"service-1", "service-2"}, deps.ServiceIDs)
	assert.Equal(t, []string{"upstream-1"}, deps.UpstreamIDs)
	assert.Equal(t, []string{"plugin-config-1", "plugin-config-2"}, deps.PluginConfigIDs)
}

func TestCollectServicePublishDependencies(t *testing.T) {
	t.Parallel()

	services := []*model.Service{
		{UpstreamID: "upstream-1"},
		{UpstreamID: "upstream-1"},
		{UpstreamID: ""},
		{UpstreamID: "upstream-2"},
	}

	deps := collectServicePublishDependencies(services)
	assert.Equal(t, []string{"upstream-1", "upstream-2"}, deps.UpstreamIDs)
}

func TestCollectUpstreamPublishDependencies(t *testing.T) {
	t.Parallel()

	upstreams := []*model.Upstream{
		{
			ResourceCommonModel: model.ResourceCommonModel{
				Config: datatypes.JSON(`{"tls":{"client_cert_id":"ssl-1"}}`),
			},
		},
		{
			ResourceCommonModel: model.ResourceCommonModel{
				Config: datatypes.JSON(`{"tls":{"client_cert_id":"ssl-1"}}`),
			},
		},
		{
			ResourceCommonModel: model.ResourceCommonModel{
				Config: datatypes.JSON(`{"tls":{}}`),
			},
		},
		{
			ResourceCommonModel: model.ResourceCommonModel{
				Config: datatypes.JSON(`{"tls":{"client_cert_id":"ssl-2"}}`),
			},
		},
	}

	deps := collectUpstreamPublishDependencies(upstreams)
	assert.Equal(t, []string{"ssl-1", "ssl-2"}, deps.SSLIDs)
}

func TestCollectConsumerPublishDependencies(t *testing.T) {
	t.Parallel()

	consumers := []*model.Consumer{
		{GroupID: "group-1"},
		{GroupID: "group-1"},
		{GroupID: ""},
		{GroupID: "group-2"},
	}

	deps := collectConsumerPublishDependencies(consumers)
	assert.Equal(t, []string{"group-1", "group-2"}, deps.ConsumerGroupIDs)
}

func TestCollectStreamRoutePublishDependencies(t *testing.T) {
	t.Parallel()

	streamRoutes := []*model.StreamRoute{
		{
			ServiceID:  "service-1",
			UpstreamID: "upstream-1",
		},
		{
			ServiceID:  "service-1",
			UpstreamID: "upstream-1",
		},
		{
			ServiceID: "service-2",
		},
	}

	deps := collectStreamRoutePublishDependencies(streamRoutes)
	assert.Equal(t, []string{"service-1", "service-2"}, deps.ServiceIDs)
	assert.Equal(t, []string{"upstream-1"}, deps.UpstreamIDs)
}
