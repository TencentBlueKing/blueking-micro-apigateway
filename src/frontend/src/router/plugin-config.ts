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
* 插件组(PluginConfig)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;
const PluginConfig = () => import(/* webpackChunkName: "PluginConfig" */ '@/views/plugin-config/plugin-config.vue');
const PluginList = () => import(/* webpackChunkName: "PluginConfig" */ '@/views/plugin-config/create.vue');

export default [
  {
    path: 'plugin-config',
    name: 'plugin-config',
    component: PluginConfig,
    meta: {
      headerTitle: t('插件组列表'),
      menuKey: 'Plugin Config',
    },
  },
  {
    path: 'plugin-config/create',
    name: 'plugin-config-create',
    component: PluginList,
    meta: {
      headerTitle: t('创建插件组'),
      menuKey: 'Plugin Config',
      showBack: true,
    },
  },
  {
    path: 'plugin-config/edit/:id',
    name: 'plugin-config-edit',
    component: PluginList,
    meta: {
      headerTitle: t('编辑插件组'),
      MenuKey: 'Plugin Config',
      showBack: true,
    },
  },
];
