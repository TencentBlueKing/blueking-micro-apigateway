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

// 获取网关列表的hooks, 多个地方用到
import { ref, watch } from 'vue';
import { getGatewaysList } from '@/http';
import { IPagination } from '@/types';

const initPagination: IPagination = {
  offset: 0,
  limit: 10000,
  count: 0,
};
const pagination = ref<IPagination>(initPagination);
const dataList = ref<any[]>([]);

export const useGetApiList = (filter: any) => {
  const getGatewaysListData = async () => {
    try {
      const parmas = {
        limit: pagination.value.limit,
        offset: pagination.value.limit * pagination.value.offset,
        ...filter.value,
      };
      const res = await getGatewaysList(parmas);
      dataList.value = res;
      return dataList.value;
    } catch (error) {}
  };

  // 监听输入框改变
  watch(
    () => filter,
    () => {
      getGatewaysListData();
    },
    { deep: true },
  );

  return {
    getGatewaysListData,
    dataList,
    pagination,
  };
};
