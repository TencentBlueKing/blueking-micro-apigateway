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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// GlobalRuleCreate ...
//
//	@ID			global_rule_create
//	@Summary	global_rule 创建
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.global_rule
//	@Param		gateway_id	path	int							true	"网关 ID"
//	@Param		request		body	serializer.GlobalRuleInfo	true	"global_rule 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/global_rules/ [post]
func GlobalRuleCreate(c *gin.Context) {
	var req serializer.GlobalRuleInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	globalRule := model.GlobalRule{
		Name: req.Name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        idx.GenResourceID(constant.GlobalRule),
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(req.Config),
			Status:    constant.ResourceStatusCreateDraft,
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserID(c),
				Updater: ginx.GetUserID(c),
			},
		},
	}
	if err := biz.CreateGlobalRule(c.Request.Context(), globalRule); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// GlobalRuleUpdate ...
//
//	@ID			global_rule_update
//	@Summary	global_rule 更新
//	@Accept		json
//	@Produce	json
//	@Tags		webapi.global_rule
//	@Param		gateway_id	path	int							true	"网关 ID"	@Param	id	path	string	true	"global_rule ID"
//	@Param		request		body	serializer.GlobalRuleInfo	true	"global_rule 更新参数"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/global_rules/{id}/ [put]
func GlobalRuleUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	req := serializer.GlobalRuleInfo{ID: pathParam.ID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}

	// if resource not changed (config and extra fields), return success directly
	if !biz.IsResourceChanged(c.Request.Context(), constant.GlobalRule, pathParam.ID, req.Config, map[string]any{
		"name": req.Name,
	}) {
		ginx.SuccessNoContentResponse(c)
		return
	}

	updateStatus, err := biz.GetResourceUpdateStatus(c.Request.Context(), constant.GlobalRule, pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	globalRule := model.GlobalRule{
		Name: req.Name,
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        pathParam.ID,
			GatewayID: pathParam.GatewayID,
			Config:    datatypes.JSON(req.Config),
			Status:    updateStatus,
			BaseModel: model.BaseModel{
				Updater: ginx.GetUserID(c),
			},
		},
	}

	if err := biz.UpdateGlobalRule(c.Request.Context(), globalRule); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// GlobalRuleList ...
//
//	@ID			global_rule_list
//	@Summary	global_rule 列表
//	@Produce	json
//	@Tags		webapi.global_rule
//	@Param		gateway_id	path		int								true	"网关 ID"
//	@Param		request		query		serializer.UpstreamListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.GlobalRuleListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/global_rules/ [get]
func GlobalRuleList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.GlobalRuleListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	queryParam := map[string]any{}
	if req.ID != "" {
		queryParam["id"] = req.ID
	}
	globalRules, total, err := biz.ListPagedGlobalRules(
		c.Request.Context(),
		queryParam,
		strings.Split(req.Status, ","),
		req.Name,
		req.Updater,
		req.OrderBy,
		biz.PageParam{
			Offset: ginx.GetOffset(c),
			Limit:  ginx.GetLimit(c),
		},
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	var results serializer.GlobalRuleListResponse
	for _, globalRule := range globalRules {
		results = append(results, serializer.GlobalRuleOutputInfo{
			AutoID:    globalRule.AutoID,
			ID:        globalRule.ID,
			GatewayID: globalRule.GatewayID,
			GlobalRuleInfo: serializer.GlobalRuleInfo{
				ID:     globalRule.ID,
				Name:   globalRule.Name,
				Config: json.RawMessage(globalRule.Config),
			},
			Status:    globalRule.Status,
			CreatedAt: globalRule.CreatedAt.Unix(),
			UpdatedAt: globalRule.UpdatedAt.Unix(),
			Creator:   globalRule.Creator,
			Updater:   globalRule.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// GlobalRuleGet ...
//
//	@ID			global_rule_get
//	@Summary	global_rule 详情
//	@Produce	json
//	@Tags		webapi.global_rule
//	@Param		gateway_id	path		int		true	"网关 id"
//	@Param		id			path		string	true	"资源 ID"
//	@Success	200			{object}	serializer.GlobalRuleOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/global_rules/{id}/ [get]
func GlobalRuleGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	globalRule, err := biz.GetGlobalRule(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.GlobalRuleOutputInfo{
		GatewayID: globalRule.GatewayID,
		AutoID:    globalRule.AutoID,
		ID:        globalRule.ID,
		GlobalRuleInfo: serializer.GlobalRuleInfo{
			ID:     globalRule.ID,
			Name:   globalRule.Name,
			Config: json.RawMessage(globalRule.Config),
		},
		CreatedAt: globalRule.CreatedAt.Unix(),
		UpdatedAt: globalRule.UpdatedAt.Unix(),
		Creator:   globalRule.Creator,
		Updater:   globalRule.Updater,
		Status:    globalRule.Status,
	}
	ginx.SuccessJSONResponse(c, output)
}

// GlobalRuleDelete ...
//
//	@ID			global_rule_delete
//	@Summary	global_rule 删除
//	@Produce	json
//	@Tags		webapi.global_rule
//	@Param		gateway_id	path	int		true	"网关 id"
//	@Param		id			path	string	true	"资源 ID"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/global_rules/{id}/ [delete]
func GlobalRuleDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	globalRule, err := biz.GetGlobalRule(c.Request.Context(), pathParam.ID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// create_draft 状态可以直接删除
	if globalRule.Status == constant.ResourceStatusCreateDraft {
		err = biz.BatchDeleteGlobalRules(c.Request.Context(), []string{globalRule.ID})
		if err != nil {
			ginx.SystemErrorJSONResponse(c, err)
			return
		}
		ginx.SuccessNoContentResponse(c)
		return
	}
	err = biz.UpdateResourceStatusWithAuditLog(c.Request.Context(),
		constant.GlobalRule, globalRule.ID, constant.ResourceStatusDeleteDraft)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// GlobalRuleDropDownList ...
//
//	@ID			global_rule_dropdown_list
//	@Summary	global_rule 下拉列表
//	@Produce	json
//	@Tags		webapi.global_rule
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.GlobalRuleDropDownListResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/global_rules-dropdown/ [get]
func GlobalRuleDropDownList(c *gin.Context) {
	rules, err := biz.ListGlobalRules(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}

	var output serializer.GlobalRuleDropDownListResponse
	for _, rule := range rules {
		desc := gjson.ParseBytes(rule.Config).Get("desc").String()
		output = append(output, serializer.GlobalRuleDropDownOutputInfo{
			AutoID: rule.AutoID,
			ID:     rule.ID,
			Name:   rule.Name,
			Desc:   desc,
		})
	}

	ginx.SuccessJSONResponse(c, output)
}

// GlobalRulePlugins ...
//
//	@ID			global_rule_plugins
//	@Summary	获取 global rule 配置的插件列表
//	@Produce	json
//	@Tags		webapi.global_rule
//	@Param		gateway_id	path		int	true	"网关 ID"
//	@Success	200			{object}	serializer.GlobalRulePluginsResponse
//	@Router		/api/v1/web/gateways/{gateway_id}/global_rules/-/plugins/ [get]
func GlobalRulePlugins(c *gin.Context) {
	globalRuleToIDMap, err := biz.GetGlobalRulePluginToID(c.Request.Context())
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.GlobalRulePluginsResponse{
		Plugins:    []json.RawMessage{},
		PluginRule: make(map[string]string),
	}
	for name, rule := range globalRuleToIDMap {
		output.Plugins = append(output.Plugins, rule.Config)
		output.PluginRule[name] = rule.ID
	}
	ginx.SuccessJSONResponse(c, output)
}
