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
    <ListContentArticle>
      <template #title>{{ t('基本信息') }}</template>
      <ListContentRow :label="t('名称')">{{ streamRoute.name }}</ListContentRow>
      <ListContentRow :label="t('ID')">{{ streamRoute.id }}</ListContentRow>
      <ListContentRow v-if="streamRoute.config.desc" :label="t('描述')">
        {{ streamRoute.config.desc }}
      </ListContentRow>
      <ListContentRow
        v-if="streamRoute.config.labels && Object.keys(streamRoute.config.labels).length"
        :label="t('标签')"
      >
        <TagLabel :labels="streamRoute.config.labels" />
      </ListContentRow>
      <ListContentRowsCommon :resource="streamRoute" />
      <ListContentRow v-if="streamRoute.service_id" :label="t('绑定服务')">
        <BkButton
          text
          theme="primary"
          @click="handleResourceLinkClick('service', streamRoute.service_id)"
        >
          {{ streamRoute.service_id }}
        </BkButton>
      </ListContentRow>
      <ListContentRow v-if="streamRoute.config.remote_addr" :label="t('上游地址')">
        {{ streamRoute.config.remote_addr }}
      </ListContentRow>
      <ListContentRow v-if="streamRoute.config.server_addr" :label="t('服务器地址')">
        {{ streamRoute.config.server_addr }}
      </ListContentRow>
      <ListContentRow v-if="streamRoute.config.server_port" :label="t('服务器端口')">
        {{ streamRoute.config.server_port }}
      </ListContentRow>
    </ListContentArticle>

    <ListContentArticle v-if="streamRoute.upstream_id">
      <template #title>{{ t('上游服务') }}</template>
      <ListContentRow :label="t('上游服务 ID')">
        <BkButton
          text
          theme="primary"
          @click="handleResourceLinkClick('upstream', streamRoute.upstream_id)"
        >
          {{ streamRoute.upstream_id }}
        </BkButton>
      </ListContentRow>
    </ListContentArticle>

    <template v-if="streamRoute.config.upstream">
      <div style="font-size: 16px;font-weight: 700;color:#313238;margin-bottom: 15px;padding-top: 16px;">
        {{ t('上游服务') }}
      </div>
      <ListDisplayUpstream :resource="streamRoute.config.upstream" />
    </template>

    <ListContentArticle v-if="streamRoute.plugin_config_id">
      <template #title>{{ t('插件组') }}</template>
      <ListContentRow v-if="streamRoute.plugin_config_id" :label="t('插件组 ID')">
        <BkButton
          text
          theme="primary"
          @click="handleResourceLinkClick('plugin-config', streamRoute.plugin_config_id)"
        >
          {{ streamRoute.plugin_config_id }}
        </BkButton>
      </ListContentRow>
    </ListContentArticle>
    <ListContentArticle v-if="streamRoute.config.plugins && Object.keys(streamRoute.config.plugins).length">
      <template #title>{{ t('已启用插件') }}</template>
      <ListContentRow v-for="(config, configName) in streamRoute.config.plugins" :key="configName" :label="configName">
        <encode-json :config="config" />
      </ListContentRow>
    </ListContentArticle>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { IStreamRoute } from '@/types/stream-route';
import TagLabel from '@/components/tag-label.vue';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';
import ListDisplayUpstream from '@/components/list-display/list-display-upstream.vue';
import ListContentRowsCommon from '@/components/list-display/components/list-content-rows-common.vue';
import EncodeJson from '@/components/list-display/components/encode-json.vue';

interface IProps {
  resource: IStreamRoute
}

const { resource: streamRoute } = defineProps<IProps>();

const { t } = useI18n();
const router = useRouter();

const handleResourceLinkClick = (routeName: string, id: string) => {
  const to = router.resolve({ name: routeName, query: { id } });
  window.open(to.href);
};
</script>
