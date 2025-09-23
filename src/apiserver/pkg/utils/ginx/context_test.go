/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
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
package ginx_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

func TestGetRequestID(t *testing.T) {
	c := &gin.Context{}

	requestID := ginx.GetRequestID(c)
	assert.Equal(t, "", requestID)
}

func TestSetRequestID(t *testing.T) {
	c := &gin.Context{}

	ginx.SetRequestID(c, "test")
	assert.Equal(t, "test", ginx.GetRequestID(c))
}

func TestSetSlogGinRequestID(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		testID := "test-request-id-123"

		ginx.SetSlogGinRequestID(c, testID)
		assert.Equal(t, testID, c.Request.Header.Get(constant.RequestIDHeaderKey))
	})

	t.Run("empty request id", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}

		ginx.SetSlogGinRequestID(c, "")
		assert.Equal(t, "", c.Request.Header.Get(constant.RequestIDHeaderKey))
	})
}

func TestGetError(t *testing.T) {
	c := &gin.Context{}

	err, ok := ginx.GetError(c)
	assert.Equal(t, false, ok)
	assert.Equal(t, nil, err)
}

func TestSetError(t *testing.T) {
	c := &gin.Context{}
	err := errors.New("test")

	ginx.SetError(c, err)
	gErr, ok := ginx.GetError(c)

	assert.Equal(t, true, ok)
	assert.Equal(t, err, gErr)
}

func TestGetUserID(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		c := &gin.Context{} // Request is nil
		ginx.SetUserID(c, "test-from-context")
		assert.Equal(t, "test-from-context", ginx.GetUserID(c))
	})

	t.Run("no userID in context", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		assert.Equal(t, "", ginx.GetUserID(c))
	})

	t.Run("invalid userID type in context", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		// Set non-string value
		ctx := context.WithValue(c.Request.Context(), constant.UserIDKey, 123)
		c.Request = c.Request.WithContext(ctx)
		assert.Equal(t, "", ginx.GetUserID(c))
	})

	t.Run("valid userID in context", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		ctx := context.WithValue(c.Request.Context(), constant.UserIDKey, "test-user")
		c.Request = c.Request.WithContext(ctx)
		assert.Equal(t, "test-user", ginx.GetUserID(c))
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	t.Run("invalid userID type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), constant.UserIDKey, 123)
		assert.Equal(t, "", ginx.GetUserIDFromContext(ctx))
	})

	t.Run("valid userID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), constant.UserIDKey, "test-user")
		assert.Equal(t, "test-user", ginx.GetUserIDFromContext(ctx))
	})
}

func TestSetUserID(t *testing.T) {
	t.Run("basic set userID", func(t *testing.T) {
		c := &gin.Context{}
		ginx.SetUserID(c, "test")
		assert.Equal(t, "test", ginx.GetUserID(c))
	})

	t.Run("set userID with request", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		userID := "test-user-123"
		ginx.SetUserID(c, userID)

		// 验证上下文中的值
		ctxUserID := c.Request.Context().Value(constant.UserIDKey)
		assert.Equal(t, userID, ctxUserID)

		// 验证通过GetUserID获取的值
		assert.Equal(t, userID, ginx.GetUserID(c))
	})

	t.Run("nil request", func(t *testing.T) {
		c := &gin.Context{} // Request is nil
		userID := "test-nil-request"
		ginx.SetUserID(c, userID)
		assert.Equal(t, userID, ginx.GetUserID(c))
	})
}

func TestGetGatewayInfo(t *testing.T) {
	t.Run("no gateway info", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		assert.Nil(t, ginx.GetGatewayInfo(c))
	})

	t.Run("with gateway info", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		testGateway := &model.Gateway{ID: 1}
		ginx.SetGatewayInfo(c, testGateway)
		assert.Equal(t, testGateway, ginx.GetGatewayInfo(c))
	})
}

func TestGetGatewayInfoFromContext(t *testing.T) {
	t.Run("no gateway info", func(t *testing.T) {
		ctx := context.Background()
		assert.Nil(t, ginx.GetGatewayInfoFromContext(ctx))
	})

	t.Run("with gateway info", func(t *testing.T) {
		testGateway := &model.Gateway{ID: 1}
		ctx := context.WithValue(context.Background(), constant.GatewayInfoKey, testGateway)
		assert.Equal(t, testGateway, ginx.GetGatewayInfoFromContext(ctx))
	})
}

func TestSetGatewayInfo(t *testing.T) {
	t.Run("set gateway info", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		testGateway := &model.Gateway{ID: 1}
		ginx.SetGatewayInfo(c, testGateway)
		assert.Equal(t, testGateway, c.Request.Context().Value(constant.GatewayInfoKey))
	})
}

func TestSetGatewayInfoToContext(t *testing.T) {
	t.Run("set gateway info to context", func(t *testing.T) {
		testGateway := &model.Gateway{ID: 1}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), testGateway)
		assert.Equal(t, testGateway, ctx.Value(constant.GatewayInfoKey))
	})
}

func TestSetResourceType(t *testing.T) {
	t.Run("set resource type", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}

		ginx.SetResourceType(c, constant.Service)
		assert.Equal(t, constant.Service, c.Request.Context().Value(constant.ResourceTypeKey))
	})
}

func TestGetResourceType(t *testing.T) {
	t.Run("no resource type", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		assert.Equal(t, constant.APISIXResource(""), ginx.GetResourceType(c))
	})

	t.Run("with resource type", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		ginx.SetResourceType(c, constant.Route)
		assert.Equal(t, constant.Route, ginx.GetResourceType(c))
	})
}

func TestGetResourceTypeFromContext(t *testing.T) {
	t.Run("no resource type", func(t *testing.T) {
		ctx := context.Background()
		assert.Equal(t, constant.APISIXResource(""), ginx.GetResourceTypeFromContext(ctx))
	})

	t.Run("with resource type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), constant.ResourceTypeKey, "upstream")
		assert.Equal(t, constant.Upstream, ginx.GetResourceTypeFromContext(ctx))
	})
}

func TestGetValidateErrorInfoFromContext(t *testing.T) {
	t.Run("no validate error", func(t *testing.T) {
		ctx := context.Background()
		assert.Nil(t, ginx.GetValidateErrorInfoFromContext(ctx))
	})

	t.Run("with validate error", func(t *testing.T) {
		err := &schema.APISIXValidateError{Err: errors.New("test error")}
		ctx := context.WithValue(context.Background(), constant.APISIXValidateErrKey, err)
		assert.Equal(t, err, ginx.GetValidateErrorInfoFromContext(ctx))
	})
}

func TestSetValidateErrorInfo(t *testing.T) {
	t.Run("set validate error", func(t *testing.T) {
		c := &gin.Context{
			Request: &http.Request{
				Header: make(http.Header),
			},
		}
		ginx.SetValidateErrorInfo(c)
		assert.NotNil(t, c.Request.Context().Value(constant.APISIXValidateErrKey))
	})
}

func TestCloneCtx(t *testing.T) {
	t.Run("clone context with gateway info", func(t *testing.T) {
		testGateway := &model.Gateway{ID: 1}
		ctx := context.WithValue(context.Background(), constant.GatewayInfoKey, testGateway)
		newCtx := ginx.CloneCtx(ctx)
		assert.Equal(t, testGateway, newCtx.Value(constant.GatewayInfoKey))
	})
}
