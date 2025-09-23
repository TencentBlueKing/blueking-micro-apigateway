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

import { defineStore } from 'pinia';
import { getPlugins } from '@/http/plugins';
import { IPlugin, IPluginGroup } from '@/types/plugin';
import { getEnums } from '@/http/enums';
import { cloneDeep } from 'lodash-es';

export const useCommon = defineStore('common', {
  state: () => ({
    // 网关id
    gatewayId: 0,
    // 网关name
    gatewayName: '',
    methodList: [
      {
        id: 'GET',
        name: 'GET',
      },
      {
        id: 'POST',
        name: 'POST',
      },
      {
        id: 'PUT',
        name: 'PUT',
      },
      {
        id: 'PATCH',
        name: 'PATCH',
      },
      {
        id: 'DELETE',
        name: 'DELETE',
      },

      {
        id: 'HEAD',
        name: 'HEAD',
      },
      {
        id: 'OPTIONS',
        name: 'OPTIONS',
      },
      {
        id: 'ANY',
        name: 'ANY',
      },
    ],
    curGatewayData: { allow_update_gateway_auth: false } as any,
    websiteConfig: {},
    noGlobalError: false, // 请求出错是否显示全局的错误Message
    // 插件列表
    plugins: [] as IPlugin[],
    pluginGroupList: [] as IPluginGroup[],
    // 系统枚举值
    enums: {} as Record<string, Record<string, string>>,
    appVersion: '',
  }),

  actions: {
    setGatewayId(gatewayId: number) {
      this.gatewayId = gatewayId;
    },
    setGatewayName(name: string) {
      this.gatewayName = name;
    },
    setCurGatewayData(data: any) {
      this.curGatewayData = data;
    },
    setWebsiteConfig(data: any) {
      this.websiteConfig = data;
    },
    setNoGlobalError(val: boolean) {
      this.noGlobalError = val;
    },

    async setPlugins({ gatewayId, query = { all: true } }: {
      gatewayId?: number,
      query?: { all?: boolean }
    } = {}) {
      const res = await getPlugins({ gatewayId: gatewayId || this.gatewayId, query });
      const groups = res || [];

      const otherGroupIndex = groups.findIndex(group => group.type === 'other' || group.type === 'other protocols');
      if (otherGroupIndex > -1) {
        const groupCopy = cloneDeep(groups[otherGroupIndex]);
        groups.splice(otherGroupIndex, 1);
        groups.push(groupCopy);
      }

      const tapisixGroupIndex = groups.findIndex(group => group.type === 'tapisix');
      if (tapisixGroupIndex > -1) {
        const groupCopy = cloneDeep(groups[tapisixGroupIndex]);
        groups.splice(tapisixGroupIndex, 1);
        groups.push(groupCopy);
      }

      const bkApisixGroupIndex = groups.findIndex(group => group.type === 'bk-apisix');
      if (bkApisixGroupIndex > -1) {
        const groupCopy = cloneDeep(groups[bkApisixGroupIndex]);
        groups.splice(bkApisixGroupIndex, 1);
        groups.push(groupCopy);
      }

      const customGroupIndex = groups.findIndex(group => group.type === 'customize plugin');
      if (customGroupIndex > -1) {
        const groupCopy = cloneDeep(groups[customGroupIndex]);
        groups.splice(customGroupIndex, 1);
        groups.push(groupCopy);
      }

      this.pluginGroupList = groups;
      this.plugins = groups.reduce((acc, group) => {
        return [...acc, ...group.plugins];
      }, []);
    },

    async setEnums() {
      this.enums = await getEnums();
    },

    setAppVersion(version: string) {
      this.appVersion = version || '1.0.0';
    },
  },
});
