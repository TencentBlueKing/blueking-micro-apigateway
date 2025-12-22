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

package model

import (
	"gorm.io/datatypes"
)

// GatewayReleaseVersion 表示数据库中的 gateway_release_version 表
type GatewayReleaseVersion struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement"` // 自增 ID
	GatewayID   string         `gorm:"column:gateway_id;type:varchar(32)"` // 对应网关 ID
	ReleaseData datatypes.JSON `gorm:"column:release_data"`                // 全量生效的资源数据 (JSON 格式)
	Version     string         `gorm:"column:version;type:varchar(32)"`    // 对应的版本号
	BaseModel
}

// TableName 设置表名
func (GatewayReleaseVersion) TableName() string {
	return "gateway_release_version"
}
