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

// Package auditlog contains shared helpers for writing operation audit rows.
package auditlog

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// AddBatchAuditLog writes operation audit rows for batch status-only changes.
func AddBatchAuditLog(
	ctx context.Context,
	operationType constant.OperationType,
	resourceType constant.APISIXResource,
	resources []*model.ResourceCommonModel,
	resourceIDStatusAfterMap map[string]constant.ResourceStatus,
) error {
	if len(resources) == 0 {
		return nil
	}

	var dataBefore []model.BatchOperationData
	var dataAfter []model.BatchOperationData
	var resourceIDs []string
	for _, resource := range resources {
		resourceIDs = append(resourceIDs, resource.ID)
		dataBefore = append(dataBefore, model.BatchOperationData{
			ID:     resource.ID,
			Status: resource.Status,
			Config: json.RawMessage(resource.Config),
		})
		if operationType != constant.OperationTypeDelete {
			dataAfter = append(dataAfter, model.BatchOperationData{
				ID:     resource.ID,
				Status: resourceIDStatusAfterMap[resource.ID],
				Config: json.RawMessage(resource.Config),
			})
		}
	}

	dataBeforeRaw, err := json.Marshal(dataBefore)
	if err != nil {
		return errors.Wrap(err, "marshal dataBefore failed")
	}
	dataAfterRaw, err := json.Marshal(dataAfter)
	if err != nil {
		return errors.Wrap(err, "marshal dataAfter failed")
	}

	operationAuditLog := &model.OperationAuditLog{
		GatewayID:     ginx.GetGatewayInfoFromContext(ctx).ID,
		ResourceType:  resourceType,
		OperationType: operationType,
		ResourceIDs:   strings.Join(resourceIDs, ","),
		DataBefore:    dataBeforeRaw,
		DataAfter:     dataAfterRaw,
		Operator:      ginx.GetUserIDFromContext(ctx),
	}
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).OperationAuditLog.WithContext(ctx).Create(operationAuditLog)
	}
	return repo.OperationAuditLog.WithContext(ctx).Create(operationAuditLog)
}

// AddRevertAuditLog writes operation audit rows for revert operations.
func AddRevertAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceIDs []string,
	beforeResources []*model.ResourceCommonModel,
	afterResources []*model.ResourceCommonModel,
) error {
	var dataBefore []model.BatchOperationData
	var dataAfter []model.BatchOperationData
	for _, resource := range beforeResources {
		dataBefore = append(dataBefore, model.BatchOperationData{
			ID:     resource.ID,
			Status: resource.Status,
			Config: json.RawMessage(resource.Config),
		})
	}
	for _, resource := range afterResources {
		dataAfter = append(dataAfter, model.BatchOperationData{
			ID:     resource.ID,
			Status: resource.Status,
			Config: json.RawMessage(resource.Config),
		})
	}

	dataBeforeRaw, err := json.Marshal(dataBefore)
	if err != nil {
		return errors.Wrap(err, "marshal dataBefore failed")
	}
	dataAfterRaw, err := json.Marshal(dataAfter)
	if err != nil {
		return errors.Wrap(err, "marshal dataAfter failed")
	}

	operationAuditLog := &model.OperationAuditLog{
		GatewayID:     ginx.GetGatewayInfoFromContext(ctx).ID,
		ResourceType:  resourceType,
		OperationType: constant.OperationTypeRevert,
		ResourceIDs:   strings.Join(resourceIDs, ","),
		DataBefore:    dataBeforeRaw,
		DataAfter:     dataAfterRaw,
		Operator:      ginx.GetUserIDFromContext(ctx),
	}
	if ginx.GetTx(ctx) != nil {
		return ginx.GetTx(ctx).OperationAuditLog.WithContext(ctx).Create(operationAuditLog)
	}
	return repo.OperationAuditLog.WithContext(ctx).Create(operationAuditLog)
}
