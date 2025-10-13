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
                <bk-input v-model="routeConfig.desc" clearable />
              </bk-form-item>
            </div>
          </form-card>
        </bk-form>

        <bk-form :model="routeConfig" class="form-element" style="margin-bottom: 16px;">
          <div class="prev-card-attachment">
            <!-- labels -->
            <bk-form-item :label="t('标签')" style="margin-bottom: 0;">
              <form-labels-new ref="labels-form-new" :labels="routeConfig.labels" />
            </bk-form-item>

            <bk-form-item :label="t('绑定服务')" class="form-item">
              <select-service
                v-model="formModel.service_id"
                :check-disabled="formModel.service_id === '__none__'"
                @change="handleServiceIdChanged"
              >
                <bk-option id="__none__" :name="t('不绑定服务')" />
              </select-service>
            </bk-form-item>

            <!-- enable_websocket -->
            <bk-form-item :label="t('启用 WebSocket')" class="form-item" property="enable_websocket">
              <bk-switcher v-model="routeConfig.enable_websocket" theme="primary" />
            </bk-form-item>
          </div>

          <form-card>
            <template #title>{{ t('匹配条件') }}</template>
            <template #subTitle>{{
              t('只有带红色星号的必填，其他字段非必填，如果填写会增加路由的匹配规则，当前仅当你知道配置项的作用，才去配置！')
            }}
            </template>
            <div>
              <!-- methods -->
              <bk-form-item :label="t('HTTP 方法')" class="form-item" property="methods">
                <bk-select
                  v-model="routeConfig.methods"
                  multiple
                  multiple-mode="tag"
                >
                  <bk-option
                    v-for="method in HTTP_METHODS_MAP"
                    :id="method"
                    :key="method"
                    :name="method"
                  />
                </bk-select>
                <div class="common-form-tips">
                  {{ t('非必填，为空则代表不做任何限制；如果填写，则代表只有对应的 HTTP 方法能命中路由') }}
                </div>
              </bk-form-item>

              <!-- uris -->
              <bk-form-item :label="t('路径')" required>
                <form-uris ref="uris-form" v-model="routeConfig.uris">
                  <template #tooltips>
                    {{
                      t('HTTP 请求路径，如 /foo/index.html，支持请求路径前缀 /foo/*。/* 代表所有路径，' +
                        '部署 APISIX 时使用不同的 Router 有不同的匹配方式，' +
                        '具体参考 Router https://apisix.apache.org/zh/docs/apisix/terminology/router/')
                    }}
                  </template>
                </form-uris>
              </bk-form-item>

              <form-advanced-switch v-model="uiConfig.showAdvanced" />

              <!-- 高级配置 -->
              <template v-if="uiConfig.showAdvanced">
                <!-- priority -->
                <bk-form-item :label="t('优先级')" class="form-item w120" property="priority">
                  <bk-input v-model="routeConfig.priority" :min="0" :precision="0" :step="1" type="number" />
                  <div class="common-form-tips">
                    {{
                      t('非必填，默认不需要设置；只有不同路由中包含了相同的uri， 才需要配置，会根据优先级匹配；值越大优先级越高。')
                    }}
                  </div>
                </bk-form-item>

                <!-- hosts -->
                <bk-form-item :label="t('域名')">
                  <form-hosts-new ref="hosts-form" :hosts="routeConfig.hosts">
                    <template #tooltips>{{
                      t('非必填，路由匹配的域名列表。支持泛域名，如：*.test.com；' +
                        '如果填写，则代表附加了匹配规则，只有命中域名的请求才会匹配，例如路径是`/get`，域名配置 `foo.com`，' +
                        '那么只有`foo.com/get`才能命中当前路由，`bar.com/get`不会命中当前路由。')
                    }}
                    </template>
                  </form-hosts-new>
                </bk-form-item>

                <!-- remote_addr -->
                <bk-form-item :label="t('客户端地址')">
                  <form-remote-addrs-new ref="remote-addrs-form" :addrs="routeConfig.remote_addrs">
                    <template #tooltips>{{
                      t('非必填，匹配规则，客户端与服务器握手时 IP，即客户端 IP，' +
                        '例如：192.168.1.101，192.168.1.0/24，::1，fe80::1，fe80::1/64；' +
                        '如果填写，则代表附加了匹配规则，只有配置的客户端 IP 才会命中路由。')
                    }}
                    </template>
                  </form-remote-addrs-new>
                </bk-form-item>
              </template>

              <!-- vars -->
              <bk-form-item v-show="uiConfig.showAdvanced" label="Vars">
                <vars-management ref="vars-management" :vars="routeConfig.vars" />
              </bk-form-item>
            </div>
          </form-card>
        </bk-form>

        <form-card>
          <template #title>{{ t('上游服务') }}</template>
          <div>
            <bk-form class="form-element">
              <bk-form-item :label="t('选择上游服务')" class="form-item">
                <select-upstream
                  v-model="formModel.upstream_id"
                  :check-disabled="!formModel.upstream_id
                    || formModel.upstream_id === '__config__'
                    || formModel.upstream_id === '__none__'"
                  @change="handleUpstreamSelect"
                >
                  <bk-option
                    id="__none__" :disabled="formModel.service_id === '__none__'"
                    :name="t('不选择（仅在已绑定了服务时可用）')"
                  />
                  <bk-option id="__config__" :name="t('手动填写（会覆盖绑定服务的配置）')" />
                </select-upstream>
              </bk-form-item>
            </bk-form>

            <!--  upstream 配置  -->
            <upstream-form
              v-if="formModel.upstream_id === '__config__'"
              ref="upstream-form"
              v-model="upstream"
              v-model:flags="flags"
              v-model:ssl_id="ssl_id"
              :desc-and-labels="false"
            />
          </div>
        </form-card>

        <!--  插件 配置  -->
        <form-card>
          <template #title>{{ t('插件') }}</template>
          <template #subTitle>{{ t('可直接使用插件组，或逐个添加插件，也可组合使用') }}</template>
          <div>
            <bk-form class="form-element">
              <bk-form-item :label="t('插件组')" class="form-item">
                <select-plugin-config v-model="routeConfig.plugin_config_id" />
                <div class="common-form-tips" style="line-height: 1.2;margin-top: 4px;">{{
                  t('可以将一组通用的插件配置提取成插件组，然后在路由中引用；' +
                    '对于同一个插件的配置，只能有一个是有效的，优先级为 Consumer > Route > Plugin Config > Service。')
                }}
                </div>
              </bk-form-item>
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
                v-model="enabledPluginList" v-model:is-show="isPluginConfigManageSliderVisible"
              />
            </div>
          </div>
        </form-card>
      </div>
    </main>
    <form-page-footer @cancel="handleCancelClick" @submit="handleSubmit" />
  </div>
</template>

<script lang="ts" setup>
import { IRoute, IRouteConfig } from '@/types/route';
import { HTTP_METHODS_MAP } from '@/enum';
import UpstreamForm, { type IFlags } from '@/components/form/form-upstream.vue';
import { IUpstream, IUpstreamConfig } from '@/types/upstream';
import { useUpstreamForm } from '@/views/upstream/use-upstream-form';
import FormPageFooter from '@/components/form/form-page-footer.vue';
import FormUris from '@/components/form/form-uris.vue';
import { Form, InfoBox, Message } from 'bkui-vue';
import Ajv from 'ajv';
import addFormats from 'ajv-formats';
import ROUTE_JSON from '@/assets/schemas/route.json';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, onMounted, ref, useTemplateRef, watch } from 'vue';
import { getRoute, postRoute, putRoute } from '@/http/route';
import { getUpstreams } from '@/http/upstream';
import { getPluginConfigs } from '@/http/plugin-config';
import { IPluginConfigDto } from '@/types/plugin-config';
import { cloneDeep, isEmpty, uniq } from 'lodash-es';
import SelectUpstream from '@/components/select/select-upstream.vue';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
import SelectPluginConfig from '@/components/select/select-plugin-config.vue';
import SelectService from '@/components/select/select-service.vue';
import useConfigFilter from '@/hooks/use-config-filter';
import useResourcePageDetector from '@/hooks/use-resource-page-detector';
import useElementScroll from '@/hooks/use-element-scroll';
import FormCard from '@/components/form-card.vue';
import ButtonIcon from '@/components/button-icon.vue';
import ManagePluginConfigNew from '@/components/manage-plugin-config-new.vue';
import FormAdvancedSwitch from '@/components/form/form-advanced-switch.vue';
import FormLabelsNew from '@/components/form/form-labels-new.vue';
import FormHostsNew from '@/components/form/form-hosts-new.vue';
import FormRemoteAddrsNew from '@/components/form/form-remote-addrs-new.vue';
import VarsManagement from '@/components/vars-management/index.vue';

interface ILocalPlugin {
  doc_url?: string
  example?: string
  id?: string
  name: string
  config: string
  enabled?: boolean
}

const ajv = new Ajv();
addFormats(ajv);
const schemaValidate = ajv.compile(ROUTE_JSON);

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

const formModel = ref<Omit<IRoute, 'config'>>({
  name: '',
  upstream_id: '__config__',
  service_id: '__none__',
});

const routeConfig = ref<Partial<IRouteConfig>>({
  methods: [
    'GET',
    'POST',
  ],
  priority: 0,
  enable_websocket: false,
  hosts: [],
  uris: [''],
  remote_addrs: [],
  labels: {},
  desc: '',
  vars: [],
  // plugin_config_id: '',
});

const ssl_id = ref('');

const plugins = ref<Record<string, any>>({});
// const upstream = ref<Partial<IUpstreamConfig>>(createDefaultUpstream());
const upstream = ref<IUpstreamConfig>(createDefaultUpstream());

const uiConfig = ref({
  showAdvanced: false,
});
const flags = ref<IFlags>({
  upstreamType: 'nodes',
  tlsType: '__disabled__',
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const labelsFormNewRef = useTemplateRef('labels-form-new');
const urisFormRef = useTemplateRef('uris-form');
const hostsFormRef = useTemplateRef('hosts-form');
const remoteAddrsFormRef = useTemplateRef('remote-addrs-form');
const upstreamFormRef = useTemplateRef('upstream-form');
const varsManagementRef = useTemplateRef('vars-management');

const rules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'blur' },
  ],
  // desc: [
  //   { maxlength: 256 },
  // ],
};

const upstreamList = ref<IUpstream[]>([]);

const pluginConfigList = ref<IPluginConfigDto[]>([]);

const enabledPluginList = ref<ILocalPlugin[]>([]);

const isPluginConfigManageSliderVisible = ref(false);

const routeDtoId = computed(() => {
  return route.params.id as string;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getRoute({ id } as { id: string });
    const { config, service_id, ...rest } = response;
    delete rest.upstream_id;
    const upstream_id = response.upstream_id || config?.upstream_id || '';

    const {
      upstream: remoteUpstream,
      plugins: remotePlugins,
      ...restConfig
    } = config;

    await getDependencies();

    routeConfig.value = { ...routeConfig.value, ...restConfig };

    if (!restConfig.methods?.length) {
      routeConfig.value.methods = [];
    }

    if (service_id) {
      formModel.value.service_id = service_id;
      // 路由绑定了服务，上游自动改为手动填写
      if (!upstream_id) {
        formModel.value.upstream_id = '__config__';
      }
    }

    if (upstream_id) {
      formModel.value.upstream_id = upstream_id;
    }

    if (remoteUpstream) {
      // upstream.value = remoteUpstream;
      upstream.value = { ...upstream.value, ...remoteUpstream };

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

const handleServiceIdChanged = () => {
  // 选择了不绑定服务，则上游不允许为“不选择”, 选择了绑定的服务，则上游自动改为手动
  const isServiceNone = formModel.value.service_id === '__none__';
  const isUpstreamNone = formModel.value.upstream_id === '__none__';
  if (formModel.value.service_id  && ((isServiceNone && isUpstreamNone) || !isServiceNone)) {
    formModel.value.upstream_id = '__config__';
  }
};

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

const handleRemoteAddrs = async () => {
  const addrs = await remoteAddrsFormRef.value?.getValue() || [];

  if (!addrs.length || addrs.every(addr => isEmpty(addr))) {
    return Promise.resolve();
  }
  return await remoteAddrsFormRef.value?.validate();
};

const handleSubmit = async () => {
  try {
    const extraPlugins = enabledPluginList.value.reduce((result, plugin) => {
      result[plugin.name] = typeof plugin.config === 'string' ? JSON.parse(plugin.config) : plugin.config;
      return result;
    }, {} as Record<string, any>);

    let config: IRouteConfig = {
      ...cloneDeep(routeConfig.value),
      plugins: {
        ...plugins.value,
        ...extraPlugins,
      },
    };

    if (!config.plugin_config_id) {
      delete config.plugin_config_id;
    }

    if (formModel.value.upstream_id === '__config__') {
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

      config = { ...config, upstream: upstreamCopy };
      delete config.upstream_id;
    } else if (formModel.value.upstream_id === '__none__' && config.upstream_id) {
      delete config.upstream_id;
    } else {
      config.upstream_id = formModel.value.upstream_id;
    }

    // 校验表单
    await Promise.all([
      formRef.value?.validate(),
      handleLabels(),
      urisFormRef.value?.validate(),
      handleHosts(),
      handleRemoteAddrs(),
    ]);

    if (config.upstream) {
      await upstreamFormRef.value?.validate();
    }

    config.labels = await labelsFormNewRef.value.getValue() || {};

    config.uris = uniq(config.uris.filter(value => !!value));

    config.hosts = await hostsFormRef.value?.getValue() || routeConfig.value.hosts || [];
    config.hosts = uniq(config.hosts.filter(value => !!value) || []);
    if (!config.hosts.length) {
      delete config.hosts;
    }

    config.remote_addrs = await remoteAddrsFormRef.value?.getValue() || routeConfig.value.remote_addrs || [];
    config.remote_addrs = uniq(config.remote_addrs?.filter(value => !!value) || []);
    if (!config.remote_addrs.length) {
      delete config.remote_addrs;
    }

    config.vars = varsManagementRef.value?.getValue() || [];

    // 过滤值为空或默认值的字段
    // if (!isEditMode.value) {
    config = filterEmpty(config);
    config = filterAdvanced(config, 'route');
    if (config.upstream) {
      config.upstream = filterEmpty(config.upstream);
      config.upstream = filterAdvanced(config.upstream, 'upstream');
    }
    // }

    // 校验 schema
    if (schemaValidate(config)) {
      const data: Partial<IRoute> = {
        config,
        name: formModel.value.name,
      };

      if (formModel.value.service_id !== '__none__') {
        data.service_id = formModel.value.service_id;
      }

      // 既没选择“手动填写” upstream，也没选择“不选择”时才传入 upstream_id
      if (!['__none__', '__config__'].includes(formModel.value.upstream_id)) {
        data.upstream_id = formModel.value.upstream_id;
      } else {
        data.upstream_id = '';
      }

      if (routeConfig.value.plugin_config_id) {
        data.plugin_config_id = routeConfig.value.plugin_config_id;
      }

      InfoBox({
        title: t('确认提交？'),
        confirmText: t('提交'),
        cancelText: t('取消'),
        onConfirm: async () => {
          if (isEditMode.value) {
            await putRoute({
              data,
              id: routeDtoId.value,
            });
          } else {
            await postRoute({ data });
          }

          Message({
            theme: 'success',
            message: t('提交成功'),
          });

          await router.push({ name: 'route', replace: true });
        },
      });
    } else {
      showSchemaErrorMessages(schemaValidate.errors);
      // throw new Error;
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

const getUpstreamList = async () => {
  const response = await getUpstreams({ query: { offset: 0, limit: 100 } });
  upstreamList.value = response.results || [];
};

const handleUpstreamSelect = () => {
  if (!formModel.value.upstream_id || formModel.value.upstream_id === '__config__' || formModel.value.upstream_id === '__none__') {
    upstream.value = createDefaultUpstream();
    return;
  }

  const _upstream = upstreamList.value.find(upstream => upstream.id === formModel.value.upstream_id);
  const { config } = _upstream;
  upstream.value = cloneDeep(config);
};

const getPluginConfigList = async () => {
  const response = await getPluginConfigs();
  pluginConfigList.value = response.results as IPluginConfigDto[] || [];
};

const getDependencies = async () => {
  await Promise.all([
    getUpstreamList(),
    getPluginConfigList(),
  ]);
};

onMounted(async () => {
  if (isEditMode.value) {
    return;
  }
  await getDependencies();
});

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
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
