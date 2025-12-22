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

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/orderedmap"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// Enum ...
//
//	@ID			get_enum_list
//	@Summary	获取系统枚举值
//	@Tags		basic
//	@Success	200	{object}	map[string]interface{}
//	@Router		/api/v1/web/enums/ [get]
func Enum(c *gin.Context) {
	// 资源类型有序返回
	resourceTypeOrderedMap := orderedmap.New()
	for _, resType := range constant.ResourceTypeOrder {
		resourceTypeOrderedMap.Set(resType.String(), constant.ResourceTypeMap[resType])
	}
	constants := map[string]any{
		"gateway_mode":           constant.GatewayModeMap,
		"resource_status":        constant.ResourceStatusMap,
		"synced_resource_status": constant.SyncedResourceStatusMap,
		"upload_status":          constant.UploadResourceStatusMap,
		"apisix_type":            constant.APISIXTypeMap,
		"resource_type":          resourceTypeOrderedMap,
		"operation_type":         constant.OperationTypeMap,
		"support_apisix_version": schema.GetSupportVersionMap(),
	}
	ginx.SuccessJSONResponse(c, constants)
}
