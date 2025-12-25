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

<template>
  <bk-config-provider :locale="bkuiLocale">
    <div id="app" :class="[systemCls]">
      <!--  demo 模式的顶部横幅  -->
      <div v-if="constantConfig.BK_APP_MODE === 'demo'" class="demo-mode-banner">
        <bk-alert
          theme="warning"
        >
          <template #title>
            <span>{{ constantConfig.BK_DEMO_INFO }}</span>
            <bk-link :href="constantConfig.BK_DEMO_DOC_URL" style="font-size: 12px;" target="_blank" theme="primary">
              {{ t('文档') }}
            </bk-link>
          </template>
        </bk-alert>
      </div>
      <!--      <notice-component-->
      <!--        v-if="showNoticeAlert && enableShowNotice"-->
      <!--        :api-url="noticeApi"-->
      <!--        @show-alert-change="handleShowAlertChange"-->
      <!--      />-->
      <bk-navigation
        :default-open="true"
        :need-menu="false"
        class="navigation-content"
        :class="{ 'has-demo-mode-banner': constantConfig.BK_APP_MODE === 'demo' }"
        navigation-type="top-bottom"
      >
        <template #side-header>
          <div
            class="flex-row align-items-center"
            @click="handleToHome"
          >
            <div>
              <img
                :src="appLogo"
                alt="API GATEWAY"
                class="api-logo"
              >
            </div>
            <div class="side-title">
              {{ sideTitle }}
            </div>
          </div>
        </template>
        <div class="content">
          <router-view></router-view>
        </div>
        <template #header>
          <div class="header">
            <div class="header-nav">
              <template v-for="(item, index) in headerList">
                <div
                  v-if="item.enabled"
                  :key="item.id"
                  :class="{ 'item-active': index === activeIndex }"
                  class="header-nav-item"
                >
                  <span
                    v-if="!isExternalLink(item.url)"
                    @click="handleToPage(item)"
                  >{{ item.name }}</span>
                  <a v-else :href="item.url" target="_blank">{{ item.name }}</a>
                </div>
              </template>
              <!-- 内部上云版的共享网关外链 -->
              <div
                v-if="envStore.edition === 'te'"
                class="header-nav-item"
              >
                <a :href="envStore.links.bk_apigateway_link" target="_blank">{{ t('共享网关') }}</a>
              </div>
            </div>
            <div class="header-aside-wrap">
              <!--              <language-toggle></language-toggle>-->
              <product-info></product-info>
              <user-info v-if="userLoaded" />
            </div>
          </div>
        </template>
      </bk-navigation>
    </div>
  </bk-config-provider>
</template>

<script lang="ts" setup>
import { ConfigProvider as BkConfigProvider, Message } from 'bkui-vue';
// @ts-ignore
import zhCn from 'bkui-vue/dist/locale/zh-cn.esm';
// @ts-ignore
import en from 'bkui-vue/dist/locale/en.esm';

// import NoticeComponent from '@blueking/notice-component';
// import '@blueking/notice-component/dist/style.css';
import UserInfo from '@/components/user-info.vue';
import ProductInfo from '@/components/product-info.vue';
import { useCommon, useUser, useEnv } from '@/store';
import { getUser } from '@/http';
import { getPlatformConfig, setDocumentTitle, setShortcutIcon } from '@blueking/platform-config';
import logoWithoutName from '@/images/APIgateway-logo.png';
import { isChinese } from '@/i18n';
import constantConfig from '@/constant/config';
import { useScriptTag } from '@vueuse/core';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, watch } from 'vue';

interface IHeaderItem {
  name: string
  id: number
  url: string
  enabled: boolean
  link: string
  routeName: string
}

const { t, locale } = useI18n();
const router = useRouter();
const route = useRoute();
const common = useCommon();
const user = useUser();
const envStore = useEnv();

// const { BK_DASHBOARD_URL } = window;

// 接入访问统计逻辑，只在上云版执行
if (constantConfig.BK_ANALYSIS_SCRIPT_SRC) {
  try {
    const { BK_ANALYSIS_SCRIPT_SRC } = constantConfig;
    if (BK_ANALYSIS_SCRIPT_SRC) {
      useScriptTag(
        BK_ANALYSIS_SCRIPT_SRC,
        // script loaded 后的回调
        () => {
          window.BKANALYSIS.init({ siteName: '' });
        },
        // script 标签的 attrs
        {
          attrs: { charset: 'utf-8' },
        },
      );
    }
  } catch {
    console.log('BKANALYSIS init fail');
  }
}

const bkuiLocaleData = {
  zhCn,
  en,
};

const bkuiLocale = computed(() => {
  if (locale.value === 'zh-cn') {
    return bkuiLocaleData.zhCn;
  }
  return bkuiLocaleData.en;
});

// t()方法的 | 符号有特殊含义，需要用插值才能正确显示
const sideTitle = computed(() => {
  // return t("蓝鲸 {pipe} API 网关", { pipe: '|'});
  return t('蓝鲸微网关');
});

const websiteConfig = ref<any>({});

const getWebsiteConfig = async () => {
  const bkSharedResUrl = window.BK_NODE_ENV === 'development'
    ? window.BK_DASHBOARD_FE_URL
    : window.BK_SHARED_RES_URL;

  if (bkSharedResUrl) {
    const url = bkSharedResUrl?.endsWith('/') ? bkSharedResUrl : `${bkSharedResUrl}/`;
    websiteConfig.value = await getPlatformConfig(`${url}${window.BK_APP_CODE || 'bk_apigateway'}/base.js`, constantConfig.SITE_CONFIG);
  } else {
    websiteConfig.value = await getPlatformConfig(constantConfig.SITE_CONFIG);
  }

  if (websiteConfig.value.i18n) {
    websiteConfig.value.i18n.appLogo = websiteConfig.value[isChinese ? 'appLogo' : 'appLogoEn'];
  }

  setShortcutIcon(websiteConfig.value?.favicon);
  setDocumentTitle(websiteConfig.value?.i18n);
  common.setWebsiteConfig(websiteConfig.value);
};
getWebsiteConfig();

const appLogo = computed(() => {
  return logoWithoutName;
});

// 加载完用户数据才会展示页面
const userLoaded = ref(false);
const activeIndex = ref(0);

// 跑马灯数据
// const showNoticeAlert = ref(true);
// const enableShowNotice = ref(false);
// const noticeApi = ref(`${BK_DASHBOARD_URL}/notice/announcements/`);

const headerList = computed<IHeaderItem[]>(() => ([
  {
    name: t('我的网关'),
    id: 1,
    url: '/',
    enabled: true,
    link: '',
    routeName: 'root',
  },
]));

const systemCls = ref('mac');
// const authRef = ref();

// const apigwId = computed(() => {
//   if (route.params.id !== undefined) {
//     return route.params.id;
//   }
//   return undefined;
// });

// const handleShowAlertChange = (payload: boolean) => {
//   showNoticeAlert.value = payload;
// };

watch(
  () => route.fullPath,
  async () => {
    // const { meta } = route;
    // let index = 0;
    // for (let i = 0; i < headerList.value.length; i++) {
    //   const item = headerList.value[i];
    //   if (item.url === meta?.topMenu) {
    //     index = i;
    //     break;
    //   }
    // }
    // activeIndex.value = index;
    const platform = window.navigator.platform.toLowerCase();
    if (platform.indexOf('win') === 0) {
      systemCls.value = 'win';
    }

    try {
      // const [useRes, flagsRes] = await Promise.all([
      //   getUser(),
      //   getFeatureFlags({ limit: 10000, offset: 0 }),
      // ]);
      const useRes = await getUser();

      user.setUser(useRes);

      // enableShowNotice.value = flagsRes?.ENABLE_BK_NOTICE || false;
      // user.setFeatureFlags(flagsRes);

      userLoaded.value = true;
    } catch (e: any) {
      console.error(e);
      if (e?.code !== 'Unauthorized') {
        Message('获取用户信息或功能权限失败，请检查后再试');
      }
    }
  },
  {
    immediate: true,
    deep: true,
  },
);

const isExternalLink = (url?: string) => /^https?:\/\//.test(url);

const handleToPage = (headerItem: IHeaderItem) => {
  router.push({ name: headerItem.routeName });
};

const handleToHome = () => {
  router.push({ name: 'root' });
};

common.setEnums();
envStore.setVars();
</script>

<style lang="scss">
@use "style/index";
// class 带 .form-element 的表单样式
@use "@/style/form-element";
// 自定义滚动条样式
@use "@/style/scroll-bar";

.bk-code-diff .d2h-code-side-linenumber {
  padding-right: 16px !important;
}
</style>

<style lang="scss" scoped>

// demo 模式的顶部横幅样式
.demo-mode-banner {
  padding-left: 24px;
  background-color: #fff4e2;

  .bk-alert {
    border: none;

    :deep(.bk-alert-wraper) {
      padding-block: 12px;
    }
  }
}

.navigation-content {
  // 页面中有 demo 模式横幅时，应减去横幅高度
  &.has-demo-mode-banner {
    height: calc(100vh - 40px);
  }

  :deep(.bk-navigation-wrapper) {
    .container-content {
      min-width: 1020px;
      // 最小宽度应为 1280px 减去左侧菜单栏展开时的宽度 260px，即为 1020px
      padding: 0 !important;
    }
  }

  .content {
    font-size: 14px;
    height: 100%;
  }

  :deep(.title-desc) {
    cursor: pointer;
    color: #eaebf0;
  }

  .api-logo {
    height: 28px;
    cursor: pointer;
    margin-top: 4px;
  }
  .side-title {
    font-weight: Bold;
    font-size: 16px;
    color: #EAEBF0;
    margin-left: 6px;
    cursor: pointer;
  }

  .header {
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    color: #96a2b9;

    .header-nav {
      display: flex;
      flex: 1;
      margin: 0;
      padding: 0;

      &-item {
        margin-right: 40px;
        list-style: none;
        color: #96a2b9;

        &.item-active {
          color: #ffffff !important;
        }

        &:hover {
          cursor: pointer;
          color: #d3d9e4;
        }

        a {
          color: #96a2b9;

          &:hover {
            color: #d3d9e4;
          }
        }
      }
    }

    .header-aside-wrap {
      display: flex;
      align-items: center;
      gap: 14px;
    }
  }
}

</style>
