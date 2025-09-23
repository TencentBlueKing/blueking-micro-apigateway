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
import { useCommon } from '@/store';
import { json2Query } from '@/common/util';
import { IStreamRoute, IStreamRouteConfig } from '@/types/stream-route';

const { BK_DASHBOARD_URL } = window;
const common = useCommon();

// 获取stream route 列表
export const getStreamRouteList = ({ gatewayId, query }: {
  gatewayId?: number
  query?: {
    id?: number,
    name?: string
    status?: string
  }
}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/stream_routes/?${json2Query(query)}`);

// 获取stream route详情
export const getStreamRoute = ({ gatewayId, id }: {
  gatewayId?: number
  id: string
}): Promise<IStreamRoute> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/stream_routes/${id}/`);

// 创建stream route
export const postStreamRoute = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data?: IStreamRouteConfig
}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/stream_routes/`, data);

// 更新stream route
export const putStreamRoute = ({
  gatewayId,
  id,
  data,
}: {
  gatewayId?: number
  id: string
  data?: Record<string, any>
}) => fetch.put(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/stream_routes/${id}/`, data);

// 删除stream route
export const deleteStreamRoute = ({ gatewayId, id }: {
  gatewayId?: number
  id: string
}) => fetch.delete(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/stream_routes/${id}/`);

// stream route 下拉列表接口
export const getStreamRouteDropdowns = ({ gatewayId }: {
  gatewayId?: number
} = {}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/stream_routes-dropdown/`);
