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
    <list-content-article>
      <template #title>{{ t('基本信息') }}</template>
      <list-content-row :label="t('名称')">{{ service.name }}</list-content-row>
      <list-content-row :label="t('ID')">{{ service.id }}</list-content-row>
      <list-content-row v-if="service.config.desc" :label="t('描述')">{{ service.config.desc }}</list-content-row>
      <list-content-row
        v-if="service.config.labels && Object.keys(service.config.labels).length" :label="t('标签')"
      >{{ service.config.labels }}
      </list-content-row>
      <list-content-row
        v-if="service.config.labels && Object.keys(service.config.labels).length"
        :label="t('标签')"
      >
        <tag-label :labels="service.config.labels" />
      </list-content-row>
      <list-content-rows-common :resource="service" />
      <list-content-row v-if="service.config.hosts && service.config.hosts.length" :label="t('匹配域名')">
        {{ service.config.hosts }}
      </list-content-row>
      <list-content-row :label="t('启用 WebSocket')">{{ service.config.enable_websocket }}</list-content-row>
    </list-content-article>

    <list-content-article v-if="service.config.plugins && Object.keys(service.config.plugins).length">
      <template #title>{{ t('已启用插件') }}</template>
      <list-content-row v-for="(config, configName) in service.config.plugins" :key="configName" :label="configName">
        <encode-json :config="config" />
      </list-content-row>
    </list-content-article>

    <list-content-article v-if="service.config.upstream_id">
      <template #title>{{ t('上游服务') }}</template>
      <list-content-row :label="t('上游服务 ID')">
        <bk-button text theme="primary" @click="handleUpstreamIdClick">{{ service.config.upstream_id }}</bk-button>
      </list-content-row>
    </list-content-article>

    <template v-if="service.config.upstream">
      <div style="font-size: 16px;font-weight: 700;color:#313238;margin-bottom: 15px;padding-top: 16px;">
        {{ t('上游服务') }}
      </div>
      <list-display-upstream :resource="service.config.upstream as IUpstream" />
    </template>

  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';
import { IService } from '@/types/service';
import ListDisplayUpstream from '@/components/list-display/list-display-upstream.vue';
import { IUpstream } from '@/types/upstream';
import ListContentRowsCommon from '@/components/list-display/components/list-content-rows-common.vue';
import { useRouter } from 'vue-router';
import TagLabel from '@/components/tag-label.vue';
import EncodeJson from '@/components/list-display/components/encode-json.vue';

interface IProps {
  resource: IService
}

const { resource: service } = defineProps<IProps>();

const { t } = useI18n();
const router = useRouter();

const handleUpstreamIdClick = () => {
  const to = router.resolve({ name: 'upstream', query: { id: service.config.upstream_id } });
  window.open(to.href);
};

</script>

<style lang="scss" scoped>

</style>
