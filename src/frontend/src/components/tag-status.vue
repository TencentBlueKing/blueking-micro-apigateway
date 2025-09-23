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
  <template v-if="status">
    <div
      class="dot-wrapper"
      v-if="['success', 'create_draft', 'update_draft', 'delete_draft'].includes(status)">
      <span :class="['ag-dot', status]"></span>
      <span>{{ STATUS_CN_MAP[status] || '--' }}</span>
    </div>

    <bk-tag v-else :theme="themeMap[status]" type="stroke">{{ STATUS_CN_MAP[status] || '--' }}</bk-tag>
  </template>
  <span v-else>--</span>
</template>

<script lang="ts" setup>
import { STATUS_CN_MAP } from '@/enum';

interface IProps {
  status?: string
}

const { status } = defineProps<IProps>();

const themeMap: Record<string, string> = {
  delete_draft: 'danger',
  create_draft: 'warning',
  conflict: 'warning',
  update_draft: 'info',
  success: 'success',
};

</script>

<style lang="scss" scoped>
.dot-wrapper {
  display: flex;
  align-items: center;

  .ag-dot {
    flex-shrink: 0;
    width: 8px;
    height: 8px;
    margin-right: 4px;
    border: 1px solid #C4C6CC;
    border-radius: 50%;

    &.success {
      background: #e5f6ea;
      border: 1px solid #3fc06d;
    }

    &.create_draft,
    &.update_draft,
    &.delete_draft {
      background: #ffe8c3;
      border: 1px solid #ff9c01;
    }
  }
}
</style>
