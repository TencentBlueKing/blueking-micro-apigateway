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
    <bk-table
      :border="['row', 'outer']"
      :data="tableData"
      show-overflow-tooltip
      style="border-bottom: 1px solid #dcdee5;"
    >
      <bk-table-column :label="t('参数位置')" prop="in">
        <template #default="{ row }">
          {{ inMap[row.in] }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('参数名称')" prop="var" />
      <bk-table-column :label="t('非(!)')">
        <template #default="{ row }">
          {{ row.not ? 'true' : 'false' }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('运算符')">
        <template #default="{ row }">
          {{ operatorMap[row.operator] }}
        </template>
      </bk-table-column>
      <bk-table-column :label="t('参数值')" prop="val" />
      <bk-table-column :label="t('操作')">
        <template #default="{ row }">
          <bk-button
            style="margin-right: 12px;"
            text
            theme="primary"
            @click="() => handleEditClick(row)"
          >{{ t('编辑') }}
          </bk-button>
          <bk-button text theme="primary" @click="() => handleDeleteClick(row)">{{ t('删除') }}</bk-button>
        </template>
      </bk-table-column>
    </bk-table>
    <div class="tips">{{ t('非必填，支持通过请求头，请求参数、Cookie 进行路由匹配，可应用于灰度发布，蓝绿测试等场景') }}
    </div>
    <div style="margin-top: 24px;">
      <button-icon
        icon-color="#3a84ff"
        style="background: #f0f5ff;border-color: transparent;border-radius: 2px;color:#3a84ff;"
        @click="handleAddClick"
      >{{
        t('添加')
      }}
      </button-icon>
    </div>
    <slider-edit v-model="isEditSliderShow" :var-data="currentVar" @confirm="handleEditConfirm" />
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import ButtonIcon from '@/components/button-icon.vue';
import SliderEdit from './components/slider-edit.vue';
import { uniqueId } from 'lodash-es';

interface IProps {
  vars?: string[][];
}

interface ITableRow {
  id: string;
  in: string;
  var: string;
  not?: boolean;
  operator: string;
  val: string;
}

const { vars } = defineProps<IProps>();

const { t } = useI18n();

const tableData = ref<ITableRow[]>([]);
const currentVar = ref<ITableRow | null>(null);
const isEditSliderShow = ref(false);

const inMap: Record<string, string> = {
  http: t('HTTP 请求头'),
  arg: t('请求参数'),
  post_arg: t('POST 请求参数'),
  cookie: 'Cookie',
  param: t('内置参数'),
};

const operatorMap: Record<string, string> = {
  '==': t('等于（==）'),
  '~=': t('不等于（~=）'),
  '>': t('大于（>）'),
  '<': t('小于（<）'),
  '~~': t('正则匹配（~~）'),
  '~*': t('不区分大小写的正则匹配（~*）'),
  IN: 'IN',
  HAS: 'HAS',
};

const getInAndVar = (inAndVar: string) => {
  const strArr = inAndVar.split('_');
  const _var = strArr.pop();
  const _in = strArr.join('_') || 'param';
  return [_in, _var];
};

watch(() => vars, () => {
  if (vars?.length) {
    vars.forEach((item) => {
      if (item.length === 3) {
        const [in_var, operator, val] = item;
        const [_in, _var] = getInAndVar(in_var);
        tableData.value.push({
          id: uniqueId(),
          in: _in,
          var: _var,
          not: false,
          operator,
          val,
        });
      } else {
        const [in_var, not, operator, val] = item;
        const [_in, _var] = getInAndVar(in_var);
        tableData.value.push({
          id: uniqueId(),
          in: _in,
          var: _var,
          not: not === '!',
          operator,
          val,
        });
      }
    });
  }
}, { deep: true, immediate: true });

const handleAddClick = () => {
  currentVar.value = null;
  isEditSliderShow.value = true;
};

const handleEditClick = (row: ITableRow) => {
  currentVar.value = { ...row };
  isEditSliderShow.value = true;
};

const handleDeleteClick = (row: ITableRow) => {
  const index = tableData.value.findIndex(item => item.id === row.id);
  tableData.value.splice(index, 1);
};

const handleEditConfirm = (form: ITableRow) => {
  if (currentVar.value) {
    const index = tableData.value.findIndex(item => item.id === currentVar.value.id);
    tableData.value.splice(index, 1, form);
  } else {
    tableData.value.push(form);
  }
  currentVar.value = null;
};

defineExpose({
  getValue: () => {
    return tableData.value.map((item) => {
      const { in: _in, var: _var, not, operator, val } = item;
      const result: string[] = [];
      if (_in === 'param') {
        result.push(_var);
      } else {
        result.push(`${_in}_${_var}`);
      }
      if (not === true) {
        result.push('!');
      }
      result.push(operator);
      result.push(val);
      return result;
    });
  },
});
</script>

<style lang="scss">
.tips {
  font-size: 12px;
  line-height: 1.2;
  margin-top: 4px;
  color: #979ba5;
}
</style>
