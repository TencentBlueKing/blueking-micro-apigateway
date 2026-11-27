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

import { cloneDeep, uniq } from 'lodash-es';
import type { ISSLConfig } from '@/types/ssl';

export interface ISSLSubmitParams {
  cert: string;
  key: string;
  snis?: string[];
  fallbackSnis?: string[];
}

const sanitizeStringArray = (values: unknown): string[] => {
  if (!Array.isArray(values)) {
    return [];
  }
  return uniq(values.filter((item): item is string => typeof item === 'string' && item.length > 0));
};

/**
 * 从 SSL 配置中提取用于表单或列表回显的 SNI 列表，兼容数组 snis 与单值 sni。
 */
export const extractSSLSnis = (config?: Partial<ISSLConfig> | null): string[] => {
  const snis = sanitizeStringArray(config?.snis);
  if (snis.length > 0) {
    return snis;
  }
  if (typeof config?.sni === 'string' && config.sni.length > 0) {
    return [config.sni];
  }
  return [];
};

/**
 * 基于原始 SSL 配置构建提交 config：
 * 保留表单未暴露的附加字段（如 certs、keys、client、ssl_protocols、labels 等），
 * 覆写最新主证书/私钥与 snis，并清理互斥单值 sni 与需由后端重算的证书派生字段。
 */
export const buildSSLSubmitConfig = (
  rawConfig: Partial<ISSLConfig> | null | undefined,
  params: ISSLSubmitParams,
): Partial<ISSLConfig> => {
  const nextConfig: Partial<ISSLConfig> = cloneDeep(rawConfig ?? {});

  // 清理单值 sni，避免与 snis 并存触发 APISIX SSL schema oneOf 校验失败
  delete nextConfig.sni;
  // 清理需由后端 ParseCert 根据当前证书与名称重新生成的字段，防止 MergeJson 反向覆盖新值
  delete nextConfig.validity_start;
  delete nextConfig.validity_end;
  delete nextConfig.name;

  nextConfig.cert = params.cert;
  nextConfig.key = params.key;

  const userSnis = sanitizeStringArray(params.snis);
  const fallbackSnis = sanitizeStringArray(params.fallbackSnis);
  const finalSnis = userSnis.length > 0 ? userSnis : fallbackSnis;

  if (finalSnis.length > 0) {
    nextConfig.snis = finalSnis;
  } else {
    delete nextConfig.snis;
  }

  return nextConfig;
};
