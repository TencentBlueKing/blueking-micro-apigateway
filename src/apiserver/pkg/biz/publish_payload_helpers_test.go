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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
)

func TestCleanupPublishPayloadFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		rawConfig    string
		wantConfig   string
	}{
		{
			name:         "consumer drops id",
			resourceType: constant.Consumer,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"consumer-id","username":"demo","plugins":{"key-auth":{"key":"demo"}}}`,
			wantConfig:   `{"username":"demo","plugins":{"key-auth":{"key":"demo"}}}`,
		},
		{
			name:         "consumer group drops name in 3.11 but keeps id",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"cg-id","name":"cg-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
			wantConfig:   `{"id":"cg-id","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		},
		{
			name:         "consumer group keeps name in 3.13",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			rawConfig:    `{"id":"cg-id","name":"cg-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
			wantConfig:   `{"id":"cg-id","name":"cg-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		},
		{
			name:         "global rule drops name",
			resourceType: constant.GlobalRule,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"gr-id","name":"global-demo","plugins":{"prometheus":{"prefer_name":true}}}`,
			wantConfig:   `{"id":"gr-id","plugins":{"prometheus":{"prefer_name":true}}}`,
		},
		{
			name:         "proto drops name in 3.11",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"proto-id","name":"demo.proto","content":"syntax = \"proto3\";"}`,
			wantConfig:   `{"id":"proto-id","content":"syntax = \"proto3\";"}`,
		},
		{
			name:         "ssl drops name and internal validity fields",
			resourceType: constant.SSL,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"ssl-id","name":"ssl-demo","validity_start":1,"validity_end":2,"cert":"x","key":"y","snis":["demo.com"]}`,
			wantConfig:   `{"id":"ssl-id","cert":"x","key":"y","snis":["demo.com"]}`,
		},
		{
			name:         "stream route drops name in 3.11 and labels",
			resourceType: constant.StreamRoute,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"sr-id","name":"stream-demo","labels":{"env":"prod"},"server_addr":"0.0.0.0","server_port":9100,"upstream":{"nodes":[{"host":"127.0.0.1","port":80,"weight":1}],"type":"roundrobin"}}`,
			wantConfig:   `{"id":"sr-id","server_addr":"0.0.0.0","server_port":9100,"upstream":{"nodes":[{"host":"127.0.0.1","port":80,"weight":1}],"type":"roundrobin"}}`,
		},
		{
			name:         "stream route keeps name in 3.13 but still drops labels",
			resourceType: constant.StreamRoute,
			version:      constant.APISIXVersion313,
			rawConfig:    `{"id":"sr-id","name":"stream-demo","labels":{"env":"prod"},"server_addr":"0.0.0.0","server_port":9100,"upstream":{"nodes":[{"host":"127.0.0.1","port":80,"weight":1}],"type":"roundrobin"}}`,
			wantConfig:   `{"id":"sr-id","name":"stream-demo","server_addr":"0.0.0.0","server_port":9100,"upstream":{"nodes":[{"host":"127.0.0.1","port":80,"weight":1}],"type":"roundrobin"}}`,
		},
		{
			name:         "route stays unchanged",
			resourceType: constant.Route,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"route-id","name":"route-demo","uris":["/demo"]}`,
			wantConfig:   `{"id":"route-id","name":"route-demo","uris":["/demo"]}`,
		},
		{
			name:         "service stays unchanged",
			resourceType: constant.Service,
			version:      constant.APISIXVersion313,
			rawConfig:    `{"id":"svc-id","name":"service-demo","upstream_id":"u-id"}`,
			wantConfig:   `{"id":"svc-id","name":"service-demo","upstream_id":"u-id"}`,
		},
		{
			name:         "upstream stays unchanged",
			resourceType: constant.Upstream,
			version:      constant.APISIXVersion313,
			rawConfig:    `{"id":"u-id","name":"upstream-demo","nodes":[]}`,
			wantConfig:   `{"id":"u-id","name":"upstream-demo","nodes":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanupPublishPayloadFields(publishPayloadCleanupInput{
				ResourceType: tt.resourceType,
				Version:      tt.version,
				RawConfig:    json.RawMessage(tt.rawConfig),
			})
			assert.JSONEq(t, tt.wantConfig, string(got))
		})
	}
}

func TestBuildPublishResourceOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      publishResourceOperationInput
		wantKey    string
		wantConfig string
	}{
		{
			name: "plugin metadata uses plugin name as key and id",
			input: publishResourceOperationInput{
				ResourceType: constant.PluginMetadata,
				ResourceKey:  "limit-count",
				BaseInfo: entity.BaseInfo{
					ID:         "limit-count",
					CreateTime: 1700000000,
					UpdateTime: 1700000001,
				},
				Version: constant.APISIXVersion311,
				RawConfig: json.RawMessage(
					`{"config":{"log_format":{"client_ip":"$remote_addr"}},"name":"limit-count"}`,
				),
			},
			wantKey:    "limit-count",
			wantConfig: `{"id":"limit-count","create_time":1700000000,"update_time":1700000001,"config":{"log_format":{"client_ip":"$remote_addr"}},"name":"limit-count"}`,
		},
		{
			name: "consumer group keeps id and removes name in 3.11",
			input: publishResourceOperationInput{
				ResourceType: constant.ConsumerGroup,
				ResourceKey:  "cg-id",
				BaseInfo: entity.BaseInfo{
					ID:         "cg-id",
					CreateTime: 1700000000,
					UpdateTime: 1700000001,
				},
				Version:   constant.APISIXVersion311,
				RawConfig: json.RawMessage(`{"id":"cg-id","name":"cg-demo","plugins":{}}`),
			},
			wantKey:    "cg-id",
			wantConfig: `{"id":"cg-id","create_time":1700000000,"update_time":1700000001,"plugins":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPublishResourceOperation(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantKey, got.Key)
			assert.Equal(t, tt.input.ResourceType, got.Type)
			assert.JSONEq(t, tt.wantConfig, string(got.Config))
		})
	}
}
