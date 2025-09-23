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

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - APIGateway) available.
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

// Package storage ...
package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/runtime"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/tls"
)

// KeyNotFoundError ...
var (
	KeyNotFoundError      = errors.New("key not found")
	ConnectionFailedError = errors.New("连接失败，请检查 etcd 地址是否正确")
	AuthFailedError       = errors.New("用户名或密码错误，或者证书错误，请重新检查后再试")
)

// SkippedValueEtcdInitDir ...
const (

	// SkippedValueEtcdInitDir indicates the init_dir
	// etcd event will be skipped.
	SkippedValueEtcdInitDir = "init_dir"

	// SkippedValueEtcdEmptyObject indicates the data with an
	// empty JSON value {}, which may be set by APISIX,
	// should be also skipped.
	//
	// Important: at present, {} is considered as invalid,
	// but may be changed in the future.
	SkippedValueEtcdEmptyObject = "{}"
	// MaxOperateNum ...
	bulkOperateSize = 100
)

// EtcdV3Storage ...
type EtcdV3Storage struct {
	client *clientv3.Client
	prefix string
}

var _ StorageInterface = &EtcdV3Storage{}

func initEtcdClient(etcdConf base.EtcdConfig) (*clientv3.Client, error) {
	config := clientv3.Config{
		Endpoints:   etcdConf.Endpoint.Endpoints(),
		DialTimeout: 5 * time.Second,
		Username:    etcdConf.Username,
		Password:    etcdConf.Password,
	}
	if etcdConf.CertCert != "" && etcdConf.CACert != "" && etcdConf.CertKey != "" {
		var err error
		config.TLS, err = tls.NewClientTLSConfig(etcdConf.CACert, etcdConf.CertCert, etcdConf.CertKey)
		if err != nil {
			return nil, err
		}
	}
	cli, err := clientv3.New(config)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			err = ConnectionFailedError
		}
		if strings.Contains(err.Error(), "etcdserver: authentication failed, invalid user ID or password") {
			err = AuthFailedError
		}
		log.Errorf("init etcd failed: %s", err)
		return nil, fmt.Errorf("etcd 初始化失败: %s", err)
	}
	return cli, nil
}

// NewEtcdStorage ...
func NewEtcdStorage(etcdConf base.EtcdConfig) (StorageInterface, error) {
	cli, err := initEtcdClient(etcdConf)
	if err != nil {
		log.Errorf("init etcd failed: %s", err)
		return nil, err
	}
	return &EtcdV3Storage{
		client: cli,
		prefix: etcdConf.Prefix,
	}, nil
}

// Get ...
func (e *EtcdV3Storage) Get(ctx context.Context, key string) (string, error) {
	resp, err := e.client.Get(ctx, fmt.Sprintf("%s/%s", e.prefix, key))
	if err != nil {
		log.Errorf("etcd get failed: %s", err)
		return "", fmt.Errorf("etcd get failed: %s", err)
	}
	if resp.Count == 0 {
		log.Warnf("key: %s is not found", key)
		return "", KeyNotFoundError
	}

	return string(resp.Kvs[0].Value), nil
}

func (e *EtcdV3Storage) txnOperate(ctx context.Context, ops []clientv3.Op) error {
	timeoutCtx, cancelFunc := context.WithTimeout(ctx, time.Second*2)
	defer cancelFunc()
	txn := e.client.Txn(timeoutCtx)
	txnRsp, err := txn.Then(ops...).Commit()
	if err != nil {
		return err
	}
	if !txnRsp.Succeeded {
		return errors.New("etcd transaction failed")
	}
	return nil
}

func (e *EtcdV3Storage) txnMultiOperate(ctx context.Context, ops []clientv3.Op) error {
	length := len(ops)
	if length <= bulkOperateSize {
		return e.txnOperate(ctx, ops)
	}

	count := length / bulkOperateSize
	if length%bulkOperateSize > 0 {
		count += 1
	}
	for i := 0; i < count; i++ {
		start := i * bulkOperateSize
		end := start + bulkOperateSize
		if end > length {
			end = length
		}
		err := e.txnOperate(ctx, ops[start:end])
		if err != nil {
			return err
		}
	}
	_ = e.client.Close()
	return nil
}

// List ...
func (e *EtcdV3Storage) List(ctx context.Context, key string) ([]KeyValuePair, error) {
	resp, err := e.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("etcd get failed: %s", err)
		return nil, fmt.Errorf("etcd get failed: %s", err)
	}
	var ret []KeyValuePair
	for i := range resp.Kvs {
		key := string(resp.Kvs[i].Key)
		value := string(resp.Kvs[i].Value)

		// Skip the data if its value is init_dir or {}
		// during fetching-all phase.
		//
		// For more complex cases, an explicit function to determine if
		// skippable would be better.
		if value == SkippedValueEtcdInitDir || value == SkippedValueEtcdEmptyObject {
			continue
		}

		data := KeyValuePair{
			Key:         key,
			Value:       value,
			ModRevision: resp.Kvs[i].ModRevision,
		}
		ret = append(ret, data)
	}

	return ret, nil
}

// Create ...
func (e *EtcdV3Storage) Create(ctx context.Context, key, val string) error {
	_, err := e.client.Put(ctx, fmt.Sprintf("%s/%s", e.prefix, key), val)
	if err != nil {
		log.Errorf("etcd put failed: %s", err)
		return fmt.Errorf("etcd put failed: %s", err)
	}
	return nil
}

// Update ...
func (e *EtcdV3Storage) Update(ctx context.Context, key, val string) error {
	_, err := e.client.Put(ctx, fmt.Sprintf("%s/%s", e.prefix, key), val)
	if err != nil {
		log.Errorf("etcd put failed: %s", err)
		return fmt.Errorf("etcd put failed: %s", err)
	}
	return nil
}

// BatchCreate ...
func (e *EtcdV3Storage) BatchCreate(ctx context.Context, resource map[string]string) error {
	var ops []clientv3.Op
	for k, v := range resource {
		ops = append(ops, clientv3.OpPut(fmt.Sprintf("%s/%s", e.prefix, k), v))
	}
	return e.txnMultiOperate(ctx, ops)
}

// BatchDelete ...
func (e *EtcdV3Storage) BatchDelete(ctx context.Context, keys []string) error {
	var ops []clientv3.Op
	for _, key := range keys {
		ops = append(ops, clientv3.OpDelete(fmt.Sprintf("%s/%s", e.prefix, key)))
	}
	return e.txnMultiOperate(ctx, ops)
}

// Watch ...
func (e *EtcdV3Storage) Watch(ctx context.Context, key string) <-chan WatchResponse {
	// NOTE: should use e.prefix here?
	eventChan := e.client.Watch(ctx, key, clientv3.WithPrefix())
	ch := make(chan WatchResponse, 1)
	go func() {
		defer runtime.HandlePanic()
		for event := range eventChan {
			if event.Err() != nil {
				log.Errorf("etcd watch error: key: %s err: %v", key, event.Err())
				close(ch)
				return
			}

			output := WatchResponse{
				Canceled: event.Canceled,
			}

			for i := range event.Events {
				key := string(event.Events[i].Kv.Key)
				value := string(event.Events[i].Kv.Value)

				// Skip the data if its value is init_dir or {}
				// during watching phase.
				//
				// For more complex cases, an explicit function to determine if
				// skippable would be better.
				if value == SkippedValueEtcdInitDir || value == SkippedValueEtcdEmptyObject {
					continue
				}

				e := Event{
					KeyValuePair: KeyValuePair{
						Key:   key,
						Value: value,
					},
				}
				switch event.Events[i].Type {
				case clientv3.EventTypePut:
					e.Type = EventTypePut
				case clientv3.EventTypeDelete:
					e.Type = EventTypeDelete
				}
				output.Events = append(output.Events, e)
			}
			if output.Canceled {
				log.Error("channel canceled")
				output.Error = fmt.Errorf("channel canceled")
			}
			ch <- output
		}

		close(ch)
	}()

	return ch
}

// Close ...
func (e *EtcdV3Storage) Close() error {
	return e.client.Close()
}

// GetClient ...
func (e *EtcdV3Storage) GetClient() *clientv3.Client {
	return e.client
}
