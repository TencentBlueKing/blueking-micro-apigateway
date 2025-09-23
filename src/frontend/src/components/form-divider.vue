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
  <bk-divider ref="divider" :align="align" :class="{ 'sub-level': level !== 1 }" v-bind="$attrs">
    <slot></slot>
  </bk-divider>
</template>

<script lang="ts" setup>
import { computed, ref, useTemplateRef, watch } from 'vue';
import { useElementSize } from '@vueuse/core';

interface IProps {
  // 文本距离容器最左侧的偏移量
  offsetLeft?: string
  marginTop?: string
  marginBottom?: string
  fontSize?: string
  level?: number
}

const { offsetLeft = '16', fontSize = '15', marginTop = '48', marginBottom = '48', level = 1 } = defineProps<IProps>();

const dividerRef = useTemplateRef('divider');
const localWidth = ref('100%');

const align = computed(() => {
  return level === 1 ? 'left' : 'center';
});

const localOffsetLeft = computed(() => {
  return `${offsetLeft}px`;
});

const localMarginTop = computed(() => {
  return `${marginTop}px`;
});

const localMarginBottom = computed(() => {
  return `${marginBottom}px`;
});

const subLevelOffsetLeft = computed(() => {
  return `${(level - 1) * 8}%`;
});

const localFontSize = computed(() => {
  return level === 1 ? `${fontSize}px` : `${Number(fontSize) - 1}px`;
});

watch(dividerRef, () => {
  if (dividerRef.value) {
    const { width } = useElementSize(dividerRef.value as HTMLElement);
    localWidth.value = `${width.value - level * Number(offsetLeft)}px`;
  }
});

</script>

<style lang="scss" scoped>

// 覆盖 bkui 组件样式
.bk-divider-horizontal {
  font-weight: 700;
  margin-top: v-bind(localMarginTop);
  margin-bottom: v-bind(localMarginBottom);

  :deep(.bk-divider-info.bk-divider-info-left) {
    font-size: v-bind(localFontSize);
    padding-left: v-bind(localOffsetLeft);
  }

  :deep(.bk-divider-info.bk-divider-info-center) {
    font-size: v-bind(localFontSize);
    //font-weight: normal;
    left: v-bind(subLevelOffsetLeft);
  }
}

.bk-divider-horizontal.sub-level {
  width: v-bind(localWidth);
  margin-inline: auto;
}


</style>
