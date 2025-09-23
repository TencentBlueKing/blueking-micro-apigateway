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
        <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
          <form-card>
            <template #title>{{ t('基本信息') }}</template>
            <div>
              <!-- name -->
              <bk-form-item :label="t('名称')" class="form-item" property="name" required>
                <bk-input v-model="formModel.name" clearable />
              </bk-form-item>

              <!-- desc -->
              <bk-form-item :label="t('描述')" class="form-item" property="desc">
                <bk-input v-model="formModel.desc" clearable />
              </bk-form-item>

              <!-- labels -->
              <bk-form-item :label="t('标签')" style="margin-bottom: 0;">
                <form-labels-new ref="labels-form-new" :labels="formModel.labels" />
              </bk-form-item>

              <form-advanced-switch v-model="uiConfig.showAdvanced" />

              <!-- 高级配置 -->
              <template v-if="uiConfig.showAdvanced">
                <!-- hosts -->
                <bk-form-item :label="t('匹配域名')">
                  <form-hosts-new ref="hosts-form" :hosts="formModel.hosts">
                    <template #tooltips>
                      {{
                        t('非必填，如果填写，则代表附加了匹配规则，只有命中域名的请求' +
                          '才会继续匹配 service 下的route; 例如 配置 *.bar.com，' +
                          '此时`a.bar.com/get`可以命中，使用`foor.com/get`则会 404')
                      }}
                    </template>
                  </form-hosts-new>
                </bk-form-item>

                <!-- enable_websocket -->
                <bk-form-item :label="t('启用 WebSocket')" class="form-item" property="enable_websocket">
                  <bk-switcher v-model="formModel.enable_websocket" theme="primary" />
                </bk-form-item>
              </template>
            </div>
          </form-card>

          <form-card>
            <template #title>{{ t('上游服务') }}</template>

            <bk-form-item :label="t('选择上游服务')" class="form-item">
              <select-upstream
                v-model="upstreamId"
                :check-disabled="!upstreamId || upstreamId === '__config__' || upstreamId === '__bind-id__'"
                :static-options="upstreamOptions"
                @change="handleUpstreamSelect"
                @check="handleCheckUpstreamConfigClick"
              />
            </bk-form-item>
          </form-card>
        </bk-form>

        <!-- 上游服务配置表单 -->
        <div class="prev-card-attachment">
          <upstream-form
            v-show="upstreamId === '__config__'"
            ref="upstream-form"
            v-model="upstream"
            v-model:flags="flags"
            v-model:ssl_id="ssl_id"
            :desc-and-labels="false"
          />
        </div>
      </div>

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
          <!-- 插件配置 -->
          <div style="margin-left: 150px;">
            <manage-plugin-config-new
              v-model="enabledPluginList"
              v-model:is-show="isPluginConfigManageSliderVisible"
            />
          </div>
        </div>
      </form-card>
    </main>
    <slider-resource-viewer
      v-model="isResourceViewerShow"
      :resource="selectedUpstream"
      :source="selectedUpstreamSource"
      resource-type="upstream"
    />
    <form-page-footer @cancel="handleCancelClick" @submit="handleSubmit" />
  </div>
</template>

<script lang="ts" setup>
import UpstreamForm, { type IFlags } from '@/components/form/form-upstream.vue';
import { IUpstream, IUpstreamConfig } from '@/types/upstream';
import { useUpstreamForm } from '@/views/upstream/use-upstream-form';
import FormPageFooter from '@/components/form/form-page-footer.vue';
import { IService, IServiceConfig } from '@/types/service';
import Ajv from 'ajv';
import SERVICE_JSON from '@/assets/schemas/service.json';
import { Form, InfoBox, Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, onMounted, ref, useTemplateRef, watch } from 'vue';
import { getService, postService, putService } from '@/http/service';
import { cloneDeep, isEmpty, uniq } from 'lodash-es';
import SelectUpstream from '@/components/select/select-upstream.vue';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
import SliderResourceViewer from '@/components/slider-resource-viewer.vue';
import { getUpstream } from '@/http/upstream';
import useConfigFilter from '@/hooks/use-config-filter';
import useResourcePageDetector from '@/hooks/use-resource-page-detector';
import useElementScroll from '@/hooks/use-element-scroll';
import FormCard from '@/components/form-card.vue';
import ButtonIcon from '@/components/button-icon.vue';
import ManagePluginConfigNew from '@/components/manage-plugin-config-new.vue';
import FormAdvancedSwitch from '@/components/form/form-advanced-switch.vue';
import FormLabelsNew from '@/components/form/form-labels-new.vue';
import FormHostsNew from '@/components/form/form-hosts-new.vue';

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
const { showSchemaErrorMessages } = useSchemaErrorMessage();
const { showFirstErrorFormItem } = useElementScroll();
const { createDefaultUpstream } = useUpstreamForm();
const {
  filterEmpty,
  filterAdvanced,
} = useConfigFilter();
const { isEditMode, isCloneMode } = useResourcePageDetector();

const ajv = new Ajv();
const schemaValidate = ajv.compile(SERVICE_JSON);

const formModel = ref<IServiceConfig>({
  name: '',
  desc: '',
  enable_websocket: false,
  hosts: [],
  labels: {},
});
const ssl_id = ref('');
const upstream = ref<Partial<IUpstreamConfig>>(createDefaultUpstream());

// 配置上游服务的方式 __config__: 手动填写 | __bind-id__：选择已绑定的上游
const upstreamId = ref('__config__');
// const upstreamList = ref<IUpstream[]>([]);
const selectedUpstream = ref<IUpstream>();
const enabledPluginList = ref<ILocalPlugin[]>([]);
const selectedUpstreamSource = ref('{}');
const isResourceViewerShow = ref(false);
const isPluginConfigManageSliderVisible = ref(false);
const flags = ref<IFlags>({
  upstreamType: 'nodes',
  tlsType: '__disabled__',
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const labelsFormNewRef = useTemplateRef('labels-form-new');
const hostsFormRef = useTemplateRef('hosts-form');
const upstreamFormRef = useTemplateRef('upstream-form');

const rules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'change' },
  ],
};

const upstreamOptions = ref([
  {
    id: '__bind-id__',
    name: '不选择（仅在绑定服务时可用）',
  },
  {
    id: '__config__',
    name: '手动填写',
  },
]);

const uiConfig = ref({
  showAdvanced: false,
});

const serviceDtoId = computed(() => {
  return route.params.id as string;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getService({ id } as { id: string });
    const { config, upstream_id } = response;
    formModel.value = { ...formModel.value, ...config };

    if (upstream_id) {
      upstreamId.value = upstream_id;
    }

    if (config.upstream) {
      upstream.value = config.upstream;

      if (config.upstream.service_name && config.upstream.discovery_type) {
        flags.value.upstreamType = 'service_discovery';
        delete upstream.value.nodes;
      }

      if (config.upstream.tls?.client_cert && config.upstream.tls?.client_key) {
        flags.value.tlsType = '__input__';
      } else if (config.upstream.tls?.client_cert_id) {
        flags.value.tlsType = '__select__';
      } else {
        flags.value.tlsType = '__disabled__';
      }

      ssl_id.value = config.upstream.tls?.client_cert_id || '';
    }

    if (config.plugins) {
      enabledPluginList.value = Object.entries(config.plugins)
        .map(([pluginName, pluginConfig]) => ({
          name: pluginName,
          config: pluginConfig,
        }));
    }

    if (isCloneMode.value) {
      formModel.value.name += '_clone';
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

const handleHosts = async () => {
  const hosts = await hostsFormRef.value?.getValue() || [];

  if (!hosts.length || hosts.every(host => isEmpty(host))) {
    return Promise.resolve();
  }
  return await hostsFormRef.value?.validate();
};

const handleSubmit = async () => {
  try {
    let config: IServiceConfig | null;
    const data: Partial<IService> = {
      name: formModel.value.name,
    };

    // 过滤属性
    if (upstreamId.value === '__config__') {
      const { upstream_id, ...rest } = formModel.value;

      const upstreamCopy = cloneDeep(upstream.value);

      if (upstreamCopy.checks) {
        if (!Object.keys(upstreamCopy.checks).length) {
          delete upstreamCopy.checks;
        } else {
          if (!upstreamCopy.checks.active && upstreamCopy.checks.passive) {
            Message({
              theme: 'error',
              message: t('打开被动检查时也必须打开主动检查'),
            });
            return false;
          }

          if (upstreamCopy.checks.active) {
            let { req_headers } = upstreamCopy.checks.active;
            req_headers = uniq(req_headers?.filter(value => !!value));

            if (!req_headers?.length) {
              delete upstreamCopy.checks.active.req_headers;
            }
          }
        }
      }

      // 写入 ssl_id
      if (flags.value.tlsType === '__select__') {
        upstreamCopy.tls = {
          client_cert_id: ssl_id.value,
        };
      }

      config = { ...rest, upstream: upstreamCopy };
    } else if (upstreamId.value === '__bind-id__') {
      const { upstream, ...rest } = formModel.value;
      config = cloneDeep(rest);
    } else {
      const { upstream, ...rest } = formModel.value;
      config = cloneDeep(rest);
      data.upstream_id = upstreamId.value;
    }

    // 写入插件 plugins
    config.plugins = enabledPluginList.value.reduce((result, plugin) => {
      result[plugin.name] = typeof plugin.config === 'string' ? JSON.parse(plugin.config) : plugin.config;
      return result;
    }, {} as Record<string, any>);

    // 校验表单
    if (upstreamId.value === '__config__') {
      await Promise.all([
        formRef.value?.validate(),
        handleLabels(),
        handleHosts(),
        upstreamFormRef.value?.validate(),
      ]);
    } else {
      await Promise.all([
        formRef.value?.validate(),
        handleLabels(),
        handleHosts(),
      ]);
    }

    config.hosts = await hostsFormRef.value?.getValue() || formModel.value.hosts || [];
    config.hosts = uniq(config.hosts.filter(value => !!value));
    if (!config.hosts.length) {
      delete config.hosts;
    }

    config.labels = await labelsFormNewRef.value.getValue() || {};

    // 过滤值为空或默认值的字段
    // if (!isEditMode.value) {
    config = filterEmpty(config);
    config = filterAdvanced(config, 'service');

    if (config.upstream) {
      config.upstream = filterEmpty(config.upstream);
      config.upstream = filterAdvanced(config.upstream, 'upstream');
    }
    // }

    // 校验 schema
    if (schemaValidate(config)) {
      InfoBox({
        title: t('确认提交？'),
        confirmText: t('提交'),
        cancelText: t('取消'),
        onConfirm: async () => {
          if (isEditMode.value) {
            await putService({
              id: serviceDtoId.value,
              data: {
                config,
                ...data,
              },
            });
          } else {
            await postService({
              data: {
                config,
                ...data,
              },
            });
          }

          Message({
            theme: 'success',
            message: t('提交成功'),
          });

          await router.push({ name: 'service', replace: true });
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

const handleUpstreamSelect = ({ upstream }: { upstream: IUpstream }) => {
  if (!upstreamId.value || upstreamId.value === '__config__' || upstreamId.value === '__bind-id__') {
    // upstream.value = createDefaultUpstream();
    return;
  }

  selectedUpstream.value = upstream;
  const { config } = upstream;
  selectedUpstreamSource.value = typeof config !== 'string' ? JSON.stringify(config) : config;
};

const handleCheckUpstreamConfigClick = async () => {
  if (!upstreamId.value || upstreamId.value === '__config__' || upstreamId.value === '__bind-id__') {
    return;
  }

  if (!selectedUpstream.value) {
    selectedUpstream.value = await getUpstream({ id: upstreamId.value });
    const { config } = selectedUpstream.value;
    selectedUpstreamSource.value = typeof config !== 'string' ? JSON.stringify(config) : config;
  }

  isResourceViewerShow.value = true;
};

const handleCancelClick = () => {
  router.back();
};

onMounted(async () => {
  if (isEditMode.value) {
    return;
  }
  // await getUpstreamOptions();
});

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
}

.advanced-setting-switch {
  font-size: 14px;
  line-height: 22px;
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #3a84ff;

  .switch-icon.is-on {
    transform: rotate(180deg);
  }
}

.prev-card-attachment {
  margin-top: -15px;
  margin-bottom: 16px;
  padding: 0 24px 16px;
  border-radius: 2px;
  background: #ffffff;
  box-shadow: 0 2px 4px 0 #1919290d;
}

</style>
