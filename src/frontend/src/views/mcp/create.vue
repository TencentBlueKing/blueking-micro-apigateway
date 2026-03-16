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
    v-model:is-show="dialogData.isShow"
    width="700"
    :before-close="handleBeforeClose"
    :title="dialogData.title"
    theme="primary">
    <bk-form ref="formRef" label-width="120" class="create-gw-form" :model="formData" :rules="rules">

      <bk-form-item
        class="form-item-name"
        :label="t('名称')"
        property="name"
        required
      >
        <bk-input
          v-model="formData.name"
          :maxlength="128"
          :placeholder="t('请输入令牌名称')"
          clearable
          autofocus
        />
      </bk-form-item>
      <span class="common-form-tips form-item-name-tips">
        {{ t('1-128 个字符，同一网关下名称不可重复') }}
      </span>
      <bk-form-item
        :label="t('描述')"
        property="description"
      >
        <bk-input
          type="textarea"
          v-model="formData.description"
          :placeholder="t('请输入令牌描述')"
          :maxlength="500"
          :clearable="false"
        />
      </bk-form-item>
      <bk-form-item
        :label="t('访问范围')"
        property="access_scope"
        required
      >
        <bk-select
          v-model="formData.access_scope"
          :placeholder="t('请选择')"
        >
          <bk-option
            v-for="item in scopeList"
            :key="item.value"
            :id="item.value"
            :name="item.label"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('过期时间')"
        property="expired_at"
        class="form-item-name"
        required
      >
        <bk-date-picker
          v-model="formData.expired_at"
          :disabled-date="disableDate"
          type="datetime"
          style="width: 100%;"
          clearable
        />
      </bk-form-item>
      <span class="common-form-tips form-item-name-tips">
        {{ t('令牌过期后将无法使用，需重新创建') }}
      </span>
    </bk-form>

    <template #footer>
      <div class="footer-actions">
        <bk-button theme="primary" :loading="dialogData.loading" @click="handleConfirmCreate">{{ t('确定') }}</bk-button>
        <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>

  <bk-dialog
    v-model:is-show="newToken.isShow"
    :title="t('令牌创建成功')"
    width="600"
    :quick-close="false"
  >
    <div>
      <bk-alert
        theme="warning"
        :title="t('请立即复制并保存此令牌。关闭此对话框后，您将无法再次查看完整令牌。')"
      />
      <div class="token-wrapper">
        {{ newToken?.data?.token }}
        <Copy
          class="default-c pointer"
          @click="() => handleCopy(newToken?.data?.token)"
        />
      </div>
      <encode-json
        :config="mcpServers"
        :is-copy="true" />
    </div>
    <template #footer>
      <BkButton
        style="width: 142px;"
        theme="primary"
        @click="newToken.isShow = false"
      >
        {{ t('我已保存，关闭') }}
      </BkButton>
    </template>
  </bk-dialog>

</template>

<script lang="ts" setup>
import { computed, nextTick, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Copy } from 'bkui-vue/lib/icon';
import { IDialog, IMcpToken } from '@/types';
import { useCommon } from '@/store';
import { postMcpToken } from '@/http/mcp';
import { useSidebar } from '@/hooks/index';
import { handleCopy } from '@/common/util';
import EncodeJson from '@/components/list-display/components/encode-json.vue';

const emit = defineEmits(['done']);

const { t } = useI18n();
const { initSidebarFormData, isSidebarClosed } = useSidebar();

const common = useCommon();

const { BK_DASHBOARD_URL } = window;

const formRef = ref(null);
const formData = ref({
  name: '',
  description: '',
  access_scope: 'read',
  expired_at: '',
});

const dialogData = ref<IDialog>({
  isShow: false,
  loading: false,
  title: t('新建 MCP Access Token'),
});

const disableDate = (date: any) => date && date.valueOf() < Date.now() - 86400;

const newToken = ref<{
  isShow: boolean;
  data: IMcpToken;
}>({
  isShow: false,
  data: {},
});
const mcpServers = computed(() => ({
  mcpServers: {
    [`bk-apisix-${common.gatewayName}`]: {
      url: `${BK_DASHBOARD_URL.replace('/web', '')}/mcp/gateways/${common.gatewayId}`,
      type: 'streamableHttp',
      headers: {
        Authorization: `Bearer ${newToken.value?.data?.token}`,
      },
    },
  },
}));

const scopeList = computed(() => {
  return [
    {
      label: t('read - 只读（允许 GET 请求和 sync 操作）'),
      value: 'read',
    },
    {
      label: t('readwrite - 读写（允许所有操作，包括 POST/PUT/DELETE）'),
      value: 'readwrite',
    },
  ];
});

const rules = {
  name: [
    {
      required: true,
      message: t('请输入令牌名称'),
      trigger: 'change',
    },
    {
      validator: (value: string) => value.length >= 1,
      message: t('不能小于1个字符'),
      trigger: 'change',
    },
    {
      validator: (value: string) => value.length <= 128,
      message: t('不能多于128个字符'),
      trigger: 'change',
    },
    {
      validator: (value: string) => {
        const reg = /^[a-z][a-z0-9-]*$/;
        return reg.test(value);
      },
      message: t('由小写字母、数字、连接符（-）组成，首字符必须是小写字母，长度大于1小于128个字符'),
      trigger: 'change',
    },
  ],
};

const handleConfirmCreate = async () => {
  try {
    await formRef.value.validate();

    dialogData.value.loading = true;
    const payload = {
      ...formData.value,
      expired_at: Math.floor(new Date(formData.value.expired_at).getTime() / 1000), // 转换为秒时间戳
    };
    newToken.value.data = await postMcpToken({
      gatewayId: common.gatewayId,
      data: payload,
    });
    dialogData.value.isShow = false;
    newToken.value.isShow = true;
    reset();
    emit('done');
  } catch (error) {
  } finally {
    dialogData.value.loading = false;
  }
};

const handleClose = () => {
  dialogData.value.isShow = false;
  reset();
};

const reset = () => {
  formData.value = {
    name: '',
    description: '',
    access_scope: 'read',
    expired_at: '',
  };
  nextTick(() => {
    formRef.value?.clearValidate();
  });
};

const handleBeforeClose = () => {
  return isSidebarClosed(JSON.stringify(formData.value));
};

const show = () => {
  initSidebarFormData(formData.value);
  dialogData.value.isShow = true;
  dialogData.value.loading = false;

  nextTick(() => {
    formRef.value?.clearValidate();
  });
};

defineExpose({
  show,
});

</script>

<style lang="scss" scoped>
.create-gw-form {
  padding: 24px;
  .form-item-name {
    :deep(.bk-form-error) {
      position: relative;
    }
  }
  .form-item-name-tips {
    position: relative;
    top: -24px;
    left: 120px;
  }
}
.footer-actions {
  width: 100%;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}
.group-wrapper {
  width: 100%;
}

.token-wrapper {
  margin: 24px 0;
  padding: 12px 4px;
  background-color: #fafbfd;
  .default-c {
    float: right;
    margin-top: 4px;
  }
}
</style>
