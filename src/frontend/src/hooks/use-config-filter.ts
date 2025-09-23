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

import { cloneDeep, isEmpty, isEqual, isNil, isObject } from 'lodash-es';

interface IDefaultConfigMap {
  [key: string]: {
    default: any
    // 是否必填
    required?: boolean
    // 是否被折叠在高级配置里
    advanced?: boolean
  }
}

export const defaultResourceConfigMap: Record<string, IDefaultConfigMap> = {
  route: {
    desc: { default: '' },
    labels: { default: {} },
    methods: {
      default: [
        'GET',
        'POST',
      ],
    },
    enable_websocket: { default: false },
    priority: {
      default: 0,
      advanced: true,
    },
    hosts: {
      default: [] as string[],
      advanced: true,
    },
    remote_addrs: {
      default: [] as string[],
      advanced: true,
    },
    plugins: { default: {} },
  },

  upstream: {
    name: { default: '' },
    desc: { default: '' },
    labels: { default: {} },
    scheme: { default: 'http' },
    type: { default: 'roundrobin' },
    hash_on: { default: '' },
    key: { default: '' },
    pass_host: { default: 'pass' },
    upstream_host: { default: '' },
    tls: { default: {} },
    checks: { default: {} },
    retries: {
      default: '',
      advanced: true,
    },
    retry_timeout: {
      default: '',
      advanced: true,
    },
    timeout: {
      default: {
        send: 6,
        connect: 6,
        read: 6,
      },
      advanced: true,
    },
    keepalive_pool: {
      default: {
        idle_timeout: 60,
        requests: 1000,
        size: 320,
      },
      advanced: true,
    },
  },

  service: {
    name: { default: '' },
    desc: { default: '' },
    labels: { default: {} },
    enable_websocket: { default: false, advanced: true },
    hosts: { default: [], advanced: true },
  },

  stream_route: {
    name: { default: '' },
    desc: { default: '' },
    labels: { default: {} },
    remote_addrs: {
      default: [] as string[],
      advanced: true,
    },
    server_addrs: {
      default: [] as string[],
      advanced: true,
    },
    server_port: {
      default: 0,
      advanced: true,
    },
    sni: {
      default: '',
      advanced: true,
    },
  },
};

export default function useConfigFilter() {
  const filterOptionalDefaultKeys = (config: Record<string, any>, resourceType: string) => {
    const _config = cloneDeep(config);
    const defaultConfigMap = defaultResourceConfigMap[resourceType];

    Object.keys(_config)
      .forEach((key) => {
        if (defaultConfigMap[key] && isEqual(_config[key], defaultConfigMap[key].default)) {
          delete _config[key];
        }
      });

    return _config;
  };

  const filterEmpty = <T = Record<any, any>>(config: Record<any, any>, exclude: string[] = []) => {
    const _config = cloneDeep(config);
    Object.keys(_config)
      .forEach((key) => {
        if (
          (isObject(_config[key]) && isEmpty(_config[key]))
          || isNil(_config[key])
          || _config[key] === ''
        ) {
          if (exclude.includes(key)) {
            return;
          }
          delete _config[key];
        }
      });

    return _config as T;
  };

  // 过滤高级配置里没改动过的配置
  const filterAdvanced = <T = Record<any, any>>(config: Record<any, any>, resourceType: string) => {
    const _config = cloneDeep(config);
    const defaultConfigMap = defaultResourceConfigMap[resourceType];

    Object.keys(_config)
      .forEach((key) => {
        if (key in defaultConfigMap && defaultConfigMap[key].advanced) {
          if (isEqual(defaultConfigMap[key].default, _config[key])) {
            delete _config[key];
          }
        }
      });

    return _config as T;
  };

  return {
    filterOptionalDefaultKeys,
    filterEmpty,
    filterAdvanced,
  };
}
