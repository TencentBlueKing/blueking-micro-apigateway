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
  <div class="form-card-wrapper">
    <div class="form-card-title">
      <div class="main-title">
        <slot name="title"></slot>
      </div>
      <div class="sub-title">
        <slot name="subTitle"></slot>
      </div>
    </div>
    <div class="form-card-content">
      <slot></slot>
    </div>
  </div>
</template>

<style lang="scss" scoped>

.form-card-wrapper {
  z-index: 10;
  margin-bottom: 16px;
  padding: 16px 24px;
  border-radius: 2px;
  background: #ffffff;
  box-shadow: 0 2px 4px 0 #1919290d;

  &:last-of-type {
    margin-bottom: 0;
  }

  .form-card-title {
    font-size: 14px;
    font-weight: 700;
    line-height: 22px;
    display: flex;
    align-items: center;
    margin-bottom: 18px;
    color: #313238;

    .main-title {
      margin-right: 8px;
    }

    .sub-title {
      font-size: 12px;
      font-weight: normal;
      line-height: 20px;
      color: #979ba5;
    }
  }
}

</style>
