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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestRouteCreateRejectsConflictingClientConfigID(t *testing.T) {
	validation.RegisterValidator()

	createCalled := false
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	patches.ApplyFunc(idx.GenResourceID, func(constant.APISIXResource) string {
		return "server-route-id"
	})
	patches.ApplyFunc(
		biz.DuplicatedResourceName,
		func(context.Context, constant.APISIXResource, string, string) bool {
			return false
		},
	)
	patches.ApplyFunc(schema.NewAPISIXSchemaValidator, func(
		constant.APISIXVersion,
		string,
	) (schema.Validator, error) {
		return routeCreateValidatorStub{}, nil
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
		return routeCreateValidatorStub{}, nil
	})
	patches.ApplyFunc(biz.CreateRoute, func(context.Context, model.Route) error {
		createCalled = true
		return nil
	})

	gateway := data.Gateway1WithBkAPISIX()
	gateway.ID = 1001
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ginx.SetGatewayInfo(c, gateway)
		ginx.SetUserID(c, "tester")
		ginx.SetValidateErrorInfo(c)
		c.Next()
	})
	router.POST("/test", RouteCreate)

	body := `{"name":"route-a","config":{"id":"client-route-id","uris":["/test"],"methods":["GET"],"upstream":{"type":"roundrobin","nodes":[{"host":"127.0.0.1","port":80,"weight":1}]}}}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	assert.Contains(t, w.Body.String(), "config.id conflicts with resolved resource id")
	assert.False(t, createCalled)
}

type routeCreateValidatorStub struct {
	validate func(json.RawMessage) error
}

func (v routeCreateValidatorStub) Validate(payload json.RawMessage) error {
	if v.validate != nil {
		return v.validate(payload)
	}
	return nil
}
