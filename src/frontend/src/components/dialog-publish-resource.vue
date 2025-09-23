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
  <bk-dialog
    v-model:is-show="isShow"
    :quick-close="true"
    title="待发布的资源"
    width="1280"
    @confirm="handleConfirm"
    @closed="handleClose"
  >
    <div v-if="diffGroupList.length">
      <bk-tab
        v-model:active="activeTab"
        tab-position="left"
        type="card"
      >
        <bk-tab-panel
          v-for="item in diffGroupList"
          :key="item.resource_type"
          :label="common.enums?.resource_type[item.resource_type]"
          :name="item.resource_type"
          :num="item.added_count + item.deleted_count + item.modified_count"
        >
          <div>
            <table-resource-to-publish
              :data="item.change_detail"
              :resource-type="item.resource_type"
              :show-del="true"
              :table-empty-type="tableEmptyType"
              :delete-api="deleteApi"
              :filter-value="filterData"
              @filter-change="(filterItem) => handleFilterChange(filterItem, item)"
              @clear-filter="handleClearFilter"
              @refresh="refresh"
              @del="(id) => emit('del', id)"
            />
          </div>
        </bk-tab-panel>
      </bk-tab>
    </div>
    <div v-else>
      <bk-exception
        :description="t('没有需要发布的资源')"
        class="exception-wrap-item exception-part"
        scene="part"
        type="empty"
      />
    </div>
  </bk-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { InfoBox } from 'bkui-vue';
import { cloneDeep } from 'lodash-es';
import { type FilterValue } from '@blueking/tdesign-ui';
import { IDiffGroup } from '@/http/publish';
import { useCommon } from '@/store';
import useConfigFilter from '@/hooks/use-config-filter';
import TableResourceToPublish from '@/components/table-resource-to-publish.vue';

interface IProps {
  diffGroupList: IDiffGroup[];
  diffGroupListBack: IDiffGroup[];
  resourceType?: string;
  deleteApi?: (...args: any[]) => Promise<unknown>;
}

const isShow = defineModel<boolean>({
  required: true,
});

const {
  diffGroupList = [],
  diffGroupListBack = [],
  deleteApi,
} = defineProps<IProps>();

const emit = defineEmits<{
  'refresh': [boolean?]
  'confirm': [void]
  'del': [string]
  'closed': [void]
}>();

const { t } = useI18n();
const common = useCommon();
const { filterEmpty } = useConfigFilter();

const activeTab = ref('');
const tableEmptyType = ref('empty');
const filterData = ref<FilterValue>({});

const handleFilterChange = (filterItem: FilterValue, row: IDiffGroup) => {
  const isExistField = Object.keys(filterEmpty(filterItem))?.length > 0;
  const filterList =  cloneDeep(diffGroupListBack);
  const currTab = filterList.find(item => item.resource_type === activeTab.value);
  filterData.value = Object.assign({}, filterItem);
  if (currTab && filterItem?.operation_type) {
    const detailList = currTab.change_detail.filter(item => filterItem.operation_type === item.operation_type);
    row.change_detail = detailList;
  }
  if (!isExistField) {
    handleClearFilter();
  }
  tableEmptyType.value = isExistField ? 'search-empty' : 'empty';
};

const handleConfirm = () => {
  if (!diffGroupList.length) {
    return emit('closed');
  }

  isShow.value = true;

  InfoBox({
    title: t('确认发布所有资源？'),
    confirmText: t('确认'),
    cancelText: t('取消'),
    onConfirm: () => {
      emit('confirm');
      isShow.value = false;
    },
  });
};

const handleClose = () => {
  filterData.value = {};
  emit('closed');
};

const handleClearFilter = () => {
  filterData.value = {};
  emit('refresh', true);
};

const refresh = (hideTips?: boolean) => {
  emit('refresh', hideTips);
};

</script>

<style lang="scss" scoped>

:deep(.bk-tab--left .bk-tab-header-nav .bk-tab-header-item) {
  min-width: 100px;
  padding-left: 0;
}

</style>
