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

package status

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

func TestNewResourceStatusOp(t *testing.T) {
	ctx := context.Background()
	type args struct {
		resourceInfo model.ResourceCommonModel
		op           constant.OperationType
		wantStatus   constant.ResourceStatus
		pass         bool
	}
	tests := []struct {
		name string
		args args
		want *ResourceStatusOp
	}{
		{
			name: "create_success",
			args: args{
				resourceInfo: model.ResourceCommonModel{},
				op:           constant.OperationTypeCreate,
				wantStatus:   constant.ResourceStatusCreateDraft,
				pass:         true,
			},
		},
		{
			name: "delete_success_with_success",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusSuccess},
				op:           constant.OperationTypeDelete,
				wantStatus:   constant.ResourceStatusDeleteDraft,
				pass:         true,
			},
		},
		{
			name: "delete_success_from_create_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusCreateDraft},
				op:           constant.OperationTypeDelete,
				wantStatus:   "",
				pass:         true,
			},
		},
		{
			name: "delete_fail_with_update_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusUpdateDraft},
				op:           constant.OperationTypeDelete,
				pass:         false,
			},
		},
		{
			name: "update_success_with_success",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusSuccess},
				op:           constant.OperationTypeUpdate,
				wantStatus:   constant.ResourceStatusUpdateDraft,
				pass:         true,
			},
		},
		{
			name: "update_success_with_update_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusUpdateDraft},
				op:           constant.OperationTypeUpdate,
				wantStatus:   constant.ResourceStatusUpdateDraft,
				pass:         true,
			},
		},
		{
			name: "update_success_with_create_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusCreateDraft},
				op:           constant.OperationTypeUpdate,
				wantStatus:   constant.ResourceStatusCreateDraft,
				pass:         true,
			},
		},
		{
			name: "update_fail",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusDeleteDraft},
				op:           constant.OperationTypeUpdate,
				pass:         false,
			},
		},
		{
			name: "revert_success_from_update_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusUpdateDraft},
				op:           constant.OperationTypeRevert,
				wantStatus:   constant.ResourceStatusSuccess,
				pass:         true,
			},
		},
		{
			name: "revert_success_from_delete_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusDeleteDraft},
				op:           constant.OperationTypeRevert,
				wantStatus:   constant.ResourceStatusSuccess,
				pass:         true,
			},
		},
		{
			name: "revert_fail",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusSuccess},
				op:           constant.OperationTypeRevert,
				pass:         false,
			},
		},
		{
			name: "publish_from_create_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusCreateDraft},
				op:           constant.OperationTypePublish,
				wantStatus:   constant.ResourceStatusSuccess,
				pass:         true,
			},
		},
		{
			name: "publish_from_delete_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusDeleteDraft},
				op:           constant.OperationTypePublish,
				wantStatus:   constant.ResourceStatusSuccess,
				pass:         true,
			},
		},
		{
			name: "publish_from_update_draft",
			args: args{
				resourceInfo: model.ResourceCommonModel{Status: constant.ResourceStatusUpdateDraft},
				op:           constant.OperationTypePublish,
				wantStatus:   constant.ResourceStatusSuccess,
				pass:         true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusOp := NewResourceStatusOp(tt.args.resourceInfo)
			err := statusOp.CanDo(ctx, tt.args.op)
			if tt.args.pass {
				assert.NoError(t, err)
				assert.Equal(t, tt.args.wantStatus, constant.ResourceStatus(statusOp.fsm.Current()))
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestResourceStatusOp_NextStatus(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus constant.ResourceStatus
		opType        constant.OperationType
		wantStatus    constant.ResourceStatus
		wantErr       bool
	}{
		{
			name:          "ignore update operation in create_draft status",
			currentStatus: constant.ResourceStatusCreateDraft,
			opType:        constant.OperationTypeUpdate,
			wantStatus:    constant.ResourceStatusCreateDraft,
			wantErr:       false,
		},
		{
			name:          "ignore update operation in update_draft status",
			currentStatus: constant.ResourceStatusUpdateDraft,
			opType:        constant.OperationTypeUpdate,
			wantStatus:    constant.ResourceStatusUpdateDraft,
			wantErr:       false,
		},
		{
			name:          "successful status transition",
			currentStatus: constant.ResourceStatusCreateDraft,
			opType:        constant.OperationTypePublish,
			wantStatus:    constant.ResourceStatusSuccess,
			wantErr:       false,
		},
		{
			name:          "invalid status transition",
			currentStatus: constant.ResourceStatusSuccess,
			opType:        constant.OperationTypeCreate,
			wantStatus:    "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := NewResourceStatusOp(model.ResourceCommonModel{
				Status: tt.currentStatus,
			})

			gotStatus, err := op.NextStatus(context.Background(), tt.opType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantStatus, gotStatus)
		})
	}
}

func TestCanSyncToEditorPreservesOtherReadOnlyRestrictions(t *testing.T) {
	gateway := &model.Gateway{ReadOnly: true}
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)
	for _, state := range []constant.ResourceStatus{
		constant.ResourceStatusSuccess, constant.ResourceStatusCreateDraft,
		constant.ResourceStatusUpdateDraft, constant.ResourceStatusDeleteDraft,
	} {
		t.Run(string(state), func(t *testing.T) {
			resource := model.ResourceCommonModel{Status: state}
			assert.NoError(t, NewResourceStatusOp(resource).CanSyncToEditor(ctx))
			assert.True(t, gateway.ReadOnly)
			for _, operation := range []constant.OperationType{
				constant.OperationTypeCreate, constant.OperationTypeUpdate,
				constant.OperationTypeDelete, constant.OperationTypeRevert, constant.OperationTypePublish,
			} {
				assert.ErrorContains(t, NewResourceStatusOp(resource).CanDo(ctx, operation), "网关只读模式")
			}
		})
	}
}

func TestCanSyncToEditorRetainsResourceGuards(t *testing.T) {
	previous := config.G
	t.Cleanup(func() { config.G = previous })
	config.G = &config.Config{
		Service: config.ServiceConfig{DemoMode: true},
		Biz:     config.BizConfig{DemoProtectResources: map[string]bool{"protected": true}},
	}
	for _, readOnly := range []bool{false, true} {
		t.Run(fmt.Sprintf("readonly=%t", readOnly), func(t *testing.T) {
			ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ReadOnly: readOnly})
			protected := model.ResourceCommonModel{ID: "protected", Status: constant.ResourceStatusSuccess}
			assert.ErrorContains(t, NewResourceStatusOp(protected).CanSyncToEditor(ctx), "被保护状态禁止修改")
			for _, state := range []constant.ResourceStatus{"", "invalid"} {
				assert.Error(
					t,
					NewResourceStatusOp(
						model.ResourceCommonModel{Status: state},
					).CanSyncToEditor(
						ctx,
					),
				)
			}
		})
	}
}
