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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// ValidationStage identifies which part of DATABASE payload validation failed.
type ValidationStage string

const (
	ValidationStageResourceSchemaBuild    ValidationStage = "resource_schema_build"
	ValidationStageResourceSchemaValidate ValidationStage = "resource_schema_validate"
	ValidationStageJSONSchemaBuild        ValidationStage = "json_schema_build"
	ValidationStageJSONSchemaValidate     ValidationStage = "json_schema_validate"
)

// ValidationStageError wraps a validation error with its stage.
type ValidationStageError struct {
	Stage ValidationStage
	Err   error
}

func (e *ValidationStageError) Error() string {
	return e.Err.Error()
}

func (e *ValidationStageError) Unwrap() error {
	return e.Err
}

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
	schemaValidator, err := newResourceSchemaPayloadValidator(version, resourceType)
	if err != nil {
		return nil, err
	}
	jsonConfigValidator, err := newJSONSchemaPayloadValidator(version, resourceType, customPluginSchemaMap)
	if err != nil {
		return nil, err
	}
	return &databasePayloadValidator{
		schemaValidator:     schemaValidator,
		jsonConfigValidator: jsonConfigValidator,
	}, nil
}

// Validate validates rawConfig with APISIX resource schema first, then DATABASE JSON schema.
func (v *databasePayloadValidator) Validate(rawConfig json.RawMessage) error {
	if err := validateResourceSchemaPayload(v.schemaValidator, rawConfig); err != nil {
		return err
	}
	return validateJSONSchemaPayload(v.jsonConfigValidator, rawConfig)
}

func newResourceSchemaPayloadValidator(
	version constant.APISIXVersion,
	resourceType constant.APISIXResource,
) (schemax.Validator, error) {
	validator, err := schemax.NewAPISIXSchemaValidator(version, "main."+resourceType.String())
	if err != nil {
		return nil, &ValidationStageError{Stage: ValidationStageResourceSchemaBuild, Err: err}
	}
	return validator, nil
}

func newJSONSchemaPayloadValidator(
	version constant.APISIXVersion,
	resourceType constant.APISIXResource,
	customPluginSchemaMap map[string]any,
) (schemax.Validator, error) {
	validator, err := schemax.NewAPISIXJsonSchemaValidator(
		version,
		resourceType,
		"main."+resourceType.String(),
		customPluginSchemaMap,
		constant.DATABASE,
	)
	if err != nil {
		return nil, &ValidationStageError{Stage: ValidationStageJSONSchemaBuild, Err: err}
	}
	return validator, nil
}

func validateResourceSchemaPayload(validator schemax.Validator, rawConfig json.RawMessage) error {
	if err := validator.Validate(rawConfig); err != nil {
		return &ValidationStageError{Stage: ValidationStageResourceSchemaValidate, Err: err}
	}
	return nil
}

func validateJSONSchemaPayload(validator schemax.Validator, rawConfig json.RawMessage) error {
	if err := validator.Validate(rawConfig); err != nil {
		return &ValidationStageError{Stage: ValidationStageJSONSchemaValidate, Err: err}
	}
	return nil
}
