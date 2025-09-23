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
  <bk-sideslider
    v-model:is-show="isShow"
    :title="`${t('插件')}: ${localPlugin.name}`"
    width="960"
    @close="resetEditor"
  >
    <template #default>
      <div class="content-wrapper">
        <div class="config-content-wrapper">
          <div class="actions">
            <!--   <bk-select v-model="sourceLanguage" :clearable="false" :filterable="false" style="width: 100px;">-->
            <!--              <bk-option id="json" name="JSON" />-->
            <!--              <bk-option id="yaml" name="YAML" />-->
            <!--            </bk-select>-->
            <button-icon icon="copy" @click="handleCopy">{{ t('复制') }}</button-icon>
            <button-icon icon="geshihua" @click="handleFormat">{{ t('格式化') }}</button-icon>
            <button-icon
              v-if="localPlugin.type !== 'customize plugin'"
              icon="link"
              @click="handleDocClick"
            >{{ t('文档') }}
            </button-icon>
          </div>
          <div class="editor">
            <monaco-editor
              :id="mountOn" ref="monaco-editor" :height="700" :source="localPlugin.config" @change="handleEditorChange"
            />
          </div>
        </div>
      </div>
    </template>
    <template #footer>
      <div class="footer-actions">
        <bk-button theme="primary" @click="handleConfirmClick">{{ t('保存') }}</bk-button>
        <bk-button @click="handleCancelClick">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<script lang="ts" setup>
import MonacoEditor from '@/components/monaco-editor.vue';
import ButtonIcon from '@/components/button-icon.vue';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useClipboard } from '@vueuse/core';
import { ref, toValue, useTemplateRef, watch } from 'vue';
import { cloneDeep } from 'lodash-es';
import { useCommon } from '@/store';

interface ILocalPlugin {
  doc_url?: string
  example?: string
  id?: string
  name: string
  config: string
  enabled?: boolean
  type: string
}

interface IProps {
  plugin: ILocalPlugin
  mountOn?: string   // monaco editor 挂载的元素 ID
}

const isShow = defineModel<boolean>({
  required: true,
  default: false,
});

const { plugin, mountOn = 'plugin-config-editor' } = defineProps<IProps>();

const emits = defineEmits<{
  (e: 'confirm', plugin: ILocalPlugin): void
}>();

const { t } = useI18n();
const common = useCommon();

const localPlugin = ref(toValue(plugin));
const localSource = ref('');
const sourceLanguage = ref('json');
const editor = useTemplateRef('monaco-editor');

watch(() => plugin, () => {
  localPlugin.value = cloneDeep(plugin);
}, { deep: true });

watch(sourceLanguage, () => {
  editor.value?.setLanguage(sourceLanguage.value);
});

const handleEditorChange = ({ source }: { source: string }) => {
  localSource.value = source;
};

const handleCopy = () => {
  const { copy, isSupported } = useClipboard({ source: localSource.value, legacy: true });

  if (isSupported.value) {
    copy(localSource.value);
    Message({
      theme: 'success',
      message: t('已复制'),
    });
  } else {
    Message({
      theme: 'warning',
      message: t('复制失败，未开启剪贴板权限'),
    });
  }
};

const handleFormat = () => {
  editor.value?.format();
};

const handleDocClick = () => {
  if (plugin.doc_url) {
    window.open(plugin.doc_url);
  } else {
    const plugin = common.plugins.find(item => item.name === localPlugin.value.name);
    window.open(plugin?.doc_url || '');
  }
};

const handleConfirmClick = async () => {
  emits('confirm', { ...localPlugin.value, config: localSource.value });
};

const handleCancelClick = () => {
  resetEditor();
  isShow.value = false;
};

const resetEditor = () => {
  localPlugin.value = {
    name: '',
    config: '{}',
    type: '',
  };
  sourceLanguage.value = 'json';
};

</script>

<style lang="scss" scoped>

.content-wrapper {
  font-size: 14px;
  padding: 24px;

  .actions {
    display: flex;
    margin-bottom: 24px;
    gap: 12px;
  }
}

.footer-actions {
  display: flex;
  gap: 12px;
}

</style>
