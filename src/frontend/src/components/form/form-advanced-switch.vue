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
  <!-- 表单 高级配置 显示开关-->
  <bk-form-item class="form-item" label="">
    <div class="advanced-setting-switch" @click="visible = !visible">
      <span>{{ ('高级配置') }}</span>
      <Icon
        :class="{ 'is-on': visible }" class="switch-icon" color="#3a84ff" name="arrows-down" size="22"
      />
    </div>
  </bk-form-item>
</template>

<script lang="ts" setup>
import Icon from '@/components/icon.vue';

const visible = defineModel<boolean>({
  default: false,
});

</script>

<style lang="scss" scoped>

.advanced-setting-switch {
  font-size: 14px;
  line-height: 22px;
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #3a84ff;

  .switch-icon.is-on {
    transform: rotate(180deg);
  }
}

</style>
