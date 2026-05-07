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

// Package resourcevalidation owns shared APISIX resource config validation orchestration.
package resourcevalidation

import (
	"context"
	"encoding/json"
	"fmt"

	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// Input describes one prepared DATABASE validation payload.
type Input struct {
	Version                constant.APISIXVersion
	ResourceType           constant.APISIXResource
	ResourceIdentification string
	RawConfig              json.RawMessage
}

// ValidateDatabaseResourceConfig validates a resource config as a DATABASE payload.
func ValidateDatabaseResourceConfig(ctx context.Context, input Input) error {
	resourceIdentification := input.ResourceIdentification
	if resourceIdentification == "" {
		resourceIdentification = schemax.GetResourceIdentification(input.RawConfig)
	}

	schemaValidator, err := schemax.NewAPISIXSchemaValidator(
		input.Version,
		"main."+input.ResourceType.String(),
	)
	if err != nil {
		return fmt.Errorf("new APISIX schema validator failed, resource:%s validate failed: %w",
			resourceIdentification, err)
	}
	if err = schemaValidator.Validate(input.RawConfig); err != nil {
		return fmt.Errorf("resource:%s validate failed: %w", resourceIdentification, err)
	}

	customizePluginSchemaMap, err := schemabiz.GetCustomizePluginSchemaMap(ctx)
	if err != nil {
		return fmt.Errorf("get customize plugin schema map failed: %w", err)
	}
	jsonConfigValidator, err := schemax.NewAPISIXJsonSchemaValidator(
		input.Version,
		input.ResourceType,
		"main."+input.ResourceType.String(),
		customizePluginSchemaMap,
		constant.DATABASE,
	)
	if err != nil {
		return fmt.Errorf("new APISIX json schema validator failed, resource:%s validate failed: %w",
			resourceIdentification, err)
	}
	if err = jsonConfigValidator.Validate(input.RawConfig); err != nil {
		return fmt.Errorf("resource config:%s validate failed, err: %w", input.RawConfig, err)
	}

	return nil
}
