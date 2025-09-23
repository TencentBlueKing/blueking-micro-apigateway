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

import {
  IBaseResource,
  IBaseResourceConfig,
  IHealthCheck,
  IKeepalivePool,
  INode,
  INodeHash,
  ITimeout,
  ITls,
} from '@/types/common';

export interface IUpstreamConfig extends IBaseResourceConfig {
  scheme?: string;
  key?: string;
  timeout?: ITimeout;
  discovery_type?: string;
  pass_host?: string;
  upstream_host?: string;
  service_name?: string;
  type?: string;
  name?: string;
  tls?: ITls;
  retry_timeout?: number;
  desc?: string;
  checks?: IHealthCheck;
  retries?: number;
  nodes?: INode[] | INodeHash;
  keepalive_pool?: IKeepalivePool;
  hash_on?: string;
}

export interface IUpstream extends IBaseResource {
  config: IUpstreamConfig;
  name?: string;
  ssl_id?: string;
}
