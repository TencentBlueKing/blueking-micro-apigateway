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

import dayjs from 'dayjs';
import i18n from '@/i18n/index';
import { Message } from 'bkui-vue';
import { useClipboard } from '@vueuse/core';

const { t } = i18n.global;

// 获取 cookie object
export function getCookies(strCookie = document.cookie) {
  if (!strCookie) {
    return {};
  }
  const arrCookie = strCookie.split('; ');// 分割
  const cookiesObj: Record<string, string> = {};
  arrCookie.forEach((cookieStr) => {
    const arr = cookieStr.split('=');
    const [key, value] = arr;
    if (key) {
      cookiesObj[key] = value;
    }
  });
  return cookiesObj;
}

// 是否为24h之前
export const is24HoursAgo = (dateString: string | number) => {
  // 将日期字符串转换为 Date 对象
  const date: any = typeof dateString === 'string' ? new Date(dateString) : new Date(dateString * 1000);

  // 获取当前时间
  const now: any = new Date();

  // 计算时间差，单位为毫秒
  const diff = now - date;

  // 将时间差转换为小时
  const hours = diff / (1000 * 60 * 60);

  // 判断时间差是否大于等于24小时
  return hours >= 24;
};

export const handleCopy = (data: any) => {
  if (!data) return Message({
    theme: 'warning',
    message: t('暂无可复制内容'),
  });

  const { copy, isSupported } = useClipboard({ source: data, legacy: true });

  if (isSupported.value) {
    copy(data);
    Message({
      theme: 'success',
      message: t('已复制'),
    });
  } else {
    Message({
      theme: 'warning',
      message: t('复制失败，未开启剪贴板权限'),
    });
  }
};

/**
 * 检查是不是 object 类型
 * @param item
 */
export function isObject(item: any) {
  return Object.prototype.toString.apply(item) === '[object Object]';
}


/**
 * 深度合并多个对象
 * @param objectArray 待合并列表
 * @returns 合并后的对象
 */
export function deepMerge(...objectArray: object[]) {
  return objectArray.reduce((acc: Record<string, any>, obj: Record<string, any>) => {
    Object.keys(obj || {})
      .forEach((key) => {
        const pVal = acc[key];
        const oVal = obj[key];

        if (isObject(pVal) && isObject(oVal)) {
          acc[key] = deepMerge(pVal, oVal);
        } else {
          acc[key] = oVal;
        }
      });

    return acc;
  }, {});
}

/**
 * 时间格式化
 * @param val 待格式化时间
 * @param format 格式
 * @returns 格式化后的时间
 */
export function timeFormatter(val: string, format = 'YYYY-MM-DD HH:mm:ss') {
  return val ? dayjs(val)
    .format(format) : '--';
}


/**
 * 对象转为 url query 字符串
 * @param {*} param 要转的参数
 * @param {string} key key
 *
 */
export function json2Query(param: any, key?: string) {
  const mappingOperator = '=';
  const separator = '&';
  let paramStr = '';
  if (
    param instanceof String
    || typeof param === 'string'
    || param instanceof Number
    || typeof param === 'number'
    || param instanceof Boolean
    || typeof param === 'boolean'
  ) {
    // @ts-ignore
    paramStr += separator + key + mappingOperator + encodeURIComponent(param);
  } else {
    if (param) {
      Object.keys(param)
        .forEach((p) => {
          const value = param[p];
          const k = key === null || key === '' || key === undefined
            ? p
            : key + (param instanceof Array ? `[${p}]` : `.${p}`);
          paramStr += separator + json2Query(value, k);
        });
    }
  }
  return paramStr.substring(1);
}

/**
 * jsonp请求
 * @param url
 * @param params
 * @param callbackName
 */

export function jsonpRequest(url: string, params: any, callbackName?: string) {
  return new Promise((resolve) => {
    const script = document.createElement('script');
    if (callbackName) {
      callbackName = callbackName + Math.floor((1 + Math.random()) * 0x10000)
        .toString(16)
        .substring(1);
    }
    Object.assign(params, callbackName ? { callback: callbackName } : {});
    const arr = Object.keys(params)
      .map(key => `${key}=${params[key]}`);
    script.src = `${url}?${arr.join('&')}`;
    document.body.appendChild(script);
    // @ts-ignore
    window[callbackName] = (data: any) => {
      resolve(data);
    };
  });
}

export function textLengthCut({ text, length = 20, parens = false }: {
  text: string,
  length?: number,
  parens?: boolean
}) {
  const result = text.length > length ? `${text.slice(0, length)}...` : text;
  return parens ? `(${result})` : result;
}
