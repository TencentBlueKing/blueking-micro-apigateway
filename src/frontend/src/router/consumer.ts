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
* 消费者(Consumer)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;

const Consumer = () => import(/* webpackChunkName: "Consumer" */ '@/views/consumer/consumer.vue');
const ConsumerCreate = () => import(/* webpackChunkName: "Consumer" */ '@/views/consumer/create.vue');

export default [
  {
    path: 'consumer',
    name: 'consumer',
    component: Consumer,
    meta: {
      headerTitle: t('消费者列表'),
      menuKey: 'Consumer',
    },
  },
  {
    path: 'consumer/create',
    name: 'consumer-create',
    component: ConsumerCreate,
    meta: {
      headerTitle: t('创建消费者'),
      menuKey: 'Consumer',
      showBack: true,
    },
  },
  {
    path: 'consumer/edit/:id',
    name: 'consumer-edit',
    component: ConsumerCreate,
    meta: {
      headerTitle: t('编辑消费者'),
      menuKey: 'Consumer',
      showBack: true,
    },
  },
  {
    path: 'consumer/clone/:id',
    name: 'consumer-clone',
    component: ConsumerCreate,
    meta: {
      headerTitle: t('克隆消费者'),
      menuKey: 'Consumer',
      showBack: true,
    },
  },
];
