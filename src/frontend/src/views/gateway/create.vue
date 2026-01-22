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
    :title="formData?.id ? t('编辑网关') : t('新建网关')"
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
          :maxlength="30"
          :disabled="!!formData?.id"
          show-word-limit
          :placeholder="t('请输入小写字母、数字、连字符(-)，以小写字母开头')"
          clearable
          autofocus
        />
      </bk-form-item>
      <span class="common-form-tips form-item-name-tips">
        {{ t('网关的唯一标识，创建后不可更改') }}
      </span>
      <bk-form-item
        :label="t('网关管理员')"
        property="maintainers"
        required
      >
        <member-select v-model="formData.maintainers" />
      </bk-form-item>
      <bk-form-item
        :label="t('描述')"
        property="description"
      >
        <bk-input
          type="textarea"
          v-model="formData.description"
          :placeholder="t('请输入网关描述')"
          :maxlength="500"
          :clearable="false"
        />
      </bk-form-item>

      <form-divider>etcd</form-divider>
      <bk-form-item
        class="form-item-name"
        :label="t('etcd 地址')"
        property="etcd_endpoints"
        required
      >
        <form-etcd-endpoints ref="etcd_endpoints-form" v-model="formData.etcd_endpoints" />
      </bk-form-item>
      <bk-form-item
        class="form-item-name"
        :label="t('etcd 前缀')"
        property="etcd_prefix"
        required
      >
        <bk-input
          v-model="formData.etcd_prefix"
          :placeholder="t('请输入 etcd 前缀')"
          :disabled="!!formData?.id"
          clearable
        />
      </bk-form-item>

      <bk-form-item
        class="form-item-name"
        :label="t('etcd 连接类型')"
        property="etcd_schema_type"
        required
      >
        <bk-radio-group
          v-model="formData.etcd_schema_type"
          :placeholder="t('请选择 etcd 连接类型')"
          type="card"
        >
          <bk-radio-button label="http">HTTP</bk-radio-button>
          <bk-radio-button label="https">HTTPS</bk-radio-button>
        </bk-radio-group>
      </bk-form-item>

      <div v-if="formData.etcd_schema_type === 'http'">
        <bk-form-item
          class="form-item-name"
          :label="t('etcd 用户名')"
          property="etcd_username"
          required
        >
          <bk-input
            v-model="formData.etcd_username"
            :placeholder="t('请输入 etcd 用户名')"
            clearable
          />
        </bk-form-item>
        <bk-form-item
          class="form-item-name"
          :label="t('etcd 密码')"
          property="etcd_password"
          required
        >
          <bk-input
            v-model="formData.etcd_password"
            :placeholder="t('请输入 etcd 密码')"
            type="password"
            class="flex-primary"
            clearable
          />
        </bk-form-item>
        <span class="common-form-tips form-item-name-tips">
          {{ t('敏感信息会加密存储确保数据安全') }}
        </span>
      </div>

      <div v-else>
        <bk-form-item
          class="form-item-name"
          :label="t('CACert')"
          property="etcd_ca_cert"
          required
        >
          <bk-input
            v-model="formData.etcd_ca_cert"
            :placeholder="t('请输入 CACert')"
            type="textarea"
          />
        </bk-form-item>
        <bk-form-item
          class="form-item-name"
          :label="t('Cert')"
          property="etcd_cert_cert"
          required
        >
          <bk-input
            v-model="formData.etcd_cert_cert"
            :placeholder="t('请输入 Cert')"
            type="textarea"
          />
        </bk-form-item>
        <bk-form-item
          class="form-item-name"
          :label="t('私钥')"
          property="etcd_cert_key"
          required
        >
          <bk-input
            v-model="formData.etcd_cert_key"
            :placeholder="t('请输入私钥')"
            type="textarea"
          />
        </bk-form-item>
        <span class="common-form-tips form-item-name-tips">
          {{ t('敏感信息会加密存储确保数据安全') }}
        </span>
      </div>

      <bk-form-item
        class="form-item-name"
        label=""
      >
        <bk-button
          theme="primary"
          class="group-wrapper"
          outline
          @click="handleConnect"
          :loading="loadingConnect"
          :disabled="disableConnect">
          {{ t('etcd 连通测试') }}
        </bk-button>
      </bk-form-item>

      <form-divider>{{ t('APISIX') }}</form-divider>
      <bk-form-item
        class="form-item-name"
        :label="t('APISIX 类型')"
        property="apisix_type"
        required
      >
        <bk-select
          v-model="formData.apisix_type"
          :placeholder="t('请选择 APISIX 类型')"
          :disabled="!!formData?.id"
          @change="handleApisixTypeChange"
        >
          <bk-option
            v-for="item in apisixTypeList"
            :key="item.value"
            :id="item.value"
            :name="item.label"
          />
        </bk-select>
      </bk-form-item>
      <bk-form-item
        class="form-item-name"
        :label="t('APISIX 版本')"
        property="apisix_version"
        required
      >
        <bk-select
          v-model="formData.apisix_version"
          :placeholder="t('请选择 APISIX 版本')"
          :disabled="!!formData?.id"
        >
          <bk-option
            v-for="item in versionList"
            :key="item.value"
            :id="item.value"
            :name="item.label"
          />
        </bk-select>
      </bk-form-item>

      <form-divider>{{ t('其他') }}</form-divider>
      <bk-form-item
        :label="t('只读模式')"
        class="form-item-name"
        property="read_only"
      >
        <bk-switcher v-model="formData.read_only" theme="primary"></bk-switcher>
      </bk-form-item>
      <span class="common-form-tips form-item-name-tips">
        {{ t('开启只读模式，可以从 etcd 同步数据，在控制面中仅展示，不可变更，不可发布') }}
      </span>
      <!-- <bk-form-item
        :label="t('模式')"
        property="mode"
        required
      >
        <bk-radio-group v-model="formData.mode">
          <bk-radio :label="1">{{ t('直管') }}</bk-radio>
          <bk-radio :label="2">{{ t('纳管') }}</bk-radio>
        </bk-radio-group>
      </bk-form-item> -->

    </bk-form>

    <template #footer>
      <div class="footer-actions">
        <bk-button theme="primary" :loading="dialogData.loading" @click="handleConfirmCreate">{{ t('确定') }}</bk-button>
        <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { cloneDeep } from 'lodash-es';
import { ICheckName, IConnectTest, ICreatePayload, IDialog } from '@/types';
import MemberSelect from '@/components/member-select';
// @ts-ignore
import FormDivider from '@/components/form-divider.vue';
// @ts-ignore
import formEtcdEndpoints from '@/components/form/form-etcd-endpoints.vue';
import { checkName, createGateway, etcdConnectTest, updateGateways } from '@/http';
import { useCommon } from '@/store';
import { useSidebar } from '@/hooks/index';

const emit = defineEmits(['done']);

const common = useCommon();
const { t } = useI18n();
const { initSidebarFormData, isSidebarClosed } = useSidebar();

const disableConnect = ref<boolean>(true);
const loadingConnect = ref<boolean>(false);
const isProbe = ref<boolean>(false);
const formRef = ref(null);
const contrastData = ref<string>('');
const etcdEndpointsFormRef = useTemplateRef('etcd_endpoints-form');
const formData = ref<ICreatePayload>({
  name: '',
  apisix_version: '',
  apisix_type: 'bk-apisix',
  mode: 1,
  etcd_endpoints: [],
  etcd_prefix: '/apisix',
  etcd_username: 'root',
  etcd_password: '',
  maintainers: [],
  description: '',
  etcd_schema_type: 'http',
  etcd_ca_cert: '',
  etcd_cert_cert: '',
  etcd_cert_key: '',
  read_only: false,
});
const dialogData = ref<IDialog>({
  isShow: false,
  title: t('新建网关'),
  loading: false,
});

const rules = {
  name: [
    {
      required: true,
      message: t('请填写名称'),
      trigger: 'change',
    },
    {
      validator: (value: string) => value.length >= 3,
      message: t('不能小于3个字符'),
      trigger: 'change',
    },
    {
      validator: (value: string) => value.length <= 30,
      message: t('不能多于30个字符'),
      trigger: 'change',
    },
    {
      validator: (value: string) => {
        const reg = /^[a-z][a-z0-9-]*$/;
        return reg.test(value);
      },
      message: t('由小写字母、数字、连接符（-）组成，首字符必须是小写字母，长度大于3小于30个字符'),
      trigger: 'change',
    },
    {
      validator: async (value: string) => {
        try {
          if (!value) return true;

          const data: ICheckName = {
            name: value,
          };
          const { id } = formData.value;
          if (id) {
            data.id = id;
          }

          const response = await checkName(data);
          return response?.status === 'ok';
        } catch (error) {
          return false;
        }
      },
      message: t('与现有的网关名称重复了'),
      trigger: 'change',
    },
  ],
};

const apisixTypeList = computed(() => {
  return Object.keys(common.enums.apisix_type).map((key: string) => ({
    label: common.enums.apisix_type[key],
    value: key,
  }));
});

const versionList = computed(() => {
  return (common.enums.support_apisix_version[formData.value.apisix_type]?.support_version ?? [])
    .map((item: string) => ({
      label: item,
      value: item,
    }));
});

watch(
  () => formData.value,
  (v) => {
    const {
      etcd_endpoints,
      etcd_prefix,
      etcd_schema_type,
      etcd_username,
      etcd_password,
      etcd_ca_cert,
      etcd_cert_cert,
      etcd_cert_key,
      id,
    } = v;

    if (!etcd_endpoints?.length || !etcd_prefix) {
      disableConnect.value = true;
      isProbe.value = false;
      return;
    }

    if (id) {
      let now: any = { etcd_endpoints, etcd_prefix };

      if (etcd_schema_type === 'http') {
        now = { ...now, etcd_username, etcd_password };
      } else {
        now = { ...now, etcd_ca_cert, etcd_cert_cert, etcd_cert_key };
      }

      if (Object.values(now)?.some((value: any) => !value?.length)) {
        disableConnect.value = true;
        isProbe.value = false;
      } else {
        if (!contrastUpdate()) {
          disableConnect.value = true;
          isProbe.value = true;
        } else {
          disableConnect.value = false;
          isProbe.value = false;
        };
      }
    } else {
      if (etcd_schema_type === 'http') {
        if (etcd_username && etcd_password) {
          disableConnect.value = false;
        } else {
          disableConnect.value = true;
        }
      } else {
        if (etcd_ca_cert && etcd_cert_cert && etcd_cert_key) {
          disableConnect.value = false;
        } else {
          disableConnect.value = true;
        }
      }
    }
  },
  {
    deep: true,
  },
);

const contrastUpdate = () => {
  let isUpdate = false;

  if (!contrastData.value) {
    return isUpdate;
  }

  const {
    etcd_endpoints,
    etcd_prefix,
    etcd_schema_type,
    etcd_username,
    etcd_password,
    etcd_ca_cert,
    etcd_cert_cert,
    etcd_cert_key,
  } = formData.value;

  const contrast = JSON.parse(contrastData.value) || {};

  let origin: any = {
    etcd_endpoints: contrast.etcd_endpoints,
    etcd_prefix: contrast.etcd_prefix,
  };

  let now: any = { etcd_endpoints, etcd_prefix };

  if (etcd_schema_type === 'http') {
    origin = {
      ...origin,
      etcd_username: contrast.etcd_username,
      etcd_password: contrast.etcd_password,
    };
    now = { ...now, etcd_username, etcd_password };
  } else {
    origin = {
      ...origin,
      etcd_ca_cert: contrast.etcd_ca_cert,
      etcd_cert_cert: contrast.etcd_cert_cert,
      etcd_cert_key: contrast.etcd_cert_key,
    };
    now = { ...now, etcd_ca_cert, etcd_cert_cert, etcd_cert_key };
  }

  isUpdate = JSON.stringify(origin) !== JSON.stringify(now);

  return isUpdate;
};

const handleApisixTypeChange = () => {
  formData.value.apisix_version = '';
};

const handleConnect = async () => {
  loadingConnect.value = true;
  try {
    if (!await etcdEndpointsFormRef.value?.validate()) return;

    const {
      etcd_username,
      etcd_password,
      etcd_prefix,
      etcd_endpoints,
      etcd_schema_type,
      etcd_ca_cert,
      etcd_cert_cert,
      etcd_cert_key,
    } = formData.value;

    const params: IConnectTest = {
      etcd_endpoints,
      etcd_prefix,
      etcd_schema_type,
    };

    if (etcd_schema_type === 'http') {
      params.etcd_username = etcd_username;
      params.etcd_password = etcd_password;
    } else {
      params.etcd_ca_cert = etcd_ca_cert;
      params.etcd_cert_cert = etcd_cert_cert;
      params.etcd_cert_key = etcd_cert_key;
    }

    if (formData.value?.id) {
      params.gateway_id = formData.value.id;
    }

    const response = await etcdConnectTest(params);

    Message({
      theme: 'success',
      message: t('连接成功'),
    });
    if (response?.apisix_version) {
      formData.value.apisix_version = response.apisix_version;
    }
    isProbe.value = true;
  } catch (e) {
    isProbe.value = false;
  } finally {
    loadingConnect.value = false;
  }
};

// 创建网关确认
const handleConfirmCreate = async () => {
  try {
    await Promise.all([
      formRef.value.validate(),
      etcdEndpointsFormRef.value?.validate(),
    ]);
    if (!isProbe.value) {
      return Message({
        message: t('请先完成 etcd 连通测试……'),
        theme: 'warning',
      });
    }

    dialogData.value.loading = true;
    if (formData.value?.id) {
      const payload = cloneDeep(formData.value);
      const {
        etcd_ca_cert,
        etcd_cert_cert,
        etcd_cert_key,
        etcd_password,
      } = payload;

      if (payload.etcd_schema_type === 'https') {
        if (etcd_ca_cert.includes('**********')) {
          payload.etcd_ca_cert = undefined;
        }
        if (etcd_cert_cert.includes('**********')) {
          payload.etcd_cert_cert = undefined;
        }
        if (etcd_cert_key.includes('**********')) {
          payload.etcd_cert_key = undefined;
        }

        payload.etcd_username = undefined;
        payload.etcd_password = undefined;
      } else {
        if (etcd_password === '******') {
          payload.etcd_password = undefined;
        }

        payload.etcd_ca_cert = undefined;
        payload.etcd_cert_cert = undefined;
        payload.etcd_cert_key = undefined;
      }

      // if (!contrastUpdate()) {
      //   payload.etcd_schema_type = undefined;
      // }

      await updateGateways(payload.id, payload);
      Message({
        message: t('更新成功'),
        theme: 'success',
      });
    } else {
      await createGateway(formData.value);
      Message({
        message: t('创建成功'),
        theme: 'success',
      });
    }
    dialogData.value.isShow = false;
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
    apisix_version: '',
    apisix_type: 'bk-apisix',
    mode: 1,
    etcd_endpoints: [],
    etcd_prefix: '/apisix',
    etcd_username: 'root',
    etcd_password: '',
    maintainers: [],
    description: '',
  };
  nextTick(() => {
    formRef.value?.clearValidate();
  });
};

const handleBeforeClose = () => {
  return isSidebarClosed(JSON.stringify(formData.value));
};

const show = (initData: ICreatePayload) => {
  formData.value = cloneDeep(initData);
  contrastData.value = JSON.stringify(initData);
  initSidebarFormData(formData.value);
  dialogData.value.isShow = true;
  dialogData.value.loading = false;

  nextTick(() => {
    formRef.value?.clearValidate();
    etcdEndpointsFormRef.value?.validate();
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
</style>
