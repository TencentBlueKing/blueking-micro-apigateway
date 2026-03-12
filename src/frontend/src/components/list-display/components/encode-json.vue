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
  <div
    v-if="!!config"
    class="encode-json"
  >
    <pre v-dompurify-html="highlightJson(JSON.stringify(config, null, 4))" />
    <Copy
      class="default-c pointer"
      v-if="isCopy"
      @click="() => handleCopy(JSON.stringify(config, null, 4))"
    />
  </div>
  <div v-else>--</div>
</template>

<script lang="ts" setup>
import highlightJs from 'highlight.js';
import 'highlight.js/styles/github.css';
import { Copy } from 'bkui-vue/lib/icon';
import { handleCopy } from '@/common/util';

interface IProps {
  config: object | null
  isCopy?: boolean
}

const {
  config,
  isCopy = false,
} = defineProps<IProps>();

const highlightJson = (value: string) => {
  if (!value) {
    return '';
  }

  return highlightJs.highlight(value, { language: 'json' }).value;
};
</script>

<style lang="scss" scoped>
.encode-json {
  height: 100%;
  background: #FAFBFD;
  padding: 8px 0px;
  position: relative;
  pre {
    white-space: pre-wrap;
  }
  .default-c {
    position: absolute;
    right: 4px;
    top: 6px;
  }
}
</style>
