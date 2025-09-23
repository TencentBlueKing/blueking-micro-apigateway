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
const StreamRoute = () => import('@/views/stream-route/stream-route.vue');
const StreamRouteCreate = () => import('@/views/stream-route/create.vue');

export default [
  {
    path: 'stream-route',
    name: 'stream-route',
    component: StreamRoute,
    meta: {
      headerTitle: t('stream 路由列表'),
      menuKey: 'Stream Route',
    },
  },
  {
    path: 'stream-route/create',
    name: 'stream-route-create',
    component: StreamRouteCreate,
    meta: {
      headerTitle: t('创建 stream 路由'),
      menuKey: 'Stream Route',
      showBack: true,
    },
  },
  {
    path: 'stream-route/edit/:id',
    name: 'stream-route-edit',
    component: StreamRouteCreate,
    meta: {
      headerTitle: t('编辑 stream 路由'),
      menuKey: 'Stream Route',
      showBack: true,
    },
  },
  {
    path: 'stream-route/clone/:id',
    name: 'stream-route-clone',
    component: StreamRouteCreate,
    meta: {
      headerTitle: t('克隆 stream 路由'),
      menuKey: 'Stream Route',
      showBack: true,
    },
  },
];
