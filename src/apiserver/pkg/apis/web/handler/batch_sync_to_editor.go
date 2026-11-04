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

package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	unifyopbiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/unifyop"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// ReadonlyEditorSyncPreview previews gateway-wide differences for bulk overwrite.
//
//	@ID			readonly_editor_sync_preview
//	@Summary	预览只读网关批量同步到编辑区
//	@Tags		webapi.unify_op
//	@Produce	json
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.ReadonlyEditorSyncPreview
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/-/sync/preview/ [get]
func ReadonlyEditorSyncPreview(c *gin.Context) {
	preview, err := unifyopbiz.PreviewReadonlyEditorSync(c.Request.Context())
	if err != nil {
		readonlyEditorSyncErrorResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, preview)
}

// ReadonlyEditorSync adds missing resources and overwrites confirmed differences.
//
//	@ID			readonly_editor_sync
//	@Summary	只读网关批量同步到编辑区
//	@Tags		webapi.unify_op
//	@Accept		json
//	@Produce	json
//	@Param		gateway_id	path		int									true	"网关 ID"
//	@Param		request		body		serializer.BatchSyncToEditorRequest	true	"确认覆盖"
//	@Success	200			{object}	serializer.ReadonlyEditorSyncResult
//	@Router		/api/v1/web/gateways/{gateway_id}/unify_op/resources/-/sync/ [post]
func ReadonlyEditorSync(c *gin.Context) {
	var req serializer.BatchSyncToEditorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	result, err := unifyopbiz.BatchSyncReadonlyEditor(c.Request.Context(), req.Resources, req.OverwriteDrafts)
	if err != nil {
		readonlyEditorSyncErrorResponse(c, err)
		return
	}
	ginx.SuccessJSONResponse(c, result)
}

func readonlyEditorSyncErrorResponse(c *gin.Context, err error) {
	switch {
	case errors.Is(err, unifyopbiz.ErrBatchSyncRequiresReadOnly),
		errors.Is(err, unifyopbiz.ErrInvalidEditorSyncItem):
		ginx.BadRequestErrorJSONResponse(c, err)
	case errors.Is(err, unifyopbiz.ErrResourceNameConflict):
		ginx.ConflictJSONResponse(c, err)
	default:
		ginx.SystemErrorJSONResponse(c, err)
	}
}
