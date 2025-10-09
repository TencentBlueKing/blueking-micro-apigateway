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
    <BkAlert
      theme="info"
      :closable="false"
      class="mb16"
    >
      <template #title>
        <div>
          <span>{{ t('与 etcd 中同步过来的快照数据对比，数据同步时间：') }}
            {{ lastTime ? dayjs.unix(lastTime).format('YYYY-MM-DD HH:mm:ss Z') : '--' }}，
          </span>
          <span class="alert-sync" @click="handleSync">{{ t('立即同步') }}</span>
        </div>
      </template>
    </BkAlert>
    <div class="header-actions">
      <div class="left">
        <BkButton
          theme="primary"
          @click="handlePublishConfirm"
          :disabled="common.curGatewayData?.read_only || !diffGroupTotal?.length"
        >
          {{ t('全局发布') }}
        </BkButton>
      </div>
      <div class="right">
        <BkSearchSelect
          v-model="searchParams"
          v-click-outside="handleSearchOutside"
          :data="searchOptions"
          :placeholder="t('搜索名称、ID、操作类型')"
          clearable
          class="table-resource-search"
          unique-select
          @click.stop="handleSearchSelectClick"
        />
      </div>
    </div>
    <div class="table-wrapper">
      <div class="side-category">
        <div
          v-for="item in diffGroupAll"
          :key="item.resource_type"
          :class="{ 'category-item': true, 'active': activeTab === item.resource_type }"
          @click="handleChange(item)"
        >
          <div class="name">
            {{ item.resource_type === 'all' ? t('全部') : common.enums?.resource_type[item.resource_type] }}
          </div>
          <div class="num">
            {{ item.added_count + item.deleted_count + item.modified_count }}
          </div>
        </div>
      </div>
      <div class="main-wrapper">
        <div>
          <TableResourceToPublish
            :data="currentResource?.change_detail || []"
            :resource-type="currentResource?.resource_type"
            :show-del="true"
            :table-empty-type="emptyType"
            :disabled="common.curGatewayData?.read_only"
            :filter-value="filterData"
            @filter-change="handleFilterChange"
            @del="refresh"
            @refresh="refresh"
            @clear-filter="handleClearFilter"
          />
        </div>
      </div>
    </div>
  </div>

  <slider-batch-publish-diff
    v-model="isDiffSliderShow"
    :list="diffGroupTotal"
    @done="refresh"
  />
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { InfoBox, Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';
import { useCommon } from '@/store';
// @ts-ignore
import TableResourceToPublish from '@/components/table-resource-to-publish.vue';
import { getDiffAll, IDiffGroup } from '@/http/publish';
import usePublishSearch from '@/hooks/use-publish-search';
import { useTableFilterChange } from '@/hooks/use-table-filter-change';
import { useSearchSelectPopoverHidden } from '@/hooks/use-search-select-popover-hidden';
import { getSyncLastTime, postGatewaySyncData } from '@/http/gateway-sync-data';
// @ts-ignore
import sliderBatchPublishDiff from '@/components/slider-batch-publish-diff.vue';

const { t } = useI18n();
const common = useCommon();
const { handleTableFilterChange } = useTableFilterChange();
const { handleSearchOutside, handleSearchSelectClick } = useSearchSelectPopoverHidden();

const lastTime = ref<number>();
const activeTab = ref<string>('');
const isDiffSliderShow = ref<boolean>(false);
const currentResource = ref<IDiffGroup>({
  resource_type: '',
  added_count: 0,
  deleted_count: 0,
  modified_count: 0,
  change_detail: [],
});
const diffGroupTotal = ref<IDiffGroup[]>([]);
const diffGroupAll = ref<IDiffGroup[]>([]);
const filterData = ref<Record<string, any>>({});
const searchParams = ref<{ id: string, name: string, values?: { id: string, name: string }[] }[]>([]);
// 表格没数据时的空状态控制变量
const emptyType = ref<'empty' | 'search-empty'>('empty');

const searchOptions = computed(() => {
  return [
    {
      id: 'name',
      name: '名称',
    },
    {
      id: 'id',
      name: 'ID',
    },
    {
      id: 'operation_type',
      name: t('操作类型'),
      children: Object.keys(common.enums?.operation_type ?? {})?.filter((key: string) => (['create', 'update', 'delete'].includes(key)))
        ?.map((key: string) => ({
          name: common.enums?.operation_type[key],
          id: key,
        })),
    },
  ];
});

watch(
  () => searchParams.value,
  () => {
    handleSearch();
  },
);

const initCategory = () => {
  setDiffGroupAll(diffGroupShow.value);
  [currentResource.value] = diffGroupAll.value;
  activeTab.value = 'all';
};

const { regroupData, diffGroupShow } = usePublishSearch({
  filterData,
  diffGroupTotal,
  searchDoneFn: initCategory,
});

const handleSearch = () => {
  const data: Record<string, any> = {};
  searchParams.value.forEach((option) => {
    if (option.values) {
      data[option.id] = option.values[0]?.id;
    } else {
      data.keywords += `&${option.id}`;
    }
  });
  filterData.value = data;
  emptyType.value = Object.keys(data)?.length > 0 ? 'search-empty' : 'empty';
  // 过滤数据
  regroupData();
};

const refresh = () => {
  getDiffGroupList();
};

const setDiffGroupAll = (diffList: IDiffGroup[]) => {
  if (!diffList?.length) {
    diffGroupAll.value = [];
    return;
  }

  const allItem: IDiffGroup = {
    added_count: 0,
    modified_count: 0,
    deleted_count: 0,
    resource_type: 'all',
    change_detail: [],
  };

  diffList.forEach((source: IDiffGroup) => {
    allItem.added_count += source.added_count;
    allItem.modified_count += source.modified_count;
    allItem.deleted_count += source.deleted_count;
    allItem.change_detail.push(...source.change_detail.map(item => ({ ...item, resource_type: source.resource_type })));
  });

  diffGroupAll.value = [allItem, ...diffList];
};

const getDiffGroupList = async () => {
  const res = await getDiffAll({ data: filterData.value });

  if (res?.length) {
    diffGroupTotal.value = res;
    diffGroupShow.value = res;
    initCategory();
  } else {
    diffGroupTotal.value = [];
    diffGroupShow.value = [];
    diffGroupAll.value = [];
    currentResource.value = {
      resource_type: '',
      added_count: 0,
      deleted_count: 0,
      modified_count: 0,
      change_detail: [],
    };
    activeTab.value = 'all';
  }
};

getDiffGroupList();

// const handleClearFilterKey = () => {
//   filterData.value = {};
// };

const handleChange = (resource: IDiffGroup) => {
  currentResource.value = resource;
  activeTab.value = resource.resource_type;
  // searchParams.value = [];
};

const handleFilterChange = (filterItem) => {
  handleTableFilterChange({
    filterItem,
    filterData,
    searchOptions,
    searchParams,
  });
  emptyType.value = Object.keys(filterItem).length > 0 ? 'search-empty' : 'empty';
  regroupData();
};

const handlePublishConfirm = async () => {
  isDiffSliderShow.value = true;
};

const getSyncTime = async () => {
  const res = await getSyncLastTime({});
  lastTime.value = res.latest_time;
};
getSyncTime();

const handleSync = async () => {
  const infoBoxRef = InfoBox({
    type: 'loading',
    title: t('正在同步…'),
    content: '请稍等',
    class: 'bk-hide-footer',
  });
  try {
    await postGatewaySyncData({ data: {} });

    getDiffGroupList();
    getSyncTime();
    Message({
      theme: 'success',
      message: t('同步成功'),
    });
  } catch (e) {} finally {
    infoBoxRef?.hide();
  }
};

const handleClearFilter = () => {
  searchParams.value = [];
};

</script>

<style lang="scss" scoped>
.page-content-wrapper {
  min-height: calc(100vh - 157px);
  padding: 16px 24px 24px;

  .table-wrapper {
    display: flex;
    align-items: flex-start;
    .side-category {
      width: 240px;
      min-width: 240px;
      padding: 12px 8px;
      height: calc(100vh - 272px);
      margin-right: 16px;
      border-radius: 2px;
      background-color: #ffffff;
      box-shadow: 0 2px 4px 0 #1919290d;
      .category-item {
        height: 40px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 8px;
        border-radius: 4px;
        cursor: pointer;
        background: #FFFFFF;
        .name {
          font-size: 14px;
          color: #313238;
        }
        .num {
          font-size: 12px;
          padding: 0px 8px;
          border-radius: 2px;
          color: #979BA5;
          background: #F0F1F5;
        }
        &.active {
          background: #F0F5FF;
          .name {
            color: #3A84FF;
          }
          .num {
            color: #FFFFFF;
            background: #A3C5FD;
          }
        }
      }
    }
    .main-wrapper {
      flex: 1;
      min-height: calc(100vh - 272px);
      border-radius: 2px;
      background-color: #ffffff;
      box-shadow: 0 2px 4px 0 #1919290d;
      padding: 12px;
      .change-total {
        margin-bottom: 12px;
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
  }
}

:deep(.table-cell-actions) {
  display: flex;
  gap: 12px;
}

.alert-sync {
  color: #3A84FF;
  cursor: pointer;
}
</style>
<style lang="scss">
.bk-hide-footer {
  .bk-modal-footer {
    height: 48px;
    .bk-infobox-footer {
      display: none;
    }
  }
}
</style>
