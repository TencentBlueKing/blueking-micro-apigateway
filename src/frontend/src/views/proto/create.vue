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
      <form-card>
        <template #title>{{ t('基本信息') }}</template>
        <div>
          <bk-form ref="form-ref-base" :model="formModel" :rules="baseRules" class="form-element">
            <!-- name -->
            <bk-form-item :label="t('名称')" class="form-item" property="name" required>
              <bk-input v-model="formModel.name" clearable />
            </bk-form-item>
          </bk-form>
        </div>
      </form-card>

      <form-card>
        <template #title>{{ t('config 配置') }}</template>
        <div>
          <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
            <bk-form-item :label="t('desc')" class="form-item" property="config.desc">
              <bk-input v-model="formModel.config.desc" clearable />
            </bk-form-item>

            <bk-form-item :label="t('content')" class="form-item icon-required" property="config.content">
              <div class="editor-wrapper">
                <monaco-editor
                  ref="monaco-editor"
                  language="yaml"
                  :height="editorHeight"
                  :source="initialSource"
                  @change="handleEditorChange"
                />
              </div>
            </bk-form-item>
          </bk-form>
        </div>
      </form-card>
    </main>
    <form-page-footer @cancel="handleCancelClick" @submit="handleSubmit" />
  </div>
</template>

<script lang="ts" setup>
import FormPageFooter from '@/components/form/form-page-footer.vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, useTemplateRef, watch } from 'vue';
import { getProtoDetails, postProto, putProto } from '@/http/proto';
import { Form, InfoBox, Message } from 'bkui-vue';
import { useElementSize } from '@vueuse/core';
import { IProto } from '@/types/proto';
import useElementScroll from '@/hooks/use-element-scroll';
import FormCard from '@/components/form-card.vue';
import MonacoEditor from '@/components/monaco-editor.vue';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { showFirstErrorFormItem } = useElementScroll();
const { height } = useElementSize(document.body);

const formModel = ref<IProto>({
  name: '',
  config: {
    desc: '',
    content: '',
  },
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const formRefBase = useTemplateRef<InstanceType<typeof Form>>('form-ref-base');
const initialSource = ref('');
const pluginConfig = ref('');

const baseRules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'blur' },
  ],
};

const rules = {
  'config.content': [
    {
      validator: () => {
        return !!pluginConfig.value?.length;
      },
      message: t('必填项'),
      trigger: 'blur',
    },
  ],
};

const protoId = computed(() => {
  return route.params.id as string;
});

const isEditMode = computed(() => {
  return !!route.params.id;
});

const editorHeight = computed(() => {
  return height.value - 502;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getProtoDetails({ id } as { id: string });
    const { config, ...rest } = response;

    formModel.value.name = rest.name;
    formModel.value.config.desc = config.desc;
    initialSource.value = config.content;
  }
}, { immediate: true });

const handleCancelClick = () => {
  router.back();
};

const handleSubmit = async () => {
  try {
    // 校验表单
    await Promise.all([
      formRef.value?.validate(),
      formRefBase.value?.validate(),
    ]);

    const data = {
      name: formModel.value.name,
      config: {
        desc: formModel.value.config.desc,
        content: pluginConfig.value,
      },
    };

    InfoBox({
      title: t('确认提交？'),
      confirmText: t('提交'),
      cancelText: t('取消'),
      onConfirm: async () => {
        if (isEditMode.value) {
          await putProto({
            data,
            id: protoId.value,
          });
        } else {
          await postProto({ data });
        }

        Message({
          theme: 'success',
          message: t('提交成功'),
        });

        await router.push({ name: 'proto', replace: true });
      },
    });
  } catch (e) {
    const error = e as Error;
    showFirstErrorFormItem();
    Message({
      theme: 'error',
      message: error.message || t('提交失败'),
    });
  }
};

const handleEditorChange = ({ source }: { source: string }) => {
  try {
    pluginConfig.value = source;
  } catch {
    pluginConfig.value = '';
  }
};

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
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
