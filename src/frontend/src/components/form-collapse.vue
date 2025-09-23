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
  <bk-collapse v-model="activeIndex">
    <bk-collapse-panel :name="name">
      <template #header>
        <div class="collapse-panel-header">
          <Icon
            class="icon-down-shape"
            :class="{ 'active-icon': isPanelActive }"
            name="down-shape"
          />
          <slot
            v-if="slots.header"
            name="header"
          />
          <span
            v-else
            class="panel-title"
          >
            {{ title }}
          </span>
        </div>
      </template>
      <template #content>
        <div class="collapse-panel-content">
          <slot name="default" />
        </div>
      </template>
    </bk-collapse-panel>
  </bk-collapse>
</template>
<script lang="ts" setup>
import Icon from '@/components/icon.vue';
import { computed, ref, watch } from 'vue';

interface Props {
  title?: string;
  name?: string;
}

interface Emits {
  (e: 'toggle', value: boolean): void;
}

interface Slots {
  default: any;
  header: any;
}

interface Exposes {
  show: () => void;
  hide: () => void;
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  name: 'default',
});

const emits = defineEmits<Emits>();

const slots = defineSlots<Slots>();

const activeIndex = ref([props.name]);

const isPanelActive = computed(() => !activeIndex.value.includes(props.name));

watch(isPanelActive, () => {
  emits('toggle', isPanelActive.value);
});

defineExpose<Exposes>({
  show: () => {
    activeIndex.value = [props.name];
  },
  hide: () => {
    activeIndex.value = [];
  },
});

</script>

<style lang="scss" scoped>

.collapse-panel-header {
  position: relative;
  display: flex;
  align-items: center;
  height: 28px;
  padding-left: 8px;
  cursor: pointer;
  border-radius: 2px;
  background: #f5f7fa;

  :deep(.icon-down-shape) {
    transition: all 0.5s;
    transform: rotateZ(0deg);
    color: #313238;
  }

  .panel-title {
    font-size: 14px;
    font-weight: 700;
    margin-left: 5px;
    color: #313238;
  }

  .active-icon {
    transition: all 0.5s;
    transform: rotateZ(-90deg);
  }
}

.collapse-panel-content {
  margin-top: 16px;
}

:deep(.bk-collapse-content) {
  padding: 0;
}

</style>
