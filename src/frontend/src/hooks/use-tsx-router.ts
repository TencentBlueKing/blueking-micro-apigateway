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
* 使用 lang="tsx" 的组件直接 import 'vue-router' 引入的路由无法使用
* 需要在组件外封装一次，此 hook 即为实现此目的创建
* 现在在 TSX 组件中引入本 hook 就可正常使用 vue-router
*
* 引入：
* import useTsxRouter from './hooks/useTsxRouter';
*
* 使用：
* const { useRouter, onBeforeRouteLeave } = useTsxRouter();
* const router = useRouter();
*
*  */
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router';

export default function useTsxRouter() {
  return {
    useRoute,
    useRouter,
    onBeforeRouteLeave,
  };
};
