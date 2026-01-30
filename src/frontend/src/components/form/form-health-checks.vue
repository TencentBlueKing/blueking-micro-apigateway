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
        :model="localChecks"
        :rules="rules"
        class="form-element"
        v-bind="$attrs"
      >
        <checkbox-collapse
          v-model="flags.active"
          :desc="t('通过预设的探针类型，主动探测上游节点的存活性')"
          :name="t('主动检查')"
        >
          <bk-form-item :label="t('类型')" class="form-item">
            <bk-select v-model="localChecks.active.type" :clearable="false" :filterable="false" style="width: 474px;">
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
              v-model="localChecks.active.timeout" :precision="1" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('并行数量')" class="form-item w180">
            <bk-input v-model="localChecks.active.concurrency" :min="0" :precision="0" :step="1" type="number" />
          </bk-form-item>
          <bk-form-item :label="t('主机名')" class="form-item w474" property="active.host">
            <bk-input v-model="localChecks.active.host" clearable />
          </bk-form-item>
          <bk-form-item :label="t('端口')" class="form-item w180">
            <bk-input
              v-model="localChecks.active.port" :max="65535" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('请求路径')" class="form-item w474">
            <bk-input v-model="localChecks.active.http_path" clearable />
          </bk-form-item>
          <bk-form-item :label="t('请求头')" class="form-item">
            <req-headers-form v-model="localChecks.active.req_headers" input-width="414" />
          </bk-form-item>

          <div class="form-divider-title">
            {{ t('健康状态（主动）') }}
          </div>

          <bk-form-item :label="t('间隔时间(s)')" class="form-item w180" property="active.healthy.interval" required>
            <bk-input
              v-model="localChecks.active.healthy.interval" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('成功次数')" class="form-item w180" property="active.healthy.successes" required>
            <bk-input
              v-model="localChecks.active.healthy.successes" :max="254" :min="1" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('状态码')" property="active.healthy.http_statuses" required>
            <form-tag-input-http-statuses
              v-model="localChecks.active.healthy.http_statuses"
              :rule="httpStatusesRule"
              width="474"
            />
          </bk-form-item>

          <div class="form-divider-title">
            {{ t('不健康状态（主动）') }}
          </div>

          <bk-form-item
            :label="t('超时时间(s)')" class="form-item w180" property="active.unhealthy.timeouts" required
          >
            <bk-input
              v-model="localChecks.active.unhealthy.timeouts" :max="254" :min="1" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('间隔时间(s)')" class="form-item w180" property="active.unhealthy.interval" required
          >
            <bk-input
              v-model="localChecks.active.unhealthy.interval" :min="1" :precision="0" :step="1" type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('状态码')" property="active.unhealthy.http_statuses" required>
            <form-tag-input-http-statuses
              v-model="localChecks.active.unhealthy.http_statuses"
              :rule="httpStatusesRule"
              width="474"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('HTTP 失败次数')" class="form-item w180" property="active.unhealthy.http_failures" required
          >
            <bk-input
              v-model="localChecks.active.unhealthy.http_failures" :max="254" :min="1" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('TCP 失败次数')" class="form-item w180" property="active.unhealthy.tcp_failures" required
          >
            <bk-input
              v-model="localChecks.active.unhealthy.tcp_failures" :max="254" :min="1" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
        </checkbox-collapse>

        <!-- 被动检查 -->

        <checkbox-collapse
          v-model="flags.passive"
          :desc="t('通过实际请求的响应状态判断节点健康情况，无需额外探针请求，但可能会延迟问题发现，导致部分请求失败。')
            + t('由于不健康的节点无法收到请求，仅使用被动健康检查策略无法重新将节点标记为健康，因此通常需要结合主动健康检查策略。')"
          :name="t('被动检查')"
          style="margin-top: 12px;"
        >
          <bk-form-item :label="t('类型')" class="form-item">
            <bk-select
              v-model="localChecks.passive.type"
              :clearable="false"
              :filterable="false"
              style="width: 474px;"
            >
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
            <form-tag-input-http-statuses
              v-model="localChecks.passive.healthy.http_statuses"
              :rule="httpStatusesRule"
              width="474"
            />
          </bk-form-item>
          <bk-form-item :label="t('成功次数')" class="form-item w180" property="passive.healthy.successes" required>
            <bk-input
              v-model="localChecks.passive.healthy.successes" :max="254" :min="0" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>

          <div class="form-divider-title">
            {{ t('不健康状态（被动）') }}
          </div>

          <bk-form-item
            :label="t('超时时间(s)')" class="form-item w180" property="passive.unhealthy.timeouts" required
          >
            <bk-input
              v-model="localChecks.passive.unhealthy.timeouts" :max="254" :min="1" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('TCP 失败次数')" class="form-item w180" property="passive.unhealthy.tcp_failures" required
          >
            <bk-input
              v-model="localChecks.passive.unhealthy.tcp_failures" :max="254" :min="0" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
          <bk-form-item
            :label="t('HTTP 失败次数')" class="form-item w180" property="passive.unhealthy.http_failures" required
          >
            <bk-input
              v-model="localChecks.passive.unhealthy.http_failures" :max="254" :min="0" :precision="0" :step="1"
              type="number"
            />
          </bk-form-item>
          <bk-form-item :label="t('状态码')" property="passive.unhealthy.http_statuses" required>
            <form-tag-input-http-statuses
              v-model="localChecks.passive.unhealthy.http_statuses"
              :rule="httpStatusesRule"
              width="474"
            />
          </bk-form-item>
        </checkbox-collapse>
      </bk-form>
    </div>
  </form-collapse>
</template>

<script lang="ts" setup>
import { Form, Message } from 'bkui-vue';
import { ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import ReqHeadersForm from '@/components/form/form-req-headers.vue';
import { IHealthCheck } from '@/types/common';
import { createDefaultHealthCheck } from '@/views/upstream/use-upstream-form';
import FormCollapse from '@/components/form-collapse.vue';
import FormTagInputHttpStatuses from '@/components/form/form-tag-input-http-statuses-.vue';
import { cloneDeep, isInteger, isPlainObject } from 'lodash-es';
import CheckboxCollapse from '@/components/checkbox-collapse.vue';

interface IProps {
  checks?: IHealthCheck;
}

const { checks = undefined } = defineProps<IProps>();

const localChecks = ref<IHealthCheck>(createDefaultHealthCheck());

const flags = ref({
  active: false,
  passive: false,
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');

const { t } = useI18n();

const rules = {
  'active.host': [
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

watch(() => checks, () => {
  if (checks && isPlainObject(checks)) {
    flags.value.active = checks.active !== undefined;
    flags.value.passive = checks.passive !== undefined;
    Object.assign(localChecks.value, cloneDeep(checks));
  }
}, { immediate: true, deep: true });

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
  getValue: () => {
    const checks = {};
    if (flags.value.active) {
      Object.assign(checks, { active: localChecks.value.active });
    }
    if (flags.value.passive) {
      Object.assign(checks, { passive: localChecks.value.passive });
    }
    return checks;
  },
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
