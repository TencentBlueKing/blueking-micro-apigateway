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
  <div v-if="methods.length" class="tag-http-method-wrapper">
    <bk-tag
      v-if="methods.length === 1"
      :theme="METHOD_THEMES[methods[0] as keyof typeof METHOD_THEMES]"
    >
      {{ methods[0] }}
    </bk-tag>
    <template v-else>
      <bk-tag :theme="METHOD_THEMES[methods[0] as keyof typeof METHOD_THEMES]">{{ methods[0] }}</bk-tag>
      <bk-popover :popover-delay="0" theme="light">
        <bk-tag
          ref="moreTagRef"
          class="more-label-tag ml4"
        >
          {{ `+${methods.length - 1}` }}
        </bk-tag>
        <template #content>
          <div v-for="(method, index) in methods" :key="index">
            <bk-tag
              v-if="index > 0"
              :theme="METHOD_THEMES[method as keyof typeof METHOD_THEMES]"
              class="mr4 mb4"
            >
              {{ method }}
            </bk-tag>
          </div>
        </template>
      </bk-popover>
    </template>
  </div>
  <span v-else>--</span>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { Tag } from 'bkui-vue';
import { METHOD_THEMES } from '@/enum';

interface IProps {
  methods?: string[]
}

const { methods = [] } = defineProps<IProps>();

const moreTagRef = ref<InstanceType<typeof Tag>>(null);

const getResizeLabelWidth = () => {
  if (methods.length) {
    const labelWidthList: number[] = [];
    // 单元格默认左右内边距各12px
    const cellPadding = 24;
    // 更多数据tag的左边距
    const numTagMargin = 4;
    const tagLabelWrapper = document.querySelectorAll('.tag-http-method-wrapper');
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
    return labelWidthList.length > 0 ? Math.max(...labelWidthList) : 80;
  }
};

defineExpose({
  getResizeLabelWidth,
});
</script>

<style lang="scss" scoped>
.tag-http-method-wrapper {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
}
</style>
