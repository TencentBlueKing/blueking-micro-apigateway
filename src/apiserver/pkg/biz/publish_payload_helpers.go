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

var publishPayloadCleanupRules = map[constant.APISIXResource][]publishPayloadCleanupRule{
	constant.Consumer: {
		{Field: "id", VersionGated: true},
	},
	constant.ConsumerGroup: {
		{Field: "name", VersionGated: true},
	},
	constant.GlobalRule: {
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
			!constant.ShouldRemoveFieldBeforeValidationOrPublish(input.ResourceType, rule.Field, input.Version) {
			continue
		}
		cleaned, _ = sjson.DeleteBytes(cleaned, rule.Field)
	}
	return cleaned
}
