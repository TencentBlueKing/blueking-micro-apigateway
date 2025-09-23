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
  <div class="navigation-main">
    <bk-navigation
      class="navigation-main-content apigw-navigation"
      default-open
      need-menu
    >
      <template #menu>
        <bk-menu
          :active-key="activeMenuKey"
          :collapse="collapse"
          unique-open
        >
          <bk-menu-group v-for="menuGroup in menuGroups" :key="menuGroup.name" :name="menuGroup.name">
            <bk-menu-item
              v-for="menu in menuGroup.menus"
              :key="menu.name"
              @click="handleMenuItemClick(menu)"
            >
              <template #icon>
                <i :class="['icon apigateway-icon', `icon-ag-${menu.icon}`]"></i>
              </template>
              <div>{{ menu.title }}</div>
            </bk-menu-item>
          </bk-menu-group>
        </bk-menu>
      </template>
      <template #side-header>
        <div class="side-header-wrapper">
          <bk-select
            ref="apigwSelect"
            v-model="gatewayId"
            :clearable="false"
            :popover-min-width="230"
            class="header-select"
            filterable
            @change="changeGateway"
          >
            <template #prefix>
              <div
                v-if="common.curGatewayData?.read_only" class="gateway-select-readonly-tag"
                style="margin-top: 5px;margin-left: 5px;"
              >
                {{ t('只读') }}
              </div>
            </template>
            <bk-option
              v-for="item in gatewaysList"
              :id="item.id"
              :key="item.id"
              :name="item.name"
            >
              <div class="gateway-select-option-item">
                <div>{{ item.name }}</div>
                <div v-if="item.read_only" class="gateway-select-readonly-tag">
                  {{ t('只读') }}
                </div>
              </div>
            </bk-option>
          </bk-select>
          <!-- <button-icon
            v-bk-tooltips="{ content: t('发布所有资源'), placement: 'bottom' }"
            icon="publish-fill"
            icon-color="white"
            style="width: 60px"
            theme="primary"
            @click="handlePublishAll"
          >{{
            t('发布')
          }}
          </button-icon> -->
        </div>
      </template>
      <div class="content-view">
        <bk-alert
          v-if="common.curGatewayData?.read_only"
          theme="warning"
        >
          {{
            t('当前网关处于只读模式，不支持任何修改操作。如需进行操作，请先在“基本信息”中关闭只读模式，注意，关闭后变更将可以发布到生产环境，请谨慎操作！')
          }}
        </bk-alert>
        <!-- 默认头部 -->
        <div class="content-header">
          <header class="content-header-title-wrapper">
            <div v-if="route.meta.showBack" class="content-header-title" @click="handleBack">
              <i class="icon apigateway-icon icon-ag-return-small"></i>
              {{ headerTitle }}
            </div>
            <div v-else class="content-header-title">
              <span>{{ headerTitle }}</span>
              <span v-if="headerIntro" class="intro-toggle" @click="toggleHeaderResourceIntro">
                <Icon
                  v-if="!showHeaderResourceIntro"
                  v-bk-tooltips="{ content: t('查看资源介绍') }"
                  color="#3a84ff" name="info" size="15"
                />
                <button-icon v-else icon="''" text theme="primary">
                  {{ t('收起') }}
                </button-icon>
              </span>
            </div>
            <div v-if="route.meta.showPageName && pageName" class="title-name">
              <span></span>
              <div class="name">{{ pageName }}</div>
            </div>
          </header>
          <main v-if="showHeaderResourceIntro && headerIntro" class="content-header-resource-intro">
            <div>
              <span class="intro-text">{{ headerIntro.text }}</span>
              <bk-link :href="headerIntro.docLink" v-if="headerIntro.docLink" target="_blank" theme="primary">
                <icon color="#3a84ff" name="jump" />
                {{ t('文档') }}
              </bk-link>
            </div>
          </main>
        </div>
        <div :class="route.meta.customHeader ? 'custom-header-view' : 'default-header-view'" class="home-view-wrapper">
          <router-view :key="homeViewKey"></router-view>
        </div>
      </div>
    </bk-navigation>
    <!-- <dialog-publish-resource
      v-model="isDialogPublishResourceShow"
      :diff-group-list="diffGroupList"
      @confirm="handlePublishConfirm"
      @refresh="handleDiffGroupRefresh"
    /> -->
  </div>
</template>

<script lang="ts" setup>
import { IGatewayItem, IMenu, IMenuGroup } from '@/types';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, watch } from 'vue';
import { useCommon } from '@/store';
import { useGetApiList } from '@/hooks';
import { getGatewaysDetail } from '@/http';
import { RESOURCE_INTRODUCTION } from '@/enum';
import Icon from '@/components/icon.vue';
import ButtonIcon from '@/components/button-icon.vue';
import { useStorage } from '@vueuse/core';
import { uniqueId } from 'lodash-es';
// import DialogPublishResource from '@/components/dialog-publish-resource.vue';
// import { getDiffAll, IDiffGroup, publishAll } from '@/http/publish';
// import { Message } from 'bkui-vue';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const common = useCommon();

const {
  getGatewaysListData,
} = useGetApiList({ name: '' });

const homeViewKey = ref(uniqueId());
const menuGroups = ref<IMenuGroup[]>([
  {
    name: t('基础资源'),
    menus: [
      {
        title: t('路由'),
        name: 'Route',
        routeName: 'route',
        icon: 'luyou',
        enabled: true,
      },
      {
        title: t('stream 路由'),
        name: 'Stream Route',
        routeName: 'stream-route',
        icon: 'stream-result',
        enabled: true,
      },
      {
        title: t('服务'),
        name: 'Service',
        routeName: 'service',
        icon: 'fuwu-2',
        enabled: true,
      },
      {
        title: t('上游'),
        name: 'Upstream',
        routeName: 'upstream',
        icon: 'micro-upstream',
        enabled: true,
      },
      {
        title: t('proto'),
        name: 'Proto',
        routeName: 'proto',
        icon: 'keguancexing',
        enabled: true,
      },
      {
        title: t('证书'),
        name: 'SSL',
        routeName: 'ssl',
        icon: 'cert-fill',
        enabled: true,
      },
    ],
  },
  {
    name: t('消费者'),
    menus: [
      {
        title: t('消费者'),
        name: 'Consumer',
        routeName: 'consumer',
        icon: 'cc-user',
        enabled: true,
      },
      {
        title: t('消费者组'),
        name: 'Consumer Group',
        routeName: 'consumer-group',
        icon: 'usergroup',
        enabled: true,
      },
    ],
  },
  {
    name: t('插件'),
    menus: [
      {
        title: t('插件元数据'),
        name: 'Plugin Metadata',
        routeName: 'plugin-metadata',
        icon: 'micro-plugin',
        enabled: true,
      },
      {
        title: t('全局规则'),
        name: 'Global Rules',
        routeName: 'global-rules',
        icon: 'system-mgr',
        enabled: true,
      },
      {
        title: t('插件组'),
        name: 'Plugin Config',
        routeName: 'plugin-config',
        icon: 'plugin-generic-fill',
        enabled: true,
      },
      {
        title: t('自定义插件'),
        name: 'Plugin Custom',
        routeName: 'plugin-custom',
        icon: 'micro-service',
        enabled: true,
      },
    ],
  },
  {
    name: t('发布'),
    menus: [
      {
        title: t('发布管理'),
        name: 'Publish',
        routeName: 'publish',
        icon: 'publish-fill',
        enabled: true,
      },
    ],
  },
  {
    name: t('同步'),
    menus: [
      {
        title: t('etcd 资源列表'),
        name: 'Gateway Sync Data',
        routeName: 'gateway-sync-data',
        icon: 'ziyuanguanli',
        enabled: true,
      },
      {
        title: t('导入导出'),
        name: 'ImportExport',
        routeName: 'import-export',
        icon: 'import-export',
        enabled: true,
      },
    ],
  },
  {
    name: t('其他'),
    menus: [
      {
        title: t('审计日志'),
        name: 'Audit',
        routeName: 'audit',
        icon: 'audit1-fill',
        enabled: true,
      },
      {
        title: t('基本信息'),
        name: 'Basic Info',
        routeName: 'basic-info',
        icon: 'doc-mgr',
        enabled: true,
      },
    ],
  },
]);
const collapse = ref(false);
const pageName = ref('');
const gatewayId = ref<number>(common.gatewayId || 0);
const gatewaysList = ref<IGatewayItem[]>([]);
// const isDialogPublishResourceShow = ref(false);
// const diffGroupList = ref<IDiffGroup[]>([]);
const showHeaderResourceIntro = useStorage('show_header_resource_intro', true);

// 页面header名
const headerTitle = computed(() => {
  return route.meta?.headerTitle || '';
});

// 页面资源介绍
const headerIntro = computed(() => {
  return RESOURCE_INTRODUCTION[route.name as string] || null;
});

// 选中的菜单
const activeMenuKey = computed(() => {
  return route.meta?.menuKey || route.name || '';
});

const setGatewayDetail = async () => {
  const curGatewayData = await getGatewaysDetail(gatewayId.value);
  common.setCurGatewayData(curGatewayData);
};

watch(() => route.params.gatewayId, async () => {
  const _gatewayId = Number(route.params.gatewayId as unknown);
  gatewayId.value = _gatewayId;
  common.setGatewayId(_gatewayId);
  await setGatewayDetail();
}, { immediate: true });

const handleMenuItemClick = (menu: IMenu) => {
  router.push({ name: menu.routeName });
};

const handleBack = () => {
  // 历史记录有上一层，正常返回
  if (history.length > 1) {
    router.back();
  } else {
    // 历史记录里没有记录时，返回资源对应的列表页
    if (route.params?.id) {
      const pathArr = route.path.split('/');
      const resourceType = pathArr[3];
      router.replace({ name: resourceType || 'root' });
    } else {
      router.replace({ name: 'root' });
    }
  }
};

const toggleHeaderResourceIntro = () => {
  showHeaderResourceIntro.value = !showHeaderResourceIntro.value;
};

const getGateways = async () => {
  gatewaysList.value = await getGatewaysListData();
};

const changeGateway = async () => {
  const gatewayObj = gatewaysList.value.find((item: IGatewayItem) => item.id === gatewayId.value);

  common.setGatewayId(gatewayId.value);
  common.setGatewayName(gatewayObj?.name);

  await setGatewayDetail();

  const { path } = route;
  router.replace('/refresh').then(() => {
    const pathArray = path.split('/');
    pathArray[2] = String(gatewayId.value);
    router.push(pathArray.join('/'));
  });
};

// const handlePublishAll = async () => {
//   await getDiffGroupList();
//   isDialogPublishResourceShow.value = true;
// };

// const getDiffGroupList = async () => {
//   const res = await getDiffAll();
//   diffGroupList.value = res || [];
// };

// const handlePublishConfirm = async () => {
//   await publishAll();

//   Message({
//     theme: 'success',
//     message: t('发布成功'),
//   });

//   refreshView();
// };

// const handleDiffGroupRefresh = async () => {
//   await getDiffGroupList();
//   refreshView();
// };

// const refreshView = () => {
//   homeViewKey.value = uniqueId();
// };

getGateways();

// 获取插件列表，写到全局状态里方便复用
common.setPlugins();

</script>

<style lang="scss" scoped>
.navigation-main {
  //height: calc(100vh - 52px);
  height: 100%;

  :deep(.navigation-nav) {
    .nav-slider {
      border-right: 1px solid #dcdee5 !important;
      background: #ffffff !important;

      .bk-navigation-title {
        flex-basis: 51px !important;
      }

      .nav-slider-list {
        border-top: 1px solid #f0f1f5;
      }
    }

    .footer-icon {
      &.is-left {
        color: #63656e;

        &:hover {
          cursor: pointer;
          color: #63656e;
          background: linear-gradient(270deg, #dee0ea, #eaecf2);
        }
      }
    }

    .bk-menu {
      background: #ffffff !important;

      .bk-menu-item {
        margin: 0;
        color: rgb(99, 101, 110);

        .item-icon {
          .default-icon {
            background-color: rgb(197, 199, 205);
          }
        }

        &:hover {
          background: #f0f1f5;
        }
      }

      .bk-menu-item.is-active {
        color: rgb(58, 132, 255);
        background: rgb(225, 236, 255);

        .item-icon {
          .default-icon {
            background-color: rgb(58, 132, 255);
          }
        }
      }
    }

    .submenu-header {
      position: relative;

      .bk-badge-main {
        position: absolute;
        top: 1px;
        left: 120px;

        .bk-badge {
          width: 6px;
          min-width: 6px;
          height: 6px;
          background-color: #ff5656;
        }
      }
    }

    .submenu-list {
      .bk-menu-item {
        .item-content {
          position: relative;

          .bk-badge-main {
            position: absolute;
            top: 6px;
            left: 56px;

            .bk-badge {
              font-size: 12px;
              line-height: 14px;
              min-width: 18px;
              height: 18px;
              padding: 0 2px;
              background-color: #ff5656;
            }
          }
        }
      }
    }

    .submenu-header-icon {
      color: rgb(99, 101, 110);
    }

    .submenu-header-content {
      color: rgb(99, 101, 110);
    }

    .submenu-header-collapse {
      font-size: 22px;
      width: 22px;
    }

    .bk-menu-submenu.is-opened {
      background: #ffffff !important;
    }

    .bk-menu-submenu .submenu-header.is-collapse {
      color: rgb(58, 132, 255);
      background: rgb(225, 236, 255);

      .submenu-header-icon {
        color: rgb(58, 132, 255);
      }
    }
  }

  :deep(.navigation-container) {
    .container-header {
      flex-basis: 0 !important;
      height: 0 !important;
      border-bottom: 0;
    }
  }

  .navigation-main-content {
    border: 1px solid #dddddd;

    .content-view {
      font-size: 14px;
      overflow: hidden;
      height: 100%;

      .content-header {
        font-size: 16px;
        box-sizing: border-box;
        padding: 0 24px;
        color: #313238;
        border-bottom: 1px solid #dcdee5;
        background: #ffffff;
        box-shadow: 0 3px 4px 0 #0000000a;

        .content-header-title-wrapper {
          display: flex;
          align-items: center;
          flex-basis: 52px;
          box-sizing: border-box;
          height: 51px;
          margin-right: auto;

          .content-header-title {
            display: flex;
            align-items: center;
            cursor: pointer;

            .intro-toggle {
              font-size: 12px;
              display: flex;
              align-items: center;
              margin-left: 6px;
            }
          }

          .icon-ag-return-small {
            font-size: 32px;
            color: #3a84ff;
          }

          .title-name {
            display: flex;
            align-items: center;
            margin-left: 8px;

            span {
              width: 1px;
              height: 14px;
              margin-right: 8px;
              background: #dcdee5;
            }

            .name {
              font-size: 14px;
              color: #979ba5;
            }
          }
        }

        .content-header-resource-intro {
          font-size: 13px;
          padding-bottom: 12px;

          .intro-text {
            color: #63656e;
          }
        }
      }

      .default-header-view {
        overflow: auto;
        //height: calc(100vh - 105px);
        height: calc(100% - 52px);
      }

      .custom-header-view {
        overflow: auto;
        height: 100%;
        margin-top: 52px;
      }
    }

    .side-header-wrapper {
      width: 100%;
    }
  }

  :deep(.header-select) {
    //width: 224px;

    .bk-input {
      border: none;
      border-radius: 2px;
      background: #f5f7fa;
      box-shadow: none;

      .bk-input--text {
        font-size: 14px;
        color: #63656e;
        background: #f5f7fa;
      }
    }

    &.is-focus {
      border: 1px solid #3a84ff;
    }
  }

  .home-view-wrapper {
    position: relative;
  }
}
</style>
<style lang="scss">
.custom-height-navigation {
  .content-header {
    border-bottom: none !important;
  }
}

.custom-side-header {
  display: flex;
  align-items: center;

  .title {
    font-size: 16px;
    font-weight: 400;
    color: #313238;
  }

  .subtitle {
    font-size: 14px;
    color: #979ba5;
  }

  span {
    width: 1px;
    height: 14px;
    margin: 0 8px;
    background: #dcdee5;
  }
}

// 网关 select 在菜单折叠时的 tooltips
.bk-popper {
  width: auto !important;
}

.gateway-select-option-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.gateway-select-readonly-tag {
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 41px;
  height: 22px;
  color: #4d4f56;
  border-radius: 2px;
  background: #dcdee5;
}

</style>
