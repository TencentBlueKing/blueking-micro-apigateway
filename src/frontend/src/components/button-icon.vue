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
  <bk-button
    v-bind="$attrs"
    @click="$emit('click')"
  >
    <Icon v-if="icon" :color="iconColor" :name="icon" style="margin-right: 6px;" />
    <slot>{{ t('增加') }}</slot>
  </bk-button>
</template>

<script lang="ts" setup>
import Icon from '@/components/icon.vue';
import { useI18n } from 'vue-i18n';

interface IProps {
  icon?: string;
  iconColor?: string;
}

const { icon = 'plus-circle-shape', iconColor = 'inherit' } = defineProps<IProps>();

defineEmits(['click']);

const { t } = useI18n();

</script>

<style lang="scss" scoped>

</style>
