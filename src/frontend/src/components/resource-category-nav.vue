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
  <div class="nav">
    <div class="change-total">
      <bk-tag theme="success" type="stroke" class="pointer" @click="handleClick({ id: 'create', name: '新增' })">
        {{ t('新增: ') }}{{ addCount }}
      </bk-tag>
      <bk-tag theme="warning" type="stroke" class="pointer" @click="handleClick({ id: 'update', name: '更新' })">
        {{ t('更新: ') }}{{ modifiedCount }}
      </bk-tag>
      <bk-tag theme="danger" type="stroke" class="pointer" @click="handleClick({ id: 'delete', name: '删除' })">
        {{ t('删除: ') }}{{ delCount }}
      </bk-tag>
    </div>
    <bk-collapse
      class="collapse-category"
      v-model="activeIndex"
    >
      <bk-collapse-panel
        v-for="item in showList"
        :key="item.resource_type"
        :name="item.resource_type"
      >
        <template #header>
          <div :class="{ 'category': true, 'active': activeIndex?.includes(item.resource_type) }">
            <right-shape class="icon" />
            <span class="name">{{ common.enums?.resource_type[item.resource_type] }}</span>
          </div>
        </template>
        <template #content>
          <div class="resource">
            <div
              v-for="resource in item.change_detail"
              :key="resource.resource_id"
              :class="{ 'highlight-wrapper': true, 'active': currentResource?.resource_id === resource.resource_id }"
              v-bk-tooltips="{ content: resource.name, placement: 'left' }"
              @click="changeResource(resource)">
              <div class="name">
                {{ resource.name }}
              </div>
            </div>
          </div>
        </template>
      </bk-collapse-panel>
    </bk-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import i18n from '@/i18n';
import { IDiffGroup, IChangeDetail } from '@/http/publish';
import { RightShape } from 'bkui-vue/lib/icon';
import { useCommon } from '@/store';

interface IProps {
  showList: IDiffGroup[]
  allList: IDiffGroup[]
}

const {
  showList,
  allList,
} = defineProps<IProps>();

const emit = defineEmits(['search']);

const { t } = i18n.global;
const common = useCommon();

const activeIndex = ref<string[]>([...showList?.map(item => item?.resource_type)]);

const currentResource = ref<IChangeDetail>(showList[0]?.change_detail[0]);

const addCount = ref<number>(0);
const delCount = ref<number>(0);
const modifiedCount = ref<number>(0);

const changeResource = (resource: IChangeDetail) => {
  currentResource.value = resource;

  const element = document.getElementById(`${resource.resource_id}-${resource.name}`);
  if (element) {
    element.scrollIntoView({
      behavior: 'smooth', // 平滑滚动
      block: 'start', // 元素顶部与视口顶部对齐
    });
  }
};

const handleClick = (type: { id: string, name: string }) => {
  emit('search', type);
};

const getTotal = () => {
  let add = 0;
  let del = 0;
  let modified = 0;

  allList?.forEach((item) => {
    add += item.added_count;
    del += item.deleted_count;
    modified += item.modified_count;
  });

  addCount.value = add;
  delCount.value = del;
  modifiedCount.value = modified;
};

watch(
  () => allList,
  () => {
    getTotal();
  },
  { deep: true, immediate: true },
);

</script>

<style lang="scss" scoped>
.change-total {
  padding-top: 4px;
  margin-bottom: 16px;
}

.bk-collapse-item {
  margin-bottom: 12px;
}

:deep(.bk-collapse-content) {
  padding: 12px 0 0;
}

.category {
  cursor: pointer;
  line-height: 20px;
  .name {
    font-size: 12px;
    color: #4D4F56;
    margin-left: 4px;
  }
  .icon {
    color: #C4C6CC;
    transform: rotate(0deg);
    transition: all .2s;
  }
  &.active {
    .name {
      font-weight: Bold;
    }
    .icon {
      transform: rotate(90deg);
    }
  }
}

.resource {
  .highlight-wrapper {
    position: relative;
    padding-right: 32px;
    padding-left: 42px;
    margin-left: -10px;
    &::before {
      content: ' ';
      position: absolute;
      left: 0px;
      top: -6px;
      width: 2px;
      height: 32px;
      z-index: 99;
    }
    &.active {
      .name {
        color: #3A84FF;
      }
      &::before {
        background-color: #3A84FF;
      }
    }
    &:not(:nth-last-child(1)) {
      margin-bottom: 12px;
    }
  }
  .name {
    font-size: 12px;
    color: #4D4F56;
    width: 102px;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
    overflow: hidden;
    cursor: pointer;
  }
}

.collapse-category {
  max-height: calc(100vh - 168px);
  overflow-y: auto;
  padding-left: 10px;
  padding-top: 12px;
  position: relative;
  &::before {
    content: ' ';
    position: absolute;
    left: 0;
    top: 0;
    width: 1px;
    height: 100%;
    background-color: #DCDEE5;
  }
}

.pointer {
  cursor: pointer;
}
</style>
