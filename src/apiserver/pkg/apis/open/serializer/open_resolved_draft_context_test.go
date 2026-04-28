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
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOpenResolvedDraftContextHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("stores and reads drafts", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)

		drafts := []OpenResolvedDraft{
			{
				ID:            "pc-fixed-id",
				StorageConfig: json.RawMessage(`{"id":"pc-fixed-id","name":"pc-demo","plugins":{}}`),
			},
		}

		SetOpenResolvedDrafts(c, drafts)

		got, ok := GetOpenResolvedDrafts(c)
		assert.True(t, ok)
		assert.Len(t, got, 1)
		assert.Equal(t, "pc-fixed-id", got[0].ID)
	})

	t.Run("context key occupied with wrong type returns ok false", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Set(openResolvedDraftsContextKey, "not-a-draft-slice")

		got, ok := GetOpenResolvedDrafts(c)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}
