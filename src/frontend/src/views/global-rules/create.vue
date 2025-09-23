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
            <!-- name -->
            <bk-form-item :label="t('名称')" class="form-item" property="name" required>
              <bk-input v-model="formModel.name" clearable />
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
              >
                {{ t('添加插件') }}
              </button-icon>
            </bk-form-item>
          </bk-form>
          <div style="margin-left: 150px;">
            <manage-plugin-config-new
              v-model="enabledPluginList"
              v-model:is-show="isPluginConfigManageSliderVisible"
              :disable-list="existedPluginNameList"
              @removed="handlePluginRemoved"
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
import { getGlobalRule, getGlobalRulePlugins, postGlobalRules, putGlobalRules } from '@/http/global-rules';
import { Form, InfoBox, Message } from 'bkui-vue';
import { IGlobalRules } from '@/types/global-rules';
import useElementScroll from '@/hooks/use-element-scroll';
import FormCard from '@/components/form-card.vue';
import ButtonIcon from '@/components/button-icon.vue';
import ManagePluginConfigNew from '@/components/manage-plugin-config-new.vue';

interface ILocalPlugin {
  doc_url?: string
  example?: string
  id?: string
  name: string
  config: string
  enabled?: boolean
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { showFirstErrorFormItem } = useElementScroll();

const formModel = ref<Omit<IGlobalRules, 'config'>>({
  name: '',
});

const pluginFormModel = ref<{ plugins: ILocalPlugin[] }>({
  plugins: [],
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const pluginFormRef = useTemplateRef<InstanceType<typeof Form>>('plugin-form-ref');

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

const enabledPluginList = ref<ILocalPlugin[]>([]);
const existedPluginNameList = ref<string[]>([]);
const isPluginConfigManageSliderVisible = ref(false);

const globalRuleId = computed(() => {
  return route.params.id as string;
});

const isEditMode = computed(() => {
  return !!route.params.id;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getGlobalRule({ id } as { id: string });
    const { config, ...rest } = response;
    formModel.value.name = rest.name;
    enabledPluginList.value = Object.entries(config.plugins)
      .map(([pluginName, pluginConfig]) => ({
        name: pluginName,
        config: typeof pluginConfig === 'string' ? pluginConfig : JSON.stringify(pluginConfig),
      }));
  }
}, { immediate: true });

watch(enabledPluginList, () => {
  pluginFormModel.value.plugins = enabledPluginList.value;
  if (pluginFormModel.value.plugins?.length) {
    pluginFormRef.value.clearValidate();
  }
}, { deep: true });

const handleCancelClick = () => {
  router.back();
};

const handleSubmit = async () => {
  try {
    // 校验表单
    await Promise.all([
      formRef.value!.validate(),
      pluginFormRef.value!.validate(),
    ]);

    const plugins = enabledPluginList.value.reduce((result, plugin) => {
      result[plugin.name] = typeof plugin.config === 'string' ? JSON.parse(plugin.config) : plugin.config;
      return result;
    }, {} as Record<string, any>);

    const data = {
      name: formModel.value.name,
      config: {
        plugins,
      },
    };

    InfoBox({
      title: t('确认提交？'),
      confirmText: t('提交'),
      cancelText: t('取消'),
      onConfirm: async () => {
        if (isEditMode.value) {
          await putGlobalRules({
            data,
            id: globalRuleId.value,
          });
        } else {
          await postGlobalRules({ data });
        }

        Message({
          theme: 'success',
          message: t('提交成功'),
        });

        await router.push({ name: 'global-rules', replace: true });
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

const handlePluginRemoved = (pluginName: string) => {
  const index = existedPluginNameList.value.findIndex(name => name === pluginName);
  existedPluginNameList.value.splice(index, 1);
};

onBeforeMount(async () => {
  const response = await getGlobalRulePlugins();
  existedPluginNameList.value = Object.keys(response.plugins?.[0] || {});
});

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
}

</style>
