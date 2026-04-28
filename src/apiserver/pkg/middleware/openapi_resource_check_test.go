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

package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

type middlewareCaptureValidator struct {
	validate func(json.RawMessage) error
}

func (v middlewareCaptureValidator) Validate(raw json.RawMessage) error {
	if v.validate != nil {
		return v.validate(raw)
	}
	return nil
}

func newOpenResourceCheckRouter(version string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	validation.RegisterValidator()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, &model.Gateway{
			ID:            7,
			APISIXVersion: version,
		})
		ginx.SetValidateErrorInfo(c)
		c.Next()
	})

	group := router.Group("/api/v1/open/gateways/:gateway_name/resources")
	group.Use(middleware.OpenAPIResourceCheck())
	group.POST("/:resource_type/", func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	return router
}

func patchOpenResourceCheckValidation(t *testing.T, onValidate func(json.RawMessage) error) *gomonkey.Patches {
	t.Helper()

	patches := gomonkey.NewPatches()
	patches.ApplyFunc(
		schemax.NewAPISIXSchemaValidator,
		func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
			return middlewareCaptureValidator{validate: onValidate}, nil
		},
	)
	patches.ApplyFunc(
		schemax.NewAPISIXJsonSchemaValidator,
		func(
			version constant.APISIXVersion,
			resourceType constant.APISIXResource,
			jsonPath string,
			customizePluginSchemaMap map[string]any,
			dataType constant.DataType,
		) (schemax.Validator, error) {
			return middlewareCaptureValidator{}, nil
		},
	)
	patches.ApplyFunc(
		biz.GetCustomizePluginSchemaMap,
		func(ctx context.Context) (map[string]any, error) {
			return map[string]any{}, nil
		},
	)
	patches.ApplyFunc(
		validation.ValidateStruct,
		func(ctx context.Context, obj any) error {
			return nil
		},
	)

	return patches
}

func TestOpenAPIResourceCheckBuildsCurrentValidationPayloads(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		path          string
		body          string
		assertPayload func(t *testing.T, payload string)
	}{
		{
			name:    "consumer group on 3.13 injects temporary id for validation",
			version: "3.13.0",
			path:    "/api/v1/open/gateways/demo/resources/consumer_groups/",
			body:    `[{"name":"cg-demo","config":{"plugins":{}}}]`,
			assertPayload: func(t *testing.T, payload string) {
				t.Helper()
				assert.NotEmpty(t, gjson.Get(payload, "id").String())
				assert.Equal(t, map[string]any{}, gjson.Get(payload, "plugins").Value())
			},
		},
		{
			name:    "proto on 3.11 strips unsupported name before validation",
			version: "3.11.0",
			path:    "/api/v1/open/gateways/demo/resources/protos/",
			body:    `[{"name":"proto-demo","config":{"name":"proto-demo","content":"syntax = \"proto3\";"}}]`,
			assertPayload: func(t *testing.T, payload string) {
				t.Helper()
				assert.False(t, gjson.Get(payload, "name").Exists())
				assert.Equal(t, `syntax = "proto3";`, gjson.Get(payload, "content").String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validationPayloads []string

			patches := patchOpenResourceCheckValidation(t, func(raw json.RawMessage) error {
				validationPayloads = append(validationPayloads, string(raw))
				return nil
			})
			defer patches.Reset()

			router := newOpenResourceCheckRouter(tt.version)
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if !assert.Equal(t, http.StatusNoContent, recorder.Code) {
				return
			}
			if !assert.Len(t, validationPayloads, 1) {
				return
			}
			tt.assertPayload(t, validationPayloads[0])
		})
	}
}

func TestOpenAPIResourceCheckDoesNotInjectIDForOldConsumerGroupSchema(t *testing.T) {
	var validationPayloads []string

	patches := patchOpenResourceCheckValidation(t, func(raw json.RawMessage) error {
		validationPayloads = append(validationPayloads, string(raw))
		return nil
	})
	defer patches.Reset()

	router := newOpenResourceCheckRouter("3.2.15")
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/open/gateways/demo/resources/consumer_groups/",
		strings.NewReader(`[{"name":"cg-demo","config":{"plugins":{}}}]`),
	)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if !assert.Equal(t, http.StatusNoContent, recorder.Code) {
		return
	}
	if !assert.Len(t, validationPayloads, 1) {
		return
	}
	assert.False(t, gjson.Get(validationPayloads[0], "id").Exists())
}
