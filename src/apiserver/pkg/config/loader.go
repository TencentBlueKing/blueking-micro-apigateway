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

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/envx"
)

var (
	pwd, _     = os.Getwd()
	exePath, _ = os.Executable()
	exeDir     = filepath.Dir(exePath)
	// BaseDir 项目根目录
	BaseDir = lo.Ternary(strings.Contains(exeDir, pwd), exeDir, pwd)
)

func loadConfigFromFile(cfgFile string) (*Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(cfgFile); err != nil {
		return nil, errors.Errorf("config file %s not found", cfgFile)
	}

	// 使用 viper 从 cfgFile 加载配置
	vp := viper.New()
	vp.SetConfigFile(cfgFile)
	if err := vp.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := vp.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// 从环境变量加载配置
func loadConfigFromEnv() (*Config, error) {
	// 服务配置
	serviceCfg, err := loadServiceConfigFromEnv()
	if err != nil {
		return nil, err
	}

	// Mysql 配置
	mysqlCfg, err := loadMysqlConfigFromEnv()
	if err != nil {
		return nil, err
	}

	// 业务配置
	bizCfg, err := loadBizConfigFromEnv()
	if err != nil {
		return nil, err
	}

	bkPlatUrl := loadBkPlatUrlFromEnv()

	crypto, err := loadCryptoFromEnv()
	if err != nil {
		return nil, err
	}
	return &Config{
		Edition:         envx.Get("EDITION", "ee"),
		Service:         serviceCfg,
		Biz:             bizCfg,
		Tracing:         loadTraceFromEnv(),
		Sentry:          loadSentryFromEnv(),
		MysqlConfig:     mysqlCfg,
		BkPlatUrlConfig: bkPlatUrl,
		Crypto:          crypto,
	}, nil
}

// 判断字符串非空
func notEmpty(str string) bool {
	return str != ""
}

// 从环境变量读取 Mysql 增强服务配置
func loadMysqlConfigFromEnv() (*MysqlConfig, error) {
	host := envx.Get("MYSQL_HOST", "")
	port := envx.Get("MYSQL_PORT", "")
	name := envx.Get("MYSQL_NAME", "")
	user := envx.Get("MYSQL_USER", "")
	passwd := envx.Get("MYSQL_PASSWORD", "")
	charset := envx.Get("MYSQL_CHARSET", "utf8")

	if ok := lo.EveryBy([]string{host, port, name, user, passwd}, notEmpty); !ok {
		return nil, nil
	}
	mysqlPort, err := cast.ToIntE(port)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid GCS_MYSQL_PORT: %s", port)
	}

	return &MysqlConfig{
		Host:     host,
		Port:     mysqlPort,
		Name:     name,
		User:     user,
		Password: passwd,
		Charset:  charset,
	}, nil
}

// 从环境变量读取服务配置
func loadServiceConfigFromEnv() (ServiceConfig, error) {
	// 是否为本地开发环境
	isLocalDev := envx.Get("BKPAAS_ENVIRONMENT", "dev") == "dev"

	allowedUsers := []string{}
	if val := envx.Get("ALLOWED_USERS", ""); val != "" {
		// 允许访问的用户在环境变量中格式如 "admin,userAlpha,userBeta"
		allowedUsers = strings.Split(val, ",")
	}
	// 默认允许任意源访问
	allowedOrigins := []string{"*"}
	if val := envx.Get("ALLOWED_ORIGINS", ""); val != "" {
		// 允许访问的源在环境变量中格式如 "http://localhost:8080,http://localhost:8081"
		allowedOrigins = strings.Split(val, ",")
	}
	return ServiceConfig{
		Server: ServerConfig{
			Port:         cast.ToInt(envx.Get("PORT", "8080")),
			GraceTimeout: cast.ToInt(envx.Get("GRACE_TIMEOUT", "30")),
			GinRunMode: envx.Get(
				"GIN_RUN_MODE",
				lo.Ternary[string](isLocalDev, gin.DebugMode, gin.ReleaseMode),
			),
		},
		Log: LogConfig{
			Level: envx.Get(
				"LOG_LEVEL",
				lo.Ternary(isLocalDev, "debug", "info"),
			),
			Dir: envx.Get(
				"LOG_BASE_DIR",
				lo.Ternary(isLocalDev, BaseDir+"/logs/", "/app/v3logs/"),
			),
			ForceToStdout: true,

			SentryReportLevel: envx.Get(
				"SENTRY_LOG_LEVEL",
				lo.Ternary(isLocalDev, "debug", "error"),
			),
		},
		AllowedOrigins: allowedOrigins,
		AllowedUsers:   allowedUsers,
		HealthzToken:   envx.Get("HEALTHZ_TOKEN", ""),
		MetricToken:    envx.Get("METRIC_TOKEN", "metric_token"),
		EnableSwagger:  cast.ToBool(envx.Get("ENABLE_SWAGGER", lo.Ternary(isLocalDev, "true", "false"))),
		DocFileBaseDir: envx.Get(
			"DOC_FILE_BASE_DIR",
			lo.Ternary(isLocalDev, BaseDir+"/docs/", "/app/docs/"),
		),
		AppCode:          envx.Get("BK_APP_CODE", ""),
		AppSecret:        envx.Get("BK_APP_SECRET", ""),
		UserTokenKey:     envx.Get("BK_USER_TOKEN_KEY", "bk_token"),
		CSRFCookieDomain: envx.Get("CSRF_COOKIE_DOMAIN", ""),
		SessionCookieAge: envx.GetDuration("SESSION_COOKIE_AGE", "24h"),
		Standalone:       envx.GetBoolean("STANDALONE", false),
		DemoMode:         envx.GetBoolean("DEMO_MODE", false),
		DemoModeWarnMsg:  envx.Get("DEMO_MODE_WARN_MSG", "demo 模式下不允许进行该操作"),
	}, nil
}

// 从环境变量读取蓝鲸平台服务地址
func loadBkPlatUrlFromEnv() BkPlatUrlConfig {
	return BkPlatUrlConfig{
		BkLogin: strings.TrimRight(envx.Get("BKPAAS_LOGIN_URL", "http://bklogin.example.com"), "/"),
	}
}

// 从环境变量读取 sentry 地址 TODO：部署文档补充相关配置/说明
func loadSentryFromEnv() Sentry {
	return Sentry{
		DSN: envx.Get("SENTRY_DSN", ""),
	}
}

// 从环境变量读取 trace 地址 TODO：部署文档补充相关配置/说明
func loadTraceFromEnv() Tracing {
	return Tracing{
		Enable:       envx.GetBoolean("TRACING_ENABLE", false),
		Endpoint:     envx.Get("TRACING_ENDPOINT", ""),
		Type:         envx.Get("TRACING_TYPE", "http"),
		Token:        envx.Get("TRACING_TOKEN", ""),
		Sampler:      envx.Get("TRACING_SAMPLER", ""),
		SamplerRatio: envx.GetFloat64("TRACING_SAMPLER_RATIO", 0.1),
		ServiceName:  envx.Get("TRACING_SERVICE_NAME", "blueking-micro-apigateway"),
		Instrument: Instrument{
			GinAPI: envx.GetBoolean("TRACING_INSTRUMENT_GIN_API", false),
			DbAPI:  envx.GetBoolean("TRACING_INSTRUMENT_DB_API", false),
		},
	}
}

// 加载业务相关配置
func loadBizConfigFromEnv() (BizConfig, error) {
	// TAPISIX 插件文档地址列表
	tapisixPluginDocUrls := envx.Get("TAPISIX_PLUGIN_DOC_URLS", "{}")
	tapisixPluginMap := make(map[string]string)
	err := json.Unmarshal([]byte(tapisixPluginDocUrls), &tapisixPluginMap)
	if err != nil {
		return BizConfig{}, errors.Wrap(err, "failed to unmarshal TAPISIX_PLUGIN_DOC_URLS")
	}

	// 蓝鲸插件文档地址列表
	bkPluginDocUrls := envx.Get("BK_PLUGIN_DOC_URLS", "{}")
	bkPluginMap := make(map[string]string)
	err = json.Unmarshal([]byte(bkPluginDocUrls), &bkPluginMap)
	if err != nil {
		return BizConfig{}, errors.Wrap(err, "failed to unmarshal BK_PLUGIN_DOC_URLS")
	}

	openapiTokenWhitelist := envx.Get("OPEN_API_TOKEN_WHITELIST", "")
	tokenList := strings.Split(openapiTokenWhitelist, ";")
	tokenMap := make(map[string]bool)
	for _, token := range tokenList {
		if token == "" {
			continue
		}
		tokenMap[token] = true
	}
	demoProtectResources := envx.Get("DEMO_PROTECT_RESOURCES", "")
	demoProtectResourcesList := strings.Split(demoProtectResources, ";")
	demoProtectResourceMap := make(map[string]bool)
	for _, r := range demoProtectResourcesList {
		if r == "" {
			continue
		}
		demoProtectResourceMap[r] = true
	}
	return BizConfig{
		SyncInterval:          envx.GetDuration("SYNC_INTERVAL", "1h"),
		TAPISIXPluginDocURLs:  tapisixPluginMap,
		BKPluginDocURLs:       bkPluginMap,
		OpenApiTokenWhitelist: tokenMap,
		DemoProtectResources:  demoProtectResourceMap,
		Links: LinkConfig{
			BKFeedBackLink:   envx.Get("BK_FEED_BACK_LINK", ""),
			BKGuideLink:      envx.Get("BK_GUIDE_LINK", ""),
			BKApigatewayLink: envx.Get("BK_APIGATEWAY_LINK", ""),
		},
	}, nil
}

// 加载 co
func loadCryptoFromEnv() (Crypto, error) {
	return Crypto{
		Nonce: envx.Get("CRYPTO_NONCE", ""),
		Key:   envx.Get("CRYPTO_KEY", ""),
	}, nil
}
