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
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

// prepareMCPCreateConfig marshals the inbound MCP create payload and injects
// the outer name field according to the resource type. Consumers use
// "username"; all other MCP resources use "name".
//
// The trailing gjson.Exists check is preserved from the original inline
// createResourceHandler implementation as a defensive guard.
func prepareMCPCreateConfig(
	resourceType constant.APISIXResource,
	inputConfig any,
	name string,
) ([]byte, error) {
	config, err := json.Marshal(inputConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	nameKey := model.GetResourceNameKey(resourceType)
	config, err = sjson.SetBytes(config, nameKey, name)
	if err != nil {
		return nil, fmt.Errorf("failed to inject name into config: %w", err)
	}
	if !gjson.GetBytes(config, nameKey).Exists() {
		return nil, fmt.Errorf("name field not found in config after injection")
	}

	return config, nil
}

// prepareMCPUpdateConfig mirrors prepareMCPCreateConfig but keeps the original
// config untouched when the caller does not provide a new outer name.
//
// Behavior change vs. the previous inline updateResourceHandler code:
// sjson.SetBytes errors are now returned to the caller instead of being
// silently ignored.
func prepareMCPUpdateConfig(
	resourceType constant.APISIXResource,
	inputConfig any,
	name string,
) ([]byte, error) {
	config, err := json.Marshal(inputConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	if name == "" {
		return config, nil
	}

	nameKey := model.GetResourceNameKey(resourceType)
	config, err = sjson.SetBytes(config, nameKey, name)
	if err != nil {
		return nil, fmt.Errorf("failed to inject name into config: %w", err)
	}

	return config, nil
}

// MCP stays local for now: we only deduplicate MCP's own config prep and
// draft assembly. It does not join the cross-domain abstraction track unless a
// later change proves very high leverage.
//
// Intentional signature drift vs. web's buildWebCreateDraft(c *gin.Context,
// ...): MCP is invoked from an MCP tool handler with no gin.Context in scope,
// so the helper takes gatewayID directly. Do not align this signature with the
// web helper.
func buildMCPCreateDraft(
	gatewayID int,
	resourceID string,
	config []byte,
) model.ResourceCommonModel {
	return model.ResourceCommonModel{
		ID:        resourceID,
		GatewayID: gatewayID,
		Config:    datatypes.JSON(config),
		Status:    constant.ResourceStatusCreateDraft,
		BaseModel: model.BaseModel{
			Creator: "mcp",
			Updater: "mcp",
		},
	}
}
