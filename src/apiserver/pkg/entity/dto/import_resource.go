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

package dto

import (
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// ImportResourceInfo ...
type ImportResourceInfo struct {
	ResourceType constant.APISIXResource `json:"resource_type,omitempty"`               // 资源类型
	ResourceID   string                  `json:"resource_id,omitempty"`                 // 资源 ID
	Name         string                  `json:"name,omitempty"`                        // 资源名称
	Config       json.RawMessage         `json:"config,omitempty" swaggertype:"object"` // 资源配置
	Status       constant.UploadStatus   `json:"status,omitempty"`                      // 资源导入状态 (add/update)
}

// GetResourceKey 获取资源 key
func (r ImportResourceInfo) GetResourceKey() string {
	if r.ResourceType == constant.PluginMetadata {
		return fmt.Sprintf(constant.ResourceKeyFormat, r.ResourceType, r.Name)
	}
	return fmt.Sprintf(constant.ResourceKeyFormat, r.ResourceType, r.ResourceID)
}

// ImportUploadInfo ...
type ImportUploadInfo struct {
	Add    map[constant.APISIXResource][]*ImportResourceInfo `json:"add,omitempty"`
	Update map[constant.APISIXResource][]*ImportResourceInfo `json:"update,omitempty"`
}

// ImportMetadata ...
type ImportMetadata struct {
	IgnoreFields map[constant.APISIXResource][]string `json:"ignore_fields"`
}
