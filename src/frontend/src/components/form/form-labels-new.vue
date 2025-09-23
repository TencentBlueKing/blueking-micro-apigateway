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
      v-for="(label, index) in localLabels" :key="index" ref="forms"
      :model="label"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <div class="multi-column-wrapper">
        <bk-form-item
          class="form-item w310"
          label=""
          property="key"
        >
          <bk-input v-model="label.key" :placeholder="t('键')" clearable />
        </bk-form-item>
        <div class="equal-mark">=</div>
        <bk-form-item
          class="form-item w310"
          label=""
          label-width="0"
          property="value"
        >
          <section class="multi-line-item has-suffix">
            <bk-input v-model="label.value" :placeholder="t('值')" clearable />
            <div class="suffix-actions">
              <icon
                v-if="localLabels.length > 1"
                class="icon-btn" color="#979BA5" name="minus-circle-shape" size="18" @click="handleRemoveItem(index)"
              />
              <icon
                class="icon-btn" color="#979BA5" name="plus-circle-shape" size="18" @click="handleAddItem"
              />
            </div>
          </section>
        </bk-form-item>
      </div>
    </bk-form>
    <icon
      v-if="!localLabels?.length"
      class="icon-btn" color="#979BA5" name="plus-circle-shape" size="18" @click="handleAddItem"
    />
  </div>
</template>

<script lang="ts" setup>
import { Form } from 'bkui-vue';
import { ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { isEmpty } from 'lodash-es';
import { ILabels } from '@/types';
import Icon from '@/components/icon.vue';

interface IProps {
  labels?: ILabels;
}

const { labels = {} } = defineProps<IProps>();

const localLabels = ref<{ key: string, value: string }[]>([]);

const formRefs = useTemplateRef<InstanceType<typeof Form>[]>('forms');

const { t } = useI18n();

const rules = {
  key: [
    { pattern: /^\S+$/, message: t('不能以空格开头或结尾'), trigger: 'change' },
  ],
  value: [
    { pattern: /^\S+$/, message: t('不能以空格开头或结尾'), trigger: 'change' },
  ],
};

watch(() => labels, () => {
  if (isEmpty(labels)) {
    localLabels.value = [{ key: '', value: '' }];
    return;
  }
  localLabels.value = Object.entries(labels)
    .map(([key, value]) => ({ key, value }));
}, { deep: true, immediate: true });

const handleAddItem = () => {
  localLabels.value.push({
    key: '',
    value: '',
  });
};

const handleRemoveItem = (index: number) => {
  localLabels.value.splice(index, 1);
};

const validate = async () => {
  await Promise.all(formRefs.value.map(formRef => formRef.validate()));
};

const getValue = async () => {
  const value = localLabels.value.reduce((result, { key, value }) => {
    if (key || value) {
      result[key] = value;
    }
    return result;
  }, {} as ILabels);
  return Promise.resolve(value);
};

defineExpose({
  validate,
  getValue,
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

.equal-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 30px;
}

</style>
