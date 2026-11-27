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

import { uniq } from 'lodash-es';
import type { IRouteConfig } from '@/types/route';

export type RouteMatchGroupKey = 'uris' | 'hosts' | 'remote_addrs';
export type RouteMatchSingleKey = 'uri' | 'host' | 'remote_addr';

export interface IRouteMatchConflict {
  key: RouteMatchGroupKey;
  singleKey: RouteMatchSingleKey;
  arrayKey: RouteMatchGroupKey;
  singleValue: string;
  arrayValues: string[];
}

export interface INormalizeRouteMatchResult {
  config: Partial<IRouteConfig>;
  conflicts: Partial<Record<RouteMatchGroupKey, IRouteMatchConflict>>;
}

export interface IRouteMatchSubmitValues {
  uris?: string[];
  hosts?: string[];
  remoteAddrs?: string[];
}

interface IMatchGroupSpec {
  singleKey: RouteMatchSingleKey;
  arrayKey: RouteMatchGroupKey;
  defaultEmpty: string[];
}

const MATCH_GROUP_SPECS: IMatchGroupSpec[] = [
  { singleKey: 'uri', arrayKey: 'uris', defaultEmpty: [''] },
  { singleKey: 'host', arrayKey: 'hosts', defaultEmpty: [] },
  { singleKey: 'remote_addr', arrayKey: 'remote_addrs', defaultEmpty: [] },
];

const sanitizeStringArray = (values: unknown): string[] => {
  if (!Array.isArray(values)) {
    return [];
  }
  return uniq(values.filter((item): item is string => typeof item === 'string' && item.length > 0));
};

/**
 * 在路由表单入口将单值匹配字段（uri / host / remote_addr）归一化为数组字段，
 * 并清理单值键以避免提交时与数组字段并存。
 * 若原始配置同时包含单值与数组且语义不一致，不静默合并，返回冲突信息交由页面显式确认。
 */
export const normalizeRouteMatchFields = (rawConfig: Partial<IRouteConfig> = {}): INormalizeRouteMatchResult => {
  const normalized: Partial<IRouteConfig> = { ...rawConfig };
  const conflicts: Partial<Record<RouteMatchGroupKey, IRouteMatchConflict>> = {};

  MATCH_GROUP_SPECS.forEach(({ singleKey, arrayKey, defaultEmpty }) => {
    const rawSingle = normalized[singleKey];
    const hasSingle = typeof rawSingle === 'string' && rawSingle.length > 0;
    const arrayValues = sanitizeStringArray(normalized[arrayKey]);

    if (hasSingle && arrayValues.length > 0) {
      if (arrayValues.length === 1 && arrayValues[0] === rawSingle) {
        normalized[arrayKey] = [rawSingle];
      } else {
        // 默认先回显数组字段，并记录冲突供用户显式选择，避免静默合并扩大匹配范围
        normalized[arrayKey] = arrayValues;
        conflicts[arrayKey] = {
          key: arrayKey,
          singleKey,
          arrayKey,
          singleValue: rawSingle,
          arrayValues,
        };
      }
    } else if (hasSingle) {
      normalized[arrayKey] = [rawSingle];
    } else if (arrayValues.length > 0) {
      normalized[arrayKey] = arrayValues;
    } else {
      normalized[arrayKey] = [...defaultEmpty];
    }

    delete normalized[singleKey];
  });

  return {
    config: normalized,
    conflicts,
  };
};

/**
 * 组装路由提交配置中的匹配条件字段：
 * 始终移除旧单值字段，非空时写入数组字段，清空可选字段时彻底省略。
 */
export const applyRouteMatchFieldsForSubmit = <T extends Partial<IRouteConfig>>(
  config: T,
  values: IRouteMatchSubmitValues,
): T => {
  const nextConfig = { ...config };
  delete nextConfig.uri;
  delete nextConfig.host;
  delete nextConfig.remote_addr;

  const uris = sanitizeStringArray(values.uris);
  if (uris.length > 0) {
    nextConfig.uris = uris;
  } else {
    delete nextConfig.uris;
  }

  const hosts = sanitizeStringArray(values.hosts);
  if (hosts.length > 0) {
    nextConfig.hosts = hosts;
  } else {
    delete nextConfig.hosts;
  }

  const remoteAddrs = sanitizeStringArray(values.remoteAddrs);
  if (remoteAddrs.length > 0) {
    nextConfig.remote_addrs = remoteAddrs;
  } else {
    delete nextConfig.remote_addrs;
  }

  return nextConfig;
};
