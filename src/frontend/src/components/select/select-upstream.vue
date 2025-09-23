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
      v-model="upstreamId"
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
        v-for="option in upstreamOptions"
        :id="option.id"
        :key="option.id"
        :name="option.name"
      />
    </bk-select>
    <bk-button
      v-if="showCheck" :disabled="checkDisabled || !upstreamId" class="check-btn" text theme="primary"
      @click="handleCheckClick"
    >
      {{ t('查看配置') }}
    </bk-button>
    <slider-resource-viewer
      v-if="viewerConfig.resource"
      v-model="viewerConfig.visible"
      :resource="viewerConfig.resource"
      :source="viewerConfig.source"
      resource-type="upstream"
    />
  </div>
</template>

<script lang="ts" setup>
import { onBeforeMount, ref } from 'vue';
import { getUpstream, getUpstreams } from '@/http/upstream';
import { IUpstream } from '@/types/upstream';
import { useI18n } from 'vue-i18n';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { textLengthCut } from '@/common/util';

interface IProps {
  staticOptions?: IOption[]
  showCheck?: boolean
  checkDisabled?: boolean
}

interface IOption {
  id: string,
  name: string
}

const upstreamId = defineModel<string>();

const { staticOptions = [], showCheck = true, checkDisabled = false } = defineProps<IProps>();

const emit = defineEmits<{
  change: [{ upstreamId: string, upstream: IUpstream }],
  check: [],
}>();

const { t } = useI18n();
const upstreamOptions = ref<IOption[]>([...staticOptions]);
const upstreamList = ref<IUpstream[]>([]);

const viewerConfig = ref<{ resource: IUpstream | null, source: string, visible: boolean }>({
  resource: null,
  source: '{}',
  visible: false,
});

const pagination = ref({
  current: 0,
  offset: 0,
  limit: 20,
  loading: false,
  lastLoaded: false,
});

const getOptions = async () => {
  if (pagination.value.lastLoaded) {
    return;
  }

  pagination.value.loading = true;
  pagination.value.current += 1;
  const { current, limit } = pagination.value;
  const response = await getUpstreams({
    query: {
      limit,
      offset: (current - 1) * limit,
    },
  });
  const results = response?.results || [];

  if (results.length < limit || !results.length) {
    pagination.value.lastLoaded = true;
  }

  upstreamOptions.value.push(...results.map((upstream: IUpstream) => {
    let desc = '';
    if (upstream.config?.desc) {
      desc = textLengthCut({ text: upstream.config.desc, parens: true });
    }
    return {
      id: upstream.id,
      name: `${upstream.name} ${desc}`,
    };
  }));
  upstreamList.value.push(...results);
  pagination.value.loading = false;
};

const handleChange = (value: string) => {
  const upstream = upstreamList.value.find(item => item.id === value);

  if (upstream) {
    viewerConfig.value.resource = upstream;
    const { config } = upstream;
    viewerConfig.value.source = typeof config !== 'string' ? JSON.stringify(config) : config;
  } else {
    viewerConfig.value.resource = null;
    viewerConfig.value.source = '{}';
  }

  emit('change', { upstreamId: value, upstream });
};

const handleCheckClick = async () => {
  if (!upstreamId.value) {
    return;
  }

  if (!viewerConfig.value.resource) {
    const upstream = await getUpstream({ id: upstreamId.value });
    viewerConfig.value.resource = upstream;
    const { config } = upstream;
    viewerConfig.value.source = typeof config !== 'string' ? JSON.stringify(config) : config;
  }

  viewerConfig.value.visible = true;
  // emit('check');
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

  .check-btn {
    position: absolute;
    left: 652px;
  }
}

</style>
