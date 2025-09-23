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
    :title="t('规则')"
    width="960"
    @close="resetEditor"
  >
    <template #default>
      <div class="content-wrapper">
        <bk-form ref="form-ref" :model="formModel" :rules="rules" class="form-element">
          <bk-form-item :label="t('参数位置')" class="form-item" property="in" required>
            <bk-select v-model="formModel.in" :clearable="false" :filterable="false">
              <bk-option
                v-for="type in inList"
                :id="type.id"
                :key="type.id"
                :name="type.name"
              />
            </bk-select>
          </bk-form-item>

          <bk-form-item :label="t('参数名称')" class="form-item" property="var" required>
            <bk-input v-model="formModel.var" clearable />
          </bk-form-item>

          <bk-form-item :label="t('非(!)')" class="form-item" property="not" required>
            <bk-switcher v-model="formModel.not" theme="primary" />
          </bk-form-item>

          <bk-form-item :label="t('运算符')" class="form-item" property="operator" required>
            <bk-select v-model="formModel.operator" :clearable="false" :filterable="false">
              <bk-option
                v-for="type in operatorList"
                :id="type.id"
                :key="type.id"
                :name="type.name"
              />
            </bk-select>
          </bk-form-item>

          <bk-form-item :label="t('参数值')" class="form-item" property="val" required>
            <bk-input v-model="formModel.val" clearable />
          </bk-form-item>
        </bk-form>
      </div>
    </template>
    <template #footer>
      <div class="footer-actions">
        <bk-button theme="primary" @click="handleConfirmClick">{{ t('保存') }}</bk-button>
        <bk-button @click="handleCancelClick">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, useTemplateRef, watch } from 'vue';
import { uniqueId } from 'lodash-es';

interface IProps {
  varData?: IForm;
}

interface IForm {
  id: string;
  in: string;
  var: string;
  not?: boolean;
  operator: string;
  val: string;
}

const isShow = defineModel<boolean>({
  required: true,
  default: false,
});

const { varData } = defineProps<IProps>();

const emits = defineEmits<{
  (e: 'confirm', form: IForm): void
}>();

const { t } = useI18n();

const formRef = useTemplateRef('form-ref');
const formModel = ref<IForm>({
  id: uniqueId(),
  in: 'http',
  var: '',
  not: false,
  operator: '==',
  val: '',
});

const rules = {
  in: [{ required: true, message: t('必填项') }],
  var: [{ required: true, message: t('必填项') }],
  not: [{ required: true, message: t('必填项') }],
  operator: [{ required: true, message: t('必填项') }],
  val: [{ required: true, message: t('必填项') }],
};

const inList = [
  {
    id: 'http',
    name: t('HTTP 请求头'),
  },
  {
    id: 'arg',
    name: t('请求参数'),
  },
  {
    id: 'post_arg',
    name: t('POST 请求参数'),
  },
  {
    id: 'cookie',
    name: 'Cookie',
  },
  {
    id: 'param',
    name: t('内置参数'),
  },
];

const operatorList = [
  {
    id: '==',
    name: t('等于（==）'),
  },
  {
    id: '~=',
    name: t('不等于（~=）'),
  },
  {
    id: '>',
    name: t('大于（>）'),
  },
  {
    id: '<',
    name: t('小于（<）'),
  },
  {
    id: '~~',
    name: t('正则匹配（~~）'),
  },
  {
    id: '~*',
    name: t('不区分大小写的正则匹配（~*）'),
  },
  {
    id: 'IN',
    name: 'IN',
  },
  {
    id: 'HAS',
    name: 'HAS',
  },
];

watch(() => varData, () => {
  if (varData) {
    formModel.value = { ...varData };
  }
}, { deep: true, immediate: true });

const handleConfirmClick = async () => {
  try {
    await formRef.value!.validate();
    emits('confirm', { ...formModel.value });
    isShow.value = false;
    resetEditor();
  } catch (error) {
    return;
  }
};

const handleCancelClick = () => {
  isShow.value = false;
  resetEditor();
};

const resetEditor = () => {
  formModel.value = {
    id: uniqueId(),
    in: 'http',
    var: '',
    not: false,
    operator: '==',
    val: '',
  };
};

</script>

<style lang="scss" scoped>

.content-wrapper {
  font-size: 14px;
  padding: 24px;

  .actions {
    display: flex;
    margin-bottom: 24px;
    gap: 12px;
  }
}

.footer-actions {
  display: flex;
  gap: 12px;
}

</style>
