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

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/resourcecodec"
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
	resourceType := constant.APISIXResource(fl.Param())
	resourceTypeName := resourceType.String()
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	draft, err := resourcecodec.PrepareRequestDraft(resourcecodec.RequestInput{
		Source:       resourcecodec.SourceWeb,
		Operation:    webValidationOperation(fl),
		GatewayID:    gatewayInfo.ID,
		ResourceType: resourceType,
		Version:      gatewayInfo.GetAPISIXVersionX(),
		PathID:       fl.Parent().FieldByName("ID").String(),
		OuterName:    getResourceNameByResourceType(resourceTypeName, fl),
		OuterFields:  webValidationOuterFields(resourceTypeName, fl),
		Config:       rawConfig,
	})
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = err
		return false
	}
	// Request-time validation now targets the prepared DATABASE payload, not raw echoed config fields.
	built, err := resourcecodec.BuildRequestPayload(draft, constant.DATABASE)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = err
		return false
	}
	rawConfig = built.Payload
	resourceIdentification := schema.GetResourceIdentification(rawConfig)
	if resourceIdentification == "" {
		resourceIdentification = draft.Identity.NameValue
	}
	if resourceIdentification == "" {
		resourceIdentification = draft.Identity.ResourceID
	}
	// 基础 schema 校验
	schemaValidator, err := schema.NewAPISIXSchemaValidator(
		gatewayInfo.GetAPISIXVersionX(),
		"main."+resourceTypeName,
	)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = fmt.Errorf("resource:%s validate failed, err: %w",
			resourceIdentification, err)
		logging.Errorf("new schema validator failed, err: %v", err)
		return false
	}
	if err = schemaValidator.Validate(rawConfig); err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = err
		logging.Errorf("schema validate failed, err: %v", err)
		return false
	}
	// 配置校验
	customizePluginSchemaMap, err := biz.GetCustomizePluginSchemaMap(ctx)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = fmt.Errorf("resource:%s validate failed, err: %w",
			resourceIdentification, err)
		logging.Errorf("get customize plugin schema map failed, err: %v", err)
		return false
	}
	jsonConfigValidator, err := schema.NewAPISIXJsonSchemaValidator(
		gatewayInfo.GetAPISIXVersionX(),
		resourceType,
		"main."+resourceTypeName,
		customizePluginSchemaMap,
		constant.DATABASE,
	)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = fmt.Errorf("resource:%s validate failed, err: %w",
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

func webValidationOuterFields(resourceType string, fl validator.FieldLevel) map[string]any {
	fields := map[string]any{}
	if id := fl.Parent().FieldByName("ID").String(); id != "" {
		fields["id"] = id
	}
	for _, fieldName := range []string{"ServiceID", "UpstreamID", "PluginConfigID", "GroupID"} {
		value := fl.Parent().FieldByName(fieldName).String()
		if value == "" {
			continue
		}
		switch fieldName {
		case "ServiceID":
			fields["service_id"] = value
		case "UpstreamID":
			if resourceType == constant.Upstream.String() {
				fields["tls.client_cert_id"] = value
			} else {
				fields["upstream_id"] = value
			}
		case "PluginConfigID":
			fields["plugin_config_id"] = value
		case "GroupID":
			fields["group_id"] = value
		}
	}
	if sslID := fl.Parent().FieldByName("SSLID").String(); sslID != "" {
		fields["tls.client_cert_id"] = sslID
	}
	return fields
}

func webValidationOperation(fl validator.FieldLevel) constant.OperationType {
	if fl.Parent().FieldByName("ID").String() != "" {
		return constant.OperationTypeUpdate
	}
	return constant.OperationTypeCreate
}

// CheckLabel 校验 label
func CheckLabel(label string) (base.LabelMap, error) {
	// 检查标签是否符合 k:v 规范
	if label == "" {
		return nil, nil
	}
	r := &GatewayLabelRequest{Label: base.LabelMap{"label": []string{label}}}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var result GatewayLabelRequest
	err = json.Unmarshal(b, &result)
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
	end := min(offset+limit, total)
	return offset, end
}

func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"apisixConfig",
		CheckAPISIXConfig,
		"{0}:{1} 无效",
	)
}
