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

package unifyop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/tidwall/gjson"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	syncdatabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/syncdata"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/sslx"
)

var (
	// ErrSyncReadOnly prevents writes to the edit area of a read-only gateway.
	ErrSyncReadOnly = errors.New("网关只读模式，不允许同步到编辑区")
	// ErrSyncConflict requires a new preview before overwriting changed data.
	ErrSyncConflict = errors.New("资源已变化，请重新对比后同步")
	// ErrSyncDraft requires explicit consent to discard unpublished edits.
	ErrSyncDraft = errors.New("同步将覆盖未发布草稿，请确认后重试")
)

// SyncResourceToEditor overwrites one existing resource from the displayed snapshot.
// It never writes etcd. Typed model hooks retain materialization, schema links and auditing.
func SyncResourceToEditor(ctx context.Context, resourceType constant.APISIXResource,
	id, token string, overwriteDraft bool,
) error {
	gateway := ginx.GetGatewayInfoFromContext(ctx)
	if gateway.ReadOnly {
		return ErrSyncReadOnly
	}
	if !slices.Contains(constant.ResourceTypeList, resourceType) {
		return fmt.Errorf("invalid resource type: %s", resourceType)
	}
	return database.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var snapshot model.GatewaySyncData
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("gateway_id = ? AND type = ? AND id = ?",
			gateway.ID, resourceType, id).First(&snapshot).Error; err != nil {
			return err
		}
		var editor model.ResourceCommonModel
		table := resourcebiz.ResourceTableName(resourceType)
		if err := tx.Table(table).Clauses(clause.Locking{Strength: "UPDATE"}).Where("gateway_id = ? AND id = ?",
			gateway.ID, id).First(&editor).Error; err != nil {
			return err
		}
		if token == "" || token != syncdatabiz.SyncToken(&snapshot, &editor) {
			return ErrSyncConflict
		}
		operation := constant.OperationTypeUpdate
		if editor.Status == constant.ResourceStatusDeleteDraft {
			operation = constant.OperationTypeRevert
		}
		if err := status.NewResourceStatusOp(editor).CanDo(ctx, operation); err != nil {
			return err
		}
		if editor.Status != constant.ResourceStatusSuccess && !overwriteDraft {
			return ErrSyncDraft
		}
		syncStatus, err := syncdatabiz.ConfigStatus(&snapshot, &editor, gateway.GetAPISIXVersionX())
		if err != nil {
			return err
		}
		if syncStatus == constant.SyncedResourceStatusSuccess {
			return nil
		}
		editor.Config, err = editorSyncConfig(&snapshot, &editor, gateway.GetAPISIXVersionX())
		if err != nil {
			return err
		}
		editor.Status = constant.ResourceStatusSuccess
		editor.Updater = ginx.GetUserIDFromContext(ctx)
		typed, err := resourcebiz.NewTypedResourceModel(resourceType, editor)
		if err != nil {
			return err
		}
		// Select zero association values too, so removed references cannot survive the overwrite.
		return tx.Model(typed).Where("gateway_id = ? AND id = ?", gateway.ID, id).
			Select("*").Omit("auto_id", "id", "gateway_id", "creator", "created_at").Updates(typed).Error
	})
}

// editorSyncConfig preserves fields that are owned only by the control plane and
// recomputes certificate metadata from the incoming certificate, never the old one.
func editorSyncConfig(snapshot *model.GatewaySyncData, editor *model.ResourceCommonModel,
	version constant.APISIXVersion,
) (datatypes.JSON, error) {
	var incoming, local map[string]json.RawMessage
	if err := json.Unmarshal(snapshot.Config, &incoming); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(editor.Config, &local); err != nil {
		return nil, err
	}
	var localFields []string
	if snapshot.Type != constant.Consumer && !constant.ResourceSupportsNameFieldForVersion(snapshot.Type, version) {
		localFields = append(localFields, "name")
	}
	if snapshot.Type == constant.StreamRoute {
		localFields = append(localFields, "labels")
	}
	for _, field := range localFields {
		if value, exists := local[field]; exists {
			incoming[field] = value
		} else {
			delete(incoming, field)
		}
	}
	if snapshot.Type == constant.SSL {
		delete(incoming, "validity_start")
		delete(incoming, "validity_end")
		cert := gjson.GetBytes(snapshot.Config, "cert").String()
		// References are resolved by APISIX. Their expiry is unknown to the control plane.
		referenced := strings.HasPrefix(cert, "$secret://") ||
			strings.HasPrefix(cert, "$env://") || strings.HasPrefix(cert, "$ENV://")
		if cert != "" && !referenced {
			validity, err := sslx.X509CertValidity(cert)
			if err != nil {
				return nil, err
			}
			start, err := json.Marshal(validity.NotBefore)
			if err != nil {
				return nil, err
			}
			end, err := json.Marshal(validity.NotAfter)
			if err != nil {
				return nil, err
			}
			incoming["validity_start"] = start
			incoming["validity_end"] = end
		}
	}
	return json.Marshal(incoming)
}
