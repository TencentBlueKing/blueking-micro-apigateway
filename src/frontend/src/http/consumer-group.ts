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
import { IConsumerGroup } from '@/types/consumer-group';

const common = useCommon();

const { BK_DASHBOARD_URL } = window;

export const getConsumerGroups = ({ gatewayId, query }: {
  gatewayId?: number
  query?: Record<string, any>
} = {}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/consumer_groups/?${json2Query(query)}`);

export const getConsumerGroup = ({ gatewayId, id }: {
  gatewayId?: number
  id: string
}): Promise<IConsumerGroup> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/consumer_groups/${id}/`);

export const postConsumerGroup = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data?: object
}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/consumer_groups/`, data);

export const putConsumerGroup = ({
  gatewayId,
  id,
  data,
}: {
  gatewayId?: number
  id: string
  data?: object
}) => fetch.put(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/consumer_groups/${id}/`, data);

export const deleteConsumerGroup = ({ gatewayId, id }: {
  gatewayId?: number
  id: string
}) => fetch.delete(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/consumer_groups/${id}/`);

export const getConsumerGroupDropdowns = ({ gatewayId }: {
  gatewayId?: number
} = {}): Promise<{
  auto_id: number
  desc: string
  id: string
  name: string
}[]> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/consumer_groups-dropdown/`);

