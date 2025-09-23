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
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
	testingutil "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/testing"
)

func TestQueryTokenAuthRight(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	c := testingutil.CreateTestContextWithDefaultRequest(w)

	q := c.Request.URL.Query()
	q.Add("token", "token_for_test")
	c.Request.URL.RawQuery = q.Encode()

	middleware.QueryTokenAuth("token_for_test")(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQueryTokenAuthBad(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	c := testingutil.CreateTestContextWithDefaultRequest(w)

	middleware.QueryTokenAuth("token_for_test")(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
