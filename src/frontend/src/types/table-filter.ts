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
import i18n from '@/i18n';

export type IFilterOption = {
  label?: string
  value?: string | number
  name?: string
  id?: string | number
};

// 表格filter增加全局无关联关系
export class FilterOptionClass {
  public displayKey?:  keyof IFilterOption;
  public displayValue?:  keyof IFilterOption;
  public filterOptions: IFilterOption[] = [];
  private extraOptions?: boolean | IFilterOption[];
  private initFilterOption: IFilterOption[];

  // options代表原列表， key、value对应显示键值， extraOption代表是否存在额外插入项
  constructor(params?: {
    options?: IFilterOption[],
    key?: keyof IFilterOption,
    value?: keyof IFilterOption,
    extra?: boolean | IFilterOption[]
  }) {
    const { key, value, options = [], extra } = params ?? {};
    this.displayKey = key || 'label';
    this.displayValue = value || 'value';
    this.filterOptions = options.map(item => ({
      [this.displayKey]: item[this.displayKey],
      [this.displayValue]: item[this.displayValue],
    }));
    // 如果有额外的option，如果是个数组代表是各模块自定义的数据，如果为true默认新增固定的option
    const isExistExtra = extra && (Array.isArray(extra) || typeof extra === 'boolean');
    if (isExistExtra) {
      this.extraOptions = extra;
      this.initFilterOption = [{
        [this.displayKey]: i18n.global.t('无关联关系'),
        [this.displayValue]: '--',
      }];
      this.filterOptions = Array.isArray(this.extraOptions)
        ? [...this.extraOptions, ...this.filterOptions]
        : [...this.initFilterOption, ...this.filterOptions];
    }
  }
}
