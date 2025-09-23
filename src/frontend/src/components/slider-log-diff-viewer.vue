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
    @closed="handleClosed"
  >
    <template #header>
      <slot>
        <div>
          <span>{{ titleConfig.title }}</span>
          <span
            v-if="log?.resource_type"
            class="slider-sub-title"
          >{{ common.enums?.resource_type[log?.resource_type] }}</span>
          <tag-operation-type v-if="log?.operation_type" :type="log.operation_type" />
        </div>
      </slot>
    </template>
    <template #default>
      <div v-if="log" class="content-wrapper">
        <div class="diff-titles">
          <div><span class="diff-title before">{{ titleConfig.before }}</span></div>
          <div><span class="diff-title after">{{ titleConfig.after }}</span></div>
        </div>
        <div v-for="(resource_id, index) in log.resource_ids" :key="index" class="diff-wrapper">
          <form-collapse>
            <template #header>{{ resource_id }}</template>
            <bk-code-diff
              :diff-context="20"
              :hljs="highlightjs"
              :new-content="formatContent((log.data_after || [])[index])"
              :old-content="formatContent((log.data_before || [])[index])"
              diff-format="side-by-side"
              language="json"
            />
          </form-collapse>
        </div>
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
import useJsonTransformer from '@/hooks/use-json-transformer';
// @ts-ignore
import TagOperationType from '@/components/tag-operation-type.vue';
import highlightjs from 'highlight.js';
import FormCollapse from '@/components/form-collapse.vue';
import { useCommon } from '@/store';
import { ref, watch } from 'vue';
import { uniqueId } from 'lodash-es';

interface ILog {
  resource_ids: string[]
  data_before: Record<string, any>[]
  data_after: Record<string, any>[]
  resource_type: string
  operation_type: string
}

interface IProps {
  showFooter?: boolean
  titleConfig?: Record<string, any>
  log: ILog | null
}

const isShow = defineModel<boolean>({
  required: true,
  default: false,
});

const {
  log,
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
  'closed': [void]
}>();

const common = useCommon();

const { t } = i18n.global;
const { formatJSON } = useJsonTransformer();
const diffCompKey = ref(uniqueId());

watch(isShow, () => {
  if (isShow.value) {
    diffCompKey.value = uniqueId();
  }
});

const formatContent = (config: Record<string, any>) => {
  return formatJSON({ source: config || {} });
};

const handleConfirmClick = () => {
  emit('confirm');
};

const handleCancelClick = () => {
  isShow.value = false;
  emit('cancel');
};

const handleClosed = () => {
  emit('closed');
};

</script>

<style lang="scss" scoped>

.slider-sub-title {
  font-size: 12px;
  line-height: 12px;
  margin-left: 8px;
  padding-left: 12px;
  color: #979ba5;
  border-left: 1px solid #979ba5;
}

.content-wrapper {
  padding: 24px 24px 0;

  .diff-titles {
    font-size: 14px;
    display: flex;
    margin-bottom: 6px;
    gap: 414px;

    .diff-title {
      font-weight: bold;
    }
  }
}

.footer-actions {
  display: flex;
  gap: 12px;
}

:deep(.collapse-panel-content) {
  margin-top: 0;
}

</style>
