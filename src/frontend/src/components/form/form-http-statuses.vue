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
      v-for="(status, index) in localStatuses" :key="index" ref="forms"
      :model="status"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <bk-form-item class="form-item w480" label="" property="value">
        <div class="multi-line-wrapper">
          <section class="multi-line-item">
            <bk-input v-model="status.value" clearable />
            <icon
              v-if="localStatuses.length > 1" class="icon-btn" name="minus-circle" size="18"
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
import FormItemButton from '@/components/form/form-item-button.vue';
import Icon from '@/components/icon.vue';
import { Form } from 'bkui-vue';
import { isInteger } from 'lodash-es';
import { ref, useTemplateRef, watch } from 'vue';

const statuses = defineModel<number[]>({
  required: true,
});

const localStatuses = ref(statuses.value.map(status => ({ value: String(status) })));

const formRefs = useTemplateRef<InstanceType<typeof Form>[]>('forms');

// const { t } = useI18n();

const rules = {
  // value: [
  //   {
  //     pattern: new RegExp('^\*?[0-9a-zA-Z-._[\]:]+$'),
  //     message: t('仅支持字母、数字、-、_和 *，但 * 需要在开头位置'),
  //     trigger: 'change',
  //   },
  // ],
};

watch(localStatuses, () => {
  statuses.value = localStatuses.value.filter(status => isInteger(Number(status.value)))
    .map(status => Number(status.value));
}, { deep: true });

const handleAddItem = () => {
  localStatuses.value.push({
    value: '',
  });
};

const handleRemoveItem = (index: number) => {
  localStatuses.value.splice(index, 1);
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
