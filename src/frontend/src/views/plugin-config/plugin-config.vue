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
  <TableResourceList
    :columns="columns"
    :delete-api="deletePluginConfig"
    :extra-search-params="searchParams"
    :query-list-params="{ apiMethod: getPluginConfigs }"
    :routes="{ create: 'plugin-config-create', edit: 'plugin-config-edit' }"
    resource-type="plugin_config"
    @check-resource="toggleResourceViewerSlider"
  />
  <SliderResourceViewer
    v-model="isResourceViewerShow"
    :resource="pluginConfig"
    :source="source"
    resource-type="plugin_config"
  />
</template>

<script lang="ts" setup>
import TableResourceList, { ISearchParam } from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { IColumn } from '@/types';
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import { deletePluginConfig, getPluginConfig, getPluginConfigs } from '@/http/plugin-config';
import { IPluginConfigDto } from '@/types/plugin-config';
import { useRoute } from 'vue-router';

const { t } = useI18n();
const route = useRoute();

const columns: IColumn[] = [
  {
    label: 'ID',
    field: 'id',
  },
  {
    label: t('描述'),
    render: ({ row }: any) => {
      return row.config?.desc || '--';
    },
  },
  {
    label: t('插件'),
    render: ({ row }: any) => {
      return Object.keys(row.config?.plugins || {})
        .join(', ') || '--';
    },
  },
];

const pluginConfig = ref<IPluginConfigDto>();
const source = ref('');
const isResourceViewerShow = ref(false);
const searchParams = ref<ISearchParam[]>([]);

const toggleResourceViewerSlider = ({ resource }: { resource: IPluginConfigDto }) => {
  pluginConfig.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

watch(() => route.query.id, async () => {
  if (route.query.id) {
    const id = route.query.id as string;
    const resource = await getPluginConfig({ id });
    toggleResourceViewerSlider({ resource });
    searchParams.value = [{
      id: 'id',
      name: 'ID',
      values: [{ id, name: id }],
    }];
  }
}, { immediate: true });

</script>
