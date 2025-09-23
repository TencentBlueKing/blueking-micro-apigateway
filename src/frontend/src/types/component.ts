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

import { Column, RenderFunctionString } from 'bkui-vue/lib/table/props';
// import 'vue/jsx';
import { Ref } from 'vue';
import { IDiffGroup } from '@/http/publish';

// 分页interface
export interface IPagination {
  small?: boolean
  offset: number
  limit: number
  count: number
  abnormal?: boolean
  limitList?: number[]
  current?: number
}

export interface IMenuGroup {
  name: string
  menus: IMenu[]
}

export interface IMenu {
  name: string
  title: string
  icon?: string
  routeName: string  // 路由名称，必须和路由配置中的 name 一致
  enabled?: boolean
  children?: IMenu[]
}

export interface ITableEmptyConfig {
  keyword: string
  isAbnormal: boolean
}

export interface IColumn<T = IBaseTableRow> extends Column {
  __name__?: string
  label?: string | (() => string)
  field?: string
  prop?: string
  width?: string
  align?: string
  rowspan?: ({ row }: { row: T }) => number
  index?: number
  type?: string
  fixed?: string
  showOverflowTooltip?: boolean
  // render?: (args: HeadRenderArgs) => Element | boolean | number | string;
  render?: RenderFunctionString;
}

export interface IBaseTableRow {
  key?: string
  value?: unknown
  rowSpan?: number
}

export interface IQueryListParams {
  apiMethod: (...args: any[]) => Promise<unknown>;
  filterData?: Ref<Record<string, any>>;
  id?: number;
  resetPageOnFilterDataChange?: boolean;
  initialPagination?: Partial<IPagination>;
}

export interface IPublishSearch {
  filterData?: Ref<Record<string, any>>;
  diffGroupTotal?: Ref<IDiffGroup[]>;
  searchDoneFn?: () => void;
}
