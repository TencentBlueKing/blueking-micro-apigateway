/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
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

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

func bindAndValidateWebCreateWithGeneratedID(
	c *gin.Context,
	req any,
	resourceType constant.APISIXResource,
	setResourceID func(string),
) error {
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}
	// A typed callback keeps the hot path simple and avoids reflective field mutation.
	setResourceID(idx.GenResourceID(resourceType))
	return validation.ValidateStruct(c.Request.Context(), req)
}

func buildWebCreateDraft(
	c *gin.Context,
	resourceID string,
	config json.RawMessage,
) model.ResourceCommonModel {
	userID := ginx.GetUserID(c)
	return model.ResourceCommonModel{
		ID:        resourceID,
		GatewayID: ginx.GetGatewayInfo(c).ID,
		Config:    datatypes.JSON(config),
		Status:    constant.ResourceStatusCreateDraft,
		BaseModel: model.BaseModel{
			Creator: userID,
			Updater: userID,
		},
	}
}
