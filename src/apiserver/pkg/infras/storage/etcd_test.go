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

package storage

// Add this import
import (
	"context"
	"reflect"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
)

type MockKV struct {
	mock.Mock
}

func (m *MockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	args := m.Called(ctx, key, opts)
	return args.Get(0).(*clientv3.GetResponse), args.Error(1)
}

func (m *MockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	args := m.Called(ctx, key, val, opts)
	return args.Get(0).(*clientv3.PutResponse), args.Error(1)
}

func (m *MockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	args := m.Called(ctx, key, opts)
	return args.Get(0).(*clientv3.DeleteResponse), args.Error(1)
}

func (m *MockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	args := m.Called(ctx, op)
	return args.Get(0).(clientv3.OpResponse), args.Error(1)
}

func (m *MockKV) Txn(ctx context.Context) clientv3.Txn {
	args := m.Called(ctx)
	return args.Get(0).(clientv3.Txn)
}

func (m *MockKV) Compact(
	ctx context.Context,
	rev int64,
	opts ...clientv3.CompactOption,
) (*clientv3.CompactResponse, error) {
	args := m.Called(ctx, rev, opts)
	return args.Get(0).(*clientv3.CompactResponse), args.Error(1)
}

var _ = Describe("EtcdV3Storage", func() {
	Context("NewEtcdStorage", func() {
		It("NewEtcdStorage: ok", func() {
			patches := gomonkey.ApplyFunc(initEtcdClient, func(base.EtcdConfig) (*clientv3.Client, error) {
				return &clientv3.Client{}, nil
			})
			defer patches.Reset()

			c, err := NewEtcdStorage(base.EtcdConfig{})
			assert.NoError(GinkgoT(), err)
			assert.NotNil(GinkgoT(), c)
		})

		It("NewEtcdStorage: error", func() {
			patches := gomonkey.ApplyFunc(initEtcdClient, func(base.EtcdConfig) (*clientv3.Client, error) {
				return nil, errors.New("error")
			})
			defer patches.Reset()

			c, err := NewEtcdStorage(base.EtcdConfig{})
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), c)
		})
	})

	Context("EtcdV3Storage", func() {
		Context("Get", func() {
			It("Get: ok", func() {
				kv := &MockKV{}
				resp := &clientv3.GetResponse{
					Kvs:   []*mvccpb.KeyValue{{Value: []byte("value")}},
					Count: 1,
				}
				kv.On("Get", context.Background(), "/key", []clientv3.OpOption(nil)).Return(resp, nil)

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client}

				_, err := etcd.Get(context.Background(), "key")
				assert.NoError(GinkgoT(), err)
			})

			It("Get: error", func() {
				kv := &MockKV{}
				kv.On("Get", context.Background(), "/key", []clientv3.OpOption(nil)).
					Return(&clientv3.GetResponse{}, errors.New("error"))

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client}

				_, err := etcd.Get(context.Background(), "key")
				assert.Error(GinkgoT(), err)
			})
		})

		Context("Create", func() {
			It("Create: ok", func() {
				kv := &MockKV{}
				kv.On("Put", context.Background(), "/key", "value", []clientv3.OpOption(nil)).
					Return(&clientv3.PutResponse{}, nil)

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client}

				err := etcd.Create(context.Background(), "key", "value")
				assert.NoError(GinkgoT(), err)
			})

			It("Create: error", func() {
				kv := &MockKV{}
				kv.On("Put", context.Background(), "/key", "value", []clientv3.OpOption(nil)).
					Return(&clientv3.PutResponse{}, errors.New("error"))

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client}

				err := etcd.Create(context.Background(), "key", "value")
				assert.Error(GinkgoT(), err)
			})
		})

		Context("Update", func() {
			It("Update: ok", func() {
				kv := &MockKV{}
				kv.On("Put", context.Background(), "/key", "value", []clientv3.OpOption(nil)).
					Return(&clientv3.PutResponse{}, nil)

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client}

				err := etcd.Update(context.Background(), "key", "value")
				assert.NoError(GinkgoT(), err)
			})

			It("Update: error", func() {
				kv := &MockKV{}
				kv.On("Put", context.Background(), "/key", "value", []clientv3.OpOption(nil)).
					Return(&clientv3.PutResponse{}, errors.New("error"))

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client}

				err := etcd.Update(context.Background(), "key", "value")
				assert.Error(GinkgoT(), err)
			})
		})

		Context("List", func() {
			It("List: ok", func() {
				kv := &MockKV{}
				resp := &clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{Key: []byte("/key1"), Value: []byte("value1")},
						{Key: []byte("/key2"), Value: []byte("value2")},
					},
					Count: 2,
				}
				kv.On("Get", context.Background(), "/prefix", mock.Anything).Return(resp, nil)

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client, prefix: "/prefix"}

				result, err := etcd.List(context.Background(), "/prefix")
				assert.NoError(GinkgoT(), err)
				assert.Len(GinkgoT(), result, 2)
				assert.Equal(GinkgoT(), "value1", result[0].Value)
				assert.Equal(GinkgoT(), "value2", result[1].Value)
			})

			It("List: error", func() {
				kv := &MockKV{}
				kv.On("Get", context.Background(), "/prefix", mock.Anything).
					Return(&clientv3.GetResponse{}, errors.New("error"))

				client := &clientv3.Client{KV: kv}
				etcd := &EtcdV3Storage{client: client, prefix: "/prefix"}

				_, err := etcd.List(context.Background(), "/prefix")
				assert.Error(GinkgoT(), err)
			})
		})

		Context("Close", func() {
			It("Close: ok", func() {
				client := &clientv3.Client{}
				patches := gomonkey.ApplyMethod(reflect.TypeOf(client), "Close", func(*clientv3.Client) error {
					return nil
				})
				defer patches.Reset()

				etcd := &EtcdV3Storage{client: client}

				err := etcd.Close()
				assert.NoError(GinkgoT(), err)
			})
		})

		Context("GetClient", func() {
			It("GetClient: ok", func() {
				client := &clientv3.Client{}
				etcd := &EtcdV3Storage{client: client}

				result := etcd.GetClient()
				assert.Equal(GinkgoT(), client, result)
			})
		})
	})
})
