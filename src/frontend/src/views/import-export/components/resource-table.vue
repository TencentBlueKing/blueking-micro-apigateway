<template>
  <div style="padding-bottom: 12px;">
    <bk-table :columns="columns" :data="data" :pagination="pagination" row-key="resource_id">
      <template #empty>
        <TableEmpty
          :type="tableEmptyType"
          @clear-filter="handleClearFilterKey"
        />
      </template>
    </bk-table>
  </div>
</template>

<script lang="tsx" setup>
import { RESOURCE_CN_MAP } from '@/enum';
import { useI18n } from 'vue-i18n';
import { computed, ref, watch } from 'vue';
import TableEmpty from '@/components/table-empty.vue';

interface IProps {
  type: 'add' | 'update' | 'uncheck'
  keywords?: string
}

interface IRow {
  name: string
  resource_id: string
  resource_type: string
  status: string
  config: Record<string, any>
  __source__?: string
}

const data = defineModel<IRow[]>('data', {
  default: [],
});

const { type, keywords } = defineProps<IProps>();

const emit = defineEmits<{
  'uncheck': [row: IRow]
  'recover': [row: IRow]
  'clear-filter': []
}>();

const { t } = useI18n();

const pagination = ref({
  current: 1,
  limit: 10,
  count: data.value.length,
  showLimit: false,
});

const tableEmptyType = ref<'empty' | 'search-empty'>('empty');

const columns = computed(() => [
  {
    label: t('名称'),
    field: 'name',
  },
  {
    label: 'ID',
    field: 'resource_id',
  },
  {
    label: t('类型'),
    field: 'resource_type',
    render: ({ row }: { row: IRow }) => RESOURCE_CN_MAP[row.resource_type] || '--',
  },
  {
    label: t('状态'),
    field: 'status',
  },
  {
    label: t('操作'),
    field: 'operation',
    render: ({ row }: { row: IRow }) => (
      <div style="display: flex;gap: 8px;">
        {type === 'uncheck'
          ? <bk-button
            text
            theme="primary"
            onClick={() => handleRecover(row)}
          >
            {t('恢复')}
          </bk-button>
          : <bk-button
            text
            theme="primary"
            onClick={() => handleUncheck(row)}
          >
            {t('不导入')}
          </bk-button>
        }
      </div>
    ),
  },
]);

const updateTableEmptyConfig = () => {
  if (keywords) {
    tableEmptyType.value = 'search-empty';
  } else {
    tableEmptyType.value = 'empty';
  }
};

watch(() => keywords, () => {
  updateTableEmptyConfig();
}, { immediate: true });

const handleClearFilterKey = () => {
  emit('clear-filter');
  updateTableEmptyConfig();
};

const handleRecover = (row: IRow) => {
  emit('recover', row);
};

const handleUncheck = (row: IRow) => {
  emit('uncheck', row);
};

</script>
