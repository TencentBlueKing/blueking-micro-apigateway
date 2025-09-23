<template>
  <div style="padding-bottom: 12px;">
    <bk-table :columns="columns" :data="data" :pagination="pagination" row-key="resource_id" />
  </div>
</template>

<script lang="tsx" setup>
import { RESOURCE_CN_MAP } from '@/enum';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';

interface IProps {
  type: 'add' | 'update' | 'uncheck'
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

const { type } = defineProps<IProps>();

const emit = defineEmits<{
  'uncheck': [row: IRow]
  'recover': [row: IRow]
}>();

const { t } = useI18n();

const pagination = ref({
  current: 1,
  limit: 10,
  count: data.value.length,
  showLimit: false,
});

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

const handleRecover = (row: IRow) => {
  emit('recover', row);
};

const handleUncheck = (row: IRow) => {
  emit('uncheck', row);
};

</script>
