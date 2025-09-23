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
  <bk-popover
    :arrow="false"
    disable-outside-click
    ext-cls="user-home"
    placement="bottom"
    theme="light"
  >
    <div class="user-name">
      {{ userInfo.username || userInfo.uid }}
      <angle-down-line class="pl5" />
    </div>
    <template #content>
      <div
        class="logout"
        @click="logout"
      >{{ t('退出登录') }}
      </div>
    </template>
  </bk-popover>
</template>

<script setup lang="ts">
import { useUser } from '@/store/user';
import { logout } from '@/common/auth';
import { AngleDownLine } from 'bkui-vue/lib/icon';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();
const user = useUser();
const userInfo = ref(user.user);
</script>

<style lang="scss" scoped>
.user-home {
  z-index: 1000;
  font-size: 14px;
}

.user-name {
  padding-left: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.user-name:hover {
  color: #fff;
}

.logout {
  display: inline-block;
  text-align: center;
  width: 80px;
  cursor: pointer;
  color: #63656e;
}
</style>
