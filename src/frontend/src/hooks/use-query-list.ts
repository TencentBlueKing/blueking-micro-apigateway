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

/*
 * 列表分页、查询hooks
 * 需要传入获取列表的方法名apiMethod、当前列表的过滤条件filterData
 */

import { IPagination, IQueryListParams } from '@/types';
import { useCommon } from '@/store';
import { sortBy, sortedUniq } from 'lodash-es';
import useMaxTableLimit from '@/hooks/use-max-table-limit';
import { onMounted, ref, watch } from 'vue';

export function useQueryList<T>({
  apiMethod,
  filterData,
  id,
  resetPageOnFilterDataChange,
  initialPagination = {},
}: IQueryListParams) {
  const common = useCommon();
  // 当前视口高度能展示最多多少条表格数据
  const maxTableLimit = useMaxTableLimit();
  const { gatewayId } = common;
  const initPagination: IPagination = {
    offset: 0,
    // 每页页数选项，这个也是 table 组件的默认值
    limitList: sortedUniq(sortBy([maxTableLimit, 10, 20, 50, 100])),
    limit: maxTableLimit,
    count: 0,
    small: false,
    // 获取接口是否异常
    abnormal: false,
    // 当前页码
    current: 1,
    ...initialPagination,
  };

  const pagination = ref<IPagination>({ ...initPagination });
  const isLoading = ref(false);
  const tableData = ref<T[]>([]);
  const getMethod = ref<any>(null);

  // 获取列表数据的方法
  const getList = async (fetchMethod = apiMethod, needLoading = true) => {
    getMethod.value = fetchMethod;
    isLoading.value = needLoading;
    // 列表参数
    const query = {
      offset: pagination.value.offset,
      limit: pagination.value.limit,
      ...filterData.value,
    };
    try {
      const response = id ? await getMethod.value({ gatewayId, id, query }) : await getMethod.value({
        gatewayId,
        query,
      });
      tableData.value = response.results || response.data || [];
      // pagination.value = Object.assign(pagination.value, {
      //   count: response.count || 0,
      //   abnormal: false,
      // });
      pagination.value = {
        ...pagination.value,
        count: response.count || 0,
        abnormal: false,
      };
    } catch (error) {
      pagination.value.abnormal = true;
    } finally {
      isLoading.value = false;
    }
  };

  // 页码变化发生的事件
  const handlePageChange = async (current: number) => {
    pagination.value.offset = pagination.value.limit * (current - 1);
    pagination.value.current = current;
    await fetchList();
  };

  // 条数变化发生的事件
  const handlePageSizeChange = async (limit: number) => {
    pagination.value.limit = limit;
    pagination.value.offset = limit * (pagination.value.current - 1);

    // 页码没变化的情况下需要手动请求一次数据
    if (pagination.value.offset <= pagination.value.count) {
      await fetchList();
    }
  };

  // 监听筛选条件的变化
  watch(
    () => filterData,
    async () => {
      if (!resetPageOnFilterDataChange) {
        pagination.value = { ...initPagination };
      }

      await fetchList();
    },
    { deep: true },
  );

  const fetchList = async () => {
    if (getMethod.value) {
      await getList(getMethod.value);
    } else {
      await getList();
    }
  };

  onMounted(async () => {
    await getList();
  });

  return {
    tableData,
    getMethod,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    getList,
    fetchList,
  };
}
