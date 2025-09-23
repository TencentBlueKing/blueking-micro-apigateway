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

import fetch from './fetch';
import { json2Query } from '@/common/util';
import { IPluginGroup } from '@/types/plugin';

const { BK_DASHBOARD_URL } = window;

export const getPlugins = ({ gatewayId, query }: {
  gatewayId?: number
  query?: { all?: boolean }
}): Promise<IPluginGroup[]> => {
  const queryParams = query && Object.keys(query)?.length > 0 ? `?${json2Query(query)}` : '';
  return fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/plugins/${queryParams}`);
};

export const getPluginSchema = ({ gatewayId, name, query }: {
  gatewayId?: number
  name: string
  query?: { schema_type: string }
}): Promise<Record<string, any>> => {
  const queryParams = query && Object.keys(query)?.length > 0 ? `?${json2Query(query)}` : '';
  return fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/schemas/plugins/${name}/${queryParams}`);
};

// 获取能给 consumer 和 consumer group 配置的插件列表
export const getConsumerPlugins = ({ gatewayId }: {
  gatewayId?: number
}): Promise<IPluginGroup[]> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/plugins/?kind=consumer`);

// 获取能配置 plugin metadata 的插件列表
export const getMetadataPlugins = ({ gatewayId }: {
  gatewayId?: number
}): Promise<IPluginGroup[]> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/plugins/?kind=metadata`);

// 获取能配置 stream route 的插件列表
export const getStreamRoutePlugins = ({ gatewayId }: {
  gatewayId?: number
}): Promise<IPluginGroup[]> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/plugins/?kind=stream`);

export const getConsumerPluginSchema = ({ gatewayId, name }: {
  gatewayId?: number
  name: string
}): Promise<Record<string, any>> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/schemas/plugins/${name}/?schema_type=consumer`);

export const getMetadataPluginSchema = ({ gatewayId, name }: {
  gatewayId?: number
  name: string
}): Promise<Record<string, any>> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/schemas/plugins/${name}/?schema_type=metadata`);

