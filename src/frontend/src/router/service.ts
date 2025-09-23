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
* 服务(Service)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;

const Service = () => import(/* webpackChunkName: "Service" */ '@/views/service/service.vue');
const ServiceCreate = () => import(/* webpackChunkName: "Service" */ '@/views/service/create.vue');

export default [
  {
    path: 'service',
    name: 'service',
    component: Service,
    meta: {
      headerTitle: t('服务列表'),
      menuKey: 'Service',
    },
  },
  {
    path: 'service/create',
    name: 'service-create',
    component: ServiceCreate,
    meta: {
      headerTitle: t('创建服务'),
      menuKey: 'Service',
      showBack: true,
    },
  },
  {
    path: 'service/edit/:id',
    name: 'service-edit',
    component: ServiceCreate,
    meta: {
      headerTitle: t('编辑服务'),
      menuKey: 'Service',
      showBack: true,
    },
  },
  {
    path: 'service/clone/:id',
    name: 'service-clone',
    component: ServiceCreate,
    meta: {
      headerTitle: t('克隆服务'),
      menuKey: 'Service',
      showBack: true,
    },
  },
];
