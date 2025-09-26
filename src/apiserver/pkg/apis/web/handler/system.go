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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

// PluginSchemaGet ...
//
//	@ID			plugin_schema_get
//	@Summary	获取插件 schema
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path		int						true	"网关 id"
//	@Param		name		path		string					true	"插件名称"
//	@Param		schema_type	query		string					false	"schema 类型：metadata/consumer/不传就获取完整 schema"
//	@Success	200			{object}	map[string]interface{}	"schema"
//	@Router		/api/v1/web/gateways/{gateway_id}/schemas/plugins/{name}/ [get]
func PluginSchemaGet(c *gin.Context) {
	var param serializer.PluginSchemaRequest
	if err := c.ShouldBind(&param); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	param.Name = c.Param("name")

	schemaInfo := schema.GetPluginSchema(ginx.GetGatewayInfo(c).GetAPISIXVersionX(), param.Name, param.SchemaType)
	if schemaInfo == nil {
		// 查询自定义插件 schema
		customizePluginSchema, _ := biz.GetSchemaByName(c.Request.Context(), param.Name)
		if customizePluginSchema == nil {
			ginx.NotFoundJSONResponse(c, errors.New("schema not found"))
			return
		}
		schemaInfo = customizePluginSchema.Schema
	}
	ginx.SuccessJSONResponse(c, schemaInfo)
}

// ResourceSchemaGet ...
//
//	@ID			resource_schema_get
//	@Summary	获取资源 schema
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path		int						true	"网关 id"
//	@Param		type		path		string					true	"资源类型:route/global_rule 等"
//	@Success	200			{object}	map[string]interface{}	"schema"
//	@Router		/api/v1/web/gateways/{gateway_id}/schemas/resources/{type}/ [get]
func ResourceSchemaGet(c *gin.Context) {
	var param serializer.ResourceSchemaRequest
	if err := c.ShouldBindUri(&param); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	schemaInfo := schema.GetResourceSchema(ginx.GetGatewayInfo(c).GetAPISIXVersionX(), param.Type)
	if schemaInfo == nil {
		ginx.NotFoundJSONResponse(c, errors.New("schema not found"))
		return
	}
	ginx.SuccessJSONResponse(c, schemaInfo)
}

// SchemaCreate ...
//
//	@ID			schema_create
//	@Summary	创建自定义插件 schema
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path	int						true	"网关 id"
//	@Param		request		body	serializer.SchemaInfo	true	"schema 创建参数"
//	@Success	201
//	@Router		/api/v1/web/gateways/{gateway_id}/schemas/ [post]
func SchemaCreate(c *gin.Context) {
	var req serializer.SchemaInfo
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	err := serializer.CheckPluginSchemaAndExample(req.Schema, req.Example)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	err = biz.CreateSchema(c.Request.Context(), &model.GatewayCustomPluginSchema{
		Name:      req.Name,
		GatewayID: ginx.GetGatewayInfo(c).ID,
		Schema:    datatypes.JSON(req.Schema),
		Example:   datatypes.JSON(req.Example),
		BaseModel: model.BaseModel{
			Creator: ginx.GetUserID(c),
			Updater: ginx.GetUserID(c),
		},
	})
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessCreateResponse(c)
}

// SchemaUpdate ...
//
//	@ID			schema_update
//	@Summary	更新自定义插件 schema
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path	int						true	"网关 id"
//	@Param		auto_id	path	int							true	"插件 id"
//	@Param		request		body	serializer.SchemaInfo	true	"schema 更新参数"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/schemas/{auto_id}/ [put]
func SchemaUpdate(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	req := serializer.SchemaInfo{AutoID: pathParam.AutoID}
	if err := validation.BindAndValidate(c, &req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	// 查询是否有资源关联
	resourceSchemas, err := biz.GetResourceSchemaAssociation(c.Request.Context(), pathParam.AutoID)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	if len(resourceSchemas) != 0 {
		var resourceList []string
		for _, s := range resourceSchemas {
			resourceList = append(resourceList, fmt.Sprintf("%s: %s", s.ResourceType.String(), s.ResourceID))
		}
		ginx.BadRequestErrorJSONResponse(c,
			fmt.Errorf("name: %s 该插件已被 [ %s ] 资源引用, 不可更新", req.Name, strings.Join(resourceList, ", ")),
		)
		return
	}
	err = serializer.CheckPluginSchemaAndExample(req.Schema, req.Example)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	gatewayCustomPluginSchema := model.GatewayCustomPluginSchema{
		AutoID:    pathParam.AutoID,
		Name:      req.Name,
		GatewayID: pathParam.GatewayID,
		Schema:    datatypes.JSON(req.Schema),
		Example:   datatypes.JSON(req.Example),
		BaseModel: model.BaseModel{
			Updater: ginx.GetUserID(c),
		},
	}
	if err = biz.UpdateSchema(c.Request.Context(), gatewayCustomPluginSchema); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// SchemaGet ...
//
//	@ID			schema_get
//	@Summary	查询自定义插件 schema
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path		int	true	"网关 id"
//	@Param		auto_id		path		int	true	"插件 id"
//	@Success	200			{object}	serializer.SchemaOutputInfo
//	@Router		/api/v1/web/gateways/{gateway_id}/schemas/{auto_id}/ [get]
func SchemaGet(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	schemaInfo, err := biz.GetSchemaByID(c.Request.Context(), pathParam.AutoID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	output := serializer.SchemaOutputInfo{
		GatewayID: schemaInfo.GatewayID,
		SchemaInfo: serializer.SchemaInfo{
			AutoID:  schemaInfo.AutoID,
			Name:    schemaInfo.Name,
			Schema:  json.RawMessage(schemaInfo.Schema),
			Example: json.RawMessage(schemaInfo.Example),
		},
		CreatedAt: schemaInfo.CreatedAt.Unix(),
		UpdatedAt: schemaInfo.UpdatedAt.Unix(),
		Creator:   schemaInfo.Creator,
		Updater:   schemaInfo.Updater,
	}

	ginx.SuccessJSONResponse(c, output)
}

// SchemaDelete ...
//
//	@ID			schema_delete
//	@Summary	删除自定义插件 schema
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path	int	true	"网关 id"
//	@Param		auto_id		path	int	true	"插件 id"
//	@Success	204
//	@Router		/api/v1/web/gateways/{gateway_id}/schemas/{auto_id}/ [delete]
func SchemaDelete(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	schemaInfo, err := biz.GetSchemaByID(c.Request.Context(), pathParam.AutoID)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// 查询是否有资源关联
	resourceSchemas, err := biz.GetResourceSchemaAssociation(c.Request.Context(), pathParam.AutoID)
	if err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	if len(resourceSchemas) != 0 {
		var resourceList []string
		for _, s := range resourceSchemas {
			resourceList = append(resourceList, fmt.Sprintf("%s: %s", s.ResourceType.String(), s.ResourceID))
		}
		ginx.BadRequestErrorJSONResponse(c,
			fmt.Errorf("name: %s 该插件已被 [ %s ] 资源引用, 不可删除", schemaInfo.Name, strings.Join(resourceList, ", ")),
		)
		return
	}
	if err := biz.DeleteSchema(c.Request.Context(), pathParam.AutoID); err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	ginx.SuccessNoContentResponse(c)
}

// SchemaList ...
//
//	@ID			schema_list
//	@Summary	自定义插件 schema 列表
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path		int								true	"网关 id"
//	@Param		request		query		serializer.SchemaListRequest	false	"查询参数"
//	@Success	200			{object}	ginx.PaginatedResponse{results=serializer.SchemaListResponse}
//	@Router		/api/v1/web/gateways/{gateway_id}/schemas/ [get]
func SchemaList(c *gin.Context) {
	var pathParam serializer.ResourceCommonPathParam
	if err := c.ShouldBindUri(&pathParam); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	var req serializer.SchemaListRequest
	if err := c.ShouldBind(&req); err != nil {
		ginx.BadRequestErrorJSONResponse(c, err)
		return
	}
	schemaList, total, err := biz.ListPagedSchema(
		c.Request.Context(),
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
	var results serializer.SchemaListResponse
	for _, s := range schemaList {
		results = append(results, serializer.SchemaOutputInfo{
			GatewayID: s.GatewayID,
			SchemaInfo: serializer.SchemaInfo{
				AutoID:  s.AutoID,
				Name:    s.Name,
				Schema:  json.RawMessage(s.Schema),
				Example: json.RawMessage(s.Example),
			},
			CreatedAt: s.CreatedAt.Unix(),
			UpdatedAt: s.UpdatedAt.Unix(),
			Creator:   s.Creator,
			Updater:   s.Updater,
		})
	}
	ginx.SuccessJSONResponse(c, ginx.NewPaginatedRespData(total, results))
}

// PluginsGet ...
//
//	@ID			plugins_get
//	@Summary	获取插件列表
//	@Produce	json
//	@Tags		webapi.system
//	@Param		gateway_id	path		int								true	"网关 id"
//	@Param		kind		query		string							false	"插件类型:plugins/consumer/metadata/stream"
//	@Success	200			{object}	serializer.PluginListResponse	"schema"
//	@Router		/api/v1/web/gateways/{gateway_id}/plugins/ [get]
func PluginsGet(c *gin.Context) {
	version := ginx.GetGatewayInfo(c).GetAPISIXVersionX()
	apisixType := ginx.GetGatewayInfo(c).APISIXType
	plugins, err := schema.GetPlugins(apisixType, version)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	// 查询自定义插件
	customizePluginExampleList, err := biz.GetCustomizePluginExampleList(
		c.Request.Context(),
		ginx.GetGatewayInfo(c).ID,
	)
	if err != nil {
		ginx.SystemErrorJSONResponse(c, err)
		return
	}
	plugins = append(plugins, customizePluginExampleList...)

	kind := c.Query("kind")
	// 按类别分组返回
	pluginTypeMap := make(map[string][]*schema.Plugin)
	for _, plugin := range plugins {
		if kind == constant.Metadata && len(plugin.MetadataExample) == 0 {
			continue
		}
		// 当查询的插件类别为 stream 时，仅获取 StreamRoutePluginMap 匹配的插件
		if kind == constant.Stream && schema.StreamRoutePluginMap[plugin.Name] == "" {
			continue
		}
		// 只有 stream route 资源才能使用 stream 类型的插件，其他资源需要排除掉
		if kind != constant.Stream && plugin.ProxyType == constant.Stream {
			continue
		}
		// 根据 apisixType 过滤
		if apisixType == constant.APISIXTypeAPISIX {
			// apisix 实例需要过滤掉 tapisix 和 bk 插件
			if plugin.Type == constant.APISIXTypeTAPISIX || plugin.Type == constant.APISIXTypeBKAPISIX {
				continue
			}
		}
		if apisixType == constant.APISIXTypeTAPISIX && plugin.Type == constant.APISIXTypeBKAPISIX {
			// tapisix 实例需要过滤掉 bk 插件
			continue
		}
		// 处理特殊插件的文档地址
		if val, ok := constant.SpecialPluginDocMap[plugin.Name]; ok {
			plugin.DocUrl = fmt.Sprintf(schema.VersionDocUrlMap[version], val)
		} else {
			plugin.DocUrl = fmt.Sprintf(schema.VersionDocUrlMap[version], plugin.Name)
		}
		if plugin.Type == constant.APISIXTypeTAPISIX {
			plugin.DocUrl = config.G.Biz.TAPISIXPluginDocURLs[plugin.Name]
		}
		if plugin.Type == constant.APISIXTypeBKAPISIX {
			plugin.DocUrl = config.G.Biz.BKPluginDocURLs[plugin.Name]
		}
		if _, ok := pluginTypeMap[plugin.Type]; !ok {
			pluginTypeMap[plugin.Type] = []*schema.Plugin{plugin}
		} else {
			pluginTypeMap[plugin.Type] = append(pluginTypeMap[plugin.Type], plugin)
		}
	}
	var pluginTypeList serializer.PluginListResponse
	for pluginType, pluginList := range pluginTypeMap {
		pluginTypeList = append(pluginTypeList, &serializer.TypePluginInfo{
			Plugins: pluginList,
			Type:    pluginType,
		})
	}
	// plugins 排序
	serializer.SortPlugins(pluginTypeList)
	ginx.SuccessJSONResponse(c, pluginTypeList)
}
