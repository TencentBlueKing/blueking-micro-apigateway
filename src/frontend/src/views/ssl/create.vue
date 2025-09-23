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
    <main class="page-content-wrapper">
      <div>
        <form-card>
          <template #title>{{ t('证书') }}</template>
          <div>
            <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
              <!-- name -->
              <bk-form-item :label="t('名称')" class="form-item" property="name" required>
                <bk-input v-model="formModel.name" clearable />
              </bk-form-item>

              <bk-form-item :label="t('方式')" class="form-item">
                <bk-radio-group v-model="type" type="card" @change="handleTypeChange">
                  <bk-radio-button label="input">{{ t('输入') }}</bk-radio-button>
                  <bk-radio-button label="upload">{{ t('上传') }}</bk-radio-button>
                </bk-radio-group>
              </bk-form-item>

              <bk-form-item v-show="type === 'upload'" :label="t('上传证书')" class="form-item" property="cert">
                <upload-text @done="(content: string) => handleUpload(content, 'cert')" />
              </bk-form-item>

              <bk-form-item :label="t('证书')" class="form-item" property="cert" required>
                <bk-input
                  v-model="formModel.cert"
                  :autosize="autoSizeConf"
                  :clearable="false"
                  :placeholder="t('请输入证书内容')"
                  :rows="4"
                  type="textarea"
                />
              </bk-form-item>

              <bk-form-item v-show="type === 'upload'" :label="t('上传私钥')" class="form-item" property="key">
                <upload-text @done="(content: string) => handleUpload(content, 'key')" />
              </bk-form-item>

              <bk-form-item :label="t('私钥')" class="form-item" property="key" required>
                <bk-input
                  v-model="formModel.key"
                  :clearable="false"
                  :placeholder="t('请输入私钥内容')"
                  :rows="4"
                  type="textarea"
                />
              </bk-form-item>

              <bk-form-item class="form-item" label="SNIS" property="snis">
                <bk-input
                  v-model="formModel.snis"
                  :placeholder="t('请输入 SNIS')"
                  clearable
                />
              </bk-form-item>

              <bk-form-item :label="t('过期时间')" class="form-item">
                <span>{{ checkStatus.expire_at || '--' }}</span>
              </bk-form-item>
            </bk-form>
          </div>
        </form-card>
      </div>
    </main>
    <form-page-footer>
      <bk-button
        v-if="!checkStatus.checked || !checkStatus.valid"
        theme="primary"
        @click="checkNameAndSSL"
      >
        {{ t('校验') }}
      </bk-button>
      <bk-button v-else theme="primary" @click="handleSubmit">{{ t('提交') }}</bk-button>
      <bk-button @click="handleCancelClick">{{ t('取消') }}</bk-button>
    </form-page-footer>
  </div>
</template>

<script lang="ts" setup>
import FormPageFooter from '@/components/form/form-page-footer.vue';
import { Form, InfoBox, Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, nextTick, ref, useTemplateRef, watch } from 'vue';
import { checkSSL, getSSL, postSSL, putSSL } from '@/http/ssl';
import FormCard from '@/components/form-card.vue';
import dayjs from 'dayjs';
import UploadText from '@/components/upload-text.vue';

export type UploadFieldType = 'cert' | 'key';

interface IFormModel {
  name: string;
  cert: string;
  key: string;
  snis: string;
}

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const autoSizeConf = {
  minRows: 4,
  maxRows: 8,
};

const type = ref<'input' | 'upload'>('input');

const formModel = ref<IFormModel>({
  name: '',
  cert: '',
  key: '',
  snis: '',
});

const fileList = ref([]);

const checkStatus = ref({
  checked: false,
  valid: false,
  expire_at: '',
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');

const rules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'blur' },
  ],
  cert: [
    { required: true, message: t('证书不能为空'), trigger: 'blur' },
    { validator: (value: string) => value.length >= 128, message: t('证书内容至少需要128个字符'), trigger: 'change' },
  ],
  key: [
    { required: true, message: t('私钥不能为空'), trigger: 'blur' },
    { validator: (value: string) => value.length >= 128, message: t('私钥内容至少需要128个字符'), trigger: 'change' },
  ],
};

const sslId = computed(() => {
  return route.params.id as string;
});

const isEditMode = computed(() => {
  return !!route.params.id;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getSSL({ id } as { id: string });
    formModel.value.name = response.name;
    formModel.value.cert = response.config.cert;
    formModel.value.key = response.config.key;
    formModel.value.snis = response.config.snis[0] || '';
    fileList.value = [];
    await nextTick(() => {
      checkStatus.value = {
        checked: true,
        valid: true,
        expire_at: dayjs.unix(response.config.validity_end)
          .format('YYYY-MM-DD HH:mm:ss'),
      };
    });
  }
}, { immediate: true });

watch([formModel, type], () => {
  checkStatus.value = {
    checked: false,
    valid: false,
    expire_at: '',
  };
}, { deep: true });

const handleTypeChange = () => {
  nextTick(() => {
    formRef.value?.clearValidate();
  });
};

const handleUpload = (content: string, field: UploadFieldType) => {
  if (field === 'cert') {
    formModel.value.cert = content;
  } else if (field === 'key') {
    formModel.value.key = content;
  }
};

const checkNameAndSSL = async () => {
  try {
    await formRef.value!.validate();
  } catch (e) {
    const error = e as Error;
    Message({
      theme: 'error',
      message: error.message || t('校验失败'),
    });
    return;
  }

  try {
    const { cert, key, name } = formModel.value;
    const checkResponse = await checkSSL({
      data: {
        cert,
        key,
        name,
      },
    });
    checkStatus.value = {
      checked: true,
      valid: true,
      expire_at: dayjs.unix(checkResponse.validity_end)
        .format('YYYY-MM-DD HH:mm:ss'),
    };
    Message({
      theme: 'success',
      message: t('证书校验成功'),
    });
  } catch {
    checkStatus.value = {
      checked: true,
      valid: false,
      expire_at: '',
    };
  }
};

const handleSubmit = async () => {
  try {
    // 校验表单
    await formRef.value!.validate();
    const checkResponse = await checkSSL({
      data: {
        ...formModel.value,
      },
    });

    const config = {
      cert: checkResponse.cert,
      key: checkResponse.key,
      snis: formModel.value.snis ? [formModel.value.snis] : checkResponse.snis,
    };

    const data = {
      config,
      name: checkResponse.name,
    };

    InfoBox({
      title: t('确认提交？'),
      confirmText: t('提交'),
      cancelText: t('取消'),
      onConfirm: async () => {
        if (isEditMode.value) {
          await putSSL({
            data: {
              ...data,
              id: sslId.value,
            },
            id: sslId.value,
          });
        } else {
          await postSSL({ data });
        }

        Message({
          theme: 'success',
          message: t('提交成功'),
        });

        await router.push({ name: 'ssl', replace: true });
      },
    });
  } catch (e) {
    const error = e as Error;
    Message({
      theme: 'error',
      message: error.message || t('提交失败'),
    });
  }
};

const handleCancelClick = () => {
  router.back();
};

</script>

<style lang="scss" scoped>

.page-content-wrapper {
  min-height: calc(100vh - 158px);
  padding: 24px;
}

</style>
