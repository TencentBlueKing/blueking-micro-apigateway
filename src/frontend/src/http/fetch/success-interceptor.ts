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

interface IErrorInfo {
  status?: Number;
  error?: any
}

// 请求成功执行拦截器
export default async (response: any, config: IFetchConfig) => {
  // 处理 http 非 200 异常
  if (!RegExp(/20/).test(String(response.status))) {
    const text = await response.text();
    const errorInfo: IErrorInfo = {
      status: response.status,
      error: {
        message: text,
      },
    };
    return Promise.reject(errorInfo);
  }

  // 处理204没有返回值的情况
  if (response.status === 204) {
    return Promise.resolve(response);
  }

  let responseInfo: any = null;
  try {
    responseInfo = await response[config.responseType]();
  } catch {
    return Promise.resolve(responseInfo);
  }

  if (response.ok) {
    const reg = RegExp(/20/);
    // 包含20x 代表请求成功
    if (reg.test(response.status)) {
      return Promise.resolve(responseInfo?.data);
    }
    return Promise.reject(response);
  }
};
