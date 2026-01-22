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
  <table-resource-list
    ref="tableRef"
    :columns="columns"
    :delete-api="deleteService"
    :extra-search-options="extraSearchOptions"
    :extra-search-params="searchParams"
    :routes="{ create: 'service-create', edit: 'service-edit', clone: 'service-clone' }"
    resource-type="service"
    :query-list-params="{ apiMethod: getServices }"
    @check-resource="toggleResourceViewerSlider"
    @clear-filter="handleTableClearFilter"
  >
    <!--    <template #extraSearch>-->
    <!--      <div style="display: flex; align-items: center; gap: 8px;">-->
    <!--        <div style="font-size: 12px;">{{ t('上游') }}</div>-->
    <!--        <bk-select-->
    <!--          v-model="relationSearchParams.upstream_id"-->
    <!--          :popover-options="{ zIndex: 999 }"-->
    <!--          filterable-->
    <!--          style="width: 180px;"-->
    <!--          @change="(id: string) => handleRelationResourceIdChange({ field: 'upstream_id', id })"-->
    <!--        >-->
    <!--          <bk-option-->
    <!--            v-for="item in upstreamSelectOptions"-->
    <!--            :id="item.value"-->
    <!--            :key="item.value"-->
    <!--            :name="item.label"-->
    <!--          >-->
    <!--            <div v-bk-tooltips="{ content: relatedResourceTooltipContent(item), placement: 'left' }">-->
    <!--              <span>{{ item.label }}</span>-->
    <!--              <span-->
    <!--                v-if="item.desc"-->
    <!--                style="padding-left: 8px;color: #979ba5;"-->
    <!--              >-->
    <!--                {{ item.desc }}-->
    <!--              </span>-->
    <!--            </div>-->
    <!--          </bk-option>-->
    <!--        </bk-select>-->
    <!--      </div>-->
    <!--    </template>-->
  </table-resource-list>
  <slider-resource-viewer
    v-model="isResourceViewerShow"
    editable
    :resource="service"
    :source="source"
    resource-type="service"
    @updated="handleUpdated"
  />
</template>

<script lang="tsx" setup>
import TableResourceList, { ISearchParam } from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { deleteService, getService, getServices } from '@/http/service';
import { computed, ref, watch } from 'vue';
import { IService } from '@/types/service';
import { FilterOptionClass, type IFilterOption } from '@/types/table-filter';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import type { PrimaryTableProps, TableRowData } from '@blueking/tdesign-ui';
import { getUpstreamDropdowns } from '@/http/upstream';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const service = ref<IService>();

// 关联资源传参
const relationSearchParams = ref<{ upstream_id?: string }>({});
// 关联的上游 select 下拉选项
const upstreamSelectOptions = ref<{ value: string, label: string, desc: string }[]>([]);

const columns = computed<PrimaryTableProps['columns']>(() => [
  {
    title: 'ID',
    colKey: 'id',
    ellipsis: true,
  },
  {
    title: t('匹配域名'),
    colKey: 'type',
    ellipsis: true,
    cell: (h, { row }: any) => row.config?.hosts?.map((host: string) => host)
      .join(', ') || '--',
  },
  {
    title: t('描述'),
    colKey: 'desc',
    ellipsis: true,
    cell: (h, { row }: any) => row.config?.desc || '--',
  },
  {
    title: t('上游'),
    colKey: 'upstream_id',
    ellipsis: true,
    filter: {
      type: 'single',
      showConfirmAndReset: true,
      list: getFilterOptions({ options: upstreamSelectOptions.value, extra: true }),
      popupProps: {
        overlayInnerClassName: 'custom-radio-filter-wrapper',
      },
    },
    cell: (h, { row }: TableRowData) => (row.upstream_id
      ? <bk-button
        text theme="primary"
        onClick={() => handleRelatedResourceIdClicked({ routeName: 'upstream', id: row.upstream_id })}
      >{upstreamNameMap[row.upstream_id]}</bk-button> : '--'),
  },
]);

const source = ref('');
const isResourceViewerShow = ref(false);
const searchParams = ref<ISearchParam[]>([]);
const tableRef = ref();

let upstreamNameMap: Record<string, string> = {};

const extraSearchOptions = computed(() => [
  {
    id: 'upstream_id',
    name: t('上游'),
    children: getFilterOptions({
      options: upstreamSelectOptions.value,
      key: 'name',
      value: 'id',
      extra: true,
    }),
  },
]);

const toggleResourceViewerSlider = ({ resource }: { resource: IService }) => {
  service.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

watch(() => route.query.id, async () => {
  if (route.query.id) {
    const id = route.query.id as string;
    const resource = await getService({ id });
    toggleResourceViewerSlider({ resource });
    searchParams.value = [{
      id: 'id',
      name: 'ID',
      values: [{ id, name: id }],
    }];
  }
}, { immediate: true });

// 根据不同键值初始化数组结构
function getFilterOptions({
  key,
  value,
  options,
  extra,
}: {
  key: string,
  value?: string | number,
  options: IFilterOption[],
  extraOption?: boolean | IFilterOption[],
})  {
  return new FilterOptionClass({ key, value, options, extra })?.filterOptions;
};

const getUpstreamSelectOptions = async () => {
  const response = await getUpstreamDropdowns();
  upstreamSelectOptions.value = (response ?? []).map(item => ({
    name: item.name,
    id: item.id,
    label: item.name,
    value: item.id,
    desc: item.desc,
  }));
  upstreamNameMap = upstreamSelectOptions.value.reduce<Record<string, string>>((acc, cur) => {
    acc[cur.id] = cur.name;
    return acc;
  }, {});
};
getUpstreamSelectOptions();

// const handleRelationResourceIdChange = ({ field, id }: { field: 'service_id' | 'upstream_id', id: string }) => {
//   if (!id) {
//     delete relationSearchParams.value[field];
//   }
//   tableRef.value!.getList({ ...relationSearchParams.value });
// };

const handleTableClearFilter = () => {
  relationSearchParams.value = {};
};

const handleRelatedResourceIdClicked = ({ routeName, id }: { routeName: string, id: string }) => {
  const to = router.resolve({ name: routeName, query: { id } });
  window.open(to.href);
};

const handleUpdated = async () => {
  tableRef.value!.getList({ ...relationSearchParams.value });
  service.value = await getService({ id: service.value.id });
};

// const relatedResourceTooltipContent = (item: {
//   label: string,
//   desc: string
// }) => (item.desc ? `${item.label}(${item.desc})` : item.label);

</script>
