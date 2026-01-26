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
    <bk-tag-input
      v-model="localStatuses"
      :create-tag-validator="tagValidator"
      :placeholder="t('HTTP 状态码')"
      allow-create
      has-delete-icon
      :style="{ width: `${width}px` }"
    />
  </div>
</template>

<script lang="ts" setup>
import { Message } from 'bkui-vue';
import { isInteger } from 'lodash-es';
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

interface IProps {
  rule?: {
    validator: (value?: any) => boolean;
    message: string
  }
  width?: number | string;
}

const statuses = defineModel<number[]>({
  required: true,
});

const {
  rule = {
    validator: () => true,
    message: '',
  },
  width = 490,
} = defineProps<IProps>();

const localStatuses = ref(statuses.value.map(status => (String(status))));

const { t } = useI18n();

watch(localStatuses, () => {
  statuses.value = localStatuses.value.filter(status => isInteger(Number(status)))
    .map(status => Number(status));
}, { deep: true });

const tagValidator = (value: string) => {
  if (!/^[1-5][0-9][0-9]$/g.test(value)) {
    Message({
      theme: 'warning',
      // message: t('{code} 不是合法的 HTTP 状态码', { code: value }),
      message: `${value} 不是合法的 HTTP 状态码`,
    });
    return false;
  }

  if (rule.validator && !rule.validator(value)) {
    Message({
      theme: 'warning',
      message: rule.message || `${value} 不是合法的 HTTP 状态码`,
    });
    return false;
  }

  return true;
};

const validate = async () => {
  try {
    return localStatuses.value.every(status => /^[1-5][0-9][0-9]$/g.test(status));
  } catch {
    return false;
  }
};

defineExpose({
  validate,
});

</script>
