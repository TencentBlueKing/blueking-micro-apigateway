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
  <div class="page-content-wrapper">
    <div class="header-actions">
      <div class="left"></div>
      <div class="right flex-row justify-content-center">
        <BkDatePicker
          :key="dateKey"
          v-model="dateValue"
          use-shortcut-text
          format="yyyy-MM-dd HH:mm:ss"
          clearable
          class="mr8"
          type="datetimerange"
          :placeholder="t('选择日期时间范围')"
          :shortcuts="shortcutsRange"
          :shortcut-selected-index="shortcutSelectedIndex"
          @change="handleChange"
          @shortcut-change="handleShortcutChange"
          @clear="handleClearDate"
          @pick-success="handlePickSuccess"
          @selection-mode-change="handleSelectionModeChange"
        />
        <BkSearchSelect
          v-model="searchParams"
          v-click-outside="handleSearchOutside"
          :data="searchOptions"
          :placeholder="t('搜索 名称、ID、资源类型、操作类型、操作人')"
          clearable
          class="table-resource-search"
          unique-select
          @click.stop="handleSearchSelectClick"
        />
      </div>
    </div>
    <div class="table-wrapper">
      <MicroAgTable
        ref="tableRef"
        v-model:table-data="tableData"
        v-model:settings="settings"
        row-key="id"
        resizable
        :disable-data-page="true"
        :filter-value="filterData"
        :api-method="getTableData"
        :columns="columns"
        @clear-filter="handleClearFilter"
        @filter-change="handleFilterChange"
      >
      </MicroAgTable>
    </div>
  </div>

  <slider-log-diff-viewer
    v-model="diffSliderConfig.visible"
    :log="diffSliderConfig.log"
    :title-config="diffSliderConfig.titleConfig"
    @closed="handleClosed"
  />
</template>

<script lang="tsx" setup>
import { computed, ref, shallowRef, watch, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { IOperationAuditLogs, ITableMethod } from '@/types';
import dayjs from 'dayjs';
import { useCommon } from '@/store';
import { useDatePicker } from '@/hooks/use-date-picker';
import { useTableFilterChange } from '@/hooks/use-table-filter-change';
import { useSearchSelectPopoverHidden } from '@/hooks/use-search-select-popover-hidden';
import { getOperationAuditLogs } from '@/http/audit';
// @ts-ignore
import TagOperationType from '@/components/tag-operation-type.vue';
import SliderLogDiffViewer from '@/components/slider-log-diff-viewer.vue';
// @ts-ignore
import MicroAgTable from '@/components/micro-ag-table/table';
import {
  type PrimaryTableProps,
  type FilterValue,
} from '@blueking/tdesign-ui';

const common = useCommon();
const { t } = useI18n();
const { handleTableFilterChange } = useTableFilterChange();
const { handleSearchOutside, handleSearchSelectClick } = useSearchSelectPopoverHidden();

const filterData = ref<Record<string, any>>({});
const searchParams = ref<{ id: string, name: string, values?: { id: string, name: string }[] }[]>([]);
const tableData = ref([]);
const tableRef = useTemplateRef<InstanceType<typeof MicroAgTable> & ITableMethod>('tableRef');

const diffSliderConfig = ref<any>({
  visible: false,
  log: null,
  titleConfig: {
    title: t('查看差异'),
    before: t('变更前'),
    after: t('变更后'),
  },
});
const settings = shallowRef({
  size: 'small',
  checked: [],
  disabled: [],
});

const {
  dateValue,
  shortcutsRange,
  shortcutSelectedIndex,
  handleChange,
  handleClear,
  handleConfirm,
  handleShortcutChange,
  handleSelectionModeChange,
} = useDatePicker(filterData);

const dateKey = ref('dateKey');

const resourceTypeFilter = computed(() => {
  const list = Object.keys(common.enums?.resource_type ?? {})?.map((key: string) => ({
    label: common.enums?.resource_type[key],
    value: key,
  }));
  return [
    {
      label: t('全部'),
      value: '',
    },
    ...list,
  ];
});

const operationTypeFilter = computed(() => {
  const list = Object.keys(common.enums?.operation_type ?? {})?.map((key: string) => ({
    label: common.enums?.operation_type[key],
    value: key,
  }));
  return [
    {
      label: t('全部'),
      value: '',
    },
    ...list,
  ];
});

const columns: PrimaryTableProps['columns'] = [
  {
    title: t('名称'),
    colKey: 'names',
    ellipsis: true,
    cell: (h, { row }) => row.names?.join(', ') || '--',
  },
  {
    title: 'ID',
    colKey: 'resource_ids',
    ellipsis: true,
    cell: (h, { row }) => row.resource_ids?.join(', ') || '--',
  },
  {
    title: t('资源类型'),
    colKey: 'resource_type',
    ellipsis: true,
    cell: (h, { row }) => common.enums?.resource_type?.[row.resource_type] ?? '--',
    filter: {
      type: 'single',
      list: resourceTypeFilter.value,
    },
  },
  {
    title: t('操作类型'),
    colKey: 'operation_type',
    ellipsis: true,
    cell: (h, { row }) => {
      return (
        <TagOperationType type={row.operation_type} />
      );
    },
    filter: {
      type: 'single',
      list: operationTypeFilter.value,
    },
  },
  {
    title: t('操作人'),
    colKey: 'operator',
    ellipsis: true,
  },
  {
    title: t('操作时间'),
    colKey: 'created_at',
    ellipsis: true,
    cell: (h, { row }) => dayjs.unix(row.created_at).format('YYYY-MM-DD HH:mm:ss Z'),
  },
  {
    title: t('操作'),
    colKey: 'opt',
    fixed: 'right',
    cell: (h, { row }) => {
      return (
        <bk-button text theme="primary" onClick={() => handleAlter(row)}>
          { t('查看变更') }
        </bk-button>
      );
    },
  },
];

const searchOptions = computed(() => {
  return [
    {
      id: 'name',
      name: t('名称'),
      multiple: false,
    },
    {
      id: 'resource_id',
      name: 'ID',
      multiple: false,
    },
    {
      id: 'resource_type',
      name: t('资源类型'),
      children: Object.keys(common.enums?.resource_type ?? {})?.map((key: string) => ({
        name: common.enums?.resource_type[key],
        id: key,
      })),
    },
    {
      id: 'operation_type',
      name: t('操作类型'),
      children: Object.keys(common.enums?.operation_type ?? {})?.map((key: string) => ({
        name: common.enums?.operation_type[key],
        id: key,
      })),
    },
    {
      id: 'operator',
      name: t('操作人'),
      multiple: false,
    },
  ];
});

watch(
  () => searchParams.value,
  () => {
    handleSearch();
  },
);

const getList = () => {
  tableRef.value!.fetchData(filterData.value, { resetPage: true });
};

const getTableData = async (params: Record<string, any> = {}) => {
  const results = await getOperationAuditLogs({ gatewayId: common.gatewayId, query: params });
  return results ?? [];
};

const handleSearch = () => {
  const data: Record<string, any> = {};
  searchParams.value.forEach((option) => {
    if (option.values) {
      data[option.id] = option.values[0]?.id;
    }
  });
  const { start_time, end_time } = filterData.value;
  filterData.value = {
    ...data,
    start_time,
    end_time,
  };
  if (!start_time) {
    delete filterData.value.start_time;
    delete filterData.value.end_time;
  }
  getList();
};

const handleClearFilter = () => {
  filterData.value = {};
  searchParams.value = [];
  dateValue.value = [];
  shortcutSelectedIndex.value = -1;
  dateKey.value = String(+new Date());
};

const handlePickSuccess = () => {
  handleConfirm();
  getList();
};

const handleClearDate = () => {
  handleClear();
  getList();
};

// 处理表头筛选联动搜索框
const handleFilterChange: PrimaryTableProps['onFilterChange'] = (filterItem: FilterValue) => {
  handleTableFilterChange({
    filterItem,
    filterData,
    searchOptions,
    searchParams,
  });
  getList();
};

const handleAlter = async (row: IOperationAuditLogs) => {
  diffSliderConfig.value.log = row;
  diffSliderConfig.value.visible = true;
};

const handleClosed = () => {
  diffSliderConfig.value.log = null;
};

</script>

<style lang="scss" scoped>
.page-content-wrapper {
  min-height: calc(100vh - 157px);
  padding: 24px;

  .table-wrapper {
    background-color: #ffffff;
  }
}

.header-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;

  .left {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

:deep(.table-cell-actions) {
  display: flex;
  gap: 12px;
}
</style>
