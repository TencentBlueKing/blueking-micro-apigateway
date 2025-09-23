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
      <list-content-row :label="t('名称')">{{ proto.name }}</list-content-row>
      <list-content-row :label="t('ID')">{{ proto.id }}</list-content-row>
      <list-content-row v-if="proto.config.desc" :label="t('描述')">{{
        proto.config.desc
      }}
      </list-content-row>
      <list-content-rows-common :resource="proto" />
    </list-content-article>

    <list-content-article v-if="proto.config.content">
      <template #title>Content</template>
      <list-content-row label="content">
        {{ proto.config.content }}
      </list-content-row>
    </list-content-article>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';
import { IProto } from '@/types/proto';
import ListContentRowsCommon from '@/components/list-display/components/list-content-rows-common.vue';

interface IProps {
  resource: IProto
}

const { resource: proto } = defineProps<IProps>();

const { t } = useI18n();

</script>

<style lang="scss" scoped>

</style>
