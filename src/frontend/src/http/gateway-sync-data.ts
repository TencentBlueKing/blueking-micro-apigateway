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

const common = useCommon();

const { BK_DASHBOARD_URL } = window;

// 获取同步数据列表
export const getGatewaySyncDataList = ({ gatewayId, query }: {
  gatewayId?: number
  query?: Record<string, any>
}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/synced/items/?${json2Query(query)}`);

// 一键同步
export const postGatewaySyncData = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data?: object
}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/sync/`, data);

// 同步资源添加到编辑区
export const addResourceToEditArea = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data?: object
}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/unify_op/resources/-/managed/`, data);

// 获取最新同步时间
export const getSyncLastTime = ({ gatewayId }: {
  gatewayId?: number
}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/synced/last_time/`);

// 导出 etcd 资源
export const exportEtcdResources = ({ gatewayId }: {
  gatewayId?: number
} = {}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/unify_op/etcd/export/`);

// 导入 etcd 资源
export const importEtcdResources = ({ gatewayId, resources }: {
  gatewayId?: number
  resources: {
    add?: Record<string, any[]>
    update?: Record<string, any[]>
  }
}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/unify_op/resources/import/`, resources);
