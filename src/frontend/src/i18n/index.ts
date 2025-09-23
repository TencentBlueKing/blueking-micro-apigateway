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

import Cookie from 'js-cookie';
import locales from './locales';
import { createI18n } from 'vue-i18n';

interface ILANG_PKG {
  [propName: string]: string;
}

const en: ILANG_PKG = {};
const zh: ILANG_PKG = {};

Object.keys(locales)
  .forEach((key) => {
    en[key] = locales[key][0] || key;
    zh[key] = locales[key][1] || key;
  });

const localLanguage = Cookie.get('blueking_language') || 'zh-cn';

const i18n = createI18n({
  silentTranslationWarn: true,
  legacy: false,
  locale: localLanguage,
  fallbackLocale: 'zh-cn',
  messages: {
    en,
    'zh-cn': zh,
  },
  missingWarn: false,
});

export const isChinese = localLanguage === 'zh-cn';

export default i18n;
