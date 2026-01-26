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

import { IUpstreamConfig } from '@/types/upstream';
import { IHealthCheck } from '@/types/common';

export const createDefaultHealthCheck = (): IHealthCheck => ({
  active: {
    type: 'http',
    timeout: 1,
    concurrency: 10,
    // port: 1,
    host: '',
    http_path: '/',
    req_headers: [],
    healthy: {
      successes: 2,
      interval: 1,
      http_statuses: [
        200,
        302,
      ],
    },
    unhealthy: {
      timeouts: 3,
      interval: 1,
      http_failures: 5,
      tcp_failures: 2,
      http_statuses: [
        429,
        404,
        500,
        501,
        502,
        503,
        504,
        505,
      ],
    },
  },
  passive: {
    type: 'http',
    healthy: {
      successes: 5,
      http_statuses: [
        200,
        201,
        202,
        203,
        204,
        205,
        206,
        207,
        208,
        226,
        300,
        301,
        302,
        303,
        304,
        305,
        306,
        307,
        308,
      ],
    },
    unhealthy: {
      timeouts: 7,
      http_failures: 2,
      tcp_failures: 2,
      http_statuses: [
        429,
        500,
        503,
      ],
    },
  },
});

export const useUpstreamForm = () => {
  const createDefaultUpstream = (override: Partial<IUpstreamConfig> = {}): Partial<IUpstreamConfig> => ({
    scheme: 'http',
    hash_on: '',
    // nodes 与 service_name 二选一
    nodes: [
      {
        host: '',
        port: 80,
        weight: 1,
      },
    ],
    // 服务 与 nodes 二选一
    // service_name: '',
    // 与 service_name 配合使用
    // discovery_type: 'dns',
    key: '',
    timeout: {
      send: 6,
      connect: 6,
      read: 6,
    },
    pass_host: 'pass',
    // upstream_host: '',
    type: 'roundrobin',
    tls: {},
    // retry_timeout: 120,
    // retries: 3,
    keepalive_pool: {
      idle_timeout: 60,
      requests: 1000,
      size: 320,
    },
    checks: {},
    labels: {},
    ...override,
  });

  return { createDefaultUpstream, createDefaultHealthCheck };
};
