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

package config

import (
	"fmt"
	"net/url"
	"time"
)

// Config SaaS 配置
type Config struct {
	// 服务配置
	Service ServiceConfig
	// sentry 配置
	Sentry Sentry
	// tracing 配置
	Tracing Tracing
	// MYSQL 配置
	MysqlConfig *MysqlConfig
	// 业务配置
	Biz BizConfig
	// Crypto Crypto 配置
	Crypto Crypto
	// 蓝鲸各平台地址
	BkPlatUrlConfig BkPlatUrlConfig
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	// Web Server 配置
	Server ServerConfig
	// 日志配置
	Log LogConfig

	// CORS 允许来源列表
	AllowedOrigins []string
	// AllowedUsers 允许访问的用户列表（UserID）
	AllowedUsers []string
	// 健康探针 Token
	HealthzToken string
	// 指标 API Token
	MetricToken string

	// 是否启用 swagger docs
	EnableSwagger bool
	// 文档文件存放目录
	DocFileBaseDir string
	// appcode
	AppCode string
	// appsecret
	AppSecret string
	// user_token key
	UserTokenKey string
	// csrf cookie domain
	CSRFCookieDomain string
	// SESSION_COOKIE_AGE
	SessionCookieAge time.Duration
	// standalone true: 代表独立部署
	Standalone bool
	// 是否开启demo模式
	DemoMode bool
	// demo模式提醒报错信息
	DemoModeWarnMsg string
}

// ServerConfig Gin Web Server 配置
type ServerConfig struct {
	// 服务端口
	Port int
	// 优雅退出等待时间
	GraceTimeout int
	// Gin 运行模式
	GinRunMode string
}

// LogConfig 日志配置
type LogConfig struct {
	// 日志级别，可选值为：debug、info、warn、error
	Level string
	// 日志目录，部署于 PaaS 平台上时，该值必须为 /app/v3logs，否则无法采集日志
	Dir string
	// 是否强制标准输出，不输出到文件（一般用于本地开发，标准输出日志查看比较方便）
	ForceToStdout bool

	// sentry level
	SentryReportLevel string
}

// Sentry sentry 配置
type Sentry struct {
	DSN string
}

// Tracing is the config for trace
type Tracing struct {
	Enable       bool
	Endpoint     string
	Type         string
	Token        string
	Sampler      string
	SamplerRatio float64
	ServiceName  string
	Instrument   Instrument
}

// Instrument  is the config for trace
type Instrument struct {
	GinAPI bool
	DbAPI  bool
}

// GinAPIEnabled get gin api trace switch
func (t Tracing) GinAPIEnabled() bool {
	return t.Enable && t.Instrument.GinAPI
}

// DBAPIEnabled get db api trace switch
func (t Tracing) DBAPIEnabled() bool {
	return t.Enable && t.Instrument.DbAPI
}

// MysqlConfig Mysql 增强服务配置
type MysqlConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	Charset  string
}

// DSN ...
func (cfg *MysqlConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true",
		cfg.User,
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Charset,
	)
}

// BkPlatUrlConfig 蓝鲸各平台服务地址
type BkPlatUrlConfig struct {
	// 蓝鲸开发者中心地址
	BkPaaS string
	// 统一登录地址
	BkLogin string
	// 组件 API 地址
	BkCompApi string
	// TODO: SaaS 开发者可按需添加诸如 BkIAM，BkLog 等服务配置
}

// BizConfig 业务相关配置
type BizConfig struct {
	SyncInterval          time.Duration     // 定时同步间隔
	TAPISIXPluginDocURLs  map[string]string // TAPISIX 插件文档地址列表
	BKPluginDocURLs       map[string]string // 蓝鲸插件文档地址列表
	OpenApiTokenWhitelist map[string]bool   // OpenAPI 接口token白名单
	DemoProtectResources  map[string]bool   // demo模式保护资源列表
	Links                 LinkConfig        // 前端需要的链接相关配置
}

type LinkConfig struct {
	BKGuideLink    string // 产品使用指南地址
	BKFeedBackLink string // 产品反馈地址
}

// Crypto  配置
type Crypto struct {
	Nonce string
	Key   string
}
