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

/**
 * 选择相关状态和事件
 */
import { nextTick, ref } from 'vue';

export type SelectionType = {
  checked: boolean;
  data: any[];
  isAll?: boolean;
  row?: any
};

export const useSelection = (disabledSelected?: (item: any) => boolean) => {
  const selections = ref([]);
  const bkTableRef = ref();

  const handleSelectionChange = (selection: SelectionType) => {
    // 选择某一个
    if (selection.checked) {
      selections.value.push(selection.row);
    }
    // 取消选择某一个
    if (!selection.checked) {
      const index = selections.value.findIndex((item: any) => item.id === selection.row.id);
      selections.value.splice(index, 1);
    }
  };

  const handleSelectAllChange = (selection: SelectionType) => {
    if (selection.checked) {
      if (disabledSelected) {
        const list = selection.data.filter((item: any) => {
          return disabledSelected(item);
        });
        selections.value = JSON.parse(JSON.stringify(list));
      } else {
        selections.value = JSON.parse(JSON.stringify(selection.data));
      }
    } else {
      selections.value = [];
    }
  };

  const resetSelections = (tableRef: { clearSelection: () => void; }) => {
    nextTick(() => {
      selections.value = [];
      tableRef?.clearSelection();
    });
  };

  return {
    selections,
    bkTableRef,
    handleSelectionChange,
    handleSelectAllChange,
    resetSelections,
  };
};
