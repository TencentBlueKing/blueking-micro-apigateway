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

	validator "github.com/go-playground/validator/v10"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// SyncRequest ...
type SyncRequest struct {
	ResourceType constant.APISIXResource `json:"resource_type"` // 资源类型：route/upstream/...如果为空，则同步所有资源
}

// SyncResponse ...
type SyncResponse map[constant.APISIXResource]int

// RevertRequest ...
type RevertRequest struct {
	ResourceType   constant.APISIXResource `json:"resource_type" binding:"required"`    // 资源类型：route/upstream/...
	ResourceIDList []string                `json:"resource_id_list" binding:"required"` // 资源ID列表
}

// DeleteRequest ...
type DeleteRequest struct {
	ResourceType   constant.APISIXResource `json:"resource_type" binding:"required"`    // 资源类型：route/upstream/...
	ResourceIDList []string                `json:"resource_id_list" binding:"required"` // 资源ID列表
}

// ResourceManagedRequest ...
type ResourceManagedRequest struct {
	ResourceIDList []string `json:"resource_id_list"` // 资源ID列表，不传则同步所有资源
}

// ResourceManagedResponse ...
type ResourceManagedResponse map[constant.APISIXResource]int

// ResourceDiffAllRequest ...
type ResourceDiffAllRequest struct {
	ID            string                   `json:"id"`                                                 // 资源ID
	Name          string                   `json:"name"`                                               // 资源名称
	ResourceType  constant.APISIXResource  `json:"resource_type"`                                      // 资源类型
	OperationType []constant.OperationType `json:"operation_type" binding:"resourceDiffOperationType"` // 操作类型
}

// ResourceDiffRequest ...
type ResourceDiffRequest struct {
	ResourceIDList []string                 `json:"resource_id_list"`                                   // 资源ID列表, 比对必须传入ID
	Name           string                   `json:"name"`                                               // 资源名称
	OperationType  []constant.OperationType `json:"operation_type" binding:"resourceDiffOperationType"` // 操作类型
}

// ResourceDiffResponse ...
type ResourceDiffResponse []dto.ResourceChangeInfo

// EtcdExportOutput ...
type EtcdExportOutput map[constant.APISIXResource][]ResourceInfo

// ResourceInfo ...
type ResourceInfo struct {
	ResourceType constant.APISIXResource `json:"resource_type,omitempty"`               // 资源类型
	ResourceID   string                  `json:"resource_id,omitempty"`                 // 资源ID
	Name         string                  `json:"name,omitempty"`                        // 资源名称
	Config       json.RawMessage         `json:"config,omitempty" swaggertype:"object"` // 资源配置
	Status       constant.UploadStatus   `json:"status,omitempty"`                      // 资源导入状态(add/update)
}

type ResourceUploadInfo struct {
	Adds   map[constant.APISIXResource][]ResourceInfo `json:"add,omitempty"`
	Update map[constant.APISIXResource][]ResourceInfo `json:"update,omitempty"`
}

// OperationTypeToResourceStatus 操作类型转换资源状态
func OperationTypeToResourceStatus(operationType []constant.OperationType) []constant.ResourceStatus {
	var resourceStatus []constant.ResourceStatus
	for _, t := range operationType {
		switch t {
		case constant.OperationTypeUpdate:
			resourceStatus = append(resourceStatus, constant.ResourceStatusUpdateDraft)
		case constant.OperationTypeCreate:
			resourceStatus = append(resourceStatus, constant.ResourceStatusCreateDraft)
		case constant.OperationTypeDelete:
			resourceStatus = append(resourceStatus, constant.ResourceStatusDeleteDraft)
		}
	}
	return resourceStatus
}

// CheckResourceDiffOperationType 校验操作类型
func CheckResourceDiffOperationType(fl validator.FieldLevel) bool {
	value := fl.Field().Interface().([]constant.OperationType)
	resourceDiffOperationTypeMap := map[constant.OperationType]bool{
		constant.OperationTypeUpdate: true,
		constant.OperationTypeCreate: true,
		constant.OperationTypeDelete: true,
	}
	for _, v := range value {
		if _, ok := resourceDiffOperationTypeMap[v]; !ok {
			return false
		}
	}
	return true
}

func init() {
	validation.AddBizFieldTagValidator(
		"resourceDiffOperationType",
		CheckResourceDiffOperationType,
		"{0}:{1} must be update,create,delete",
	)
}
