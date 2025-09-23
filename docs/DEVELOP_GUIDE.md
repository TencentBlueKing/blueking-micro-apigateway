# 本地开发环境搭建

本地开发时，你可以根据需要，为实际开发的模块（apiserver / frontend）准备开发环境。

在开始开发前，你需要为整个项目安装并初始化 `pre-commit`，

```bash
# 假设你当前在项目的根目录下

❯ pre-commit install
```

## apiserver

`apiserver` 基于 gin 框架开发，为网关产品提供后端接口。本地开发环境搭建请参考 [本地开发文档](../src/apiserver/README.md)

## frontend

`dashboard-front` 为基于 Vue.js 的前端项目。本地开发环境搭建请参考 [本地开发文档](../src/frontend/README.md)
