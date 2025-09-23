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
      v-for="(uri, index) in uris" :key="index" ref="forms"
      :model="{ value: uri }"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <bk-form-item
        class="form-item w480" label="" property="value"
      >
        <div class="multi-line-wrapper">
          <section class="multi-line-item">
            <bk-input v-model="uris[index]" clearable />
            <icon
              v-if="uris.length > 1" class="icon-btn" name="minus-circle" size="18"
              @click="handleRemoveItem(index)"
            />
          </section>
        </div>
      </bk-form-item>
    </bk-form>
    <form-item-button margin-top="0" @click="handleAddItem" />
  </div>
</template>

<script lang="ts" setup>
// @ts-ignore
import FormItemButton from '@/components/form/form-item-button.vue';
// @ts-ignore
import Icon from '@/components/icon.vue';
import { Form } from 'bkui-vue';
import { useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';

const uris = defineModel<string[]>();

const { t } = useI18n();

const formRefs = useTemplateRef<InstanceType<typeof Form>[]>('forms');

const rules = {
  value: [
    {
      required: true,
      message: t('必填项'),
      trigger: 'blur',
    },
    {
      validator: (value: string) => {
        const reg = /^https?:\/\//;
        return reg.test(value);
      },
      message: t('请输入 http:// 或者 https:// 开头的地址'),
      trigger: 'blur',
    },
  ],
};

const handleAddItem = () => {
  if (!uris.value) {
    uris.value = [''];
  } else {
    uris.value.push('');
  }
};

const handleRemoveItem = (index: number) => {
  uris.value.splice(index, 1);
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
