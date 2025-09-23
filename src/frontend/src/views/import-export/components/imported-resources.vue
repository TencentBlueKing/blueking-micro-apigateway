<template>
  <div class="imported-resources-wrap">
    <header class="res-counter-banner">
      <main>
        <Icon
          color="#3A84FF"
          name="info"
          size="16"
          style="margin-right: 5px;"
        />
        <span>{{ t('共') }}<span
          style="font-weight: bold;padding-inline: 5px;"
        >{{ resourceCount }}</span>{{ t('个资源，新增') }}<span
          style="font-weight: bold;color: #34d97b;padding-inline: 5px;"
        >{{ rawTableDataToAdd.length }}</span>{{ t('个，更新') }}<span
          style="font-weight: bold;color: #ffb400;padding-inline: 5px;"
        >{{ rawTableDataToUpdate.length }}</span>{{ t('个，取消导入') }}<span
          style="font-weight: bold;color: #ff5656;padding-inline: 5px;"
        >{{
          uncheckedResources.length
        }}</span>{{ t('个') }}</span>
      </main>
      <aside>
        <bk-button
          text
          theme="primary"
          @click="handleRecoverAll"
        >
          <Icon
            color="#3A84FF"
            name="undo-2"
            size="16"
            style="margin-right: 4px;"
          />
          {{ t('恢复取消导入的资源') }}
        </bk-button>
      </aside>
    </header>
    <main class="res-content-wrap">
      <bk-collapse
        v-model="panelNamesList"
        class="collapse-cls"
        use-card-theme
      >
        <!--  新增的资源  -->
        <bk-collapse-panel name="add">
          <template #header>
            <div
              ref="panelHeadAddRef"
              class="panel-header"
            >
              <main style="display: flex;align-items:center;">
                <AngleUpFill
                  :class="[panelNamesList.includes('add') ? 'panel-header-show' : 'panel-header-hide']"
                />
                <div class="title">
                  {{ t('新增的资源（共') }}
                  <span style="color: #34b97b;">{{ rawTableDataToAdd.length }}</span>
                  {{ t('个）') }}
                </div>
              </main>
              <aside @click.stop.prevent>
                <bk-input
                  v-model="filterInputAddClone"
                  :placeholder="t('请输入资源名称/ID，按Enter搜索')"
                  clearable
                  style="width: 578px;"
                  @clear="() => filterData('add')"
                  @enter="() => filterData('add')"
                >
                  <template #prefix>
                    <aside
                      style="display: flex;align-items:center;margin-left: 10px;"
                      @click="() => filterData('add')"
                    >
                      <Icon name="search" />
                    </aside>
                  </template>
                </bk-input>
              </aside>
            </div>
          </template>
          <template #content>
            <div>
              <!--  新增资源 table  -->
              <resource-table :data="tableDataToAdd" type="update" @uncheck="(row) => handleUncheck(row, 'add')" />
            </div>
          </template>
        </bk-collapse-panel>
        <!--  更新的资源  -->
        <bk-collapse-panel name="update">
          <template #header>
            <div
              ref="panelHeadUpdateRef"
              class="panel-header"
            >
              <main style="display: flex;align-items:center;">
                <AngleUpFill
                  :class="[panelNamesList.includes('update') ? 'panel-header-show' : 'panel-header-hide']"
                />
                <div class="title">
                  {{ t('更新的资源（共') }}
                  <span style="color: #ffb400;">{{ rawTableDataToUpdate.length }}</span>
                  {{ t('个）') }}
                </div>
              </main>
              <aside @click.stop.prevent>
                <bk-input
                  v-model="filterInputUpdateClone"
                  :placeholder="t('请输入资源名称/ID')"
                  clearable
                  style="width: 578px;"
                  @clear="() => filterData('update')"
                  @enter="() => filterData('update')"
                >
                  <template #prefix>
                    <aside
                      style="display: flex;align-items:center;margin-left: 10px;"
                      @click="() => filterData('update')"
                    >
                      <Icon name="search" />
                    </aside>
                  </template>
                </bk-input>
              </aside>
            </div>
          </template>
          <template #content>
            <div>
              <!--  更新资源 table  -->
              <resource-table
                :data="tableDataToUpdate"
                type="update"
                @uncheck="(row) => handleUncheck(row, 'update')"
              />
            </div>
          </template>
        </bk-collapse-panel>
        <!--  不导入的资源  -->
        <bk-collapse-panel name="uncheck">
          <template #header>
            <div class="panel-header">
              <main style="display: flex;align-items:center;">
                <AngleUpFill
                  :class="[panelNamesList.includes('uncheck') ? 'panel-header-show' : 'panel-header-hide']"
                />
                <div class="title">
                  {{ t('不导入的资源（共') }}
                  <span style="color: #ff5656;">{{ uncheckedResources.length }}</span>
                  {{ t('个）') }}
                </div>
              </main>
            </div>
          </template>
          <template #content>
            <div>
              <!--  不导入资源 table  -->
              <resource-table :data="uncheckedResources" type="uncheck" @recover="(row: IRow) => handleRecover(row)" />
            </div>
          </template>
        </bk-collapse-panel>
      </bk-collapse>
    </main>
  </div>
</template>

<script lang="tsx" setup>
import { useI18n } from 'vue-i18n';
import { ref, computed, watch } from 'vue';
import Icon from '@/components/icon.vue';
import { AngleUpFill } from 'bkui-vue/lib/icon';
import ResourceTable from './resource-table.vue';

interface IProps {
  data: {
    add?: IRow[],
    update?: IRow[],
  }
}

interface IRow {
  name: string
  resource_id: string
  resource_type: string
  status: string
  config: Record<string, any>
  __source__?: string
}

const { data = {} } = defineProps<IProps>();

const { t } = useI18n();

const rawTableDataToAdd = ref<IRow[]>([]);
const rawTableDataToUpdate = ref<IRow[]>([]);
const panelNamesList = ref(['add', 'update', 'uncheck']);
const filterInputAdd = ref('');
const filterInputAddClone = ref('');
const filterInputUpdate = ref('');
const filterInputUpdateClone = ref('');

const uncheckedResources = ref<IRow[]>([]);

// 展示在“新增的资源”一栏的资源
// eslint-disable-next-line max-len
const tableDataToAdd = computed(() => rawTableDataToAdd.value.filter(data => data.name?.includes(filterInputAdd.value) || data.resource_id?.includes(filterInputAdd.value)));

// 展示在“更新的资源”一栏的资源
// eslint-disable-next-line max-len
const tableDataToUpdate = computed(() => rawTableDataToUpdate.value.filter(data => data.name?.includes(filterInputUpdate.value) || data.resource_id?.includes(filterInputUpdate.value)));

// eslint-disable-next-line max-len
const resourceCount = computed(() => rawTableDataToAdd.value.length + rawTableDataToUpdate.value.length + uncheckedResources.value.length);

watch(() => data, () => {
  rawTableDataToAdd.value = data.add ? data.add.map(row => ({ ...row, __source__: 'add' })) : [];
  rawTableDataToUpdate.value = data.update ? data.update.map(row => ({ ...row, __source__: 'update' })) : [];
}, { deep: true, immediate: true });

const filterData = (action: string) => {
  if (action === 'add') {
    filterInputAdd.value = filterInputAddClone.value;
  }

  if (action === 'update') {
    filterInputUpdate.value = filterInputUpdateClone.value;
  }
};

// 还原所有不导入的资源
const handleRecoverAll = () => {
  uncheckedResources.value.forEach((row) => {
    if (row.__source__ === 'add') {
      rawTableDataToAdd.value.push(row);
    } else {
      rawTableDataToUpdate.value.push(row);
    }
  });
  uncheckedResources.value = [];
};

const handleRecover = (row: IRow) => {
  const index = uncheckedResources.value.findIndex(d => d.resource_id === row.resource_id);
  if (index > -1) {
    uncheckedResources.value.splice(index, 1);
    if (row.__source__ === 'add') {
      rawTableDataToAdd.value.push(row);
    } else {
      rawTableDataToUpdate.value.push(row);
    }
    delete row.__source__;
  }
};

const handleUncheck = (row: IRow, source?: 'add' | 'update') => {
  if (source === 'add') {
    const index = rawTableDataToAdd.value.findIndex(d => d.resource_id === row.resource_id);
    if (index > -1) {
      rawTableDataToAdd.value.splice(index, 1);
    }
  } else if (source === 'update') {
    const index = rawTableDataToUpdate.value.findIndex(d => d.resource_id === row.resource_id);
    if (index > -1) {
      rawTableDataToUpdate.value.splice(index, 1);
    }
  }

  uncheckedResources.value.push({
    ...row,
    __source__: source,
  });
};

defineExpose({
  getValue: () => {
    const result: {
      add?: Record<string, Omit<IRow, '__source__'>[]>,
      update?: Record<string, Omit<IRow, '__source__'>[]>,
    } = {};
    if (rawTableDataToAdd.value.length) {
      result.add = rawTableDataToAdd.value.reduce<Record<string, any[]>>((acc, item) => {
        delete item.__source__;
        if (acc[item.resource_type]) {
          acc[item.resource_type].push(item);
        } else {
          acc[item.resource_type] = [item];
        }
        return acc;
      }, {});
    }
    if (rawTableDataToUpdate.value.length) {
      result.update = rawTableDataToUpdate.value.reduce<Record<string, any[]>>((acc, item) => {
        delete item.__source__;
        if (acc[item.resource_type]) {
          acc[item.resource_type].push(item);
        } else {
          acc[item.resource_type] = [item];
        }
        return acc;
      }, {});
    }
    return result;
  },
});
</script>

<style lang="scss" scoped>
.imported-resources-wrap {
  .res-counter-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 40px;
    margin-bottom: 16px;
    padding: 0 24px 0 12px;
    border-radius: 2px;
    background: #ffffff;
    box-shadow: 0 2px 4px 0 #1919290d;
  }

  .res-content-wrap {
    overflow-y: scroll;
    height: calc(100vh - 285px);

    &::-webkit-scrollbar {
      width: 4px;
      height: 4px;
    }

    &::-webkit-scrollbar-thumb {
      border-radius: 20px;
      background: #dddddd;
      box-shadow: inset 0 0 6px #cccccc4d;
    }

    :deep(.collapse-cls) {
      margin-bottom: 24px;

      .bk-collapse-item {
        margin-bottom: 16px;
        background: #ffffff;
        box-shadow: 0 2px 4px 0 #1919290d;
      }

      // 折叠让 .panel-header 粘滞固定到顶部

      .panel-head-sticky-top {
        position: sticky;
        z-index: 3;
        top: 0;
        background-color: #ffffff;
      }
    }

    :deep(.bk-collapse-content) {
      padding-top: 0 !important;
      padding-bottom: 0 !important;
      padding-inline: 24px !important;
    }

    .panel-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 19px 24px;
      cursor: pointer;

      .title {
        font-size: 14px;
        font-weight: 700;
        margin-left: 8px;
        color: #313238;
      }

      .panel-header-show {
        transition: .2s;
        transform: rotate(0deg);
      }

      .panel-header-hide {
        transition: .2s;
        transform: rotate(-90deg);
      }
    }
  }
}
</style>
