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

interface IPublishBody {
  'resource_id_list': string[]
  'resource_type': string
}

const common = useCommon();

const { BK_DASHBOARD_URL } = window;

export const publish = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data: IPublishBody
}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/publish/`, data);

export const publishAll = ({
  gatewayId,
}: {
  gatewayId?: number
} = {}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/publish/all/`, {});

export interface IDiffGroup {
  added_count: number;
  change_detail: IChangeDetail[];
  deleted_count: number;
  modified_count: number;
  resource_type: string;
}

export interface IChangeDetail {
  after_status: string;
  before_status: string;
  name: string;
  operation_type: string;
  resource_id: string;
  updated_at: number;
  resource_type?: string;
}

export const getDiffAll = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data?: any,
} = {}): Promise<IDiffGroup[]> => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/unify_op/resources/-/diff/`, data);

export const getDiffByType = ({
  gatewayId,
  type,
  data,
}: {
  gatewayId?: number
  type: string
  data: Pick<IPublishBody, 'resource_id_list'>
}): Promise<IDiffGroup[]> => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/unify_op/resources/${type}/diff/`, data);

export interface IResourceDiffResponse {
  etcd_config: Record<string, any>
  editor_config: Record<string, any>
}

export const getResourceDiff = ({
  gatewayId,
  type,
  id,
}: {
  gatewayId?: number
  type: string
  id: string
}): Promise<IResourceDiffResponse> => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/unify_op/resources/${type}/diff/${id}/`, {});


