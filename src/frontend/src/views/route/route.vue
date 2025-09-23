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
    :extra-search-options="extraSearchOptions"
    :delete-api="deleteRoute"
    :query-list-params="{ apiMethod: getRoutes }"
    :routes="{ create: 'route-create', edit: 'route-edit', clone: 'route-clone' }"
    resource-type="route"
    @check-resource="toggleResourceViewerSlider"
    @clear-filter="handleTableClearFilter"
  >
    <!--    <template #extraSearch>-->
    <!--      <div style="display: flex; align-items: center; gap: 12px;">-->
    <!--        <div style="display: flex; align-items: center; gap: 8px;">-->
    <!--          <div style="font-size: 12px;">{{ t('服务') }}</div>-->
    <!--          <bk-select-->
    <!--            v-model="relationSearchParams.service_id"-->
    <!--            :popover-options="{ zIndex: 999 }"-->
    <!--            filterable-->
    <!--            style="width: 180px;"-->
    <!--            @change="(id: string) => handleRelationResourceIdChange({ field: 'service_id', id })"-->
    <!--          >-->
    <!--            <bk-option-->
    <!--              v-for="item in serviceSelectOptions"-->
    <!--              :id="item.value"-->
    <!--              :key="item.value"-->
    <!--              :name="item.label"-->
    <!--            >-->
    <!--              <div v-bk-tooltips="{ content: relatedResourceTooltipContent(item), placement: 'left' }">-->
    <!--                <span>{{ item.label }}</span>-->
    <!--                <span-->
    <!--                  v-if="item.desc"-->
    <!--                  style="padding-left: 8px;color: #979ba5;"-->
    <!--                >-->
    <!--                  {{ item.desc }}-->
    <!--                </span>-->
    <!--              </div>-->
    <!--            </bk-option>-->
    <!--          </bk-select>-->
    <!--        </div>-->
    <!--        <div style="display: flex; align-items: center; gap: 8px;">-->
    <!--          <div style="font-size: 12px;">{{ t('上游') }}</div>-->
    <!--          <bk-select-->
    <!--            v-model="relationSearchParams.upstream_id"-->
    <!--            :popover-options="{ zIndex: 999 }"-->
    <!--            filterable-->
    <!--            style="width: 180px;"-->
    <!--            @change="(id: string) => handleRelationResourceIdChange({ field: 'upstream_id', id })"-->
    <!--          >-->
    <!--            <bk-option-->
    <!--              v-for="item in upstreamSelectOptions"-->
    <!--              :id="item.value"-->
    <!--              :key="item.value"-->
    <!--              :name="item.label"-->
    <!--            >-->
    <!--              <div v-bk-tooltips="{ content: relatedResourceTooltipContent(item), placement: 'left' }">-->
    <!--                <span>{{ item.label }}</span>-->
    <!--                <span-->
    <!--                  v-if="item.desc"-->
    <!--                  style="padding-left: 8px;color: #979ba5;"-->
    <!--                >-->
    <!--                  {{ item.desc }}-->
    <!--                </span>-->
    <!--              </div>-->
    <!--            </bk-option>-->
    <!--          </bk-select>-->
    <!--        </div>-->
    <!--      </div>-->
    <!--    </template>-->
  </table-resource-list>
  <slider-resource-viewer
    v-model="isResourceViewerShow"
    editable
    :resource="route"
    :source="source"
    resource-type="route"
    @updated="handleUpdated"
  />
</template>

<script lang="tsx" setup>
import TableResourceList from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { deleteRoute, getRoutes, getRoute } from '@/http/route';
import { getServiceDropdowns } from '@/http/service';
import { getUpstreamDropdowns } from '@/http/upstream';
import { computed, ref, shallowRef } from 'vue';
import { IRoute } from '@/types/route';
import { FilterOptionClass, type IFilterOption } from '@/types/table-filter';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import type { PrimaryTableProps, TableRowData } from '@blueking/tdesign-ui';
import { METHOD_THEMES } from '@/enum';
import TagHttpMethod from '@/components/tag-http-method.vue';

const { t } = useI18n();
const router = useRouter();

const route = ref<IRoute>();

// 关联资源传参
const relationSearchParams = ref<{ service_id?: string, upstream_id?: string }>({});
// 关联的服务 select 下拉选项
const serviceSelectOptions = ref<{ name: string, label: string, desc: string }[]>([]);
// 关联的上游 select 下拉选项
const upstreamSelectOptions = ref<{ value: string, label: string, desc: string }[]>([]);

const source = ref('');
const isResourceViewerShow = ref(false);
const tableRef = ref();
let serviceNameMap: Record<string, string> = {};
let upstreamNameMap: Record<string, string> = {};

const methodList = computed(() => Object.keys(METHOD_THEMES)
  .map(method => ({
    value: method,
    label: method,
  })));

const columns = shallowRef<PrimaryTableProps['columns']>([
  {
    title: 'ID',
    colKey: 'id',
    ellipsis: true,
  },
  {
    title: t('路径'),
    colKey: 'uris',
    ellipsis: true,
    cell: (h, { row }: TableRowData) => row.config?.uris?.join(', ') || '--',
  },
  {
    colKey: 'method',
    title: t('方法'),
    width: 130,
    cell: (h, { row }) => <TagHttpMethod methods={row.config?.methods || []} />,
    filter: {
      type: 'single',
      showConfirmAndReset: true,
      list: methodList.value,
      popupProps: {
        overlayInnerClassName: 'custom-radio-filter-wrapper',
      },
    },
  },
  {
    title: t('描述'),
    colKey: 'desc',
    ellipsis: true,
    cell: (h, { row }: TableRowData) => row.config?.desc || '--',
  },
  {
    title: t('服务'),
    colKey: 'service_id',
    ellipsis: true,
    filter: {
      type: 'single',
      showConfirmAndReset: true,
      list: getFilterOptions({ options: serviceSelectOptions.value, extra: true }),
      popupProps: {
        overlayInnerClassName: 'custom-radio-filter-wrapper',
      },
    },
    cell: (h, { row }: TableRowData) => (row.service_id
      ? <bk-button
        text theme="primary"
        onClick={() => handleRelatedResourceIdClicked({ routeName: 'service', id: row.service_id })}
      >{serviceNameMap[row.service_id]}</bk-button> : '--'),
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

const extraSearchOptions = computed(() => [
  {
    id: 'path',
    name: t('路径'),
  },
  {
    id: 'method',
    name: t('方法'),
    children: Object.keys(METHOD_THEMES)
      .map(method => ({
        name: method,
        id: method,
      })),
  },
  {
    id: 'service_id',
    name: t('服务'),
    children: getFilterOptions({
      options: serviceSelectOptions.value,
      key: 'name',
      value: 'id',
      extra: true,
    }),
  },
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

const getServiceSelectOptions = async () => {
  const response = await getServiceDropdowns();
  serviceSelectOptions.value = (response ?? []).map(item => ({
    name: item.name,
    id: item.id,
    desc: item.desc,
  }));
  const filterOptions = getFilterOptions({ options: serviceSelectOptions.value, extra: true });
  const groupCol = columns.value.find(col => ['service_id'].includes(col.colKey));
  if (groupCol) {
    groupCol.filter.list = filterOptions;
  }
  serviceNameMap = serviceSelectOptions.value.reduce<Record<string, string>>((acc, cur) => {
    acc[cur.id] = cur.name;
    return acc;
  }, {});
};
getServiceSelectOptions();

const getUpstreamSelectOptions = async () => {
  const response = await getUpstreamDropdowns();
  upstreamSelectOptions.value = (response ?? []).map(item => ({
    name: item.name,
    id: item.id,
    desc: item.desc,
  }));
  const filterOptions = getFilterOptions({ options: upstreamSelectOptions.value, extra: true });
  const groupCol = columns.value.find(col => ['upstream_id'].includes(col.colKey));
  if (groupCol) {
    groupCol.filter.list = filterOptions;
  }
  upstreamNameMap = upstreamSelectOptions.value.reduce<Record<string, string>>((acc, cur) => {
    acc[cur.id] = cur.name;
    return acc;
  }, {});
};
getUpstreamSelectOptions();

const handleTableClearFilter = () => {
  relationSearchParams.value = {};
};

const handleRelatedResourceIdClicked = ({ routeName, id }: { routeName: string, id: string }) => {
  const to = router.resolve({ name: routeName, query: { id } });
  window.open(to.href);
};

const toggleResourceViewerSlider = ({ resource }: { resource: IRoute }) => {
  route.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

const handleUpdated = async () => {
  tableRef.value!.getList({ ...relationSearchParams.value });
  route.value = await getRoute({ id: route.value.id });
};

// const relatedResourceTooltipContent = (item: {
//   label: string,
//   desc: string
// }) => (item.desc ? `${item.label}(${item.desc})` : item.label);

</script>
