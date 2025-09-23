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

import { ref } from 'vue';
import { IDiffGroup, IChangeDetail } from '@/http/publish';
import { IPublishSearch } from '@/types';

export default function usePublishSearch({
  filterData,
  diffGroupTotal,
  searchDoneFn,
}: IPublishSearch) {
  const diffGroupShow = ref<IDiffGroup[]>([]);

  const regroupData = () => {
    const keywordsItems = filterData.value?.keywords?.split('&')?.slice(1);

    const {
      id: searchId = '',
      name: searchName = '',
      operation_type: searchOperationType = '',
    } = filterData.value;

    if (!keywordsItems?.length && !searchId && !searchName && !searchOperationType) {
      diffGroupShow.value = diffGroupTotal.value;

      searchDoneFn && searchDoneFn();
      return;
    }

    let list: any = diffGroupTotal.value?.map((source: IDiffGroup) => {
      return source.change_detail?.map((item) => {
        item.resource_type = source.resource_type;
        return item;
      }) || [];
    });

    list = list.flat();

    list = list?.filter((item: IChangeDetail) => {
      const { resource_id, name, operation_type } = item;

      if (keywordsItems?.some((keywords: string) => (!resource_id?.includes(keywords)
      && !name?.includes(keywords)
      && !operation_type?.includes(keywords)))) {
        return false;
      }

      if (searchId && !resource_id?.includes(searchId)) {
        return false;
      }

      if (searchName && !name?.includes(searchName)) {
        return false;
      }

      if (searchOperationType && !operation_type?.includes(searchOperationType)) {
        return false;
      }

      return true;
    });

    const sourceMap: any = {};

    list?.forEach((item: IChangeDetail) => {
      if (sourceMap[item.resource_type]) {
        sourceMap[item.resource_type] = [...sourceMap[item.resource_type], item];
      } else {
        sourceMap[item.resource_type] = [item];
      }
    });

    const diffList: IDiffGroup[] = [];
    Object.keys(sourceMap)?.forEach((type: string) => {
      const diff: IDiffGroup = {
        added_count: 0,
        deleted_count: 0,
        modified_count: 0,
        resource_type: type,
        change_detail: [],
      };

      sourceMap[type]?.forEach((item: IChangeDetail) => {
        if (item.operation_type === 'create') {
          diff.added_count += 1;
        }

        if (item.operation_type === 'delete') {
          diff.deleted_count += 1;
        }

        if (item.operation_type === 'update') {
          diff.modified_count += 1;
        }

        diff.change_detail.push(item);
      });

      diffList.push(diff);
    });

    diffGroupShow.value = diffList;

    searchDoneFn && searchDoneFn();
  };

  return {
    regroupData,
    diffGroupShow,
  };
}
