# apiserver

## 常用命令

```shell
make doc # 生成 Swagger API 文档。此命令依赖于 swaggo 工具。
make test # 执行测试
make lint # 执行代码检查
make fmt # 格式化代码: golines gofumpt goimports-reviser
make vet # 执行代码静态分析
make build # 构建项目
make docker-build # 构建 Docker 镜像
```

## 项目结构

```shell
.
├── ChangeLog.md          # 记录项目变更历史的文档
├── Dockerfile            # 用于构建 Docker 镜像的配置文件
├── Makefile              # 包含项目构建和管理任务的 Makefile
├── README.md             # 项目的 README 文件，通常包含项目介绍和使用说明
├── cmd                   # 存放命令行工具的源代码
│   ├── gen.go            # 生成相关代码或配置的命令
│   ├── init.go           # 初始化项目的命令
│   ├── migrate.go        # 数据库迁移的命令
│   ├── root.go           # 命令行工具的根命令
│   ├── scheduler.go      # 调度器相关命令
│   ├── version.go        # 显示项目版本的命令
│   └── webserver.go      # 启动 Web 服务器的命令
├── configs               # 存放配置文件
│   └── config.yaml       # 主配置文件，通常为 YAML 格式
├── docs                  # 存放项目文档
│   ├── docs.go           # 生成文档的 Go 代码
│   ├── swagger.json      # Swagger API 文档的 JSON 格式
│   └── swagger.yaml      # Swagger API 文档的 YAML 格式
├── go.mod                # Go 模块文件，定义项目依赖
├── go.sum                # Go 模块依赖校验文件
├── license_header.txt    # 代码文件头部的许可证声明
├── main.go               # 项目的主入口文件
└── pkg                   # 存放项目包（模块）的目录
    ├── account           # user 相关
    │   ├── account.go
    │   ├── bk_ticket.go
    │   ├── bk_token.go
    │   ├── mocks.go
    │   └── types.go
    ├── apis              # API 相关代码
    │   ├── basic         # 基础 API 实现
    │   └── web           # Web API 实现
    ├── biz               # 业务逻辑相关代码
    │   ├── common.go     # 通用业务逻辑
    │   ├── consumer.go
    ├── config            # 配置管理相关代码
    │   ├── config.go     # 配置加载和管理
    │   ├── loader.go     # 配置加载器
    │   └── types.go      # 定义配置相关数据类型
    ├── constant          # 定义项目常量
    │   ├── apisix.go     # APISIX 相关常量
    │   ├── enum.go       # 枚举类型常量
    │   └── system.go     # 系统相关常量
    ├── entity            # 实体层代码，定义数据模型
    │   ├── apisix        # APISIX 实体
    │   ├── base          # 基础实体
    │   ├── dto           # 数据传输对象
    │   └── model         # 数据模型
    ├── infras            # 基础设施层代码
    │   ├── database      # 数据库相关代码
    │   ├── leaderelection # 领导者选举相关代码
    │   ├── logging       # 日志记录相关代码
    │   ├── sentry        # Sentry 错误跟踪集成
    │   ├── storage       # 存储相关代码
    │   └── trace         # 跟踪和监控相关代码
    ├── middleware        # 中间件相关代码
    │   ├── access_control.go # 访问控制中间件
    ├── publisher         # 发布者相关代码
    │   ├── etcd.go       # ETCD 发布者实现
    │   └── type.go       # 发布者类型定义
    ├── repo              # 代码生成的仓库层代码
    │   ├── consumer.gen.go # 生成的消费者仓库代码
    ├── router            # 路由相关代码
    │   └── router.go     # 路由定义和管理
    ├── status            # 状态管理相关代码
    │   ├── status.go     # 状态管理逻辑
    │   └── status_test.go # 状态管理测试
    ├── utils             # 工具类代码
    │   ├── envx          # 环境变量相关工具
    └── version           # 版本相关代码
        └── version.go    # 版本信息定义
```

> 调用链路： handler-> biz -> repo -> infra(database)
