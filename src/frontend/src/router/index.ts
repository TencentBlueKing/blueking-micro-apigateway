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

import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router';
import routeRoutes from './route';
import streamRoute from '@/router/stream-route';
import serviceRoutes from './service';
import upstreamRoutes from './upstream';
import protoRoutes from './proto';
import sslRoutes from '@/router/ssl';
import consumerRoutes from './consumer';
import consumerGroupRoutes from '@/router/consumer-group';
import pluginMetadataRoutes from './plugin-metadata';
import globalRulesRoutes from './global-rules';
import pluginConfigRoutes from './plugin-config';
import pluginCustomRoutes from './plugin-custom';
import publishRoutes from '@/router/publish';
import gatewaySyncDataRoutes from '@/router/gateway-sync-data';
import importExportRoutes from '@/router/import-export';
import auditRoutes from '@/router/audit';
import basicInfoRoutes from '@/router/basic-info';

const Gateway = () => import(/* webpackChunkName: "Gateway" */ '@/views/gateway/gateway.vue');

const Home = () => import(/* webpackChunkName: "Home" */ '@/views/home.vue');

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'root',
    component: Gateway,
    // redirect: '/gateway',
  },
  {
    path: '/gateway/:gatewayId',
    name: 'gateway',
    component: Home,
    // redirect: 'route',
    children: [
      ...routeRoutes,
      ...streamRoute,
      ...serviceRoutes,
      ...upstreamRoutes,
      ...protoRoutes,
      ...sslRoutes,
      ...consumerRoutes,
      ...consumerGroupRoutes,
      ...pluginMetadataRoutes,
      ...globalRulesRoutes,
      ...pluginConfigRoutes,
      ...pluginCustomRoutes,
      ...gatewaySyncDataRoutes,
      ...importExportRoutes,
      ...publishRoutes,
      ...auditRoutes,
      ...basicInfoRoutes,
    ],
  },
];

const router = createRouter({
  history: createWebHistory(window.SITE_URL),
  routes,
});

export default router;
