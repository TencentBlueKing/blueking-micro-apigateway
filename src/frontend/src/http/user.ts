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

const { BK_DASHBOARD_URL } = window;

interface IVersionLogItem {
  content: string;
  data: string;
  version: string
}

export const getUser = () => fetch.get(`${BK_DASHBOARD_URL}/accounts/userinfo/`);

export const getFeatureFlags = (data: Record<string, any>) => fetch.get(`${BK_DASHBOARD_URL}/settings/feature_flags/?${json2Query(data)}`);

export const getVersionLog = (): Promise<IVersionLogItem[]> => fetch.get(`${BK_DASHBOARD_URL}/version-log/`);

export const getEnvVars = (): Promise<{
  links: {
    bk_feed_back_link: string,
    bk_guide_link: string
  }
}> => fetch.get(`${BK_DASHBOARD_URL}/env-vars/`);

