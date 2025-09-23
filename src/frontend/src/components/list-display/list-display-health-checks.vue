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
  <div>
    <template v-if="checks?.active">
      <list-content-article>
        <template #title>{{ `${t('健康检查')} - ${t('主动检查')}` }}</template>
        <list-content-row :label="t('类型')">{{ checks.active.type }}</list-content-row>
        <list-content-row :label="t('超时时间(s)')">{{ checks.active.timeout }}</list-content-row>
        <list-content-row :label="t('并行数量')">{{ checks.active.concurrency }}</list-content-row>
        <list-content-row :label="t('主机名')">{{ checks.active.host }}</list-content-row>
        <list-content-row :label="t('端口')">{{ checks.active.port }}</list-content-row>
        <list-content-row :label="t('请求路径')">{{ checks.active.http_path }}</list-content-row>
        <list-content-row :label="t('请求头')">{{ checks.active.req_headers }}</list-content-row>
      </list-content-article>

      <list-content-article>
        <template #title>{{ `${t('健康检查')} - ${t('主动检查')} - ${t('健康状态')}` }}</template>
        <list-content-row :label="t('间隔时间(s)')">{{ checks.active.healthy.interval }}</list-content-row>
        <list-content-row :label="t('成功次数')">{{ checks.active.healthy.successes }}</list-content-row>
        <list-content-row :label="t('状态码')">{{ checks.active.healthy.http_statuses }}</list-content-row>
      </list-content-article>

      <list-content-article>
        <template #title>{{ `${t('健康检查')} - ${t('主动检查')} - ${t('不健康状态')}` }}</template>
        <list-content-row :label="t('超时时间(s)')">{{ checks.active.unhealthy.timeouts }}</list-content-row>
        <list-content-row :label="t('间隔时间(s)')">{{ checks.active.unhealthy.interval }}</list-content-row>
        <list-content-row :label="t('状态码')">{{ checks.active.unhealthy.http_statuses }}</list-content-row>
        <list-content-row :label="t('HTTP 失败次数')">{{ checks.active.unhealthy.http_failures }}</list-content-row>
        <list-content-row :label="t('TCP 失败次数')">{{ checks.active.unhealthy.tcp_failures }}</list-content-row>
      </list-content-article>
    </template>

    <template v-if="checks?.passive">
      <list-content-article>
        <template #title>{{ `${t('健康检查')} - ${t('被动检查')}` }}</template>
        <list-content-row :label="t('类型')">{{ checks.passive.type }}</list-content-row>
      </list-content-article>

      <list-content-article>
        <template #title>{{ `${t('健康检查')} - ${t('被动检查')} - ${t('健康状态')}` }}</template>
        <list-content-row :label="t('状态码')">{{ checks.passive.healthy.http_statuses }}</list-content-row>
        <list-content-row :label="t('成功次数')">{{ checks.passive.healthy.successes }}</list-content-row>
      </list-content-article>

      <list-content-article>
        <template #title>{{ `${t('健康检查')} - ${t('被动检查')} - ${t('不健康状态')}` }}</template>
        <list-content-row :label="t('超时时间(s)')">{{ checks.passive.unhealthy.timeouts }}</list-content-row>
        <list-content-row :label="t('TCP 失败次数')">{{ checks.passive.unhealthy.tcp_failures }}</list-content-row>
        <list-content-row :label="t('HTTP 失败次数')">{{ checks.passive.unhealthy.http_failures }}</list-content-row>
        <list-content-row :label="t('状态码')">{{ checks.passive.unhealthy.http_statuses }}</list-content-row>
      </list-content-article>
    </template>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { IHealthCheck } from '@/types';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';

interface IProps {
  checks: IHealthCheck
}

const { checks } = defineProps<IProps>();

const { t } = useI18n();

</script>
