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
	"fmt"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/publisher"
)

var batchUpdateResourceStatus = resourcebiz.BatchUpdateResourceStatus

func persistPublishedOperations(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceIDs []string,
	ops []publisher.ResourceOperation,
	errMessage string,
) error {
	if err := batchCreateEtcdResource(ctx, ops); err != nil {
		return err
	}
	if err := batchUpdateResourceStatus(
		ctx,
		resourceType,
		resourceIDs,
		constant.ResourceStatusSuccess,
	); err != nil {
		logging.ErrorFWithContext(ctx, "%s status change err: %s", resourceType, err.Error())
		return fmt.Errorf("%s：%w", errMessage, err)
	}
	return nil
}
