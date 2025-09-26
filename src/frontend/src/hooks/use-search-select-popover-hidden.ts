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

import { nextTick } from 'vue';

type IElement =  HTMLElement | null;

/**
 * 兼容处理searchSelect组件点击任意处一直弹窗
 */
export function useSearchSelectPopoverHidden() {
  function handleSearchOutside()  {
    const searchSelectEl = document.querySelector('.bk-search-select-popover');
    if (searchSelectEl) {
      nextTick(() => {
        // 有menuList的多选列表
        const parentEl = searchSelectEl?.parentNode as IElement;
        if (parentEl) {
          parentEl.style.display = 'none';
        }
      });
    } else {
      // 单个输入
      setTimeout(() => {
        const popoverVisible = Array.from(document.querySelectorAll('.bk-popover.bk-pop2-content.visible'))
          .find((elem) => {
            const classes = elem.className.trim().split(/\s+/);
            return classes.length === 3 && classes.every(cl => ['bk-popover', 'bk-pop2-content', 'visible'].includes(cl));
          }) as IElement;
        if (popoverVisible?.classList?.value === 'bk-popover bk-pop2-content visible') {
          popoverVisible.classList.remove('visible');
          popoverVisible.classList.add('hidden');
          popoverVisible.style.display = 'none';
          popoverVisible.style.visibility = 'hidden';
        }
      }, 0);
    }
  };

  function handleSearchSelectClick()  {
    nextTick(() => {
      const searchSelectFocus = document.querySelector('.table-resource-search > .bk-search-select-container.is-focus');
      const popover = document.querySelector('.bk-popover.bk-pop2-content');
      if (searchSelectFocus) {
        const searchSelectEl = document.querySelector('.bk-search-select-popover');
        const parentEl = searchSelectEl?.parentNode as IElement;
        if (parentEl) {
          parentEl.style.display = 'block';
        }
      }
      // 单个输入
      setTimeout(() => {
        if (popover.classList.contains('hidden')) {
          popover.classList.add('visible');
        }
      }, 0);
    });
  };

  return {
    handleSearchOutside,
    handleSearchSelectClick,
  };
}
