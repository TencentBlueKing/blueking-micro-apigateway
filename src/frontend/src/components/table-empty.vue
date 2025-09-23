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
  <div class="gateways-empty-table-search">
    <bk-exception :type="type" class="exception-wrap-item exception-part" scene="part">
      <div class="exception-part-title">
        <slot name="title">{{ title }}</slot>
      </div>
      <template v-if="type === 'search-empty'">
        <div class="search-empty-tips">
          {{ t('可以尝试 调整关键词 或') }}
          <span class="clear-search" @click="handlerClearFilter">
            {{ t('清空搜索条件') }}
          </span>
        </div>
      </template>
    </bk-exception>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

interface IProps {
  type: 'empty' | 'search-empty';
}

const {
  type = 'empty',
} = defineProps<IProps>();

const emit = defineEmits(['clear-filter']);

const { t } = useI18n();

const title = computed(() => {
  if (type === 'search-empty') {
    return t('搜索结果为空');
  }
  return t('暂无数据');
});


const handlerClearFilter = () => {
  emit('clear-filter');
};

</script>

<style lang="scss" scoped>
.gateways-empty-table-search {
  display: flex;
  align-items: center;
  width: auto !important;
  max-height: 280px;
  margin: 0 auto;
  padding-bottom: 24px;

  .search-empty-tips {
    font-size: 12px;
    margin-top: 8px;
    color: #979ba5;

    .clear-search {
      cursor: pointer;
      color: #3a84ff;
    }
  }

  .empty-tips {
    color: #63656e;
  }

  .exception-part-title {
    font-size: 14px;
    margin-bottom: 5px;
    color: #63656e;
  }

  .refresh-tips {
    cursor: pointer;
    color: #3a84ff;
  }

  .exception-wrap-item .bk-exception-img.part-img {
    height: 130px;
  }

  .bk-table-empty-text {
    padding: 0 !important;
  }

  :deep(.bk-exception-footer) {
    margin-top: 0
  }
}
</style>
