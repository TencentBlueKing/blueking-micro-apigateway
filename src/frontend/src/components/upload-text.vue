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
  <bk-upload
    :custom-request="handleUpload"
    :limit="1"
    :multiple="false"
    type="binary"
    v-bind="$attrs"
    @error="handleError"
  />
</template>
<script lang="ts" setup>
import { UploadRequestOptions } from 'bkui-vue/lib/upload/upload.type';
import { Message } from 'bkui-vue';
import { onMounted } from 'vue';

const emit = defineEmits<{
  'done': [content: string]
  'error': [void]
}>();

const reader = new FileReader();

const handleUpload = (options: UploadRequestOptions) => {
  try {
    reader.readAsText(options.file);
  } catch (e) {
    const error = e as Error;
    showErrorMsg(error.message);
  }
};

const readerLoadedHandler = (event: ProgressEvent<FileReader>) => {
  const { result } = event.target;
  emit('done', result as string);
};

const handleError = (file: any, fileList: any, error: Error) => {
  showErrorMsg(error?.message);
  emit('error');
};

const showErrorMsg = (msg?: string) => {
  Message({
    theme: 'danger',
    message: `上传失败: ${msg || 'Unknown'}`,
  });
};

onMounted(() => {
  reader.addEventListener('load', readerLoadedHandler);
});

</script>
<style lang="scss" scoped>

</style>
