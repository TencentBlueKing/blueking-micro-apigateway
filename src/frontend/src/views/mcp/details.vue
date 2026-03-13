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
  <bk-sideslider
    v-model:is-show="isShow"
    width="960"
    :title="t('令牌详情')"
  >
    <template #default>
      <div class="content-wrapper mcp">
        <list-content-article>
          <list-content-row :label="t('ID')">{{ data?.id }}</list-content-row>
          <list-content-row :label="t('名称')">{{ data?.name }}</list-content-row>
          <list-content-row :label="t('令牌')">{{ data?.masked_token }}</list-content-row>
          <list-content-row :label="t('描述')">{{ data?.description }}</list-content-row>
          <list-content-row :label="t('访问范围')">
            <bk-tag
              radius="8px"
              :theme="data?.access_scope === 'readwrite' ? 'success' : 'info'"
            >
              {{ data?.access_scope }}
            </bk-tag>
          </list-content-row>
          <list-content-row :label="t('状态')">
            <bk-tag
              radius="8px"
              :theme="data?.is_expired ? 'danger' : 'success'"
            >
              {{ data?.is_expired ? t('已过期') : t('活跃') }}
            </bk-tag>
          </list-content-row>
          <list-content-row :label="t('过期时间')">
            {{ dayjs.unix(data?.expired_at).format('YYYY-MM-DD HH:mm:ss Z') }}
          </list-content-row>
          <list-content-row :label="t('最近使用时间')">
            <span v-if="data?.last_used_at === null">{{ t('从未使用') }}</span>
            <span v-else>
              {{ dayjs.unix(data?.last_used_at).format('YYYY-MM-DD HH:mm:ss Z') }}
              ({{ dayjs.unix(data?.last_used_at).fromNow() }})
            </span>
          </list-content-row>
          <list-content-row :label="t('创建时间')">
            {{ dayjs.unix(data?.created_at).format('YYYY-MM-DD HH:mm:ss Z') }}
          </list-content-row>
          <list-content-row :label="t('创建人')">{{ data?.creator }}</list-content-row>
          <list-content-row :label="t('更新时间')">
            {{ dayjs.unix(data?.updated_at).format('YYYY-MM-DD HH:mm:ss Z') }}
          </list-content-row>
          <list-content-row :label="t('更新人')">{{ data?.updater }}</list-content-row>
        </list-content-article>
      </div>
    </template>
  </bk-sideslider>
</template>

<script lang="ts" setup>
import dayjs from 'dayjs';
import { useI18n } from 'vue-i18n';
import { IMcpToken } from '@/types';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';

interface IProps {
  data?: IMcpToken
}

const isShow = defineModel<boolean>({
  required: true,
  default: false,
});

const {
  data = {},
} = defineProps<IProps>();

const { t } = useI18n();

</script>

<style lang="scss" scoped>

.content-wrapper.mcp {
  font-size: 14px;
  padding-top: 24px;

  :deep(.content-group-title) {
    display: none !important;
  }
}

</style>
