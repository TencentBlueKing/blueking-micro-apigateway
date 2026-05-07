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
	"errors"
	"testing"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

func TestNewDatabasePayloadValidatorBuildsDatabaseValidators(t *testing.T) {
	customSchemaMap := map[string]any{"custom-plugin": map[string]any{"type": "object"}}
	var gotSchemaVersion constant.APISIXVersion
	var gotSchemaPath string
	var gotJSONVersion constant.APISIXVersion
	var gotResourceType constant.APISIXResource
	var gotJSONPath string
	var gotCustomSchemaMap map[string]any
	var gotDataType constant.DataType
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(
		schemax.NewAPISIXSchemaValidator,
		func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
			gotSchemaVersion = version
			gotSchemaPath = jsonPath
			return captureValidator{}, nil
		},
	)
	patches.ApplyFunc(
		schemax.NewAPISIXJsonSchemaValidator,
		func(
			version constant.APISIXVersion,
			resourceType constant.APISIXResource,
			jsonPath string,
			customizePluginSchemaMap map[string]any,
			dataType constant.DataType,
		) (schemax.Validator, error) {
			gotJSONVersion = version
			gotResourceType = resourceType
			gotJSONPath = jsonPath
			gotCustomSchemaMap = customizePluginSchemaMap
			gotDataType = dataType
			return captureValidator{}, nil
		},
	)

	validator, err := NewDatabasePayloadValidator(
		constant.APISIXVersion313,
		constant.Route,
		customSchemaMap,
	)

	assert.NoError(t, err)
	assert.NotNil(t, validator)
	assert.Equal(t, constant.APISIXVersion313, gotSchemaVersion)
	assert.Equal(t, "main.route", gotSchemaPath)
	assert.Equal(t, constant.APISIXVersion313, gotJSONVersion)
	assert.Equal(t, constant.Route, gotResourceType)
	assert.Equal(t, "main.route", gotJSONPath)
	assert.Equal(t, customSchemaMap, gotCustomSchemaMap)
	assert.Equal(t, constant.DATABASE, gotDataType)
}

func TestDatabasePayloadValidatorValidateRunsSchemaThenJSONSchema(t *testing.T) {
	var validateCalls []string
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(
		schemax.NewAPISIXSchemaValidator,
		func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
			return captureValidator{
				validate: func(raw json.RawMessage) error {
					validateCalls = append(validateCalls, "schema")
					return nil
				},
			}, nil
		},
	)
	patches.ApplyFunc(
		schemax.NewAPISIXJsonSchemaValidator,
		func(
			version constant.APISIXVersion,
			resourceType constant.APISIXResource,
			jsonPath string,
			customizePluginSchemaMap map[string]any,
			dataType constant.DataType,
		) (schemax.Validator, error) {
			return captureValidator{
				validate: func(raw json.RawMessage) error {
					validateCalls = append(validateCalls, "json_schema")
					return nil
				},
			}, nil
		},
	)

	validator, err := NewDatabasePayloadValidator(
		constant.APISIXVersion313,
		constant.Route,
		nil,
	)
	if !assert.NoError(t, err) {
		return
	}
	err = validator.Validate(json.RawMessage(`{"uri":"/demo"}`))

	assert.NoError(t, err)
	assert.Equal(t, []string{"schema", "json_schema"}, validateCalls)
}

func TestDatabasePayloadValidatorValidateStopsWhenSchemaValidationFails(t *testing.T) {
	expectedErr := errors.New("schema validate failed")
	jsonValidateCalled := false
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(
		schemax.NewAPISIXSchemaValidator,
		func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
			return captureValidator{
				validate: func(raw json.RawMessage) error {
					return expectedErr
				},
			}, nil
		},
	)
	patches.ApplyFunc(
		schemax.NewAPISIXJsonSchemaValidator,
		func(
			version constant.APISIXVersion,
			resourceType constant.APISIXResource,
			jsonPath string,
			customizePluginSchemaMap map[string]any,
			dataType constant.DataType,
		) (schemax.Validator, error) {
			return captureValidator{
				validate: func(raw json.RawMessage) error {
					jsonValidateCalled = true
					return nil
				},
			}, nil
		},
	)

	validator, err := NewDatabasePayloadValidator(
		constant.APISIXVersion313,
		constant.Route,
		nil,
	)
	if !assert.NoError(t, err) {
		return
	}
	err = validator.Validate(json.RawMessage(`{"uri":"/demo"}`))

	var stageErr *ValidationStageError
	assert.ErrorAs(t, err, &stageErr)
	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, ValidationStageResourceSchemaValidate, stageErr.Stage)
	assert.False(t, jsonValidateCalled)
}

func TestDatabasePayloadValidatorReturnsStageErrors(t *testing.T) {
	tests := []struct {
		name      string
		patch     func(*gomonkey.Patches, error)
		validate  bool
		wantStage ValidationStage
	}{
		{
			name: "resource schema build error",
			patch: func(patches *gomonkey.Patches, expectedErr error) {
				patches.ApplyFunc(
					schemax.NewAPISIXSchemaValidator,
					func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
						return nil, expectedErr
					},
				)
			},
			wantStage: ValidationStageResourceSchemaBuild,
		},
		{
			name: "json schema build error",
			patch: func(patches *gomonkey.Patches, expectedErr error) {
				patches.ApplyFunc(
					schemax.NewAPISIXSchemaValidator,
					func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
						return captureValidator{}, nil
					},
				)
				patches.ApplyFunc(
					schemax.NewAPISIXJsonSchemaValidator,
					func(
						version constant.APISIXVersion,
						resourceType constant.APISIXResource,
						jsonPath string,
						customizePluginSchemaMap map[string]any,
						dataType constant.DataType,
					) (schemax.Validator, error) {
						return nil, expectedErr
					},
				)
			},
			wantStage: ValidationStageJSONSchemaBuild,
		},
		{
			name: "json schema validate error",
			patch: func(patches *gomonkey.Patches, expectedErr error) {
				patches.ApplyFunc(
					schemax.NewAPISIXSchemaValidator,
					func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
						return captureValidator{}, nil
					},
				)
				patches.ApplyFunc(
					schemax.NewAPISIXJsonSchemaValidator,
					func(
						version constant.APISIXVersion,
						resourceType constant.APISIXResource,
						jsonPath string,
						customizePluginSchemaMap map[string]any,
						dataType constant.DataType,
					) (schemax.Validator, error) {
						return captureValidator{
							validate: func(raw json.RawMessage) error {
								return expectedErr
							},
						}, nil
					},
				)
			},
			validate:  true,
			wantStage: ValidationStageJSONSchemaValidate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedErr := errors.New("database payload validator failed")
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			tt.patch(patches, expectedErr)

			validator, err := NewDatabasePayloadValidator(
				constant.APISIXVersion313,
				constant.Route,
				nil,
			)
			if tt.validate && err == nil {
				err = validator.Validate(json.RawMessage(`{"uri":"/demo"}`))
			}

			var stageErr *ValidationStageError
			assert.ErrorAs(t, err, &stageErr)
			assert.ErrorIs(t, err, expectedErr)
			assert.Equal(t, tt.wantStage, stageErr.Stage)
		})
	}
}
