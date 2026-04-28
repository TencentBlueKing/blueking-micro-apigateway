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

	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/publisher"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
)

type publishPayloadCleanupInput struct {
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	RawConfig    json.RawMessage
}

type publishPayloadCleanupRule struct {
	Field        string
	VersionGated bool
}

type publishResourceOperationInput struct {
	ResourceType constant.APISIXResource
	ResourceKey  string
	BaseInfo     entity.BaseInfo
	Version      constant.APISIXVersion
	RawConfig    json.RawMessage
}

var publishPayloadCleanupRules = map[constant.APISIXResource][]publishPayloadCleanupRule{
	constant.Consumer: {
		{Field: "id", VersionGated: true},
	},
	constant.ConsumerGroup: {
		// Keep both id/name here to mirror the pre-refactor review points.
		// Under current version rules, id is retained and name is removed only before APISIX 3.13.
		{Field: "id", VersionGated: true},
		{Field: "name", VersionGated: true},
	},
	constant.GlobalRule: {
		{Field: "name", VersionGated: true},
	},
	constant.PluginConfig: {
		// Keep both id/name here to make the checked fields explicit.
		// Under current version rules, neither field is removed, so these entries are a no-op today.
		{Field: "id", VersionGated: true},
		{Field: "name", VersionGated: true},
	},
	constant.Proto: {
		{Field: "name", VersionGated: true},
	},
	constant.SSL: {
		{Field: "name", VersionGated: true},
		{Field: "validity_start"},
		{Field: "validity_end"},
	},
	constant.StreamRoute: {
		{Field: "name", VersionGated: true},
		{Field: "labels"},
	},
}

func cleanupPublishPayloadFields(input publishPayloadCleanupInput) json.RawMessage {
	cleaned := append(json.RawMessage(nil), input.RawConfig...)
	for _, rule := range publishPayloadCleanupRules[input.ResourceType] {
		if rule.VersionGated &&
			!constant.ShouldRemoveFieldBeforeValidationOrPublish(
				input.ResourceType,
				rule.Field,
				input.Version,
			) {
			continue
		}
		cleaned, _ = sjson.DeleteBytes(cleaned, rule.Field)
	}
	return cleaned
}

func buildPublishResourceOperation(input publishResourceOperationInput) (publisher.ResourceOperation, error) {
	baseConfig, err := json.Marshal(input.BaseInfo)
	if err != nil {
		return publisher.ResourceOperation{}, err
	}
	merged, err := jsonx.MergeJson(input.RawConfig, baseConfig)
	if err != nil {
		return publisher.ResourceOperation{}, err
	}
	cleaned := cleanupPublishPayloadFields(publishPayloadCleanupInput{
		ResourceType: input.ResourceType,
		Version:      input.Version,
		RawConfig:    merged,
	})
	return publisher.ResourceOperation{
		Key:    input.ResourceKey,
		Config: cleaned,
		Type:   input.ResourceType,
	}, nil
}
