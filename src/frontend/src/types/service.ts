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

import { IBaseResource, IBaseResourceConfig } from '@/types/common';
import { IUpstreamConfig } from '@/types/upstream';

export interface IServiceConfig extends IBaseResourceConfig {
  name?: string;
  desc?: string;
  hosts?: string[];
  enable_websocket?: boolean;
  upstream?: IUpstreamConfig;
  upstream_id?: string | number;
  plugins?: Record<string, any>;
}

export interface IService extends IBaseResource {
  name: string;
  config: IServiceConfig;
  upstream_id?: string;
}
