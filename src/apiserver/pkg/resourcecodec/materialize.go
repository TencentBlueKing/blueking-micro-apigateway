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

	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	entity "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/apisix"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
)

// StoredRowInput captures the authoritative stored row used to rebuild a canonical draft.
type StoredRowInput struct {
	GatewayID      int
	ResourceType   constant.APISIXResource
	Version        constant.APISIXVersion
	ResourceID     string
	NameKey        string
	NameValue      string
	Associations   map[string]string
	Config         json.RawMessage
	CreateTime     int64
	UpdateTime     int64
	Labels         map[string]string
	LegacyDetected bool
}

// DraftFromStoredRow rebuilds a canonical draft from an existing stored resource row.
func DraftFromStoredRow(input StoredRowInput) CanonicalDraft {
	configSpec, err := DematerializeStoredConfigSpec(input.ResourceType, input.Config)
	if err != nil {
		configSpec = CloneRawMessage(input.Config)
	}
	legacyDetected := input.LegacyDetected || DetectLegacyEchoes(input.ResourceType, input.Config)
	return CanonicalDraft{
		GatewayID:    input.GatewayID,
		ResourceType: input.ResourceType,
		Version:      input.Version,
		Identity: ResolvedIdentity{
			ResourceType:   input.ResourceType,
			ResourceID:     input.ResourceID,
			NameKey:        input.NameKey,
			NameValue:      input.NameValue,
			ResolvedFrom:   "legacy row",
			Associations:   CloneStringMap(input.Associations),
			LegacyDetected: legacyDetected,
		},
		ConfigSpec:     configSpec,
		ExistingConfig: CloneRawMessage(input.Config),
		LegacyEchoes:   legacyDetected,
		CreateTime:     input.CreateTime,
		UpdateTime:     input.UpdateTime,
		Labels:         cloneLabels(input.Labels),
	}
}

// MaterializeStoredDraft converts a stored canonical draft into the payload used for validation/publish.
func MaterializeStoredDraft(draft CanonicalDraft, profile constant.DataType) (MaterializedPayload, error) {
	payload := CloneRawMessage(draft.ConfigSpec)
	var err error

	switch draft.ResourceType {
	case constant.Route:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			Name:       draft.Identity.NameValue,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.Service:
		if associationValue(draft, "upstream_id") != "" {
			payload, err = mergeBaseInfo(payload, entity.BaseInfo{
				ID:         draft.Identity.ResourceID,
				CreateTime: draft.CreateTime,
				UpdateTime: draft.UpdateTime,
			})
		}
	case constant.Upstream:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.PluginConfig:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.PluginMetadata:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.NameValue,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.Consumer:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.ConsumerGroup:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.GlobalRule:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.Proto:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.SSL:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	case constant.StreamRoute:
		payload, err = mergeBaseInfo(payload, entity.BaseInfo{
			ID:         draft.Identity.ResourceID,
			CreateTime: draft.CreateTime,
			UpdateTime: draft.UpdateTime,
		})
	}
	if err != nil {
		return MaterializedPayload{}, err
	}

	payload, err = injectStoredDraftAuthoritativeFields(payload, draft)
	if err != nil {
		return MaterializedPayload{}, err
	}

	payload = cleanupMaterializedPayload(draft.ResourceType, draft.Version, NormalizeProfile(profile), payload)
	return MaterializedPayload{
		Profile:      NormalizeProfile(profile),
		ResourceType: draft.ResourceType,
		Version:      draft.Version,
		Payload:      payload,
	}, nil
}

func injectStoredDraftAuthoritativeFields(payload json.RawMessage, draft CanonicalDraft) (json.RawMessage, error) {
	var err error

	idValue := draft.Identity.ResourceID
	if draft.ResourceType == constant.PluginMetadata {
		idValue = draft.Identity.NameValue
	}
	if draft.ResourceType != constant.Consumer && idValue != "" {
		payload, err = sjson.SetBytes(payload, "id", idValue)
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

func mergeBaseInfo(payload json.RawMessage, baseInfo entity.BaseInfo) (json.RawMessage, error) {
	baseConfig, err := json.Marshal(baseInfo)
	if err != nil {
		return nil, err
	}
	merged, err := jsonx.MergeJson(payload, baseConfig)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(merged), nil
}

func cleanupMaterializedPayload(
	resourceType constant.APISIXResource,
	version constant.APISIXVersion,
	profile constant.DataType,
	payload json.RawMessage,
) json.RawMessage {
	_ = profile
	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "id", version) {
		payload, _ = sjson.DeleteBytes(payload, "id")
	}
	if constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "name", version) {
		payload, _ = sjson.DeleteBytes(payload, "name")
	}

	switch resourceType {
	case constant.SSL:
		payload, _ = sjson.DeleteBytes(payload, "validity_start")
		payload, _ = sjson.DeleteBytes(payload, "validity_end")
	case constant.StreamRoute:
		payload, _ = sjson.DeleteBytes(payload, "labels")
	}
	return payload
}

func associationValue(draft CanonicalDraft, key string) string {
	if draft.Identity.Associations == nil {
		return ""
	}
	return draft.Identity.Associations[key]
}

func cloneLabels(labels map[string]string) map[string]string {
	if labels == nil {
		return nil
	}
	cloned := make(map[string]string, len(labels))
	maps.Copy(cloned, labels)
	return cloned
}
