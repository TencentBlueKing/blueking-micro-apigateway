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
  <bk-sideslider
    v-model:is-show="isShow"
    width="960"
  >
    <template #header>
      <slot>
        <div>
          <span>{{ titleConfig.title }}</span>
          <tag-operation-type v-if="operationType" :type="operationType" />
        </div>
      </slot>
    </template>
    <template #default>
      <div v-if="!isContentEqual" class="content-wrapper">
        <div class="diff-titles">
          <div><span class="diff-title before">{{ titleConfig.before }}</span></div>
          <div><span class="diff-title after">{{ titleConfig.after }}</span>
          </div>
        </div>
        <div class="diff-wrapper">
          <bk-code-diff
            :key="dateKey"
            :diff-context="20"
            :hljs="highlightjs"
            :new-content="newContent"
            :old-content="oldContent"
            diff-format="side-by-side"
            language="json"
          />
        </div>
      </div>
      <div v-else class="diff-wrapper">
        <bk-exception
          :description="t('没有差异')"
          class="exception-wrap-item"
          scene="part"
          type="empty"
        />
      </div>
    </template>
    <template v-if="showFooter" #footer>
      <div class="footer-actions">
        <bk-button theme="primary" @click="handleConfirmClick">{{ t('确定') }}</bk-button>
        <bk-button @click="handleCancelClick">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<script lang="ts" setup>
import i18n from '@/i18n';
import { ref, computed } from 'vue';
import highlightjs from 'highlight.js';
import useJsonTransformer from '@/hooks/use-json-transformer';
// @ts-ignore
import TagOperationType from '@/components/tag-operation-type.vue';
import { isEqual } from 'lodash-es';

interface IProps {
  beforeConfig: Record<string, any>
  afterConfig: Record<string, any>
  operationType?: string
  showFooter?: boolean
  titleConfig?: Record<string, any>
}

const isShow = defineModel<boolean>({
  required: true,
  default: false,
});

const {
  beforeConfig,
  afterConfig,
  operationType,
  showFooter = false,
  titleConfig = {
    title: i18n.global.t('查看变更'),
    before: i18n.global.t('变更前'),
    after: i18n.global.t('变更后'),
  },
} = defineProps<IProps>();

const emit = defineEmits<{
  'confirm': [void]
  'cancel': [void]
}>();

const { t } = i18n.global;
const { formatJSON } = useJsonTransformer();

const dateKey = ref<number>(+new Date());

const newContent = computed(() => {
  dateKey.value = +new Date();
  return formatJSON({ source: afterConfig });
});

const oldContent = computed(() => {
  dateKey.value = +new Date();
  return formatJSON({ source: beforeConfig });
});

const isContentEqual = computed(() => {
  return isEqual(oldContent.value, newContent.value);
});

const handleConfirmClick = () => {
  emit('confirm');
};

const handleCancelClick = () => {
  isShow.value = false;
  emit('cancel');
};

</script>

<style lang="scss" scoped>

.content-wrapper {
  padding: 24px 24px 0;

  .diff-titles {
    font-size: 14px;
    position: relative;
    display: flex;
    align-items: center;
    height: 40px;
    margin-bottom: 8px;
    background: #dcdee5;
    gap: 414px;

    &::after {
      position: absolute;
      top: 8px;
      left: 50%;
      margin-left: -1px;
      width: 1px;
      height: 24px;
      content: "";
      background: #FFFFFF;
    }

    .diff-title {
      color: #313238;
      font-weight: bold;
      font-size: 14px;
      margin-left: 12px;
    }
  }
}

.footer-actions {
  display: flex;
  gap: 12px;
}

.exception-wrap-item {
  // margin-top: 30%;
}

.diff-wrapper {
  height: calc(100vh - 172px);
  overflow-y: auto;
  :deep(.d2h-file-wrapper) {
    border-radius: 0px;
  }
}

</style>
