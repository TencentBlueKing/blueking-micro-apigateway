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

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// OperationAuditLog 操作审计表
type OperationAuditLog struct {
	ID            int                    `gorm:"primaryKey;autoIncrement" json:"id"`
	GatewayID     int                    `gorm:"column:gateway_id" json:"gateway_id"`
	CreatedAt     time.Time              `gorm:"column:created_at" json:"created_at"`
	OperationType constant.OperationType `gorm:"column:operation_type;type:varchar(64);not null" json:"operation_type"`
	Operator      string                 `gorm:"column:operator;type:varchar(50)" json:"operator"`
	// 资源id，多个用逗号分隔
	ResourceIDs string `gorm:"column:resource_ids;type:text" json:"resource_ids"`
	// route/service/upstream
	ResourceType constant.APISIXResource `gorm:"column:resource_type"`
	DataBefore   datatypes.JSON          `gorm:"type:json" json:"data_before"`
	DataAfter    datatypes.JSON          `gorm:"type:json" json:"data_after"`
}

// BatchOperationData 批量操data格式
type BatchOperationData struct {
	ID     string                  `json:"id"`
	Status constant.ResourceStatus `json:"status"`
	Config json.RawMessage         `json:"config,omitempty"`
}

// TableName 定义表名
func (OperationAuditLog) TableName() string {
	return "operation_audit_log"
}

// 定义一个通用的回调
func auditCallback(db *gorm.DB, gatewayID int, resourceID string, operator string,
	status constant.ResourceStatus, operationType constant.OperationType, resourceType constant.APISIXResource,
	dataBefore datatypes.JSON, dataAfter datatypes.JSON,
) error {
	var dataBeforeList []BatchOperationData
	var dataAfterList []BatchOperationData

	if operationType != constant.OperationTypeCreate {
		dataBeforeList = append(dataBeforeList, BatchOperationData{
			ID:     resourceID,
			Status: status,
			Config: json.RawMessage(dataBefore),
		})
	}

	if operationType != constant.OperationTypeDelete {
		dataAfterList = append(dataAfterList, BatchOperationData{
			ID:     resourceID,
			Status: status,
			Config: json.RawMessage(dataAfter),
		})
	}

	dataBeforeRaw, err := json.Marshal(dataBeforeList)
	if err != nil {
		return errors.Wrap(err, "marshal dataBefore failed")
	}

	dataAfterRaw, err := json.Marshal(dataAfterList)
	if err != nil {
		return errors.Wrap(err, "marshal dataAfter failed")
	}

	log := OperationAuditLog{
		GatewayID:     gatewayID,
		CreatedAt:     time.Now(),
		OperationType: operationType,
		Operator:      operator,
		ResourceIDs:   resourceID,
		ResourceType:  resourceType,
		DataBefore:    dataBeforeRaw,
		DataAfter:     dataAfterRaw,
	}
	if result := db.Create(&log); result.Error != nil {
		return result.Error
	}
	return nil
}
