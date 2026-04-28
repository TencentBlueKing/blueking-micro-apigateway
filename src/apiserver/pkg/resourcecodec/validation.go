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
	"fmt"
	"reflect"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

const (
	SourceWeb     = "web"
	SourceOpenAPI = "openapi"
	SourceImport  = "import"
)

// ValidationInput captures the compatibility behavior needed to build request/import validation payloads.
type ValidationInput struct {
	Source       string
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	Config       json.RawMessage
	ResourceID   string
	OuterName    string
}

// ValidationPayloadInput describes an already-built payload that should match a shared profile shape.
type ValidationPayloadInput struct {
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	Profile      constant.DataType
	Payload      json.RawMessage
}

// PrepareValidationPayload centralizes request/import validation payload shaping while preserving current behavior.
func PrepareValidationPayload(input ValidationInput) json.RawMessage {
	payload := CloneRawMessage(input.Config)

	switch input.Source {
	case SourceWeb:
		payload = prepareWebValidationPayload(input, payload)
	case SourceOpenAPI:
		payload = prepareOpenAPIValidationPayload(input, payload)
	default:
		payload = CleanupValidationPayload(input.ResourceType, input.Version, payload)
	}

	return payload
}

// ValidateBuiltPayloadShape verifies that a payload already matches the shared built payload shape.
func ValidateBuiltPayloadShape(input ValidationPayloadInput) (BuiltPayload, error) {
	payload := CloneRawMessage(input.Payload)
	profile := NormalizeProfile(input.Profile)
	expected := cleanupBuiltPayload(
		input.ResourceType,
		input.Version,
		profile,
		CloneRawMessage(payload),
	)
	same, err := sameJSONShape(payload, expected)
	if err != nil {
		return BuiltPayload{}, err
	}
	if !same {
		return BuiltPayload{}, fmt.Errorf("schema 验证失败: payload is not in %s built form", profile)
	}
	return BuiltPayload{
		Profile:      profile,
		ResourceType: input.ResourceType,
		Version:      input.Version,
		Payload:      payload,
	}, nil
}

// CleanupValidationPayload keeps the historical DATABASE-validation cleanup rules used by import/openapi flows.
func CleanupValidationPayload(
	resourceType constant.APISIXResource,
	version constant.APISIXVersion,
	payload json.RawMessage,
) json.RawMessage {
	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "id", version) {
		payload, _ = sjson.DeleteBytes(payload, "id")
	}
	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "name", version) {
		payload, _ = sjson.DeleteBytes(payload, "name")
	}
	return payload
}

func prepareWebValidationPayload(input ValidationInput, payload json.RawMessage) json.RawMessage {
	if constant.ResourceRequiresIDInSchemaForVersion(input.ResourceType, input.Version) &&
		input.ResourceID != "" &&
		!gjson.GetBytes(payload, "id").Exists() {
		payload, _ = sjson.SetBytes(payload, "id", input.ResourceID)
	}

	if getResourceIdentification(payload) == "" &&
		input.OuterName != "" &&
		(resourceSupportsValidationNameInjection(input.ResourceType, input.Version)) {
		payload, _ = sjson.SetBytes(payload, validationNameKey(input.ResourceType), input.OuterName)
	}

	if input.ResourceType == constant.PluginMetadata && input.OuterName != "" {
		payload, _ = sjson.SetBytes(payload, "id", input.OuterName)
	}

	return payload
}

func prepareOpenAPIValidationPayload(input ValidationInput, payload json.RawMessage) json.RawMessage {
	if constant.ResourceRequiresIDInSchema(input.ResourceType) &&
		input.ResourceID != "" &&
		!gjson.GetBytes(payload, "id").Exists() {
		payload, _ = sjson.SetBytes(payload, "id", input.ResourceID)
	}
	return CleanupValidationPayload(input.ResourceType, input.Version, payload)
}

func getResourceIdentification(config json.RawMessage) string {
	if value := gjson.GetBytes(config, "id").String(); value != "" {
		return value
	}
	if value := gjson.GetBytes(config, "name").String(); value != "" {
		return value
	}
	return gjson.GetBytes(config, "username").String()
}

func sameJSONShape(left, right json.RawMessage) (bool, error) {
	var leftValue any
	if err := json.Unmarshal(left, &leftValue); err != nil {
		return false, err
	}
	var rightValue any
	if err := json.Unmarshal(right, &rightValue); err != nil {
		return false, err
	}
	return reflect.DeepEqual(leftValue, rightValue), nil
}

func resourceSupportsValidationNameInjection(
	resourceType constant.APISIXResource,
	version constant.APISIXVersion,
) bool {
	return resourceType == constant.Consumer ||
		constant.ResourceSupportsNameFieldForVersion(resourceType, version)
}

func validationNameKey(resourceType constant.APISIXResource) string {
	return codecConfigFor(resourceType).nameKey
}
