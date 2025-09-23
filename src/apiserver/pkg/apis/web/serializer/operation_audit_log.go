/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
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

package serializer

import (
	"encoding/json"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// OperationAuditLogListRequest 操作审计日志列表请求
type OperationAuditLogListRequest struct {
	Name          string                  `form:"name" json:"name,omitempty"`
	ResourceType  constant.APISIXResource `form:"resource_type" json:"resource_type,omitempty"`   // 资源类型
	ResourceID    string                  `form:"resource_id" json:"resource_id,omitempty"`       // 资源ID
	Operator      string                  `form:"operator" json:"operator,omitempty"`             // 操作人
	OperationType constant.OperationType  `form:"operation_type" json:"operation_type,omitempty"` // 操作类型
	TimeStart     int                     `form:"start_time" json:"start_time,omitempty"`
	TimeEnd       int                     `form:"end_time" json:"end_time,omitempty"`
	Offset        int                     `form:"offset" json:"offset,omitempty"`
	Limit         int                     `form:"limit" json:"limit,omitempty"`
}

// OperationAuditLogListResponse 操作审计日志列表响应
type OperationAuditLogListResponse struct {
	ID            int                     `json:"id"`                                // ID
	Names         []string                `json:"names"`                             // 名称列表
	ResourceIDs   []string                `json:"resource_ids"`                      // 资源ID列表
	ResourceType  constant.APISIXResource `json:"resource_type"`                     // 资源类型
	OperationType constant.OperationType  `json:"operation_type"`                    // 操作类型
	Operator      string                  `json:"operator"`                          // 操作人
	CreatedAt     int64                   `json:"created_at"`                        // 创建时间
	DataBefore    json.RawMessage         `json:"data_before" swaggertype:"object"`  // 操作前数据
	DataAfter     json.RawMessage         `json:"data_after"   swaggertype:"object"` // 操作后数据
}
