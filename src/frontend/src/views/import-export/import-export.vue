<template>
  <div class="import-export-page-wrapper">
    <div class="action-section import">
      <div class="icon-wrapper">
        <img :src="ImportIcon" alt="Import" class="icon-img">
      </div>
      <div class="help-text">{{ t('导入文件资源数据到编辑区') }}</div>
      <div class="button-wrapper">
        <div class="btn" @click="() => toPage('import-export-upload')">{{ t('导入') }}</div>
      </div>
    </div>
    <div class="action-section export">
      <div class="icon-wrapper">
        <img :src="ExportIcon" alt="Import" class="icon-img">
      </div>
      <div class="help-text">{{ t('导出 etcd 资源配置') }}</div>
      <div class="button-wrapper">
        <div class="btn" @click="handleExportClicked">{{ t('导出') }}</div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import ImportIcon from '@/images/import.svg';
import ExportIcon from '@/images/export.svg';
import { useRouter } from 'vue-router';
import { useCommon } from '@/store';
import { computed } from 'vue';
import Cookie from 'js-cookie';

const { BK_DASHBOARD_URL } = window;

const { t } = useI18n();
const router = useRouter();
const commonStore = useCommon();


// const downloadLink = computed(() => `/gateways/${commonStore.gatewayId}/unify_op/etcd/export/`);
const fileName = computed(() => `${commonStore.curGatewayData?.name || 'gateway_resource'}.json`);

const toPage = (name: string) => {
  router.push({ name });
};

const handleExportClicked = async () => {
  const response = await fetch(
    `${BK_DASHBOARD_URL}/gateways/${commonStore.gatewayId}/unify_op/etcd/export/`,
    {
      headers: {
        'X-CSRF-TOKEN': Cookie.get(window.BK_DASHBOARD_CSRF_COOKIE_NAME),
      },
      credentials: 'include',
    },
  );
  const j = await response.json();
  const blob = new Blob([JSON.stringify(j, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = fileName.value;
  document.body.appendChild(a);
  a.click();
};

</script>

<style lang="scss" scoped>
.import-export-page-wrapper {
  display: flex;
  justify-content: center;
  padding-top: 160px;
  gap: 32px;

  .action-section {
    display: flex;
    align-items: center;
    flex-direction: column;
    justify-content: center;
    width: 300px;
    background-color: #ffffff;
    padding-block: 50px;

    .icon-wrapper {
      margin-bottom: 19px;

      .icon-img {
        width: 65px;
        height: 65px;
      }
    }

    .help-text {
      font-size: 12px;
      line-height: 20px;
      margin-bottom: 24px;
      text-align: center;
      color: #63656e;
    }

    .button-wrapper {

      .btn {
        font-size: 14px;
        display: flex;
        align-items: center;
        justify-content: center;
        width: 160px;
        height: 36px;
        cursor: pointer;
        color: #4d4f56;
        border: 1px solid #c4c6cc;
        border-radius: 18px;
      }

      .btn:hover {
        color: #3a84ff;
        border-color: #3a84ff;
        background: #f0f5ff;
        box-shadow: 0 2px 6px 0 #3a84ff1a;
      }
    }
  }
}
</style>
