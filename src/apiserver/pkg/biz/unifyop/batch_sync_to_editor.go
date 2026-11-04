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
	"errors"
	"fmt"
	"slices"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	syncdatabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/syncdata"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ErrBatchSyncRequiresReadOnly restricts bulk overwrites to read-only gateways.
var ErrBatchSyncRequiresReadOnly = errors.New("仅只读网关允许批量覆盖编辑区差异")

// ErrInvalidEditorSyncItem identifies invalid or duplicate batch confirmation items.
var ErrInvalidEditorSyncItem = errors.New("invalid or duplicate sync item")

// PreviewReadonlyEditorSync returns all differences with their existing confirmation tokens.
func PreviewReadonlyEditorSync(ctx context.Context) (*dto.ReadonlyEditorSyncPreview, error) {
	gateway := ginx.GetGatewayInfoFromContext(ctx)
	if !gateway.ReadOnly {
		return nil, ErrBatchSyncRequiresReadOnly
	}
	snapshots, err := syncdatabiz.QuerySyncedItems(ctx, map[string]any{"gateway_id": gateway.ID})
	if err != nil {
		return nil, err
	}
	editors := make(map[constant.APISIXResource]map[string]*model.ResourceCommonModel)
	preview := &dto.ReadonlyEditorSyncPreview{Resources: []dto.EditorSyncItem{}}
	for _, snapshot := range snapshots {
		if !slices.Contains(constant.ResourceTypeList, snapshot.Type) {
			continue
		}
		if _, ok := editors[snapshot.Type]; !ok {
			resources, err := resourcebiz.BatchGetResources(ctx, snapshot.Type, nil)
			if err != nil {
				return nil, err
			}
			editors[snapshot.Type] = make(map[string]*model.ResourceCommonModel)
			for _, resource := range resources {
				editors[snapshot.Type][resource.ID] = resource
			}
		}
		editor := editors[snapshot.Type][snapshot.ID]
		state, err := syncdatabiz.ConfigStatus(snapshot, editor, gateway.GetAPISIXVersionX())
		if err != nil {
			return nil, err
		}
		switch state {
		case constant.SyncedResourceStatusMiss:
			preview.Missing++
		case constant.SyncedResourceStatusDifferent:
			preview.Resources = append(preview.Resources, dto.EditorSyncItem{
				ResourceType: snapshot.Type, ID: snapshot.ID,
				SyncToken: syncdatabiz.SyncToken(snapshot, editor), EditorStatus: editor.Status,
			})
		}
	}
	return preview, nil
}

// BatchSyncReadonlyEditor adds missing resources and overwrites only confirmed differences.
// Existing add and single-resource sync operations retain their own transaction boundaries.
func BatchSyncReadonlyEditor(ctx context.Context, items []dto.EditorSyncItem,
	overwriteDrafts bool,
) (*dto.ReadonlyEditorSyncResult, error) {
	if !ginx.GetGatewayInfoFromContext(ctx).ReadOnly {
		return nil, ErrBatchSyncRequiresReadOnly
	}
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		key := fmt.Sprintf("%s/%s", item.ResourceType, item.ID)
		if !slices.Contains(constant.ResourceTypeList, item.ResourceType) || item.ID == "" ||
			len(item.SyncToken) != 64 || seen[key] {
			return nil, fmt.Errorf("%w: %s", ErrInvalidEditorSyncItem, key)
		}
		seen[key] = true
	}
	// Add dependencies before applying updates. Existing resources are never overwritten here.
	added, err := AddSyncedResources(ctx, nil)
	if err != nil {
		if errors.Is(err, ErrResourceNameConflict) && len(items) > 0 {
			return syncReadonlyEditorWithNameConflicts(ctx, items, overwriteDrafts)
		}
		return nil, err
	}
	result := &dto.ReadonlyEditorSyncResult{Added: added, Failed: []dto.EditorSyncFailure{}}
	for _, item := range items {
		if err := SyncResourceToEditor(
			ctx,
			item.ResourceType,
			item.ID,
			item.SyncToken,
			overwriteDrafts,
		); err != nil {
			result.Failed = append(result.Failed, dto.EditorSyncFailure{
				ResourceType: item.ResourceType, ID: item.ID, Error: err.Error(),
			})
			continue
		}
		result.Synced++
	}
	return result, nil
}

// missingEditorSyncResources selects by identity, not name: existing rows may only
// be changed by the confirmed single-resource sync operation.
func missingEditorSyncResources(ctx context.Context) ([]*model.GatewaySyncData, error) {
	snapshots, err := syncdatabiz.QuerySyncedItems(ctx, map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	})
	if err != nil {
		return nil, err
	}
	editors := make(map[constant.APISIXResource]map[string]bool)
	var missing []*model.GatewaySyncData
	for _, snapshot := range snapshots {
		if !slices.Contains(constant.ResourceTypeList, snapshot.Type) {
			continue
		}
		if _, ok := editors[snapshot.Type]; !ok {
			resources, err := resourcebiz.BatchGetResources(ctx, snapshot.Type, nil)
			if err != nil {
				return nil, err
			}
			editors[snapshot.Type] = make(map[string]bool, len(resources))
			for _, resource := range resources {
				editors[snapshot.Type][resource.ID] = true
			}
		}
		if !editors[snapshot.Type][snapshot.ID] {
			missing = append(missing, snapshot)
		}
	}
	return missing, nil
}

// A new resource can reuse the old name of an editor resource awaiting overwrite.
// Import independent dependencies first, then retry blocked additions/overwrites
// only while another operation succeeds. Never retry successful confirmation tokens.
func syncReadonlyEditorWithNameConflicts(ctx context.Context, items []dto.EditorSyncItem,
	overwriteDrafts bool,
) (*dto.ReadonlyEditorSyncResult, error) {
	missing, err := missingEditorSyncResources(ctx)
	if err != nil {
		return nil, err
	}
	result := &dto.ReadonlyEditorSyncResult{
		Added: make(map[constant.APISIXResource]int), Failed: []dto.EditorSyncFailure{},
	}
	for {
		var pendingMissing []*model.GatewaySyncData
		var pendingItems []dto.EditorSyncItem
		var blocked []dto.EditorSyncFailure
		progressed := false
		for _, snapshot := range missing {
			added, err := AddSyncedResourcesOfType(ctx, []string{snapshot.ID}, snapshot.Type)
			if err != nil {
				failure := dto.EditorSyncFailure{
					ResourceType: snapshot.Type,
					ID:           snapshot.ID,
					Error:        "添加失败: " + err.Error(),
				}
				if errors.Is(err, ErrResourceNameConflict) {
					pendingMissing = append(pendingMissing, snapshot)
					blocked = append(blocked, failure)
				} else {
					result.Failed = append(result.Failed, failure)
				}
				continue
			}
			for resourceType, count := range added {
				result.Added[resourceType] += count
			}
			progressed = true
		}
		for _, item := range items {
			err := SyncResourceToEditor(ctx, item.ResourceType, item.ID, item.SyncToken, overwriteDrafts)
			if err != nil {
				failure := dto.EditorSyncFailure{
					ResourceType: item.ResourceType,
					ID:           item.ID,
					Error:        err.Error(),
				}
				if errors.Is(err, ErrSyncDependencyMissing) {
					pendingItems = append(pendingItems, item)
					blocked = append(blocked, failure)
				} else {
					result.Failed = append(result.Failed, failure)
				}
				continue
			}
			result.Synced++
			progressed = true
		}
		if !progressed || len(blocked) == 0 {
			result.Failed = append(result.Failed, blocked...)
			return result, nil
		}
		missing, items = pendingMissing, pendingItems
	}
}
