# 本地开发环境搭建

本地开发时，你可以根据需要，为实际开发的模块（apiserver / frontend）准备开发环境。

在开始开发前，你需要为整个项目安装并初始化 `pre-commit`，

```bash
# 假设你当前在项目的根目录下

❯ pre-commit install
```

## apiserver

`apiserver` 基于 gin 框架开发，为网关产品提供后端接口。本地开发环境搭建请参考 [本地开发文档](../src/apiserver/README.md)

配置校验重构相关的后端验证建议直接在 `src/apiserver/` 下执行：

```bash
GOTOOLCHAIN=auto go test ./pkg/apis/web/serializer ./pkg/apis/open/serializer ./pkg/middleware ./pkg/apis/common ./pkg/biz ./pkg/entity/model ./pkg/publisher ./pkg/resourcecodec
GOTOOLCHAIN=auto go test ./...
```

原因：当前 `go.mod` 版本要求高于本机默认工具链，且这组命令可以同时覆盖 WebAPI、OpenAPI、导入校验、模型持久化投影、发布装配与最终 ETCD JSON Schema 校验。

## frontend

`dashboard-front` 为基于 Vue.js 的前端项目。本地开发环境搭建请参考 [本地开发文档](../src/frontend/README.md)
