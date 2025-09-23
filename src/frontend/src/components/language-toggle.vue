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

<template>
  <bk-popover
    :arrow="false"
    :padding="0"
    disable-outside-click
    placement="bottom"
    theme="light"
  >
    <div class="toggle-language-icon">
      <span
        :class="locale === 'en' ? 'icon-ag-toggle-english' : 'icon-ag-toggle-chinese'"
        class="icon apigateway-icon f22"
      ></span>
    </div>
    <template #content>
      <div
        class="language-item"
        @click="toggleLanguage('zh-cn')"
      >
        <span class="icon apigateway-icon icon-ag-toggle-chinese"></span>
        <span>中文</span>
      </div>
      <div
        class="language-item"
        @click="toggleLanguage('en')"
      >
        <span class="icon apigateway-icon icon-ag-toggle-english"></span>
        <span>English</span>
      </div>
    </template>
  </bk-popover>
</template>

<script lang="ts" setup>
import jsCookie from 'js-cookie';
import { jsonpRequest } from '@/common/util';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

interface IJsonpResponse {
  code: number;
  message: string;
  data: unknown;
  request_id: string;
  result: boolean;
}

const { locale } = useI18n();
const router = useRouter();

const toggleLanguage = async (targetLanguage: 'en' | 'zh-cn') => {
  if (targetLanguage === locale.value) {
    return;
  }

  const res = await jsonpRequest(
    `${window.BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fe_update_user_language/`,
    {
      language: targetLanguage,
    },
    'languageToggle',
  ) as IJsonpResponse;

  console.log(res);

  // if (res.code === 0) {
  jsCookie.set('blueking_language', targetLanguage, {
    domain: window.BK_DOMAIN,
    path: '/',
  });
  router.go(0);
  // }
};
</script>

<style lang="scss" scoped>

.toggle-language-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  cursor: pointer;
  border-radius: 50%;

  .icon {
    font-size: 16px;
    vertical-align: middle;
  }
}

.toggle-language-icon:hover {
  color: #ffffff;
  background-color: #303d55;
}

.language-item {
  font-size: 12px;
  display: block;
  margin-top: 5px;
  padding: 4px 12px;
  cursor: pointer;
  color: #63656e;

  &:hover {
    background-color: #f3f6f9;
  }

  &:nth-of-type(1) {
    margin-top: 0;
  }

  .icon {
    font-size: 18px;
    vertical-align: bottom;
  }
}

</style>
