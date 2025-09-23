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

import { IRoute } from '@/types/route';
import { IService } from '@/types/service';
import { IConsumerGroup } from '@/types/consumer-group';
import { IPluginConfigDto } from '@/types/plugin-config';
import { IPluginMetadataDto } from '@/types/plugin-metadata';
import { IGlobalRules } from '@/types/global-rules';
import { IConsumer } from '@/types/consumer';
import { IUpstream } from '@/types/upstream';
import { Ref } from 'vue';

export type ResourceType =
  IRoute
  | IService
  | IUpstream
  | IConsumer
  | IConsumerGroup
  | IPluginConfigDto
  | IPluginMetadataDto
  | IGlobalRules;

// 基础资源对象
export interface IBaseResource {
  auto_id?: number;
  id?: string;
  updated_at?: number;
  created_at?: number;
  update_time?: number;
  create_time?: number;
  create_by?: string;
  update_by?: string;
  status?: string;
}

// 基础资源 config
export interface IBaseResourceConfig {
  id?: string;
  labels?: { [key: string]: string };
}

export interface IHTTPMethod {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'HEAD' | 'OPTIONS' | 'CONNECT' | 'TRACE' | 'PURGE';
}

export interface IDialog {
  isShow: boolean
  title?: string
  loading?: boolean
}

export enum IStaffType {
  RTX = 'rtx',
}

export interface IStaff {
  english_name: string;
  chinese_name: string;
  username: string;
  display_name: string;
}

export interface ITimeout {
  send?: number;
  connect?: number;
  read?: number;
}

export interface ITls {
  client_key?: string;
  client_cert_id?: string;
  verify?: boolean;
  client_cert?: string;
}

export interface IHealthCheck {
  passive?: IPassiveHealthCheck;
  active?: IActiveHealthCheck;
}

export interface IPassiveHealthCheck {
  healthy?: {
    successes?: number;
    http_statuses?: number[];
  };
  unhealthy?: {
    tcp_failures?: number;
    timeouts?: number;
    http_failures?: number;
    http_statuses?: number[];
  };
  type?: string;
}

export interface IActiveHealthCheck {
  concurrency?: number;
  healthy?: {
    successes?: number;
    interval?: number;
    http_statuses?: number[];
  };
  https_verify_certificate?: boolean;
  host?: string;
  timeout?: number;
  unhealthy?: {
    tcp_failures?: number;
    timeouts?: number;
    interval?: number;
    http_statuses?: number[];
    http_failures?: number;
  };
  req_headers?: string[];
  port?: number;
  type?: string;
  http_path?: string;
}

export interface INode {
  priority?: number;
  weight?: number;
  port?: number;
  host?: string;
  metadata?: any;
}

export interface INodeHash {
  [key: string]: number;
}

export interface IKeepalivePool {
  requests?: number;
  size?: number;
  idle_timeout?: number;
}

export interface ILabels {
  [key: string]: string
}


export type ITableMethod = {
  loading: boolean
  TDesignTableRef: Ref,
  setPagination: () => void
  getPagination: () => void
  resetPaginationTheme: () => void
  setPaginationTheme: () => void
  fetchData: () => void
  refresh: () => void
};

export type ITableSettings  = {
  columns: string[]
  fontSize: string
  rowSize: string
};

export type ReturnRecordType<T, U> = Record<string, (arg?: T) => U>;

export interface IDropList {
  value: string
  label: string
  theme?: string
  disabled?: boolean
  method?: () => void,
}
