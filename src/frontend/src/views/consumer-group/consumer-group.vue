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
  <table-resource-list
    :columns="columns"
    :delete-api="deleteConsumerGroup"
    :extra-search-params="searchParams"
    :query-list-params="{ apiMethod: getConsumerGroups }"
    :routes="{ create: 'consumer-group-create', edit: 'consumer-group-edit', clone: 'consumer-group-clone' }"
    resource-type="consumer_group"
    @check-resource="toggleResourceViewerSlider"
  />
  <slider-resource-viewer
    v-model="isResourceViewerShow"
    :resource="consumerGroup"
    :source="source"
    resource-type="consumer_group"
  />
</template>

<script lang="ts" setup>
import TableResourceList, { ISearchParam } from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { type PrimaryTableProps, type TableRowData } from '@blueking/tdesign-ui';
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import { deleteConsumerGroup, getConsumerGroup, getConsumerGroups } from '@/http/consumer-group';
import { IConsumerGroup } from '@/types/consumer-group';
import { useRoute } from 'vue-router';

const { t } = useI18n();
const route = useRoute();

const columns:  PrimaryTableProps['columns']  = [
  {
    title: 'ID',
    colKey: 'id',
  },
  {
    title: t('描述'),
    colKey: 'desc',
    cell: (h, { row }: TableRowData) => {
      return row.config?.desc || '--';
    },
  },
];

const consumerGroup = ref<IConsumerGroup>();
const source = ref('');
const isResourceViewerShow = ref(false);
const searchParams = ref<ISearchParam[]>([]);

const toggleResourceViewerSlider = ({ resource }: { resource: IConsumerGroup }) => {
  consumerGroup.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

watch(() => route.query.id, async () => {
  if (route.query.id) {
    const id = route.query.id as string;
    const resource = await getConsumerGroup({ id });
    toggleResourceViewerSlider({ resource });
    searchParams.value = [{
      id: 'id',
      name: 'ID',
      values: [{ id, name: id }],
    }];
  }
}, { immediate: true });

</script>
