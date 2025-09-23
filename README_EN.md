![img](./docs/resource/img/bk_micro_apigateway_en.png)
---

[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://github.com/TencentBlueKing/blueking-micro-apigateway/blob/main/LICENSE.txt) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/TencentBlueKing/blueking-micro-apigateway/pulls)

Simplified Chinese | [English](README_EN.md)

## Overview

BlueKing Micro API Gateway (BK Micro APIGateway) consists of a control plane and a data plane. The control plane handles API configuration, publishing, comparison, synchronization, etc., while the data plane manages API traffic forwarding and security protection. The data plane `bk-apisix` is customized based on [Apache APISIX](https://github.com/apache/apisix), providing a professional gateway solution that supports official plugins + BlueKing plugins (partial).

This project is the `BlueKing Micro Gateway - Control Plane`.

**Core Open-Source Projects of BlueKing Micro API Gateway**

- BlueKing Micro Gateway - [Control Plane](https://github.com/TencentBlueKing/blueking-micro-apigateway)
    - apiserver: Backend of the micro-gateway control plane
    - frontend: Frontend of the micro-gateway control plane

## Features

- Supports managing data planes of different versions such as `apisix` and `bk-apisix`
- Supports read-only mode management
- Supports full lifecycle management for 11 types of APISIX resources (route / service / upstream / consumer / consumer_group / plugin_config / global_rule / plugin_metadata / protobuf / ssl / stream_route)
- Provides complete `OpenAPI` interfaces for gateway registration and resource publishing
- Supports custom plugin registration and configuration

## Quick Start

- [Local Development and Deployment Guide](docs/DEVELOP_GUIDE.md)

## Support

- [BlueKing API Gateway Product White Paper](https://bk.tencent.com/docs/document/7.0/171/13974)
- [BlueKing - Learning Community](https://bk.tencent.com/s-mart/community)
- [BlueKing DevOps Online Video Tutorials](https://bk.tencent.com/s-mart/video)
- Join the technical QQ group:

![img](./docs/resource/img/bk_qq_group.png)

## BlueKing Community
- [BK-APIGATEWAY](https://github.com/TencentBlueKing/blueking-apigateway): BlueKing API Gateway (API Gateway), a high-performance and highly available API hosting service.
- [BK-CI](https://github.com/TencentBlueKing/bk-ci): BlueKing Continuous Integration Platform is an open-source CI/CD system that easily visualizes your R&D workflow.
- [BK-BCS](https://github.com/TencentBlueKing/bk-bcs): BlueKing Container Management Platform is a foundational service platform for microservice orchestration based on container technology.
- [BK-SOPS](https://github.com/TencentBlueKing/bk-sops): Standard Operations (SOPS) is a system for task orchestration and execution through a visual interface, serving as a lightweight scheduling SaaS product in the BlueKing ecosystem.
- [BK-CMDB](https://github.com/TencentBlueKing/bk-cmdb): BlueKing Configuration Platform is an enterprise-level configuration management platform for assets and applications.
- [BK-JOB](https://github.com/TencentBlueKing/bk-job): BlueKing Job Platform (Job) is an Ops script management system with massive task concurrency capabilities.

## Contribution

If you have suggestions or feedback, feel free to submit Issues or Pull Requests to contribute to the BlueKing open-source community. For branch/Issue/PR guidelines, see [CONTRIBUTING](docs/CONTRIBUTING.md).

The [Tencent Open Source Incentive Plan](https://opensource.tencent.com/contribution) encourages developer participation and contributions. We look forward to your involvement.

## Collaborators

<a href="https://apisix.apache.org/" target="_blank"><img src="https://github.com/apache/apisix/blob/master/logos/apisix-white-bg.jpg" alt="APISIX logo" height="150px" /></a>

## License

Licensed under the MIT License. For details, see [LICENSE](LICENSE.txt).