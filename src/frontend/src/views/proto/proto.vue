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
    :columns="columns"
    :delete-api="deleteProto"
    :query-list-params="{ apiMethod: getProtoList }"
    :routes="{ create: 'proto-create', edit: 'proto-edit' }"
    resource-type="proto"
    :exclude-columns="['label']"
    @check-resource="toggleResourceViewerSlider"
  />
  <slider-resource-viewer
    v-model="isResourceViewerShow"
    :resource="proto"
    :source="source"
    resource-type="proto"
  />
</template>

<script lang="ts" setup>
import TableResourceList from '@/components/table-resource-list.vue';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { type PrimaryTableProps } from '@blueking/tdesign-ui';
import { ref } from 'vue';
import { deleteProto, getProtoList } from '@/http/proto';
import { IProto } from '@/types/proto';

const columns: PrimaryTableProps['columns'] = [
  {
    title: 'ID',
    colKey: 'id',
  },
];

const proto = ref<IProto>();
const source = ref('');
const isResourceViewerShow = ref(false);

const toggleResourceViewerSlider = ({ resource }: { resource: IProto }) => {
  proto.value = resource;
  source.value = JSON.stringify(resource.config);
  isResourceViewerShow.value = true;
};

</script>
