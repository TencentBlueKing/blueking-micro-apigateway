# 更多配置字段说明：pkg/config/types.go

# 版本，可选项：ee、te
edition: ee

# 蓝鲸平台相关配置
platform:
  # 蓝鲸应用 ID
  appID: apiserver
  # 蓝鲸应用密钥
  appSecret: <masked>
  # 应用模块名称
  moduleName: default
  # 运行环境
  runEnv: stag
  # 蓝鲸应用版本
  region: default
  # 数据库加密方式
  cryptoType: CLASSIC
  # 蓝鲸域名
  bkDomain: example.com
  # API 地址模板
  apiUrlTmpl: http://{api_name}.apigw.example.com
  # 蓝鲸平台访问地址
  bkPlatUrl:
    bkPaaS: http://bkpaas.example.com
    bkLogin: http://bklogin.example.com
    bkCompApi: http://bkapi.example.com
  # 增强服务
  addons:
    # MySQL 数据库服务
    mysql:
      host: localhost
      port: 3306
      name: bk-micro-apigateway
      user: root
      password: <masked>
      charset: utf8mb4
    # RabbitMQ 消息队列服务
    rabbitMQ:
      host: localhost
      port: 5672
      user: bk-micro-apigateway
      vhost: bk-micro-apigateway
      password: <masked>
    # Redis 服务
    redis:
      username: ""
      host: localhost
      port: 6379
      password: <masked>
    # 蓝鲸制品库（对象存储）
    bkRepo:
      endpointUrl: http://localhost
      project: bksaas-addons
      username: gin-demo
      password: <masked>
      bucket: gin-demo
      publicBucket: gin-demo-public
      privateBucket: gin-demo-private
    # 蓝鲸 APM（监控提供的 OpenTelemetry）
    bkAPM:
      otelTrace: true
      otelGrpcUrl: http://localhost:4317
      otelBkDataToken: <masked>
      otelSample: always_on
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
  # 默认允许其他来源访问
  allowedOrigins: ["*"]
  # 默认允许所有用户访问
  allowedUsers: []
  # 健康检查 API Token
  healthzToken: <masked>
  # 指标 API Token
  metricToken: <masked>
  # 是否启用 Swagger 服务
  enableSwagger: false
  # 文档，静态文件，模板的基础目录
  docFileBaseDir: docs
  staticFileBaseDir: static
  appCode: "demo"
  appSecret: "123"
  userTokenKey: "bk_token"
  csrfCookieDomain: ""
  sessionCookieAge: 24h
  standalone: false
  demoMode: false
  tmplFileBaseDir: templates
# 业务相关配置
biz:
  links:
    bkGuideLink: http://example.com/guide
    bkFeedBackLink: http://example.com/feedback
    bkApigatewayLink: http://apigw.example.com

mysqlconfig:
  host: localhost
  port: 3306
  name: bk-micro-apigateway
  user: root
  password: <masked>
  charset: utf8mb4

crypto:
  nonce: k2dbCGetyusW
  key: jxi18GX5w2qgHwfZCFpn07q8FScXJOd3

