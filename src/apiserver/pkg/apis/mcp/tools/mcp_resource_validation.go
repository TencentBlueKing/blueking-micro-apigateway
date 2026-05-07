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
	"fmt"

	resourcevalidationbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resourcevalidation"
	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func validateMCPDatabaseResourceConfig(
	ctx context.Context,
	version constant.APISIXVersion,
	resourceType constant.APISIXResource,
	rawConfig json.RawMessage,
) error {
	customizePluginSchemaMap, err := schemabiz.GetCustomizePluginSchemaMap(ctx)
	if err != nil {
		return fmt.Errorf("get customize plugin schema map failed: %w", err)
	}

	databaseValidator, err := resourcevalidationbiz.NewDatabasePayloadValidator(
		version,
		resourceType,
		customizePluginSchemaMap,
	)
	if err != nil {
		return err
	}
	if err = databaseValidator.Validate(rawConfig); err != nil {
		return err
	}

	return nil
}
