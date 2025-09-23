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
* 自定义插件(PluginCustom)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;
const PluginCustom = () => import(/* webpackChunkName: "PluginCustom" */ '@/views/plugin-custom/plugin-custom.vue');
const PluginCustomCreate = () => import(/* webpackChunkName: "PluginCustom" */ '@/views/plugin-custom/create.vue');

export default [
  {
    path: 'plugin-custom',
    name: 'plugin-custom',
    component: PluginCustom,
    meta: {
      headerTitle: t('自定义插件列表'),
      menuKey: 'Plugin Custom',
    },
  },
  {
    path: 'plugin-custom/create',
    name: 'plugin-custom-create',
    component: PluginCustomCreate,
    meta: {
      headerTitle: t('创建自定义插件'),
      menuKey: 'Plugin Custom',
      showBack: true,
    },
  },
  {
    path: 'plugin-custom/edit/:id',
    name: 'plugin-custom-edit',
    component: PluginCustomCreate,
    meta: {
      headerTitle: t('编辑自定义插件'),
      MenuKey: 'Plugin Custom',
      showBack: true,
    },
  },
];
