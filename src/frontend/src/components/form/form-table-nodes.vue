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
    <table>
      <thead>
        <tr>
          <th>{{ t('主机名') }}<span class="required-mark">*</span></th>
          <th>{{ t('端口') }}<span class="required-mark">*</span></th>
          <th>{{ t('权重') }}<span class="required-mark">*</span></th>
          <th>{{ t('操作') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(node, index) in nodes" :key="index">
          <td>
            <bk-input
              v-model="node.host"
              :class="{ 'is-error': isErrorField('host', index) }" :style="inputStyle" clearable
              @input="clearValidationState('host', index)"
            />
          </td>
          <td>
            <bk-input
              v-model="node.port"
              :class="{ 'is-error': isErrorField('port', index) }"
              :min="0" :precision="0" :step="1" :style="inputStyle" clearable type="number"
              @change="clearValidationState('port', index)"
            />
          </td>
          <td>
            <bk-input
              v-model="node.weight" :class="{ 'is-error': isErrorField('weight', index) }"
              :min="0" :precision="0" :step="1"
              :style="inputStyle" type="number"
              @change="clearValidationState('weight', index)"
            />
          </td>
          <td>
            <div class="cell-actions">
              <icon
                v-if="nodes.length > 1"
                class="icon-btn" color="#979BA5" name="minus-circle-shape" size="18" @click="handleRemoveItem(index)"
              />
              <icon
                class="icon-btn" color="#979BA5" name="plus-circle-shape" size="18" @click="handleAddItem"
              />
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="validationErrors.length" class="error-message">{{ validationErrors[0].message }}</div>
  </div>
</template>
<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import Icon from '@/components/icon.vue';
import { INode } from '@/types/common';
import { isInteger } from 'lodash-es';

interface IValidationError {
  message: string;
  rowIndex: number;
  key: string;
}

const nodes = defineModel<INode[]>({
  required: true,
});

const { t } = useI18n();

const validationErrors = ref<IValidationError[]>([]);

const inputStyle = ref({
  border: 'none',
  height: '40px',
});

const isErrorField = (key: string, index: number) => {
  if (!validationErrors.value.length) {
    return false;
  }
  return validationErrors.value.some((error) => {
    return error.rowIndex === index && error.key === key;
  });
};

const validate = async () => {
  clearValidationState();
  try {
    nodes.value.forEach((node, index) => {
      // 校验 host
      if (node.host === '') {
        validationErrors.value.push({
          message: t('请输入主机名或IP'),
          rowIndex: index,
          key: 'host',
        });
      } else if (!/^\*?[0-9a-zA-Z-._[\]:]+$/.test(node.host)) {
        validationErrors.value.push({
          message: t('主机名仅支持字母、数字、-、_和 *，但 * 需要在开头位置'),
          rowIndex: index,
          key: 'host',
        });
      }
      // 校验 port
      // @ts-ignore
      if (node.port === '' || node.port === null || !isInteger(Number(node.port))) {
        validationErrors.value.push({
          message: t('端口必须为整数'),
          rowIndex: index,
          key: 'port',
        });
      }
      // 校验 weight
      // @ts-ignore
      if (node.weight === '') {
        validationErrors.value.push({
          message: t('请输入权重'),
          rowIndex: index,
          key: 'weight',
        });
      } else
        // @ts-ignore
        if (!(node.weight === '' || isInteger(Number(node.weight)))) {
          validationErrors.value.push({
            message: t('权重必须为整数'),
            rowIndex: index,
            key: 'weight',
          });
        }
    });

    if (validationErrors.value.length) {
      throw Error(validationErrors.value[0].message, {
        cause: validationErrors,
      });
    }
  } catch (e) {
    throw e;
  }
};

const handleAddItem = () => {
  clearValidationState();
  nodes.value.push({
    host: '',
    port: null,
    weight: 1,
  });
};

const handleRemoveItem = (index: number) => {
  clearValidationState('', index);
  nodes.value.splice(index, 1);
};

const clearValidationState = (key = '', index = -1) => {
  if (key) {
    if (index !== -1) {
      const errorIndex = validationErrors.value.findIndex(error => error.key === key && error.rowIndex === index);
      if (errorIndex !== -1) {
        validationErrors.value.splice(errorIndex, 1);
      }
    } else {
      validationErrors.value = validationErrors.value.filter(error => error.key !== key);
    }
  } else if (index !== -1) {
    validationErrors.value = validationErrors.value.filter(error => error.rowIndex !== index);
  } else {
    validationErrors.value = [];
  }
};

defineExpose({
  validate,
});

</script>
<style lang="scss" scoped>

.required-mark {
  font-size: 14px;
  padding-left: 4px;
  color: #ea3636;
}

table {
  font-size: 12px;
  font-weight: normal;
  width: 640px;
  color: #313238;

  border: 1px solid #dcdee5;

  tr {
    height: 40px;
  }

  thead {
    th {
      font-weight: normal;
      width: 25%;
      padding-left: 12px;
      color: #313238;
      border-right: 1px solid #dcdee5;
      border-bottom: 1px solid #dcdee5;
      background: #f5f7fa;

      &:nth-child(1) {
        width: 35%;
      }

      &:nth-child(2) {
        width: 20%;
      }

      &:nth-child(3) {
        width: 20%;
      }

      &:nth-child(4) {
        width: 25%;
      }
    }
  }

  tbody {
    tr:not(:last-child) {
      td {
        border-bottom: 1px solid #dcdee5;
      }
    }

    td {
      border-right: 1px solid #dcdee5;
    }
  }
}

.cell-actions {
  display: flex;
  align-items: center;
  padding-left: 12px;
  gap: 12px;
}

.error-message {
  font-size: 12px;
  line-height: 1;
  padding-top: 4px;
  color: #ea3636;
}

// input 错误态样式
.bk-input.is-error {
  background: #fff0f1;

  :deep(input) {
    background: transparent;
  }

  :deep(.bk-input--clear-icon) {
    background: #fff0f1;
  }
}

</style>
