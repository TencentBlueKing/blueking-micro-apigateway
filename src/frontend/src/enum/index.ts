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

import i18n from '@/i18n';

export const HTTP_METHODS_MAP = {
  GET: 'GET',
  POST: 'POST',
  PUT: 'PUT',
  DELETE: 'DELETE',
  PATCH: 'PATCH',
  HEAD: 'HEAD',
  OPTIONS: 'OPTIONS',
  CONNECT: 'CONNECT',
  TRACE: 'TRACE',
  PURGE: 'PURGE',
};

export const METHOD_THEMES = {
  POST: 'info',
  GET: 'success',
  DELETE: 'danger',
  PUT: 'warning',
  PATCH: 'info',
  ANY: 'success',
};

export const RESOURCE_CN_MAP: Record<string, string> = {
  consumer: i18n.global.t('消费者'),
  consumer_group: i18n.global.t('消费者组'),
  global_rule: i18n.global.t('全局规则'),
  plugin_config: i18n.global.t('插件组'),
  plugin_metadata: i18n.global.t('插件元数据'),
  route: i18n.global.t('路由'),
  service: i18n.global.t('服务'),
  upstream: i18n.global.t('上游'),
  gateway: i18n.global.t('网关'),
  stream_route: i18n.global.t('stream 路由'),
  proto: 'proto',
  ssl: 'ssl',
  schema: i18n.global.t('自定义插件'),
};

export const RESOURCE_ICON_MAP: Record<string, string> = {
  consumer: 'cc-user',
  consumer_group: 'usergroup',
  global_rule: 'system-mgr',
  plugin_config: 'plugin-generic-fill',
  plugin_metadata: 'micro-plugin',
  route: 'luyou',
  service: 'fuwu-2',
  upstream: 'micro-upstream',
};

export const STATUS_CN_MAP: Record<string, string> = {
  delete_draft: i18n.global.t('删除待发布'),
  create_draft: i18n.global.t('新增待发布'),
  update_draft: i18n.global.t('更新待发布'),
  conflict: i18n.global.t('冲突'),
  success: i18n.global.t('已发布'),
};

export const PLUGIN_TYPE_CN_MAP: Record<string, string> = {
  authentication: i18n.global.t('身份验证'),
  'bk-apisix': i18n.global.t('蓝鲸插件'),
  general: i18n.global.t('通用'),
  transformation: i18n.global.t('转换'),
  observability: i18n.global.t('可观测性'),
  security: i18n.global.t('安全防护'),
  serverless: i18n.global.t('无服务器架构'),
  traffic: i18n.global.t('流量控制'),
  tapisix: i18n.global.t('TAPISIX 插件'),
  other: i18n.global.t('其他'),
  'other protocols': i18n.global.t('其他'),
  'customize plugin': i18n.global.t('自定义插件'),
};

export const OPERATION_TYPE_CN_MAP: Record<string, string> = {
  create: i18n.global.t('新增'),
  update: i18n.global.t('更新'),
  delete: i18n.global.t('删除'),
};

// 键名需和 router 中的路由 name 字段一致
export const RESOURCE_INTRODUCTION: Record<string, { text: string, docLink: string }> = {
  route: {
    text: '路由 Route，是请求的入口点，它定义了客户端请求与服务之间的匹配规则，根据匹配结果加载并执行相应的插件，最后将请求转发给到指定的上游服务。路由中主要包含三部分内容：匹配规则，插件配置和上游信息。路由的上游信息可以直接配置在路由中，或者通过绑定 Service 或 Upstream 配置。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/route/',
  },
  service: {
    text: '服务 Service， 一组 Route 的抽象，可以将路由中公共的插件配置、上游目标信息抽象成一个服务；多个路由绑定同一个服务，这个服务可以对应一组上游节点。从而减少路由中的冗余配置。（Route N <-> 1 Service <-> 1 Upstream）',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/service/',
  },
  upstream: {
    text: '上游 Upstream，即后端服务，是对虚拟主机抽象，即应用层服务或节点的抽象； 可以对上游服务的多个目标节点进行负载均衡和健康检查以及重试。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/upstream/',
  },
  consumer: {
    text: '消费者 Consumer，是某类服务的消费方，需要与用户认证配合才可以使用；形式可能是最终用户，开发者或者 API 调用方等。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/consumer/',
  },
  'consumer-group': {
    text: '消费者组 Consumer Groups，可以在同一个消费者组中启用任意数量的插件，并在一个或者多个消费者中引用该消费者组。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/consumer-group/',
  },
  'global-rules': {
    text: '全局规则 Global Rules，可以在Global Rules 中启用任意数量的插件，插件对所有的请求生效。相对于 Route、Service、Plugin Config、Consumer 中的插件配置，Global Rules 中的插件总是优先执行。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/global-rule/',
  },
  'plugin-config': {
    text: '插件组 Plugin Config，一组通用插件配置的抽象，配置后可以在不同的路由中直接引用该插件组，从而实现复用同一组插件配置。对于同一个插件的配置，只能有一个是有效的，优先级为 Consumer > Route > Plugin Config > Service。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/plugin-metadata/',
  },
  'plugin-metadata': {
    text: '插件元数据 Plugin Metadata， 配置通用的插件元数据属性，可以作用于包含该元数据插件的所有路由及服务中。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/terminology/plugin-metadata/',
  },
  'plugin-custom': {
    text: '自定义插件在 config.yaml 启用并且挂载后可以在这里进行自定义插件的 schema 上传，上传后即可在资源配置的插件列表看到该插件进行配置。',
    docLink: '',
  },
  'gateway-sync-data': {
    text: '数据来源是 etcd，同步过来的数据会作为一份独立快照数据，可以同当前编辑区资源对比，etcd 中由其他控制面新增同步过来的资源可以加入到编辑区进行维护',
    docLink: '',
  },
  publish: {
    text: '将编辑区资源下发到 etcd 当中并且发布完之后会同步到 etcd 资源列表中',
    docLink: '',
  },
  proto: {
    text: 'Protocol Buffers 是 Google 用于序列化结构化数据的框架，它具有语言中立、平台中立、可扩展机制的特性，您只需定义一次数据的结构化方式，然后就可以使用各种语言通过特殊生成的源代码轻松地将结构化数据写入和读取各种数据流。Protocol Buffers 列表包含了已创建的 proto 文件，在启用 grpc-transcode 插件时可配置 ID 读取对应的 proto 文件内容。',
    docLink: '',
  },
  ssl: {
    text: '证书被网关用于处理加密请求，它将与 SNI 关联，并与路由中主机名绑定。',
    docLink: '',
  },
  'stream-route': {
    text: 'Stream Route 在传输层运行，处理基于 TCP 和 UDP 协议的流式流量。TCP 用于许多应用程序和服务，如 LDAP、MySQL 和 RTMP。UDP 用于许多流行的非事务性应用程序，如 DNS、syslog 和 RADIUS。',
    docLink: 'https://apisix.apache.org/zh/docs/apisix/stream-proxy/',
  },
};
