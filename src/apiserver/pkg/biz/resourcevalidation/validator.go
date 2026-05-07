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

// Package resourcevalidation owns shared APISIX resource config validation helpers.
package resourcevalidation

import (
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// DatabasePayloadValidator validates an already-prepared DATABASE payload.
type DatabasePayloadValidator interface {
	Validate(rawConfig json.RawMessage) error
}

type databasePayloadValidator struct {
	schemaValidator     schemax.Validator
	jsonConfigValidator schemax.Validator
}

// NewDatabasePayloadValidator builds a DATABASE payload validator for one APISIX resource type.
func NewDatabasePayloadValidator(
	version constant.APISIXVersion,
	resourceType constant.APISIXResource,
	customPluginSchemaMap map[string]any,
) (DatabasePayloadValidator, error) {
	schemaValidator, err := schemax.NewAPISIXSchemaValidator(version, "main."+resourceType.String())
	if err != nil {
		return nil, fmt.Errorf("new APISIX schema validator failed, resource type:%s validate failed: %w",
			resourceType, err)
	}
	jsonConfigValidator, err := schemax.NewAPISIXJsonSchemaValidator(
		version,
		resourceType,
		"main."+resourceType.String(),
		customPluginSchemaMap,
		constant.DATABASE,
	)
	if err != nil {
		return nil, fmt.Errorf("NewAPISIXJsonSchemaValidator failed, resource type:%s validate failed, err: %w",
			resourceType, err)
	}
	return &databasePayloadValidator{
		schemaValidator:     schemaValidator,
		jsonConfigValidator: jsonConfigValidator,
	}, nil
}

// Validate validates rawConfig with APISIX resource schema first, then DATABASE JSON schema.
func (v *databasePayloadValidator) Validate(rawConfig json.RawMessage) error {
	if err := v.schemaValidator.Validate(rawConfig); err != nil {
		resourceIdentification := schemax.GetResourceIdentification(rawConfig)
		if resourceIdentification != "" {
			return fmt.Errorf("resource:%s validate failed: %w", resourceIdentification, err)
		}
		return fmt.Errorf("resource validate failed: %w", err)
	}
	if err := v.jsonConfigValidator.Validate(rawConfig); err != nil {
		return fmt.Errorf("resource config:%s validate failed, err: %w", rawConfig, err)
	}
	return nil
}
