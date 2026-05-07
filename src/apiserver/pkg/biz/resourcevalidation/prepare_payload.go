/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
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

package resourcevalidation

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// PrepareOpenValidationPayload builds the validation-only DATABASE payload for Open API resource requests.
func PrepareOpenValidationPayload(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw string,
) json.RawMessage {
	resourceID := ""
	// check if id exists, if not, create on by idx.GenResourceID(resourceType)
	if constant.ResourceRequiresIDInSchemaForVersion(resourceType, apisixVersion) &&
		!gjson.Get(configRaw, "id").Exists() {
		resourceID = idx.GenResourceID(resourceType)
	}
	validationRaw := injectGeneratedIDForValidation(
		apisixVersion,
		resourceType,
		json.RawMessage(configRaw),
		resourceID,
	)
	return buildConfigRawForValidation(apisixVersion, resourceType, string(validationRaw))
}

// PrepareImportValidationPayload builds the validation-only DATABASE payload for imported resources.
func PrepareImportValidationPayload(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw string,
) json.RawMessage {
	return buildConfigRawForValidation(apisixVersion, resourceType, configRaw)
}

// PrepareWebValidationPayload builds the validation-only payload for Web API serializer checks.
func PrepareWebValidationPayload(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw string,
	resourceID string,
	name string,
	fallbackIdentity string,
) (json.RawMessage, string) {
	validationConfig := injectGeneratedIDForValidation(
		apisixVersion,
		resourceType,
		json.RawMessage(configRaw),
		resourceID,
	)

	resourceIdentification, usedFallback := resolveWebValidationIdentity(validationConfig, fallbackIdentity)
	// FIXME: config modified logical
	if usedFallback && shouldInjectResourceNameForValidation(apisixVersion, resourceType) {
		validationConfig, _ = sjson.SetBytes(
			validationConfig,
			model.GetResourceNameKey(resourceType),
			resourceIdentification,
		)
	}
	// FIXME: config modified logical
	if resourceType == constant.PluginMetadata {
		validationConfig, _ = sjson.SetBytes(validationConfig, "id", name)
	}
	return validationConfig, resourceIdentification
}

// PrepareMCPDatabaseValidationPayload builds the validation-only DATABASE payload for MCP resource writes.
func PrepareMCPDatabaseValidationPayload(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw string,
	resourceID string,
	name string,
) (json.RawMessage, error) {
	// if create, the resourceID is idx.GenResourceID(resourceType)
	// if update, get from request
	validationConfig := json.RawMessage(configRaw)
	validationConfig = injectGeneratedIDForValidation(
		apisixVersion,
		resourceType,
		validationConfig,
		resourceID,
	)

	if schemax.GetResourceIdentification(validationConfig) == "" &&
		name != "" &&
		shouldInjectResourceNameForValidation(apisixVersion, resourceType) {
		var err error
		validationConfig, err = sjson.SetBytes(
			validationConfig,
			model.GetResourceNameKey(resourceType),
			name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to inject name into validation payload: %w", err)
		}
	}
	if resourceType == constant.PluginMetadata && name != "" {
		var err error
		validationConfig, err = sjson.SetBytes(validationConfig, "id", name)
		if err != nil {
			return nil, fmt.Errorf("failed to inject plugin metadata id into validation payload: %w", err)
		}
	}

	return buildConfigRawForValidation(apisixVersion, resourceType, string(validationConfig)), nil
}

func resolveWebValidationIdentity(configRaw json.RawMessage, fallbackIdentity string) (string, bool) {
	if identity := schemax.GetResourceIdentification(configRaw); identity != "" {
		return identity, false
	}
	return fallbackIdentity, true
}

// injectGeneratedIDForValidation injects a server-side resource ID only for validation time.
// Callers decide the ID source; this helper only applies the schema/version rule consistently.
func injectGeneratedIDForValidation(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw json.RawMessage,
	resourceID string,
) json.RawMessage {
	if !constant.ResourceRequiresIDInSchemaForVersion(resourceType, apisixVersion) || resourceID == "" {
		return configRaw
	}
	if gjson.GetBytes(configRaw, "id").Exists() {
		return configRaw
	}
	// FIXME: config modified logical
	configRaw, _ = sjson.SetBytes(configRaw, "id", resourceID)
	return configRaw
}

// buildConfigRawForValidation builds a validation-only config payload.
func buildConfigRawForValidation(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw string,
) json.RawMessage {
	configRawForValidationBytes := make([]byte, len(configRaw))
	copy(configRawForValidationBytes, configRaw)
	configRawForValidation := json.RawMessage(configRawForValidationBytes)

	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "id", apisixVersion) {
		configRawForValidation, _ = sjson.DeleteBytes(configRawForValidation, "id")
	}
	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "name", apisixVersion) {
		configRawForValidation, _ = sjson.DeleteBytes(configRawForValidation, "name")
	}

	return configRawForValidation
}

func shouldInjectResourceNameForValidation(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
) bool {
	return resourceType == constant.Consumer ||
		constant.ResourceSupportsNameFieldForVersion(resourceType, apisixVersion)
}
