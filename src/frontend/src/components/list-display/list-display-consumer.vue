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
      <list-content-row :label="t('名称')">{{ consumer.name }}</list-content-row>
      <list-content-row :label="t('ID')">{{ consumer.id }}</list-content-row>
      <list-content-row v-if="consumer.config.desc" :label="t('描述')">{{ consumer.config.desc }}</list-content-row>
      <list-content-row
        v-if="consumer.config.labels && Object.keys(consumer.config.labels).length"
        :label="t('标签')"
      >
        <tag-label :labels="consumer.config.labels" />
      </list-content-row>
      <list-content-row v-if="consumer.group_id || consumer.config?.group_id" :label="t('消费者组 ID')">
        <bk-button
          text
          theme="primary"
          @click="handleConsumerGroupIdClick(consumer.group_id || consumer.config?.group_id)"
        >
          {{ consumer.group_id || consumer.config?.group_id }}
        </bk-button>
      </list-content-row>
      <list-content-rows-common :resource="consumer" />
    </list-content-article>

    <list-content-article v-if="consumer.config.plugins && Object.keys(consumer.config.plugins).length">
      <template #title>{{ t('已启用插件') }}</template>
      <list-content-row
        v-for="(config, configName) in consumer.config.plugins" :key="configName" :label="configName as string"
      >
        {{ config }}
      </list-content-row>
    </list-content-article>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';
import { IConsumer } from '@/types/consumer';
import ListContentRowsCommon from '@/components/list-display/components/list-content-rows-common.vue';
import { useRouter } from 'vue-router';
import TagLabel from '@/components/tag-label.vue';

interface IProps {
  resource: IConsumer
}

const { resource: consumer } = defineProps<IProps>();

const { t } = useI18n();
const router = useRouter();

const handleConsumerGroupIdClick = (id: string) => {
  const to = router.resolve({ name: 'consumer-group', query: { id } });
  window.open(to.href);
};

</script>

<style lang="scss" scoped>

</style>
