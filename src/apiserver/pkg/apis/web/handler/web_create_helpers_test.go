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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

func TestBindAndValidateWebCreateWithGeneratedID(t *testing.T) {
	initWebCreateHandlerTestEnv()

	gateway := &model.Gateway{ID: 2201, APISIXVersion: "3.13.0"}

	t.Run("plugin config gets id before validation", func(t *testing.T) {
		body := mustJSONBody(t, map[string]any{
			"name": "pc-demo",
			"config": map[string]any{
				"plugins": map[string]any{
					"authz-casbin": map[string]any{
						"model":    "path/to/model.conf",
						"policy":   "path/to/policy.csv",
						"username": "admin",
					},
				},
			},
		})

		c, _ := newWebCreateTestContext(t, body, gateway, "helper-tester")

		var req serializer.PluginConfigInfo
		err := bindAndValidateWebCreateWithGeneratedID(c, &req, constant.PluginConfig, func(resourceID string) {
			req.ID = resourceID
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, req.ID)
	})

	t.Run("consumer group gets id before validation", func(t *testing.T) {
		body := mustJSONBody(t, map[string]any{
			"name": "cg-demo",
			"config": map[string]any{
				"plugins": map[string]any{
					"limit-count": map[string]any{
						"count":       100,
						"time_window": 60,
						"key":         "remote_addr",
						"policy":      "local",
					},
				},
			},
		})

		c, _ := newWebCreateTestContext(t, body, gateway, "helper-tester")

		var req serializer.ConsumerGroupInfo
		err := bindAndValidateWebCreateWithGeneratedID(c, &req, constant.ConsumerGroup, func(resourceID string) {
			req.ID = resourceID
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, req.ID)
	})

	t.Run("global rule gets id before validation", func(t *testing.T) {
		body := mustJSONBody(t, map[string]any{
			"name": "gr-demo",
			"config": map[string]any{
				"plugins": map[string]any{
					"authz-casbin": map[string]any{
						"model":    "path/to/model.conf",
						"policy":   "path/to/policy.csv",
						"username": "admin",
					},
				},
			},
		})

		c, _ := newWebCreateTestContext(t, body, gateway, "helper-tester")
		c.Request.Method = http.MethodPost

		var req serializer.GlobalRuleInfo
		err := bindAndValidateWebCreateWithGeneratedID(c, &req, constant.GlobalRule, func(resourceID string) {
			req.ID = resourceID
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, req.ID)
	})
}

func TestBuildWebCreateDraft(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	c, _ := newWebCreateTestContext(t, "{}", &model.Gateway{ID: 12}, "tester")

	draft := buildWebCreateDraft(
		c,
		"resource-id",
		json.RawMessage(`{"plugins":{"limit-count":{"count":1}}}`),
	)

	assert.Equal(t, "resource-id", draft.ID)
	assert.Equal(t, 12, draft.GatewayID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, draft.Status)
	assert.Equal(t, "tester", draft.Creator)
	assert.Equal(t, "tester", draft.Updater)
	assert.JSONEq(t, `{"plugins":{"limit-count":{"count":1}}}`, string(draft.Config))
}
