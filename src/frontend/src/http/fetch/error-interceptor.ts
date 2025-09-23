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

import type { IFetchConfig } from './index';
import { Message } from 'bkui-vue';
import { showLoginModal } from '@blueking/login-modal';
import { useCommon } from '@/store';

const { BK_LOGIN_URL } = window;

// 请求执行失败拦截器
export default (errorData: any, config: IFetchConfig) => {
  const {
    status,
    error,
  } = errorData;
  const loginCallbackURL = `${window.BK_DASHBOARD_FE_URL}/static/login_success.html?is_ajax=1`;
  const siteLoginUrl = BK_LOGIN_URL || '';
  const loginUrl = `${BK_LOGIN_URL}?app_code=1&c_url=${encodeURIComponent(loginCallbackURL)}`;
  let iframeWidth = 700;
  let iframeHeight = 500;
  switch (status) {
    // 参数错误
    case 400:
      break;
    // 用户登录状态失效
    case 401:
      // if (error?.data?.login_plain_url) {
      if (error?.data?.loginUrl) {
        const { width, height } = error.data;
        iframeWidth = width;
        iframeHeight = height;
      }
      if (!siteLoginUrl) {
        console.error('Login URL not configured!');
        return;
      }
      // 增加encodeURIComponent防止回调地址特殊字符被转义
      showLoginModal({ loginUrl, width: iframeWidth, height: iframeHeight });
      break;
  }
  // 全局捕获错误给出提示
  if (config.globalError) {
    if (error?.code !== 'Unauthorized' && !useCommon()?.noGlobalError) {
      Message({ theme: 'error', message: error?.message || 'Error' });
    }
  }
  console.log('=>(error-interceptor.ts:46) errorData', errorData);
  return Promise.reject(error);
};
