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

// Package publish contains APISIX publish orchestration and payload preparation helpers.
package publish

import "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"

type routePublishDependencies struct {
	ServiceIDs      []string
	UpstreamIDs     []string
	PluginConfigIDs []string
}

type servicePublishDependencies struct {
	UpstreamIDs []string
}

type upstreamPublishDependencies struct {
	SSLIDs []string
}

type consumerPublishDependencies struct {
	ConsumerGroupIDs []string
}

type streamRoutePublishDependencies struct {
	ServiceIDs  []string
	UpstreamIDs []string
}

func collectRoutePublishDependencies(routes []*model.Route) routePublishDependencies {
	deps := routePublishDependencies{}
	for _, route := range routes {
		if route.ServiceID != "" {
			deps.ServiceIDs = append(deps.ServiceIDs, route.ServiceID)
		}
		if route.UpstreamID != "" {
			deps.UpstreamIDs = append(deps.UpstreamIDs, route.UpstreamID)
		}
		if route.PluginConfigID != "" {
			deps.PluginConfigIDs = append(deps.PluginConfigIDs, route.PluginConfigID)
		}
	}
	return deps
}

func collectServicePublishDependencies(services []*model.Service) servicePublishDependencies {
	deps := servicePublishDependencies{}
	for _, service := range services {
		if service.UpstreamID != "" {
			deps.UpstreamIDs = append(deps.UpstreamIDs, service.UpstreamID)
		}
	}
	return deps
}

func collectUpstreamPublishDependencies(upstreams []*model.Upstream) upstreamPublishDependencies {
	deps := upstreamPublishDependencies{}
	for _, upstream := range upstreams {
		if upstream.GetSSLID() != "" {
			deps.SSLIDs = append(deps.SSLIDs, upstream.GetSSLID())
		}
	}
	return deps
}

func collectConsumerPublishDependencies(consumers []*model.Consumer) consumerPublishDependencies {
	deps := consumerPublishDependencies{}
	for _, consumer := range consumers {
		if consumer.GroupID != "" {
			deps.ConsumerGroupIDs = append(deps.ConsumerGroupIDs, consumer.GroupID)
		}
	}
	return deps
}

func collectStreamRoutePublishDependencies(streamRoutes []*model.StreamRoute) streamRoutePublishDependencies {
	deps := streamRoutePublishDependencies{}
	for _, streamRoute := range streamRoutes {
		if streamRoute.ServiceID != "" {
			deps.ServiceIDs = append(deps.ServiceIDs, streamRoute.ServiceID)
		}
		if streamRoute.UpstreamID != "" {
			deps.UpstreamIDs = append(deps.UpstreamIDs, streamRoute.UpstreamID)
		}
	}
	return deps
}
