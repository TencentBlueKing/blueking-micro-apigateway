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
  <div class="enabled-plugins">
    <div class="plugin-groups">
      <div class="plugin-group">
        <div class="plugin-items">
          <div v-for="plugin in enabledPluginList" :key="plugin.name" class="plugin-item">
            <div class="plugin-icon">
              <Icon color="#3a84ff" name="plugin-generic-fill" size="20" />
            </div>
            <div class="plugin-name">{{ plugin.name }}</div>
            <div class="plugin-action">
              <bk-button size="small" @click="handleEditClick(plugin)">{{
                t('编辑')
              }}
              </bk-button>
            </div>
            <bk-pop-confirm
              :title="t('确认删除？')"
              trigger="click"
              width="288"
              @confirm="handleRemovePlugin({ name: plugin.name })"
            >
              <Icon class="plugin-item-delete" color="#979BA5" name="close-circle-filled" size="18" />
            </bk-pop-confirm>
          </div>
        </div>
      </div>
    </div>
  </div>
  <bk-sideslider
    v-model:is-show="isShow"
    :title="t('添加插件')"
    width="960"
    @closed="keyword = ''"
  >
    <template #default>
      <div class="content-wrapper">
        <div class="plugin-filter">
          <bk-input v-model="keyword" :placeholder="t('搜索插件名称')" class="search-input" clearable type="search" />
        </div>
        <div class="plugin-and-side-anchor">
          <main ref="scroll-target-parent-t" class="plugin-wrapper">
            <div ref="scroll-target" class="plugin-groups custom-scroll-bar">
              <template v-if="filteredPluginList.length">
                <div
                  v-for="group in pluginGroupList"
                  :id="group.type"
                  :key="group.type"
                  ref="scroll-anchor-refs"
                  class="plugin-group"
                >
                  <div class="plugin-group-title">{{ PLUGIN_TYPE_CN_MAP[group.type] || group.type }}</div>
                  <div class="plugin-items">
                    <div
                      v-for="plugin in group.plugins"
                      :key="plugin.name"
                      v-bk-tooltips="{ content: t('已配置 Global Rule'), disabled: !isDisabled(plugin.name) }"
                      :class="{ disabled: isDisabled(plugin.name) }"
                      class="plugin-item"
                    >
                      <div class="plugin-icon">
                        <Icon color="#cccccc" name="plugin-generic-fill" size="20" />
                      </div>
                      <div class="plugin-name">{{ plugin.name }}</div>
                      <div class="plugin-action">
                        <bk-button
                          :disabled="isDisabled(plugin.name)"
                          size="small"
                          theme="primary"
                          @click="handleEnableClick(plugin)"
                        >
                          {{ t('启用') }}
                        </bk-button>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
              <table-empty
                v-else
                :type="emptyStatusType"
                @clear-filter="handleClearFilter"
              >
                <template #title>{{ availablePluginList.length ? t('暂无数据') : t('没有可启用的插件') }}</template>
              </table-empty>
            </div>
          </main>
          <aside class="side-anchor">
            <side-nav-plugin
              v-model="activeGroupId"
              v-model:scroll-y="scrollY"
              :container="scrollTargetRef"
              :elements="scrollAnchorElementsRefs"
              :list="groupNavList"
              :parent="scrollTargetParentRef"
              width="90"
            />
          </aside>
        </div>
      </div>
    </template>
    <template v-if="showFooter" #footer>
      <div class="footer-actions">
        <bk-button theme="primary" @click="handleConfirmClick">{{ t('保存') }}</bk-button>
        <bk-button @click="handleCancelClick">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>
  <slider-plugin-config-editor v-model="isConfigSliderShow" :plugin="editingPlugin" @confirm="handleEditConfirm" />
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { computed, nextTick, onBeforeMount, ref, useTemplateRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import { refDebounced } from '@vueuse/core';
import Ajv from 'ajv';
import addFormats from 'ajv-formats';
import { IPlugin } from '@/types/plugin';
import { useCommon } from '@/store';
import useSchemaErrorMessage from '@/hooks/use-schema-error-message';
import { PLUGIN_TYPE_CN_MAP } from '@/enum';
import {
  getConsumerPlugins,
  getConsumerPluginSchema,
  getMetadataPlugins,
  getMetadataPluginSchema,
  getPluginSchema,
} from '@/http/plugins';
import { Message } from 'bkui-vue';
import SideNavPlugin from '@/components/side-nav-plugin.vue';
import Icon from '@/components/icon.vue';
import SliderPluginConfigEditor from '@/components/slider-plugin-config-editor.vue';
import { cloneDeep } from 'lodash-es';
import useJsonTransformer from '@/hooks/use-json-transformer';
import TableEmpty from '@/components/table-empty.vue';

interface IProps {
  type?: 'plugins' | 'consumer' | 'metadata'
  disableList?: string[]
  showFooter?: boolean
  pluginQuery?: { kind: string }
}

const isShow = defineModel<boolean>('isShow', {
  required: true,
  default: false,
});

const enabledPluginList = defineModel<ILocalPlugin[]>({
  required: true,
});

const {
  type = 'plugins',
  disableList = [],
  showFooter = false,
  pluginQuery = {},
} = defineProps<IProps>();

const emit = defineEmits<{
  'removed': [name: string],
}>();

interface ILocalPlugin {
  doc_url?: string
  example?: string
  id?: string
  name: string
  config: string
  enabled?: boolean
  type: string
}

interface IPluginGroup {
  type: string
  plugins: IPlugin[]
}

const { t } = useI18n();
const common = useCommon();
const { showSchemaErrorMessages } = useSchemaErrorMessage();
const { formatJSON } = useJsonTransformer();
const route = useRoute();

const keyword = ref('');
const keywordDebounced = refDebounced(keyword, 150);
const emptyStatusType = ref<'empty' | 'search-empty'>('empty');

const activeGroupId = ref('');
const scrollY = ref(0);
const scrollAnchorElementsRefs = useTemplateRef<HTMLDivElement[]>('scroll-anchor-refs');
const scrollTargetRef = useTemplateRef('scroll-target');
const scrollTargetParentRef = useTemplateRef<HTMLDivElement>('scroll-target-parent-t');

const isConfigSliderShow = ref(false);

const editingPlugin = ref<ILocalPlugin>({
  name: '',
  config: '{}',
  type: '',
});

const consumerPluginList = ref<IPlugin[]>([]);
const metadataPluginList = ref<IPlugin[]>([]);

const schema = ref<Record<string, any>>({});

// 完整的插件列表
const allPluginList = computed(() => {
  if (type === 'plugins') {
    return common.plugins || [];
  }
  if (type === 'consumer') {
    return consumerPluginList.value;
  }
  if (type === 'metadata') {
    return metadataPluginList.value;
  }

  return common.plugins;
});

// 可启用的插件列表
const availablePluginList = computed(() => {
  return allPluginList.value.filter(plugin => !enabledPluginList.value.map(item => item.name)
    .includes(plugin.name));
});

const filteredPluginList = computed(() => {
  if (!keyword.value) {
    return availablePluginList.value;
  }

  return availablePluginList.value.filter(plugin => plugin.name.includes(keywordDebounced.value));
});

const pluginGroupList = computed(() => {
  const rawGroups = filteredPluginList.value.reduce<IPluginGroup[]>((groupList, plugin) => {
    const group = groupList.find(_group => _group.type === plugin.type);
    if (group) {
      groupList.find(group => group.type === plugin.type)
        .plugins
        .push(plugin);
    } else {
      groupList.push({
        type: plugin.type,
        plugins: [plugin],
      });
    }
    return groupList;
  }, []);

  return sortGroups(rawGroups);
});

const groupNavList = computed(() => {
  return pluginGroupList.value.map(group => ({
    id: group.type,
    name: PLUGIN_TYPE_CN_MAP[group.type] || group.type,
  }));
});

watch(keyword, () => {
  nextTick(() => {
    scrollY.value = 0;
  });
});

watch(filteredPluginList, () => {
  if (!filteredPluginList.value.length) {
    if (keyword.value) {
      emptyStatusType.value = 'search-empty';
    } else {
      emptyStatusType.value = 'empty';
    }
  } else {
    emptyStatusType.value = 'empty';
  }
});

const getSchema = async (name: string) => {
  if (type === 'plugins') {
    const query = ['stream-route-create', 'stream-route-clone', 'stream-route-edit'].includes(route.name) ? { schema_type: 'stream' } : {};
    return await getPluginSchema({ name, gatewayId: common.gatewayId, query });
  }
  if (type === 'consumer') {
    return await getConsumerPluginSchema({ name, gatewayId: common.gatewayId });
  }
  if (type === 'metadata') {
    return await getMetadataPluginSchema({ name, gatewayId: common.gatewayId });
  }

  return {};
};

const handleEditClick = async (plugin: ILocalPlugin) => {
  schema.value = await getSchema(plugin.name);

  editingPlugin.value = {
    name: plugin.name,
    config: formatJSON({ source: plugin.config }),
    doc_url: plugin.doc_url,
    type: plugin.type,
  };
  isConfigSliderShow.value = true;
};

const handleEnableClick = async (plugin: IPlugin) => {
  schema.value = await getSchema(plugin.name);

  let config = '{}';

  if (type === 'plugins') {
    config = JSON.stringify(plugin.example);
  } else if (type === 'consumer') {
    config = JSON.stringify(plugin.consumer_example || plugin.example);
  } else if (type === 'metadata') {
    config = JSON.stringify(plugin.metadata_example || plugin.example);
  }

  editingPlugin.value = {
    config,
    name: plugin.name,
    doc_url: plugin.doc_url,
    type: plugin.type,
  };
  isConfigSliderShow.value = true;
};

const isDisabled = (pluginName: string) => {
  return disableList.includes(pluginName);
};

const handleEditConfirm = (plugin: ILocalPlugin) => {
  try {
    // 校验 schema
    const ajv = new Ajv();
    addFormats(ajv);
    const schemaValidate = ajv.compile(schema.value);
    if (schemaValidate(JSON.parse(plugin.config))) {
      const targetPlugin = enabledPluginList.value.find(item => item.name === plugin.name);

      if (targetPlugin) {
        targetPlugin.config = plugin.config;
      } else {
        enabledPluginList.value.push(plugin);
      }

      isConfigSliderShow.value = false;
      Message({
        theme: 'success',
        message: t('配置成功'),
      });
      isShow.value = false;
      keyword.value = '';
    } else {
      showSchemaErrorMessages(schemaValidate.errors);
      // throw Error(t('配置不正确'));
    }
  } catch (err) {
    const error = err as Error;
    Message({
      theme: 'error',
      message: error.message || t('配置不正确'),
    });
  }
};

const handleRemovePlugin = ({ name }: { name: string }) => {
  const index = enabledPluginList.value.findIndex(item => item.name === name);

  if (index > -1) {
    enabledPluginList.value.splice(index, 1);
    emit('removed', name);
  }
};

const sortGroups = (groups: IPluginGroup[]) => {
  const otherGroupIndex = groups.findIndex(group => group.type === 'other' || group.type === 'other protocols');
  if (otherGroupIndex > -1) {
    moveToLast(groups, otherGroupIndex);
  }

  const tapisixGroupIndex = groups.findIndex(group => group.type === 'tapisix');
  if (tapisixGroupIndex > -1) {
    moveToLast(groups, tapisixGroupIndex);
  }

  const bkApisixGroupIndex = groups.findIndex(group => group.type === 'bk-apisix');
  if (bkApisixGroupIndex > -1) {
    moveToLast(groups, bkApisixGroupIndex);
  }

  const customGroupIndex = groups.findIndex(group => group.type === 'customize plugin');
  if (customGroupIndex > -1) {
    moveToLast(groups, customGroupIndex);
  }

  return groups;
};

const moveToLast = (list: IPluginGroup[], index: number) => {
  const target = cloneDeep(list[index]);
  list.splice(index, 1);
  list.push(target);
};

const handleClearFilter = () => {
  keyword.value = '';
  emptyStatusType.value = 'empty';
};

const tidyPlugins = (response: IPluginGroup[]) => {
  return response.reduce((acc, cur) => {
    acc = [...acc, ...cur.plugins];
    return acc;
  }, [] as IPlugin[]);
};

onBeforeMount(async () => {
  if (type === 'plugins') {
    await common.setPlugins({ query: pluginQuery });
  } else if (type === 'consumer') {
    const response = await getConsumerPlugins({ gatewayId: common.gatewayId });
    consumerPluginList.value = tidyPlugins(response || []);
  } else if (type === 'metadata') {
    const response = await getMetadataPlugins({ gatewayId: common.gatewayId });
    metadataPluginList.value = tidyPlugins(response || []);
  }
  activeGroupId.value = filteredPluginList.value?.[0]?.type || '';
});

const handleConfirmClick = async () => {
  isShow.value = false;
};

const handleCancelClick = () => {
  isShow.value = false;
};

</script>

<style lang="scss" scoped>

.bk-modal-content {
  overflow: hidden;
}

.content-wrapper {
  font-size: 14px;
  padding: 24px 24px 0;

  .plugin-filter {
    margin-bottom: 12px;
  }

  .plugin-and-side-anchor {
    display: flex;
    gap: 18px;

    .plugin-wrapper {
      position: relative;
      display: flex;
      width: 776px;
      //height: calc(100vh - 200px);
      max-height: calc(100vh - 168px);
    }
  }
}

.plugin-groups {
  overflow-y: scroll;
  flex-grow: 1;
  background-color: #ffffff;

  .plugin-group {
    margin-bottom: 24px;

    .plugin-group-title {
      font-size: 14px;
      font-weight: 700;
      line-height: 28px;
      height: 28px;
      margin-right: 12px;
      margin-bottom: 12px;
      padding-left: 8px;
      color: #313238;
      border-radius: 2px;
      background: #f0f1f5;
    }

    .plugin-items {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;

      .plugin-item {
        position: relative;
        display: flex;
        align-items: center;
        width: 246px;
        height: 48px;
        padding: 8px;
        cursor: pointer;
        border-radius: 4px;
        background-color: #f5f7fb;

        &:hover {
          box-shadow: 0 2px 4px 0 #0000001a, 0 2px 4px 0 #1919290d;
        }

        .plugin-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 32px;
          height: 32px;
          border-radius: 2px;
          background: #ffffff;
        }

        .plugin-name {
          font-weight: 500;
          flex-grow: 1;
          padding-inline: 12px;
        }

        .plugin-action {
          display: flex;
          align-items: center;
          flex-shrink: 0;
          justify-content: center;
          width: 52px;
        }

        .plugin-item-delete {
          position: absolute;
          top: -8px;
          right: -8px;
          cursor: pointer;

          &:hover {
            color: #aaaaaa !important;
          }
        }
      }

      .plugin-item.disabled {
        box-shadow: none;
      }
    }
  }
}

.enabled-plugins {
  .plugin-groups {
    overflow: visible;
  }
}

.footer-actions {
  display: flex;
  gap: 12px;
}

</style>
