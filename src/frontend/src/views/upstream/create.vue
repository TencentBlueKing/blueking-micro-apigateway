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
          <template #title>{{ t('基本信息') }}</template>
          <div>
            <!--  基本信息  -->
            <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
              <!-- name -->
              <bk-form-item :label="t('名称')" class="form-item" property="name" required>
                <bk-input v-model="formModel.name" :placeholder="t('名称')" clearable />
              </bk-form-item>
            </bk-form>

            <!--  upstream 配置  -->
            <form-upstream
              ref="upstream-form"
              v-model="upstream"
              v-model:flags="flags"
              v-model:ssl_id="formModel.ssl_id"
            />
          </div>
        </form-card>
      </div>
    </main>
    <form-page-footer @cancel="handleCancelClick" @submit="handleSubmit" />
  </div>
</template>

<script lang="ts" setup>
import { IUpstream, IUpstreamConfig } from '@/types/upstream';
import FormUpstream, { type IFlags } from '@/components/form/form-upstream.vue';
import { useUpstreamForm } from '@/views/upstream/use-upstream-form';
import { Form, InfoBox, Message } from 'bkui-vue';
import UPSTREAM_JSON from '@/assets/schemas/upstream.json';
import Ajv from 'ajv';
import FormPageFooter from '@/components/form/form-page-footer.vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, useTemplateRef, watch } from 'vue';
import { getUpstream, postUpstream, putUpstream } from '@/http/upstream';
import { cloneDeep, uniq, isPlainObject } from 'lodash-es';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
import useConfigFilter from '@/hooks/use-config-filter';
import useResourcePageDetector from '@/hooks/use-resource-page-detector';
import useElementScroll from '@/hooks/use-element-scroll';
import FormCard from '@/components/form-card.vue';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { showSchemaErrorMessages } = useSchemaErrorMessage();
const { showFirstErrorFormItem } = useElementScroll();
const { createDefaultUpstream } = useUpstreamForm();
const {
  filterEmpty,
  filterAdvanced,
} = useConfigFilter();
const { isEditMode, isCloneMode } = useResourcePageDetector();

const ajv = new Ajv();
const schemaValidate = ajv.compile(UPSTREAM_JSON);

const formModel = ref<Omit<IUpstream, 'config'>>({
  name: '',
  ssl_id: '',
});

const upstream = ref<IUpstreamConfig>(createDefaultUpstream());
const flags = ref<IFlags>({
  upstreamType: 'nodes',
  tlsType: '__disabled__',
});

const formRef = useTemplateRef<InstanceType<typeof Form>>('form-ref');
const upstreamFormRef = useTemplateRef('upstream-form');

const rules = {
  name: [
    { required: true, message: t('必填项'), trigger: 'change' },
  ],
};

const upstreamDtoId = computed(() => {
  return route.params.id as string;
});

watch(() => route.params.id, async (id: unknown) => {
  if (id) {
    const response = await getUpstream({ id } as { id: string });
    const { config, ...rest } = response;
    formModel.value = rest;
    // upstream.value = config;
    upstream.value = createDefaultUpstream(config);

    // 转换 nodes 格式
    if (config.nodes && isPlainObject(config.nodes)) {
      upstream.value.nodes = Object.entries(config.nodes)
        .map(([host, weight]) => {
          let port = 80;
          let hostArray: string[] = [];
          hostArray = host.split(':');

          if (host.includes(':')) {
            port = Number(hostArray.pop());
          }
          return {
            host: hostArray.join(''),
            port,
            weight,
          };
        });
    }

    if (config.service_name && config.discovery_type) {
      flags.value.upstreamType = 'service_discovery';
      delete upstream.value.nodes;
    }

    if (config.tls?.client_cert && config.tls?.client_key) {
      flags.value.tlsType = '__input__';
    } else if (config.tls?.client_cert_id) {
      flags.value.tlsType = '__select__';
    } else {
      flags.value.tlsType = '__disabled__';
    }

    if (!upstream.value.checks) {
      upstream.value.checks = {};
    }

    if (isCloneMode.value) {
      formModel.value.name += '_clone';
    }
  }
}, { immediate: true });

const handleSubmit = async () => {
  try {
    let upstreamCopy = cloneDeep(upstream.value);

    if (upstreamCopy.checks) {
      if (!Object.keys(upstreamCopy.checks).length) {
        delete upstreamCopy.checks;
      } else {
        if (!upstreamCopy.checks.active && upstreamCopy.checks.passive) {
          Message({
            theme: 'error',
            message: t('打开被动检查时也必须打开主动检查'),
          });
          return false;
        }

        if (upstreamCopy.checks.active) {
          let { req_headers } = upstreamCopy.checks.active;
          req_headers = uniq(req_headers?.filter(value => !!value));

          if (!req_headers?.length) {
            delete upstreamCopy.checks.active.req_headers;
          }
        }
      }
    }

    // 校验表单
    await Promise.all([
      formRef.value?.validate(),
      upstreamFormRef.value?.validate(),
    ]);

    upstreamCopy.labels = await upstreamFormRef.value?.setLabels();

    // if (!isEditMode.value) {
    // 过滤值为空或默认值的字段
    upstreamCopy = filterEmpty(upstreamCopy);
    // 过滤高级配置里没改动的值
    upstreamCopy = filterAdvanced(upstreamCopy, 'upstream');
    // }

    // 校验 schema
    if (schemaValidate(upstreamCopy)) {
      InfoBox({
        title: t('确认提交？'),
        confirmText: t('提交'),
        cancelText: t('取消'),
        onConfirm: async () => {
          const data = {
            name: formModel.value.name,
            ssl_id: formModel.value.ssl_id,
            config: upstreamCopy,
          };

          if (isEditMode.value) {
            await putUpstream({
              data,
              id: upstreamDtoId.value,
            });
          } else {
            await postUpstream({ data });
          }

          Message({
            theme: 'success',
            message: t('提交成功'),
          });

          await router.push({ name: 'upstream', replace: true });
        },
      });
    } else {
      showSchemaErrorMessages(schemaValidate.errors);
    }
  } catch (e) {
    const error = e as Error;
    showFirstErrorFormItem();
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
