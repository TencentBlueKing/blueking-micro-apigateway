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

package publish

import (
	"context"
	"encoding/json"
	"testing"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/publisher"
)

func TestPersistPublishedOperations(t *testing.T) {
	t.Parallel()

	var (
		calledCreate bool
		calledStatus bool
	)

	patches := gomonkey.ApplyFunc(
		batchCreateEtcdResource,
		func(context.Context, []publisher.ResourceOperation) error {
			calledCreate = true
			return nil
		},
	)
	defer patches.Reset()

	patches.ApplyFunc(
		batchUpdateResourceStatus,
		func(
			context.Context,
			constant.APISIXResource,
			[]string,
			constant.ResourceStatus,
		) error {
			calledStatus = true
			return nil
		},
	)

	err := persistPublishedOperations(
		context.Background(),
		constant.PluginConfig,
		[]string{"pc-id"},
		[]publisher.ResourceOperation{
			{
				Type:   constant.PluginConfig,
				Key:    "pc-id",
				Config: json.RawMessage(`{"id":"pc-id","plugins":{}}`),
			},
		},
		"插件组发布错误",
	)
	assert.NoError(t, err)
	assert.True(t, calledCreate)
	assert.True(t, calledStatus)
}
