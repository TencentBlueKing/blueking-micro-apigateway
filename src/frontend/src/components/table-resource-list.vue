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
    <div class="page-content-wrapper">
      <div class="header-actions">
        <div class="left">
          <BkButton
            :disabled="isReadonlyGateway"
            style="width: 142px;"
            theme="primary"
            @click="handleCreateClick"
          >
            {{ t('新建') }}
          </BkButton>
          <BkDropdown
            trigger="click"
            ref="dropdownRef"
            @show="handleOpenDropdown"
            @hide="isOpenDropdown = false"
          >
            <BkButton
              v-if="selectionColumns.length"
              :disabled="disabledOperate"
            >
              {{ t('批量操作') }}
              <DownSmall
                :class="[
                  'micro-apigateway-select-icon',
                  { 'is-open': isOpenDropdown }
                ]"
              />
            </BkButton>
            <template #content>
              <BkDropdownMenu>
                <BkDropdownItem
                  v-for="item in dropdownList"
                  :key="item.value"
                  @click.stop="handleDropdownClick(item)"
                >
                  <BkPopover
                    :popover-delay="0"
                    :content="item.tooltip"
                    :disabled="!item.disabled"
                  >
                    <BkButton
                      :theme="item.theme"
                      :disabled="item.disabled"
                      text
                    >
                      {{ item.label }}
                    </BkButton>
                  </BkPopover>
                </BkDropdownItem>
              </BkDropdownMenu>
            </template>
          </BkDropdown>
        </div>
        <div class="right">
          <slot name="extraSearch" />
          <BkSearchSelect
            v-model="searchParams"
            v-click-outside="handleSearchOutside"
            :data="localSearchOptions"
            :placeholder="t('搜索 {options}', { options: localSearchOptions.map(option => option.name).join(', ') })"
            clearable
            class="table-resource-search"
            unique-select
            @search="handleSearch"
            @keyup.enter="handleSearch"
            @click.stop="handleSearchSelectClick"
          />
        </div>
      </div>
      <div class="table-wrapper">
        <MicroAgTable
          ref="tableRef"
          v-model:selected-row-keys="selectionsRowKeys"
          v-model:table-data="tableData"
          v-model:settings="settings"
          v-bind="tableProps"
          :resizable="true"
          :row-key="tableRowKey"
          :sort="sortData"
          :filter-value="filterData"
          :api-method="getTableData"
          :columns="tableColumn"
          :is-show-first-full-row="selections.length > 0"
          :table-layout="tableLayout"
          @request-done="handleRequestDone"
          @select-change="handleSelectionChange"
          @clear-filter="handleClearFilter"
          @filter-change="handleFilterChange"
          @sort-change="handleSortChange"
        >
          <template #firstFullRow>
            <div class="table-first-full-row">
              <span class="normal-text">
                <span>{{ t('已选') }}</span>
                <span class="count">{{ selections.length }}</span>
                <span>{{ t('条') }}</span>
                <span class="mr4">,</span>
              </span>
              <span class="hight-light-text" @click="handleResetSelection">
                {{ t('清除选择') }}
              </span>
            </div>
          </template>
        </MicroAgTable>
      </div>
    </div>
    <SliderPublishDiff
      v-model="publishSliderConfig.visible"
      :after-config="publishSliderConfig.after_config"
      :before-config="publishSliderConfig.before_config"
      :show-footer="publishSliderConfig.footer"
      :source-info="publishSliderConfig.info"
      @done="handlePublishDone"
    />
    <DialogPublishResource
      v-model="isDialogPublishResourceShow"
      :diff-group-list="diffGroupList"
      :diff-group-list-back="diffGroupListBack"
      :delete-api="deleteApi"
      :table-row-key="tableRowKey"
      @confirm="handleMultiPublish"
      @refresh="handleDiffGroupListRefresh"
      @del="handleDelDone"
      @closed="handlePublishClosed"
    />
  </div>
</template>

<script lang="tsx" setup>
import { computed, ref, watch, shallowRef, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { cloneDeep } from 'lodash-es';
import { IQueryListParams, IDropList } from '@/types';
import { Message, InfoBox, Checkbox, Dropdown } from 'bkui-vue';
import { DownSmall } from 'bkui-vue/lib/icon';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import {
  type PrimaryTableProps,
  type TableRowData,
  type SortInfo,
  type FilterValue,
} from '@blueking/tdesign-ui';
import { useCommon } from '@/store';
import { STATUS_CN_MAP } from '@/enum';
import { useTDesignSelection } from '@/hooks/use-tdesign-selection';
import { useSearchSelectPopoverHidden } from '@/hooks/use-search-select-popover-hidden';
import { getLabels } from '@/http/labels';
import { deleteResource } from '@/http/delete';
import { revertResource } from '@/http/revert';
import { getDiffByType, getResourceDiff, IDiffGroup, publish } from '@/http/publish';
import dayjs from 'dayjs';
import i18n from '@/i18n';
import useTsxRouter from '@/hooks/use-tsx-router';
import { useTableFilterChange } from '@/hooks/use-table-filter-change';
import { useTableSortChange } from '@/hooks/use-table-sort-change';
import TagStatus from '@/components/tag-status.vue';
import DialogPublishResource from '@/components/dialog-publish-resource.vue';
import TagLabel from '@/components/tag-label.vue';
import RenderDropdown from '@/components/render-dropdown.vue';
import SliderPublishDiff from '@/components/slider-publish-diff.vue';
import MicroAgTable from '@/components/micro-ag-table/table.vue';

interface IProps {
  queryListParams?: IQueryListParams;
  tableProps?: Partial<PrimaryTableProps>;
  routes?: {
    create?: string,
    edit?: string,
    clone?: string,
  };
  columns?: PrimaryTableProps['columns'],
  selectionColumn?: PrimaryTableProps['columns'];
  actionColumn?: PrimaryTableProps['columns'];
  excludeColumns?: string[];
  extraSearchOptions?: ISearchItem[];
  extraSearchParams?: ISearchParam[];
  actionColumnConfig?: {
    edit?: boolean,
    delete?: boolean,
    code?: boolean,
    clone?: boolean,
  };
  resourceType?: string;
  tableRowKey?: string;
  tableLayout?: string;
  nameColKey?: string;
  isDisabledCheckSelection?: (...args: TableRowData) => void;
  deleteApi?: (...args: any[]) => Promise<unknown>;
}

export interface ISearchParam {
  id: string,
  name: string,
  values?: { id: string, name: string }[]
}

const {
  routes = {},
  queryListParams = {},
  tableProps = {},
  columns = [
    {
      title: 'ID',
      colKey: 'id',
      ellipsis: true,
    },
    {
      title: i18n.global.t('路径'),
      colKey: 'uris',
      cell: (h, { row }: TableRowData) => row.config?.uris?.join(',') || '--',
    },
    {
      title: i18n.global.t('描述'),
      colKey: 'desc',
      ellipsis: true,
      cell: (h, { row }: TableRowData) => row.config?.desc || '--',
    },
  ],
  selectionColumn = [],
  actionColumn = [],
  excludeColumns = [],
  extraSearchOptions = [],
  extraSearchParams = [],
  resourceType,
  // 表格唯一标识
  tableRowKey = 'id',
  // 表格布局方式
  tableLayout = 'fixed',
  // 表格sort列字段模块之间会存在不同key
  nameColKey = 'name',
  // 禁止勾选复选框的条件
  isDisabledCheckSelection = (_value: TableRowData) => {
    return false;
  },
  deleteApi,
} = defineProps<IProps>();

const emit = defineEmits(['check-resource', 'clear-filter']);

// const router = useRouter();
const { useRouter } = useTsxRouter();
const router = useRouter();
const { t } = useI18n();
const common = useCommon();
const { handleTableFilterChange } = useTableFilterChange();
const { handleTableSortChange } = useTableSortChange();
const { handleSearchOutside, handleSearchSelectClick } = useSearchSelectPopoverHidden();
const {
  isAllSelection,
  selections,
  selectionsRowKeys,
  resetSelections,
  handleSelectionChange,
  handleCustomSelectChange,
  handleCustomSelectAllChange,
} = useTDesignSelection();

const settings = shallowRef({
  size: 'small',
  checked: [],
  disabled: [],
});

const allowSortField = shallowRef(['name', 'username', 'updated_at']);

const isOpenDropdown = shallowRef(false);

const tableRef = ref<InstanceType<typeof MicroAgTable>>(null);

const tagLabelRef = ref<InstanceType<typeof TagLabel>>(null);

const dropdownRef = ref<InstanceType<typeof Dropdown>>(null);

const filterData = ref<FilterValue>({});

const sortData =  ref<SortInfo>({});

const tableData = ref([]);

const labelList = ref<string[]>([]);

// 设置表格半选效果
const setIndeterminate = computed(() => {
  const isExistCheck = tableData.value.some(item => item.isCustomCheck);
  return isExistCheck && selections.value.length > 0 && !isAllSelection.value;
});

// 这里采用自定义checkbox是为了后续功能扩展，用自带的无法自定义渲染函数(暂时支持跨页选择，不支持跨页全选)
const selectionColumns = shallowRef(selectionColumn?.length ? selectionColumn : [{
  colKey: 'row-select',
  type: 'custom-checkbox',
  align: 'center',
  fixed: 'left',
  width: 60,
  title: () => {
    return (
       <Checkbox
          v-model={isAllSelection.value}
          v-bk-tooltips={{
            content: t('当前网关为只读'),
            disabled: !isReadonlyGateway.value,
          }}
          disabled={disabledSelected.value}
          indeterminate={setIndeterminate.value}
          onChange={() => {
            tableData.value.forEach((item) => {
              if (!isDisabledCheckSelection?.(item)) {
                item.isCustomCheck = isAllSelection.value;
              }
            });
            const tables = tableData.value.filter(item => !isDisabledCheckSelection?.(item));
            handleCustomSelectAllChange({ isCheck: isAllSelection.value, tableRowKey, tables });
          }}
        />
    );
  },
  cell: (h, { row }) => {
    return (
      <Checkbox
        v-model={row.isCustomCheck}
        v-bk-tooltips={{
          content: t('当前网关为只读'),
          disabled: !isReadonlyGateway.value,
        }}
        disabled={disabledSelected.value}
        onChange={(isCheck: boolean) => {
          // 这里可以增加disabled逻辑
          handleCustomSelectChange({ isCheck, tableRowKey, row });
          const selectionTable = tableData.value.filter(item => !isDisabledCheckSelection?.(item));
          const checkedIds = tableData.value
            .filter(item => selectionsRowKeys.value.includes(item[tableRowKey]))
            .map(check => check[tableRowKey]);
          isAllSelection.value = checkedIds.length > 0 && checkedIds.length === selectionTable.length;
          tableData.value.forEach((item) => {
            if (!isDisabledCheckSelection?.(item) && row[tableRowKey] === item[tableRowKey]) {
              item.isCustomCheck = row.isCustomCheck;
            }
          });
        }}
      />
    );
  },
}]);

const searchParams = ref<ISearchParam[]>([]);

const dateKey = ref<number>(+new Date());

const statusFilter = computed(() => {
  const list = Object.entries(STATUS_CN_MAP)
    .map(([key, value]) => ({
      label: value,
      value: key,
    }));

  return list;
});

const labelFilter = computed(() => {
  const results = labelList.value.map((label) => {
    return {
      name: label,
      id: label,
    };
  });
  return results;
});

const searchOptions = ref<ISearchItem[]>([
  {
    id: 'name',
    name: t('名称'),
  },
  {
    id: 'id',
    name: 'ID',
  },
  {
    id: 'updater',
    name: t('更新人'),
  },
  {
    id: 'status',
    name: t('状态'),
    multiple: true,
    children: Object.entries(STATUS_CN_MAP)
      .map(([key, value]) => ({
        id: key,
        name: value,
      })),
  },
]);

const searchOptionsPluginCustom = ref<ISearchItem[]>([
  {
    id: 'name',
    name: t('名称'),
  },
  {
    id: 'updater',
    name: t('更新人'),
  },
]);

const commonColumns = ref<PrimaryTableProps['columns']>([
  {
    __name__: 'label',
    title: t('标签'),
    colKey: 'label',
    width: 'auto',
    ellipsis: false,
    cellStyle: {
      whiteSpace: 'normal',
      wordBreak: 'break-all',
    },
    filter: {
      type: 'multiple',
      showConfirmAndReset: true,
      list: [],
    },
    cell: (h, { row }) => {
      const isExistLabel = Object.keys(row?.config?.labels ?? {}).length > 0;
      if (isExistLabel) {
        return (
          <TagLabel
            ref={tagLabelRef}
            labels={row.config.labels}
         />
        );
      }
      return '--';
    },
  },
  {
    __name__: 'updated_at',
    title: t('更新时间'),
    colKey: 'updated_at',
    ellipsis: true,
    sorter: true,
    cell: (h, { row }) => {
      return dayjs.unix(row.updated_at as number)
        .format('YYYY-MM-DD HH:mm:ss Z');
    },
  },
  {
    __name__: 'updater',
    title: t('更新人'),
    colKey: 'updater',
    ellipsis: true,
    width: 80,
    cell: (h, { row }) => <span>{row.updater || row.update_by || '--'}</span>,
  },
  {
    __name__: 'status',
    title: t('状态'),
    colKey: 'status',
    ellipsis: true,
    filter: {
      type: 'multiple',
      showConfirmAndReset: true,
      list: statusFilter.value,
    },
    width: 110,
    cell: (h, { row }) => <TagStatus status={row.status as string} />,
  },
  {
    __name__: 'actions',
    title: t('操作'),
    colKey: 'opt',
    fixed: 'right',
    width: 160,
    cell: (h, { row }) => {
      return (
        <div class="table-cell-actions">
          <bk-popover
            popoverDelay={0}
            content={getDisabledToolTip(row.status, 'edit')}
            disabled={!isReadonlyGateway.value && !['delete_draft'].includes(row.status)}
          >
            <bk-button
              text={true}
              theme="primary"
              disabled={isReadonlyGateway.value || ['delete_draft'].includes(row.status)}
              onClick={() => handleEditClick(row)}
            >
              {t('编辑')}
            </bk-button>
          </bk-popover>
          {
            resourceType !== 'plugin_custom'
              ? (
                <div class="table-cell-actions">
                  <bk-popover
                    popoverDelay={0}
                    content={getDisabledToolTip(row.status, 'diff')}
                    disabled={!isReadonlyGateway.value && !['success'].includes(row.status)}
                  >
                    <bk-button
                      text={true}
                      disabled={isReadonlyGateway.value || ['success'].includes(row.status)}
                      theme="primary"
                      onClick={() => handleDiffClick(row)}
                    >
                      {t('差异')}
                    </bk-button>
                  </bk-popover>
                   <bk-popover
                    popoverDelay={0}
                    content={getDisabledToolTip(row.status, 'publish')}
                    disabled={!isReadonlyGateway.value && !['success', 'conflict'].includes(row.status)}
                  >
                    <bk-button
                      text={true}
                      theme="primary"
                      disabled={isReadonlyGateway.value || ['success', 'conflict'].includes(row.status)}
                      onClick={() => {
                        handlePublishClick(row);
                      }}
                    >
                      {t('发布')}
                    </bk-button>
                  </bk-popover>
                  <RenderDropdown
                    row={row}
                    clone={!!routes.clone}
                    onClone={() => handleCloneClick(row)}
                    onRevert={() => handleRevertClick(row)}
                    onDelete={() => handleDeleteClick(row)}
                  />
                </div>
              )
              : (
                <bk-pop-confirm
                  content={t('删除操作无法撤回，请谨慎操作！')}
                  title={t('确认删除？')}
                  trigger="click"
                  width="288"
                  onConfirm={() => handleDeleteClick(row)}
                >
                  <bk-button
                    class="dropdown-item-btn"
                    text
                    theme="danger"
                    disabled={isReadonlyGateway.value}
                  >
                    {t('删除')}
                  </bk-button>
                </bk-pop-confirm>
              )
          }
        </div>
      );
    },
  },
]);

const tableColumn = ref<PrimaryTableProps['columns']>([
  ...selectionColumns.value,
  {
    title: t('名称'),
    colKey: nameColKey,
    sorter: true,
    fixed: 'left',
    cell: (h, { row }) => {
      const isNoPlugin =  !['plugin_custom'].includes(resourceType);
      return (
        <div class="flex-row align-items-center">
          <div
            class="single-ellipse"
            class={[{ 'name-active': isNoPlugin }]}
            v-bk-tooltips={{
              content: row.name,
              placement: 'top',
              disabled: !row.isOverflow,
            }}
            onClick={() => handleCheckCodeClick(row, isNoPlugin)}
           >
            {row.name}
          </div>
        </div>
      );
    },
  },
  ...columns,
  ...commonColumns.value.filter(column => !excludeColumns.includes(column.colKey)),
  ...actionColumn,
]);

const publishSliderConfig = ref({
  visible: false,
  after_config: {},
  before_config: {},
  operationType: '',
  info: {
    type: '',
    name: '--',
    id: '',
  },
  footer: true,
});

const isDialogPublishResourceShow = ref(false);
const diffGroupList = ref<IDiffGroup[]>([]);
const diffGroupListBack = ref<IDiffGroup[]>([]);
const singlePublishId = ref<string>('');
const publishDelIds = ref<string[]>([]);

const localSearchOptions = computed<ISearchItem[]>(() => {
  if (resourceType === 'plugin_custom') {
    return [
      ...searchOptionsPluginCustom.value,
    ];
  }

  return excludeColumns.includes('label') ? [
    ...searchOptions.value,
    ...extraSearchOptions,
  ] : [
    ...searchOptions.value,
    ...extraSearchOptions,
    {
      id: 'label',
      name: t('标签'),
      multiple: true,
      children: labelFilter.value,
    },
  ];
});

const isReadonlyGateway = computed(() => common.curGatewayData?.read_only);

const disabledSelected = computed(() => {
  return !tableData.value?.length || isReadonlyGateway.value;
});

const disabledOperate = computed(() => {
  return !selections.value?.length || isReadonlyGateway.value;
});

const dropdownList = shallowRef<IDropList>([
  {
    label: t('发布'),
    value: 'publish',
    theme: '',
    disabled: true,
    method: handleMultiPublishClick,
  },
  {
    label: t('撤销'),
    value: 'revert',
    theme: '',
    disabled: true,
    method: handleMultiRevertClick,
  },
  {
    label: t('删除'),
    value: 'delete',
    theme: 'danger',
    disabled: true,
    method: handleMultiDeleteClick,
  },
]);

watch(searchParams, (newParams) => {
  if (!newParams?.length) {
    filterData.value = {};
    handleSearch();
    return;
  }
  // 如果没有extraSearchParams选项，直接调用搜索接口
  if (!extraSearchParams?.length) {
    handleSearch();
    return;
  }
});

watch(() => extraSearchParams, () => {
  if (extraSearchParams.length) {
    searchParams.value = [...searchParams.value, ...extraSearchParams];
    handleSearch();
  }
}, { deep: true });

const getList = (params: Record<string, any> = {}) => {
  const queryParams = { ...params };
  Object.keys(filterData.value).forEach((key) => {
    queryParams[key] = Array.isArray(filterData.value[key]) ? filterData.value[key].join() : filterData.value[key];
  });
  tableRef.value!.fetchData(queryParams, { resetPage: true });
};

const getLabelData = async () => {
  if (!excludeColumns.includes('label')) {
    const res = await getLabels({ type: resourceType });
    const labelData = Object.keys(res ?? {});
    if (labelData?.length) {
      labelData.forEach((label) => {
        if (!labelList.value.includes(label)) {
          labelList.value.push(label);
        }
      });
      const labelCol = tableColumn.value.find(col => ['label'].includes(col.colKey));
      if (labelCol) {
        labelCol.filter.list  = labelFilter.value.map((label) => {
          return {
            label: label.name,
            value: label.id,
          };
        });
      }
    }
  }
};
getLabelData();

const getTableData = async (params: Record<string, any> = {}) => {
  const results = await queryListParams?.apiMethod({ gatewayId: common.gatewayId, query: params });
  return results ?? [];
};

// 自适应标签宽度
const getLabelWidth = () => {
  const labelCol = tableColumn.value.find(col => ['label'].includes(col.colKey));
  nextTick(() => {
    if (labelCol && tagLabelRef.value) {
      const labelWidth = tagLabelRef.value?.getResizeLabelWidth();
      labelCol.width = labelWidth;
    }
  });
};

// 回显选中数据
const getSelectionData = () => {
  if (selectionsRowKeys.value?.length && tableData.value?.length) {
    const selectionTable = tableData.value.filter(item => !isDisabledCheckSelection?.(item));
    const checkedIds = tableData.value
      .filter(item => selectionsRowKeys.value.includes(item[tableRowKey]))
      .map(check => check[tableRowKey]);
    isAllSelection.value = checkedIds.length > 0 && checkedIds.length === selectionTable.length;
    tableData.value.forEach((item) => {
      if (!isDisabledCheckSelection?.(item)) {
        item.isCustomCheck = checkedIds.includes(item[tableRowKey]);
      }
    });
  } else {
    isAllSelection.value = false;
  }
};

// 处理不同disabled的tooltip
const getDisabledToolTip = (status?: string, type?: string) => {
  if (isReadonlyGateway.value) {
    return t('当前网关为只读');
  }

  if (['success'].includes(status)) {
    if (['publish'].includes(type)) {
      return t('该资源已发布');
    }
    if (['diff'].includes(type)) {
      return t('该资源已发布, 无需进行差异对比');
    }
  }

  if (['conflict'].includes(status)) {
    if (['publish'].includes(type)) {
      return t('该资源存在冲突，不能发布');
    }
  }

  if (['delete_draft'].includes(status) && ['edit'].includes(type)) {
    return t('该资源删除待发布, 无法编辑');
  }

  return '';
};

// 处理表头筛选联动搜索框
const handleFilterChange: PrimaryTableProps['onFilterChange'] = (filterItem) => {
  handleTableFilterChange({
    filterItem,
    filterData,
    searchOptions: localSearchOptions,
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

const handleOpenDropdown = () => {
  isOpenDropdown.value = true;
  // 不可撤销
  const isNoRevert = selections.value.every(resource => !['update_draft', 'delete_draft'].includes(resource.status));
  // 不可发布
  const isNoPublish = selections.value.every(resource => ['success'].includes(resource.status));
  // 不可删除
  const isNoDelete = selections.value.every(resource => !['create_draft', 'success'].includes(resource.status));

  dropdownList.value.forEach((item) => {
    item.disabled = disabledOperate.value;
    if (['publish'].includes(item.value) && isNoPublish) {
      item.tooltip = t('选中资源均已发布');
      item.disabled = true;
    }
    if (['revert'].includes(item.value) && isNoRevert) {
      item.tooltip = t('选中资源均不是可撤销状态');
      item.disabled = true;
    }
    if (['delete'].includes(item.value) && isNoDelete) {
      item.tooltip = t('选中资源均不是可删除状态');
      item.disabled = true;
    }
  });
};

const handleDropdownClick = (row: IDropList) => {
  if (!row.disabled) {
    dropdownRef.value?.popoverRef?.hide();
    row.method?.();
  };
};

const handleRequestDone = () => {
  getLabelWidth();
  getSelectionData();
};

const handleClearFilter = () => {
  filterData.value = {};
  sortData.value = {};
  searchParams.value = [];
  emit('clear-filter');
};

const handleCreateClick = () => {
  const name = routes.create;
  if (name) {
    router.push({ name });
  }
};

const handleEditClick = (row: TableRowData) => {
  const name = routes.edit;
  if (name) {
    router.push({ name, params: { id: row[tableRowKey] } });
  }
};

const handleCloneClick = (row: TableRowData) => {
  const name = routes.clone;
  if (name) {
    router.push({ name, params: { id: row[tableRowKey] } });
  }
};

const handleDiffClick = async (row: Record<string, any>) => {
  const res = await getResourceDiff({ type: resourceType, id: row[tableRowKey] });
  const config = res || {
    editor_config: {},
    etcd_config: {},
  };

  publishSliderConfig.value.before_config = sortObjectKeys(filterTimeKeys(config.etcd_config || {}));
  publishSliderConfig.value.after_config = sortObjectKeys(filterTimeKeys(config.editor_config || {}));
  publishSliderConfig.value.info.name = row.name;
  publishSliderConfig.value.info.id = row[tableRowKey];
  publishSliderConfig.value.info.type = resourceType;
  publishSliderConfig.value.footer = false;
  publishSliderConfig.value.visible = true;
};

const handlePublishDone = async () => {
  await getList();
};

const handlePublishClick = async (row: Record<string, any>) => {
  if (row.status === 'success') {
    Message({
      theme: 'info',
      message: t('该资源已发布'),
    });
    return;
  }

  if (row.status === 'conflict') {
    Message({
      theme: 'warning',
      message: t('该资源存在冲突，不能发布'),
    });
    return;
  }

  if (!row[tableRowKey] || !resourceType) {
    Message({
      theme: 'error',
      message: t('获取资源ID或资源类型失败'),
    });
    return;
  }

  singlePublishId.value = row[tableRowKey];

  const res = await getDiffByType({
    type: resourceType,
    data: {
      resource_id_list: [singlePublishId.value],
    },
  });

  diffGroupList.value = res || [];
  diffGroupListBack.value = cloneDeep(res ?? []);
  isDialogPublishResourceShow.value = true;
};

const handleDelDone = async (id: string) => {
  publishDelIds.value.push(id);

  let resource_id_list: string[] = [];

  if (!selections.value.length) {
    resource_id_list = [singlePublishId.value];
  } else {
    resource_id_list = selections.value.filter(resource => !publishDelIds.value.includes(resource[tableRowKey]))
      .map(resource => resource[tableRowKey]);
  }

  const res = await getDiffByType({
    type: resourceType,
    data: {
      resource_id_list,
    },
  });

  diffGroupList.value = res || [];
  diffGroupListBack.value = cloneDeep(res ?? []);
};

const handleMultiPublish = async () => {
  let resource_id_list: string[] = [];

  if (!selections.value.length) {
    resource_id_list = [singlePublishId.value];
  } else {
    resource_id_list = selections.value.filter(resource => !['success', 'conflict'].includes(resource.status)
       && !publishDelIds.value.includes(resource[tableRowKey]))
      .map(resource => resource[tableRowKey]);
  }

  await publish({
    data: {
      resource_id_list,
      resource_type: resourceType,
    },
  });

  Message({
    theme: 'success',
    message: t('已发布'),
  });

  handleResetSelection();
  singlePublishId.value = '';
  publishDelIds.value = [];

  await getList();
  isDialogPublishResourceShow.value = false;
};

const handlePublishClosed = async () => {
  await getList();
  handleResetSelection();
  singlePublishId.value = '';
  publishDelIds.value = [];
  dateKey.value = +new Date();
};

const handleDiffGroupListRefresh = async (hideTips?: boolean) => {
  await getDiffGroupList();

  if (!diffGroupList.value.length) {
    handleResetSelection();
    singlePublishId.value = '';
    publishDelIds.value = [];
    await getList();
    isDialogPublishResourceShow.value = false;

    if (!hideTips) {
      Message({
        theme: 'info',
        message: t('没有需要发布的资源'),
      });
    }
  }
};

const getDiffGroupList = async () => {
  let resource_id_list: string[] = [];

  if (!selections.value.length) {
    resource_id_list = [singlePublishId.value];
  } else {
    resource_id_list = selections.value.filter(resource => resource.status !== 'success' && !publishDelIds.value.includes(resource[tableRowKey]))
      .map(resource => resource[tableRowKey]);
  }

  const res = await getDiffByType({
    type: resourceType,
    data: {
      resource_id_list,
    },
  });

  diffGroupList.value = res || [];
  diffGroupListBack.value = cloneDeep(res || []);
};

const handleRevertClick = async (row: Record<string, any>) => {
  const rowId = row[tableRowKey];
  if (!rowId || !resourceType) {
    Message({
      theme: 'error',
      message: t('获取资源ID或资源类型失败'),
    });
    return;
  }

  await revertResource({
    data: {
      resource_id_list: [rowId],
      resource_type: resourceType,
    },
  });

  Message({
    theme: 'success',
    message: t('已撤销'),
  });

  await getList();
};

const handleDeleteClick = async (row: TableRowData) => {
  if (deleteApi) {
    await deleteApi({ id: row[tableRowKey] } as Record<string, string>);

    Message({
      theme: 'success',
      message: t('已删除'),
    });

    await getList();
  }
};

const handleCheckCodeClick = (row: TableRowData, isNoPlugin: boolean) => {
  if (isNoPlugin) {
    emit('check-resource', { resource: row });
  }
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

const sortObjectKeys = (obj: Record<string, any>) => Object.keys(obj)
  .sort()
  .reduce((result, key) => {
    result[key] = obj[key];
    return result;
  }, {} as Record<string, any>);

const handleSearch = () => {
  const data: Record<string, any> = {};
  searchParams.value.forEach((option) => {
    if (option.values) {
      if (['label', 'status', 'path'].includes(option.id)) {
        data[option.id] = option.values.map(label => label.id);
      } else {
        data[option.id] = option.values[0]?.id;
      }
    }
  });
  filterData.value = data;
  getList();
};

async function handleMultiPublishClick()  {
  if (!selections.value.length) {
    Message({
      theme: 'warning',
      message: t('请选择要发布的资源'),
    });
    return;
  }

  await getDiffGroupList();
  isDialogPublishResourceShow.value = true;
};

async function handleMultiRevertClick() {
  if (!selections.value.length) {
    Message({
      theme: 'warning',
      message: t('请选择要撤销的资源'),
    });
    return;
  }

  InfoBox({
    title: t('确认撤销？'),
    confirmText: t('撤销'),
    cancelText: t('取消'),
    onConfirm: async () => {
      await revertResource({
        data: {
          resource_id_list: selections.value.filter(resource => ['update_draft', 'delete_draft'].includes(resource.status))
            .map(resource => resource[tableRowKey]),
          resource_type: resourceType,
        },
      });

      Message({
        theme: 'success',
        message: t('已撤销'),
      });

      handleResetSelection();
      await getList();
    },
  });
}

async function handleMultiDeleteClick() {
  if (!selections.value.length) {
    Message({
      theme: 'warning',
      message: t('请选择要删除的资源'),
    });
    return;
  }

  InfoBox({
    type: 'warning',
    title: t('确认删除？'),
    confirmText: t('删除'),
    cancelText: t('取消'),
    onConfirm: async () => {
      await deleteResource({
        data: {
          resource_id_list: selections.value.filter(resource => ['create_draft', 'success'].includes(resource.status))
            .map(resource => resource[tableRowKey]),
          resource_type: resourceType,
        },
      });

      Message({
        theme: 'success',
        message: t('已删除'),
      });

      handleResetSelection();
      await getList();
    },
  });
}

const handleResetSelection = () => {
  isAllSelection.value = false;
  tableData.value.forEach((item) => {
    item.isCustomCheck = false;
  });
  resetSelections();
};

defineExpose({
  getList,
});

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

  .right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.micro-apigateway-select-icon {
  font-size: 20px !important;
  transition: transform .5s;

  &.is-open {
    transform: rotate(180deg) !important;
  }
}

:deep(.table-cell-actions) {
  display: flex;
  gap: 12px;
}

:deep(.name-active) {
  cursor: pointer;
  color: #3a84ff;
}

</style>
