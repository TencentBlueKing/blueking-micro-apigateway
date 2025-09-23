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
  <div class="home-container">
    <div class="title-container">
      <div class="title-total">
        {{ t('我的网关') }} ({{ showGateways.length }})
      </div>
      <div class="flex-row justify-content-between">
        <div class="flex-1 left">
          <bk-button
            theme="primary"
            @click="showAddDialog"
          >
            <plus class="f22" />
            {{ t('新建网关') }}
          </bk-button>
        </div>
        <div class="flex-1 flex-row">
          <bk-select
            v-model="filterKey"
            :clearable="false"
            :filterable="false"
            trigger="manual"
            ref="selectRef"
            class="select-cls"
            @change="handleChange"
          >
            <template #trigger="{ selected }">
              <div class="select-trigger flex-row align-items-center justify-content-between">
                <span class="label" @click.stop="handleToggle">{{ selected[0]?.label }}</span>
                <div class="suffix-cls flex-row align-items-center">
                  <span class="line"></span>
                  <div :class="{ 'icon-container': true, 'active': sort === 'asc' }" @click.stop="handleDirection">
                    <i class="icon apigateway-icon icon-ag-jiangxu"></i>
                  </div>
                </div>
              </div>
            </template>
            <bk-option
              class="custom-select-option"
              v-for="(item, index) in filterData"
              :key="index"
              :value="item.value"
              :label="item.label" />
          </bk-select>
          <bk-input class="search-input" v-model="filterNameData.keyword" :placeholder="t('请输入网关名称')" />
        </div>
      </div>
    </div>
    <div class="table-container" v-bkloading="{ loading: isLoading, opacity: 1, color: '#f5f7fb' }">
      <section v-if="showGateways.length">
        <div class="table-header flex-row">
          <div class="flex-1 of2">{{ t('网关名') }}</div>
          <div class="flex-1 of2">{{ t('ETCD 前缀') }}</div>
          <div class="flex-1 of2">{{ t('APISIX 类型') }}</div>
          <div class="flex-1 of1">{{ t('APISIX 版本') }}</div>
          <div class="flex-1 of08 text-r">{{ t('路由数量') }}</div>
          <div class="flex-1 of08 text-r">{{ t('服务数量') }}</div>
          <div class="flex-1 of08 text-r">{{ t('上游数量') }}</div>
          <div class="flex-1 of2 text-c">{{ t('操作') }}</div>
        </div>
        <div class="table-list">
          <div
            class="table-item flex-row align-items-center"
            v-for="item in showGateways" :key="item.id">
            <div class="flex-1 flex-row align-items-center of2">
              <div
                class="name-logo mr10"
                @click="handleGoPage('route', item)">
                {{ item.name[0].toUpperCase() }}
              </div>
              <span
                class="name mr10"
                @click="handleGoPage('route', item)">
                {{ item.name }}
              </span>
              <!-- <bk-tag theme="info" v-if="item.is_official">{{ t('官方') }}</bk-tag> -->
              <!-- <bk-tag v-if="item.status === 0">{{ t('已停用') }}</bk-tag> -->
            </div>
            <div class="flex-1 of2">{{ item?.etcd?.prefix }}</div>
            <div class="flex-1 of2">{{ item?.apisix?.type }}</div>
            <div class="flex-1 of1">{{ item?.apisix?.version }}</div>
            <!-- <div
              :class="[
                'flex-1 of2 text-c',
                { 'default-c': item.hasOwnProperty('count') }
              ]"
            >
              <template v-for="sourceName in Object.keys(item?.count)" :key="sourceName">
                <router-link :to="{ name: sourceName, params: { gatewayId: item.id } }" target="_blank">
                  <span :style="{ color: item.count[sourceName] === 0 ? '#c4c6cc' : '#3a84ff' }">
                    {{ common.enums?.resource_type[sourceName] }}: {{ item.count[sourceName] }}
                  </span>
                  <br />
                </router-link>
              </template>
            </div> -->
            <div class="flex-1 of08 text-r default-c">
              <router-link :to="{ name: 'route', params: { gatewayId: item.id } }" target="_blank">
                <span :style="{ color: item.count.route === 0 ? '#c4c6cc' : '#3a84ff' }">
                  {{ item.count.route }}
                </span>
              </router-link>
            </div>
            <div class="flex-1 of08 text-r default-c">
              <router-link :to="{ name: 'service', params: { gatewayId: item.id } }" target="_blank">
                <span :style="{ color: item.count.service === 0 ? '#c4c6cc' : '#3a84ff' }">
                  {{ item.count.service }}
                </span>
              </router-link>
            </div>
            <div class="flex-1 of08 text-r default-c">
              <router-link :to="{ name: 'upstream', params: { gatewayId: item.id } }" target="_blank">
                <span :style="{ color: item.count.upstream === 0 ? '#c4c6cc' : '#3a84ff' }">
                  {{ item.count.upstream }}
                </span>
              </router-link>
            </div>
            <div class="flex-1 of2 text-c">
              <bk-button
                text
                theme="primary"
                class="ml30"
                @click="handleGoPage('route', item)">{{ t('资源配置') }}</bk-button>
              <!-- <bk-button
                text
                theme="primary"
                class="pl20"
                @click="handleEdit(item.id)"
              >{{ t('编辑网关') }}</bk-button> -->
            </div>
          </div>
        </div>
      </section>
      <div class="empty-container" v-else>
        <div class="table-header flex-row">
          <div class="flex-1 of2">{{ t('网关名') }}</div>
          <div class="flex-1 of2">{{ t('ETCD 前缀') }}</div>
          <div class="flex-1 of2">{{ t('APISIX 类型') }}</div>
          <div class="flex-1 of1">{{ t('APISIX 版本') }}</div>
          <div class="flex-1 of08 text-r">{{ t('路由数量') }}</div>
          <div class="flex-1 of08 text-r">{{ t('服务数量') }}</div>
          <div class="flex-1 of08 text-r">{{ t('上游数量') }}</div>
          <div class="flex-1 of2 text-c">{{ t('操作') }}</div>
        </div>
        <TableEmpty
          :type="tableEmptyType"
          @clear-filter="handleClearFilterKey"
        />
      </div>
    </div>
    <div class="footer-container">
      <p class="contact" v-dompurify-html="contact"></p>
      <p class="copyright">{{copyright}}</p>
    </div>

    <create ref="createRef" @done="init()" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useUser } from '@/store/user';
import { useRouter } from 'vue-router';
import { useGetApiList } from '@/hooks';
// import { is24HoursAgo } from '@/common/util';
import { useCommon } from '@/store';
// @ts-ignore
import TableEmpty from '@/components/table-empty.vue';
// @ts-ignore
import Create from './create.vue';
import { IGatewayItem } from '@/types';
import { Plus } from 'bkui-vue/lib/icon';
// import { getGatewaysDetail } from '@/http';

const { t } = useI18n();
const user = useUser();
const router = useRouter();
const common = useCommon();

const filterKey = ref<string>('updated_at');
const filterNameData = ref({ keyword: '' });
const createRef = ref<InstanceType<typeof Create>>();

// const globalProperties = useGetGlobalProperties();
// const { GLOBAL_CONFIG } = globalProperties;

const tableEmptyType = ref<'empty' | 'search-empty'>('empty');

const isLoading = ref(true);

const selectRef = ref();
const sort = ref<string>('desc');
const isShowPopover = ref<boolean>(false);

// 网关列表数据
const gatewaysList = ref<IGatewayItem[]>([]);

const filterData = ref([
  { value: 'updated_at', label: t('更新时间') },
  { value: 'created_at', label: t('创建时间') },
  { value: 'name', label: t('首字母') },
]);

// 获取网关数据方法
const {
  getGatewaysListData,
  dataList,
} = useGetApiList({});

const contact = computed(() => {
  return (common?.websiteConfig as any)?.i18n?.footerInfoHTML;
});

const copyright = computed(() => {
  const content = (common?.websiteConfig as any)?.footerCopyrightContent;
  const version = common.appVersion || '1.0.0';
  return content.replace('{version_placeholder}', version);
});

const showGateways = computed(() => {
  if (!filterNameData.value.keyword) {
    return gatewaysList.value;
  }
  return gatewaysList.value.filter((item: IGatewayItem) => item.name.indexOf(filterNameData.value.keyword) !== -1);
});

watch(() => dataList.value, (val: IGatewayItem[]) => {
  gatewaysList.value = handleGatewaysList(val);
});

watch(
  () => showGateways.value, () => {
    updateTableEmptyConfig();
  },
  {
    deep: true,
  },
);

// 处理列表项
const handleGatewaysList = (arr: IGatewayItem[]) => {
  if (!arr) return [];

  // arr?.forEach((item: IGatewayItem) => {
  // item.is24HoursAgo = is24HoursAgo(item.created_at);
  // item.tagOrder = '3';
  // item.stages?.sort((a: any, b: any) => (b.released - a.released));
  // item.labelTextData = item.stages?.reduce((prev: any, label: any, index: number) => {
  //   if (index > item.tagOrder - 1) {
  //     prev.push({ name: label.name, released: label.released });
  //   }
  //   return prev;
  // }, []);
  // });

  return arr;
};

// 页面初始化
const init = async () => {
  isLoading.value = true;
  const list = await getGatewaysListData();
  gatewaysList.value = handleGatewaysList(list);
  setTimeout(() => {
    isLoading.value = false;
    handleSort();
  }, 100);
};
init();

const showAddDialog = () => {
  createRef.value?.show({
    maintainers: [user.user.username],
    mode: 1,
    apisix_type: 'bk-apisix',
    etcd_prefix: '/apisix',
    etcd_username: 'root',
    etcd_schema_type: 'http',
    read_only: false,
  });
};

const handleGoPage = (routeName: string, gateway: IGatewayItem) => {
  common.setGatewayId(gateway.id);
  common.setGatewayName(gateway.name);
  common.setCurGatewayData(gateway);

  // router.push({
  //   name: 'home',
  //   params: {
  //     gatewayId: `${gateway.id}`,
  //   },
  // });
  router.push(`/gateway/${gateway.id}/route`);
};

// const handleEdit = async (id: number) => {
//   const details = await getGatewaysDetail(id);

//   createRef.value?.show({
//     ...details,
//     apisix_type: details.apisix.type,
//     apisix_version: details.apisix.version,
//     etcd_endpoints: details.etcd.endpoints,
//     etcd_password: details.etcd.password,
//     etcd_prefix: details.etcd.prefix,
//     etcd_username: details.etcd.username,
//   });
// };

const handleToggle = () => {
  const dom = document.querySelector('.select-cls .label');
  if (!dom) {
    return;
  }

  if (isShowPopover.value) {
    dom.classList.remove('sel');
    isShowPopover.value = false;
    selectRef.value?.hidePopover();
  } else {
    dom.classList.add('sel');
    isShowPopover.value = true;
    selectRef.value?.showPopover();
  }
};

document.addEventListener('click', (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  const excludedButton = document.querySelector('.select-cls .label');

  if (excludedButton && !excludedButton.contains(target)) {
    excludedButton.classList.remove('sel');
    isShowPopover.value = false;
    selectRef.value?.hidePopover();
  }
});

const handleDirection = () => {
  sort.value = sort.value === 'desc' ? 'asc' : 'desc';
  handleSort();
};

const handleChange = () => {
  sort.value = 'desc';
  handleSort();
};

const handleSort = () => {
  switch (filterKey.value) {
    case 'created_at':
      if (sort.value === 'desc') {
        // @ts-ignore
        showGateways.value.sort((a: IGatewayItem, b: IGatewayItem) => new Date(b.created_at) - new Date(a.created_at));
      } else {
        // @ts-ignore
        showGateways.value.sort((a: IGatewayItem, b: IGatewayItem) => new Date(a.created_at) - new Date(b.created_at));
      }
      break;
    case 'updated_at':
      if (sort.value === 'desc') {
        // @ts-ignore
        showGateways.value.sort((a: IGatewayItem, b: IGatewayItem) => new Date(b.updated_at) - new Date(a.updated_at));
      } else {
        // @ts-ignore
        showGateways.value.sort((a: IGatewayItem, b: IGatewayItem) => new Date(a.updated_at) - new Date(b.updated_at));
      }
      break;
    case 'name':
      if (sort.value === 'desc') {
        // @ts-ignore
        showGateways.value.sort((a: IGatewayItem, b: IGatewayItem) => a.name.charAt(0).localeCompare(b.name.charAt(0)));
      } else {
        // @ts-ignore
        showGateways.value.sort((a: IGatewayItem, b: IGatewayItem) => b.name.charAt(0).localeCompare(a.name.charAt(0)));
      }
      break;
    default:
      break;
  }
};

const handleClearFilterKey = () => {
  filterNameData.value = { keyword: '' };
  filterKey.value = 'updated_at';
  getGatewaysListData();
  updateTableEmptyConfig();
};

const updateTableEmptyConfig = () => {
  const searchParams = {
    ...filterNameData.value,
  };
  const list = Object.values(searchParams).filter(item => item !== '');

  if (list.length) {
    tableEmptyType.value = 'search-empty';
  } else {
    tableEmptyType.value = 'empty';
  }
};
</script>

<style lang="scss" scoped>
.home-container{
  width: 80%;
  margin: 0 auto;
  font-size: 14px;
  min-width: 1200px;
  .title-container {
    width: 100%;
    padding-bottom: 18px;
    position: sticky;
    top: 0;
    z-index: 9;
    background-color: #f5f7fa;
    .title-total {
      font-size: 24px;
      color: #313238;
      font-weight: Medium;
      padding: 16px 0;
    }
    .left {
      font-size: 20px;
      color: #313238;
      flex: 0 0 60%;
    }
  }
  .select-cls {
    flex-shrink: 0;
    width: 126px;
    margin-right: 16px;
    :deep(.bk-select-trigger) {
      background: #EAEBF0;
      border-radius: 2px;
    }
  }
  .select-trigger {
    font-size: 12px;
    color: #63656E;
    cursor: pointer;

    .label {
      flex: 1;
      padding: 6px 8px;
      border-radius: 2px 0 0 2px;
      border: 1px solid transparent;
      &:hover {
        background: #DCDEE5;
      }
      &.sel {
        background: #FFFFFF;
        border: 1px solid #3A84FF;
      }
    }

    .suffix-cls {
      background: #EAEBF0;
      color: #979BA5;

      .line {
        width: 1px;
        height: 14px;
        background: #DCDEE5;
      }
      .icon-container {
        padding: 6px 8px;
        border-radius: 0 2px 2px 0;
        &:hover {
          background: #DCDEE5;
        }
        &.active {
          color: #3A84FF;
        }
        i {
          font-size: 16px;
        }
      }
    }
  }
  .table-container {
    width: 100%;
    min-height: calc(100vh - 192px);
    .table-header {
      width: 100%;
      color: #979ba5;
      padding: 0 16px 10px 16px;
      position: sticky;
      top: 120px;
      background-color: #f5f7fa;
    }
    .table-list{
      height: calc(100% - 45px);
      overflow-y: auto;
      .table-item{
        width: 100%;
        height: 80px;
        background: #FFFFFF;
        box-shadow: 0 2px 4px 0 #1919290d;
        border-radius: 2px;
        padding: 0 16px;
        margin: 12px 0px;
        .name-logo{
          width: 48px;
          height: 48px;
          line-height: 48px;
          text-align: center;
          background: #F0F5FF;
          border-radius: 4px;
          color: #3A84FF;
          font-size: 26px;
          font-weight: 700;
          cursor: pointer;
        }
        .name{
          font-weight: 700;
          color: #313238;
          cursor: pointer;
          &:hover{
            color: #3a84ff;
          }
        }
        .env{
          overflow: hidden;
        }
        .environment-tag {
          margin-right: 8px;
        }
      }
      .table-item:nth-of-type(1) {
        margin-top: 0px
       };

      //  .newly-item{
      //   background: #F2FFF4;
      //  }
    }
    .of08{
        flex: 0 0 8%;
      }
    .of1{
        flex: 0 0 10%;
      }
    .of3{
      flex: 0 0 30%;
    }

    .empty-table {
      :deep(.bk-table-head) {
        display: none;
      }
    }
    .empty-container {
      .empty-exception {
        background-color: #fff;
        padding-bottom: 40px;
      }
    }
  }

  .footer-container{
    position: relative;
    left: 0;
    height: 50px;
    line-height: 20px;
    // padding: 20px 0;
    display: flex;
    flex-flow: column;
    align-items: center;
    font-size: 12px;
  }

  .deact {
    background: #EAEBF0 !important;
    color: #fff !important;
    &-name{
      color: #979BA5 !important;
    }
  }

  .default-c {
    cursor: pointer;
  }
}
.ag-dot{
    width: 8px;
    height: 8px;
    display: inline-block;
    vertical-align: middle;
    border-radius: 50%;
    border: 1px solid #C4C6CC;
  }
.success{
  background: #e5f6ea;
  border: 1px solid #3fc06d;
}
.tips-cls{
  background: #f0f1f5;
  padding: 3px 8px;
  border-radius: 2px;
  cursor: default;
  &:hover{
    background: #d7d9e1 !important;
  }
}
.custom-select-option {
  padding-right: 0px !important;
  margin-right: 12px;
}
</style>
