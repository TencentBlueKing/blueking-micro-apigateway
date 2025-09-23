declare interface Window {
  SITE_URL: string,
  BK_DOMAIN: string
  BK_DASHBOARD_URL: string
  BK_LOGIN_URL: string
  BK_DASHBOARD_CSRF_COOKIE_NAME: string
  BK_LIST_USERS_API_URL: string
  BK_API_RESOURCE_URL_TMPL: string
  BK_DASHBOARD_FE_URL: string
  BK_DOCS_URL_PREFIX: string
  BK_DOCS_URL_PREFIX_MARKDOWN: string
  BK_APIGATEWAY_VERSION: string
  GLOBAL_CONFIG: any
  BK_COMPONENT_API_URL: string
  BK_APP_VERSION: string
  CREATE_CHAT_API: string
  SEND_CHAT_API: string
  BK_SHARED_RES_URL: string
  BK_APP_CODE: string
  BK_NODE_ENV: string
  BK_ANALYSIS_SCRIPT_SRC: string
  BKANALYSIS?: {
    [key: string]: any
  }
  BK_APP_MODE: string
  BK_DEMO_INFO: string
  BK_DEMO_DOC_URL: string
}

declare module '*.svg' {
  const content: any;
  export default content;
}

declare module '*.png';

declare module '@blueking/login-modal';
// declare module '@blueking/notice-component';
declare module '@blueking/platform-config';
declare module '@blueking/release-note';
