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
* 消费者组(ConsumerGroup)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;
const ConsumerGroup = () => import(/* webpackChunkName: "ConsumerGroup" */ '@/views/consumer-group/consumer-group.vue');
const ConsumerGroupCreate = () => import(/* webpackChunkName: "ConsumerGroup" */ '@/views/consumer-group/create.vue');

export default [
  {
    path: 'consumer-group',
    name: 'consumer-group',
    component: ConsumerGroup,
    meta: {
      headerTitle: t('消费者组列表'),
      menuKey: 'Consumer Group',
    },
  },
  {
    path: 'consumer-group/create',
    name: 'consumer-group-create',
    component: ConsumerGroupCreate,
    meta: {
      headerTitle: t('创建消费者组'),
      menuKey: 'Consumer Group',
      showBack: true,
    },
  },
  {
    path: 'consumer-group/edit/:id',
    name: 'consumer-group-edit',
    component: ConsumerGroupCreate,
    meta: {
      headerTitle: t('编辑消费者组'),
      MenuKey: 'Consumer Group',
      showBack: true,
    },
  },
  {
    path: 'consumer-group/clone/:id',
    name: 'consumer-group-clone',
    component: ConsumerGroupCreate,
    meta: {
      headerTitle: t('克隆消费者组'),
      MenuKey: 'Consumer Group',
      showBack: true,
    },
  },
];
