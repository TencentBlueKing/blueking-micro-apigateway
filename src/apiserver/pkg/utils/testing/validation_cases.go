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

package testing

import (
	"encoding/json"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// ValidationSurface identifies which lifecycle surface a regression case targets.
type ValidationSurface string

const (
	ValidationSurfaceWeb         ValidationSurface = "web"
	ValidationSurfaceOpenAPI     ValidationSurface = "openapi"
	ValidationSurfaceImport      ValidationSurface = "import"
	ValidationSurfaceStoredDraft ValidationSurface = "stored-draft"
	ValidationSurfacePublish     ValidationSurface = "publish"
)

// ValidationInput stores the request or stored-config payload that drives a regression case.
type ValidationInput struct {
	PathID       string
	OuterFields  map[string]any
	Config       json.RawMessage
	StoredConfig json.RawMessage
}

// ValidationExpected stores the observable outcome that a regression case should assert.
type ValidationExpected struct {
	Accepted          bool
	ErrorContains     string
	ResolvedID        string
	ResolvedName      string
	ValidationProfile constant.DataType
	StoredConfig      json.RawMessage
	PublishConfig     json.RawMessage
}

// ValidationCase represents one reusable validation regression scenario.
type ValidationCase struct {
	Name         string
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	Surface      ValidationSurface
	Input        ValidationInput
	Expected     ValidationExpected
}

// NewValidationCase creates a reusable regression case with a safely cloned config payload.
func NewValidationCase(
	name string,
	resourceType constant.APISIXResource,
	version constant.APISIXVersion,
	surface ValidationSurface,
	config json.RawMessage,
) ValidationCase {
	return ValidationCase{
		Name:         name,
		ResourceType: resourceType,
		Version:      version,
		Surface:      surface,
		Input: ValidationInput{
			Config: CloneRawMessage(config),
		},
	}
}

// Clone returns a copy that tests can mutate without changing the shared catalog entry.
func (c ValidationCase) Clone() ValidationCase {
	return ValidationCase{
		Name:         c.Name,
		ResourceType: c.ResourceType,
		Version:      c.Version,
		Surface:      c.Surface,
		Input: ValidationInput{
			PathID:       c.Input.PathID,
			OuterFields:  CloneMap(c.Input.OuterFields),
			Config:       CloneRawMessage(c.Input.Config),
			StoredConfig: CloneRawMessage(c.Input.StoredConfig),
		},
		Expected: ValidationExpected{
			Accepted:          c.Expected.Accepted,
			ErrorContains:     c.Expected.ErrorContains,
			ResolvedID:        c.Expected.ResolvedID,
			ResolvedName:      c.Expected.ResolvedName,
			ValidationProfile: c.Expected.ValidationProfile,
			StoredConfig:      CloneRawMessage(c.Expected.StoredConfig),
			PublishConfig:     CloneRawMessage(c.Expected.PublishConfig),
		},
	}
}
