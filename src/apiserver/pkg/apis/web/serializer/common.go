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
	"context"
	"encoding/json"
	"fmt"

	validator "github.com/go-playground/validator/v10"
	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// ResourceCommonPathParam 资源通用参数
type ResourceCommonPathParam struct {
	ID        string                  `json:"id" uri:"id"`
	AutoID    int                     `json:"auto_id" uri:"auto_id"`
	GatewayID int                     `json:"gateway_id" uri:"gateway_id" binding:"required"`
	Type      constant.APISIXResource `json:"type" uri:"type"`
}

// CheckAPISIXConfig 校验 APISIX 配置 schema
func CheckAPISIXConfig(ctx context.Context, fl validator.FieldLevel) bool {
	rawConfig, ok := fl.Field().Interface().(json.RawMessage)
	if !ok {
		return false
	}
	if jsonx.IsJSONEmpty(rawConfig) {
		return false
	}
	resourceType := fl.Param()
	resourceIdentification := schema.GetResourceIdentification(rawConfig)
	if resourceIdentification == "" {
		// 兼容第一次创建没有 id 的情况以及 rawConfig 没有 name 的情况
		resourceIdentification = getResourceNameByResourceType(resourceType, fl)
		rawConfig, _ = sjson.SetBytes(
			rawConfig,
			model.GetResourceNameKey(constant.APISIXResource(resourceType)),
			resourceIdentification,
		)
	}
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	// 基础 schema 校验
	schemaValidator, err := schema.NewAPISIXSchemaValidator(gatewayInfo.GetAPISIXVersionX(), "main."+resourceType)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = fmt.Errorf("resource:%s validate failed, err: %v",
			resourceIdentification, err)
		logging.Errorf("new schema validator failed, err: %v", err)
		return false
	}
	// metadata 校验需要带上插件 name
	if resourceType == constant.PluginMetadata.String() {
		rawConfig, _ = sjson.SetBytes(rawConfig, "id", fl.Parent().FieldByName("Name").String())
	}
	if err = schemaValidator.Validate(rawConfig); err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = err
		logging.Errorf("schema validate failed, err: %v", err)
		return false
	}
	// 配置校验
	customizePluginSchemaMap, err := biz.GetCustomizePluginSchemaMap(ctx)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = fmt.Errorf("resource:%s validate failed, err: %v",
			resourceIdentification, err)
		logging.Errorf("get customize plugin schema map failed, err: %v", err)
		return false
	}
	jsonConfigValidator, err := schema.NewAPISIXJsonSchemaValidator(
		gatewayInfo.GetAPISIXVersionX(),
		constant.APISIXResource(
			resourceType,
		),
		"main."+string(resourceType),
		customizePluginSchemaMap,
		constant.DATABASE,
	)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = fmt.Errorf("resource:%s validate failed, err: %v",
			resourceIdentification, err)
		logging.Errorf("new schema config validator failed, err: %v", err)
		return false
	}
	if err = jsonConfigValidator.Validate(rawConfig); err != nil { // 校验 json schema
		ginx.GetValidateErrorInfoFromContext(ctx).Err = err
		logging.Errorf("json schema validate failed, err: %v", err)
		return false
	}
	return true
}

func getResourceNameByResourceType(resourceType string, fl validator.FieldLevel) string {
	if resourceType == constant.Consumer.String() {
		return fl.Parent().FieldByName("Username").String()
	}
	return fl.Parent().FieldByName("Name").String()
}

// CheckLabel 校验 label
func CheckLabel(label string) (base.LabelMap, error) {
	// 检查标签是否符合 k:v 规范
	if label == "" {
		return nil, nil
	}
	r := &GatewayLabelRequest{Label: base.LabelMap{"label": []string{label}}}
	b, _ := json.Marshal(r)
	var result GatewayLabelRequest
	err := json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return result.Label, nil
}

// PaginateResults 处理分页
func PaginateResults(total, offset, limit int) (int, int) {
	if offset >= total {
		return 0, 0
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return offset, end
}

func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"apisixConfig",
		CheckAPISIXConfig,
		"{0}:{1} 无效",
	)
}
