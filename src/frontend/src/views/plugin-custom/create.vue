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
      <div>
        <form-card>
          <template #title>{{ t('基本信息') }}</template>
          <div>
            <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
              <bk-form-item :label="t('名称')" class="form-item" property="name" required>
                <bk-input v-model="formModel.name" clearable />
              </bk-form-item>
            </bk-form>
          </div>
        </form-card>

        <form-card>
          <template #title>{{ t('配置') }}</template>
          <div class="config-wrapper">
            <bk-form ref="schema-form-ref" :model="formModel" :rules="schemaRules" class="form-element config-item">
              <bk-form-item label="schema" class="form-item config-content icon-required" property="schema">
                <div>
                  <button-icon class="mb10" icon="geshihua" @click="handleSchemaFormat">{{ t('格式化') }}</button-icon>
                  <monaco-editor
                    ref="monaco-editor-schema"
                    id="monaco-editor-schema"
                    :height="editorHeight"
                    :source="schemaSource"
                    @change="handleSchemaChange"
                  />
                </div>
              </bk-form-item>
            </bk-form>

            <bk-form ref="example-form-ref" :model="formModel" :rules="exampleRules" class="form-element config-item">
              <bk-form-item label="example" class="form-item config-content icon-required" property="example">
                <div>
                  <button-icon class="mb10" icon="geshihua" @click="handleExampleFormat">{{ t('格式化') }}</button-icon>
                  <monaco-editor
                    ref="monaco-editor-example"
                    id="monaco-editor-example"
                    :height="editorHeight"
                    :source="exampleSource"
                    @change="handleExampleChange"
                  />
                </div>
              </bk-form-item>
            </bk-form>
          </div>
        </form-card>
      </div>
    </main>
    <form-page-footer @cancel="handleCancelClick" @submit="handleSubmit" />
  </div>
</template>

<script lang="ts" setup>
import FormPageFooter from '@/components/form/form-page-footer.vue';
import { Form, InfoBox, Message } from 'bkui-vue';
import MonacoEditor from '@/components/monaco-editor.vue';
import { IPluginCustomConfig, IPluginCustomDto } from '@/types/plugin-custom';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, useTemplateRef, watch } from 'vue';
import {
  getPluginCustom,
  postPluginCustom,
  putPluginCustom,
} from '@/http/plugin-custom';
import ButtonIcon from '@/components/button-icon.vue';
import FormCard from '@/components/form-card.vue';
import { useElementSize } from '@vueuse/core';
import Ajv from 'ajv';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const ajv = new Ajv();
const validateSchema = ajv.getSchema('http://json-schema.org/draft-07/schema#');

const { height } = useElementSize(document.body);

const formModel = ref<Omit<IPluginCustomDto, 'config'>>({
  name: '',
  example: {},
  schema: {},
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const schemaFormRef = useTemplateRef<InstanceType<typeof Form>>('schema-form-ref');
const exampleFormRef = useTemplateRef<InstanceType<typeof Form>>('example-form-ref');
const exampleEditor = useTemplateRef('monaco-editor-example');
const schemaEditor = useTemplateRef('monaco-editor-schema');
const exampleConfig = ref<IPluginCustomConfig>({});
const schemaConfig = ref<IPluginCustomConfig>({});
const exampleSource = ref('{}');
const schemaSource = ref('{}');

const rules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'blur' },
  ],
};

const schemaRules = {
  schema: [
    {
      validator: () => {
        return !!Object.keys(schemaConfig.value)?.length;
      },
      message: t('必填项'),
      trigger: 'blur',
    },
  ],
};

const exampleRules = {
  example: [
    {
      validator: () => {
        return !!Object.keys(exampleConfig.value)?.length;
      },
      message: t('必填项'),
      trigger: 'blur',
    },
  ],
};

const pluginCustomId = computed(() => {
  return route.params.id as string;
});

const isEditMode = computed(() => {
  return !!route.params.id;
});

const editorHeight = computed(() => {
  return height.value - 466;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getPluginCustom({ id } as { id: string });
    const { example, schema, ...rest } = response;

    formModel.value.name = rest.name;
    exampleSource.value = JSON.stringify(example);
    schemaSource.value = JSON.stringify(schema);
  }
}, { immediate: true });

const handleExampleChange = ({ source }: { source: string }) => {
  try {
    exampleConfig.value = JSON.parse(source);
  } catch {
    exampleConfig.value = {};
  }
};

const handleSchemaChange = ({ source }: { source: string }) => {
  try {
    schemaConfig.value = JSON.parse(source);
  } catch {
    schemaConfig.value = {};
  }
};

const handleSubmit = async () => {
  try {
    // 校验表单
    await Promise.all([
      formRef.value?.validate(),
      schemaFormRef.value?.validate(),
      exampleFormRef.value?.validate(),
    ]);

    // 验证 schema 是否为有效的 JSON Schema
    if (!validateSchema(schemaConfig.value)) {
      Message({
        theme: 'error',
        message: t('提供的 schema 不是有效的 JSON Schema'),
      });
      return;
    }

    // 使用 schema 验证 example
    const validateExample = ajv.compile(schemaConfig.value);
    if (!validateExample(exampleConfig.value)) {
      Message({
        theme: 'error',
        message: t('example 不符合 schema 格式'),
      });
      return;
    }

    const data = {
      name: formModel.value.name,
      example: exampleConfig.value,
      schema: schemaConfig.value,
    };

    InfoBox({
      title: t('确认提交？'),
      confirmText: t('提交'),
      cancelText: t('取消'),
      onConfirm: async () => {
        if (isEditMode.value) {
          await putPluginCustom({
            data,
            id: pluginCustomId.value,
          });
        } else {
          await postPluginCustom({ data });
        }

        Message({
          theme: 'success',
          message: t('提交成功'),
        });

        await router.push({ name: 'plugin-custom', replace: true });
      },
    });
  } catch (e) {
    const error = e as Error;
    Message({
      theme: 'error',
      message: error.message || t('提交失败'),
    });
  }
};

const handleExampleFormat = () => {
  exampleEditor.value?.format();
};

const handleSchemaFormat = () => {
  schemaEditor.value?.format();
};

const handleCancelClick = () => {
  router.back();
};

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
}

.config-wrapper {
  width: 100%;
  display: flex;

  .config-item {
    flex: 1;
    display: flex;

    .config-content {
      width: 100%;
    }
  }
}

.icon-required {
  :deep(.bk-form-label::after) {
    position: absolute;
    top: 0;
    width: 14px;
    color: #ea3636;
    text-align: center;
    content: "*";
  }
}

</style>
