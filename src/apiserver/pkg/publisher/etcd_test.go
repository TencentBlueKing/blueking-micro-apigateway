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

package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage/mock"
)

const (
	validateError    = "validate error"
	batchCreateError = "batch create error"
)

var ctx context.Context

var _ = Describe("EtcdPublisher", func() {
	Describe("NewEtcdPublisher", func() {
		var (
			mockEtcdStore *mock.MockStorageInterface
			ctx           context.Context
			gateway       *model.Gateway
			ctrl          *gomock.Controller
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockEtcdStore = mock.NewMockStorageInterface(ctrl)
			ctx = context.Background()
			gateway = &model.Gateway{}
		})

		It("Test NewEtcdPublisher: ok", func() {
			patches := gomonkey.ApplyFunc(
				storage.NewEtcdStorage,
				func(base.EtcdConfig) (storage.StorageInterface, error) {
					return mockEtcdStore, nil
				},
			)
			defer patches.Reset()

			p, err := NewEtcdPublisher(ctx, gateway)
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), p)
			assert.Equal(GinkgoT(), mockEtcdStore, p.etcdStore)
		})

		It("Test NewEtcdPublisher: fail", func() {
			patches := gomonkey.ApplyFunc(
				storage.NewEtcdStorage,
				func(base.EtcdConfig) (storage.StorageInterface, error) {
					return nil, errors.New("error")
				},
			)
			defer patches.Reset()

			p, err := NewEtcdPublisher(ctx, gateway)
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), p)
		})
	})

	Describe("Test EtcdPublisher", func() {
		var ctrl *gomock.Controller
		var patches *gomonkey.Patches

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
		})

		AfterEach(func() {
			ctrl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		Describe("Get", func() {
			It("Test Get: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Get(gomock.Any(), "key").Return("value", nil)
				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}
				result, err := p.Get(context.Background(), "key")
				assert.NoError(GinkgoT(), err)
				assert.Equal(GinkgoT(), "value", result)
			})

			It("Test Get: fail", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Get(gomock.Any(), "key").Return("", errors.New("error"))
				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}
				result, err := p.Get(context.Background(), "key")
				assert.Error(GinkgoT(), err)
				assert.Empty(GinkgoT(), result)
			})
		})

		Describe("List", func() {
			It("Test List: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().
					List(gomock.Any(), "prefix").
					Return(
						[]storage.KeyValuePair{
							{Key: "key1", Value: "value1"},
							{Key: "key2", Value: "value2"},
						},
						nil,
					)
				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}
				result, err := p.List(context.Background(), "prefix")
				assert.NoError(GinkgoT(), err)
				assert.Equal(
					GinkgoT(),
					[]storage.KeyValuePair{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
					result,
				)
			})

			It("Test List: fail", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().List(gomock.Any(), "prefix").Return(nil, errors.New("error"))
				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}
				result, err := p.List(context.Background(), "prefix")
				assert.Error(GinkgoT(), err)
				assert.Nil(GinkgoT(), result)
			})
		})

		Describe("Create", func() {
			It("Test Create: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Create(gomock.Any(), "/key", "value").Return(nil)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resource := ResourceOperation{
					Key:    "key",
					Config: json.RawMessage("value"),
				}

				err := p.Create(context.Background(), resource)
				assert.NoError(GinkgoT(), err)
			})

			It("Test Create: Validate error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return errors.New(validateError)
					},
				)

				resource := ResourceOperation{
					Key:    "key",
					Config: json.RawMessage("value"),
				}

				err := p.Create(context.Background(), resource)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), validateError, err.Error())
			})

			It("Test Create: Create error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Create(
					gomock.Any(),
					"/key",
					"value",
				).Return(
					errors.New("create error"),
				)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resource := ResourceOperation{
					Key:    "key",
					Config: json.RawMessage("value"),
				}

				err := p.Create(context.Background(), resource)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), "create error", err.Error())
			})
		})

		Describe("Update", func() {
			It("Test Update: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Get(gomock.Any(), "/key").Return("value", nil)
				mockEtcdStore.EXPECT().Update(gomock.Any(), "key", "value").Return(nil)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resource := ResourceOperation{
					Key:    "key",
					Config: json.RawMessage("value"),
				}

				err := p.Update(context.Background(), resource, false)
				assert.NoError(GinkgoT(), err)
			})

			It("Test Update: Validate error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return errors.New(validateError)
					},
				)

				resource := ResourceOperation{
					Key:    "key",
					Config: json.RawMessage("value"),
				}

				err := p.Update(context.Background(), resource, false)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), validateError, err.Error())
			})

			It("Test Update: Get error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Get(gomock.Any(), "/key").Return("", errors.New("get error"))

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resource := ResourceOperation{
					Key:    "key",
					Config: json.RawMessage("value"),
				}

				err := p.Update(context.Background(), resource, false)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), "get error", err.Error())
			})

			It("Test Update: Update error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Get(gomock.Any(), "/key").Return("value", nil)
				mockEtcdStore.EXPECT().Update(
					gomock.Any(),
					"key",
					"value",
				).Return(
					errors.New("update error"),
				)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resource := ResourceOperation{
					Key:    "key",
					Config: json.RawMessage("value"),
				}

				err := p.Update(context.Background(), resource, false)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), "update error", err.Error())
			})
		})

		Describe("BatchCreate", func() {
			It("Test BatchCreate: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().BatchCreate(
					gomock.Any(),
					map[string]string{"/key": "value"},
				).Return(
					nil,
				)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resources := []ResourceOperation{{
					Key:    "key",
					Config: json.RawMessage("value"),
				}}

				err := p.BatchCreate(context.Background(), resources)
				assert.NoError(GinkgoT(), err)
			})

			It("Test BatchCreate: Validate error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return errors.New(validateError)
					},
				)

				resources := []ResourceOperation{{
					Key:    "key",
					Config: json.RawMessage("value"),
				}}

				err := p.BatchCreate(context.Background(), resources)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), validateError, err.Error())
			})

			It("Test BatchCreate: BatchCreate error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().
					BatchCreate(gomock.Any(), map[string]string{"/key": "value"}).
					Return(errors.New(batchCreateError))

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resources := []ResourceOperation{{
					Key:    "key",
					Config: json.RawMessage("value"),
				}}

				err := p.BatchCreate(context.Background(), resources)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), batchCreateError, err.Error())
			})
		})

		Describe("BatchUpdate", func() {
			It("Test BatchUpdate: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().BatchCreate(
					gomock.Any(),
					map[string]string{"/key": "value"},
				).Return(
					nil,
				)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resources := []ResourceOperation{{
					Key:    "key",
					Config: json.RawMessage("value"),
				}}

				err := p.BatchUpdate(context.Background(), resources)
				assert.NoError(GinkgoT(), err)
			})

			It("Test BatchUpdate: Validate error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return errors.New(validateError)
					},
				)

				resources := []ResourceOperation{{
					Key:    "key",
					Config: json.RawMessage("value"),
				}}

				err := p.BatchUpdate(context.Background(), resources)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), validateError, err.Error())
			})

			It("Test BatchUpdate: BatchCreate error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().
					BatchCreate(gomock.Any(), map[string]string{"/key": "value"}).
					Return(errors.New(batchCreateError))

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				patches = gomonkey.ApplyMethod(
					reflect.TypeOf(p),
					"Validate",
					func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
						return nil
					},
				)

				resources := []ResourceOperation{{
					Key:    "key",
					Config: json.RawMessage("value"),
				}}

				err := p.BatchUpdate(context.Background(), resources)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), batchCreateError, err.Error())
			})
		})

		Describe("BatchDelete", func() {
			It("Test BatchDelete: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().BatchDelete(gomock.Any(), []string{"/key"}).Return(nil)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				resources := []ResourceOperation{{
					Key: "key",
				}}

				err := p.BatchDelete(context.Background(), resources)
				assert.NoError(GinkgoT(), err)
			})

			It("Test BatchDelete: BatchDelete error", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().
					BatchDelete(gomock.Any(), []string{"/key"}).
					Return(errors.New("batch delete error"))

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				resources := []ResourceOperation{{
					Key: "key",
				}}

				err := p.BatchDelete(context.Background(), resources)
				assert.Error(GinkgoT(), err)
				assert.Equal(GinkgoT(), "batch delete error", err.Error())
			})
		})

		Describe("Close", func() {
			It("Test Close: ok", func() {
				mockEtcdStore := mock.NewMockStorageInterface(ctrl)
				mockEtcdStore.EXPECT().Close().Return(nil)

				p := &EtcdPublisher{
					etcdStore: mockEtcdStore,
				}

				err := p.Close()
				assert.NoError(GinkgoT(), err)
			})
		})
	})
})
