# 更多配置字段说明：pkg/config/types.go

# 版本，可选项：ee、te
edition: ee

# 服务相关配置
service:
  # Gin Web 服务
  server:
    port: 8080
    graceTimeout: 30
    ginRunMode: debug
  # 日志配置
  log:
    # 日志级别，可选项：debug、info、warn、error
    level: info
    dir: v3logs
    forceToStdout: true
    sentryReportLevel: error
  # 默认允许其他来源访问
  allowedOrigins: ["*"]
  # 默认允许所有用户访问
  allowedUsers: []
  # 健康检查 API Token
  healthzToken: ""
  # 指标 API Token
  metricToken: ""
  # 是否启用 Swagger 服务
  enableSwagger: false
  # 文档文件存放目录
  docFileBaseDir: docs
  appCode: "demo"
  appSecret: "123"
  userTokenKey: "bk_token"
  csrfCookieDomain: ""
  sessionCookieAge: 24h
  standalone: false
  demoMode: false
  demoModeWarnMsg: "demo 模式下不允许进行该操作"

# Sentry 错误追踪配置
sentry:
  dsn: ""

# 链路追踪配置
tracing:
  enable: false
  endpoint: ""
  type: http
  token: ""
  sampler: ""
  samplerRatio: 0.1
  serviceName: blueking-micro-apigateway
  instrument:
    ginAPI: false
    dbAPI: false

# 蓝鲸平台访问地址
bkPlatUrlConfig:
  bkLogin: http://bklogin.example.com

# 业务相关配置
biz:
  syncInterval: 1h
  tapisixPluginDocUrls: {}
  bkPluginDocUrls: {}
  openApiTokenWhitelist: {}
  demoProtectResources: {}
  links:
    bkGuideLink: http://example.com/guide
    bkFeedBackLink: http://example.com/feedback
    bkApigatewayLink: http://apigw.example.com

# MySQL 数据库配置
mysqlconfig:
  host: localhost
  port: 3306
  name: bk-micro-apigateway
  user: root
  password: ""
  charset: utf8mb4

# 加密配置
crypto:
  nonce: ""
  key: ""
