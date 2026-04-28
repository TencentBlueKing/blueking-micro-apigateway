package handler

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
	openmiddleware "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestResourceBatchCreateKeepsValidationAndPersistedIdentityAligned(t *testing.T) {
	var capturedValidationPayload string
	var capturedResources []*model.ResourceCommonModel

	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(schema.NewAPISIXSchemaValidator, func(
		constant.APISIXVersion,
		string,
	) (schema.Validator, error) {
		return resourceCreateValidatorStub{validate: func(payload json.RawMessage) error {
			capturedValidationPayload = string(payload)
			return nil
		}}, nil
	})
	patches.ApplyFunc(biz.GetCustomizePluginSchemaMap, func(context.Context) (map[string]any, error) {
		return map[string]any{}, nil
	})
	patches.ApplyFunc(schema.NewAPISIXJsonSchemaValidator, func(
		constant.APISIXVersion,
		constant.APISIXResource,
		string,
		map[string]any,
		constant.DataType,
	) (schema.Validator, error) {
		return resourceCreateValidatorStub{}, nil
	})
	patches.ApplyFunc(validation.ValidateStruct, func(context.Context, any) error {
		return nil
	})
	patches.ApplyFunc(
		biz.BatchCheckNameDuplication,
		func(context.Context, constant.APISIXResource, []string) (bool, error) {
			return false, nil
		},
	)
	patches.ApplyFunc(
		biz.BatchCreateResources,
		func(_ context.Context, _ constant.APISIXResource, resources []*model.ResourceCommonModel) error {
			capturedResources = resources
			return nil
		},
	)

	gateway := data.Gateway1WithBkAPISIX()
	gateway.ID = 1001
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, gateway)
		ginx.SetUserID(c, "tester")
		c.Next()
	})
	router.Use(openmiddleware.OpenAPIResourceCheck())
	router.POST("/api/v1/open/gateways/:gateway_name/resources/:resource_type/", ResourceBatchCreate)

	body := `[{"name":"route-a","config":{"uris":["/test"],"methods":["GET"],"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}}]`
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/open/gateways/gateway1/resources/routes/",
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	if assert.Len(t, capturedResources, 1) {
		assert.Equal(t, gjson.Get(capturedValidationPayload, "id").String(), capturedResources[0].ID)
	}
}

type resourceCreateValidatorStub struct {
	validate func(json.RawMessage) error
}

func (v resourceCreateValidatorStub) Validate(payload json.RawMessage) error {
	if v.validate != nil {
		return v.validate(payload)
	}
	return nil
}
