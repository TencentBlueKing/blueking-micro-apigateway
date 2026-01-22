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
    :delete-api="deleteSSL"
    :extra-search-params="searchParams"
    :columns="columns"
    :exclude-columns="['label']"
    :query-list-params="{ apiMethod: getSSLList }"
    :routes="{ create: 'ssl-create', edit: 'ssl-edit' }"
    :selection-column="[]"
    resource-type="ssl"
    @check-resource="toggleResourceViewerSlider"
  />
  <slider-resource-viewer
    v-model="isResourceViewerShow"
    :resource="ssl"
    :source="source"
    resource-type="ssl"
  />
</template>

<script lang="tsx" setup>
import TableResourceList, { ISearchParam } from '@/components/table-resource-list.vue';
import { deleteSSL, getSSL, getSSLList } from '@/http/ssl';
import { type PrimaryTableProps, type TableRowData } from '@blueking/tdesign-ui';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { ref, watch } from 'vue';
import { ISSL } from '@/types/ssl';
import { useRoute } from 'vue-router';

const { t } = useI18n();
const route = useRoute();

const columns: PrimaryTableProps['columns'] = [
  {
    title: 'ID',
    colKey: 'id',
  },
  {
    title: 'SNIS',
    colKey: 'snis',
    cell: (h, { row }: TableRowData) => {
      return row.config?.snis?.map((sni: string) => sni)
        .join(', ') || '--';
    },
  },
  {
    title: t('过期时间'),
    colKey: 'validity_end',
    ellipsis: true,
    cell: (h, { row }: TableRowData) => {
      return dayjs.unix(row.config?.validity_end as number)
        .format('YYYY-MM-DD HH:mm:ss Z');
    },
  },
  {
    title: t('有效期'),
    colKey: 'expiration_end',
    ellipsis: true,
    cell: (h, { row }: TableRowData) => {
      const ts = row.config?.validity_end as number | undefined;
      if (!ts) return '--';
      const expiry = dayjs.unix(ts);
      const now = dayjs();
      const ms = expiry.valueOf() - now.valueOf();
      const days = Math.ceil(ms / 86400000);
      if (ms <= 0) {
        return <span style={{ color: '#9e9e9e' }}>{t('已过期')}</span>;
      }
      if (days < 7) {
        return <span style={{ color: '#f44336' }}>{days} {t('天')}</span>;
      }
      if (days < 30) {
        return <span style={{ color: '#ff9800' }}>{days} {t('天')}</span>;
      }
      return <span>{days} {t('天')}</span>;
    },
  },
];

const ssl = ref<ISSL>();
const source = ref('');
const isResourceViewerShow = ref(false);
const searchParams = ref<ISearchParam[]>([]);

const toggleResourceViewerSlider = ({ resource }: { resource: ISSL }) => {
  ssl.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

watch(() => route.query.id, async () => {
  if (route.query.id) {
    const id = route.query.id as string;
    const resource = await getSSL({ id });
    toggleResourceViewerSlider({ resource });
    searchParams.value = [{
      id: 'id',
      name: 'ID',
      values: [{ id, name: id }],
    }];
  }
}, { immediate: true });

</script>
