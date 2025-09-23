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
    <bk-form
      ref="form-ref"
      :model="checks"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <form-divider>健康检查</form-divider>

      <!-- 主动检查 -->
      <bk-form-item :label="t('主动检查')" class="form-item">
        <bk-switcher v-model="flags.active" theme="primary" @change="handleActiveCheckChange" />
      </bk-form-item>
      <template v-if="flags.active">
        <bk-form-item :label="t('类型')" class="form-item">
          <bk-select v-model="checks.active.type" :clearable="false" :filterable="false">
            <bk-option
              v-for="type in healthCheckerTypeOptions"
              :id="type.id"
              :key="type.id"
              :name="type.name"
            />
          </bk-select>
        </bk-form-item>
        <bk-form-item :label="t('超时时间(s)')" class="form-item w240">
          <bk-input
            v-model="checks.active.timeout" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('并行数量')" class="form-item w240">
          <bk-input v-model="checks.active.concurrency" :precision="1" :step="1" type="number" />
        </bk-form-item>
        <bk-form-item :label="t('主机名')" class="form-item" property="active.host" required>
          <bk-input v-model="checks.active.host" clearable />
        </bk-form-item>
        <bk-form-item :label="t('端口')" class="form-item w240">
          <bk-input v-model="checks.active.port" :precision="1" :step="1" type="number" />
        </bk-form-item>
        <bk-form-item :label="t('请求路径')" class="form-item">
          <bk-input v-model="checks.active.http_path" clearable />
        </bk-form-item>
        <bk-form-item :label="t('请求头')" class="form-item">
          <req-headers-form v-model="checks.active.req_headers" />
        </bk-form-item>

        <form-divider>健康状态</form-divider>

        <bk-form-item :label="t('间隔时间(s)')" class="form-item w240" required>
          <bk-input
            v-model="checks.active.healthy.interval" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('成功次数')" class="form-item w240" required>
          <bk-input
            v-model="checks.active.healthy.successes" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('状态码')" required>
          <http-statuses-form v-model="checks.active.healthy.http_statuses" />
        </bk-form-item>

        <form-divider>不健康状态</form-divider>

        <bk-form-item :label="t('超时时间(s)')" class="form-item w240" required>
          <bk-input
            v-model="checks.active.unhealthy.timeouts" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('间隔时间(s)')" class="form-item w240" required>
          <bk-input
            v-model="checks.active.unhealthy.interval" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('状态码')" required>
          <http-statuses-form v-model="checks.active.unhealthy.http_statuses" />
        </bk-form-item>
        <bk-form-item :label="t('HTTP 失败次数')" class="form-item w240" required>
          <bk-input
            v-model="checks.active.unhealthy.http_failures" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('TCP 失败次数')" class="form-item w240" required>
          <bk-input
            v-model="checks.active.unhealthy.tcp_failures" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
      </template>

      <!-- 被动检查 -->
      <form-divider v-if="flags.active">被动检查</form-divider>

      <bk-form-item :label="t('被动检查')" class="form-item">
        <bk-switcher v-model="flags.passive" theme="primary" @change="handlePassiveCheckChange" />
      </bk-form-item>
      <template v-if="flags.passive">
        <bk-form-item :label="t('类型')" class="form-item">
          <bk-select v-model="checks.passive.type" :clearable="false" :filterable="false">
            <bk-option
              v-for="type in healthCheckerTypeOptions"
              :id="type.id"
              :key="type.id"
              :name="type.name"
            />
          </bk-select>
        </bk-form-item>

        <form-divider>健康状态</form-divider>

        <bk-form-item :label="t('状态码')" required>
          <http-statuses-form v-model="checks.passive.healthy.http_statuses" />
        </bk-form-item>
        <bk-form-item :label="t('成功次数')" class="form-item w240" required>
          <bk-input
            v-model="checks.passive.healthy.successes" :precision="1" :step="1" type="number"
          />
        </bk-form-item>

        <form-divider>不健康状态</form-divider>

        <bk-form-item :label="t('超时时间(s)')" class="form-item w240" required>
          <bk-input
            v-model="checks.passive.unhealthy.timeouts" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('TCP 失败次数')" class="form-item w240" required>
          <bk-input
            v-model="checks.passive.unhealthy.tcp_failures" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('HTTP 失败次数')" class="form-item w240" required>
          <bk-input
            v-model="checks.passive.unhealthy.http_failures" :precision="1" :step="1" type="number"
          />
        </bk-form-item>
        <bk-form-item :label="t('状态码')" required>
          <http-statuses-form v-model="checks.passive.unhealthy.http_statuses" />
        </bk-form-item>
      </template>
    </bk-form>
  </div>
</template>

<script lang="ts" setup>
import { Form, Message } from 'bkui-vue';
import { ref, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import ReqHeadersForm from '@/components/form-req-headers.vue';
import HttpStatusesForm from '@/components/form-http-statuses.vue';
import FormDivider from '@/components/form-divider.vue';
import { IHealthCheck } from '@/types/common';
import { useUpstreamForm } from '@/views/upstream/use-upstream-form';

const checks = defineModel<IHealthCheck>({
  default: () => ({}),
  // set(value) {
  //   flags.value.active = !!value.active;
  //   flags.value.passive = !!value.passive;
  //   return value;
  // },
  get(value) {
    flags.value.active = !!value.active;
    flags.value.passive = !!value.passive;
    return value;
  },
});

const { createDefaultHealthCheck } = useUpstreamForm();

const flags = ref({
  active: false,
  passive: false,
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');

const { t } = useI18n();

const rules = {
  'active.host': [
    { required: true, message: '必填项', trigger: 'blur' },
    {
      pattern: /^\*?[0-9a-zA-Z-._[\]:]+$/,
      message: t('仅支持字母、数字、-、_和 *，但 * 需要在开头位置'),
      trigger: 'change',
    },
  ],
};

const healthCheckerTypeOptions = [
  {
    id: 'http',
    name: 'HTTP',
  },
  {
    id: 'https',
    name: 'HTTPS',
  },
  {
    id: 'tcp',
    name: 'TCP',
  },
];

const handleActiveCheckChange = (value: boolean) => {
  if (value) {
    checks.value.active = createDefaultHealthCheck().active;
  } else {
    delete checks.value.active;
  }
};

const handlePassiveCheckChange = (value: boolean) => {
  if (value) {
    checks.value.passive = createDefaultHealthCheck().passive;
  } else {
    delete checks.value.passive;
  }
};

const validate = async () => {
  if (!flags.value.active) {
    Message({
      theme: 'error',
      message: t('必须打开主动检查'),
    });
    return false;
  }

  if (!flags.value.active && flags.value.passive) {
    Message({
      theme: 'error',
      message: t('打开被动检查时也必须打开主动检查'),
    });
    return false;
  }

  try {
    await formRef.value.validate();
    return true;
  } catch {
    return false;
  }
};

defineExpose({
  validate,
});
</script>
