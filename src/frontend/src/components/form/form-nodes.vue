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
      v-for="(node, index) in nodes" :key="index" ref="forms"
      :model="node"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <div class="multi-column-wrapper">
        <bk-form-item
          :label="t('主机名')"
          class="form-item w360"
          label-position="left"
          label-width="70"
          property="host"
          required
        >
          <bk-input v-model="node.host" clearable />
        </bk-form-item>
        <bk-form-item
          :label="t('端口')" class="form-item w240" label-width="90"
          property="port"
        >
          <bk-input v-model="node.port" :min="0" :precision="0" :step="1" clearable type="number" />
        </bk-form-item>
        <bk-form-item
          :label="t('权重')" class="form-item w240" label-width="90" property="weight"
          required
        >
          <section class="multi-line-item">
            <bk-input v-model="node.weight" :min="0" :precision="0" :step="1" type="number" />
            <icon
              v-if="nodes.length > 1" class="icon-btn" name="minus-circle" size="18"
              @click="handleRemoveItem(index)"
            />
          </section>
        </bk-form-item>
      </div>
    </bk-form>
    <form-item-button margin-top="0" @click="handleAddItem" />
  </div>
</template>

<script lang="ts" setup>
import FormItemButton from '@/components/form/form-item-button.vue';
import Icon from '@/components/icon.vue';
import { Form } from 'bkui-vue';
import { INode } from '@/types/common';
import { isInteger } from 'lodash-es';
import { useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';

const nodes = defineModel<INode[]>({
  required: true,
});

const formRefs = useTemplateRef<InstanceType<typeof Form>[]>('forms');

const { t } = useI18n();

const rules = {
  host: [
    { required: true, message: '请输入主机名或IP', trigger: 'blur' },
    {
      pattern: /^\*?[0-9a-zA-Z-._[\]:]+$/,
      message: '仅支持字母、数字、-、_和 *，但 * 需要在开头位置',
      trigger: 'change',
    },
  ],
  port: [
    {
      validator: (value: string) => value === '' || isInteger(Number(value)),
      message: '必须为整数',
      trigger: 'change',
    },
  ],
  weight: [
    { required: true, message: '请输入权重', trigger: 'blur' },
    {
      validator: (value: string) => value === '' || isInteger(Number(value)),
      message: '必须为整数',
      trigger: 'change',
    },
  ],
};

const handleAddItem = () => {
  nodes.value.push({
    host: '',
    port: null,
    weight: 1,
  });
};

const handleRemoveItem = (index: number) => {
  nodes.value.splice(index, 1);
};

const validate = async () => {
  try {
    await Promise.all(formRefs.value.map(nodeForm => nodeForm.validate()));
    return true;
  } catch {
    return false;
  }
};

defineExpose({
  validate,
});
</script>
