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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// RequestInput captures the external request/import shape before draft preparation.
type RequestInput struct {
	Source       string
	Operation    constant.OperationType
	GatewayID    int
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	PathID       string
	OuterName    string
	OuterFields  map[string]any
	Config       json.RawMessage
}

// ExistingResource captures authoritative stored state used during update/legacy normalization.
type ExistingResource struct {
	ID             string
	Name           string
	ServiceID      string
	UpstreamID     string
	PluginConfigID string
	GroupID        string
	Config         json.RawMessage
}

// ResolvedIdentity is the single identity/association resolution result for one lifecycle operation.
type ResolvedIdentity struct {
	ResourceType   constant.APISIXResource
	ResourceID     string
	NameKey        string
	NameValue      string
	ResolvedFrom   string
	Associations   map[string]string
	Generated      bool
	LegacyDetected bool
}

// ResourceDraft is the internal edit-state representation used by validation and publish payload building.
type ResourceDraft struct {
	GatewayID      int
	ResourceType   constant.APISIXResource
	Version        constant.APISIXVersion
	Identity       ResolvedIdentity
	ConfigSpec     json.RawMessage
	LegacyEchoes   bool
	ExistingConfig json.RawMessage
	CreateTime     int64
	UpdateTime     int64
	Labels         map[string]string
}

// BuiltPayload is the version-aware payload emitted for schema validation or publish.
type BuiltPayload struct {
	Profile      constant.DataType
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	Payload      json.RawMessage
	Dependencies []string
}

// Codec resolves external input into a prepared draft and built APISIX payloads.
type Codec interface {
	ResolveIdentity(input RequestInput, existing *ExistingResource) (ResolvedIdentity, error)
	PrepareDraft(
		input RequestInput,
		identity ResolvedIdentity,
		existing *ExistingResource,
	) (ResourceDraft, error)
	BuildPayload(draft ResourceDraft, profile constant.DataType) (BuiltPayload, error)
}
