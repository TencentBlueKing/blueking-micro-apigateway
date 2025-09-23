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
import { useCommon } from '@/store';
import { IPluginCustomDto } from '@/types/plugin-custom';

const common = useCommon();

const { BK_DASHBOARD_URL } = window;

export const getPluginCustomList = ({ gatewayId, query }: {
  gatewayId?: number
  query?: Record<string, any>
}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/schemas/?${json2Query(query)}`);

export const getPluginCustom = ({ gatewayId, id }: {
  gatewayId?: number
  id: string
}): Promise<IPluginCustomDto> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/schemas/${id}/`);

export const postPluginCustom = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data?: object
}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/schemas/`, data);

export const putPluginCustom = ({
  gatewayId,
  id,
  data,
}: {
  gatewayId?: number
  id: string
  data?: object
}) => fetch.put(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/schemas/${id}/`, data);

export const deletePluginCustom = ({ gatewayId, id }: {
  gatewayId?: number
  id: string
}) => fetch.delete(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/schemas/${id}/`);

export const getCustomPluginSchema = ({ gatewayId, name }: {
  gatewayId?: number
  name: string
}): Promise<Record<string, any>> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/schemas/plugins/${name}/`);

