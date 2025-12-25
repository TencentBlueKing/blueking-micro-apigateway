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
  <div>
    <main class="page-content-wrapper">
      <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
        <bk-form-item :label="t('选择插件')" class="form-item" property="name" required label-width="80">
          <bk-select
            v-model="formModel.name"
            :clearable="false"
            :disabled="isEditMode"
            filterable
            style="width: 350px"
            @change="handlePluginChange"
          >
            <bk-option
              v-for="type in allPluginList"
              :id="type.name"
              :key="type.name"
              :disabled="existedMetadataPluginNameList?.includes(type.name)"
              :name="type.name"
            />
          </bk-select>
        </bk-form-item>
      </bk-form>
      <!-- 配置 -->
      <div class="editor-wrapper">
        <div class="editor-top-bar flex-row align-items-center justify-content-between">
          <div class="plugin-name">
            {{ formModel.name }}
          </div>
          <div class="flex-row align-items-center">
            <span class="line"></span>
            <div class="format-btn" @click="handleFormat">
              <i class="icon apigateway-icon icon-ag-geshihua"></i>
              {{ t('格式化') }}
            </div>
          </div>
        </div>
        <monaco-editor
          ref="monaco-editor" :height="editorHeight" :source="initialSource" @change="handleEditorChange"
        />
      </div>
    </main>
    <form-page-footer @cancel="handleCancelClick" @submit="handleSubmit" />
  </div>
</template>

<script lang="ts" setup>
import FormPageFooter from '@/components/form/form-page-footer.vue';
import { Form, InfoBox, Message } from 'bkui-vue';
import MonacoEditor from '@/components/monaco-editor.vue';
import { IPluginMetadataConfig, IPluginMetadataDto } from '@/types/plugin-metadata';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, onBeforeMount, ref, useTemplateRef, watch } from 'vue';
import { useCommon } from '@/store';
import {
  getPluginMetadata,
  getPluginMetadataDropdownList,
  postPluginMetadata,
  putPluginMetadata,
} from '@/http/plugin-metadata';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
import Ajv from 'ajv';
import { getMetadataPlugins, getMetadataPluginSchema } from '@/http/plugins';
import { IPlugin } from '@/types/plugin';
import { useElementSize } from '@vueuse/core';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const common = useCommon();
const { showSchemaErrorMessages } = useSchemaErrorMessage();
const { height } = useElementSize(document.body);

const formModel = ref<Omit<IPluginMetadataDto, 'config'>>({
  name: '',
});

const pluginConfig = ref<IPluginMetadataConfig>({});
const initialSource = ref('{}');
const schema = ref<Record<string, any>>({});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const editor = useTemplateRef('monaco-editor');

const rules = {
  name: [
    { required: true, message: t('必选项'), trigger: 'blur' },
  ],
};

// 完整的插件列表
// const allPluginList = computed(() => {
//   return common.plugins || [];
// });
const allPluginList = ref<IPlugin[]>([]);
const existedMetadataPluginNameList = ref<string[]>([]);

const pluginMetadataId = computed(() => {
  return route.params.id as string;
});

const isEditMode = computed(() => {
  return !!route.params.id;
});

const editorHeight = computed(() => {
  return height.value - 392;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getPluginMetadata({ id } as { id: string });
    const { config, ...rest } = response;
    formModel.value.name = rest.name;
    schema.value = await getMetadataPluginSchema({ name: rest.name, gatewayId: common.gatewayId });
    initialSource.value = JSON.stringify(config);
  }
}, { immediate: true });

const handleEditorChange = ({ source }: { source: string }) => {
  try {
    pluginConfig.value = JSON.parse(source);
  } catch {
    pluginConfig.value = {};
  }
};

const handleSubmit = async () => {
  try {
    // 校验表单
    await Promise.all([
      formRef.value?.validate(),
    ]);

    // 校验 schema
    const ajv = new Ajv();
    const schemaValidate = ajv.compile(schema.value);

    const config = {
      ...pluginConfig.value,
    };

    if (schemaValidate(config)) {
      const data = {
        config,
        name: formModel.value.name,
      };

      InfoBox({
        title: t('确认提交？'),
        confirmText: t('提交'),
        cancelText: t('取消'),
        onConfirm: async () => {
          if (isEditMode.value) {
            await putPluginMetadata({
              data,
              id: pluginMetadataId.value,
            });
          } else {
            await postPluginMetadata({ data });
          }

          Message({
            theme: 'success',
            message: t('提交成功'),
          });

          await router.push({ name: 'plugin-metadata', replace: true });
        },
      });
    } else {
      showSchemaErrorMessages(schemaValidate.errors);
    }
  } catch (e) {
    const error = e as Error;
    Message({
      theme: 'error',
      message: error.message || t('提交失败'),
    });
  }
};

const handlePluginChange = async (pluginName: string) => {
  const plugin = allPluginList.value.find(plugin => plugin.name === pluginName);

  if (plugin) {
    schema.value = await getMetadataPluginSchema({ name: pluginName, gatewayId: common.gatewayId });
    initialSource.value = JSON.stringify(plugin.metadata_example || {});
  } else {
    schema.value = {};
    initialSource.value = '{}';
  }
};

const handleFormat = () => {
  editor.value?.format();
};

const handleCancelClick = () => {
  router.back();
};

onBeforeMount(async () => {
  const [pluginRes, metadataRes] = await Promise.all([
    getMetadataPlugins({ gatewayId: common.gatewayId }),
    getPluginMetadataDropdownList({ gatewayId: common.gatewayId }),
  ]);
  allPluginList.value = pluginRes.reduce((acc, cur) => {
    acc = [...acc, ...cur.plugins];
    return acc;
  }, [] as IPlugin[]);
  existedMetadataPluginNameList.value = metadataRes?.map(item => item.name);
});

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
}

.editor-top-bar {
  height: 40px;
  background: #2E2E2E;
  box-shadow: 0 2px 4px 0 #00000029;
  border-radius: 2px 2px 0 0;
  padding: 0 24px;
  color: #C4C6CC;
  font-size: 14px;
  .line {
    width: 1px;
    height: 14px;
    background: #45464D;
    margin-right: 12px;
  }
  .format-btn {
    cursor: pointer;
    &:hover {
      color: #409CFF;
    }
  }
}

</style>
