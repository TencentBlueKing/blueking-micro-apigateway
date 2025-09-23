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

// Package publisher ...
package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/version"
)

// EtcdPublisher ...
type EtcdPublisher struct {
	ctx       context.Context
	etcdStore storage.StorageInterface // etcd client
	// nolint:unused
	cancel context.CancelFunc
	Prefix string
	// nolint:unused
	closing     bool
	gatewayInfo *model.Gateway
}

var _ PInterface = &EtcdPublisher{}

// NewEtcdPublisher 创建 etcd publisher
func NewEtcdPublisher(ctx context.Context, gatewayInfo *model.Gateway) (*EtcdPublisher, error) {
	etcdStore, err := storage.NewEtcdStorage(gatewayInfo.EtcdConfig.EtcdConfig)
	if err != nil {
		log.ErrorFWithContext(ctx, "init etcd failed: %s", err)
		return nil, fmt.Errorf("init etcd failed: %s", err)
	}
	return &EtcdPublisher{
		ctx:         ctx,
		Prefix:      gatewayInfo.EtcdConfig.Prefix,
		etcdStore:   etcdStore,
		gatewayInfo: gatewayInfo,
	}, nil
}

// Get 获取
func (s *EtcdPublisher) Get(ctx context.Context, key string) (any, error) {
	ret, err := s.etcdStore.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// List 获取
func (s *EtcdPublisher) List(ctx context.Context, prefix string) (any, error) {
	return s.etcdStore.List(ctx, prefix)
}

// Validate 验证
func (s *EtcdPublisher) Validate(resourceType constant.APISIXResource, config json.RawMessage) (err error) {
	apisixVersion, _ := version.ToXVersion(s.gatewayInfo.APISIXVersion)
	customizePluginSchemaMap := GetCustomizePluginSchemaMap(s.ctx, s.gatewayInfo.ID)
	validator, err := schema.NewAPISIXJsonSchemaValidator(
		apisixVersion,
		resourceType,
		"main."+string(resourceType),
		customizePluginSchemaMap,
		constant.ETCD,
	)
	if err != nil {
		return err
	}
	return validator.Validate(config)
}

// Create 创建
func (s *EtcdPublisher) Create(ctx context.Context, resource ResourceOperation) error {
	if err := s.Validate(resource.Type, resource.Config); err != nil {
		return err
	}

	if err := s.etcdStore.Create(ctx, resource.GetKey(), string(resource.Config)); err != nil {
		return err
	}

	return nil
}

// Update 更新
func (s *EtcdPublisher) Update(ctx context.Context, resource ResourceOperation, createIfNotExist bool) error {
	if err := s.Validate(resource.Type, resource.Config); err != nil {
		return err
	}
	// 如果不存在不更新的话
	if !createIfNotExist {
		_, err := s.Get(ctx, resource.GetKey())
		if err != nil {
			return err
		}
	}

	if err := s.etcdStore.Update(ctx, resource.Key, string(resource.Config)); err != nil {
		return err
	}

	return nil
}

// BatchCreate 批量创建
func (s *EtcdPublisher) BatchCreate(ctx context.Context, resources []ResourceOperation) error {
	resourcesMap := make(map[string]string)
	for _, resource := range resources {
		if err := s.Validate(resource.Type, resource.Config); err != nil {
			return err
		}
		resourcesMap[resource.GetKey()] = string(resource.Config)
	}
	if err := s.etcdStore.BatchCreate(ctx, resourcesMap); err != nil {
		return err
	}
	return nil
}

// BatchUpdate 批量更新
func (s *EtcdPublisher) BatchUpdate(ctx context.Context, resources []ResourceOperation) error {
	resourcesMap := make(map[string]string)
	for _, resource := range resources {
		if err := s.Validate(resource.Type, resource.Config); err != nil {
			return err
		}
		resourcesMap[resource.GetKey()] = string(resource.Config)
	}
	if err := s.etcdStore.BatchCreate(ctx, resourcesMap); err != nil {
		return err
	}
	return nil
}

// BatchDelete 批量删除
func (s *EtcdPublisher) BatchDelete(ctx context.Context, resources []ResourceOperation) error {
	keys := make([]string, 0, len(resources))
	for _, resource := range resources {
		keys = append(keys, resource.GetKey())
	}
	return s.etcdStore.BatchDelete(ctx, keys)
}

// nolint:unused
// func (s *EtcdPublisher) listAndWatch() error {
// 	lc, lcancel := context.WithTimeout(context.TODO(), 5*time.Second)
// 	defer lcancel()
// 	ret, err := s.etcdStore.List(lc, s.Prefix)
// 	if err != nil {
// 		return err
// 	}
// 	for i := range ret {
// 		key := ret[i].Key[len(s.Prefix)+1:]
// 		print(key)
// 		// todo: 同步逻辑
// 	}

// 	// start watch
// 	s.cancel = s.watch()

// 	return nil
// }

// nolint:unused
// func (s *EtcdPublisher) watch() context.CancelFunc {
// 	c, cancel := context.WithCancel(context.TODO())
// 	ch := s.etcdStore.Watch(c, s.Prefix)
// 	go func() {
// 		defer func() {
// 			if !s.closing {
// 				log.Errorf("etcd watch exception closed, restarting: prefix: %s", s.Prefix)
// 			}
// 		}()
// 		defer runtime.HandlePanic()
// 		for event := range ch {
// 			if event.Canceled {
// 				log.Warnf("etcd watch failed: %s", event.Error)
// 				return
// 			}
// 			for i := range event.Events {
// 				switch event.Events[i].Type {
// 				case storage.EventTypePut:
// 					// todo: 同步逻辑
// 				case storage.EventTypeDelete:
// 					// todo: 同步逻辑
// 				}
// 			}
// 		}
// 	}()
// 	return cancel
// }

// Close 关闭
func (s *EtcdPublisher) Close() error {
	return s.etcdStore.Close()
}

// GetCustomizePluginSchemaMap is duplicated with biz.GetCustomizePluginSchemaMap,
// because the biz and publisher are in the same layer, so we can directly call the biz function
// FIXME: but it's not a good practice, so we need to move the function to the right place
func GetCustomizePluginSchemaMap(ctx context.Context, gatewayID int) map[string]any {
	u := repo.GatewayCustomPluginSchema
	schemaList, err := repo.GatewayCustomPluginSchema.WithContext(ctx).Where(u.GatewayID.Eq(gatewayID)).Find()
	if err != nil {
		return nil
	}
	pluginSchemaMap := map[string]any{}
	for _, s := range schemaList {
		var schemaInfo map[string]any
		err = json.Unmarshal(s.Schema, &schemaInfo)
		if err != nil {
			return nil
		}
		pluginSchemaMap[s.Name] = schemaInfo
	}
	return pluginSchemaMap
}
