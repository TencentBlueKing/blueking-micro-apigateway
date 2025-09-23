![img](./docs/resource/img/bk_micro_apigateway_zh.png)
---

[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://github.com/TencentBlueKing/blueking-micro-apigateway/blob/main/LICENSE.txt) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/TencentBlueKing/blueking-micro-apigateway/pulls)

简体中文 | [English](README_EN.md)

## 概览

蓝鲸微网关（BK Micro APIGateway）分为控制面和数据面，控制面负责 API 的配置、发布、对比、同步等功能，数据面负责 API 的流量转发、安全防护等功能。其中数据面`bk-apisix`是基于 [Apache APISIX](https://github.com/apache/apisix) 定制，提供专业的网关解决方案， 支持官方插件 + 蓝鲸插件（部分）。

本项目是 `蓝鲸微网关 - 控制面`。

**蓝鲸微网关核心服务开源项目**

- 蓝鲸微网关 - [控制面](https://github.com/TencentBlueKing/blueking-micro-apigateway)
    - apiserver: 微网关控制面后端
    - frontend: 微网关控制面前端

## 功能特性

- 支持纳管 `apisix`、 `bk-apisix` 等不同版本的数据面的
- 支持只读模式纳管
- 支持 11 类 APISIX 资源 (route / service / upstream / consumer / consumer_group / plugin_config / global_rule / plugin_metadata / protobuf / ssl / stream_route ) 完整的生命周期管理
- 提供完整的`OpenAPI`接口，支持网关的注册及资源的注册发布。
- 支持自定义插件注册及配置。


## 快速开始

- [本地开发部署指引](docs/DEVELOP_GUIDE.md)

## 支持

- [蓝鲸 API 网关产品白皮书](https://bk.tencent.com/docs/document/7.0/171/13974)
- [蓝鲸智云 - 学习社区](https://bk.tencent.com/s-mart/community)
- [蓝鲸 DevOps 在线视频教程](https://bk.tencent.com/s-mart/video)
- 加入技术交流 QQ 群：

![img](./docs/resource/img/bk_qq_group.png)

## 蓝鲸社区
- [BK-APIGATEWAY](https://github.com/TencentBlueKing/blueking-apigateway)：蓝鲸 API 网关（API Gateway），是一种高性能、高可用的 API 托管服务。
- [BK-CI](https://github.com/TencentBlueKing/bk-ci)：蓝鲸持续集成平台是一个开源的持续集成和持续交付系统，可以轻松将你的研发流程呈现到你面前。
- [BK-BCS](https://github.com/TencentBlueKing/bk-bcs)：蓝鲸容器管理平台是以容器技术为基础，为微服务业务提供编排管理的基础服务平台。
- [BK-SOPS](https://github.com/TencentBlueKing/bk-sops)：标准运维（SOPS）是通过可视化的图形界面进行任务流程编排和执行的系统，是蓝鲸体系中一款轻量级的调度编排类
  SaaS 产品。
- [BK-CMDB](https://github.com/TencentBlueKing/bk-cmdb)：蓝鲸配置平台是一个面向资产及应用的企业级配置管理平台。
- [BK-JOB](https://github.com/TencentBlueKing/bk-job)：蓝鲸作业平台（Job）是一套运维脚本管理系统，具备海量任务并发处理能力。

## 贡献

如果你有好的意见或建议，欢迎给我们提 Issues 或 PullRequests，为蓝鲸开源社区贡献力量。关于分支 / Issue 及 PR,
请查看 [CONTRIBUTING](docs/CONTRIBUTING.md)。

[腾讯开源激励计划](https://opensource.tencent.com/contribution) 鼓励开发者的参与和贡献，期待你的加入。

## 合作方

<a href="https://apisix.apache.org/" target="_blank"><img src="https://github.com/apache/apisix/blob/master/logos/apisix-white-bg.jpg" alt="APISIX logo" height="150px" /></a>

## 证书

基于 MIT 协议，详细请参考 [LICENSE](LICENSE.txt)

