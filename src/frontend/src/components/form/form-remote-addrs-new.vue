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
      v-for="(addr, index) in localAddrs" :key="index" ref="forms"
      :model="addr"
      :rules="rules"
      class="form-element"
      v-bind="$attrs"
    >
      <bk-form-item
        :class="{ 'mb-0': localAddrs?.length > 1 && index === localAddrs.length - 1 }"
        class="form-item w640" label="" property="value"
      >
        <div class="multi-line-wrapper">
          <section class="multi-line-item has-suffix">
            <bk-input v-model="addr.value" clearable />
            <div class="suffix-actions" v-if="showAddIcon">
              <icon
                v-if="localAddrs.length > 1"
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
      v-if="!localAddrs?.length && showAddIcon"
      class="icon-btn" color="#979BA5" name="plus-circle-shape" size="18" @click="handleAddItem"
    />
    <div :class="{ 'mt--24': localAddrs?.length === 1 }" class="common-form-tips form-tips">
      <slot name="tooltips">
        {{ t('客户端与服务器握手时 IP，即客户端 IP，例如：192.168.1.101，192.168.1.0/24，::1，fe80::1，fe80::1/64') }}
      </slot>
    </div>
  </div>
</template>

<script lang="ts" setup>
import Icon from '@/components/icon.vue';
import { Form } from 'bkui-vue';
import { ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { isEmpty } from 'lodash-es';

interface IProps {
  addrs?: string[];
  showAddIcon?: true
}

const { addrs = [''], showAddIcon = true } = defineProps<IProps>();

const localAddrs = ref<{ value: string }[]>([]);

const formRefs = useTemplateRef<InstanceType<typeof Form>[]>('forms');

const { t } = useI18n();

const IPv4_REGEX = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
const IPv4_CIDR_REGEX = /^([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\/([12]?[0-9]|3[0-2])$/;
const IPv6_REGEX = /^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/;
const IPv6_CIDR_REGEX = /^([a-fA-F0-9]{0,4}:){1,8}(:[a-fA-F0-9]{0,4}){0,8}([a-fA-F0-9]{0,4})?\/[0-9]{1,3}$/;

const rules = {
  value: [
    {
      validator: (value: string) => (IPv4_REGEX.test(value)
        || IPv4_CIDR_REGEX.test(value)
        || IPv6_REGEX.test(value)
        || IPv6_CIDR_REGEX.test(value)),
      message: t('不是 IPv4 或 IPv6 格式'),
      trigger: 'change',
    },
  ],
};

watch(() => addrs, () => {
  if (isEmpty(addrs)) {
    localAddrs.value = [{ value: '' }];
    return;
  }
  localAddrs.value = addrs.map(addr => ({ value: addr }));
}, { deep: true, immediate: true });

const handleAddItem = () => {
  localAddrs.value.push({ value: '' });
};

const handleRemoveItem = (index: number) => {
  localAddrs.value.splice(index, 1);
};

const validate = async () => {
  return await Promise.all(formRefs.value.map(formRef => formRef.validate()));
};

const getValue = async () => {
  const value = localAddrs.value.reduce<string[]>((result, { value }) => {
    if (value) {
      result.push(value);
    }
    return result;
  }, []);
  return Promise.resolve(value || []);
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
