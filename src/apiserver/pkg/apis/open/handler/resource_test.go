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

package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	openhandler "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/open/handler"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	openmiddleware "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	schemax "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

type captureValidator struct {
	validate func(json.RawMessage) error
}

func (v captureValidator) Validate(raw json.RawMessage) error {
	if v.validate != nil {
		return v.validate(raw)
	}
	return nil
}

func newOpenBatchCreateRouter(version string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	validation.RegisterValidator()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, &model.Gateway{
			ID:            42,
			APISIXVersion: version,
		})
		ginx.SetUserID(c, "openapi-user")
		ginx.SetValidateErrorInfo(c)
		c.Next()
	})

	group := router.Group("/api/v1/open/gateways/:gateway_name/resources")
	group.Use(openmiddleware.OpenAPIResourceCheck())
	group.POST("/:resource_type/", openhandler.ResourceBatchCreate)
	group.PUT("/:resource_type/:id/", openhandler.ResourceUpdate)

	return router
}

func patchOpenValidation(t *testing.T, onValidate func(json.RawMessage) error) *gomonkey.Patches {
	t.Helper()

	patches := gomonkey.NewPatches()
	patches.ApplyFunc(
		schemax.NewAPISIXSchemaValidator,
		func(version constant.APISIXVersion, jsonPath string) (schemax.Validator, error) {
			return captureValidator{validate: onValidate}, nil
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
			return captureValidator{}, nil
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

func TestResourceBatchCreateInjectsNameAndUsernameAtHandlerSeam(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		path         string
		resourceType constant.APISIXResource
		body         string
		assertConfig func(t *testing.T, resource *model.ResourceCommonModel)
	}{
		{
			name:         "route injects name when config omits it",
			version:      "3.13.0",
			path:         "/api/v1/open/gateways/demo/resources/routes/",
			resourceType: constant.Route,
			body:         `[{"name":"route-demo","config":{"uri":"/demo"}}]`,
			assertConfig: func(t *testing.T, resource *model.ResourceCommonModel) {
				t.Helper()
				assert.Equal(t, "route-demo", gjson.GetBytes(resource.Config, "name").String())
				assert.Equal(t, "/demo", gjson.GetBytes(resource.Config, "uri").String())
			},
		},
		{
			name:         "consumer injects username instead of name",
			version:      "3.13.0",
			path:         "/api/v1/open/gateways/demo/resources/consumers/",
			resourceType: constant.Consumer,
			body:         `[{"name":"consumer-demo","config":{"plugins":{}}}]`,
			assertConfig: func(t *testing.T, resource *model.ResourceCommonModel) {
				t.Helper()
				assert.Equal(t, "consumer-demo", gjson.GetBytes(resource.Config, "username").String())
				assert.False(t, gjson.GetBytes(resource.Config, "name").Exists())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created []*model.ResourceCommonModel

			patches := patchOpenValidation(t, nil)
			defer patches.Reset()

			patches.ApplyFunc(
				biz.BatchCheckNameDuplication,
				func(ctx context.Context, resourceType constant.APISIXResource, names []string) (bool, error) {
					assert.Equal(t, tt.resourceType, resourceType)
					assert.Len(t, names, 1)
					return false, nil
				},
			)
			patches.ApplyFunc(
				biz.BatchCreateResources,
				func(
					ctx context.Context,
					resourceType constant.APISIXResource,
					resources []*model.ResourceCommonModel,
				) error {
					assert.Equal(t, tt.resourceType, resourceType)
					created = resources
					return nil
				},
			)

			var genCount int
			patches.ApplyFunc(idx.GenResourceID, func(resourceType constant.APISIXResource) string {
				genCount++
				return fmt.Sprintf("generated-%d", genCount)
			})

			router := newOpenBatchCreateRouter(tt.version)
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, req)

			if !assert.Equal(t, http.StatusOK, recorder.Code) {
				return
			}
			if !assert.Len(t, created, 1) {
				return
			}
			assert.Equal(t, "generated-1", created[0].ID)
			assert.Equal(t, 42, created[0].GatewayID)
			assert.Equal(t, constant.ResourceStatusCreateDraft, created[0].Status)
			tt.assertConfig(t, created[0])
			assert.Equal(t, created[0].ID, gjson.Get(recorder.Body.String(), "data.0.id").String())
			assert.Equal(t, gjson.GetBytes(created[0].Config, model.GetResourceNameKey(tt.resourceType)).String(),
				gjson.Get(recorder.Body.String(), "data.0.name").String())
		})
	}
}

func TestResourceBatchCreateReusesResolvedIDsAcrossValidationAndPersistence(t *testing.T) {
	var (
		createdResources   []*model.ResourceCommonModel
		validationPayloads []string
	)

	patches := patchOpenValidation(t, func(raw json.RawMessage) error {
		validationPayloads = append(validationPayloads, string(raw))
		return nil
	})
	defer patches.Reset()

	patches.ApplyFunc(
		biz.BatchCheckNameDuplication,
		func(ctx context.Context, resourceType constant.APISIXResource, names []string) (bool, error) {
			assert.Equal(t, constant.ConsumerGroup, resourceType)
			assert.Equal(t, []string{"cg-demo"}, names)
			return false, nil
		},
	)
	patches.ApplyFunc(
		biz.BatchCreateResources,
		func(
			ctx context.Context,
			resourceType constant.APISIXResource,
			resources []*model.ResourceCommonModel,
		) error {
			assert.Equal(t, constant.ConsumerGroup, resourceType)
			createdResources = resources
			return nil
		},
	)

	var genCount int
	patches.ApplyFunc(idx.GenResourceID, func(resourceType constant.APISIXResource) string {
		genCount++
		return fmt.Sprintf("generated-%d", genCount)
	})

	router := newOpenBatchCreateRouter("3.13.0")
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/open/gateways/demo/resources/consumer_groups/",
		strings.NewReader(`[{"name":"cg-demo","config":{"plugins":{}}}]`),
	)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if !assert.Equal(t, http.StatusOK, recorder.Code) {
		return
	}
	if !assert.Len(t, validationPayloads, 1) {
		return
	}
	if !assert.Len(t, createdResources, 1) {
		return
	}

	validationID := gjson.Get(validationPayloads[0], "id").String()
	persistedID := createdResources[0].ID

	assert.Equal(t, "generated-1", validationID)
	assert.Equal(t, "generated-1", persistedID)
	assert.Equal(t, validationID, persistedID)
	assert.Equal(t, persistedID, gjson.Get(recorder.Body.String(), "data.0.id").String())
}

func TestResourceUpdateWritesOuterNameBackIntoConfig(t *testing.T) {
	var updatedResource *model.ResourceCommonModel

	patches := patchOpenValidation(t, nil)
	defer patches.Reset()

	patches.ApplyFunc(
		biz.GetResourceByID,
		func(
			ctx context.Context,
			resourceType constant.APISIXResource,
			id string,
		) (model.ResourceCommonModel, error) {
			assert.Equal(t, constant.Route, resourceType)
			assert.Equal(t, "route-id", id)
			return model.ResourceCommonModel{
				ID:     id,
				Status: constant.ResourceStatusSuccess,
			}, nil
		},
	)
	patches.ApplyFunc(
		biz.DuplicatedResourceName,
		func(ctx context.Context, resourceType constant.APISIXResource, id string, name string) bool {
			assert.Equal(t, constant.Route, resourceType)
			assert.Equal(t, "route-id", id)
			assert.Equal(t, "route-demo", name)
			return false
		},
	)
	patches.ApplyFunc(
		biz.GetResourceUpdateStatus,
		func(ctx context.Context, resourceType constant.APISIXResource, id string) (constant.ResourceStatus, error) {
			assert.Equal(t, constant.Route, resourceType)
			assert.Equal(t, "route-id", id)
			return constant.ResourceStatusUpdateDraft, nil
		},
	)
	patches.ApplyFunc(
		biz.UpdateResource,
		func(
			ctx context.Context,
			resourceType constant.APISIXResource,
			id string,
			resource *model.ResourceCommonModel,
		) error {
			assert.Equal(t, constant.Route, resourceType)
			assert.Equal(t, "route-id", id)
			updatedResource = resource
			return nil
		},
	)

	router := newOpenBatchCreateRouter("3.13.0")
	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/open/gateways/demo/resources/routes/route-id/",
		strings.NewReader(`{"name":"route-demo","config":{"uri":"/demo"}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if !assert.Equal(t, http.StatusNoContent, recorder.Code) {
		return
	}
	if !assert.NotNil(t, updatedResource) {
		return
	}
	assert.Equal(t, "route-id", updatedResource.ID)
	assert.Equal(t, 42, updatedResource.GatewayID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedResource.Status)
	assert.Equal(t, "openapi-user", updatedResource.Updater)
	assert.Equal(t, "/demo", gjson.GetBytes(updatedResource.Config, "uri").String())
	assert.Equal(t, "route-demo", gjson.GetBytes(updatedResource.Config, "name").String())
}
