/*
 * TencentBlueKing is pleased to support the open source community by making
 * BlueKing - Micro APIGateway available.
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

package handler

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	apicommon "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/common"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

func toTypedResourceModel[T any](
	resource *model.ResourceCommonModel,
	resourceType constant.APISIXResource,
) (T, error) {
	typed, ok := resource.ToResourceModel(resourceType).(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("unexpected resource model type for %s", resourceType)
	}
	return typed, nil
}

func prepareWebResourceCommonModel(
	c *gin.Context,
	resourceType constant.APISIXResource,
	operation constant.OperationType,
	pathID string,
	name string,
	outerFields map[string]any,
	config json.RawMessage,
	status constant.ResourceStatus,
	creator string,
	updater string,
) (*model.ResourceCommonModel, error) {
	prepared, err := apicommon.PrepareStoredResource(resourcecodec.RequestInput{
		Source:       resourcecodec.SourceWeb,
		Operation:    operation,
		GatewayID:    ginx.GetGatewayInfo(c).ID,
		ResourceType: resourceType,
		Version:      ginx.GetGatewayInfo(c).GetAPISIXVersionX(),
		PathID:       pathID,
		OuterName:    name,
		OuterFields:  outerFields,
		Config:       config,
	})
	if err != nil {
		return nil, err
	}
	return apicommon.BuildResourceCommonModel(prepared, status, creator, updater), nil
}
