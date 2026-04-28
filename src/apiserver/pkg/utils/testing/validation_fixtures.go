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

package testing

import (
	"encoding/json"
	"maps"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// StoredResourceFixture mirrors the model-column and config state used by legacy compatibility tests.
type StoredResourceFixture struct {
	ID             string
	Name           string
	ServiceID      string
	UpstreamID     string
	PluginConfigID string
	GroupID        string
	Config         json.RawMessage
}

// HistoricalValidationFixture captures a stored row plus its target version for regression tests.
type HistoricalValidationFixture struct {
	Name           string
	ResourceType   constant.APISIXResource
	Version        constant.APISIXVersion
	Stored         StoredResourceFixture
	DatabaseConfig json.RawMessage
	PublishConfig  json.RawMessage
}

// CloneRawMessage returns a detached copy that tests can mutate safely.
func CloneRawMessage(raw json.RawMessage) json.RawMessage {
	if raw == nil {
		return nil
	}
	cloned := make(json.RawMessage, len(raw))
	copy(cloned, raw)
	return cloned
}

// CloneMap returns a JSON-cloned map for test fixtures with nested values.
func CloneMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	raw, err := json.Marshal(values)
	if err != nil {
		maps.Copy(cloned, values)
		return cloned
	}
	if err := json.Unmarshal(raw, &cloned); err != nil {
		maps.Copy(cloned, values)
	}
	return cloned
}

// HistoricalValidationFixtures returns representative legacy shapes that must remain import-valid and publishable.
func HistoricalValidationFixtures() []HistoricalValidationFixture {
	return []HistoricalValidationFixture{
		{
			Name:         "route legacy echoed relation fields",
			ResourceType: constant.Route,
			Version:      constant.APISIXVersion311,
			Stored: StoredResourceFixture{
				ID:        "route-id",
				Name:      "route-a",
				ServiceID: "service-a",
				Config: json.RawMessage(
					`{"id":"legacy-route","name":"legacy-route","service_id":"legacy-service","uris":["/test"]}`,
				),
			},
			DatabaseConfig: json.RawMessage(
				`{"id":"route-id","name":"route-a","service_id":"service-a","uris":["/test"]}`,
			),
			PublishConfig: json.RawMessage(
				`{"id":"route-id","name":"route-a","service_id":"service-a","uris":["/test"]}`,
			),
		},
		{
			Name:         "consumer legacy id echo",
			ResourceType: constant.Consumer,
			Version:      constant.APISIXVersion313,
			Stored: StoredResourceFixture{
				ID:      "consumer-id",
				Name:    "consumer-a",
				GroupID: "group-a",
				Config: json.RawMessage(
					`{
						"id":"legacy-consumer",
						"username":"legacy-consumer",
						"group_id":"legacy-group",
						"plugins":{"key-auth":{"key":"demo"}}
					}`,
				),
			},
			DatabaseConfig: json.RawMessage(
				`{"username":"consumer-a","group_id":"group-a","plugins":{"key-auth":{"key":"demo"}}}`,
			),
			PublishConfig: json.RawMessage(
				`{"username":"consumer-a","group_id":"group-a","plugins":{"key-auth":{"key":"demo"}}}`,
			),
		},
		{
			Name:         "plugin metadata legacy id and name",
			ResourceType: constant.PluginMetadata,
			Version:      constant.APISIXVersion313,
			Stored: StoredResourceFixture{
				ID:     "plugin-metadata-id",
				Name:   "jwt-auth",
				Config: json.RawMessage(`{"id":"basic-auth","name":"basic-auth","key":"value"}`),
			},
			DatabaseConfig: json.RawMessage(`{"id":"jwt-auth","key":"value"}`),
			PublishConfig:  json.RawMessage(`{"id":"jwt-auth","key":"value"}`),
		},
		{
			Name:         "stream route 3.11 drops unsupported name",
			ResourceType: constant.StreamRoute,
			Version:      constant.APISIXVersion311,
			Stored: StoredResourceFixture{
				ID:         "stream-route-id",
				Name:       "stream-a",
				UpstreamID: "upstream-a",
				Config: json.RawMessage(
					`{"name":"legacy-stream","upstream_id":"legacy-upstream","remote_addr":"127.0.0.1","server_port":9100}`,
				),
			},
			DatabaseConfig: json.RawMessage(
				`{"id":"stream-route-id","upstream_id":"upstream-a","remote_addr":"127.0.0.1","server_port":9100}`,
			),
			PublishConfig: json.RawMessage(
				`{"id":"stream-route-id","upstream_id":"upstream-a","remote_addr":"127.0.0.1","server_port":9100}`,
			),
		},
	}
}
