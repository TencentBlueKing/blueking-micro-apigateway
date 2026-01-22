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
  <bk-sideslider
    v-model:is-show="isShow"
    width="1200"
  >
    <template #header>
      <slot>
        <div class="header">
          <span>{{ t('全局发布') }}</span>
        </div>
      </slot>
    </template>
    <template #default>
      <div class="main-wrapper">
        <div class="lf-wrapper">
          <div class="content-wrapper">
            <bk-search-select
              v-model="searchParams"
              :data="searchOptions"
              :placeholder="t('搜索 资源名称、ID')"
              clearable
              class="mb12"
              unique-select
            />
            <div class="diff-titles">
              <div><span class="diff-title before">{{ titleConfig.before }}</span></div>
              <div><span class="diff-title after">{{ titleConfig.after }}</span></div>
            </div>

            <div class="scroll-wrapper">
              <div v-for="item in diffList" :key="item.id">
                <div class="diff-wrapper">
                  <div :class="['resource-title', item.operationType]" :id="`${item.id}-${item.name}`">
                    <i :class="['icon', 'apigateway-icon', `icon-ag-${RESOURCE_ICON_MAP[item.resourceType]}`]"></i>
                    <span class="name">{{ common.enums?.resource_type[item.resourceType] }}：{{ item.name }}</span>
                  </div>

                  <bk-code-diff
                    v-if="!isEqual(formatJSON({ source: item.beforeConfig }), formatJSON({ source: item.afterConfig }))"
                    :diff-context="20"
                    :hljs="highlightjs"
                    :new-content="formatJSON({ source: item.afterConfig })"
                    :old-content="formatJSON({ source: item.beforeConfig })"
                    diff-format="side-by-side"
                    language="json"
                  />
                  <bk-exception
                    v-else
                    :description="t('没有差异')"
                    class="exception-wrap-item"
                    scene="part"
                    type="empty"
                  />
                </div>
              </div>

              <TableEmpty
                v-if="!diffList?.length"
                :type="emptyStatusType"
                @clear-filter="handleClearFilterKey"
              />
            </div>
          </div>
        </div>
        <div class="rg-wrapper">
          <resource-category-nav
            :show-list="diffGroupShow"
            :all-list="diffListAll"
            @search="handleOperateSearch" />
        </div>
      </div>
    </template>
    <template v-if="showFooter" #footer>
      <div class="footer-actions">
        <bk-button theme="primary" :loading="isLoading" @click="handleConfirmClick">{{ t('确定发布') }}</bk-button>
        <bk-button @click="handleCancelClick">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import i18n from '@/i18n';
import { useCommon } from '@/store';
import { Message } from 'bkui-vue';
import { RESOURCE_ICON_MAP } from '@/enum';
import highlightjs from 'highlight.js';
import useJsonTransformer from '@/hooks/use-json-transformer';
import usePublishSearch from '@/hooks/use-publish-search';
import { cloneDeep, isEqual } from 'lodash-es';
import { getResourceDiff, IDiffGroup, publishAll } from '@/http/publish';
// @ts-ignore
import ResourceCategoryNav from '@/components/resource-category-nav.vue';
import TableEmpty from '@/components/table-empty.vue';

interface IProps {
  list?: IDiffGroup[]
  showFooter?: boolean
  titleConfig?: Record<string, any>
}

interface RecombinationResource {
  id: string
  name: string
  operationType: string
  resourceType: string
}

interface DiffItem extends RecombinationResource {
  beforeConfig: Record<string, any>
  afterConfig: Record<string, any>
}

const isShow = defineModel<boolean>({
  required: true,
  default: false,
});

const {
  list,
  showFooter = true,
  titleConfig = {
    title: i18n.global.t('发布'),
    before: i18n.global.t('发布前'),
    after: i18n.global.t('发布后'),
  },
} = defineProps<IProps>();

const emit = defineEmits<{
  'done': [void]
  'cancel': [void]
}>();

const { t } = i18n.global;
const common = useCommon();
const { formatJSON } = useJsonTransformer();

const diffListAll = ref<IDiffGroup[]>([]);
const diffList = ref<DiffItem[]>([]);
const filterData = ref<Record<string, any>>({});
const searchParams = ref<{ id: string, name: string, values?: { id: string, name: string }[] }[]>([]);
const emptyStatusType = ref<'empty' | 'search-empty'>('empty');
const isLoading = ref<boolean>(false);

const searchOptions = computed(() => {
  return [
    {
      id: 'name',
      name: '资源名称',
      multiple: false,
    },
    {
      id: 'id',
      name: 'ID',
      multiple: false,
    },
    {
      id: 'operation_type',
      name: t('操作类型'),
      children: Object.keys(common.enums?.operation_type ?? {})?.filter((key: string) => (['create', 'update', 'delete'].includes(key)))
        ?.map((key: string) => ({
          name: common.enums?.operation_type[key],
          id: key,
        })),
    },
  ];
});

const getConfigList = async () => {
  diffList.value = [];
  await getDiffList();
};

const { regroupData, diffGroupShow } = usePublishSearch({
  filterData,
  diffGroupTotal: diffListAll,
  searchDoneFn: getConfigList,
});

const handleOperateSearch = ({ id, name }: { id: string, name: string }) => {
  const list = searchParams.value?.filter(param => param.id !== 'operation_type');

  searchParams.value = [
    ...list,
    {
      name: '操作类型',
      id: 'operation_type',
      values: [{ id, name }],
    },
  ];
};

const handleConfirmClick = async () => {
  try {
    isLoading.value = true;

    await publishAll();

    Message({
      theme: 'success',
      message: t('发布成功'),
    });

    isShow.value = false;
    emit('done');
  } finally {
    isLoading.value = false;
  };
};

const handleCancelClick = () => {
  isShow.value = false;
  emit('cancel');
};

const filterTimeKeys = (config: Record<string, any>) => {
  const _config = cloneDeep(config);
  ['created_at', 'updated_at', 'update_time', 'create_time'].forEach((key) => {
    if (key in _config) {
      delete _config[key];
    }
  });
  return _config;
};

const getDiffConfig = async (item: RecombinationResource) => {
  const { id, name, operationType, resourceType } = item;

  const res = await getResourceDiff({ id, type: resourceType });

  const config = res || {
    editor_config: {},
    etcd_config: {},
  };

  diffList.value.push({
    id,
    name,
    operationType,
    resourceType,
    beforeConfig: filterTimeKeys(config.etcd_config || {}),
    afterConfig: filterTimeKeys(config.editor_config || {}),
  });
};

const getDiffList = async () => {
  try {
    const resourceList: RecombinationResource[] = [];

    diffGroupShow.value?.forEach((item) => {
      item?.change_detail?.forEach((subItem) => {
        resourceList.push({
          id: subItem.resource_id,
          name: subItem.name,
          operationType: subItem.operation_type,
          resourceType: item.resource_type,
        });
      });
    });

    resourceList.forEach(async (item) => {
      await getDiffConfig(item);
    });
  } catch (e) {
    console.error(e);
  }
};

const handleSearch = () => {
  if (!isShow.value) {
    return;
  }

  const data: Record<string, any> = {};
  searchParams.value.forEach((option) => {
    if (option.values) {
      data[option.id] = option.values[0]?.id;
    } else {
      data.keywords += `&${option.id}`;
    }
  });
  filterData.value = data;

  // 过滤数据
  regroupData();
  updateTableEmptyConfig();
};

const updateTableEmptyConfig = () => {
  if (searchParams.value.length) {
    emptyStatusType.value = 'search-empty';
  } else {
    emptyStatusType.value = 'empty';
  }
};

const handleClearFilterKey = () => {
  searchParams.value = [];
  emptyStatusType.value = 'empty';
};

watch(
  () => searchParams.value,
  () => {
    handleSearch();
  },
);

watch(
  () => [isShow.value, list],
  () => {
    if (isShow.value && list?.length) {
      diffListAll.value = list;
      diffList.value = [];
      regroupData();
      updateTableEmptyConfig();
    }
    if (!isShow.value) {
      filterData.value = {};
      searchParams.value = [];
    }
  },
);

</script>

<style lang="scss" scoped>

.content-wrapper {
  padding: 24px 16px 0 30px;

  .diff-titles {
    font-size: 14px;
    position: relative;
    display: flex;
    align-items: center;
    height: 40px;
    margin-bottom: 8px;
    background: #dcdee5;
    gap: 414px;

    &::after {
      position: absolute;
      top: 8px;
      left: 50%;
      margin-left: -8px;
      width: 1px;
      height: 24px;
      content: "";
      background: #FFFFFF;
    }

    .diff-title {
      color: #313238;
      font-weight: bold;
      font-size: 14px;
      margin-left: 12px;
    }
  }
}

.footer-actions {
  display: flex;
  gap: 12px;
  padding-left: 6px;
}

.exception-wrap-item {
  padding-bottom: 12px;
  border: 1px solid #DCDEE5;
  margin-bottom: 1em;
}

.header {
  display: flex;
  align-items: center;

  .subtitle {
    font-size: 14px;
    color: #979BA5;
  }

  .line {
    width: 1px;
    height: 14px;
    background: #DCDEE5;
    margin: 0 10px;
  }
}

.diff-wrapper {
  :deep(.d2h-file-wrapper) {
    border-radius: 0px;
  }
}

.main-wrapper {
  display: flex;
  align-items: flex-start;
  .lf-wrapper {
    // flex: 1;
    width: calc(100% - 194px);
    .scroll-wrapper {
      height: calc(100vh - 216px);
      overflow-y: auto;
    }
  }
  .rg-wrapper {
    padding-top: 24px;
    width: 194px;
  }
}

.resource-title {
  height: 36px;
  line-height: 36px;
  background: #FDF4E8;
  border: 1px solid #DCDEE5;
  border-radius: 2px 2px 0 0;
  border-bottom: 0px;
  padding-left: 12px;
  .name {
    font-size: 14px;
    color: #4D4F56;
    font-weight: Bold;
  }
  .icon {
    font-size: 16px;
    margin-right: 4px;
  }
  &.create {
    background: #EBFAF0;
    .icon {
      color: #2CAF5E;
    }
  }
  &.update {
    background: #FDF4E8;
    .icon {
      color: #F59500;
    }
  }
  &.delete {
    background: #FFF0F0;
    .icon {
      color: #EA3636;
    }
  }
}
</style>
