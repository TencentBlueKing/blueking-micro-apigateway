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

// SyncedItemListRequestRequest ...
type SyncedItemListRequestRequest struct {
	ID           string                  `json:"id,omitempty" form:"id"`
	ResourceType constant.APISIXResource `json:"resource_type,omitempty" form:"resource_type"` // 资源类型：route/upstream
	Name         string                  `json:"name,omitempty" form:"name"`
	Status       constant.SyncStatus     `json:"status,omitempty" form:"status"` // 同步状态：success/miss
	OrderBy      string                  `json:"order_by" form:"order_by"`
	Offset       int                     `json:"offset" form:"offset"`
	Limit        int                     `json:"limit" form:"limit"`
}

// SyncDataListResponse nolint:staticcheck
// swagger:response serializer.SyncDataListResponse
// SyncDataListResponse 同步资源列表
type SyncDataListResponse []SyncDataOutputInfo

// SyncDataOutputInfo ...
type SyncDataOutputInfo struct {
	ID            string                  `json:"id"`
	GatewayID     int                     `json:"gateway_id"`                  // 网关ID
	Name          string                  `json:"name"`                        // 资源名称
	ResourceType  constant.APISIXResource `json:"resource_type"`               // 资源类型
	ModeRevision  int                     `json:"mode_revision"`               // 同步版本
	Config        json.RawMessage         `json:"config" swaggertype:"object"` // 同步资源配置
	Status        constant.SyncStatus     `json:"status"`                      // 同步状态:success/miss
	PublishSource string                  `json:"publish_source"`              // 发布来源: "bk_micro(为网关)/others"
	CreatedAt     int64                   `json:"created_at"`
	UpdatedAt     int64                   `json:"updated_at"` // 同步时间
}

// SyncedTimeOutputInfo ...
type SyncedTimeOutputInfo struct {
	LatestTime int64 `json:"latest_time"` // 同步时间
}
