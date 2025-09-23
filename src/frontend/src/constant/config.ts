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

import index, { isChinese } from '@/i18n';
import Cookie from 'js-cookie';

const { t } = index.global;
// 当前年份
const curYear = (new Date()).getFullYear();
const {
  BK_LOGIN_URL,
  BK_LIST_USERS_API_URL,
  BK_API_RESOURCE_URL_TMPL,
  BK_DASHBOARD_FE_URL,
  BK_APIGATEWAY_VERSION,
  BK_ANALYSIS_SCRIPT_SRC,
  BK_DOCS_URL_PREFIX,
  BK_APP_MODE,
  BK_DEMO_INFO,
  BK_DEMO_DOC_URL,
} = window;
let { BK_DOCS_URL_PREFIX_MARKDOWN } = window;

// 根据文档前缀、语言和版本拼接 BK_DOCS_URL_PREFIX_MARKDOWN 的值
// 执行一个立即执行函数获取，避免了声明会暴露到全局的变量
(function () {
  const langMap: Record<string, string> = {
    'zh-cn': 'ZH',
    en: 'EN',
  };
  // 当前 cookie 中使用的语言 zh-cn | en
  const currentLang = Cookie.get('blueking_language') || 'zh-cn';
  // 获取文档语言 ZH | EN
  const lang = langMap[currentLang];
  // 获取当前版本的 major 和 minor 版本，如：1.13.1 -> 1.13
  const docVersion = (BK_APIGATEWAY_VERSION || '1.0.0').split('.')
    .slice(0, 2)
    .join('.');
  BK_DOCS_URL_PREFIX_MARKDOWN = `${BK_DOCS_URL_PREFIX}/markdown/${lang}/APIGateway/${docVersion}`;
}());

export default {
  // 登录
  BK_LOGIN_URL,

  // 访问统计script地址
  BK_ANALYSIS_SCRIPT_SRC,

  // 是否为 demo 模式，是的话值为 demo，否则为空
  BK_APP_MODE,
  // demo 模式下的信息和文档地址
  BK_DEMO_INFO,
  BK_DEMO_DOC_URL,

  // 助手
  HELPER: {
    name: '',
    href: '',
  },

  // 底部信息
  FOOT_INFO: {
    NAME: t('技术支持'),
    NAMEHREF: 'https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true',
    COMMUNITY: t('社区论坛'),
    COMMUNITYHREF: 'https://bk.tencent.com/s-mart/community/',
    PRODUCT: t('产品官网'),
    PRODUCTHREF: 'https://bk.tencent.com/index/',
    VERSION: BK_APIGATEWAY_VERSION,
  },

  // 人员列表接口地址，外部版本必填
  BK_LIST_USERS_API_URL,

  // 环境访问地址域名
  STAGE_DOMAIN: BK_API_RESOURCE_URL_TMPL,

  // 开源社区 or 加入圈子
  MARKER: 'https://github.com/TencentBlueKing/blueking-micro-apigateway',

  OA_DOMAIN: '',

  WOA_DOMAIN: '',

  IED_DOMAIN: '',

  // 网关管理
  APIGW: BK_DASHBOARD_FE_URL,

  // 旧版地址
  OLD_SITE_URL: '',

  // 常用工具
  TOOLS: '',

  // createChat api
  CREATE_CHAT_API: '',

  // sendChat api
  SEND_CHAT_API: '',

  PREV_URL: '/docs',

  DOC: {
    // “请求流水查询规则”
    QUERY_USE: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/Explanation/access-log.md`,

    // 蓝鲸用户认证
    USER_VERIFY: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/Explanation/authorization.md`,

    // API资源模板变量
    TEMPLATE_VARS: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/Explanation/template-var.md`,

    // 网关认证
    AUTH: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/Explanation/authorization.md`,

    // Swagger说明文档
    SWAGGER: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/HowTo/Connect/swagger-explain.md`,

    // 跨域资源共享(CORS)
    CORS: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/HowTo/Plugins/cors.md`,

    // 断路器
    BREAKER: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/HowTo/Plugins/circuit-breaker.md`,

    // 频率控制
    RATELIMIT: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/HowTo/Plugins/rate-limit.md`,

    // JWT
    JWT: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/Explanation/jwt.md`,

    // 用户类型
    USER_TYPE: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/README.md`,

    // API网关错误码
    ERROR_CODE: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/FAQ/error-response.md`,

    // 组件频率控制
    COMPONENT_RATE_LIMIT: `${BK_DOCS_URL_PREFIX_MARKDOWN}/component/reference/rate-limit.md`,

    // 如何开发和发布组件
    COMPONENT_CREATE_API: `${BK_DOCS_URL_PREFIX_MARKDOWN}/component/quickstart/create-api.md`,

    // 文档导入详情
    IMPORT_RESOURCE_DOCS: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/HowTo/Connect/manage-document.md`,

    // 实例类型
    INSTANCE_TYPE: `${BK_DOCS_URL_PREFIX_MARKDOWN}/`,

    // 调用API
    USER_API: `${BK_DOCS_URL_PREFIX_MARKDOWN}/UserGuide/HowTo/call-gateway-api.md`,

    // 升级到 1.13 的指引说明
    UPGRADE_TO_113_TIP: '',
  },

  // 网站默认配置
  SITE_CONFIG: {
    bkAppCode: 'bk_apigateway', // appcode
    name: 'API Gateway', // 站点的名称，通常显示在页面左上角，也会出现在网页title中
    nameEn: 'API Gateway', // 站点的名称-英文
    appLogo: isChinese ? '/static/images/APIgataway-c.png' : '/static/images/APIgataway-en.png', // 站点logo
    appLogoEn: '/static/images/APIgataway-en.png', // 站点logo
    favicon: '/static/images/favicon.png', // 站点favicon
    helperText: '联系 BK 助手',
    helperTextEn: 'Contact BK Assistant',
    helperLink: 'wxwork://message/?username=BK%E5%8A%A9%E6%89%8B',
    brandImg: '/static/images/brand.png',
    brandImgEn: '/static/images/brand.png',
    brandName: '蓝鲸智云', // 品牌名，会用于拼接在站点名称后面显示在网页title中
    brandNameEn: 'Tencent BlueKing', // 品牌名-英文
    footerInfo: '[技术支持](https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true) | [社区论坛](https://bk.tencent.com/s-mart/community/) | [产品官网](https://bk.tencent.com/index/)', // 页脚的内容，仅支持 a 的 markdown 内容格式
    footerInfoEn: '[Support](https://wpa1.qq.com/KziXGWJs?_type=wpa&qidian=true) | [Forum](https://bk.tencent.com/s-mart/community/) | [Official](https://bk.tencent.com/index/)', // 页脚的内容-英文
    footerCopyright: `Copyright © 2012-${curYear} Tencent BlueKing. All Rights Reserved. {version_placeholder}`, // 版本信息，包含变量，展示在页脚内容下方
  },
};
