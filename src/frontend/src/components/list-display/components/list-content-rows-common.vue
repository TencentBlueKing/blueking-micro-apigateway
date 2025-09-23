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
  <div>
    <bk-row v-if="resource.created_at">
      <bk-col :span="4">
        <div class="content-label">{{ t('创建时间') }}:</div>
      </bk-col>
      <bk-col :span="10">
        <div class="content-value">{{ formatTime(resource.created_at) }}</div>
      </bk-col>
    </bk-row>
    <bk-row v-if="resource.create_by">
      <bk-col :span="4">
        <div class="content-label">{{ t('创建人') }}:</div>
      </bk-col>
      <bk-col :span="10">
        <div class="content-value">{{ resource.create_by }}</div>
      </bk-col>
    </bk-row>
    <bk-row v-if="resource.updated_at">
      <bk-col :span="4">
        <div class="content-label">{{ t('更新时间') }}:</div>
      </bk-col>
      <bk-col :span="10">
        <div class="content-value">{{ formatTime(resource.updated_at) }}</div>
      </bk-col>
    </bk-row>
    <bk-row v-if="resource.update_by">
      <bk-col :span="4">
        <div class="content-label">{{ t('更新人') }}:</div>
      </bk-col>
      <bk-col :span="10">
        <div class="content-value">{{ resource.update_by }}</div>
      </bk-col>
    </bk-row>
    <bk-row v-if="resource.status">
      <bk-col :span="4">
        <div class="content-label">{{ t('状态') }}:</div>
      </bk-col>
      <bk-col :span="10">
        <div class="content-value">
          <tag-status :status="resource.status" />
        </div>
      </bk-col>
    </bk-row>
  </div>
</template>

<script lang="ts" setup>
import { ResourceType } from '@/types';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';
import TagStatus from '@/components/tag-status.vue';

interface IProps {
  resource?: ResourceType
}

const { resource } = defineProps<IProps>();

const { t } = useI18n();

const formatTime = (timeStamp: number) => {
  return dayjs.unix(timeStamp)
    .format('YYYY-MM-DD HH:mm:ss Z');
};

</script>


<style lang="scss" scoped>
.content-label {
  padding-right: 0;
  text-align: right;
}

.content-value {
  color: #313238;
}

.bk-grid-row {
  margin-bottom: 12px;
}
</style>
