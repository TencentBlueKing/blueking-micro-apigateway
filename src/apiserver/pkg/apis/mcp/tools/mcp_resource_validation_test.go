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

package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	resourcevalidationbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resourcevalidation"
	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

type mcpCaptureDatabasePayloadValidator struct {
	validate func(json.RawMessage) error
}

func (v mcpCaptureDatabasePayloadValidator) Validate(rawConfig json.RawMessage) error {
	if v.validate != nil {
		return v.validate(rawConfig)
	}
	return nil
}

func TestValidateMCPDatabaseResourceConfigBuildsSharedValidatorWithCustomSchemaMap(t *testing.T) {
	customSchemaMap := map[string]any{"custom-plugin": map[string]any{"type": "object"}}
	rawConfig := json.RawMessage(`{"id":"route-1","name":"route-1","uris":["/route-1"]}`)
	var gotVersion constant.APISIXVersion
	var gotResourceType constant.APISIXResource
	var gotCustomSchemaMap map[string]any
	var gotPayload json.RawMessage

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(
		schemabiz.GetCustomizePluginSchemaMap,
		func(ctx context.Context) (map[string]any, error) {
			return customSchemaMap, nil
		},
	)
	patches.ApplyFunc(
		resourcevalidationbiz.NewDatabasePayloadValidator,
		func(
			version constant.APISIXVersion,
			resourceType constant.APISIXResource,
			customizePluginSchemaMap map[string]any,
		) (resourcevalidationbiz.DatabasePayloadValidator, error) {
			gotVersion = version
			gotResourceType = resourceType
			gotCustomSchemaMap = customizePluginSchemaMap
			return mcpCaptureDatabasePayloadValidator{
				validate: func(raw json.RawMessage) error {
					gotPayload = append(json.RawMessage(nil), raw...)
					return nil
				},
			}, nil
		},
	)

	err := validateMCPDatabaseResourceConfig(
		context.Background(),
		constant.APISIXVersion313,
		constant.Route,
		rawConfig,
	)

	assert.NoError(t, err)
	assert.Equal(t, constant.APISIXVersion313, gotVersion)
	assert.Equal(t, constant.Route, gotResourceType)
	assert.Equal(t, customSchemaMap, gotCustomSchemaMap)
	assert.JSONEq(t, string(rawConfig), string(gotPayload))
}

func TestValidateMCPDatabaseResourceConfigReturnsSharedValidatorError(t *testing.T) {
	expectedErr := errors.New("validation failed")

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(
		schemabiz.GetCustomizePluginSchemaMap,
		func(ctx context.Context) (map[string]any, error) {
			return map[string]any{}, nil
		},
	)
	patches.ApplyFunc(
		resourcevalidationbiz.NewDatabasePayloadValidator,
		func(
			version constant.APISIXVersion,
			resourceType constant.APISIXResource,
			customizePluginSchemaMap map[string]any,
		) (resourcevalidationbiz.DatabasePayloadValidator, error) {
			return mcpCaptureDatabasePayloadValidator{
				validate: func(raw json.RawMessage) error {
					return expectedErr
				},
			}, nil
		},
	)

	err := validateMCPDatabaseResourceConfig(
		context.Background(),
		constant.APISIXVersion313,
		constant.Route,
		json.RawMessage(`{"id":"route-1"}`),
	)

	assert.ErrorIs(t, err, expectedErr)
}
