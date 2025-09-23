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
      v-for="(host, index) in hosts" :key="index" ref="forms"
      :model="{ value: host }"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <bk-form-item
        :class="{ 'mb-0': hosts?.length > 1 && index === hosts.length - 1 }"
        class="form-item w640" label="" property="value"
      >
        <div class="multi-line-wrapper">
          <section class="multi-line-item has-suffix">
            <bk-input v-model="hosts[index]" clearable />
            <div class="suffix-actions">
              <icon
                v-if="hosts.length > 1"
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
      v-if="!hosts?.length"
      class="icon-btn" color="#979BA5" name="plus-circle-shape" size="18" @click="handleAddItem"
    />
    <div :class="{ 'mt--24': hosts?.length === 1 }" class="common-form-tips form-tips">
      <slot name="tooltips">{{ t('路由匹配的域名列表。支持泛域名，如：*.test.com') }}</slot>
    </div>
  </div>
</template>

<script lang="ts" setup>
import Icon from '@/components/icon.vue';
import { Form } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useTemplateRef } from 'vue';

const hosts = defineModel<string[]>();

const formRefs = useTemplateRef<InstanceType<typeof Form>[]>('forms');

const { t } = useI18n();

const rules = {
  value: [
    {
      pattern: /^\*?[0-9a-zA-Z-._[\]:]+$/,
      message: t('仅支持字母、数字、-、_和 *，但 * 需要在开头位置'),
      trigger: 'change',
    },
  ],
};

const handleAddItem = () => {
  if (!hosts.value) {
    hosts.value = [''];
  } else {
    hosts.value.push('');
  }
};

const handleRemoveItem = (index: number) => {
  hosts.value.splice(index, 1);
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

.form-item {
  :deep(.bk-form-error) {
    position: relative;
  }
}

.form-tips {
  line-height: 1.2;
  position: relative;
  top: 4px;
  width: 640px;

  &.mt--24 {
    top: -20px;
  }
}

</style>
