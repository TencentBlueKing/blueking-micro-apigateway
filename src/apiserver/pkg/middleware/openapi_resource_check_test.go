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

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"testing"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestOpenAPIResourceCheckInjectsGeneratedIDForSchemaValidation(t *testing.T) {
	var capturedSchemaPayload string
	var capturedJSONPayload string

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(schema.NewAPISIXSchemaValidator, func(
		_ constant.APISIXVersion,
		_ string,
	) (schema.Validator, error) {
		return validatorStub{validate: func(payload json.RawMessage) error {
			capturedSchemaPayload = string(payload)
			return nil
		}}, nil
	})
	patches.ApplyFunc(biz.GetCustomizePluginSchemaMap, func(context.Context) (map[string]any, error) {
		return map[string]any{}, nil
	})
	patches.ApplyFunc(schema.NewAPISIXJsonSchemaValidator, func(
		_ constant.APISIXVersion,
		_ constant.APISIXResource,
		_ string,
		_ map[string]any,
		_ constant.DataType,
	) (schema.Validator, error) {
		return validatorStub{validate: func(payload json.RawMessage) error {
			capturedJSONPayload = string(payload)
			return nil
		}}, nil
	})
	patches.ApplyFunc(validation.ValidateStruct, func(context.Context, any) error {
		return nil
	})

	gateway := data.Gateway1WithBkAPISIX()
	gateway.ID = 1001
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, gateway)
		c.Next()
	})
	r.Use(OpenAPIResourceCheck())
	r.POST("/api/v1/open/gateways/:gateway_name/resources/:resource_type/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	body := `[{"name":"group-a","config":{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}}]`
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/open/gateways/gateway1/resources/consumer_groups/",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
	assert.Contains(t, capturedSchemaPayload, `"id":"`)
	assert.Contains(t, capturedJSONPayload, `"id":"`)
	assert.NotContains(t, body, `"id":"`)
}

func TestOpenAPIResourceCheckSkipsSchemaValidationForPublishHandler(t *testing.T) {
	called := false
	handlerName := testHandlerShortName(openAPIPublishHandlerStub)
	previousValue, hadExisting := noneValidateSchemaHandlerMap[handlerName]
	noneValidateSchemaHandlerMap[handlerName] = true
	defer func() {
		if hadExisting {
			noneValidateSchemaHandlerMap[handlerName] = previousValue
			return
		}
		delete(noneValidateSchemaHandlerMap, handlerName)
	}()

	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(
		schema.NewAPISIXSchemaValidator,
		func(_ constant.APISIXVersion, _ string) (schema.Validator, error) {
			called = true
			return validatorStub{}, nil
		},
	)

	gateway := data.Gateway1WithBkAPISIX()
	gateway.ID = 1001
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, gateway)
		c.Next()
	})
	r.Use(OpenAPIResourceCheck())
	r.POST("/api/v1/open/gateways/:gateway_name/resources/:resource_type/publish/", openAPIPublishHandlerStub)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/open/gateways/gateway1/resources/routes/publish/",
		strings.NewReader(`{"ids":["id-1"]}`),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.False(t, called)
}

func TestOpenAPIResourceCheckRejectsConflictingSchemaPayload(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(validation.ValidateStruct, func(context.Context, any) error { return nil })
	patches.ApplyFunc(schema.NewAPISIXSchemaValidator, func(
		_ constant.APISIXVersion,
		_ string,
	) (schema.Validator, error) {
		return validatorStub{validate: func(payload json.RawMessage) error {
			if strings.Contains(string(payload), `"missing-service"`) {
				return errors.New("schema conflict")
			}
			return nil
		}}, nil
	})

	gateway := data.Gateway1WithBkAPISIX()
	gateway.ID = 1001
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, gateway)
		c.Next()
	})
	r.Use(OpenAPIResourceCheck())
	r.POST("/api/v1/open/gateways/:gateway_name/resources/:resource_type/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	body := `[{"name":"route-a","config":{"service_id":"missing-service"}}]`
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/open/gateways/gateway1/resources/routes/",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "config validate failed")
}

func TestOpenAPIResourceCheckRejectsConflictingNameBeforeSchemaValidation(t *testing.T) {
	schemaValidatorCalled := false
	jsonValidatorCalled := false
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(validation.ValidateStruct, func(context.Context, any) error { return nil })
	patches.ApplyFunc(schema.NewAPISIXSchemaValidator, func(
		_ constant.APISIXVersion,
		_ string,
	) (schema.Validator, error) {
		schemaValidatorCalled = true
		return validatorStub{}, nil
	})
	patches.ApplyFunc(biz.GetCustomizePluginSchemaMap, func(context.Context) (map[string]any, error) {
		return map[string]any{}, nil
	})
	patches.ApplyFunc(schema.NewAPISIXJsonSchemaValidator, func(
		_ constant.APISIXVersion,
		_ constant.APISIXResource,
		_ string,
		_ map[string]any,
		_ constant.DataType,
	) (schema.Validator, error) {
		jsonValidatorCalled = true
		return validatorStub{}, nil
	})

	gateway := data.Gateway1WithBkAPISIX()
	gateway.ID = 1001
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, gateway)
		ginx.SetValidateErrorInfo(c)
		c.Next()
	})
	r.Use(OpenAPIResourceCheck())
	r.POST("/api/v1/open/gateways/:gateway_name/resources/:resource_type/", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	body := `[{"name":"route-a","config":{"name":"route-b","uris":["/test"]}}]`
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/open/gateways/gateway1/resources/routes/",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "conflicts with request name")
	assert.False(t, jsonValidatorCalled)
	assert.True(t, schemaValidatorCalled)
}

func TestOpenAPIRequestDraftParity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        resourcecodec.RequestInput
		wantDatabase string
	}{
		{
			name: "route create generates id and injects outer name",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				OuterName:    "route-a",
				Config: json.RawMessage(
					`{"uris":["/test"],"methods":["GET"],"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
				),
			},
			wantDatabase: `{"id":"@nonempty","name":"route-a","uris":["/test"],"methods":["GET"],"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}`,
		},
		{
			name: "consumer create uses username and strips generated id",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.Consumer,
				Version:      constant.APISIXVersion313,
				OuterName:    "consumer-a",
				Config:       json.RawMessage(`{"plugins":{"key-auth":{"key":"token-a"}}}`),
			},
			wantDatabase: `{"username":"consumer-a","plugins":{"key-auth":{"key":"token-a"}}}`,
		},
		{
			name: "plugin metadata derives id from outer name",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    constant.OperationTypeCreate,
				GatewayID:    1001,
				ResourceType: constant.PluginMetadata,
				Version:      constant.APISIXVersion313,
				OuterName:    "jwt-auth",
				Config:       json.RawMessage(`{"foo":"bar"}`),
			},
			wantDatabase: `{"id":"jwt-auth","foo":"bar"}`,
		},
		{
			name: "route update uses path id authority and keeps matching name",
			input: resourcecodec.RequestInput{
				Source:       resourcecodec.SourceOpenAPI,
				Operation:    constant.OperationTypeUpdate,
				GatewayID:    1001,
				ResourceType: constant.Route,
				Version:      constant.APISIXVersion311,
				PathID:       "route-id",
				OuterName:    "route-a",
				Config:       json.RawMessage(`{"name":"route-a","uris":["/test"]}`),
			},
			wantDatabase: `{"id":"route-id","name":"route-a","uris":["/test"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft, err := resourcecodec.PrepareRequestDraft(tt.input)
			assert.NoError(t, err)

			builtPayload, err := resourcecodec.BuildRequestPayload(draft, constant.DATABASE)
			assert.NoError(t, err)
			assertOpenAPIJSON(t, tt.wantDatabase, string(builtPayload.Payload))
		})
	}
}

func assertOpenAPIJSON(t *testing.T, want, got string) {
	t.Helper()

	var gotObj map[string]any
	assert.NoError(t, json.Unmarshal([]byte(got), &gotObj))

	var wantObj map[string]any
	assert.NoError(t, json.Unmarshal([]byte(want), &wantObj))

	if wantID, ok := wantObj["id"].(string); ok && wantID == "@nonempty" {
		gotID, ok := gotObj["id"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, gotID)
		delete(wantObj, "id")
		delete(gotObj, "id")
	}

	gotNormalized, err := json.Marshal(gotObj)
	assert.NoError(t, err)
	wantNormalized, err := json.Marshal(wantObj)
	assert.NoError(t, err)
	assert.JSONEq(t, string(wantNormalized), string(gotNormalized))
}

type validatorStub struct {
	validate func(json.RawMessage) error
}

func (v validatorStub) Validate(payload json.RawMessage) error {
	if v.validate != nil {
		return v.validate(payload)
	}
	return nil
}

func openAPIPublishHandlerStub(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func testHandlerShortName(handler any) string {
	fullHandlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	lastSlashIndex := strings.LastIndex(fullHandlerName, "/")
	return fullHandlerName[lastSlashIndex+1:]
}
