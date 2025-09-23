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

	validator "github.com/go-playground/validator/v10"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// GlobalRuleInfo GlobalRule 基本信息
type GlobalRuleInfo struct {
	ID   string `json:"-"`                                                 // 资源apisix资源id
	Name string `json:"name" binding:"required" validate:"globalRuleName"` // GlobalRule名称
	// 配置数据(json格式)
	Config json.RawMessage `json:"config" validate:"apisixConfig=global_rule,global_rule_plugins" swaggertype:"object"`
}

// GlobalRuleListRequest GlobalRule 列表请求参数
type GlobalRuleListRequest struct {
	ID      string `json:"id,omitempty" form:"id"`
	Name    string `json:"name,omitempty" form:"name"`
	Updater string `json:"updater,omitempty" form:"updater"`
	Status  string `json:"status" form:"status" binding:"resourceStatus"`
	OrderBy string `json:"order_by" form:"order_by"`
	Offset  int    `json:"offset" form:"offset"`
	Limit   int    `json:"limit" form:"limit"`
}

// GlobalRuleListResponse GlobalRule 列表
type GlobalRuleListResponse []GlobalRuleOutputInfo

// GlobalRuleOutputInfo GlobalRule 基本信息
type GlobalRuleOutputInfo struct {
	AutoID    int    `json:"auto_id"`
	ID        string `json:"id"`
	GatewayID int    `json:"gateway_id"` // 网关 ID
	GlobalRuleInfo
	CreatedAt int64                   `json:"created_at"`
	UpdatedAt int64                   `json:"updated_at"`
	Creator   string                  `json:"creator"`
	Updater   string                  `json:"updater"`
	Status    constant.ResourceStatus `json:"status"` // 发布状态
}

// GlobalRuleDropDownListResponse GlobalRule 下拉列表
type GlobalRuleDropDownListResponse []GlobalRuleDropDownOutputInfo

// GlobalRuleDropDownOutputInfo GlobalRule 下拉列表输出信息
type GlobalRuleDropDownOutputInfo struct {
	AutoID int    `json:"auto_id"` // 自增 ID
	ID     string `json:"id"`      // 资源 apisix 资源 id
	Name   string `json:"name"`    // 路由名称
	Desc   string `json:"desc"`    // 路由描述
}

// GlobalRulePluginsResponse GlobalRule 插件列表
type GlobalRulePluginsResponse struct {
	Plugins    []json.RawMessage `json:"plugins" swaggertype:"array,object"`      // 插件列表
	PluginRule map[string]string `json:"plugin_rule"  swaggertype:"array,object"` // plugin->rule_id 映射
}

// ValidateGlobalRulePlugin 校验配置的 plugin 不能被重复添加到不同的 rule 里面
func ValidateGlobalRulePlugin(ctx context.Context, fl validator.FieldLevel) bool {
	plugins := gjson.ParseBytes(fl.Field().Bytes()).Get("plugins")
	// globalRule 的 plugins 插件只能配置一个
	if len(plugins.Array()) != 1 {
		return false
	}
	var pluginName string
	plugins.ForEach(func(key, value gjson.Result) bool {
		pluginName = key.String()
		return false
	})
	// 校验 plugin 是否重复绑定到不同的 rule
	globalRuleToIDMap, err := biz.GetGlobalRulePluginToID(ctx, ginx.GetGatewayInfoFromContext(ctx).ID)
	if err != nil {
		ginx.GetValidateErrorInfoFromContext(ctx).Err = err
		return false
	}
	for name, globalRulePlugin := range globalRuleToIDMap {
		// 排除自己
		if fl.Parent().FieldByName("ID").String() == globalRulePlugin.ID {
			continue
		}
		if name == pluginName {
			return false
		}
	}
	return true
}

// ValidateGlobalRuleName 校验 global_rule 资源名称
func ValidateGlobalRuleName(ctx context.Context, fl validator.FieldLevel) bool {
	globalRuleName := fl.Field().String()
	if globalRuleName == "" {
		return false
	}
	return biz.DuplicatedResourceName(
		ctx,
		constant.GlobalRule,
		fl.Parent().FieldByName("ID").String(),
		globalRuleName,
	)
}

// 注册校验器
func init() {
	validation.AddBizFieldTagValidatorWithCtx("global_rule_plugins", ValidateGlobalRulePlugin,
		"{0}:{1} 无效: 插件长度应为1，并且插件必须是全局唯一的，请检查插件是否被占用")
	validation.AddBizFieldTagValidatorWithCtx(
		"globalRuleName",
		ValidateGlobalRuleName,
		"{0}: {1} 该资源名称已经被存在的 global_rule 资源占用",
	)
}
