<template>
  <!--使用 PrimaryTable满足大部分需求。涉及到非常复杂的需求后，
    比如要实现表格某列是树形结构等功能。替换成EnhancedTable， 它包含 BaseTable 和 PrimaryTable 的所有功能
  -->
  <PrimaryTable
    ref="primaryTableRef"
    v-model:selected-row-keys="selectedRowKeys"
    :class="[
      'primary-table-wrapper',
      {
        'primary-table-no-data': !localTableData.length
      }
    ]"
    :data="localTableData"
    :columns="columns"
    :pagination="pagination"
    :loading="loading"
    :filter-row="null"
    :bk-ui-settings="tableSetting"
    @bk-ui-settings-change="handleSettingChange"
    @row-mouseenter="handleRowEnter"
    @row-mouseleave="handleRowLeave"
    @page-change="handlePageChange"
    v-bind="$attrs"
  >
    <template
      v-if="slots.firstFullRow"
      #firstFullRow="slotProps"
    >
      <slot
        v-if="isShowFirstFullRow && localTableData.length > 0"
        name="firstFullRow"
        v-bind="slotProps"
      />
    </template>
    <template
      v-if="slots.expandedRow"
      #expandedRow="slotProps"
    >
      <slot
        name="expandedRow"
        v-bind="slotProps"
      />
    </template>
    <template #loading>
      <BkLoading :loading="loading" />
    </template>
    <template #empty>
      <TableEmpty
        :error="error"
        :empty-type="tableEmptyType"
        :query-list-params="params"
        @clear-filter="handlerClearFilter"
        @refresh="handleRefresh"
      />
    </template>
  </PrimaryTable>
</template>

<script setup lang="ts">
import {
  ref,
  unref,
  useTemplateRef,
  computed,
  watch,
  onMounted,
  onBeforeUnmount,
  useSlots,
  type ShallowRef,
} from 'vue';
import {
  PrimaryTable,
  type PrimaryTableProps,
  type TableRowData,
  type PrimaryTableInstance,
} from '@blueking/tdesign-ui';
import { ITableSettings } from '@/types';
import { useRequest } from 'vue-request';
import { cloneDeep } from 'lodash-es';
import useTableChangeSetting from '@/hooks/use-table-change-setting';
import TableEmpty from './empty-exception';

interface IProps {
  apiMethod?: (params?: Record<string, any>) => Promise<unknown>
  columns?: PrimaryTableProps['columns']
  immediate?: boolean
  localPage?: boolean
  isShowFirstFullRow?: boolean
  tableEmptyType?: 'empty' | 'search-empty'
}

const selectedRowKeys = defineModel<any[]>('selectedRowKeys', { default: () => [] });

const tableData = defineModel<any[]>('tableData', { default: () => [] });

const tableSetting = defineModel<null | ShallowRef<ITableSettings>>('settings', { default: () => null });

const isAllSelection = defineModel<boolean>('isAllSelection', { default: () => false });

const {
  apiMethod = undefined,
  columns = [],
  // 是否首次加载
  immediate = true,
  // 是否需要本地分页
  localPage = false,
  // 是否显示自定义首行内容
  isShowFirstFullRow = false,
  // 本地筛选查询状态
  tableEmptyType = 'empty',
} = defineProps<IProps>();

const emit = defineEmits<{
  'row-mouseenter': { e?: MouseEvent, row?: TableRowData }
  'row-mouseleave': { e?: MouseEvent, row?: TableRowData }
  'request-done': void
  'clear-filter': void
  'refresh': void
}>();

const slots = useSlots();

const TDesignTableRef = useTemplateRef<PrimaryTableInstance & ITableMethod>('primaryTableRef');

const filterEl = ref<HTMLElement | null>(null);

const radioEl = ref<HTMLElement | null>(null);

let radioClickHandler: ((e: Event) => void) | null = null;

const localTableData = ref<any[]>([]);

const pagination = ref<PrimaryTableProps['pagination']>({
  current: 1,
  pageSize: 10,
  defaultCurrent: 1,
  defaultPageSize: 10,
  total: 0,
  theme: 'default',
  showPageSize: true,
});

const { changeTableSetting, isDiffSize } = useTableChangeSetting(tableSetting.value);

let paramsMemo: Record<string, any> = {};

const offsetAndLimit = computed(() => {
  return {
    offset: pagination.value!.pageSize! * (pagination.value!.current! - 1) || 0,
    limit: pagination.value!.pageSize || 10,
  };
});

/**
 * 请求表格数据
 * @param {Object} params 请求数据
 * @param {Boolean} loading 加载状态
 * @param {Object | Null} error 错误信息
 * @param run 手动触发请求的函数
 */
const { params, loading, error, refresh, run } = useRequest(apiMethod, {
  manual: true,
  // 是否立即执行请求
  immediate,
  defaultParams: [offsetAndLimit.value],
  onSuccess: (response: {
    results: any[]
    count: number
  }) => {
    const results = response?.results ?? [];
    paramsMemo = { ...params.value?.[0] };
    pagination.value!.total = response?.count ?? 0;
    tableData.value = [...results];
    // 处理接口调用成功后抛出事件，为每个页面提供单独业务处理
    emit('request-done');
  },
  onError: (error) => {
    tableData.value = [];
    pagination.value!.total = 0;
    isAllSelection.value = false;
    console.error(error);
  },
});

watch(tableData, () => {
  localTableData.value = cloneDeep(tableData.value || []);
  if (localPage) {
    pagination.value = Object.assign(pagination.value, {
      current: 1,
      total: localTableData.value.length,
    });
  }
}, {
  immediate: true,
  deep: true,
});

const fetchData = (
  params: Record<string, any> = {},
  options: {
    resetPage?: boolean
  } = {
    resetPage: false,
  },
) => {
  if (options.resetPage) {
    pagination.value!.current = 1;
  }
  run({
    ...params,
    ...offsetAndLimit.value,
  });
};

const handleSettingChange = (setting: ITableSettings) => {
  tableSetting.value = { ...setting };
  const isExistDiff = isDiffSize(setting);
  changeTableSetting(setting);
  if (!isExistDiff) {
    // 这里处理高级设置事件回调后需要处理的业务
    return;
  }
};

const handleRowEnter = ({ e, row }: { e: MouseEvent, row: TableRowData }) => {
  const truncateNode = e.target?.querySelector('.single-ellipse');
  if (truncateNode) {
    row.isOverflow = truncateNode?.scrollWidth > truncateNode.clientWidth;
  }
  emit('row-mouseenter', { e, row });
};

const handleRowLeave = ({ e, row }: { e: MouseEvent, row: TableRowData }) => {
  delete row.isOverflow;
  emit('row-mouseleave', { e, row });
};

const handlePageChange = ({ current, pageSize }: {
  current: number
  pageSize: number
}) => {
  pagination.value!.current = current;
  pagination.value!.pageSize = pageSize;
  if (!localPage) {
    fetchData({
      ...paramsMemo,
      ...offsetAndLimit.value,
    });
  }
};

// 处理自定义重置功能和点击单选直接关闭弹框
const handleRadioFilterClick = () => {
  setTimeout(() => {
    const filterPopup = document.querySelector('.t-table__filter-pop-content');
    radioEl.value = filterPopup?.querySelector('.t-radio-group');
    if (radioEl.value) {
      const confirmBtn = document.querySelector('.t-table__filter--bottom-buttons > .t-button--theme-primary');
      radioClickHandler = (event: MouseEvent) => {
        const radioLabel = event.target.closest('label.t-radio');
        const radioInput = radioLabel.querySelector('input.t-radio__former');
        if (radioInput.checked) {
          confirmBtn.click();
        }
      };
      radioEl.value.addEventListener('click', radioClickHandler);
    }
  }, 0);
};

const handleListenerRadio = () => {
  const table = unref(TDesignTableRef);
  if (!table) return;

  // 获取表头filter筛选框容器元素
  filterEl.value = table.$el.querySelector('.t-table__filter-icon-wrap');
  if (!filterEl.value) {
    return;
  }
  document.addEventListener('click', handleRadioFilterClick);
};

const getPagination = () => {
  return pagination.value;
};

const setPagination = ({ current, pageSize }: {
  current: number
  pageSize: number
}) => {
  handlePageChange({
    current,
    pageSize,
  });
};

const setPaginationTheme = ({ theme, showPageSize }: {
  theme: 'default' | 'simple'
  showPageSize?: boolean
}) => {
  Object.assign(pagination.value!, {
    theme,
    showPageSize: showPageSize ?? true,
  });
};

const resetPaginationTheme = () => {
  pagination.value!.theme = 'default';
  pagination.value!.showPageSize = true;
};

// 清空过滤条件
const handlerClearFilter = () => {
  emit('clear-filter');
};

// 异常刷新
const handleRefresh = () => {
  refresh();
  emit('refresh');
};

onMounted(() => {
  if (immediate && !localPage) {
    fetchData({ ...offsetAndLimit.value });
  }
  handleListenerRadio();
});

onBeforeUnmount(() => {
  document.removeEventListener('click', handleRadioFilterClick);
  radioEl.value?.removeEventListener('click', radioClickHandler);
  filterEl.value = null;
  radioEl.value = null;
  radioClickHandler = null;
});

defineExpose({
  TDesignTableRef,
  loading,
  fetchData,
  getPagination,
  setPagination,
  setPaginationTheme,
  resetPaginationTheme,
  refresh,
});

</script>

<style lang="scss">
.primary-table-wrapper {
  font-size: 12px;

  .single-ellipse {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    &.is-color-active {
      color: #3a84ff;
      cursor: pointer;
    }
  }

  .table-first-full-row {
    width: 100%;
    height: 32px;
    line-height: 32px;
    background-color: #f0f1f5;
    text-align: center;
    font-size: 12px;

    .normal-text {
      color: #4d4f56;

      .count {
        font-weight: 700;
      }
    }

    .hight-light-text {
      color: #3a84ff;
      cursor: pointer;
    }
  }

  .t-table__body {
    color: #63656e;
  }

  .t-table__pagination {
    .t-pagination {
      color: #63656e;

      .t-input--focused {
        box-shadow: none;
      }
    }

    .t-pagination__number.t-is-current {
      background-color: #e1ecff;
      color: #3a84ff;
      border: none;
    }
  }

  // 默认的 loading 图标
  .t-loading svg.t-icon-loading {
    display: none !important;
  }

  .t-table__row--full.t-table__first-full-row {
    background-color: #f0f1f5;

    td {
      border: none;
    }

    .t-table__row-full-element {
      padding: 0;
    }
  }

  &.primary-table-no-data {
    .t-table__row--full.t-table__first-full-row {
      height: 0;
    }
  }
}

.custom-radio-filter-wrapper {
  .t-table__filter--bottom-buttons {
    .t-button:nth-child(2) {
      display: none !important;
    }
  }
}
</style>
