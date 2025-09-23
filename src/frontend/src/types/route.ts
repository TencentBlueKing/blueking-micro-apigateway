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

export interface IRouteConfig extends IBaseResourceConfig {
  desc?: string;
  enable_websocket?: boolean;
  filter_func?: string;
  host?: string;
  hosts?: string[];
  methods?: string[];
  plugins?: Record<string, any>;
  plugin_config_id?: string;
  priority?: number;
  remote_addr?: string;
  remote_addrs?: string[];
  script?: string;
  script_id?: string | number;
  status?: string;
  timeout?: ITimeout;
  upstream?: IUpstreamConfig;
  upstream_id?: string;
  uri?: string;
  uris?: string[];
  vars?: string[][];
}

export interface IRoute extends IBaseResource {
  name: string;
  config: IRouteConfig;
  route_id?: string;
  service_id?: string;
  upstream_id?: string;
  plugin_config_id?: string;
  gateway_id?: number;
}
