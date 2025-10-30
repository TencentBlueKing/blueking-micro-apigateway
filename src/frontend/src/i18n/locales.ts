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

/* eslint-disable quote-props */

interface ILocales {
  [zhCnText: string]: string[];
}

const locales: ILocales = {
  '路由': ['Route'],
  '创建路由': ['Create Route'],
  '路由列表': ['Route List'],
  '服务': ['Service'],
  '上游': ['Upstream'],
  '微网关': ['Micro API Gateway'],
  '格式化': ['Format'],
  '保存': ['Save'],
  '取消': ['Cancel'],
  '复制': ['Copy'],
  '编辑器': ['Editor'],
  '新建': ['Create'],
  '查看': ['Check'],
  '插件组': ['Plugin Configuration'],
  '插件配置': ['Plugin Configuration'],
  '消费者组': ['Consumer Group'],
  '插件元数据': ['Plugin Metadata'],
  '网关同步源代码': ['Gateway Sync Data'],
  '描述': ['Description'],
  '更新': ['Update'],
  '操作': ['Action'],
  '服务列表': ['Services'],
  '创建服务': ['Create Service'],
  '消费者': ['Consumer'],
  '消费者列表': ['Consumers'],
  '创建消费者': ['Create Consumer'],
  '上游列表': ['Upstreams'],
  '创建上游': ['Create Upstream'],
  '插件组列表': ['Plugin Configurations'],
  '创建插件组': ['Create Plugin Configuration'],
  '创建插件配置': ['Create Plugin Configuration'],
  '插件元数据列表': ['Plugin Metadata'],
  '创建插件元数据': ['Create Plugin Metadata'],
  'Global Rules 列表': ['Global Rules'],
  '创建 Global Rules': ['Create Global Rules'],
  '创建消费者组': ['Create Consumer Group'],
  '网关同步资源': ['Gateway Sync Data'],
  '创建网关同步资源': ['Create Gateway Sync Data'],
  '技术支持': ['Support'],
  '社区论坛': ['Forum'],
  '产品官网': ['Official'],
  '名称': ['Name'],
  '更新时间': ['Update Time'],
  '批量操作': ['Multiple Action'],
  '类型': ['Type'],
  '消费者组ID': ['Consumer Group ID'],
  '消费者组列表': ['Consumer Groups'],
  '网关同步资源列表': ['Gateway Sync Data'],
  '暂无数据': ['No Data'],
  '网关列表': ['Gateways'],
  '手动填写': ['Config Now'],
  '重试次数': ['Retries'],
  '端口': ['Port'],
  '协议': ['Scheme'],
  '启用': ['Enable'],
  '插件': ['Plugins'],
  '已启用': ['Enabled'],
  '未启用': ['Disabled'],
  '已启用插件': ['Enabled Plugins'],
  '未启用插件': ['Disabled Plugins'],
  '搜索 {options}': ['Search {options}'],
  '已下线': ['Disabled'],
  '标签': ['Labels'],
  '状态': ['Status'],
  '草稿：已删除': ['Draft Deleted'],
  '草稿': ['Draft'],
  '草稿：已修改': ['Draft Modified'],
  '发布失败': ['Failed'],
  '发布中': ['Doing'],
  '发布成功': ['Success'],
  '服务名称：{name}': ['Service name: {name}'],
  '已选 {count} 条，': ['Selected {count} items, '],
  'stream 路由': ['Stream Route'],
  '自定义插件': ['Schema'],
};

export default locales;
