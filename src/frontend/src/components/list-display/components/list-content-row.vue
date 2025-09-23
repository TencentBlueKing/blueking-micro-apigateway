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
  <bk-row>
    <bk-col :span="4">
      <div class="content-label">{{ label }}:</div>
    </bk-col>
    <bk-col :span="10">
      <div class="content-value">
        <slot></slot>
      </div>
    </bk-col>
  </bk-row>
</template>

<script lang="ts" setup>

interface IProps {
  label: string
}

const { label } = defineProps<IProps>();

</script>


<style lang="scss" scoped>
.content-label {
  padding-right: 0;
  text-align: right;
}

.content-value {
  word-break: break-all;
  color: #313238;
}

.bk-grid-row {
  margin-bottom: 12px;
}
</style>
