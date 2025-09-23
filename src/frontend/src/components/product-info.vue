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
  <bk-popover
    placement="bottom"
    theme="light"
    :arrow="false"
    :padding="0"
    :always="false"
    disable-outside-click
  >
    <div class="info-icon">
      <span class="icon apigateway-icon icon-ag-help-document-fill f18"></span>
    </div>
    <template #content>
      <bk-link :href="envStore.links.bk_guide_link" class="info-item" target="_blank">
        {{ t('产品文档') }}
      </bk-link>
      <span text class="info-item" @click="showVersionLog">
        {{ t('版本日志') }}
      </span>
      <bk-link
        v-if="GLOBAL_CONFIG.HELPER.href && GLOBAL_CONFIG.HELPER.name"
        :href="GLOBAL_CONFIG.HELPER.href"
        class="info-item"
        target="_blank"
      >
        {{ t(GLOBAL_CONFIG.HELPER.name) }}
      </bk-link>
      <bk-link :href="envStore.links.bk_feed_back_link" class="info-item" target="_blank">
        {{ t('问题反馈') }}
      </bk-link>
      <bk-link :href="GLOBAL_CONFIG.MARKER" class="info-item" target="_blank">
        {{ t('开源社区') }}
      </bk-link>
    </template>
    <release-note v-model:show="showSyncReleaseNote" :list="syncReleaseList" />
  </bk-popover>
</template>

<script setup lang="ts">
import semver from 'semver';
import ReleaseNote from '@blueking/release-note';
import '@blueking/release-note/vue3/vue3.css';

import { useGetGlobalProperties } from '@/hooks';
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { getVersionLog } from '@/http';
import { useCommon, useEnv } from '@/store';

const { t } = useI18n();
const envStore = useEnv();
const commonStore = useCommon();

const globalProperties = useGetGlobalProperties();
const { GLOBAL_CONFIG } = globalProperties;

const showVersionLog = () => {
  showSyncReleaseNote.value = true;
};

const showSyncReleaseNote = ref(false);
const syncReleaseList = ref([]);

onMounted(async () => {
  try {
    const list = await getVersionLog();
    list.forEach((item) => {
      syncReleaseList.value.push({
        title: item.version,
        detail: item.content,
        ...item,
      });
    });

    const latestVersion = list[0].version;
    commonStore.setAppVersion(latestVersion);
    const localLatestVersion = localStorage.getItem('latest-version');
    if (!localLatestVersion
      || semver.compare(localLatestVersion.replace(/^V/, ''), latestVersion.replace(/^V/, '')) === -1
    ) {
      localStorage.setItem('latest-version', latestVersion);
      showVersionLog();
    }
  } catch {
    syncReleaseList.value = [];
  }
});

</script>

<style lang="scss" scoped>
.info-icon {
  width: 32px;
  height: 32px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 50%;
  cursor: pointer;
}
.info-icon:hover {
  background-color: #303d55;
  color: #fff;
}
.info-item {
  display: block;
  color: #63656E;
  padding: 8px 15px;
  margin-top: 5px;
  font-size: 12px;
  user-select: none;
  cursor: pointer;
  &:hover {
    color: #979ba5;
  }
}
.info-item:hover {
  background: #f3f6f9;
}
.info-item:nth-of-type(1) {
    margin-top: 0px;
}
</style>
