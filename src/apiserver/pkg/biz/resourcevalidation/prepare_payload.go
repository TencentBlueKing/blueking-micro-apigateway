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
	validationRaw := injectRequiredResourceIDForValidation(
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
	validationConfig := injectRequiredResourceIDForValidation(
		apisixVersion,
		resourceType,
		json.RawMessage(configRaw),
		resourceID,
	)

	resourceIdentification, usedFallback := resolveWebValidationIdentity(validationConfig, fallbackIdentity)
	if usedFallback {
		validationConfig, _ = injectResourceNameForValidation(
			apisixVersion,
			resourceType,
			validationConfig,
			resourceIdentification,
		)
	}
	if resourceType == constant.PluginMetadata {
		validationConfig, _ = injectPluginMetadataIDForValidation(validationConfig, name)
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
	validationConfig = injectRequiredResourceIDForValidation(
		apisixVersion,
		resourceType,
		validationConfig,
		resourceID,
	)

	if schemax.GetResourceIdentification(validationConfig) == "" &&
		name != "" {
		var err error
		validationConfig, err = injectResourceNameForValidation(
			apisixVersion,
			resourceType,
			validationConfig,
			name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to inject name into validation payload: %w", err)
		}
	}
	if resourceType == constant.PluginMetadata && name != "" {
		var err error
		validationConfig, err = injectPluginMetadataIDForValidation(validationConfig, name)
		if err != nil {
			return nil, fmt.Errorf("failed to inject plugin metadata id into validation payload: %w", err)
		}
	}

	return buildConfigRawForValidation(apisixVersion, resourceType, string(validationConfig)), nil
}

// =====================================

// injectRequiredResourceIDForValidation injects a server-side resource ID only for validation time.
// Callers decide the ID source; this helper only applies the schema/version rule consistently.
func injectRequiredResourceIDForValidation(
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
	configRaw, _ = sjson.SetBytes(configRaw, "id", resourceID)
	return configRaw
}

// =====================================

func shouldInjectResourceNameForValidation(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
) bool {
	return resourceType == constant.Consumer ||
		constant.ResourceSupportsNameFieldForVersion(resourceType, apisixVersion)
}

func injectResourceNameForValidation(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw json.RawMessage,
	name string,
) (json.RawMessage, error) {
	if !shouldInjectResourceNameForValidation(apisixVersion, resourceType) {
		return configRaw, nil
	}
	return sjson.SetBytes(configRaw, model.GetResourceNameKey(resourceType), name)
}

// =====================================

func injectPluginMetadataIDForValidation(configRaw json.RawMessage, name string) (json.RawMessage, error) {
	return sjson.SetBytes(configRaw, "id", name)
}

// =====================================

// buildConfigRawForValidation builds a validation-only config payload.
func buildConfigRawForValidation(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw string,
) json.RawMessage {
	configRawForValidationBytes := make([]byte, len(configRaw))
	copy(configRawForValidationBytes, configRaw)
	configRawForValidation := json.RawMessage(configRawForValidationBytes)

	return cleanupUnsupportedFieldsForValidation(apisixVersion, resourceType, configRawForValidation)
}

func cleanupUnsupportedFieldsForValidation(
	apisixVersion constant.APISIXVersion,
	resourceType constant.APISIXResource,
	configRaw json.RawMessage,
) json.RawMessage {
	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "id", apisixVersion) {
		configRaw, _ = sjson.DeleteBytes(configRaw, "id")
	}
	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "name", apisixVersion) {
		configRaw, _ = sjson.DeleteBytes(configRaw, "name")
	}

	return configRaw
}

// =====================================

func resolveWebValidationIdentity(configRaw json.RawMessage, fallbackIdentity string) (string, bool) {
	if identity := schemax.GetResourceIdentification(configRaw); identity != "" {
		return identity, false
	}
	return fallbackIdentity, true
}
