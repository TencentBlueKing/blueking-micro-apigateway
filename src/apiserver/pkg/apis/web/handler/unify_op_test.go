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
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
)

func TestResourcesDiffReturnsEmptyDiffForUnsupportedTypeBeforeBindingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "type", Value: "plugin_custom"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"resource_id_list":[33,16]}`))
	c.Request.Header.Set("Content-Type", "application/json")

	ResourcesDiff(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":[]`)
	assert.NotContains(t, w.Body.String(), "cannot unmarshal")
}

func TestSyncedResourceManagedReturnsConflictForDuplicateNamesWithinBatch(t *testing.T) {
	initWebCreateHandlerTestEnv()

	suffix := time.Now().UnixNano()
	gateway := &model.Gateway{ID: int(suffix % 1000000000)}
	name := fmt.Sprintf("duplicate-route-name-%d", suffix)
	firstID := fmt.Sprintf("route-first-%d", suffix)
	secondID := fmt.Sprintf("route-second-%d", suffix)
	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(context.Background()).CreateInBatches(
		[]*model.GatewaySyncData{
			{
				ID:        firstID,
				GatewayID: gateway.ID,
				Type:      constant.Route,
				Config:    datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/first"}`, name)),
			},
			{
				ID:        secondID,
				GatewayID: gateway.ID,
				Type:      constant.Route,
				Config:    datatypes.JSON(fmt.Sprintf(`{"name":%q,"uri":"/second"}`, name)),
			},
		},
		100,
	))

	body := fmt.Sprintf(`{"resource_id_list":[%q,%q]}`, firstID, secondID)
	c, w := newWebCreateTestContext(t, body, gateway, "sync-tester")
	SyncedResourceManaged(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"Conflict"`)
	assert.Contains(t, w.Body.String(), name)
	assert.NotContains(t, w.Body.String(), "InternalServerError")
}
