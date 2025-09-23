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

package model

import "gorm.io/datatypes"

// SystemConfig ...
type SystemConfig struct {
	ID    int64          `json:"id" gorm:"column:id;primaryKey;autoIncrement;not null"`
	Key   string         `json:"key" gorm:"column:key;type:varchar(255);uniqueIndex:idx_key"`
	Value datatypes.JSON `gorm:"column:config;type:json"` // route raw config
	BaseModel
}

// TableName 设置表名
func (SystemConfig) TableName() string {
	return "system_config"
}
