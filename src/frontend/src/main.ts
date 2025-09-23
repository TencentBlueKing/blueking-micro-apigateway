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

import { createApp } from 'vue';
import { createPinia } from 'pinia';
import router from './router';
import App from './app.vue';
import VueDOMPurifyHTML from 'vue-dompurify-html';
import { clickOutsideDirective } from '@/directive';
import globalConfig from '@/constant/config';
import i18n from '@/i18n';
import '@/style/index.scss';

// 全量引入 bkui-vue
import bkui, { bkTooltips } from 'bkui-vue';
// 全量引入 bkui-vue 样式
import 'bkui-vue/dist/style.css';
// iconfont 图标
import './assets/iconfont/style.css';
// tdesign 表格样式
import '@blueking/tdesign-ui/vue3/index.css';

const app = createApp(App);
app.config.globalProperties.GLOBAL_CONFIG = globalConfig;

app.use(router)
  .use(createPinia())
  .use(i18n)
  .use(bkui)
  .use(VueDOMPurifyHTML)
  .directive('bk-tooltips', bkTooltips)
  .directive('click-outside', clickOutsideDirective)
  .mount('.app');
