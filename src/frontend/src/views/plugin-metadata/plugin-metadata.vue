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
    :delete-api="deletePluginMetadata"
    :query-list-params="{ apiMethod: getPluginMetadataList }"
    :routes="{ create: 'plugin-metadata-create', edit: 'plugin-metadata-edit' }"
    resource-type="plugin_metadata"
    :exclude-columns="['label']"
    @check-resource="toggleResourceViewerSlider"
  />
  <SliderResourceViewer
    v-model="isResourceViewerShow"
    :resource="pluginMetadata"
    :source="source"
    resource-type="plugin_metadata"
  />
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash-es';
import { ref } from 'vue';
import { type PrimaryTableProps } from '@blueking/tdesign-ui';
import { IPluginMetadataDto } from '@/types/plugin-metadata';
import { deletePluginMetadata, getPluginMetadataList } from '@/http/plugin-metadata';
import TableResourceList from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';

const columns: PrimaryTableProps['columns'] = [
  {
    title: 'ID',
    colKey: 'id',
  },
];

const source = ref('');
const isResourceViewerShow = ref(false);
const pluginMetadata = ref<IPluginMetadataDto>();

const toggleResourceViewerSlider = ({ resource }: { resource: IPluginMetadataDto }) => {
  pluginMetadata.value = resource;
  const displayedConfig = { config: cloneDeep(resource.config), id: resource.id };
  // plugin_metadata 的 config 比较特殊，需要隐藏避免和 row.id 冲突
  if (displayedConfig.config.id) {
    delete displayedConfig.config.id;
  }
  source.value = JSON.stringify(displayedConfig);
  isResourceViewerShow.value = true;
};
</script>
