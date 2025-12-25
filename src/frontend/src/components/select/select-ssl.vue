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
  <div class="wrapper">
    <bk-select
      v-model="certId"
      :clearable="false"
      :filterable="false"
      :scroll-loading="pagination.loading"
      class="select-component"
      v-bind="$attrs"
      @change="handleChange"
      @scroll-end="getOptions"
    >
      <slot></slot>
      <bk-option
        v-for="option in certOptions"
        :id="option.id"
        :key="option.id"
        :name="option.name"
      />
    </bk-select>
  </div>
</template>

<script lang="ts" setup>
import { onBeforeMount, ref } from 'vue';
import { getSSLDropdownList } from '@/http/ssl';

interface IProps {
  staticOptions?: IOption[]
  showCheck?: boolean
  checkDisabled?: boolean
}

interface IOption {
  id: string,
  name: string
}

const certId = defineModel<string>();

const { staticOptions = [] } = defineProps<IProps>();

const emit = defineEmits<{
  change: [{ certId: string }],
  check: [],
}>();

const certOptions = ref<IOption[]>([...staticOptions]);

const pagination = ref({
  current: 0,
  offset: 0,
  limit: 100,
  loading: false,
  lastLoaded: false,
});

const getOptions = async () => {
  const response = await getSSLDropdownList({});
  const results = response || [];

  certOptions.value = results.map(cert => ({
    id: cert.id,
    name: cert.name,
  }));
};

const handleChange = (certId: string) => {
  emit('change', { certId });
};

onBeforeMount(async () => {
  await getOptions();
});

</script>

<style lang="scss" scoped>

.wrapper {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;

  .select-component {
    flex-grow: 1;
  }
}

</style>
