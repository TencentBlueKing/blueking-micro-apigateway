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
import { ICreatePayload, IConnectTest, ICheckName } from '@/types';

const { BK_DASHBOARD_URL } = window;

// 获取网关列表
export const getGatewaysList = (data: object) => fetch.get(`${BK_DASHBOARD_URL}/gateways/?${json2Query(data)}`);

// 新建网关
export const createGateway = (data: ICreatePayload) => fetch.post(`${BK_DASHBOARD_URL}/gateways/`, data);

// 更新网关
export const updateGateways = (apigwId: number, data: ICreatePayload) => fetch.put(`${BK_DASHBOARD_URL}/gateways/${apigwId}/`, data);

// 获取网关详情
export const getGatewaysDetail = (gatewayId: number) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/`);

// etcd 连通性测试
export const etcdConnectTest = (data: IConnectTest) => fetch.post(`${BK_DASHBOARD_URL}/gateways/etcd/test_connection/`, data);

// 网关重名检测
export const checkName = (data: ICheckName) => fetch.post(`${BK_DASHBOARD_URL}/gateways/check_name/`, data);

// 网关删除
export const deleteGateway = (gatewayId: number) => fetch.delete(`${BK_DASHBOARD_URL}/gateways/${gatewayId}/`);
