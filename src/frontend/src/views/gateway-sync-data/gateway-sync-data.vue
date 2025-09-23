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
      <div class="left">
        <BkButton
          theme="primary"
          :loading="isSyncLoading"
          @click="handleSync"
        >
          <i class="icon apigateway-icon icon-ag-download-line" />
          {{ t('一键同步') }}
        </BkButton>
        <!-- <bk-dropdown trigger="click">
          <bk-button :disabled="!selections.length">{{ t('批量操作') }}</bk-button>
          <template #content>
            <bk-dropdown-menu>
              <bk-dropdown-item @click="handleMultiAdd">
                <bk-button text>
                  {{ t('添加到编辑区') }}
                </bk-button>
              </bk-dropdown-item>
            </bk-dropdown-menu>
          </template>
        </bk-dropdown> -->
      </div>
      <div class="right">
        <BkSearchSelect
          v-model="searchParams"
          v-click-outside="handleSearchOutside"
          :data="searchOptions"
          :placeholder="t('搜索 名称、ID、资源类型、状态')"
          class="table-resource-search"
          clearable
          unique-select
          @click.stop="handleSearchSelectClick"
        />
      </div>
    </div>
    <div class="table-wrapper">
      <div class="table-top-total">
        <span class="equal">{{ equalCount }}</span>
        <span>{{ t('一致') }}</span>
        <span class="line"></span>
        <span class="lose">{{ loseCount }}</span>
        <span>{{ t('缺失') }}</span>
        <template v-if="loseCount">
          <span> , </span>
          <span class="add-opt" @click="handleBatchAdd">{{ t('一键添加到编辑区维护') }}</span>
        </template>
      </div>
      <MicroAgTable
        ref="tableRef"
        v-model:selected-row-keys="selectionHook.selectionsRowKeys.value"
        v-model:table-data="tableData"
        v-model:settings="settings"
        v-model:is-all-selection="selectionHook.isAllSelection.value"
        row-key="id"
        resizable
        :disable-data-page="true"
        :filter-value="filterData"
        :sort="sortData"
        :api-method="getTableData"
        :columns="columns"
        :is-show-first-full-row="selectionHook.selections.value.length > 0"
        @request-done="handleRequestDone"
        @select-change="selectionHook.handleSelectionChange"
        @clear-filter="handleClearFilter"
        @filter-change="handleFilterChange"
        @sort-change="handleSortChange"
      >
        <template #firstFullRow>
          <div class="table-first-full-row">
            <span class="normal-text">
              <span>{{ t('已选') }}</span>
              <span class="count">{{ selectionHook.selections.value.length }}</span>
              <span>{{ t('条') }}</span>
              <span class="mr4">,</span>
            </span>
            <span class="hight-light-text" @click="handleMultiAdd">{{ t('添加到编辑区') }}</span>
          </div>
        </template>
      </MicroAgTable>
    </div>
  </div>

  <SliderResourceViewer
    v-model="isResourceViewerShow"
    :resource="resource"
    :source="source"
    :resource-type="resourceType"
  />

  <SliderResourceDiffViewer
    v-model="diffSliderConfig.visible"
    :after-config="diffSliderConfig.after_config"
    :before-config="diffSliderConfig.before_config"
    :operation-type="diffSliderConfig.operationType"
    :title-config="diffSliderConfig.titleConfig"
  />
</template>

<script lang="tsx" setup>
import { computed, shallowRef, ref, watch, h, useTemplateRef } from 'vue';
import { useRouter } from 'vue-router';
import { ITableMethod } from '@/types';
import { IGatewaySyncDataDto } from '@/types/gateway-sync-data';
import { cloneDeep } from 'lodash-es';
import { useCommon } from '@/store';
import { useI18n } from 'vue-i18n';
import { Message, InfoBox, Checkbox } from 'bkui-vue';
import { useTDesignSelection } from '@/hooks/use-tdesign-selection';
import { useTableFilterChange } from '@/hooks/user-table-filter-change';
import { useTableSortChange } from '@/hooks/use-table-sort-change';
import { useSearchSelectPopoverHidden } from '@/hooks/use-search-select-popover-hidden';
import {
  type PrimaryTableProps,
  type SortInfo,
  type FilterValue,
} from '@blueking/tdesign-ui';
import dayjs from 'dayjs';
// @ts-ignore
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
// @ts-ignore
import SliderResourceDiffViewer from '@/components/slider-resource-diff-viewer.vue';
import MicroAgTable from '@/components/micro-ag-table/table';
import { addResourceToEditArea, getGatewaySyncDataList, postGatewaySyncData } from '@/http/gateway-sync-data';
import { getResourceDiff } from '@/http/publish';
import { Copy } from 'bkui-vue/lib/icon';
import { handleCopy } from '@/common/util';

const router = useRouter();
const common = useCommon();
const selectionHook = useTDesignSelection();
const { t } = useI18n();
const { handleTableFilterChange } = useTableFilterChange();
const { handleTableSortChange } = useTableSortChange();
const { handleSearchOutside, handleSearchSelectClick } = useSearchSelectPopoverHidden();

const isSyncLoading = ref<boolean>(false);
const filterData = ref<FilterValue>({});
const sortData =  ref<SortInfo>({});
const searchParams = ref<{ id: string, name: string, values?: { id: string, name: string }[] }[]>([]);
const isResourceViewerShow = ref<boolean>(false);
const resource = ref();
const source = ref<string>('');
const resourceType = ref<string>('');
const diffSliderConfig = ref({
  visible: false,
  after_config: {},
  before_config: {},
  operationType: '',
  titleConfig: {
    title: t('查看差异'),
    before: 'etcd',
    after: t('本地'),
  },
});
const tableData = ref([]);
const tableRef = useTemplateRef<InstanceType<typeof MicroAgTable> & ITableMethod>('tableRef');
const settings = shallowRef({
  size: 'small',
  checked: [],
  disabled: [],
});
const allowSortField = shallowRef(['name']);

const isDisabledSelected = (item) => {
  return !['miss'].includes(item.status);
};

const isAllowCheckSelection = (item) => {
  return ['miss'].includes(item.status);
};

const statusList = [
  {
    name: t('一致'),
    id: 'success',
  },
  {
    name: t('缺失'),
    id: 'miss',
  },
];

const resourceTypeFilter = computed(() => {
  const list = Object.keys(common.enums?.resource_type ?? {})?.map((key: string) => ({
    label: common.enums?.resource_type[key],
    value: key,
  }))
    ?.filter(item => !['gateway'].includes(item.value));
  return [
    {
      label: t('全部'),
      value: '',
    },
    ...list,
  ];
});

const statusFilter = computed(() => {
  return [
    {
      label: t('全部'),
      value: '',
    },
    ...statusList.map((st) => {
      return {
        label: st.name,
        value: st.id,
      };
    }),
  ];
});

// 设置表格半选效果
const setIndeterminate = computed(() => {
  const isExistCheck = tableData.value.some(item => item.isCustomCheck);
  return isExistCheck && selectionHook.selections.value.length > 0 && !selectionHook.isAllSelection.value;
});

const columns: PrimaryTableProps['columns'] = [
  {
    colKey: 'row-select',
    type: 'custom-checkbox',
    align: 'center',
    width: 60,
    fixed: 'left',
    title: () => {
      return (
        <Checkbox
          v-model={selectionHook.isAllSelection.value}
          onChange={() => {
            tableData.value.forEach((item) => {
              if (isAllowCheckSelection(item)) {
                item.isCustomCheck = selectionHook.isAllSelection.value;
              }
            });
            const tables = tableData.value.filter(item => isAllowCheckSelection(item));
            selectionHook.handleCustomSelectAllChange({ isCheck: selectionHook.isAllSelection.value, tableRowKey: 'id', tables });
          }}
          disabled={disabledSelected.value}
          indeterminate={setIndeterminate.value}
        />
      );
    },
    cell: (h, { row }) => {
      return (
        <Checkbox
          v-model={row.isCustomCheck}
          v-bk-tooltips={{
            content: t('资源状态与编辑区一致，无需添加'),
            placement: 'top',
            disabled: !isDisabledSelected(row),
          }}
          disabled= {isDisabledSelected(row)}
          onChange={(isCheck: boolean) => {
            const results = tableData.value.filter(item => ['miss'].includes(item.status));
            selectionHook.handleCustomSelectChange({ isCheck, tableRowKey: 'id', row });
            const checkedIds = tableData.value
              .filter(item => selectionHook.selectionsRowKeys.value.includes(item.id))
              .map(check => check.id);
            selectionHook.isAllSelection.value = checkedIds.length > 0 && checkedIds.length === results.length;
            tableData.value.forEach((item) => {
              if (['miss'].includes(item.status) && row.id === item.id) {
                item.isCustomCheck = row.isCustomCheck;
              }
            });
          }}
        />
      );
    },
  },
  {
    title: t('名称'),
    colKey: 'name',
    fixed: 'left',
    sorter: true,
    cell: (h, { row }) => {
      if (!row?.name) {
        return '--';
      }
      return (
        <div class="flex-row align-items-center">
          <div
            class="single-ellipse"
            v-bk-tooltips={{
              content: row.name,
              placement: 'top',
              disabled: !row.isOverflow,
            }}
           >
            {row.name}
          </div>
          <Copy
            class="default-c pointer ml8"
            onClick={() => handleCopy(row.name)}
           />
        </div>
      );
    },
  },
  {
    title: t('ID'),
    colKey: 'id',
    cell: (h, { row }) => {
      return (
        <div class="flex-row align-items-center">
          <div
            class="single-ellipse"
            v-bk-tooltips={{
              content: row.id,
              placement: 'top',
              disabled: !row.isOverflow,
            }}
           >
            {row.id}
          </div>
          <Copy
            class="default-c pointer ml8"
            onClick={() => handleCopy(row.id)}
           />
        </div>
      );
    },
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
    title: t('同步版本'),
    colKey: 'mode_revision',
  },
  {
    title: t('同步时间'),
    colKey: 'updated_at',
    ellipsis: true,
    cell: (h, { row }) => {
      return dayjs.unix(row.updated_at).format('YYYY-MM-DD HH:mm:ss Z');
    },
  },
  {
    title: t('发布来源'),
    colKey: 'publish_source',
    ellipsis: true,
  },
  {
    title: t('状态'),
    colKey: 'status',
    cell: (h, { row }) => <bk-tag theme={row.status === 'success' ? 'success' : 'danger'}>{ row.status === 'success' ? '一致' : '缺失' }</bk-tag>,
    filter: {
      type: 'single',
      list: statusFilter.value,
    },
  },
  {
    title: t('操作'),
    colKey: 'opt',
    fixed: 'right',
    cell: (h, { row }) => {
      return (
        <div class="table-cell-actions">
          <bk-button text theme="primary" onClick={() => handleCheck(row)}>
            {t('查看')}
          </bk-button>

          {
            row?.status === 'success'
            && <bk-button text theme="primary" onClick={() => handleCheckDiff(row)}>
              {t('对比')}
            </bk-button>
          }

          {
            row?.status === 'miss'
            && <bk-button text theme="primary" onClick={() => handleAdd(row)}>
              {t('添加')}
            </bk-button>
          }
        </div>
      );
    },
  },
];

const searchOptions = computed(() => {
  return [
    {
      id: 'name',
      name: t('名称'),
    },
    {
      id: 'id',
      name: 'ID',
    },
    {
      id: 'resource_type',
      name: t('资源类型'),
      children: Object.keys(common.enums?.resource_type ?? {})?.map((key: string) => ({
        name: common.enums?.resource_type[key],
        id: key,
      }))
        ?.filter(item => item.id !== 'gateway'),
    },
    {
      id: 'status',
      name: t('状态'),
      children: statusList,
    },
  ];
});

const showDataList = computed(() => {
  if (!tableData.value?.length) return [];

  return tableData.value.map((item: IGatewaySyncDataDto) => {
    if (item.resource_type === 'consumer') {
      item.name = item.config.username;
    } else {
      item.name = item.config.name;
    }

    return item;
  });
});

const equalCount = computed(() => {
  return showDataList.value.filter(item => item.status === 'success').length;
});

const loseCount = computed(() => {
  return showDataList.value.filter(item => item.status === 'miss').length;
});

const disabledSelected = computed(() => {
  return !tableData.value?.length || tableData.value?.every(item => !['miss'].includes(item.status));
});

watch(
  () => searchParams.value,
  () => {
    handleSearch();
  },
);

const getTableData = async (params: Record<string, any> = {}) => {
  const results = await getGatewaySyncDataList({ gatewayId: common.gatewayId, query: params });
  return results ?? [];
};

const getList = () => {
  tableRef.value!.fetchData(filterData.value, { resetPage: true });
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

const handleSortChange: PrimaryTableProps['onSortChange'] = (orderBy: SortInfo) => {
  handleTableSortChange({
    orderBy,
    filterData,
    sortData,
    allowSortField,
  });
  getList();
};

const handleRequestDone = () => {
  // 回显选中数据
  if (selectionHook.selectionsRowKeys.value?.length && tableData.value?.length) {
    const selectionTable = tableData.value.filter(item => isAllowCheckSelection(item));
    const checkedIds = tableData.value
      .filter(item => selectionHook.selectionsRowKeys.value.includes(item.id))
      .map(check => check.id);
    selectionHook.isAllSelection.value = checkedIds.length > 0 && checkedIds.length === selectionTable.length;
    tableData.value.forEach((item) => {
      if (isAllowCheckSelection(item)) {
        item.isCustomCheck = checkedIds.includes(item.id);
      }
    });
  } else {
    selectionHook.isAllSelection.value = false;
  }
};

function handleSearch() {
  const data: Record<string, any> = {};
  searchParams.value.forEach((option) => {
    if (option.values) {
      data[option.id] = option.values[0]?.id;
    }
  });
  filterData.value = data;
  getList();
};

const handleClearFilter = () => {
  filterData.value = {};
  sortData.value = {};
  searchParams.value = [];
};

const handleSync = async () => {
  try {
    isSyncLoading.value = true;
    await postGatewaySyncData({ data: filterData.value });
    getList();
    Message({
      theme: 'success',
      message: t('ETCD 资源已同步到列表'),
    });
  } catch (e) {} finally {
    isSyncLoading.value = false;
  }
};

const handleCheck = (row: IGatewaySyncDataDto) => {
  resource.value = row;
  source.value = JSON.stringify(row.config);
  resourceType.value = row.resource_type;
  isResourceViewerShow.value = true;
};

const filterTimeKeys = (config: Record<string, any>) => {
  const _config = cloneDeep(config);
  ['created_at', 'updated_at', 'update_time', 'create_time'].forEach((key) => {
    if (key in _config) {
      delete _config[key];
    }
  });
  return _config;
};

const handleCheckDiff = async (row: IGatewaySyncDataDto) => {
  const res = await getResourceDiff({ type: row.resource_type, id: row.id });
  const config = res || {
    editor_config: {},
    etcd_config: {},
  };
  diffSliderConfig.value.before_config = filterTimeKeys(config.etcd_config || {});
  diffSliderConfig.value.after_config = filterTimeKeys(config.editor_config || {});
  diffSliderConfig.value.visible = true;
};

const addResource = async (ids?: string[]) => {
  try {
    const response = await addResourceToEditArea(ids ? { data: { resource_id_list: ids } } : {});
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const validEntries = Object.entries(response).filter(([_, value]) => value);
    getList();
    handleResetSelection();

    InfoBox({
      type: 'success',
      title: t('资源添加成功'),
      confirmText: t('完成'),
      content: h(
        'div',
        {
          class: 'info-box-content',
        },
        [
          h(
            'div', { class: ['normal-text'] },
            [
              h('span', { class: ['normal-text'] }, t('已同步')),

              validEntries.map(([key, value], index) => {
                return h(
                  'span',
                  [
                    h('span', { style: { margin: '0 4px' } }, value as number),
                    t('个'),
                    h('span', { class: ['active-text'], onClick: () => {
                      const routeData = router.resolve({
                        name: key,
                        params: { id: common.gatewayId },
                      });
                      window.open(routeData.href, '_blank');
                    } }, common.enums?.resource_type[key]),
                    t('资源'),
                    index !== validEntries.length - 1 ? t('，') : null,
                  ].filter(Boolean),
                );
              }),
            ],
          ),
          h('div', { class: ['normal-text'] }, t('可到对应编辑区进行维护')),
        ],
      ),
    });
  } catch (e) {}
};

const handleAdd = (row: IGatewaySyncDataDto) => {
  addResource([row.id]);
};

const handleBatchAdd = () => {
  addResource();
};

const handleMultiAdd = async () => {
  if (selectionHook?.selections.value.every(item => item.status === 'success')) {
    Message({
      theme: 'info',
      message: t('所有资源均已添加'),
    });
    return;
  }

  const ids = selectionHook?.selections.value.map((item: IGatewaySyncDataDto) => item.id);
  await addResource(ids);
  handleResetSelection();
};


const handleResetSelection = () => {
  selectionHook.isAllSelection.value = false;
  selectionHook.resetSelections();
  tableData.value.forEach((item) => {
    item.isCustomCheck = false;
  });
};
</script>

<style lang="scss" scoped>
.page-content-wrapper {
  min-height: calc(100vh - 157px);
  padding: 24px;

  .table-wrapper {
    background-color: #ffffff;

    .table-top-total {
      width: 100%;
      height: 42px;
      line-height: 42px;
      background: #EAEBF0;
      padding-left: 16px;
      display: flex;
      align-items: center;
      span {
        color: #4D4F56;
        font-size: 14px;
        &.equal {
         color: #2CAF5E;
         font-weight: Bold;
         margin-right: 4px;
        }
        &.lose {
          color: #EA3636;
          font-weight: Bold;
          margin-right: 4px;
        }
        &.add-opt {
          color: #3A84FF;
          cursor: pointer;
          margin-left: 8px;
        }
        &.line {
          width: 1px;
          height: 16px;
          background: #D8D8D8;
          margin: 0 12px;
        }
      }
    }
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
    .icon {
      margin-right: 4px;
      font-size: 16px;
    }
  }
}

.table-resource-search {
  width: 450px;
  background-color: #ffffff;
}

:deep(.table-cell-actions) {
  display: flex;
  gap: 12px;
}
</style>

<style lang="scss">
.info-box-content {
  background: #F5F7FA;
  border-radius: 2px;
  padding: 12px 16px;
  text-align: left;
  .normal-text {
    color: #4D4F56;
    font-size: 14px;
  }
  .active-text {
    color: #3A84FF;
    cursor: pointer;
  }
}
</style>
