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

package serializer

import (
	"context"
	"encoding/json"
	"fmt"

	validator "github.com/go-playground/validator/v10"
	"github.com/spf13/cast"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap/buffer"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// PluginSchemaRequest ...
type PluginSchemaRequest struct {
	Name       string `json:"name"`
	SchemaType string `json:"schema_type" form:"schema_type"` // 插件schema类型：metadata/consumer/不传就获取完整schema
}

// ResourceSchemaRequest ...
type ResourceSchemaRequest struct {
	Type string `json:"type" uri:"type" binding:"required"` // 资源名称:service/route/global_rule等
}

// SchemaInfo ...
type SchemaInfo struct {
	AutoID  int             `json:"auto_id"`                                       // 自增ID
	Name    string          `json:"name" binding:"required" validate:"schemaName"` // 插件名称
	Schema  json.RawMessage `json:"schema"  swaggertype:"object"`                  // 插件 schema (json格式)
	Example json.RawMessage `json:"example" swaggertype:"object"`                  // 插件示例 (json格式)
}

// SchemaListRequest ...
type SchemaListRequest struct {
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// SchemaListResponse ...
type SchemaListResponse []SchemaOutputInfo

// SchemaOutputInfo ...
type SchemaOutputInfo struct {
	GatewayID int `json:"gateway_id"` // 网关 ID
	SchemaInfo
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Creator   string `json:"creator"`
	Updater   string `json:"updater"`
}

// ValidationSchemaName 验证插件名称
func ValidationSchemaName(ctx context.Context, fl validator.FieldLevel) bool {
	schemaName := fl.Field().String()
	if schemaName == "" {
		return false
	}
	// 查询插件名称是否与官方插件重复
	schemaInfo := schema.GetPluginSchema(ginx.GetGatewayInfoFromContext(ctx).GetAPISIXVersionX(), schemaName, "")
	if schemaInfo != nil {
		return false
	}
	return biz.DuplicatedSchemaName(
		ctx,
		cast.ToInt(fl.Parent().FieldByName("AutoID").Int()),
		schemaName,
	)
}

// CheckPluginSchemaAndExample 检查 schema 和 example 配置
func CheckPluginSchemaAndExample(schema json.RawMessage, example json.RawMessage) error {
	schemaRaw, _ := schema.MarshalJSON()
	exampleRaw, _ := example.MarshalJSON()
	if schema == nil || jsonx.IsJSONEmpty(schemaRaw) {
		return fmt.Errorf("schema 不可为空")
	}
	if example == nil || jsonx.IsJSONEmpty(exampleRaw) {
		return fmt.Errorf("插件示例不可为空")
	}
	s, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(schemaRaw)))
	if err != nil {
		return fmt.Errorf("实例化 schema 失败: %s", err)
	}
	ret, err := s.Validate(gojsonschema.NewBytesLoader(exampleRaw))
	if err != nil {
		return fmt.Errorf("插件示例验证失败: %s", err)
	}
	if !ret.Valid() {
		errString := buffer.Buffer{}
		for i, vErr := range ret.Errors() {
			if i != 0 {
				errString.AppendString("\n")
			}
			errString.AppendString(vErr.String())
		}
		return fmt.Errorf("schema 验证失败: %s", errString.String())
	}
	return nil
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx(
		"schemaName",
		ValidationSchemaName,
		"{0}: {1} 该插件名称已创建, 或官方插件已存在",
	)
}
