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
    width="960"
    @closed="handleClose"
  >
    <template #header>
      <div class="custom-header-wrapper">
        <div class="resource-type">{{ displayedResourceType }}</div>
        <template v-if="resource.name">
          <div class="divider"></div>
          <div class="resource-id">{{ resource.name }}</div>
        </template>
      </div>
    </template>
    <template #default>
      <div class="content-wrapper">
        <bk-tab
          v-model:active="activeTabName"
          type="card-grid"
        >
          <bk-tab-panel
            :label="t('列表')"
            name="list"
          >
            <div class="tab-panel-content">
              <component :is="listDisplayComponents[resourceType]" :resource="resource"></component>
            </div>
          </bk-tab-panel>
          <bk-tab-panel
            :label="t('源码模式')"
            name="source"
          >
            <div class="tab-panel-content">
              <div class="actions">
                <!--<bk-select v-model="sourceLanguage" :clearable="false" :filterable="false" style="width: 100px;">-->
                <!--            <bk-option id="json" name="JSON" />-->
                <!--            <bk-option id="yaml" name="YAML" />-->
                <!--          </bk-select>-->
                <button-icon icon="copy" @click="handleCopy">{{ t('复制') }}</button-icon>
                <button-icon v-if="showFormat" icon="geshihua" @click="handleFormat">{{ t('格式化') }}</button-icon>
                <button-icon
                  v-if="editable"
                  :disabled="common.curGatewayData?.read_only || isEditing"
                  icon="edit-line"
                  theme="primary"
                  @click="handleStartEdit"
                >{{
                  t('编辑')
                }}
                </button-icon>
              </div>
              <div class="editor">
                <monaco-editor
                  v-if="activeTabName === 'source'"
                  ref="monaco-editor"
                  :height="editorHeight"
                  :source="source"
                  :id="mountOn"
                  :read-only="!editable || !isEditing || common.curGatewayData?.read_only"
                  @change="handleEditorChange"
                />
              </div>
            </div>
          </bk-tab-panel>
        </bk-tab>
      </div>
    </template>
    <template v-if="isEditing" #footer>
      <div class="footer-actions">
        <bk-button theme="primary" @click="handleSaveEdit">{{ t('保存') }}</bk-button>
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
import { useClipboard, useWindowSize } from '@vueuse/core';
import { Component, computed, ref, useTemplateRef, watch } from 'vue';
import ListDisplayUpstream from '@/components/list-display/list-display-upstream.vue';
import ListDisplayService from '@/components/list-display/list-display-service.vue';
import { IStreamRoute, ResourceType } from '@/types';
import ListDisplayRoute from '@/components/list-display/list-display-route.vue';
import ListDisplayStreamRoute from '@/components/list-display/list-display-stream-route.vue';
import ListDisplayConsumer from '@/components/list-display/list-display-consumer.vue';
import ListDisplayConsumerGroup from '@/components/list-display/list-display-consumer-group.vue';
import ListDisplayPluginMetadata from '@/components/list-display/list-display-plugin-metadata.vue';
import ListDisplayGlobalRules from '@/components/list-display/list-display-global-rules.vue';
import ListDisplayPluginConfigs from '@/components/list-display/list-display-plugin-configs.vue';
import ListDisplaySSL from '@/components/list-display/list-display-ssl.vue';
import ListDisplayProto from '@/components/list-display/list-display-proto.vue';
import { useCommon } from '@/store';
import useJsonTransformer from '@/hooks/use-json-transformer';
import { putRoute } from '@/http/route';
import { putService } from '@/http/service';
import { putUpstream } from '@/http/upstream';
import { IRoute } from '@/types/route';
import { putStreamRoute } from '@/http/stream-route';

interface IProps {
  source?: string
  resource?: ResourceType
  resourceType: string
  showFormat?: boolean  // 是否展示格式化按钮
  mountOn?: string   // monaco editor 挂载的元素 ID
  editable?: boolean  // 是否可编辑
}

const isShow = defineModel<boolean>({
  required: true,
  default: false,
});

const {
  source = '{}',
  resource,
  resourceType,
  showFormat = false,
  mountOn = 'resource-viewer-editor',
  editable = false,
} = defineProps<IProps>();

const emit = defineEmits<{
  updated: [void];
}>();

const { t } = useI18n();
const common = useCommon();
const { height: windowHeight } = useWindowSize();
const { formatJSON } = useJsonTransformer();

const listDisplayComponents: { [key: string]: Component } = {
  upstream: ListDisplayUpstream,
  service: ListDisplayService,
  route: ListDisplayRoute,
  consumer: ListDisplayConsumer,
  consumer_group: ListDisplayConsumerGroup,
  plugin_metadata: ListDisplayPluginMetadata,
  global_rule: ListDisplayGlobalRules,
  plugin_config: ListDisplayPluginConfigs,
  ssl: ListDisplaySSL,
  proto: ListDisplayProto,
  stream_route: ListDisplayStreamRoute,
};

const activeTabName = ref<'list' | 'source'>('list');
const editor = useTemplateRef('monaco-editor');
const localSource = ref('');
const sourceLanguage = ref('json');
const isEditing = ref(false);

const displayedResourceType = computed(() => {
  return common.enums?.resource_type[resourceType];
});

const editorHeight = computed(() => {
  const footerHeight = isEditing.value ? 32 : 0;
  const height = windowHeight.value - footerHeight - 228;
  return height >= 540 ? height : 540;
});

watch(sourceLanguage, () => {
  editor.value?.setLanguage(sourceLanguage.value);
});

watch(activeTabName, (tab: string) => {
  if (editor.value && ['source'].includes(tab)) {
    editor.value?.initEditor();
  }
});

watch(
  () => source.value,
  (newSource) => {
    if (['source'].includes(activeTabName.value) && editor.value) {
      editor.value.setValue(newSource);
    }
  },
);

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

const handleStartEdit = () => {
  if (common.curGatewayData?.read_only) {
    return;
  }
  isEditing.value = true;
};

const handleCancelClick = () => {
  isEditing.value = false;
};

const handleClose = () => {
  sourceLanguage.value = 'json';
  activeTabName.value = 'list';
  editor.value?.setValue(formatJSON({ source }));
  localSource.value = source;
  isEditing.value = false;
};

const handleFormat = () => {
  editor.value?.format();
};

const handleSaveEdit = async () => {
  let config: any = {};
  try {
    config = JSON.parse(localSource.value);
  } catch {
    return;
  }
  if (resourceType === 'route') {
    await putRoute({
      data: {
        config,
        name: config.name,
        upstream_id: (resource as IRoute).upstream_id || '',
        service_id: (resource as IRoute).service_id || '',
        plugin_config_id: (resource as IRoute).plugin_config_id || '',
      },
      id: resource.id,
    });
  } else if (resourceType === 'stream_route') {
    await putStreamRoute({
      data: {
        config,
        name: config.name,
        upstream_id: (resource as IStreamRoute).upstream_id || '',
        service_id: (resource as IStreamRoute).service_id || '',
        plugin_config_id: (resource as IStreamRoute).plugin_config_id || '',
      },
      id: resource.id,
    });
  } else if (resourceType === 'service') {
    const data = {
      name: config.name,
    };
    if (config.upstream_id) {
      Object.assign(data, { upstream_id: config.upstream_id });
    }
    await putService({
      data: {
        config,
        ...data,
      },
      id: resource.id,
    });
  } else if (resourceType === 'upstream') {
    const data = {
      name: config.name,
    };
    if (config.tls?.client_cert_id) {
      Object.assign(data, { ssl_id: config.tls.client_cert_id });
    }
    await putUpstream({
      data: {
        config,
        ...data,
      },
      id: resource.id,
    });
  }
  Message({
    theme: 'success',
    message: t('提交成功'),
  });
  emit('updated');
  isEditing.value = false;
};

</script>

<style lang="scss" scoped>

:deep(.bk-sideslider-content) {
  background-color: #f5f7fa;
}

:deep(.bk-tab-header-nav) {
  margin-left: 40px;
}

:deep(.bk-tab-content) {
  padding: 0;
  box-shadow: none;
}

.custom-header-wrapper {
  display: flex;
  align-items: center;

  .divider {
    width: 1px;
    height: 14px;
    margin-right: 8px;
    margin-left: 8px;
    background: #dcdee5;
  }

  .resource-id {
    font-size: 14px;
    color: #979ba5;
  }
}

.content-wrapper {
  font-size: 14px;
  padding-top: 24px;

  .tab-panel-content {
    padding: 24px 32px 0 40px;
  }

  .actions {
    display: flex;
    margin-bottom: 24px;
    gap: 12px;
  }
}

.footer-actions {
  display: flex;
  padding-left: 16px;
  gap: 12px;
}

</style>
