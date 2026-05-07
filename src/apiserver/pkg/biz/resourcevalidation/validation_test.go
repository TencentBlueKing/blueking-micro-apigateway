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
	"context"
	"encoding/json"
	"errors"
	"testing"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

type captureValidator struct {
	validate func(json.RawMessage) error
}

func (v captureValidator) Validate(raw json.RawMessage) error {
	if v.validate != nil {
		return v.validate(raw)
	}
	return nil
}

func TestValidateDatabaseResourceConfigAcceptsValidRoute(t *testing.T) {
	ctx := newResourceValidationTestContext()
	route := data.Route1WithNoRelationResource(&model.Gateway{ID: 1}, constant.ResourceStatusCreateDraft)

	err := ValidateDatabaseResourceConfig(ctx, Input{
		Version:                constant.APISIXVersion313,
		ResourceType:           constant.Route,
		ResourceIdentification: "route-validation-name",
		RawConfig:              json.RawMessage(route.Config),
	})

	assert.NoError(t, err)
}

func TestValidateDatabaseResourceConfigRejectsInvalidRoute(t *testing.T) {
	ctx := newResourceValidationTestContext()

	err := ValidateDatabaseResourceConfig(ctx, Input{
		Version:                constant.APISIXVersion313,
		ResourceType:           constant.Route,
		ResourceIdentification: "route-validation-name",
		RawConfig:              json.RawMessage(`{"uri":123}`),
	})

	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "validate failed")
}

func TestValidateDatabaseResourceConfigDoesNotRewritePayload(t *testing.T) {
	ctx := newResourceValidationTestContext()
	rawConfig := json.RawMessage(`{"id":"consumer-id","username":"consumer-demo","name":"keep-me","plugins":{}}`)
	originalRaw := append(json.RawMessage(nil), rawConfig...)
	var schemaPayload json.RawMessage
	var jsonSchemaPayload json.RawMessage
	patches := patchResourceValidationValidators(
		t,
		func(raw json.RawMessage) error {
			schemaPayload = append(json.RawMessage(nil), raw...)
			return nil
		},
		func(raw json.RawMessage) error {
			jsonSchemaPayload = append(json.RawMessage(nil), raw...)
			return nil
		},
	)
	defer patches.Reset()

	err := ValidateDatabaseResourceConfig(ctx, Input{
		Version:                constant.APISIXVersion313,
		ResourceType:           constant.Consumer,
		ResourceIdentification: "consumer-demo",
		RawConfig:              rawConfig,
	})

	if !assert.NoError(t, err) {
		return
	}
	assert.JSONEq(t, string(originalRaw), string(schemaPayload))
	assert.JSONEq(t, string(originalRaw), string(jsonSchemaPayload))
	assert.JSONEq(t, string(originalRaw), string(rawConfig))
}

func TestValidateDatabaseResourceConfigReturnsCustomSchemaMapError(t *testing.T) {
	ctx := newResourceValidationTestContext()
	expectedErr := errors.New("custom schema lookup failed")
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(
		schemax.NewAPISIXSchemaValidator,
		func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
			return captureValidator{}, nil
		},
	)
	patches.ApplyFunc(
		schemabiz.GetCustomizePluginSchemaMap,
		func(ctx context.Context) (map[string]any, error) {
			return nil, expectedErr
		},
	)

	err := ValidateDatabaseResourceConfig(ctx, Input{
		Version:                constant.APISIXVersion313,
		ResourceType:           constant.Route,
		ResourceIdentification: "route-validation-name",
		RawConfig:              json.RawMessage(`{"uri":"/demo"}`),
	})

	if !assert.Error(t, err) {
		return
	}
	assert.Contains(t, err.Error(), "get customize plugin schema map failed")
	assert.ErrorIs(t, err, expectedErr)
}

func TestValidateDatabaseResourceConfigReturnsValidatorCreationError(t *testing.T) {
	tests := []struct {
		name          string
		patch         func(*gomonkey.Patches, error)
		expectedError string
	}{
		{
			name: "schema validator creation fails",
			patch: func(patches *gomonkey.Patches, expectedErr error) {
				patches.ApplyFunc(
					schemax.NewAPISIXSchemaValidator,
					func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
						return nil, expectedErr
					},
				)
			},
			expectedError: "new APISIX schema validator failed",
		},
		{
			name: "json schema validator creation fails",
			patch: func(patches *gomonkey.Patches, expectedErr error) {
				patches.ApplyFunc(
					schemax.NewAPISIXSchemaValidator,
					func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
						return captureValidator{}, nil
					},
				)
				patches.ApplyFunc(
					schemabiz.GetCustomizePluginSchemaMap,
					func(ctx context.Context) (map[string]any, error) {
						return map[string]any{}, nil
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
			expectedError: "new APISIX json schema validator failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newResourceValidationTestContext()
			expectedErr := errors.New("validator creation failed")
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			tt.patch(patches, expectedErr)

			err := ValidateDatabaseResourceConfig(ctx, Input{
				Version:                constant.APISIXVersion313,
				ResourceType:           constant.Route,
				ResourceIdentification: "route-validation-name",
				RawConfig:              json.RawMessage(`{"uri":"/demo"}`),
			})

			if !assert.Error(t, err) {
				return
			}
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.ErrorIs(t, err, expectedErr)
		})
	}
}

func newResourceValidationTestContext() context.Context {
	util.InitEmbedDb()
	return ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{
		ID:            10001,
		APISIXVersion: string(constant.APISIXVersion313),
	})
}

func patchResourceValidationValidators(
	t *testing.T,
	onSchemaValidate func(json.RawMessage) error,
	onJSONSchemaValidate func(json.RawMessage) error,
) *gomonkey.Patches {
	t.Helper()

	patches := gomonkey.NewPatches()
	patches.ApplyFunc(
		schemax.NewAPISIXSchemaValidator,
		func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
			return captureValidator{validate: onSchemaValidate}, nil
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
			return captureValidator{validate: onJSONSchemaValidate}, nil
		},
	)
	patches.ApplyFunc(
		schemabiz.GetCustomizePluginSchemaMap,
		func(ctx context.Context) (map[string]any, error) {
			return map[string]any{}, nil
		},
	)

	return patches
}
