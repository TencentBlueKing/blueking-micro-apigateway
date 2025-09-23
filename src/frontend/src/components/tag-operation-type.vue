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
  <bk-tag v-if="type" :theme="themeMap[type]" type="stroke">{{ common.enums?.operation_type[type] || '--' }}</bk-tag>
  <span v-else>--</span>
</template>

<script lang="ts" setup>
import { useCommon } from '@/store';

interface IProps {
  type?: string
}

const { type } = defineProps<IProps>();

const themeMap: Record<string, string> = {
  create: 'success',
  publish: 'success',
  delete: 'danger',
  revert: 'danger',
  update: 'warning',
  fix_conflict: 'warning',
};

const common = useCommon();

</script>
