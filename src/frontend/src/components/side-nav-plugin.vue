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
  <!--  插件列表的导航目录  -->
  <div class="side-nav-wrapper">
    <aside class="highlight-glyph">
      <div class="highlight-bar"></div>
    </aside>
    <ul ref="itemWrapperRef" class="list-wrapper">
      <li v-for="nav of list" :key="nav.id">
        <span
          :class="{ active: activeId === nav.id }"
          class="nav"
          @click="handleNavItemClick(nav)"
        >
          {{ nav.name }}
        </span>
      </li>
    </ul>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useElementBounding, useElementSize, useScroll } from '@vueuse/core';
import { minBy } from 'lodash-es';

interface INavItem {
  id: string;
  name: string;
}

interface IProps {
  list?: INavItem[]
  width?: string | number
  container: HTMLDivElement
  parent: HTMLDivElement
  elements: HTMLDivElement[] | HTMLElement[] | null
}

const activeId = defineModel<string>({
  default: '',
});

const scrollY = defineModel<number>('scrollY');

const {
  list = [],
  width = 160,
  container,
  parent,
  elements = [],
} = defineProps<IProps>();

const itemWrapperRef = ref<HTMLDivElement>();

const skipScrollStopCallback = ref(false);

const { y } = useScroll(() => container, {
  // 监听容器的滚动结束事件，获取距离容器最上方且可见的标题元素
  onStop: () => {
    if (skipScrollStopCallback.value) {
      scrollY.value = y.value;
      skipScrollStopCallback.value = false;
      return;
    }

    const topVisibleGroup = minBy(elements, (el) => {
      const { top, bottom } = useElementBounding(el);
      const { top: parentTop, height: parentHeight } = useElementBounding(parent);
      const offsetTop = top.value - parentTop.value;
      const bottomFromParentTop = bottom.value - parentTop.value;

      if (offsetTop < 0) {
        if (bottomFromParentTop < 0 || bottomFromParentTop <= (parentHeight.value / 2)) {
          return Infinity;
        }
        return bottomFromParentTop;
      }
      return offsetTop;
    });
    activeId.value = topVisibleGroup?.id || '';
    scrollY.value = y.value;
  },
  idle: 50,
});

const localWidth = computed(() => {
  return `${width}px`;
});

const itemHeight = computed(() => {
  const { height: parentHeight } = useElementSize(itemWrapperRef);
  return Math.floor(parentHeight.value / (list.length || 1));
});

const itemHeightPx = computed(() => {
  return `${itemHeight.value}px`;
});

const highlightBarTop = computed(() => {
  const index = list.findIndex(item => item.id === activeId.value);
  return `${index * itemHeight.value}px`;
});

const handleNavItemClick = (nav: INavItem) => {
  if (nav.id === activeId.value) {
    return;
  }

  const element = document.getElementById(nav.id);

  if (element) {
    activeId.value = nav.id;
    skipScrollStopCallback.value = true;
    element.scrollIntoView({
      behavior: 'smooth', // 平滑滚动
      block: 'start', // 元素顶部与视口顶部对齐
    });
  }
};

// 目录变更时，默认高亮第一个目录
watch(() => list, () => {
  activeId.value = list[0]?.id || '';
}, { deep: true });

watch(scrollY, () => {
  if (scrollY.value === y.value) {
    return;
  }

  y.value = scrollY.value;
});

</script>

<style lang="scss" scoped>

.side-nav-wrapper {
  display: flex;
}

.highlight-glyph {
  position: relative;
  width: 1px;
  background-color: #dcdee5;

  .highlight-bar {
    position: absolute;
    top: v-bind(highlightBarTop);
    width: 100%;
    height: v-bind(itemHeightPx);
    transition: top 0.2s ease-in-out;
    background-color: #3a84ff;
  }
}

.list-wrapper {
  font-size: 12px;
  line-height: 28px;
  width: v-bind(localWidth);
  text-align: left;
  color: #979ba5;

  .nav {
    display: block;
    padding-left: 16px;
    cursor: pointer;
    text-decoration: none;
    color: #63656e;

    &.active {
      color: #3a84ff;
    }
  }
}

</style>
