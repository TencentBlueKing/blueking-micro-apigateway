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

package serializer

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

// OpenResolvedDraft carries middleware-computed identity plus the normalized storage payload.
type OpenResolvedDraft struct {
	ID            string
	StorageConfig json.RawMessage
}

const openResolvedDraftsContextKey = "openapi_resolved_drafts"

// SetOpenResolvedDrafts stores resolved Open drafts in gin.Context for later serializer reuse.
func SetOpenResolvedDrafts(c *gin.Context, drafts []OpenResolvedDraft) {
	c.Set(openResolvedDraftsContextKey, drafts)
}

// GetOpenResolvedDrafts reads resolved Open drafts from gin.Context.
func GetOpenResolvedDrafts(c *gin.Context) ([]OpenResolvedDraft, bool) {
	value, ok := c.Get(openResolvedDraftsContextKey)
	if !ok {
		return nil, false
	}

	drafts, ok := value.([]OpenResolvedDraft)
	if !ok {
		return nil, false
	}
	return drafts, true
}
