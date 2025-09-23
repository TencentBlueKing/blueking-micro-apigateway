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
  <div class="list-wrapper">
    <div class="enabled-plugins">
      <form-divider>已启用</form-divider>
      <div v-if="enabledPluginList.length" class="plugin-items">
        <div v-for="plugin in enabledPluginList" :key="plugin.name" class="plugin-item">
          <div class="plugin-icon">
            <Icon color="#cccccc" name="micro-plugin-generic" size="32" />
          </div>
          <div class="plugin-name">{{ plugin.name }}</div>
          <div class="plugin-action">
            <bk-button @click="showEditSlider({ name: plugin.name, config: plugin.config })">编辑</bk-button>
          </div>
        </div>
      </div>
      <div v-else>{{ t('未启用插件') }}</div>

      <form-divider>未启用</form-divider>
      <div class="plugin-items">
        <div v-for="plugin in pluginOptions" :key="plugin.name" class="plugin-item">
          <div class="plugin-icon">
            <Icon color="#cccccc" name="micro-plugin-generic" size="32" />
          </div>
          <div class="plugin-name">{{ plugin.name }}</div>
          <div class="plugin-action">
            <bk-button theme="primary" @click="showEditSlider({ name: plugin.name })">启用</bk-button>
          </div>
        </div>
      </div>
    </div>
    <slider-plugin-config-editor v-model="isConfigSliderShow" :plugin="editingPlugin" @confirm="handleEditConfirm" />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useCommon } from '@/store';
import FormDivider from '@/components/form-divider.vue';
import Icon from '@/components/icon.vue';
import SliderPluginConfigEditor from '@/components/slider-plugin-config-editor.vue';
import { Message } from 'bkui-vue';
import Ajv from 'ajv';
import addFormats from 'ajv-formats';

const enabledPluginList = defineModel<ILocalPlugin[]>({
  required: true,
});

const ajv = new Ajv();
addFormats(ajv);

interface ILocalPlugin {
  id?: string
  name: string
  config: string
  enabled?: boolean
}

const { t } = useI18n();
const common = useCommon();

const isConfigSliderShow = ref(false);

const editingPlugin = ref<ILocalPlugin>({
  name: '',
  config: `{

}`,
});

const schema = computed(() => {
  return allPluginList.value.find(plugin => plugin.name === editingPlugin.value.name)?.schema;
});

// 完整的插件列表
const allPluginList = computed(() => {
  return common.plugins || [];
});

// 可启用的插件列表
const pluginOptions = computed(() => {
  return allPluginList.value.filter(plugin => !enabledPluginList.value.map(item => item.name)
    .includes(plugin.name));
});

const showEditSlider = ({ name, config }: { name: string, config?: string }) => {
  editingPlugin.value = {
    name,
    config: config || `{

}`,
  };
  isConfigSliderShow.value = true;
};

const handleEditConfirm = (plugin: ILocalPlugin) => {
  try {
    // 校验 schema
    const schemaValidate = ajv.compile(schema.value);
    if (schemaValidate(JSON.parse(plugin.config))) {
      const targetPlugin = enabledPluginList.value.find(item => item.name === plugin.name);

      if (targetPlugin) {
        targetPlugin.config = plugin.config;
      } else {
        enabledPluginList.value.push(plugin);
      }

      isConfigSliderShow.value = false;
      Message({
        theme: 'success',
        message: t('配置成功'),
      });
    } else {
      throw Error(t('配置不正确'));
    }
  } catch (err) {
    const error = err as Error;
    Message({
      theme: 'error',
      message: error.message || t('配置不正确'),
    });
  }
};

</script>

<style lang="scss" scoped>

.list-wrapper {
  padding: 24px;
  background-color: #ffffff;

  .enabled-plugins {
    .plugin-items {
      display: flex;
      flex-wrap: wrap;
      gap: 24px;

      .plugin-item {
        display: flex;
        align-items: center;
        width: 320px;
        border: 1px solid var(--border-color);

        .plugin-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 48px;
          height: 48px;
          border-right: 1px solid var(--border-color);
        }

        .plugin-name {
          font-weight: 500;
          flex-grow: 1;
          border-right: 1px solid var(--border-color);
          padding-inline: 12px;
        }

        .plugin-action {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 72px;
        }
      }
    }
  }
}

</style>
