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
import { IBaseResource, IBaseResourceConfig, ITimeout } from '@/types/common';

export interface IStreamRouteConfig extends IBaseResourceConfig {
  desc?: string;
  protocol?: Record<string, any>;
  plugins?: Record<string, any>;
  timeout?: ITimeout;
  upstream?: IUpstreamConfig;
  sni?: string;
  plugin_config_id?: string;
  status?: string;
  upstream_id?: string;
  remote_addr?: string;
  server_addr?: string,
  remote_addrs?: string[];
  server_addrs?: string[];
  server_port?: number;
  service_id?: string;
}

export interface IStreamRoute extends IBaseResource {
  name: string;
  config: IStreamRouteConfig;
  service_id?: string;
  upstream_id?: string;
  plugin_config_id?: string;
  gateway_id?: number;
}
