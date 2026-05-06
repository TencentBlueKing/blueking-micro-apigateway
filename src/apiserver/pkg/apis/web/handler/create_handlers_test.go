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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

var webCreateHandlerTestOnce sync.Once

func initWebCreateHandlerTestEnv() {
	webCreateHandlerTestOnce.Do(func() {
		gin.SetMode(gin.TestMode)
		util.InitEmbedDb()
		validation.RegisterValidator()
	})
}

func newWebCreateTestContext(
	t *testing.T,
	body string,
	gateway *model.Gateway,
	userID string,
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	ginx.SetGatewayInfo(c, gateway)
	ginx.SetUserID(c, userID)
	ginx.SetValidateErrorInfo(c)
	return c, w
}

func mustJSONBody(t *testing.T, payload any) string {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	return string(body)
}

func uniqueWebCreateName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func assertCreateDraftCommonFields(
	t *testing.T,
	resourceID string,
	gatewayID int,
	creator string,
	status constant.ResourceStatus,
) {
	t.Helper()

	assert.NotEmpty(t, resourceID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, status)
	assert.Equal(t, creator, creator)
	assert.Equal(t, gatewayID, gatewayID)
}

func TestPluginConfigCreateCurrentBehavior(t *testing.T) {
	initWebCreateHandlerTestEnv()

	gateway := &model.Gateway{ID: 2101, APISIXVersion: "3.13.0"}
	userID := "plugin-config-tester"
	name := uniqueWebCreateName("plugin-config")
	body := mustJSONBody(t, map[string]any{
		"name": name,
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

	c, w := newWebCreateTestContext(t, body, gateway, userID)
	PluginConfigCreate(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	items, err := resourcebiz.QueryPluginConfigs(c.Request.Context(), map[string]any{"name": name})
	assert.NoError(t, err)
	if !assert.Len(t, items, 1) {
		return
	}

	created := items[0]
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, gateway.ID, created.GatewayID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, created.Status)
	assert.Equal(t, userID, created.Creator)
	assert.Equal(t, userID, created.Updater)
	assert.Equal(t, name, created.Name)
	assert.Equal(t, created.ID, gjson.GetBytes(created.Config, "id").String())
}

func TestConsumerGroupCreateCurrentBehavior(t *testing.T) {
	initWebCreateHandlerTestEnv()

	gateway := &model.Gateway{ID: 2102, APISIXVersion: "3.13.0"}
	userID := "consumer-group-tester"
	name := uniqueWebCreateName("consumer-group")
	body := mustJSONBody(t, map[string]any{
		"name": name,
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

	c, w := newWebCreateTestContext(t, body, gateway, userID)
	ConsumerGroupCreate(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	items, err := resourcebiz.QueryConsumerGroups(c.Request.Context(), map[string]any{"name": name})
	assert.NoError(t, err)
	if !assert.Len(t, items, 1) {
		return
	}

	created := items[0]
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, gateway.ID, created.GatewayID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, created.Status)
	assert.Equal(t, userID, created.Creator)
	assert.Equal(t, userID, created.Updater)
	assert.Equal(t, name, created.Name)
	assert.Equal(t, created.ID, gjson.GetBytes(created.Config, "id").String())
	assert.Equal(t, name, gjson.GetBytes(created.Config, "name").String())
}

func TestGlobalRuleCreateCurrentBehavior(t *testing.T) {
	initWebCreateHandlerTestEnv()

	gateway := &model.Gateway{ID: 2103, APISIXVersion: "3.13.0"}
	userID := "global-rule-tester"
	name := uniqueWebCreateName("global-rule")
	body := mustJSONBody(t, map[string]any{
		"name": name,
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

	c, w := newWebCreateTestContext(t, body, gateway, userID)
	GlobalRuleCreate(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	items, err := resourcebiz.QueryGlobalRules(c.Request.Context(), map[string]any{"name": name})
	assert.NoError(t, err)
	if !assert.Len(t, items, 1) {
		return
	}

	created := items[0]
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, gateway.ID, created.GatewayID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, created.Status)
	assert.Equal(t, userID, created.Creator)
	assert.Equal(t, userID, created.Updater)
	assert.Equal(t, name, created.Name)
	assert.Equal(t, created.ID, gjson.GetBytes(created.Config, "id").String())
}

func TestRouteCreateCurrentBehavior(t *testing.T) {
	initWebCreateHandlerTestEnv()

	gateway := &model.Gateway{ID: 2104, APISIXVersion: "3.11.0"}
	userID := "route-tester"
	name := uniqueWebCreateName("route")
	body := mustJSONBody(t, map[string]any{
		"name": name,
		"config": map[string]any{
			"uris":    []string{"/get"},
			"methods": []string{"GET"},
			"upstream": map[string]any{
				"type": "roundrobin",
				"nodes": []map[string]any{
					{
						"host":   "httpbin.org",
						"port":   80,
						"weight": 1,
					},
				},
				"scheme": "http",
			},
		},
	})

	c, w := newWebCreateTestContext(t, body, gateway, userID)
	RouteCreate(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	items, err := resourcebiz.QueryRoutes(c.Request.Context(), map[string]any{"name": name})
	assert.NoError(t, err)
	if !assert.Len(t, items, 1) {
		return
	}

	created := items[0]
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, gateway.ID, created.GatewayID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, created.Status)
	assert.Equal(t, userID, created.Creator)
	assert.Equal(t, userID, created.Updater)
	assert.Equal(t, name, created.Name)
	assert.Equal(t, created.ID, gjson.GetBytes(created.Config, "id").String())
	assert.Equal(t, name, gjson.GetBytes(created.Config, "name").String())
}

func TestSSLCreateCurrentBehavior(t *testing.T) {
	initWebCreateHandlerTestEnv()

	gateway := &model.Gateway{ID: 2105, APISIXVersion: "3.13.0"}
	userID := "ssl-tester"
	name := uniqueWebCreateName("ssl")
	sslFixture := data.SSL1(gateway, constant.ResourceStatusCreateDraft)
	body := mustJSONBody(t, map[string]any{
		"name":   name,
		"config": json.RawMessage(sslFixture.Config),
	})

	c, w := newWebCreateTestContext(t, body, gateway, userID)
	SSLCreate(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	items, err := resourcebiz.QuerySSL(c.Request.Context(), map[string]any{"name": name})
	assert.NoError(t, err)
	if !assert.Len(t, items, 1) {
		return
	}

	created := items[0]
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, gateway.ID, created.GatewayID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, created.Status)
	assert.Equal(t, userID, created.Creator)
	assert.Equal(t, userID, created.Updater)
	assert.Equal(t, name, created.Name)
	assert.Equal(t, created.ID, gjson.GetBytes(created.Config, "id").String())
	assert.Equal(t, name, gjson.GetBytes(created.Config, "name").String())
	assert.NotEmpty(t, gjson.GetBytes(created.Config, "cert").String())
	assert.NotEmpty(t, gjson.GetBytes(created.Config, "key").String())
	assert.Greater(t, gjson.GetBytes(created.Config, "validity_start").Int(), int64(0))
	assert.Greater(t, gjson.GetBytes(created.Config, "validity_end").Int(), int64(0))
	assert.NotEmpty(t, gjson.GetBytes(created.Config, "snis.0").String())
}
