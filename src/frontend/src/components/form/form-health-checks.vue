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
  <form-collapse :title="t('健康检查')" style="width: 640px;">
    <div>
      <bk-form
        ref="form-ref"
        :model="checks"
        :rules="rules"
        class="form-element"
        v-bind="$attrs"
      >
        <!-- 主动检查 -->
        <bk-form-item class="form-item">
          <template #label><span style="font-weight: bold;">{{ t('主动检查') }}</span></template>
          <bk-switcher v-model="flags.active" theme="primary" @change="handleActiveCheckChange" />
        </bk-form-item>
        <template v-if="flags.active">
          <bk-form-item :label="t('类型')" class="form-item">
            <bk-select v-model="checks.active.type" :clearable="false" :filterable="false" style="width: 490px;">
              <bk-option
                v-for="type in healthCheckerTypeOptions"
                :id="type.id"
                :key="type.id"
                :name="type.name"
              />
            </bk-select>
          </bk-form-item>
          <bk-form-item :label="t('超时时间(s)')" class="form-item w180">
            <bk-input
              v-model="checks.active.timeout" :precision="1" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('并行数量')" class="form-item w180">
            <bk-input v-model="checks.active.concurrency" :min="0" :precision="0" :step="1" type="number" />
          </bk-form-item>
          <bk-form-item :label="t('主机名')" class="form-item w490" property="active.host" required>
            <bk-input v-model="checks.active.host" clearable />
          </bk-form-item>
          <bk-form-item :label="t('端口')" class="form-item w180">
            <bk-input v-model="checks.active.port" :max="65535" :min="1" :precision="0" :step="1" type="number" />
          </bk-form-item>
          <bk-form-item :label="t('请求路径')" class="form-item w490">
            <bk-input v-model="checks.active.http_path" clearable />
          </bk-form-item>
          <bk-form-item :label="t('请求头')" class="form-item">
            <req-headers-form v-model="checks.active.req_headers" />
          </bk-form-item>

          <div class="form-divider-title">
            {{ t('健康状态（主动）') }}
          </div>

          <bk-form-item :label="t('间隔时间(s)')" class="form-item w180" property="active.healthy.interval" required>
            <bk-input
              v-model="checks.active.healthy.interval" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('成功次数')" class="form-item w180" property="active.healthy.successes" required>
            <bk-input
              v-model="checks.active.healthy.successes" :max="254" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('状态码')" property="active.healthy.http_statuses" required>
            <form-tag-input-http-statuses
              v-model="checks.active.healthy.http_statuses" :rule="httpStatusesRule"
            />
          </bk-form-item>

          <div class="form-divider-title">
            {{ t('不健康状态（主动）') }}
          </div>

          <bk-form-item :label="t('超时时间(s)')" class="form-item w180" property="active.unhealthy.timeouts" required>
            <bk-input
              v-model="checks.active.unhealthy.timeouts" :max="254" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('间隔时间(s)')" class="form-item w180" property="active.unhealthy.interval" required>
            <bk-input
              v-model="checks.active.unhealthy.interval" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('状态码')" property="active.unhealthy.http_statuses" required>
            <form-tag-input-http-statuses v-model="checks.active.unhealthy.http_statuses" :rule="httpStatusesRule" />
          </bk-form-item>
          <bk-form-item
            :label="t('HTTP 失败次数')" class="form-item w180" property="active.unhealthy.http_failures" required
          >
            <bk-input
              v-model="checks.active.unhealthy.http_failures" :max="254" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('TCP 失败次数')" class="form-item w180" property="active.unhealthy.tcp_failures" required
          >
            <bk-input
              v-model="checks.active.unhealthy.tcp_failures" :max="254" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
        </template>

        <!-- 被动检查 -->

        <bk-form-item class="form-item">
          <template #label><span style="font-weight: bold;">{{ t('被动检查') }}</span></template>
          <bk-switcher v-model="flags.passive" theme="primary" @change="handlePassiveCheckChange" />
        </bk-form-item>
        <template v-if="flags.passive">
          <bk-form-item :label="t('类型')" class="form-item">
            <bk-select v-model="checks.passive.type" :clearable="false" :filterable="false" style="width: 490px;">
              <bk-option
                v-for="type in healthCheckerTypeOptions"
                :id="type.id"
                :key="type.id"
                :name="type.name"
              />
            </bk-select>
          </bk-form-item>

          <div class="form-divider-title">
            {{ t('健康状态（被动）') }}
          </div>

          <bk-form-item :label="t('状态码')" property="passive.healthy.http_statuses" required>
            <form-tag-input-http-statuses v-model="checks.passive.healthy.http_statuses" :rule="httpStatusesRule" />
          </bk-form-item>
          <bk-form-item :label="t('成功次数')" class="form-item w180" property="passive.healthy.successes" required>
            <bk-input
              v-model="checks.passive.healthy.successes" :max="254" :min="0" :precision="0" :step="1" type="number"
            />
          </bk-form-item>

          <div class="form-divider-title">
            {{ t('不健康状态（被动）') }}
          </div>

          <bk-form-item :label="t('超时时间(s)')" class="form-item w180" property="passive.unhealthy.timeouts" required>
            <bk-input
              v-model="checks.passive.unhealthy.timeouts" :max="254" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('TCP 失败次数')" class="form-item w180" property="passive.unhealthy.tcp_failures" required
          >
            <bk-input
              v-model="checks.passive.unhealthy.tcp_failures" :max="254" :min="0" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('HTTP 失败次数')" class="form-item w180" property="passive.unhealthy.http_failures" required
          >
            <bk-input
              v-model="checks.passive.unhealthy.http_failures" :max="254" :min="0" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('状态码')" property="passive.unhealthy.http_statuses" required>
            <form-tag-input-http-statuses v-model="checks.passive.unhealthy.http_statuses" :rule="httpStatusesRule" />
          </bk-form-item>
        </template>
      </bk-form>
    </div>
  </form-collapse>
</template>

<script lang="ts" setup>
import { Form, Message } from 'bkui-vue';
import { nextTick, ref, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import ReqHeadersForm from '@/components/form/form-req-headers.vue';
import { IHealthCheck } from '@/types/common';
import { useUpstreamForm } from '@/views/upstream/use-upstream-form';
import FormCollapse from '@/components/form-collapse.vue';
import FormTagInputHttpStatuses from '@/components/form/form-tag-input-http-statuses-.vue';
import { isInteger } from 'lodash-es';

const checks = defineModel<IHealthCheck>({
  // default: () => ({}),
  // set(value) {
  //   flags.value.active = !!value.active;
  //   flags.value.passive = !!value.passive;
  //   return value;
  // },
  get(value) {
    flags.value.active = !!value?.active;
    flags.value.passive = !!value?.passive;
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
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      pattern: /^\*?[0-9a-zA-Z-._[\]:]+$/,
      message: t('仅支持字母、数字、-、_和 *，但 * 需要在开头位置'),
      trigger: 'change',
    },
  ],
  'active.healthy.http_statuses': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string[]) => value.every(code => Number(code) >= 200 && Number(code) <= 599),
      message: t('状态码必须大于等于 200，小于等于 599'),
      trigger: 'change',
    },
  ],
  'active.unhealthy.http_statuses': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string[]) => value.every(code => Number(code) >= 200 && Number(code) <= 599),
      message: t('状态码必须大于等于 200，小于等于 599'),
      trigger: 'change',
    },
  ],
  'passive.healthy.http_statuses': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string[]) => value.every(code => Number(code) >= 200 && Number(code) <= 599),
      message: t('状态码必须大于等于 200，小于等于 599'),
      trigger: 'change',
    },
  ],
  'passive.unhealthy.http_statuses': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string[]) => value.every(code => Number(code) >= 200 && Number(code) <= 599),
      message: t('状态码必须大于等于 200，小于等于 599'),
      trigger: 'change',
    },
  ],
  'active.healthy.interval': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 1,
      message: t('必须是大于等于 1 的整数'),
      trigger: 'change',
    },
  ],
  'active.healthy.successes': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 1 && Number(value) <= 254,
      message: t('必须是大于等于 1，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
  'active.unhealthy.timeouts': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 1 && Number(value) <= 254,
      message: t('必须是大于等于 1，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
  'passive.unhealthy.timeouts': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 1 && Number(value) <= 254,
      message: t('必须是大于等于 1，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
  'passive.healthy.successes': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 0 && Number(value) <= 254,
      message: t('必须是大于等于 0，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
  'active.unhealthy.interval': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 1,
      message: t('必须是大于等于 1 的整数'),
      trigger: 'change',
    },
  ],
  'active.unhealthy.tcp_failures': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 1 && Number(value) <= 254,
      message: t('必须是大于等于 1，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
  'passive.unhealthy.tcp_failures': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 0 && Number(value) <= 254,
      message: t('必须是大于等于 0，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
  'active.unhealthy.http_failures': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 1 && Number(value) <= 254,
      message: t('必须是大于等于 1，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
  'passive.unhealthy.http_failures': [
    { required: true, message: t('必填项'), trigger: 'blur' },
    {
      validator: (value: string) => isInteger(Number(value)) && Number(value) >= 0 && Number(value) <= 254,
      message: t('必须是大于等于 0，小于等于 254 的整数'),
      trigger: 'change',
    },
  ],
};

const httpStatusesRule = {
  validator: (value: string) => Number(value) >= 200 && Number(value) <= 599,
  message: t('状态码必须大于等于 200，小于等于 599'),
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
  if (!checks.value) {
    checks.value = {};
  }
  nextTick(() => {
    if (value) {
      checks.value.active = createDefaultHealthCheck().active;
    } else {
      delete checks.value.active;
    }
  });
};

const handlePassiveCheckChange = (value: boolean) => {
  if (!checks.value) {
    checks.value = {};
  }
  nextTick(() => {
    if (value) {
      checks.value.passive = createDefaultHealthCheck().passive;
    } else {
      delete checks.value.passive;
    }
  });
};

const validate = async () => {
  if (!flags.value.active && flags.value.passive) {
    Message({
      theme: 'error',
      message: t('打开被动检查时也必须打开主动检查'),
    });
    return Promise.reject();
  }

  await formRef.value.validate();
};

defineExpose({
  validate,
});
</script>

<style lang="scss" scoped>

.form-divider-title {
  font-size: 14px;
  font-weight: 700;
  margin-bottom: 24px;
  margin-left: 150px;
  color: #313238;
}

</style>
