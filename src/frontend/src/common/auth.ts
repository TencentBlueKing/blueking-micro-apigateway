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

import { showLoginModal } from '@blueking/login-modal';

interface ILoginData {
  loginUrl?: string
}

// 获取登录地址
const getLoginUrl = (url: string, cUrl: string, isFromLogout: boolean) => {
  const loginUrl = new URL(url);
  if (isFromLogout) {
    loginUrl.searchParams.append('is_from_logout', '1');
  }
  loginUrl.searchParams.append('c_url', cUrl);
  return loginUrl.href;
};

// 跳转到登录页
export const login = (data: ILoginData = {}) => {
  location.href = data.loginUrl || getLoginUrl(window.BK_LOGIN_URL, location.origin, false);
};

// 打开登录弹框
export const loginModal = () => {
  const loginUrl = getLoginUrl(
    `${window.BK_LOGIN_URL}/plain`,
    `${location.origin + window.SITE_URL}/static/login_success.html`,
    false,
  );
  showLoginModal({ loginUrl });
};

// 退出登录
export const logout = () => {
  location.href = getLoginUrl(window.BK_LOGIN_URL, location.origin, true);
};
