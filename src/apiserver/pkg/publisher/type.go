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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// ResourceOperation ...
type ResourceOperation struct {
	Key    string
	Config json.RawMessage
	Type   constant.APISIXResource
}

// GetKey 获取key
func (r *ResourceOperation) GetKey() string {
	return constant.ResourceTypePrefixMap[r.Type] + "/" + r.Key
}

// PInterface ...
type PInterface interface {
	Get(ctx context.Context, key string) (any, error)
	List(ctx context.Context, prefix string) (any, error)
	Create(ctx context.Context, resource ResourceOperation) error
	Update(ctx context.Context, resource ResourceOperation, createIfNotExist bool) error
	BatchCreate(ctx context.Context, resources []ResourceOperation) error
	BatchUpdate(ctx context.Context, resources []ResourceOperation) error
	BatchDelete(ctx context.Context, resources []ResourceOperation) error
}
