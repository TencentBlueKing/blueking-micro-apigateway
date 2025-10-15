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
    :extra-search-options="extraSearchOptions"
    :delete-api="deleteConsumer"
    :routes="{ create: 'consumer-create', edit: 'consumer-edit', clone: 'consumer-clone' }"
    :query-list-params="{ apiMethod: getConsumers }"
    name-col-key="username"
    resource-type="consumer"
    @check-resource="toggleResourceViewerSlider"
  />
  <SliderResourceViewer
    v-model="isResourceViewerShow"
    :resource="consumer"
    :source="source"
    resource-type="consumer"
  />
</template>

<script lang="tsx" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { FilterOptionClass, type IFilterOption } from '@/types/table-filter';
import { IConsumer } from '@/types/consumer';
import { deleteConsumer, getConsumers } from '@/http/consumer';
import { getConsumerGroupDropdowns } from '@/http/consumer-group';
import { type PrimaryTableProps, type TableRowData } from '@blueking/tdesign-ui';
import TableResourceList from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';

const { t } = useI18n();
const router = useRouter();

const consumer = ref<IConsumer>();
// 关联的消费者组 select 下拉选项
const consumerGroupSelectOptions = ref<IFilterOption[]>([]);
// const selectedConsumerGroupId = ref('');

const source = ref('');
const isResourceViewerShow = ref(false);

let consumerGroupNameMap: Record<string, string> = {};

const columns = computed<PrimaryTableProps['columns']>(() => [
  {
    title: 'ID',
    colKey: 'id',
  },
  {
    title: t('消费者组'),
    colKey: 'group_id',
    filter: {
      type: 'single',
      showConfirmAndReset: true,
      popupProps: {
        overlayInnerClassName: 'custom-radio-filter-wrapper',
      },
      list: [],
    },
    cell: (h, { row }: TableRowData) => {
      const groupName = consumerGroupNameMap[row.group_id || row.config?.group_id];
      if (!groupName) {
        return '--';
      }
      return (
        <div class="flex-row">
          <span
            class="single-ellipse is-color-active"
            v-bk-tooltips={{
              content: groupName,
              placement: 'top',
              disabled: !row.isOverflow,
            }}
            onClick={() => handleRoute(row)}
          >
            {groupName}
          </span>
        </div>
      );
    },
  },
]);

const extraSearchOptions = computed(() => [
  {
    id: 'group_id',
    name: t('消费者组'),
    children: getFilterOptions({
      options: consumerGroupSelectOptions.value,
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

const getConsumerGroupSelectOptions = async () => {
  const response = await getConsumerGroupDropdowns();
  consumerGroupSelectOptions.value = (response ?? []).map(item => ({
    name: item.name,
    id: item.id,
    label: item.name,
    value: item.id,
    desc: item.desc,
  }));
  const filterOptions = getFilterOptions({ options: consumerGroupSelectOptions.value, extra: true });
  const groupCol = columns.value.find(col => ['group_id'].includes(col.colKey));
  if (groupCol) {
    groupCol.filter.list = filterOptions;
  }
  consumerGroupNameMap = consumerGroupSelectOptions.value.reduce((acc, cur) => {
    acc[cur.id] = cur.name;
    return acc;
  }, {});
};
getConsumerGroupSelectOptions();

const toggleResourceViewerSlider = ({ resource }: { resource: IConsumer }) => {
  consumer.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

const handleRoute = (row: TableRowData) => {
  const link = router.resolve({
    name: 'consumer-group',
    query: {
      id: row?.group_id || row?.config?.group_id,
    },
  });
  window.open(link.href, '_blank');
};
</script>
