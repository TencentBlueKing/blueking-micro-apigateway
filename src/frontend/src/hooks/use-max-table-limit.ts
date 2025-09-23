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

// 计算当前视口高度能放几条表格数据
import { toValue } from 'vue';
import { useWindowSize } from '@vueuse/core';

export default function useMaxTableLimit(heightTaken = 347, lineHeight = 42) {
  const viewportHeight = toValue(useWindowSize().height);
  const heightToUse = viewportHeight - heightTaken;
  const limit = Math.floor(heightToUse / lineHeight);
  return limit > 0 ? limit : 1;
}
