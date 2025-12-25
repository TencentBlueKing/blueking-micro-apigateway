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
          <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
            <!-- username -->
            <bk-form-item :label="t('名称')" class="form-item" property="username" required>
              <bk-input v-model="formModel.username" clearable />
            </bk-form-item>

            <!-- desc -->
            <bk-form-item :label="t('描述')" class="form-item" property="desc">
              <bk-input v-model="consumer.desc" clearable />
            </bk-form-item>

            <!-- labels -->
            <bk-form-item :label="t('标签')" style="margin-bottom: 0;">
              <form-labels-new ref="labels-form-new" :labels="consumer.labels" />
            </bk-form-item>

            <bk-form-item :label="t('消费者组')" class="form-item">
              <select-consumer-group v-model="formModel.group_id" />
            </bk-form-item>
          </bk-form>
        </div>
      </form-card>

      <!--  插件 配置  -->
      <form-card>
        <template #title>{{ t('配置插件') }}</template>
        <div>
          <bk-form class="form-element">
            <bk-form-item :label="t('插件')" class="form-item">
              <button-icon
                icon-color="#3a84ff"
                style="background: #f0f5ff;border-color: transparent;border-radius: 2px;color:#3a84ff;"
                @click="isPluginConfigManageSliderVisible = true"
              >{{
                t('添加插件')
              }}
              </button-icon>
            </bk-form-item>
          </bk-form>
          <div style="margin-left: 150px;">
            <manage-plugin-config-new
              v-model="enabledPluginList" v-model:is-show="isPluginConfigManageSliderVisible" type="consumer"
            />
          </div>
        </div>
      </form-card>
    </main>
    <form-page-footer @cancel="handleCancelClick" @submit="handleSubmit" />
  </div>
</template>

<script lang="ts" setup>
import FormPageFooter from '@/components/form/form-page-footer.vue';
import { Form, InfoBox, Message } from 'bkui-vue';
import { IConsumer, IConsumerConfig } from '@/types/consumer';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, onBeforeMount, ref, useTemplateRef, watch } from 'vue';
import { getConsumer, postConsumer, putConsumer } from '@/http/consumer';
import SelectConsumerGroup from '@/components/select/select-consumer-group.vue';
import Ajv from 'ajv';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
import { getResourceSchema } from '@/http/schema';
import useConfigFilter from '@/hooks/use-config-filter';
import useResourcePageDetector from '@/hooks/use-resource-page-detector';
import useElementScroll from '@/hooks/use-element-scroll';
import FormCard from '@/components/form-card.vue';
import ButtonIcon from '@/components/button-icon.vue';
import ManagePluginConfigNew from '@/components/manage-plugin-config-new.vue';
import FormLabelsNew from '@/components/form/form-labels-new.vue';
import { isEmpty } from 'lodash-es';

interface ILocalPlugin {
  doc_url?: string
  example?: string
  id?: string
  name: string
  config: string
  enabled?: boolean
}

const ajv = new Ajv();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { showSchemaErrorMessages } = useSchemaErrorMessage();
const { showFirstErrorFormItem } = useElementScroll();
const { filterEmpty } = useConfigFilter();
const { isEditMode, isCloneMode } = useResourcePageDetector();

const formModel = ref<Omit<IConsumer, 'config'>>({
  username: '',
  group_id: '',
});

const consumer = ref<IConsumerConfig>({
  desc: '',
  labels: {},
});

const enabledPluginList = ref<ILocalPlugin[]>([]);
const schema = ref<Record<string, any>>({});
const isPluginConfigManageSliderVisible = ref(false);

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const labelsFormNewRef = useTemplateRef('labels-form-new');

const rules = {
  username: [
    { required: true, message: t('必填项'), trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_]+$/, message: t('只能包含大小写字母、数字和下划线'), trigger: 'change' },
  ],
};

const consumerId = computed(() => {
  return route.params.id as string;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getConsumer({ id } as { id: string });
    const { config, ...rest } = response;
    // await getConsumerGroupOptions();
    formModel.value = rest;
    formModel.value.username = rest.username || rest.name;
    formModel.value.group_id = rest.group_id || config.group_id;
    const { plugins, ...restConfig } = config;
    consumer.value = restConfig;

    if (plugins) {
      enabledPluginList.value = Object.entries(plugins)
        .map(([pluginName, pluginConfig]) => ({
          name: pluginName,
          config: typeof pluginConfig === 'string' ? pluginConfig : JSON.stringify(pluginConfig),
        }));
    } else {
      enabledPluginList.value = [];
    }

    if (isCloneMode.value) {
      formModel.value.username += '_clone';
    }
  }
}, { immediate: true });

const handleLabels = async () => {
  const labels = await labelsFormNewRef.value.getValue();

  if (isEmpty(labels)) {
    return Promise.resolve();
  }
  return await labelsFormNewRef.value.validate();
};

const handleSubmit = async () => {
  try {
    // 校验表单
    await Promise.all([
      formRef.value?.validate(),
      handleLabels(),
    ]);

    const plugins = enabledPluginList.value.reduce((result, plugin) => {
      result[plugin.name] = typeof plugin.config === 'string' ? JSON.parse(plugin.config) : plugin.config;
      return result;
    }, {} as Record<string, any>);

    let config: Record<any, any> = {
      ...consumer.value,
      plugins,
      username: formModel.value.username,
    };

    if (formModel.value.group_id) {
      Object.assign(config, { group_id: formModel.value.group_id });
    }

    config.labels = await labelsFormNewRef.value.getValue() || {};

    // 过滤值为空或默认值的字段
    if (!isEditMode.value) {
      config = filterEmpty(config);
    }

    const schemaValidate = ajv.compile(schema.value);

    if (schemaValidate(config)) {
      const data = {
        config,
        username: formModel.value.username,
        name: formModel.value.username,
        group_id: formModel.value.group_id,
      };

      InfoBox({
        title: t('确认提交？'),
        confirmText: t('提交'),
        cancelText: t('取消'),
        onConfirm: async () => {
          if (isEditMode.value) {
            await putConsumer({
              data,
              id: consumerId.value,
            });
          } else {
            await postConsumer({ data });
          }

          Message({
            theme: 'success',
            message: t('提交成功'),
          });

          await router.push({ name: 'consumer', replace: true });
        },
      });
    } else {
      showSchemaErrorMessages(schemaValidate.errors);
    }
  } catch (e) {
    const error = e as Error;
    showFirstErrorFormItem();
    Message({
      theme: 'error',
      message: error.message || t('提交失败'),
    });
  }
};

const handleCancelClick = () => {
  router.back();
};

onBeforeMount(async () => {
  schema.value = await getResourceSchema({ type: 'consumer' });
});

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
}

</style>
