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
* 导入导出(ImportExport)页面的路由配置
*  */

import i18n from '@/i18n';

const { t } = i18n.global;
const ImportExport = () => import(/* webpackChunkName: "ImportExport" */ '@/views/import-export/import-export.vue');
const Upload = () => import(/* webpackChunkName: "ImportExport" */ '@/views/import-export/upload.vue');

export default [
  {
    path: 'import-export',
    name: 'import-export',
    component: ImportExport,
    meta: {
      headerTitle: t('导入导出'),
      menuKey: 'ImportExport',
    },
  },
  {
    path: 'import-export/upload',
    name: 'import-export-upload',
    component: Upload,
    meta: {
      headerTitle: t('上传'),
      menuKey: 'ImportExport',
    },
  },
];
