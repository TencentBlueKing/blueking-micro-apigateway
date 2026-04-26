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

package resourcecodec

import (
	"encoding/json"
	"maps"
	"reflect"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// CloneRawMessage returns a detached JSON payload copy for legacy normalization helpers.
func CloneRawMessage(raw json.RawMessage) json.RawMessage {
	if raw == nil {
		return nil
	}
	cloned := make(json.RawMessage, len(raw))
	copy(cloned, raw)
	return cloned
}

// CloneStringMap returns a detached copy for association and duplicate-field bookkeeping.
func CloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]string, len(values))
	maps.Copy(cloned, values)
	return cloned
}

// IsLegacyDuplicate reports whether the config copy matches the authoritative value for a duplicated field.
func IsLegacyDuplicate(authoritative, configValue string) bool {
	return authoritative != "" && configValue != "" && authoritative == configValue
}

// HasLegacyConflict reports whether the config copy conflicts with the authoritative value for a duplicated field.
func HasLegacyConflict(authoritative, configValue string) bool {
	return authoritative != "" && configValue != "" && authoritative != configValue
}

// SameAssociations reports whether two association maps are semantically equal.
func SameAssociations(left, right map[string]string) bool {
	return reflect.DeepEqual(CloneStringMap(left), CloneStringMap(right))
}

// DetectLegacyEchoes reports whether stored config still carries duplicated server-owned fields.
func DetectLegacyEchoes(resourceType constant.APISIXResource, config json.RawMessage) bool {
	for _, fieldName := range codecConfigFor(resourceType).stripFields {
		if gjson.GetBytes(config, fieldName).Exists() {
			return true
		}
	}
	return false
}

// DematerializeStoredConfigSpec removes duplicated server-owned fields from a stored config row.
func DematerializeStoredConfigSpec(
	resourceType constant.APISIXResource,
	config json.RawMessage,
) (json.RawMessage, error) {
	payload := CloneRawMessage(config)
	var err error
	for _, fieldName := range codecConfigFor(resourceType).stripFields {
		if !gjson.GetBytes(payload, fieldName).Exists() {
			continue
		}
		payload, err = sjson.DeleteBytes(payload, fieldName)
		if err != nil {
			return nil, err
		}
	}
	return payload, nil
}
