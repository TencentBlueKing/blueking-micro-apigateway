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

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

package storage

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// StorageInterface ...
type StorageInterface interface {
	Get(ctx context.Context, key string) (string, error)
	List(ctx context.Context, key string) ([]KeyValuePair, error)
	Create(ctx context.Context, key, val string) error
	Update(ctx context.Context, key, val string) error
	BatchDelete(ctx context.Context, keys []string) error
	BatchCreate(ctx context.Context, resource map[string]string) error
	Watch(ctx context.Context, key string) <-chan WatchResponse
	Close() error

	// NOTE: this is a temporary method to get the etcd client
	GetClient() *clientv3.Client
}

// WatchResponse ...
type WatchResponse struct {
	Events   []Event
	Error    error
	Canceled bool
}

// KeyValuePair ...
type KeyValuePair struct {
	Key         string
	Value       string
	ModRevision int64
}

// Event ...
type Event struct {
	KeyValuePair
	Type EventType
}

// EventType ...
type EventType string

// EventTypePut ...
var (
	EventTypePut    EventType = "put"
	EventTypeDelete EventType = "delete"
)
