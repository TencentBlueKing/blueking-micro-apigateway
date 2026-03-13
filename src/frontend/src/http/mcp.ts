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

const common = useCommon();

const { BK_DASHBOARD_URL } = window;

// 获取当前网关下所有 MCP Access Token 列表
export const getMcpTokens = ({ gatewayId }: {
  gatewayId?: number
}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/mcp/tokens/`);

// 创建
export const postMcpToken = ({
  gatewayId,
  data,
}: {
  gatewayId?: number
  data?: object
} = {}) => fetch.post(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/mcp/tokens/`, data);

// 详情
export const getMcpTokensDetails = ({ gatewayId, id  }: {
  gatewayId?: number
  id: number
}) => fetch.get(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/mcp/tokens/${id}/`);

// 删除
export const deleteMcpToken = ({ gatewayId, id }: {
  gatewayId?: number
  id: number
}) => fetch.delete(`${BK_DASHBOARD_URL}/gateways/${gatewayId || common.gatewayId}/mcp/tokens/${id}/`);
