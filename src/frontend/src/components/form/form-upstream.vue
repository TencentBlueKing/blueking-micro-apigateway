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
    <!--  upstream 配置  -->
    <bk-form ref="bk-form" :model="upstream" :rules="rules" class="form-element">
      <template v-if="descAndLabels">
        <!-- desc -->
        <bk-form-item :label="t('描述')" class="form-item" property="desc">
          <bk-input v-model="upstream.desc" :placeholder="t('描述')" clearable />
        </bk-form-item>

        <!-- labels -->
        <bk-form-item :label="t('标签')">
          <form-labels-new ref="labels-form-new" :labels="upstream.labels" />
        </bk-form-item>
      </template>

      <!-- type -->
      <bk-form-item :label="t('负载均衡算法')" class="form-item" property="type" required>
        <bk-select v-model="upstream.type" :clearable="false" :filterable="false" @change="handleTypeChange">
          <bk-option
            v-for="type in typeOptions"
            :id="type.id"
            :key="type.id"
            :name="type.name"
          />
        </bk-select>
      </bk-form-item>

      <!-- hash_on -->
      <bk-form-item
        v-if="upstream.type === 'chash'" :label="t('哈希位置')" class="form-item" property="hash_on" required
      >
        <bk-select v-model="upstream.hash_on" :clearable="false" :filterable="false" @change="handleHashOnChange">
          <bk-option
            v-for="type in hashOnOptions"
            :id="type.id"
            :key="type.id"
            :name="type.name"
          />
        </bk-select>
      </bk-form-item>

      <!-- key -->
      <bk-form-item
        v-if="upstream.type === 'chash' && upstream.hash_on !== 'consumer'" class="form-item" label="Key"
        property="key" required
      >
        <bk-dropdown
          placement="bottom-start"
          style="width: 100%;"
        >
          <bk-input v-model="upstream.key" />
          <template v-if="upstream.hash_on !== 'header' && upstream.hash_on !== 'cookie'" #content>
            <bk-dropdown-menu ext-cls="dropdown-menu-wrapper">
              <bk-dropdown-item
                v-for="option in hashOnKeyOptions"
                :key="option.id" @click="handleHashOnKeyClick(option.id)"
              >
                {{ option.name }}
              </bk-dropdown-item>
            </bk-dropdown-menu>
          </template>
        </bk-dropdown>
      </bk-form-item>

      <!-- 上游类型 -->
      <bk-form-item :label="t('上游类型')" class="form-item" required>
        <bk-select
          v-model="flags.upstreamType" :clearable="false" :filterable="false" @change="handleUpstreamTypeChange"
        >
          <bk-option
            v-for="type in upstreamTypeOptions"
            :id="type.id"
            :key="type.id"
            :name="type.name"
          />
        </bk-select>
      </bk-form-item>

      <!-- nodes -->
      <bk-form-item v-if="flags.upstreamType === 'nodes'" label="目标节点">
        <form-table-nodes ref="nodes-table-form" v-model="upstream.nodes"></form-table-nodes>
      </bk-form-item>

      <template v-if="flags.upstreamType === 'service_discovery'">
        <bk-form-item :label="t('服务发现类型')" class="form-item" property="discovery_type" required>
          <bk-select v-model="upstream.discovery_type" :clearable="false" :filterable="false">
            <bk-option
              v-for="type in serviceDiscoveryTypeOptions"
              :id="type.id"
              :key="type.id"
              :name="type.name"
            />
          </bk-select>
          <div class="common-form-tips" style="line-height: 1.2;margin-top: 4px;">{{
            t('必须确保选中类型在部署 APISIX 时有配置 `conf/config.yaml` 的 `discovery`, 具体参考 集成服务发现注册中心 https://apisix.apache.org/zh/docs/apisix/discovery/')
          }}
          </div>
        </bk-form-item>
        <bk-form-item :label="t('服务名称')" class="form-item" property="service_name" required>
          <bk-input v-model="upstream.service_name" clearable />
          <div class="common-form-tips" style="line-height: 1.2;margin-top: 4px;">{{
            t('必填, 请填写对应服务注册中心中配置的服务名称')
          }}
          </div>
        </bk-form-item>
      </template>

      <!-- scheme -->
      <bk-form-item :label="t('协议')" class="form-item" property="scheme">
        <bk-select v-model="upstream.scheme" :clearable="false" :filterable="false" @change="handleSchemeChange">
          <bk-option
            v-for="type in schemeOptions"
            :id="type.id"
            :key="type.id"
            :name="type.name"
          />
        </bk-select>
      </bk-form-item>

      <!-- TLS & SSL -->
      <bk-form-item
        v-if="upstream.scheme === 'https' || upstream.scheme === 'grpcs'"
        :label="t('TLS')"
        class="form-item"
      >
        <bk-select v-model="flags.tlsType" :clearable="false" :filterable="false" @change="handleTlsChange">
          <bk-option id="__disabled__" :name="t('禁用')" />
          <bk-option id="__input__" :name="t('输入证书')" />
          <bk-option id="__select__" :name="t('关联已有证书')" />
        </bk-select>
      </bk-form-item>

      <template v-if="flags.tlsType === '__input__'">
        <bk-form-item :label="t('方式')" class="form-item">
          <bk-radio-group v-model="sslInputType" type="card" @change="handleTypeChange">
            <bk-radio-button label="input">{{ t('输入') }}</bk-radio-button>
            <bk-radio-button label="upload">{{ t('上传') }}</bk-radio-button>
          </bk-radio-group>
        </bk-form-item>

        <bk-form-item v-if="sslInputType === 'upload'" :label="t('上传客户端证书')" class="form-item" property="cert">
          <upload-text @done="(content: string) => handleSSLUpload(content, 'cert')" />
        </bk-form-item>

        <bk-form-item :label="t('客户端证书')" class="form-item" property="tls.client_cert" required>
          <bk-input
            v-model="upstream.tls.client_cert"
            :autosize="uiConfig.autoSizeConf"
            :clearable="false"
            :placeholder="t('请输入客户端证书')"
            :rows="4"
            type="textarea"
          />
        </bk-form-item>

        <bk-form-item v-if="sslInputType === 'upload'" :label="t('上传客户端私钥')" class="form-item" property="cert">
          <upload-text @done="(content: string) => handleSSLUpload(content, 'key')" />
        </bk-form-item>

        <bk-form-item :label="t('客户端私钥')" class="form-item" property="tls.client_key" required>
          <bk-input
            v-model="upstream.tls.client_key"
            :autosize="uiConfig.autoSizeConf"
            :clearable="false"
            :placeholder="t('请输入客户端私钥')"
            :rows="4"
            type="textarea"
          />
        </bk-form-item>
      </template>

      <bk-form-item
        v-if="flags.tlsType === '__select__'"
        :label="t('客户端证书')"
        class="form-item"
        property="tls.client_cert_id"
        required
      >
        <select-ssl v-model="ssl_id" />
      </bk-form-item>

      <!-- pass_host -->
      <bk-form-item :label="t('Host 请求头')" class="form-item">
        <bk-select v-model="upstream.pass_host" :clearable="false" :filterable="false">
          <bk-option
            v-for="type in passHostOptions"
            :id="type.id"
            :key="type.id"
            :name="type.name"
          />
        </bk-select>
      </bk-form-item>

      <form-advanced-switch v-model="uiConfig.showAdvanced" />

      <!-- 高级配置 -->
      <template v-if="uiConfig.showAdvanced">
        <form-collapse :title="t('重试')" style="margin-left: 150px;width: 640px;">
          <div>
            <!-- retries -->
            <bk-form-item :label="t('重试次数')" class="form-item w180" property="retries">
              <bk-input
                v-model="upstream.retries" :min="0" :placeholder="t('重试次数')" :precision="0" :step="1" type="number"
              />
            </bk-form-item>

            <!-- retry_timeout -->
            <bk-form-item :label="t('重试超时时间')" class="form-item w180" property="retry_timeout">
              <bk-input
                v-model="upstream.retry_timeout" :placeholder="t('重试超时时间')" :precision="1" :step="1" type="number"
              />
            </bk-form-item>
          </div>
        </form-collapse>

        <form-collapse :title="t('超时时间')" style="margin-left: 150px;width: 640px;">
          <!-- timeout.connect -->
          <bk-form-item :label="t('连接超时(s)')" class="form-item w180" property="timeout.connect">
            <bk-input
              v-model="upstream.timeout.connect" :precision="1" :step="1" type="number"
            />
          </bk-form-item>

          <!-- timeout.send -->
          <bk-form-item :label="t('发送超时(s)')" class="form-item w180" property="timeout.send">
            <bk-input
              v-model="upstream.timeout.send" :precision="1" :step="1" type="number"
            />
          </bk-form-item>

          <!-- timeout.read -->
          <bk-form-item :label="t('接收超时(s)')" class="form-item w180" property="timeout.read">
            <bk-input
              v-model="upstream.timeout.read" :precision="1" :step="1" type="number"
            />
          </bk-form-item>
        </form-collapse>

        <form-collapse :title="t('连接池')" style="margin-left: 150px;width: 640px;">
          <bk-form-item :label="t('容量')" class="form-item w180">
            <bk-input v-model="upstream.keepalive_pool.size" :precision="1" :step="1" clearable type="number" />
          </bk-form-item>
          <bk-form-item :label="t('空闲超时时间')" class="form-item w180">
            <bk-input
              v-model="upstream.keepalive_pool.idle_timeout" :precision="1" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('请求数量')" class="form-item w180">
            <bk-input
              v-model="upstream.keepalive_pool.requests" :min="0" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
        </form-collapse>

        <!-- 健康检查 -->
        <div style="margin-left: 150px;">
          <form-health-checks ref="health-check-form" v-model="upstream.checks" />
        </div>
      </template>
    </bk-form>
  </div>
</template>

<script lang="ts" setup>
import { IUpstreamConfig } from '@/types/upstream';
import { Form } from 'bkui-vue';
import { ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import FormHealthChecks from '@/components/form/form-health-checks.vue';
import { useUpstreamForm } from '@/views/upstream/use-upstream-form';
import useElementScroll from '@/hooks/use-element-scroll';
import FormTableNodes from '@/components/form/form-table-nodes.vue';
import FormCollapse from '@/components/form-collapse.vue';
import FormAdvancedSwitch from '@/components/form/form-advanced-switch.vue';
import FormLabelsNew from '@/components/form/form-labels-new.vue';
import { isEmpty } from 'lodash-es';
import SelectSsl from '@/components/select/select-ssl.vue';
import UploadText from '@/components/upload-text.vue';
import { type UploadFieldType } from '@/views/ssl/create.vue';

export interface IFlags {
  upstreamType: 'nodes' | 'service_discovery';
  // TLS 类型：禁用 | 输入 | 关联已有证书
  tlsType: '__disabled__' | '__input__' | '__select__';
}

interface IProps {
  descAndLabels?: boolean;
}

const upstream = defineModel<Partial<IUpstreamConfig>>({
  required: true,
  default: () => ({}),
});

const flags = defineModel<IFlags>('flags', {
  default: () => ({
    upstreamType: 'nodes',
    tlsType: '__disabled__',
  }),
});

const ssl_id = defineModel<string>('ssl_id', {
  default: '',
});

const {
  // 是否需要填写描述和标签
  descAndLabels = true,
} = defineProps<IProps>();

const { t } = useI18n();

const { createDefaultUpstream } = useUpstreamForm();
const { showFirstErrorFormItem } = useElementScroll();

const formRef = useTemplateRef<InstanceType<typeof Form>>('bk-form');
const nodesTableFormRef = useTemplateRef<InstanceType<typeof FormTableNodes>>('nodes-table-form');
const healthCheckFormRef = useTemplateRef('health-check-form');
const labelsFormNewRef = useTemplateRef('labels-form-new');

const typeOptions = [
  {
    id: 'roundrobin',
    name: '带权轮询 (Round Robin)',
  },
  {
    id: 'chash',
    name: '一致性哈希（CHash）',
  },
  {
    id: 'ewma',
    name: '指数加权移动平均法（EWMA）',
  },
  {
    id: 'least_conn',
    name: '最小连接数（least_conn）',
  },
];

const hashOnOptions = [
  {
    id: 'vars',
    name: 'vars',
  },
  {
    id: 'header',
    name: 'header',
  },
  {
    id: 'cookie',
    name: 'cookie',
  },
  {
    id: 'consumer',
    name: 'consumer',
  },
  // {
  //   id: 'vars_combinations',
  //   name: 'vars_combinations',
  // },
];

const hashOnKeyOptions = [
  {
    id: 'uri',
    name: 'uri',
  },
  {
    id: 'server_name',
    name: 'server_name',
  },
  {
    id: 'server_addr',
    name: 'server_addr',
  },
  {
    id: 'request_uri',
    name: 'request_uri',
  },
  {
    id: 'remote_port',
    name: 'remote_port',
  },
  {
    id: 'remote_addr',
    name: 'remote_addr',
  },
  {
    id: 'query_string',
    name: 'query_string',
  },
  {
    id: 'host',
    name: 'host',
  },
  {
    id: 'hostname',
    name: 'hostname',
  },
  {
    id: 'arg_***',
    name: 'arg_***',
  },
];

const serviceDiscoveryTypeOptions = [
  {
    id: 'dns',
    name: 'DNS',
  },
  {
    id: 'consul_kv',
    name: 'Consul KV',
  },
  {
    id: 'nacos',
    name: 'Nacos',
  },
  {
    id: 'eureka',
    name: 'Eureka',
  },
  {
    id: 'kubernetes',
    name: 'Kubernetes',
  },
];

const passHostOptions = [
  {
    id: 'pass',
    name: '保持与客户端请求一致的主机名',
  },
  {
    id: 'node',
    name: '使用目标节点列表中的主机名或IP',
  },
];

const upstreamTypeOptions = [
  {
    id: 'nodes',
    name: '节点',
  },
  {
    id: 'service_discovery',
    name: '服务发现',
  },
];

const schemeOptions = [
  {
    id: 'http',
    name: 'HTTP',
  },
  {
    id: 'https',
    name: 'HTTPS',
  },
  {
    id: 'grpc',
    name: 'gRPC',
  },
  {
    id: 'grpcs',
    name: 'gRPCs',
  },
  {
    id: 'tcp',
    name: 'TCP',
  },
  {
    id: 'tls',
    name: 'TLS',
  },
  {
    id: 'udp',
    name: 'UDP',
  },
  {
    id: 'kafka',
    name: 'Kafka',
  },
];

const rules = {
  service_name: [{ required: true, message: t('必填，最大长度为256个字符'), trigger: 'blur' }],
  key: [
    {
      validator: (value: string) => {
        if (upstream.value.hash_on === 'header') {
          if (!value.startsWith('http_')) {
            return false;
          }
        }
        return true;
      },
      message: t('必须以 http_ 开头'),
      trigger: 'change',
    },
    {
      validator: (value: string) => {
        if (upstream.value.hash_on === 'cookie') {
          if (!value.startsWith('cookie_')) {
            return false;
          }
        }
        return true;
      },
      message: t('必须以 cookie_ 开头'),
      trigger: 'change',
    },
    {
      validator: (value: string) => {
        if (upstream.value.hash_on === 'vars') {
          if (value.startsWith('arg_')) {
            return /arg_[0-9a-zA-z_-]+/.test(value);
          }
        }
        return true;
      },
      message: t('arg_ 后只能跟数字、字母、下划线、减号'),
      trigger: 'change',
    },
  ],
  'tls.client_cert': [
    { required: true, message: t('证书不能为空'), trigger: 'blur' },
    {
      validator: (value: string) => value.length >= 128,
      message: t('证书内容至少需要128个字符'),
      trigger: 'change',
    },
  ],
  'tls.client_key': [
    { required: true, message: t('私钥不能为空'), trigger: 'blur' },
    { validator: (value: string) => value.length >= 128, message: t('私钥内容至少需要128个字符'), trigger: 'change' },
  ],
  'tls.client_cert_id': [
    { required: true, message: t('请选择客户端证书'), trigger: 'change' },
  ],
  // desc: [
  //   { maxlength: 256, message: t('最大长度为256个字符'), trigger: 'change' },
  // ],
};

const uiConfig = ref({
  showAdvanced: false,
  autoSizeConf: {
    minRows: 4,
    maxRows: 8,
  },
});

const sslInputType = ref<'input' | 'upload'>('input');

watch(ssl_id, (id) => {
  if (!upstream.value.tls) {
    upstream.value.tls = {};
  }
  if (id) {
    upstream.value.tls.client_cert_id = id;
  }
});

const handleUpstreamTypeChange = (value: string) => {
  if (value === 'service_discovery') {
    upstream.value.service_name = '';
    upstream.value.discovery_type = '';
    delete upstream.value.nodes;
  } else {
    upstream.value.nodes = createDefaultUpstream().nodes;
    delete upstream.value.service_name;
    delete upstream.value.discovery_type;
  }
};

const handleTypeChange = (value: string) => {
  if (value === 'chash') {
    upstream.value.hash_on = 'vars';
    upstream.value.key = 'remote_addr';
  } else {
    upstream.value.hash_on = '';
    upstream.value.key = '';
  }
};

const handleHashOnChange = (value: string) => {
  if (value === 'consumer') {
    upstream.value.key = '';
  }
  if (value === 'header') {
    upstream.value.key = 'http_';
  }
  if (value === 'cookie') {
    upstream.value.key = 'cookie_';
  }
};

const handleHashOnKeyClick = (key: string) => {
  upstream.value.key = key;
};

const handleSchemeChange = (value: string) => {
  if (value !== 'https' && value !== 'grpcs') {
    flags.value.tlsType = '__disabled__';
    upstream.value.tls = {};
  }
};

const handleTlsChange = (value: string) => {
  ssl_id.value = '';

  if (value === '__disabled__') {
    upstream.value.tls = {};
  } else if (value === '__input__') {
    upstream.value.tls = {
      client_cert: '',
      client_key: '',
    };
  } else if (value === '__select__') {
    upstream.value.tls = {
      client_cert_id: '',
    };
  }
};

const handleSSLUpload = (content: string, field: UploadFieldType) => {
  if (field === 'cert') {
    upstream.value.tls.client_cert = content;
  } else if (field === 'key') {
    upstream.value.tls.client_key = content;
  }
};

const handleLabels = async () => {
  if (!descAndLabels) {
    return Promise.resolve();
  }
  const labels = await labelsFormNewRef.value.getValue();

  if (isEmpty(labels)) {
    return Promise.resolve();
  }
  return await labelsFormNewRef.value.validate();
};

const setLabels = async () => {
  if (descAndLabels) {
    const labels = await labelsFormNewRef.value.getValue() || {};
    upstream.value.labels = labels;
    return labels;
  }
  return {};
};

const validate = async () => {
  try {
    const tasks = [
      formRef.value.validate(),
      handleLabels(),
    ];

    if (flags.value.upstreamType === 'nodes') {
      tasks.push(nodesTableFormRef.value.validate());
    }

    if (upstream.value.checks && Object.keys(upstream.value.checks).length && healthCheckFormRef.value) {
      tasks.push(healthCheckFormRef.value.validate());
    }

    await Promise.all(tasks);
    await setLabels();
  } catch (error) {
    showFirstErrorFormItem();
    throw error;
  }
};

defineExpose({
  validate,
  setLabels,
});

</script>

<style lang="scss" scoped>

.dropdown-menu-wrapper {
  width: 640px;
}

</style>
