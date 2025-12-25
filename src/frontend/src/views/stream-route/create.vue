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
        <BkForm ref="form-ref" :model="formModel" :rules="rules" class="form-element">
          <FormCard>
            <template #title>{{ t('基本信息') }}</template>
            <div>
              <BkFormItem :label="t('名称')" class="form-item" property="name" required>
                <BkInput v-model="formModel.name" clearable />
              </BkFormItem>
              <BkFormItem :label="t('描述')" class="form-item" property="desc">
                <BkInput v-model="routeConfig.desc" clearable />
              </BkFormItem>
            </div>
          </FormCard>
        </BkForm>

        <BkForm :model="routeConfig" class="form-element" style="margin-bottom: 16px;">
          <div class="prev-card-attachment">
            <BkFormItem :label="t('标签')" style="margin-bottom: 0;">
              <FormLabelsNew ref="labels-form-new" :labels="routeConfig.labels" />
            </BkFormItem>

            <BkFormItem :label="t('绑定服务')" class="form-item">
              <SelectService
                v-model="formModel.service_id"
                :check-disabled="formModel.service_id === '__none__'"
                @change="handleServiceChange"
              >
                <BkOption id="__none__" :name="t('不绑定服务')" />
              </SelectService>
            </BkFormItem>

            <BkFormItem :label="t('上游地址')">
              <FormRemoteAddressNew
                ref="remote-addr-form"
                :addrs="routeConfig.remote_addrs"
                :show-add-icon="false"
              >
                <template #tooltips>
                  {{ t('发出请求的客户端地址。') }}
                </template>
              </FormRemoteAddressNew>
            </BkFormItem>

            <BkFormItem :label="t('服务器地址')">
              <FormRemoteAddressNew
                ref="server-addr-form"
                :addrs="routeConfig.server_addrs"
                :show-add-icon="false"
              >
                <template #tooltips>
                  {{ t('接受 Stream Route 连接的 APISIX 服务器的地址。') }}
                </template>
              </FormRemoteAddressNew>
            </BkFormItem>

            <BkFormItem :label="t('服务器端口')" class="form-item w120">
              <BkInput
                v-model="routeConfig.server_port"
                type="number"
                :min="0"
                :precision="0"
                :step="1"
              />
              <div class="form-item-tip">
                {{ t('接受 Stream Route 连接的 APISIX 服务器的端口。') }}
              </div>
            </BkFormItem>
          </div>
        </BkForm>

        <FormCard>
          <template #title>{{ t('上游服务') }}</template>
          <div>
            <BkForm class="form-element">
              <BkFormItem :label="t('选择上游服务')" class="form-item">
                <SelectUpstream
                  v-model="formModel.upstream_id"
                  :check-disabled="!formModel.upstream_id
                    || ['__config__', '__none__'].includes(formModel.upstream_id)"
                  @change="handleUpstreamSelect"
                >
                  <BkOption
                    id="__none__"
                    :disabled="['__none__'].includes(formModel.service_id)"
                    :name="t('不选择（仅在已绑定了服务时可用）')"
                  />
                  <BkOption
                    id="__config__"
                    :name="t('手动填写（会覆盖绑定服务的配置）')"
                  />
                </SelectUpstream>
              </BkFormItem>
            </BkForm>

            <!--  upstream 配置  -->
            <UpstreamForm
              v-if="formModel.upstream_id === '__config__'"
              ref="upstream-form"
              v-model="upstream"
              v-model:flags="flags"
              v-model:ssl_id="ssl_id"
              :desc-and-labels="false"
            />
          </div>
        </FormCard>

        <!--  插件 配置  -->
        <FormCard>
          <template #title>{{ t('插件') }}</template>
          <template #subTitle>{{ t('可直接使用插件组，或逐个添加插件，也可组合使用') }}</template>
          <div>
            <BkForm class="form-element">
              <BkFormItem :label="t('插件')" class="form-item">
                <ButtonIcon
                  icon-color="#3a84ff"
                  style="background: #f0f5ff;border-color: transparent;border-radius: 2px;color:#3a84ff;"
                  @click="isPluginConfigManageSliderVisible = true"
                >
                  {{ t('添加插件') }}
                </ButtonIcon>
              </BkFormItem>
            </BkForm>
            <div style="margin-left: 150px;">
              <ManagePluginConfigNew
                v-model="enabledPluginList"
                v-model:is-show="isPluginConfigManageSliderVisible"
                :plugin-query="{ kind: 'stream' }"
              />
            </div>
          </div>
        </FormCard>
      </div>
    </main>
    <FormPageFooter
      @cancel="handleCancelClick"
      @submit="handleSubmit"
    />
  </div>
</template>

<script lang="ts" setup>
import { cloneDeep, isEmpty, uniq } from 'lodash-es';
import { computed, onMounted, ref, useTemplateRef, watch } from 'vue';
import { Form, InfoBox, Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { IStreamRoute, IStreamRouteConfig } from '@/types/stream-route';
import { IUpstream, IUpstreamConfig } from '@/types/upstream';
import { getUpstreams } from '@/http/upstream';
import { getStreamRoute, postStreamRoute, putStreamRoute } from '@/http/stream-route';
import { useUpstreamForm } from '@/views/upstream/use-upstream-form';
import Ajv from 'ajv';
import addFormats from 'ajv-formats';
import STREAM_ROUTE_JSON from '@/assets/schemas/stream-route.json';
import useConfigFilter from '@/hooks/use-config-filter';
import useElementScroll from '@/hooks/use-element-scroll';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
import useResourcePageDetector from '@/hooks/use-resource-page-detector';
import UpstreamForm, { type IFlags } from '@/components/form/form-upstream.vue';
import ButtonIcon from '@/components/button-icon.vue';
import ManagePluginConfigNew from '@/components/manage-plugin-config-new.vue';
import FormCard from '@/components/form-card.vue';
import FormLabelsNew from '@/components/form/form-labels-new.vue';
import FormRemoteAddressNew from '@/components/form/form-remote-addrs-new.vue';
import FormPageFooter from '@/components/form/form-page-footer.vue';
import SelectUpstream from '@/components/select/select-upstream.vue';
import SelectService from '@/components/select/select-service.vue';

interface ILocalPlugin {
  doc_url?: string
  example?: string
  id?: string
  name?: string
  config?: string
  enabled?: boolean
}

const ajv = new Ajv();
addFormats(ajv);
const schemaValidate = ajv.compile(STREAM_ROUTE_JSON);

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

const formModel = ref<Omit<IStreamRoute, 'config'>>({
  name: '',
  upstream_id: '__config__',
  service_id: '__none__',
});

const routeConfig = ref<Partial<IStreamRouteConfig>>({
  name: '',
  desc: '',
  sni: '',
  server_port: 9090,
  labels: {},
  protocol: {},
  remote_addrs: [],
  server_addrs: [],
});

const ssl_id = ref('');
const plugins = ref<ILocalPlugin>({});
const upstream = ref<IUpstreamConfig>(createDefaultUpstream());
const flags = ref<IFlags>({
  upstreamType: 'nodes',
  tlsType: '__disabled__',
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const labelsFormNewRef = useTemplateRef<InstanceType<typeof FormLabelsNew>>('labels-form-new');
const remoteAddrFormRef =  useTemplateRef<InstanceType<typeof FormRemoteAddressNew>>('remote-addr-form');
const serverAddrFormRef = useTemplateRef<InstanceType<typeof FormRemoteAddressNew>>('server-addr-form');
const upstreamFormRef =  useTemplateRef<InstanceType<typeof UpstreamForm>>('upstream-form');

const rules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'blur' },
  ],
};

const upstreamList = ref<IUpstream[]>([]);

const enabledPluginList = ref<ILocalPlugin[]>([]);

const isPluginConfigManageSliderVisible = ref(false);

const routeDtoId = computed(() => {
  return route.params.id as string;
});

watch(() => route.params.id, async (id: string | null) => {
  if (id) {
    const response = await getStreamRoute({ id } as { id: string });
    const { config, service_id, upstream_id, ...rest } = response;
    const {
      upstream: remoteUpstream,
      plugins: remotePlugins,
      ...restConfig
    } = config;

    await getDependencies();

    routeConfig.value = { ...routeConfig.value, ...restConfig };

    if (routeConfig.value.remote_addr) {
      routeConfig.value.remote_addrs = [routeConfig.value.remote_addr];
      delete routeConfig.value.remote_addr;
    }

    if (routeConfig.value.server_addr) {
      routeConfig.value.server_addrs = [routeConfig.value.server_addr];
      delete routeConfig.value.server_addr;
    }

    if (service_id) {
      formModel.value.service_id = service_id;
      // 路由绑定了服务，上游服务有数据且upstream_id不存在改为手动填写，上游服务无数据且upstream_id不存在改为不选择
      if (!upstream_id) {
        formModel.value.upstream_id = remoteUpstream ? '__config__' :  '__none__';
      }
    }

    if (upstream_id) {
      formModel.value.upstream_id = upstream_id;
    }

    if (remoteUpstream) {
      upstream.value = remoteUpstream;

      if (remoteUpstream.service_name && remoteUpstream.discovery_type) {
        flags.value.upstreamType = 'service_discovery';
        delete upstream.value.nodes;
      }

      if (remoteUpstream.tls?.client_cert && remoteUpstream.tls?.client_key) {
        flags.value.tlsType = '__input__';
      } else if (remoteUpstream.tls?.client_cert_id) {
        flags.value.tlsType = '__select__';
      } else {
        flags.value.tlsType = '__disabled__';
      }

      ssl_id.value = remoteUpstream.tls?.client_cert_id || '';
    }

    if (remotePlugins) {
      enabledPluginList.value = Object.entries(remotePlugins)
        .map(([pluginName, pluginConfig]) => ({
          name: pluginName,
          config: pluginConfig,
        }));
    }

    formModel.value = {
      ...formModel.value,
      ...rest,
    };

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

const handleServerAddress = async () => {
  const services = await serverAddrFormRef.value?.getValue() || [];
  if (!services.length || services.every(sv => isEmpty(sv))) {
    return Promise.resolve();
  }
  return await serverAddrFormRef.value?.validate();
};

const handleRemoteAddress = async () => {
  const address = await remoteAddrFormRef.value?.getValue() || [];
  if (!address.length || address.every(addr => isEmpty(addr))) {
    return Promise.resolve();
  }
  return await remoteAddrFormRef.value?.validate();
};

const handleSubmit = async () => {
  try {
    const extraPlugins = enabledPluginList.value.reduce((result, plugin) => {
      result[plugin.name] = typeof plugin.config === 'string' ? JSON.parse(plugin.config) : plugin.config;
      return result;
    }, {} as ILocalPlugin);

    let config: IStreamRouteConfig = {
      ...cloneDeep(routeConfig.value),
      plugins: {
        ...plugins.value,
        ...extraPlugins,
      },
    };

    const upstreamId = formModel.value.upstream_id;
    config.upstream_id = upstreamId;
    if (['__config__'].includes(upstreamId)) {
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

          if (upstreamCopy?.checks?.active) {
            let { req_headers } = upstreamCopy.checks.active;
            req_headers = uniq(req_headers?.filter(value => !!value));

            if (!req_headers?.length) {
              delete upstreamCopy.checks.active.req_headers;
            }
          }
        }
      }

      // 写入 ssl_id
      if (['__select__'].includes(flags.value.tlsType)) {
        upstreamCopy.tls = {
          client_cert_id: ssl_id.value,
        };
      }

      config = { ...config, upstream: upstreamCopy };
      delete config.upstream_id;
    }

    if ((['__none__'].includes(upstreamId)) || !config.upstream_id) {
      delete config.upstream_id;
    }

    // 校验表单
    await Promise.all([
      formRef.value?.validate(),
      handleLabels(),
      handleServerAddress(),
      handleRemoteAddress(),
    ]);

    if (config.upstream) {
      await upstreamFormRef.value?.validate();
    }

    config.labels = await labelsFormNewRef.value.getValue() || {};

    config.server_addrs = await serverAddrFormRef.value?.getValue() || routeConfig.value.server_addrs || [];
    config.server_addrs = uniq(config.server_addrs.filter(value => !!value) || []);
    if (!config.server_addrs.length) {
      delete config.server_addrs;
    }

    config.remote_addrs = await remoteAddrFormRef.value?.getValue() || routeConfig.value.remote_addrs || [];
    config.remote_addrs = uniq(config.remote_addrs?.filter(value => !!value) || []);
    if (!config.remote_addrs.length) {
      delete config.remote_addrs;
    }

    // 过滤值为空或默认值的字段
    config = filterEmpty(config);
    config = filterAdvanced(config, 'stream_route');
    if (config.upstream) {
      config.upstream = filterEmpty(config.upstream);
      config.upstream = filterAdvanced(config.upstream, 'upstream');
    }

    if (config?.remote_addrs?.length) {
      config.remote_addr = config.remote_addrs.join();
      delete config.remote_addrs;
    }

    if (config?.server_addrs?.length) {
      config.server_addr = config.server_addrs.join();
      delete config.server_addrs;
    }

    // 校验 schema
    if (schemaValidate(config)) {
      const data: Partial<IStreamRoute> = {
        config,
        name: formModel.value.name,
      };

      if (!['__none__'].includes(formModel.value.service_id)) {
        data.service_id = formModel.value.service_id;
        if (formModel.value.service_id) {
          data.config.service_id = formModel.value.service_id;
        }
      }

      // 既没选择“手动填写” upstream，也没选择“不选择”时才传入 upstream_id
      if (!['__none__', '__config__'].includes(formModel.value.upstream_id)) {
        data.upstream_id = formModel.value.upstream_id;
      }

      InfoBox({
        title: t('确认提交？'),
        confirmText: t('提交'),
        cancelText: t('取消'),
        onConfirm: async () => {
          if (isEditMode.value) {
            await putStreamRoute({
              data,
              id: routeDtoId.value,
            });
          } else {
            await postStreamRoute({ data });
          }

          Message({
            theme: 'success',
            message: t('提交成功'),
          });

          await router.push({ name: 'stream-route', replace: true });
        },
      });
    } else {
      showSchemaErrorMessages(schemaValidate.errors);
    }
  } catch (error: Error) {
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

const getUpstreamList = async () => {
  const response = await getUpstreams({ query: { offset: 0, limit: 100 } });
  upstreamList.value = response.results || [];
};

// 绑定服务联动选择上游服务
const handleServiceChange = () => {
  // 选择了不绑定服务，则上游不允许为“不选择”, 选择了绑定的服务，则上游自动改为不选择
  const isServiceNone = formModel.value.service_id === '__none__';
  if (formModel.value.service_id) {
    formModel.value.upstream_id = isServiceNone ? '__config__' : '__none__';
  }
};

const handleUpstreamSelect = () => {
  if (!formModel.value.upstream_id || ['__config__', '__none__'].includes(formModel.value.upstream_id)) {
    upstream.value = createDefaultUpstream();
    return;
  }

  const curUpstream = upstreamList.value.find(upstream => upstream.id === formModel.value.upstream_id) ?? {};
  if (curUpstream) {
    upstream.value = cloneDeep(curUpstream.config);
  }
};

const getDependencies = async () => {
  await getUpstreamList();
};

onMounted(() => {
  if (isEditMode.value) {
    return;
  }
  getDependencies();
});

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;

  .form-item-tip {
    font-size: 12px;
    color: #979ba5;
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
