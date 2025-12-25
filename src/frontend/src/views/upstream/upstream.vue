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
    ref="tableRef"
    :columns="columns"
    :delete-api="deleteUpstream"
    :extra-search-params="searchParams"
    :query-list-params="{ apiMethod: getUpstreams }"
    :routes="{ create: 'upstream-create', edit: 'upstream-edit', clone: 'upstream-clone' }"
    resource-type="upstream"
    @check-resource="toggleResourceViewerSlider"
  />
  <SliderResourceViewer
    v-model="isResourceViewerShow"
    editable
    :resource="upstream"
    :source="source"
    resource-type="upstream"
    @updated="handleUpdated"
  />
</template>

<script lang="ts" setup>
import { type PrimaryTableProps } from '@blueking/tdesign-ui';
import TableResourceList, { ISearchParam } from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { deleteUpstream, getUpstream, getUpstreams } from '@/http/upstream';
import { ref, watch } from 'vue';
import { IUpstream } from '@/types/upstream';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

const { t } = useI18n();
const route = useRoute();

const columns: PrimaryTableProps['columns']  = [
  {
    title: 'ID',
    colKey: 'id',
  },
  {
    title: t('类型'),
    colKey: 'config.type',
  },
];

const upstream = ref<IUpstream>();
const source = ref('');
const isResourceViewerShow = ref(false);
const searchParams = ref<ISearchParam[]>([]);
const tableRef = ref();

const toggleResourceViewerSlider = ({ resource }: { resource: IUpstream }) => {
  upstream.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

watch(() => route.query.id, async () => {
  if (route.query.id) {
    const id = route.query.id as string;
    const resource = await getUpstream({ id });
    toggleResourceViewerSlider({ resource });
    searchParams.value = [{
      id: 'id',
      name: 'ID',
      values: [{ id, name: id }],
    }];
  }
}, { immediate: true });

const handleUpdated = async () => {
  tableRef.value!.getList();
  upstream.value = await getUpstream({ id: upstream.value.id });
};

</script>
