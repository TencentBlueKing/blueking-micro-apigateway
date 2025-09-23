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

// @ts-check
import { defineStore } from 'pinia';
import QueryString from 'qs';
import { shallowRef } from 'vue';
import { unionBy } from 'lodash-es';

const { BK_LIST_USERS_API_URL } = window;

export const useStaffStore = defineStore({
  id: 'staffStore',
  state: () => ({
    fetching: false,
    list: shallowRef([]),
  }),
  actions: {
    async fetchStaffs(name?: string) {
      if (this.fetching) return;
      this.fetching = true;
      const usersListPath = `${BK_LIST_USERS_API_URL}`;
      const params: any = {
        app_code: 'bk-magicbox',
        page: 1,
        page_size: 200,
        callback: 'callbackStaff',
      };
      if (name) {
        params.fuzzy_lookups = name;
      }
      const scriptTag = document.createElement('script');
      scriptTag.setAttribute('type', 'text/javascript');
      scriptTag.setAttribute('src', `${usersListPath}?${QueryString.stringify(params)}`);

      const headTag = document.getElementsByTagName('head')[0];
      // @ts-ignore
      window[params.callback] = ({ data, result }: { data: any, result: boolean }) => {
        if (result) {
          this.fetching = false;
          this.list = unionBy(this.list, data.results, 'id');
        }
        headTag.removeChild(scriptTag);
        // @ts-ignore
        delete window[params.callback];
      };
      headTag.appendChild(scriptTag);
    },
  },
});
