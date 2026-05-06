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

package importflow

import (
	"context"
	"encoding/json"
	"fmt"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// ValidateImportedResources validates imported resources with schema and
// association checks before they enter the upload transaction flow.
func ValidateImportedResources(
	ctx context.Context,
	resources map[constant.APISIXResource][]*model.GatewaySyncData,
	allResourceIDs map[string]struct{},
	allPluginSchemaMap map[string]any,
) error {
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	for resourceType, resource := range resources {
		schemaValidator, err := schema.NewAPISIXSchemaValidator(
			gatewayInfo.GetAPISIXVersionX(),
			"main."+resourceType.String(),
		)
		if err != nil {
			return err
		}
		jsonConfigValidator, err := schema.NewAPISIXJsonSchemaValidator(
			gatewayInfo.GetAPISIXVersionX(),
			resourceType,
			"main."+string(resourceType),
			allPluginSchemaMap,
			constant.DATABASE,
		)
		if err != nil {
			return err
		}
		for _, r := range resource {
			configRawForValidation := resourcebiz.BuildConfigRawForValidation(
				string(r.Config),
				resourceType,
				gatewayInfo.GetAPISIXVersionX(),
			)

			if err = schemaValidator.Validate(configRawForValidation); err != nil {
				logging.Errorf("schema validate failed, err: %v", err)
				return err
			}
			if err = jsonConfigValidator.Validate(configRawForValidation); err != nil {
				return fmt.Errorf("resource config:%s validate failed, err: %w", r.Config, err)
			}

			var resourceAssociateIDInfo dto.ResourceAssociateID
			err = json.Unmarshal(r.Config, &resourceAssociateIDInfo)
			if err != nil {
				return err
			}
			if resourceAssociateIDInfo.ServiceID != "" {
				if _, ok := allResourceIDs[resourceAssociateIDInfo.GetResourceKey(
					constant.Service,
					resourceAssociateIDInfo.ServiceID,
				)]; !ok {
					return fmt.Errorf(
						"associated service [id:%s] not found",
						resourceAssociateIDInfo.ServiceID,
					)
				}
			}
			if resourceAssociateIDInfo.UpstreamID != "" {
				if _, ok := allResourceIDs[resourceAssociateIDInfo.GetResourceKey(
					constant.Upstream,
					resourceAssociateIDInfo.UpstreamID,
				)]; !ok {
					return fmt.Errorf(
						"associated upstream [id:%s] not found",
						resourceAssociateIDInfo.UpstreamID,
					)
				}
			}
			if resourceAssociateIDInfo.PluginConfigID != "" {
				if _, ok := allResourceIDs[resourceAssociateIDInfo.GetResourceKey(
					constant.PluginConfig,
					resourceAssociateIDInfo.PluginConfigID,
				)]; !ok {
					return fmt.Errorf(
						"associated plugin_config [id:%s] not found",
						resourceAssociateIDInfo.PluginConfigID,
					)
				}
			}
			if resourceAssociateIDInfo.GroupID != "" {
				if _, ok := allResourceIDs[resourceAssociateIDInfo.GetResourceKey(
					constant.ConsumerGroup,
					resourceAssociateIDInfo.GroupID,
				)]; !ok {
					return fmt.Errorf(
						"associated consumer_group [id:%s] not found",
						resourceAssociateIDInfo.GroupID,
					)
				}
			}
		}
	}
	return nil
}
