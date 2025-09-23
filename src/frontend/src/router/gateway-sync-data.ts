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
* 网关同步资源(GatewaySyncData)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;
const GatewaySyncData = () => import(/* webpackChunkName: "GatewaySyncData" */ '@/views/gateway-sync-data/gateway-sync-data.vue');
const GatewaySyncDataCreate = () => import(/* webpackChunkName: "GatewaySyncData" */ '@/views/gateway-sync-data/create.vue');

export default [
  {
    path: 'gateway-sync-data',
    name: 'gateway-sync-data',
    component: GatewaySyncData,
    meta: {
      headerTitle: t('etcd 资源列表'),
      menuKey: 'Gateway Sync Data',
    },
  },
  {
    path: 'gateway-sync-data/create',
    name: 'gateway-sync-data-create',
    component: GatewaySyncDataCreate,
    meta: {
      headerTitle: t('创建 etcd 资源'),
      menuKey: 'Gateway Sync Data',
      showBack: true,
    },
  },
];
