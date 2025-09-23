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
  <div class="table-wrapper">
    <MicroAgTable
      v-model:selected-row-keys="selectionsRowKeys"
      v-model:is-all-selection="isAllSelection"
      v-model:table-data="tableData"
      v-model:settings="settings"
      :row-key="tableRowKey"
      :hover="false"
      resizable
      local-page
      :filter-value="filterValue"
      :table-empty-type="tableEmptyType"
      :columns="columns"
      @clear-filter="handleClearFilter"
      @filter-change="handleFilterChange"
    />

    <SliderResourceDiffViewer
      v-model="isDiffSliderShow"
      :after-config="configDiff.after_config"
      :before-config="configDiff.before_config"
      :operation-type="operationType"
    />

    <SliderPublishDiff
      v-model="isPublishDiffShow"
      :after-config="configDiff.after_config"
      :before-config="configDiff.before_config"
      :source-info="sourceInfo"
      @done="emit('refresh', true);"
    />
  </div>
</template>

<script lang="tsx" setup>
import { ref, watch, computed, shallowRef } from 'vue';
import { cloneDeep } from 'lodash-es';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useCommon } from '@/store';
import { revertResource } from '@/http/revert';
import { getResourceDiff, IChangeDetail } from '@/http/publish';
import { useTDesignSelection } from '@/hooks/use-tdesign-selection';
import { type PrimaryTableProps } from '@blueking/tdesign-ui';
import fetch from '@/http/fetch';
import dayjs from 'dayjs';
// @ts-ignore
import SliderResourceDiffViewer from '@/components/slider-resource-diff-viewer.vue';
// @ts-ignore
import SliderPublishDiff from '@/components/slider-publish-diff.vue';
import MicroAgTable from '@/components/micro-ag-table/table';
import TagOperationType from '@/components/tag-operation-type.vue';

interface IProps {
  data: IChangeDetail[];
  resourceType: string;
  showDel?: boolean;
  deleteApi?: (...args: any[]) => Promise<unknown>;
  tableRowKey?: string;
  tableEmptyType?: 'empty' | 'search-empty'
  disabled?: boolean;
  filterValue?: Record<string, string> | null
}

const {
  data = [],
  resourceType = '',
  tableRowKey = 'resource_id',
  tableEmptyType = 'empty',
  showDel = false,
  disabled = false,
  filterValue = {},
  deleteApi,
} = defineProps<IProps>();

const emit = defineEmits<{
  'check-resource': [void]
  'refresh': [boolean?]
  'del': [string]
  'clear-filter': [void]
  'filter-change': Record<string, string>
}>();

watch(
  () => data,
  (v) => {
    tableData.value = v;
  },
);

const { t } = useI18n();
const common = useCommon();
const {
  selectionsRowKeys,
  isAllSelection,
} = useTDesignSelection();

const { BK_DASHBOARD_URL } = window;

const operationTypeFilter = computed(() => {
  const list =  Object.keys(common.enums?.operation_type ?? {})
    ?.filter((key: string) => (['create', 'update', 'delete'].includes(key)))
    ?.map((key: string) => ({
      label: common.enums?.operation_type[key],
      value: key,
    }));
  return list;
});

const columns: PrimaryTableProps['columns'] = [
  {
    title: 'ID',
    colKey: 'resource_id',
    ellipsis: true,
    fixed: 'left',
  },
  {
    title: t('名称'),
    colKey: 'name',
    cell: (h, { row }) => {
      if (!row?.name) {
        return '--';
      }
      return (
        <div class="flex-row">
          <div
            class="is-color-active single-ellipse"
            v-bk-tooltips={{
              content: row.name,
              placement: 'top',
              disabled: !row.isOverflow,
            }}
            onClick={() => handleCheckCodeClick(row)}
           >
            {row.name}
          </div>
        </div>
      );
    },
  },
  {
    title: t('操作类型'),
    colKey: 'operation_type',
    filter: {
      type: 'single',
      showConfirmAndReset: true,
      popupProps: {
        overlayInnerClassName: 'custom-radio-filter-wrapper',
      },
      list: operationTypeFilter.value,
    },
    cell: (h, { row }) =>  <TagOperationType type={row.operation_type} />,
  },
  {
    title: t('更新时间'),
    colKey: 'updated_at',
    ellipsis: true,
    cell: (h, { row }) => {
      return dayjs.unix(row.updated_at).format('YYYY-MM-DD HH:mm:ss Z');
    },
  },
  {
    title: t('操作'),
    colKey: 'opt',
    fixed: 'right',
    cell: (h, { row }) => {
      return (
        <div class="table-cell-actions">
          <bk-button
            text
            theme="primary"
            disabled={disabled}
            onClick={() => handlePublishClick(row)}
          >
            { t('发布') }
          </bk-button>
          <bk-button
            text
            theme="primary"
            disabled={disabled}
            onClick={() => handleCheckCodeClick(row)}
          >
            { t('差异') }
          </bk-button>
          <bk-pop-confirm
            title={t('确认撤销该资源变更？')}
            width={288}
            trigger="click"
            onConfirm={() => handleRevertClick(row)}
          >
          {{
            content: () => (
              <div>
                <div class="mb4">{ t('服务名称：{name}', { name: row.name }) }</div>
                <div class="mb20">{ t('撤销后，不可恢复，请谨慎操作。') }</div>
              </div>
            ),
            default: () => (
              <bk-button
                text
                theme="primary"
                disabled={revertDisabled(row)}
                v-bk-tooltips={{
                  content: t('新增状态只能删除'),
                  disabled: !revertDisabled(row),
                }}
              >
                { t('撤销') }
              </bk-button>
            ),
          }}
          </bk-pop-confirm>
          <bk-pop-confirm
            content={!['success'].includes(row.status) ? t('删除操作无法撤回，请谨慎操作！') : ''}
            title={t('确认删除？')}
            trigger="click"
            width={288}
            onConfirm={() => handleDeleteClick(row)}
          >
            {
              showDel && ['create'].includes(row.operation_type)
              && (
                <bk-button
                  text
                  theme="danger"
                  disabled={disabled}
                >
                  { t('删除') }
                </bk-button>
              )
            }
          </bk-pop-confirm>
        </div>
      );
    },
  },
];
const isDiffSliderShow = ref(false);
const isPublishDiffShow = ref(false);
const configDiff = ref({
  after_config: {},
  before_config: {},
});
const operationType = ref('');
const sourceInfo = ref<{type?: string; name?: string; id?: string}>({});
const tableData = ref<IChangeDetail[]>(data);
const settings = shallowRef({
  size: 'small',
  checked: [],
  disabled: [],
});

const revertDisabled = (row: IChangeDetail) => {
  return (disabled || !['update_draft', 'delete_draft'].includes(row.before_status));
};

const handleFilterChange: PrimaryTableProps['onFilterChange'] = (filterItem) => {
  emit('filter-change', filterItem);
};

const handleRevertClick = async (row: IChangeDetail) => {
  if (!row.resource_id || !resourceType) {
    Message({
      theme: 'error',
      message: t('获取资源ID或资源类型失败'),
    });
    return;
  }

  await revertResource({
    data: {
      resource_id_list: [row.resource_id],
      resource_type: resourceType === 'all' ? row.resource_type : resourceType,
    },
  });

  Message({
    theme: 'success',
    message: t('已撤销'),
  });

  emit('refresh', true);
};

const handleDeleteClick = async (row: Record<string, any>) => {
  try {
    if (deleteApi) {
      await deleteApi({ id: row.resource_id } as { id: string });
    } else {
      await fetch.delete(`${BK_DASHBOARD_URL}/gateways/${common.gatewayId}/${row.resource_type}s/${row.resource_id}/`);
    }

    Message({
      theme: 'success',
      message: t('已删除'),
    });

    emit('del', row.resource_id);
  } catch (e) {
    Message({
      theme: 'error',
      message: t('删除失败'),
    });
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

const getDiffData = async (row: IChangeDetail) => {
  const res = await getResourceDiff({
    type: resourceType === 'all' ? row.resource_type : resourceType,
    id: row.resource_id,
  });
  const config = res || {
    editor_config: {},
    etcd_config: {},
  };
  configDiff.value.before_config = filterTimeKeys(config.etcd_config || {});
  configDiff.value.after_config = filterTimeKeys(config.editor_config || {});
};

const handleCheckCodeClick = async (row: IChangeDetail) => {
  await getDiffData(row);
  operationType.value = row.operation_type;
  isDiffSliderShow.value = true;
};

const handlePublishClick = async (row: IChangeDetail) => {
  await getDiffData(row);
  sourceInfo.value = {
    type: resourceType === 'all' ? row.resource_type : resourceType,
    name: row.name,
    id: row.resource_id,
  };
  isPublishDiffShow.value = true;
};

const handleClearFilter = () => {
  emit('clear-filter');
};

</script>

<style lang="scss" scoped>

:deep(.table-cell-actions) {
  display: flex;
  gap: 12px;
}

</style>
