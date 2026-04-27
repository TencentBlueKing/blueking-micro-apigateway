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

package common

import (
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
)

// PreparedStoredResource carries the stripped storage config plus resolved fields
// shared by the Web/OpenAPI/MCP/import write adapters.
type PreparedStoredResource struct {
	GatewayID      int
	ResourceID     string
	StorageConfig  datatypes.JSON
	ResolvedValues model.ResourceResolvedValues
}

// PrepareStoredResource resolves one request input through the strict request-draft path.
func PrepareStoredResource(input resourcecodec.RequestInput) (PreparedStoredResource, error) {
	draft, err := resourcecodec.PrepareRequestDraft(input)
	if err != nil {
		return PreparedStoredResource{}, err
	}
	return PrepareStoredResourceFromDraft(draft)
}

// PrepareStoredResourceFromDraft converts a pre-resolved request draft into the stored shape.
func PrepareStoredResourceFromDraft(draft resourcecodec.ResourceDraft) (PreparedStoredResource, error) {
	storageConfig, err := resourcecodec.BuildStorageConfig(draft)
	if err != nil {
		return PreparedStoredResource{}, err
	}
	return PreparedStoredResource{
		GatewayID:     draft.GatewayID,
		ResourceID:    draft.Identity.ResourceID,
		StorageConfig: datatypes.JSON(storageConfig),
		ResolvedValues: model.NewResourceResolvedValues(
			draft.Identity.NameValue,
			draft.Identity.Associations,
		),
	}, nil
}

// BuildFallbackStoredResource keeps the best-effort fallback semantics used by
// OpenAPI, MCP, and import normalization when strict request-draft preparation fails.
func BuildFallbackStoredResource(input resourcecodec.RequestInput) PreparedStoredResource {
	identity, _ := resourcecodec.ResolveRequestIdentity(input)
	storageConfig, err := resourcecodec.ExtractStoredConfigSpec(input.ResourceType, input.Config)
	if err != nil {
		storageConfig = resourcecodec.CloneRawMessage(input.Config)
	}
	resourceID := identity.ResourceID
	if resourceID == "" {
		resourceID = input.PathID
	}
	return PreparedStoredResource{
		GatewayID:     input.GatewayID,
		ResourceID:    resourceID,
		StorageConfig: datatypes.JSON(storageConfig),
		ResolvedValues: model.NewResourceResolvedValues(
			identity.NameValue,
			identity.Associations,
		),
	}
}

// BuildResourceCommonModel wraps a prepared stored resource into the shared
// ResourceCommonModel used by the various write adapters.
func BuildResourceCommonModel(
	prepared PreparedStoredResource,
	status constant.ResourceStatus,
	creator string,
	updater string,
) *model.ResourceCommonModel {
	resource := &model.ResourceCommonModel{
		ID:        prepared.ResourceID,
		GatewayID: prepared.GatewayID,
		Config:    prepared.StorageConfig,
		Status:    status,
		BaseModel: model.BaseModel{
			Creator: creator,
			Updater: updater,
		},
	}
	resource.ApplyResolvedValues(prepared.ResolvedValues)
	return resource
}
