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
      v-model="pluginConfigId"
      clearable
      filterable
      class="select-component"
      v-bind="$attrs"
      @change="handleChange"
    >
      <slot></slot>
      <bk-option
        v-for="option in pluginConfigOptions"
        :id="option.id"
        :key="option.id"
        :name="option.name"
      />
    </bk-select>
    <bk-button
      v-if="showCheck" :disabled="checkDisabled || !pluginConfigId" class="check-btn" text theme="primary"
      @click="handleCheckClick"
    >
      {{ t('查看配置') }}
    </bk-button>
    <slider-resource-viewer
      v-if="viewerConfig.resource"
      v-model="viewerConfig.visible"
      :resource="viewerConfig.resource"
      :source="viewerConfig.source"
      resource-type="plugin_config"
    />
  </div>
</template>

<script lang="ts" setup>
import { onBeforeMount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { IPluginConfigDto } from '@/types/plugin-config';
import { getPluginConfig, getPluginConfigs } from '@/http/plugin-config';
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

const pluginConfigId = defineModel<string>();

const { staticOptions = [], showCheck = true, checkDisabled = false } = defineProps<IProps>();

const emit = defineEmits<{
  change: [{ pluginConfigId: string, pluginConfig: IPluginConfigDto }],
  check: [],
}>();

const { t } = useI18n();
const pluginConfigOptions = ref<IOption[]>([...staticOptions]);
const pluginConfigList = ref<IPluginConfigDto[]>([]);

const viewerConfig = ref<{ resource: IPluginConfigDto | null, source: string, visible: boolean }>({
  resource: null,
  source: '{}',
  visible: false,
});

const getOptions = async () => {
  const response = await getPluginConfigs({ query: { limit: 100, offset: 0 } });
  const results = response?.results || [];
  pluginConfigOptions.value.push(...results.map((pluginConfig: IPluginConfigDto) => {
    let desc = '';
    if (pluginConfig.config?.desc) {
      desc = textLengthCut({ text: pluginConfig.config.desc, parens: true });
    }
    return {
      id: pluginConfig.id,
      name: `${pluginConfig.name} ${desc}`,
    };
  }));
  pluginConfigList.value.push(...results);
};

const handleChange = (value: string) => {
  const pluginConfig = pluginConfigList.value.find(item => item.id === value);

  if (pluginConfig) {
    viewerConfig.value.resource = pluginConfig;
    const { config } = pluginConfig;
    viewerConfig.value.source = typeof config !== 'string' ? JSON.stringify(config) : config;
  } else {
    viewerConfig.value.resource = null;
    viewerConfig.value.source = '{}';
  }

  emit('change', { pluginConfigId: value, pluginConfig });
};

const handleCheckClick = async () => {
  if (!viewerConfig.value.resource) {
    const pluginConfig = await getPluginConfig({ id: pluginConfigId.value });
    viewerConfig.value.resource = pluginConfig;
    const { config } = pluginConfig;
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
    width: 640px;
  }

  .check-btn {
    position: absolute;
    left: 652px;
  }
}

</style>
