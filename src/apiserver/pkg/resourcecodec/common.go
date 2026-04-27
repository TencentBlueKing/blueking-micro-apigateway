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
	"strconv"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// NormalizeProfile defaults unknown validation profiles to DATABASE for request-time use.
func NormalizeProfile(profile constant.DataType) constant.DataType {
	if profile == constant.ETCD {
		return constant.ETCD
	}
	return constant.DATABASE
}

type resourceCodecConfig struct {
	nameKey           string
	associationFields []string
	stripFields       []string
}

func codecConfigFor(resourceType constant.APISIXResource) resourceCodecConfig {
	switch resourceType {
	case constant.Route:
		return routeCodecConfig()
	case constant.Service:
		return serviceCodecConfig()
	case constant.Upstream:
		return upstreamCodecConfig()
	case constant.Consumer:
		return consumerCodecConfig()
	case constant.ConsumerGroup:
		return consumerGroupCodecConfig()
	case constant.PluginConfig:
		return pluginConfigCodecConfig()
	case constant.GlobalRule:
		return globalRuleCodecConfig()
	case constant.PluginMetadata:
		return pluginMetadataCodecConfig()
	case constant.Proto:
		return protoCodecConfig()
	case constant.SSL:
		return sslCodecConfig()
	case constant.StreamRoute:
		return streamRouteCodecConfig()
	default:
		return resourceCodecConfig{nameKey: "name", stripFields: []string{"id", "name"}}
	}
}

// ConflictInput describes request-side fields that must agree with config duplicates.
type ConflictInput struct {
	ResourceType constant.APISIXResource
	ResourceID   string
	OuterName    string
	OuterFields  map[string]string
	Config       []byte
}

// DetectRequestConflicts rejects duplicated semantic fields that disagree with request-side fields.
func DetectRequestConflicts(input ConflictInput) error {
	nameKey := validationNameKey(input.ResourceType)

	if input.ResourceID != "" {
		configID := gjson.GetBytes(input.Config, "id").String()
		if HasLegacyConflict(input.ResourceID, configID) {
			return fmt.Errorf("config.id conflicts with resolved resource id")
		}
	}

	if input.OuterName != "" {
		configName := gjson.GetBytes(input.Config, nameKey).String()
		if HasLegacyConflict(input.OuterName, configName) {
			return fmt.Errorf("config.%s conflicts with request name", nameKey)
		}
		if input.ResourceType == constant.PluginMetadata {
			configID := gjson.GetBytes(input.Config, "id").String()
			if HasLegacyConflict(input.OuterName, configID) {
				return fmt.Errorf("plugin_metadata.id conflicts with request name")
			}
		}
	}

	for fieldName, requestValue := range input.OuterFields {
		if requestValue == "" {
			continue
		}
		configValue := gjson.GetBytes(input.Config, fieldName).String()
		if HasLegacyConflict(requestValue, configValue) {
			return fmt.Errorf("config.%s conflicts with request field", fieldName)
		}
	}

	return nil
}

// PrepareRequestDraft resolves one request identity and strips duplicated server-owned fields.
func PrepareRequestDraft(input RequestInput) (ResourceDraft, error) {
	identity, err := ResolveRequestIdentity(input)
	if err != nil {
		return ResourceDraft{}, err
	}

	if input.Source != SourceImport {
		if err = DetectRequestConflicts(ConflictInput{
			ResourceType: input.ResourceType,
			ResourceID:   requestIdentityID(input, identity),
			OuterName:    requestIdentityName(input, identity),
			OuterFields:  requestAssociationFields(input),
			Config:       input.Config,
		}); err != nil {
			return ResourceDraft{}, err
		}
	}

	configSpec, err := normalizeRequestConfigSpec(input.Config, input.ResourceType)
	if err != nil {
		return ResourceDraft{}, err
	}

	return ResourceDraft{
		GatewayID:    input.GatewayID,
		ResourceType: input.ResourceType,
		Version:      input.Version,
		Identity:     identity,
		ConfigSpec:   configSpec,
	}, nil
}

// ResolveRequestIdentity computes one request identity for validation and storage preparation.
func ResolveRequestIdentity(input RequestInput) (ResolvedIdentity, error) {
	identity := ResolvedIdentity{
		ResourceType: input.ResourceType,
		NameKey:      validationNameKey(input.ResourceType),
		Associations: resolvedAssociations(input),
	}

	if input.PathID != "" {
		identity.ResourceID = input.PathID
		identity.ResolvedFrom = "path"
		identity.Generated = false
	} else if requestID, ok := stringField(input.OuterFields, "id"); ok && requestID != "" {
		identity.ResourceID = requestID
		identity.ResolvedFrom = "structured field"
		identity.Generated = false
	} else if configID := gjson.GetBytes(input.Config, "id").String(); configID != "" {
		identity.ResourceID = configID
		identity.ResolvedFrom = "config"
		identity.Generated = false
	} else if shouldGenerateRequestID(input) {
		identity.ResourceID = idx.GenResourceID(input.ResourceType)
		identity.ResolvedFrom = "generated"
		identity.Generated = true
	}

	identity.NameValue = resolvedName(input, identity.NameKey)
	return identity, nil
}

// BuildRequestPayload builds the version-aware payload used by request-side validation or OpenAPI body shaping.
func BuildRequestPayload(
	draft ResourceDraft,
	profile constant.DataType,
) (BuiltPayload, error) {
	payload, err := buildRequestPayload(draft)
	if err != nil {
		return BuiltPayload{}, err
	}

	payload = cleanupBuiltPayload(draft.ResourceType, draft.Version, NormalizeProfile(profile), payload)
	return BuiltPayload{
		Profile:      NormalizeProfile(profile),
		ResourceType: draft.ResourceType,
		Version:      draft.Version,
		Payload:      payload,
	}, nil
}

// BuildStorageConfig builds the config shape currently expected by persistence/update code paths.
func BuildStorageConfig(draft ResourceDraft) (json.RawMessage, error) {
	return CloneRawMessage(draft.ConfigSpec), nil
}

func buildRequestPayload(draft ResourceDraft) (json.RawMessage, error) {
	payload := CloneRawMessage(draft.ConfigSpec)
	var err error

	if draft.ResourceType == constant.PluginMetadata {
		if draft.Identity.NameValue != "" {
			payload, err = sjson.SetBytes(payload, "id", draft.Identity.NameValue)
			if err != nil {
				return nil, err
			}
		}
	} else if draft.Identity.ResourceID != "" {
		payload, err = sjson.SetBytes(payload, "id", draft.Identity.ResourceID)
		if err != nil {
			return nil, err
		}
	}

	if draft.Identity.NameValue != "" {
		payload, err = sjson.SetBytes(payload, draft.Identity.NameKey, draft.Identity.NameValue)
		if err != nil {
			return nil, err
		}
	}

	for fieldName, value := range draft.Identity.Associations {
		if value == "" {
			continue
		}
		payload, err = sjson.SetBytes(payload, fieldName, value)
		if err != nil {
			return nil, err
		}
	}

	return payload, nil
}

func requestIdentityID(input RequestInput, identity ResolvedIdentity) string {
	if input.PathID != "" {
		return input.PathID
	}
	if requestID, ok := stringField(input.OuterFields, "id"); ok && requestID != "" {
		return requestID
	}
	if input.Source == SourceImport {
		return input.PathID
	}
	if shouldGenerateRequestID(input) {
		return identity.ResourceID
	}
	return ""
}

func requestIdentityName(input RequestInput, identity ResolvedIdentity) string {
	if input.OuterName != "" {
		return input.OuterName
	}
	nameKey := validationNameKey(input.ResourceType)
	if value, ok := stringField(input.OuterFields, nameKey); ok {
		return value
	}
	if nameKey == "username" {
		if value, ok := stringField(input.OuterFields, "name"); ok {
			return value
		}
	}
	if input.Source == SourceImport {
		return identity.NameValue
	}
	return ""
}

func requestAssociationFields(input RequestInput) map[string]string {
	fields := map[string]string{}
	for _, fieldName := range codecConfigFor(input.ResourceType).associationFields {
		if value, ok := stringField(input.OuterFields, fieldName); ok && value != "" {
			fields[fieldName] = value
		}
	}
	return fields
}

func resolvedAssociations(input RequestInput) map[string]string {
	associations := map[string]string{}
	for _, fieldName := range codecConfigFor(input.ResourceType).associationFields {
		if value, ok := stringField(input.OuterFields, fieldName); ok && value != "" {
			associations[fieldName] = value
			continue
		}
		if value := gjson.GetBytes(input.Config, fieldName).String(); value != "" {
			associations[fieldName] = value
		}
	}
	return associations
}

func resolvedName(input RequestInput, nameKey string) string {
	if input.OuterName != "" {
		return input.OuterName
	}
	if value, ok := stringField(input.OuterFields, nameKey); ok && value != "" {
		return value
	}
	if input.ResourceType == constant.PluginMetadata {
		if value := gjson.GetBytes(input.Config, "id").String(); value != "" {
			return value
		}
	}
	if nameKey == "username" {
		if value, ok := stringField(input.OuterFields, "name"); ok && value != "" {
			return value
		}
	}
	if value := gjson.GetBytes(input.Config, nameKey).String(); value != "" {
		return value
	}
	return ""
}

func shouldGenerateRequestID(input RequestInput) bool {
	if input.PathID != "" || input.Operation == constant.OperationTypeUpdate {
		return false
	}
	if input.Source == SourceOpenAPI {
		return true
	}
	return input.Source == SourceWeb &&
		constant.ResourceRequiresIDInSchemaForVersion(input.ResourceType, input.Version)
}

func normalizeRequestConfigSpec(config json.RawMessage, resourceType constant.APISIXResource) (json.RawMessage, error) {
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

func stringField(fields map[string]any, key string) (string, bool) {
	if len(fields) == 0 {
		return "", false
	}
	value, ok := fields[key]
	if !ok || value == nil {
		return "", false
	}
	switch typed := value.(type) {
	case string:
		return typed, true
	case fmt.Stringer:
		return typed.String(), true
	case json.RawMessage:
		return string(typed), true
	case []byte:
		return string(typed), true
	case int:
		return strconv.Itoa(typed), true
	case int64:
		return strconv.FormatInt(typed, 10), true
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(typed), true
	default:
		return fmt.Sprintf("%v", typed), true
	}
}
