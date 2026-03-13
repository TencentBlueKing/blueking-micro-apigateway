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
          :disabled="isReadonlyGateway"
          style="width: 142px;"
          theme="primary"
          @click="handleCreateClick"
        >
          {{ t('新建') }}
        </BkButton>
      </div>
      <div class="right flex-row justify-content-center">
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
        :pagination="false"
        :filter-value="{}"
        :api-method="getTableData"
        :columns="columns"
      />
    </div>
  </div>

  <Create
    ref="createRef"
    @done="getList()" />

  <Details
    v-model="diffSliderConfig.visible"
    :data="diffSliderConfig.data" />
</template>

<script lang="tsx" setup>
import { computed, ref, shallowRef, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { ITableMethod, IMcpToken } from '@/types';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/zh-cn';
import { useCommon } from '@/store';
import {
  getMcpTokens,
  // getMcpTokensDetails,
  deleteMcpToken,
} from '@/http/mcp';
// @ts-ignore
import Create from './create.vue';
import Details from './details.vue';
// @ts-ignore
import MicroAgTable from '@/components/micro-ag-table/table';
import type {
  PrimaryTableProps,
} from '@blueking/tdesign-ui';

const common = useCommon();
const { t } = useI18n();

dayjs.extend(relativeTime);
dayjs.locale('zh-cn');

const tableData = ref<IMcpToken[]>([]);
const tableRef = useTemplateRef<InstanceType<typeof MicroAgTable> & ITableMethod>('tableRef');
const createRef = ref<InstanceType<typeof Create>>();

const diffSliderConfig = ref<any>({
  visible: false,
  data: {},
});
const settings = shallowRef({
  size: 'small',
  checked: [],
  disabled: [],
});

const isReadonlyGateway = computed(() => common.curGatewayData?.read_only);

const columns: PrimaryTableProps['columns'] = [
  {
    title: t('名称'),
    colKey: 'name',
    ellipsis: true,
  },
  {
    title: t('令牌'),
    colKey: 'masked_token',
    ellipsis: true,
    width: 410,
  },
  {
    title: t('描述'),
    colKey: 'description',
    ellipsis: true,
  },
  {
    title: t('访问范围'),
    colKey: 'access_scope',
    ellipsis: true,
    width: 100,
    cell: (h, { row }) => (
    <bk-tag
      radius="8px"
      theme={row.access_scope === 'readwrite' ? 'success' : 'info'}
    >
      {row.access_scope ?? '--'}
    </bk-tag>
    ),
  },
  {
    title: t('状态'),
    colKey: 'is_expired',
    ellipsis: true,
    width: 100,
    cell: (h, { row }) => {
      return (
        <bk-tag
          radius="8px"
          theme={row.is_expired ? 'danger' : 'success'}
        >
          {row.is_expired ? t('已过期') : t('活跃')}
        </bk-tag>
      );
    },
  },
  {
    title: t('过期时间'),
    colKey: 'expired_at',
    ellipsis: true,
    cell: (h, { row }) => dayjs.unix(row.expired_at).format('YYYY-MM-DD HH:mm:ss Z'),
  },
  {
    title: t('最近使用'),
    colKey: 'last_used_at',
    ellipsis: true,
    width: 100,
    cell: (h, { row }) => {
      if (row.last_used_at === null) {
        return t('从未使用');
      }
      return dayjs.unix(row.last_used_at).fromNow();
    },
  },
  {
    title: t('创建时间'),
    colKey: 'created_at',
    ellipsis: true,
    cell: (h, { row }) => dayjs.unix(row.created_at).format('YYYY-MM-DD HH:mm:ss Z'),
  },
  {
    title: t('创建人'),
    colKey: 'creator',
    ellipsis: true,
    width: 100,
  },
  {
    title: t('操作'),
    colKey: 'opt',
    fixed: 'right',
    width: 100,
    cell: (h, { row }) => {
      return (
        <>
        <bk-button text theme="primary" class="mr8" onClick={() => handleDetails(row as IMcpToken)}>
          { t('详情') }
        </bk-button>
        <bk-pop-confirm
          width="288"
          content={t('此操作不可恢复，使用该令牌的 MCP 客户端将无法继续访问。')}
          title={t(`确定要删除令牌 ${row.name} 吗？`)}
          trigger="click"
          onConfirm={() => handleDel(row as IMcpToken)}
        >
          <bk-button text theme="danger">
              { t('删除') }
            </bk-button>
        </bk-pop-confirm>
        </>
      );
    },
  },
];

const getList = () => {
  tableRef.value!.fetchData({}, { resetPage: true });
};

const getTableData = async () => {
  const results = await getMcpTokens({ gatewayId: common.gatewayId });
  return {
    count: results?.length ?? 0,
    results: results ?? [],
  };
};

const handleCreateClick = () => {
  createRef.value?.show();
};

const handleDel = async (row: IMcpToken) => {
  await deleteMcpToken({
    gatewayId: common.gatewayId,
    id: row.id,
  });
  Message({
    message: t('删除成功'),
    theme: 'success',
  });
  getList();
};

const handleDetails = async (row: IMcpToken) => {
  diffSliderConfig.value.data = row;
  diffSliderConfig.value.visible = true;
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
