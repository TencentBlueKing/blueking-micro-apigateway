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

// Package status ...
package status

import (
	"context"
	"errors"
	"fmt"

	"github.com/looplab/fsm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ResourceStatusOp ...
type ResourceStatusOp struct {
	resourceInfo model.ResourceCommonModel
	fsm          *fsm.FSM
}

var events = []fsm.EventDesc{
	//  create-> create_draft
	{
		Name: constant.OperationTypeCreate.String(),
		Src:  []string{""},
		Dst:  constant.ResourceStatusCreateDraft.String(),
	},
	// delete-> delete_draft
	{
		Name: constant.OperationTypeDelete.String(),
		Src:  []string{constant.ResourceStatusSuccess.String()},
		Dst:  constant.ResourceStatusDeleteDraft.String(),
	},
	// create_draft-> delete
	{
		Name: constant.OperationTypeDelete.String(),
		Src:  []string{constant.ResourceStatusCreateDraft.String()},
		Dst:  "",
	},
	// update-> update_draft
	{
		Name: constant.OperationTypeUpdate.String(),
		Src: []string{
			constant.ResourceStatusSuccess.String(),
			constant.ResourceStatusUpdateDraft.String(),
		},
		Dst: constant.ResourceStatusUpdateDraft.String(),
	},
	// create_draft update-> create_draft
	{
		Name: constant.OperationTypeUpdate.String(),
		Src: []string{
			constant.ResourceStatusCreateDraft.String(),
		},
		Dst: constant.ResourceStatusCreateDraft.String(),
	},
	// revert-> success
	{
		Name: constant.OperationTypeRevert.String(),
		Src: []string{
			constant.ResourceStatusUpdateDraft.String(),
			constant.ResourceStatusDeleteDraft.String(),
		},
		Dst: constant.ResourceStatusSuccess.String(),
	},
	// publish-> success
	{
		Name: constant.OperationTypePublish.String(),
		Src: []string{
			constant.ResourceStatusUpdateDraft.String(),
			constant.ResourceStatusCreateDraft.String(),
			constant.ResourceStatusDeleteDraft.String(),
		},
		Dst: string(constant.ResourceStatusSuccess),
	},
}

// NewResourceStatusOp ...
func NewResourceStatusOp(resourceInfo model.ResourceCommonModel) *ResourceStatusOp {
	d := &ResourceStatusOp{
		resourceInfo: resourceInfo,
	}
	d.fsm = fsm.NewFSM(
		resourceInfo.Status.String(),
		events,
		fsm.Callbacks{},
	)
	return d
}

// CanDo 判断是否可以进行操作
func (s *ResourceStatusOp) CanDo(ctx context.Context, operationType constant.OperationType) error {
	// demo 站点进行特殊处理，部分资源不允许进行任何操作
	if config.IsDemoMode() {
		if config.G.Biz.DemoProtectResources[s.resourceInfo.ID] {
			msg := fmt.Sprintf("%s。该资源处于被保护状态禁止修改，如需体验请新建对应资源后操作", config.G.Service.DemoModeWarnMsg)
			return errors.New(msg)
		}
	}
	// 如果网关是只读模式，则不允许进行任何操作
	if ginx.GetGatewayInfoFromContext(ctx) != nil && ginx.GetGatewayInfoFromContext(ctx).ReadOnly {
		return errors.New("网关只读模式，不允许进行任何变更操作")
	}

	if s.ignoreSpecialOp(operationType) {
		return nil
	}
	return s.fsm.Event(ctx, operationType.String())
}

// NextStatus 获取下一个状态
func (s *ResourceStatusOp) NextStatus(
	ctx context.Context,
	operationType constant.OperationType,
) (constant.ResourceStatus, error) {
	if s.ignoreSpecialOp(operationType) {
		return s.resourceInfo.Status, nil
	}
	err := s.fsm.Event(ctx, operationType.String())
	if err != nil {
		return "", err
	}
	return constant.ResourceStatus(s.fsm.Current()), nil
}

// ignoreSpecialOp 判断是否需要忽略特殊操作
func (s *ResourceStatusOp) ignoreSpecialOp(operationType constant.OperationType) bool {
	// fms 不支持同状态之间的转换
	if operationType == constant.OperationTypeUpdate &&
		(s.resourceInfo.Status == constant.ResourceStatusCreateDraft ||
			s.resourceInfo.Status == constant.ResourceStatusUpdateDraft) {
		return true
	}
	return false
}
