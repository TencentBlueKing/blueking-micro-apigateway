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
          <bk-form
            ref="form-ref" :model="formModel" :rules="rules" class="form-element"
          >
            <!-- name -->
            <bk-form-item :label="t('名称')" class="form-item" property="name" required>
              <bk-input v-model="formModel.name" clearable />
            </bk-form-item>

            <!-- desc -->
            <bk-form-item :label="t('描述')" class="form-item" property="desc">
              <bk-input v-model="pluginConfig.desc" clearable />
            </bk-form-item>

            <!-- labels -->
            <bk-form-item :label="t('标签')" style="margin-bottom: 0;">
              <form-labels-new ref="labels-form-new" :labels="pluginConfig.labels" />
            </bk-form-item>
          </bk-form>
        </div>
      </form-card>

      <!--  插件 配置  -->
      <form-card>
        <template #title>{{ t('配置插件') }}</template>
        <div>
          <bk-form ref="plugin-form-ref" :model="pluginFormModel" :rules="pluginFormRules" class="form-element">
            <bk-form-item :label="t('插件')" class="form-item" property="plugins" required>
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
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, onBeforeMount, ref, useTemplateRef, watch } from 'vue';
import { getConsumerGroup, postConsumerGroup, putConsumerGroup } from '@/http/consumer-group';
import { Form, InfoBox, Message } from 'bkui-vue';
import { IConsumerGroup, IConsumerGroupConfig } from '@/types/consumer-group';
import { getResourceSchema } from '@/http/schema';
import Ajv from 'ajv';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
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

const formModel = ref<Omit<IConsumerGroup, 'config'>>({
  name: '',
});

const pluginFormModel = ref<{ plugins: ILocalPlugin[] }>({
  plugins: [],
});

const pluginConfig = ref<Omit<IConsumerGroupConfig, 'plugins'>>({
  desc: '',
  labels: {},
  // plugins: {},
});

const enabledPluginList = ref<ILocalPlugin[]>([]);
const schema = ref<Record<string, any>>({});
const isPluginConfigManageSliderVisible = ref(false);

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const pluginFormRef = useTemplateRef<InstanceType<typeof Form>>('plugin-form-ref');
const labelsFormNewRef = useTemplateRef('labels-form-new');

const rules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'blur' },
  ],
};

const pluginFormRules = {
  plugins: [
    { validator: (plugins: ILocalPlugin[]) => !!plugins.length, message: t('必须配置插件'), required: true },
  ],
};

const consumerGroupId = computed(() => {
  return route.params.id as string;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getConsumerGroup({ id } as { id: string });
    const { config, ...rest } = response;
    formModel.value.name = rest.name;
    const { plugins, ...restConfig } = config;
    pluginConfig.value = restConfig;
    enabledPluginList.value = Object.entries(plugins)
      .map(([pluginName, pluginConfig]) => ({
        name: pluginName,
        config: typeof pluginConfig === 'string' ? pluginConfig : JSON.stringify(pluginConfig),
      }));

    if (isCloneMode.value) {
      formModel.value.name += '_clone';
    }
  }
}, { immediate: true });

watch(enabledPluginList, () => {
  pluginFormModel.value.plugins = enabledPluginList.value;
}, { deep: true });

const handleCancelClick = () => {
  router.back();
};

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
      formRef.value!.validate(),
      pluginFormRef.value!.validate(),
      handleLabels(),
    ]);

    const plugins = enabledPluginList.value.reduce((result, plugin) => {
      result[plugin.name] = typeof plugin.config === 'string' ? JSON.parse(plugin.config) : plugin.config;
      return result;
    }, {} as Record<string, any>);

    let config: Record<any, any> = {
      ...pluginConfig.value,
      plugins,
    };

    config.labels = await labelsFormNewRef.value.getValue() || {};

    // 过滤值为空或默认值的字段
    if (!isEditMode.value) {
      config = filterEmpty(config, ['plugins']);
    }

    const schemaValidate = ajv.compile(schema.value);

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
            await putConsumerGroup({
              data,
              id: consumerGroupId.value,
            });
          } else {
            await postConsumerGroup({ data });
          }

          Message({
            theme: 'success',
            message: t('提交成功'),
          });

          await router.push({ name: 'consumer-group', replace: true });
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

onBeforeMount(async () => {
  schema.value = await getResourceSchema({ type: 'consumer_group' });
});

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
}

</style>
