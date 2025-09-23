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
* 路由(Route)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;
const Route = () => import(/* webpackChunkName: "Route" */ '@/views/route/route.vue');
const RouteCreate = () => import(/* webpackChunkName: "Route" */ '@/views/route/create.vue');

export default [
  {
    path: 'route',
    name: 'route',
    component: Route,
    meta: {
      headerTitle: t('路由列表'),
      menuKey: 'Route',
    },
  },
  {
    path: 'route/create',
    name: 'route-create',
    component: RouteCreate,
    meta: {
      headerTitle: t('创建路由'),
      menuKey: 'Route',
      showBack: true,
    },
  },
  {
    path: 'route/edit/:id',
    name: 'route-edit',
    component: RouteCreate,
    meta: {
      headerTitle: t('编辑路由'),
      menuKey: 'Route',
      showBack: true,
    },
  },
  {
    path: 'route/clone/:id',
    name: 'route-clone',
    component: RouteCreate,
    meta: {
      headerTitle: t('克隆路由'),
      menuKey: 'Route',
      showBack: true,
    },
  },
];
