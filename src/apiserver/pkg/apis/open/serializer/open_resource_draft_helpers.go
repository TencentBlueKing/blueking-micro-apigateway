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

package serializer

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
)

// buildOpenUpdateDraft assumes caller already wrote any outer name back into config.
func buildOpenUpdateDraft(
	c *gin.Context,
	resourceID string,
	status constant.ResourceStatus,
	config json.RawMessage,
) *model.ResourceCommonModel {
	return &model.ResourceCommonModel{
		ID:        resourceID,
		GatewayID: ginx.GetGatewayInfo(c).ID,
		Config:    datatypes.JSON(config),
		Status:    status,
		BaseModel: model.BaseModel{
			Updater: ginx.GetUserID(c),
		},
	}
}
