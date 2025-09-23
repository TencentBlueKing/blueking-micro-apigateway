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
  <bk-dropdown :popover-options="{ trigger: 'click' }" v-bind="$attrs">
    <Icon
      v-bk-tooltips="{ content: getDisabledToolTip(), disabled: !isReadonlyGateway }"
      class="dropdown-trigger"
      name="more-fill"
      size="16"
      @click="handleTriggerClicked"
    />
    <template #content>
      <bk-dropdown-menu>
        <bk-dropdown-item v-if="clone">
          <bk-button
            v-bk-tooltips="{ content: t('当前网关为只读'), disabled: !isReadonlyGateway }"
            class="dropdown-item-btn"
            text
            @click="handleCloneClick"
            :disabled="isReadonlyGateway"
          >
            {{ t('克隆') }}
          </bk-button>
        </bk-dropdown-item>
        <bk-dropdown-item>
          <bk-pop-confirm
            :title="t('确认撤销？')"
            trigger="click"
            width="288"
            @confirm="handleRevert"
          >
            <bk-button
              v-bk-tooltips="{ content: getDisabledToolTip(row.status, 'revert'), disabled: !isReadonlyGateway }"
              :disabled="isReadonlyGateway || !['update_draft', 'delete_draft'].includes(row.status)"
              class="dropdown-item-btn"
              text
            >
              {{ t('撤销') }}
            </bk-button>
          </bk-pop-confirm>
        </bk-dropdown-item>
        <bk-dropdown-item>
          <bk-pop-confirm
            :content="!['success'].includes(row.status) ? t('删除操作无法撤回，请谨慎操作！') : undefined"
            :title="t('确认删除？')"
            trigger="click"
            width="288"
            @confirm="handleDelete"
          >
            <bk-button
              v-bk-tooltips="{ content: getDisabledToolTip(row.status, 'delete'), disabled: !isReadonlyGateway }"
              :disabled="isReadonlyGateway || !['create_draft', 'success'].includes(row.status)"
              class="dropdown-item-btn"
              text
              theme="danger"
            >
              {{ t('删除') }}
            </bk-button>
          </bk-pop-confirm>
        </bk-dropdown-item>
      </bk-dropdown-menu>
    </template>
  </bk-dropdown>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useCommon } from '@/store';
import Icon from '@/components/icon.vue';
import { computed } from 'vue';

interface IProps {
  row: any
  clone: boolean
}

const isShow = defineModel<boolean>();

const { row, clone } = defineProps<IProps>();

const emit = defineEmits<{
  'clone': [void]
  'revert': [void]
  'delete': [void]
}>();

const { t } = useI18n();
const common = useCommon();

const isReadonlyGateway = computed(() => common.curGatewayData?.read_only);

// 处理不同disabled的tooltip
const getDisabledToolTip = (status?: string, type?: string) => {
  if (isReadonlyGateway.value) {
    return t('当前网关为只读');
  }

  if (!['update_draft', 'delete_draft'].includes(status) && ['revert'].includes(type)) {
    return t('该资源为更新待发布或删除待发布时才可撤销');
  }

  if (!['create_draft', 'success'].includes(status) && ['delete'].includes(type)) {
    return t('该资源为新增待发布或已发布时才可删除');
  }

  return '';
};

const handleCloneClick = () => {
  isShow.value = false;
  emit('clone');
};

const handleRevert = () => {
  isShow.value = false;
  emit('revert');
};

const handleDelete = () => {
  isShow.value = false;
  emit('delete');
};

const handleTriggerClicked = () => {
  if (common.curGatewayData?.read_only) {
    return;
  }
  isShow.value = true;
};

</script>

<style lang="scss" scoped>

.dropdown-trigger {
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  cursor: pointer;
  border-radius: 2px;

  &:hover {
    background: #f0f1f5;
  }
}

.bk-dropdown-popover .bk-dropdown-item {
  padding-inline: 0;
}

.dropdown-item-btn {
  display: block;
  height: 100%;
  padding-inline: 16px;
}

</style>
