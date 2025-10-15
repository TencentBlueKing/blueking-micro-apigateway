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
    :extra-search-options="extraSearchOptions"
    :delete-api="deleteStreamRoute"
    :query-list-params="{ apiMethod: getStreamRouteList }"
    :routes="{ create: 'stream-route-create', edit: 'stream-route-edit', clone: 'stream-route-clone' }"
    :resource-type="'stream_route'"
    @check-resource="toggleResourceViewerSlider"
  />
  <!-- 发布/差异slider -->
  <SliderResourceViewer
    v-model="isResourceViewerShow"
    editable
    :resource="streamRoute"
    :source="source"
    resource-type="stream_route"
    @updated="handleUpdated"
  />
</template>

<script lang="tsx" setup>
import { computed, ref } from 'vue';
import { IStreamRoute } from '@/types';
import { FilterOptionClass, type IFilterOption } from '@/types/table-filter';
import { type PrimaryTableProps, type TableRowData } from '@blueking/tdesign-ui';
import { deleteStreamRoute, getStreamRouteList, getStreamRoute } from '@/http/stream-route';
import TableResourceList from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { useI18n } from 'vue-i18n';
import { getServiceDropdowns } from '@/http/service';
import { getUpstreamDropdowns } from '@/http/upstream';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const router = useRouter();

const tableRef = ref<InstanceType<typeof TableResourceList>>(null);
const streamRoute = ref<IStreamRoute>();

// 关联的服务 select 下拉选项
const serviceSelectOptions = ref<{ value: string, label: string, desc: string }[]>([]);
// 关联的上游 select 下拉选项
const upstreamSelectOptions = ref<{ value: string, label: string, desc: string }[]>([]);

const isResourceViewerShow = ref(false);
const source = ref('');

let serviceNameMap: Record<string, string> = {};
let upstreamNameMap: Record<string, string> = {};

const columns = computed<PrimaryTableProps['columns']>(() => {
  return [
    {
      title: 'ID',
      colKey: 'id',
      ellipsis: true,
    },
    {
      title: t('描述'),
      colKey: 'desc',
      ellipsis: true,
      cell: (h, { row }: TableRowData) => row.config?.desc || '--',
    },
    {
      title: t('服务'),
      colKey: 'service_id',
      ellipsis: true,
      width: 120,
      filter: {
        type: 'single',
        showConfirmAndReset: true,
        list: getFilterOptions({
          options: serviceSelectOptions.value,
          extra: true,
        }),
        popupProps: {
          overlayInnerClassName: 'custom-radio-filter-wrapper',
        },
      },
      cell: (h, { row }: TableRowData) => (row.service_id
        ? <bk-button
        text theme="primary"
        onClick={() => handleRelatedResourceIdClicked({ routeName: 'service', id: row.service_id })}
      >{serviceNameMap[row.service_id]}</bk-button> : '--'),
    },
    {
      title: t('上游'),
      colKey: 'upstream_id',
      ellipsis: true,
      width: 120,
      filter: {
        type: 'single',
        showConfirmAndReset: true,
        list: getFilterOptions({ options: upstreamSelectOptions.value, extra: true }),
        popupProps: {
          overlayInnerClassName: 'custom-radio-filter-wrapper',
        },
      },
      cell: (h, { row }: TableRowData) => (row.upstream_id
        ? <bk-button
        text theme="primary"
        onClick={() => handleRelatedResourceIdClicked({ routeName: 'upstream', id: row.upstream_id })}
      >{upstreamNameMap[row.upstream_id]}</bk-button> : '--'),
    },
  ];
});

const extraSearchOptions = computed(() => [
  {
    id: 'service_id',
    name: t('服务'),
    children: getFilterOptions({
      options: serviceSelectOptions.value,
      key: 'name',
      value: 'id',
      extra: true,
    }),
  },
  {
    id: 'upstream_id',
    name: t('上游'),
    children: getFilterOptions({
      options: upstreamSelectOptions.value,
      key: 'name',
      value: 'id',
      extra: true,
    }),
  },
]);

// 根据不同键值初始化数组结构
function getFilterOptions({
  key,
  value,
  options,
  extra,
}: {
  key: string,
  value?: string | number,
  options: IFilterOption[],
  extraOption?: boolean | IFilterOption[],
})  {
  return new FilterOptionClass({ key, value, options, extra })?.filterOptions;
};

const getServiceSelectOptions = async () => {
  const response = await getServiceDropdowns();
  serviceSelectOptions.value = (response ?? []).map(item => ({
    name: item.name,
    id: item.id,
    label: item.name,
    value: item.id,
    desc: item.desc,
  }));
  serviceNameMap = serviceSelectOptions.value.reduce<Record<string, string>>((acc, cur) => {
    acc[cur.id] = cur.name;
    return acc;
  }, {});
};
getServiceSelectOptions();

const getUpstreamSelectOptions = async () => {
  const response = await getUpstreamDropdowns();
  upstreamSelectOptions.value = (response ?? []).map(item => ({
    name: item.name,
    id: item.id,
    label: item.name,
    value: item.id,
    desc: item.desc,
  }));
  const filterOptions = getFilterOptions({ options: upstreamSelectOptions.value, extra: true });
  const groupCol = columns.value.find(col => ['upstream_id'].includes(col.colKey));
  if (groupCol) {
    groupCol.filter.list = filterOptions;
  }
  upstreamNameMap = upstreamSelectOptions.value.reduce<Record<string, string>>((acc, cur) => {
    acc[cur.id] = cur.name;
    return acc;
  }, {});
};
getUpstreamSelectOptions();

const handleRelatedResourceIdClicked = ({ routeName, id }: { routeName: string, id: string }) => {
  const to = router.resolve({ name: routeName, query: { id } });
  window.open(to.href);
};

const toggleResourceViewerSlider = ({ resource }: { resource: IStreamRoute }) => {
  streamRoute.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

const handleUpdated = async () => {
  tableRef.value!.getList();
  streamRoute.value = await getStreamRoute({ id: streamRoute.value.id });
};
</script>
