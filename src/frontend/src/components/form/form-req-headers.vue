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
      v-for="(req_header, index) in localReqHeaders" :key="index" ref="forms"
      :model="req_header"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <bk-form-item
        :class="{ 'mb-0': localReqHeaders?.length > 1 && index === localReqHeaders.length - 1 }"
        class="form-item" label="" property="value"
      >
        <div class="multi-line-wrapper">
          <section class="multi-line-item has-suffix">
            <bk-input v-model="req_header.value" clearable />
            <div class="suffix-actions">
              <icon
                v-if="localReqHeaders.length > 1"
                class="icon-btn" color="#979BA5" name="minus-circle-shape" size="18" @click="handleRemoveItem(index)"
              />
              <icon
                class="icon-btn" color="#979BA5" name="plus-circle-shape" size="18" @click="handleAddItem"
              />
            </div>
          </section>
        </div>
      </bk-form-item>
    </bk-form>
    <icon
      v-if="!localReqHeaders?.length"
      class="icon-btn" color="#979BA5" name="plus-circle-shape" size="18" @click="handleAddItem"
    />
  </div>
</template>

<script lang="ts" setup>
import Icon from '@/components/icon.vue';
import { Form } from 'bkui-vue';
import { ref, useTemplateRef, watch } from 'vue';
import { isEqual } from 'lodash-es';

const req_headers = defineModel<string[]>({
  required: true,
  default: () => [],
});

const localReqHeaders = ref<{ value: string }[]>([]);

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

watch(req_headers, (value, oldValue) => {
  if (isEqual(value, oldValue)) {
    return;
  }
  localReqHeaders.value = req_headers.value.map(req_header => ({ value: req_header }));
}, { deep: true, immediate: true });

watch(localReqHeaders, () => {
  req_headers.value = localReqHeaders.value.map(req_header => req_header.value);
}, { deep: true });

const handleAddItem = () => {
  localReqHeaders.value.push({
    value: '',
  });
};

const handleRemoveItem = (index: number) => {
  localReqHeaders.value.splice(index, 1);
};

const validate = async () => {
  try {
    await Promise.all(formRefs.value.map(formRef => formRef.validate()));
    return true;
  } catch {
    return false;
  }
};

defineExpose({
  validate,
});

</script>

<style lang="scss" scoped>

.has-suffix {
  position: relative;

  .suffix-actions {
    position: absolute;
    right: -12px;
    display: flex;
    align-items: center;
    transform: translateX(100%);
    gap: 12px;
  }
}

.form-item.mb-0 {
  margin-bottom: 0;
}

</style>
