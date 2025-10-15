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

// Package web ...
package web

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/account"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/handler"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/middleware"
)

// RegisterWebApi 注册 web 路由
func RegisterWebApi(path string, router *gin.RouterGroup) {
	group := router.Group(path)
	// middleware: session
	store := cookie.NewStore([]byte(config.G.Service.AppSecret))
	store.Options(sessions.Options{MaxAge: int(config.G.Service.SessionCookieAge.Seconds())})
	group.Use(sessions.Sessions(fmt.Sprintf("%s-session", config.G.Service.AppCode), store))

	//  csrf
	group.Use(middleware.CSRF(config.G.Service.AppCode, config.G.Service.AppSecret))
	group.Use(middleware.CSRFToken(config.G.Service.AppCode, config.G.Service.CSRFCookieDomain))

	// user auth
	authBackend := account.GetAuthBackend()
	group.Use(middleware.UserAuth(authBackend))
	group.Use(middleware.Permission())
	group.GET("/enums/", handler.Enum)
	group.GET("/accounts/userinfo/", handler.GetUserInfo)
	group.GET("/version-log/", handler.GetVersionLog)
	group.GET("/env-vars/", handler.EnvVars)

	// gateway
	group.POST("/gateways/", handler.GatewayCreate)
	group.GET("/gateways/", handler.GatewayList)
	group.POST("/gateways/check_name/", handler.GatewayCheckName)
	group.POST("/gateways/etcd/test_connection/", handler.EtcdTestConnection)

	// gateway:gateway_id
	gatewayGroup := group.Group("/gateways/:gateway_id")
	gatewayGroup.Use(middleware.GatewayAccess())
	gatewayGroup.Use(middleware.ResourceOperationCheck())

	gatewayGroup.PUT("/", handler.GatewayUpdate)

	gatewayGroup.GET("/", handler.GatewayGet)
	gatewayGroup.DELETE("/", handler.GatewayDelete)

	// labels
	gatewayGroup.GET("/labels/:type/", handler.GatewayLabelList)

	// route
	gatewayGroup.POST("/routes/", handler.RouteCreate)
	gatewayGroup.PUT("/routes/:id/", handler.RouteUpdate)
	gatewayGroup.GET("/routes/:id/", handler.RouteGet)
	gatewayGroup.DELETE("/routes/:id/", handler.RouteDelete)
	gatewayGroup.GET("/routes/", handler.RouteList)
	gatewayGroup.GET("/routes-dropdown/", handler.RouteDropDownList)

	// service
	gatewayGroup.POST("/services/", handler.ServiceCreate)
	gatewayGroup.PUT("/services/:id/", handler.ServiceUpdate)
	gatewayGroup.GET("/services/:id/", handler.ServiceGet)
	gatewayGroup.DELETE("/services/:id/", handler.ServiceDelete)
	gatewayGroup.GET("/services/", handler.ServiceList)
	gatewayGroup.GET("/services-dropdown/", handler.ServiceDropDownList)

	// upstream
	gatewayGroup.POST("/upstreams/", handler.UpstreamCreate)
	gatewayGroup.PUT("/upstreams/:id/", handler.UpstreamUpdate)
	gatewayGroup.GET("/upstreams/:id/", handler.UpstreamGet)
	gatewayGroup.DELETE("/upstreams/:id/", handler.UpstreamDelete)
	gatewayGroup.GET("/upstreams/", handler.UpstreamList)
	gatewayGroup.GET("/upstreams-dropdown/", handler.UpstreamDropDownList)

	// ssl
	gatewayGroup.POST("/ssls/", handler.SSLCreate)
	gatewayGroup.POST("/ssls/check/", handler.SSLCheck)
	gatewayGroup.PUT("/ssls/:id/", handler.SSLUpdate)
	gatewayGroup.GET("/ssls/:id/", handler.SSLGet)
	gatewayGroup.DELETE("/ssls/:id/", handler.SSLDelete)
	gatewayGroup.GET("/ssls/", handler.SSLList)
	gatewayGroup.GET("/ssls-dropdown/", handler.SSLDropDownList)

	// global_rule
	gatewayGroup.POST("/global_rules/", handler.GlobalRuleCreate)
	gatewayGroup.PUT("/global_rules/:id/", handler.GlobalRuleUpdate)
	gatewayGroup.GET("/global_rules/:id/", handler.GlobalRuleGet)
	gatewayGroup.DELETE("/global_rules/:id/", handler.GlobalRuleDelete)
	gatewayGroup.GET("/global_rules/", handler.GlobalRuleList)
	gatewayGroup.GET("/global_rules/-/plugins/", handler.GlobalRulePlugins)
	gatewayGroup.GET("/global_rules-dropdown/", handler.GlobalRuleDropDownList)

	// consumer
	gatewayGroup.POST("/consumers/", handler.ConsumerCreate)
	gatewayGroup.PUT("/consumers/:id/", handler.ConsumerUpdate)
	gatewayGroup.GET("/consumers/:id/", handler.ConsumerGet)
	gatewayGroup.DELETE("/consumers/:id/", handler.ConsumerDelete)
	gatewayGroup.GET("/consumers/", handler.ConsumerList)
	gatewayGroup.GET("/consumers-dropdown/", handler.ConsumerDropDownList)

	// consumer_group
	gatewayGroup.POST("/consumer_groups/", handler.ConsumerGroupCreate)
	gatewayGroup.PUT("/consumer_groups/:id/", handler.ConsumerGroupUpdate)
	gatewayGroup.GET("/consumer_groups/:id/", handler.ConsumerGroupGet)
	gatewayGroup.DELETE("/consumer_groups/:id/", handler.ConsumerGroupDelete)
	gatewayGroup.GET("/consumer_groups/", handler.ConsumerGroupList)
	gatewayGroup.GET("/consumer_groups-dropdown/", handler.ConsumerGroupDropDownList)

	// plugin_config
	gatewayGroup.POST("/plugin_configs/", handler.PluginConfigCreate)
	gatewayGroup.PUT("/plugin_configs/:id/", handler.PluginConfigUpdate)
	gatewayGroup.GET("/plugin_configs/:id/", handler.PluginConfigGet)
	gatewayGroup.DELETE("/plugin_configs/:id/", handler.PluginConfigDelete)
	gatewayGroup.GET("/plugin_configs/", handler.PluginConfigList)
	gatewayGroup.GET("/plugin_configs-dropdown/", handler.PluginConfigDropDownList)

	// plugin_metadata
	gatewayGroup.POST("/plugin_metadatas/", handler.PluginMetadataCreate)
	gatewayGroup.PUT("/plugin_metadatas/:id/", handler.PluginMetadataUpdate)
	gatewayGroup.GET("/plugin_metadatas/:id/", handler.PluginMetadataGet)
	gatewayGroup.DELETE("/plugin_metadatas/:id/", handler.PluginMetadataDelete)
	gatewayGroup.GET("/plugin_metadatas/", handler.PluginMetadataList)
	gatewayGroup.GET("/plugin_metadatas-dropdown/", handler.PluginMetadataDropDownList)

	// proto
	gatewayGroup.POST("/protos/", handler.ProtoCreate)
	gatewayGroup.PUT("/protos/:id/", handler.ProtoUpdate)
	gatewayGroup.GET("/protos/:id/", handler.ProtoGet)
	gatewayGroup.DELETE("/protos/:id/", handler.ProtoDelete)
	gatewayGroup.GET("/protos/", handler.ProtoList)
	gatewayGroup.GET("/protos-dropdown/", handler.ProtoDropDownList)

	// stream_route
	gatewayGroup.POST("/stream_routes/", handler.StreamRouteCreate)
	gatewayGroup.PUT("/stream_routes/:id/", handler.StreamRouteUpdate)
	gatewayGroup.GET("/stream_routes/:id/", handler.StreamRouteGet)
	gatewayGroup.DELETE("/stream_routes/:id/", handler.StreamRouteDelete)
	gatewayGroup.GET("/stream_routes/", handler.StreamRouteList)
	gatewayGroup.GET("/stream_routes-dropdown/", handler.StreamRouteDropDownList)

	// operation_audit_log
	gatewayGroup.GET("/audits/logs/", handler.OperationAuditLogList)

	// sync_data
	gatewayGroup.GET("/synced/items/", handler.SyncedItemList)
	gatewayGroup.GET("/synced/summary/", handler.SyncedItemSummary)
	gatewayGroup.GET("/synced/last_time/", handler.SyncedLastTime)

	// unify_op
	gatewayGroup.POST("/unify_op/resources/:type/revert/", handler.ResourceRevert)
	gatewayGroup.POST("/unify_op/resources/-/managed/", handler.SyncedResourceManaged)
	gatewayGroup.POST("/unify_op/resources/-/diff/", handler.ResourcesDiffAll)
	gatewayGroup.POST("/unify_op/resources/:type/diff/", handler.ResourcesDiff)
	gatewayGroup.GET("/unify_op/resources/:type/diff/:id/", handler.ResourceConfigDiffDetail)
	gatewayGroup.DELETE("/unify_op/resources/:type/", handler.ResourceDelete)
	gatewayGroup.GET("/unify_op/resources/labels/:type/", handler.ResourceLabelsList)
	gatewayGroup.GET("/unify_op/etcd/export/", handler.EtcdExport)
	gatewayGroup.POST("/unify_op/resources/upload/", handler.ResourceUpload)
	gatewayGroup.POST("/unify_op/resources/import/", handler.ResourceImport)

	// schema
	gatewayGroup.GET("/schemas/plugins/:name/", handler.PluginSchemaGet)
	gatewayGroup.GET("/schemas/resources/:type/", handler.ResourceSchemaGet)
	gatewayGroup.POST("/schemas/", handler.SchemaCreate)
	gatewayGroup.PUT("/schemas/:auto_id/", handler.SchemaUpdate)
	gatewayGroup.GET("/schemas/:auto_id/", handler.SchemaGet)
	gatewayGroup.DELETE("/schemas/:auto_id/", handler.SchemaDelete)
	gatewayGroup.GET("/schemas/", handler.SchemaList)
	gatewayGroup.GET("/plugins/", handler.PluginsGet)

	// publish
	gatewayGroup.POST("/publish/", handler.PublishResource)
	gatewayGroup.POST("/publish/all/", handler.PublishResourceAll)
	gatewayGroup.POST("/sync/", handler.ResourceSync)
}
