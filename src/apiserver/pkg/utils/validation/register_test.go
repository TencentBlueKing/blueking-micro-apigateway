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
package validation_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

func TestValidation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validation Suite")
}

var _ = BeforeSuite(func() {
	validation.RegisterValidator()
})

var _ = Describe("Validation", func() {
	Describe("BindAndValidate", func() {
		Context("with valid JSON and struct", func() {
			It("should not return error", func() {
				type TestStruct struct {
					Name string `json:"name" validate:"required"`
				}

				router := gin.Default()
				router.POST("/test", func(c *gin.Context) {
					var obj TestStruct
					err := validation.BindAndValidate(c, &obj)
					Expect(err).To(BeNil())
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(strings.NewReader(`{"name":"test"}`)) // 使用有效的JSON数据替换
				router.ServeHTTP(w, req)
			})
		})

		Context("with invalid JSON", func() {
			It("should return error", func() {
				type TestStruct struct {
					Name string `json:"name" validate:"required"`
				}

				router := gin.Default()
				router.POST("/test", func(c *gin.Context) {
					var obj TestStruct
					err := validation.BindAndValidate(c, &obj)
					Expect(err).NotTo(BeNil())
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(strings.NewReader(`invalid json`)) // 使用无效的JSON数据替换
				router.ServeHTTP(w, req)
			})
		})

		Context("with invalid struct", func() {
			It("should return validation error", func() {
				type TestStruct struct {
					Name string `json:"name" validate:"required"`
				}

				router := gin.Default()
				router.POST("/test", func(c *gin.Context) {
					var obj TestStruct
					err := validation.BindAndValidate(c, &obj)
					Expect(err).NotTo(BeNil())
				})

				w := httptest.NewRecorder()
				req, _ := http.NewRequest("POST", "/test", nil)
				req.Header.Set("Content-Type", "application/json")
				req.Body = io.NopCloser(strings.NewReader(`{}`)) // // 使用缺少必填字段的JSON数据替换
				router.ServeHTTP(w, req)
			})
		})
	})

	Describe("ValidateStruct", func() {
		Context("with valid struct", func() {
			It("should not return error", func() {
				type TestStruct struct {
					Name string `validate:"required"`
				}

				obj := TestStruct{Name: "test"}
				err := validation.ValidateStruct(context.Background(), &obj)
				Expect(err).To(BeNil())
			})
		})

		Context("with invalid struct", func() {
			It("should return validation error", func() {
				type TestStruct struct {
					Name string `validate:"required"`
				}

				obj := TestStruct{}
				err := validation.ValidateStruct(context.Background(), &obj)
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
