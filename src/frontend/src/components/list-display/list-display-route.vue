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
      <list-content-row :label="t('名称')">{{ route.name }}</list-content-row>
      <list-content-row :label="t('ID')">{{ route.id }}</list-content-row>
      <list-content-row v-if="route.config.desc" :label="t('描述')">{{ route.config.desc }}</list-content-row>
      <list-content-row
        v-if="route.config.labels && Object.keys(route.config.labels).length"
        :label="t('标签')"
      >
        <tag-label :labels="route.config.labels" />
      </list-content-row>
      <list-content-rows-common :resource="route" />
      <list-content-row v-if="route.service_id" :label="t('绑定服务')">
        <bk-button
          text
          theme="primary"
          @click="handleResourceLinkClick('service', route.service_id)"
        >
          {{ route.service_id }}
        </bk-button>
      </list-content-row>
      <list-content-row :label="t('启用 WebSocket')">{{ route.config.enable_websocket }}</list-content-row>
    </list-content-article>

    <list-content-article>
      <template #title>{{ t('匹配条件') }}</template>
      <list-content-row v-if="route.config.hosts" :label="t('域名')">{{ route.config.hosts }}</list-content-row>
      <list-content-row :label="t('路径')">{{ route.config.uris }}</list-content-row>
      <list-content-row v-if="route.config.remote_addrs" :label="t('客户端地址')">
        {{ route.config.remote_addrs }}
      </list-content-row>
      <list-content-row :label="t('HTTP 方法')">{{ route.config.methods }}</list-content-row>
      <list-content-row v-if="route.config.priority && route.config.priority !== 0" :label="t('优先级')">
        {{ route.config.priority }}
      </list-content-row>
      <list-content-row v-if="route.config?.vars?.length" label="Vars">
        <div v-for="(item, index) in route.config.vars" :key="index">{{ item }}</div>
      </list-content-row>
    </list-content-article>

    <list-content-article v-if="route.upstream_id">
      <template #title>{{ t('上游服务') }}</template>
      <list-content-row :label="t('上游服务 ID')">
        <bk-button
          text
          theme="primary"
          @click="handleResourceLinkClick('upstream', route.upstream_id)"
        >
          {{ route.upstream_id }}
        </bk-button>
      </list-content-row>
    </list-content-article>

    <template v-if="route.config.upstream">
      <div style="font-size: 16px;font-weight: 700;color:#313238;margin-bottom: 15px;padding-top: 16px;">
        {{ t('上游服务') }}
      </div>
      <list-display-upstream :resource="route.config.upstream as IUpstream" />
    </template>

    <list-content-article v-if="route.plugin_config_id">
      <template #title>{{ t('插件组') }}</template>
      <list-content-row v-if="route.plugin_config_id" :label="t('插件组 ID')">
        <bk-button
          text
          theme="primary"
          @click="handleResourceLinkClick('plugin-config', route.plugin_config_id)"
        >
          {{ route.plugin_config_id }}
        </bk-button>
      </list-content-row>
    </list-content-article>
    <list-content-article v-if="route.config.plugins && Object.keys(route.config.plugins).length">
      <template #title>{{ t('已启用插件') }}</template>
      <list-content-row v-for="(config, configName) in route.config.plugins" :key="configName" :label="configName">
        <encode-json :config="config" />
      </list-content-row>
    </list-content-article>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';
import { IRoute } from '@/types/route';
import ListDisplayUpstream from '@/components/list-display/list-display-upstream.vue';
import { IUpstream } from '@/types/upstream';
import ListContentRowsCommon from '@/components/list-display/components/list-content-rows-common.vue';
import { useRouter } from 'vue-router';
import TagLabel from '@/components/tag-label.vue';
import EncodeJson from '@/components/list-display/components/encode-json.vue';

interface IProps {
  resource: IRoute
}

const { resource: route } = defineProps<IProps>();

const { t } = useI18n();
const router = useRouter();

const handleResourceLinkClick = (routeName: string, id: string) => {
  const to = router.resolve({ name: routeName, query: { id } });
  window.open(to.href);
};

</script>

<style lang="scss" scoped>

</style>
