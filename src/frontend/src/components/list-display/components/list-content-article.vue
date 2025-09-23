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
  <article class="content-group">
    <header class="content-group-title">
      <slot name="title"></slot>
    </header>
    <bk-container :col="14" :margin="6">
      <slot></slot>
    </bk-container>
  </article>
</template>

<style lang="scss" scoped>

.content-group {
  margin-bottom: 24px;
}

.content-group-title {
  font-size: 14px;
  font-weight: 700;
  line-height: 24px;
  height: 24px;
  margin-bottom: 12px;
  padding-left: 8px;
  color: #313238;
  border-radius: 2px;
  background: #f5f7fa;
}

</style>
