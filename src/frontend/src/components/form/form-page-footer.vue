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
  <footer class="action-footer">
    <div class="btn-wrapper">
      <slot>
        <bk-button theme="primary" @click="handleSubmitClick">{{ t('提交') }}</bk-button>
        <bk-button @click="handleCancelClick">{{ t('取消') }}</bk-button>
      </slot>
    </div>
  </footer>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

const emits = defineEmits(['submit', 'cancel']);

const { t } = useI18n();

const handleSubmitClick = () => {
  emits('submit');
};

const handleCancelClick = () => {
  emits('cancel');
};

</script>

<style lang="scss" scoped>
.action-footer {
  line-height: 52px;
  position: sticky;
  z-index: 2;
  bottom: 0;
  display: flex;
  align-items: center;
  height: 52px;
  border-top: 1px solid #dcdee5;
  background-color: #ffffff;

  .btn-wrapper {
    display: flex;
    align-items: center;
    padding-left: 24px;
    gap: 12px;
  }
}
</style>
