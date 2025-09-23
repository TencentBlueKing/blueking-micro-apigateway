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
  <div v-if="labelStrings.length" class="tag-label-wrapper">
    <BkTag
      v-if="labelStrings.length === 1"
      theme="info"
      @click="handleLabelClick(labelStrings[0])"
    >
      {{ labelStrings[0] }}
    </BkTag>
    <template v-else>
      <BkTag theme="info" @click="handleLabelClick(labelStrings[0])">{{ labelStrings[0] }}</BkTag>
      <BkPopover theme="light" :popover-delay="0">
        <BkTag
          ref="moreTagRef"
          theme="info"
          class="more-label-tag ml4"
          @click="handleLabelClick(labelStrings[0])"
        >
          {{ `+${labelStrings.length - 1}` }}
        </BkTag>
        <template #content>
          <div v-for="(label, index) in labelStrings" :key="index">
            <BkTag v-if="index > 0" class="mr4 mb4" theme="info" @click="handleLabelClick(label)">{{ label }}</BkTag>
          </div>
        </template>
      </BkPopover>
    </template>
  </div>
  <span v-else>--</span>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { isEmpty } from 'lodash-es';
import { useI18n } from 'vue-i18n';
import { Message, Tag } from 'bkui-vue';
import { useClipboard } from '@vueuse/core';

interface IProps {
  labels?: { [key: string]: string }
  needCopy?: boolean
}

const { labels = {}, needCopy = false } = defineProps<IProps>();

const { t } = useI18n();
const { copy } = useClipboard({ legacy: true });

const labelStrings = computed(() => {
  return isEmpty(labels) ? [] : Object.entries(labels)
    .map(([key, value]) => `${key}:${value}`);
});

const moreTagRef = ref<InstanceType<typeof Tag>>(null);

const handleLabelClick = async (label: string) => {
  if (needCopy) {
    try {
      await copy(label);
      Message({
        theme: 'success',
        message: t('已复制'),
      });
    } catch {
      Message({
        theme: 'error',
        message: t('复制失败'),
      });
    }
  }
};

const getResizeLabelWidth = () => {
  if (labelStrings.value?.length) {
    const labelWidthList = [];
    // 单元格默认左右内边距各12px
    const cellPadding = 24;
    // 更多数据tag的左边距
    const numTagMargin = 4;
    const tagLabelWrapper = document.querySelectorAll('.tag-label-wrapper');
    if (tagLabelWrapper?.length) {
      tagLabelWrapper.forEach((item) => {
        const tagScrollWidth = item?.scrollWidth;
        const tagClientWidth = item?.clientWidth ?? 0;
        const moreTagWidth = item?.querySelector('.more-label-tag')?.offsetWidth ?? 0;
        if (tagScrollWidth >= tagClientWidth) {
          labelWidthList.push(item?.scrollWidth + cellPadding + moreTagWidth + numTagMargin);
        }
      });
    }
    // 取当前列表最大宽度
    const maxValue = labelWidthList.length > 0 ? Math.max(...labelWidthList) : 80;
    return maxValue;
  }
};

defineExpose({
  getResizeLabelWidth,
});
</script>

<style lang="scss" scoped>
.tag-label-wrapper {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
}
</style>
