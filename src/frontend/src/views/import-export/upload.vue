<template>
  <div class="upload-page-wrapper">
    <div class="page-header">
      <div class="header-time-line">
        <bk-steps :cur-step="currentStep" :steps="steps" />
      </div>
    </div>
    <div class="page-content">
      <form-card v-if="currentStep === 1">
        <template #title>{{ t('上传文件') }}</template>
        <div class="drop-zone-wrapper">
          <bk-upload
            :handle-res-code="handleResCode"
            :header="{ name: 'X-CSRF-TOKEN', value: CSRFToken }"
            :is-show-preview="false"
            :limit="1"
            :tip="t('仅支持 .json 格式的文件')"
            :url="uploadUrl"
            accept=".json"
            class="upload-cls"
            name="resource_file"
            with-credentials
            :size="30"
            @success="handleUploadDone"
          />
        </div>
      </form-card>
      <imported-resources
        v-if="currentStep === 2"
        ref="importedResourcesRef"
        :data="resources"
      />
    </div>
    <footer v-if="currentStep === 2" class="page-actions-wrap">
      <main class="page-actions">
        <bk-button @click="handleBack">
          {{ t('上一步') }}
        </bk-button>
        <bk-button
          theme="primary"
          @click="handleConfirm"
        >
          {{ t('确定导入') }}
        </bk-button>
        <bk-button @click="handleCancel">
          {{ t('取消') }}
        </bk-button>
      </main>
    </footer>
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { computed, ref, useTemplateRef } from 'vue';
import FormCard from '@/components/form-card.vue';
import { useCommon } from '@/store';
import Cookie from 'js-cookie';
import ImportedResources from '@/views/import-export/components/imported-resources.vue';
import { useRouter } from 'vue-router';
import { importEtcdResources } from '@/http/gateway-sync-data';
import { InfoBox, Message } from 'bkui-vue';

interface IImportedResources {
  add?: any[],
  update?: any[],
}

const { t } = useI18n();
const router = useRouter();
const commonStore = useCommon();

const { BK_DASHBOARD_URL } = window;
const CSRFToken = Cookie.get(window.BK_DASHBOARD_CSRF_COOKIE_NAME);

const currentStep = ref(1);
const resources = ref<IImportedResources>({});
const importedResourcesRef = useTemplateRef('importedResourcesRef');

const steps = [
  {
    title: t('上传文件'),
  },
  {
    title: t('资源信息确认'),
  },
];

const uploadUrl = computed(() => {
  return `${BK_DASHBOARD_URL}/gateways/${commonStore.gatewayId}/unify_op/resources/upload/`;
});

const handleResCode = (res: { data: { add: Record<string, any[]>, update: Record<string, any[]> } }) => {
  if (res.data?.add) {
    resources.value.add = Object.values(res.data.add)
      .reduce((acc, cur) => [...acc, ...cur], []);
  }
  if (res.data?.update) {
    resources.value.update = Object.values(res.data.update)
      .reduce((acc, cur) => [...acc, ...cur], []);
  }
  return !!res.data;
};

const handleUploadDone = () => {
  currentStep.value = 2;
};

const handleConfirm = async () => {
  InfoBox({
    title: t('确认导入？'),
    confirmText: t('导入'),
    cancelText: t('取消'),
    onConfirm: async () => {
      const resources = importedResourcesRef.value.getValue();
      await importEtcdResources({ resources });
      Message({
        theme: 'success',
        message: t('导入成功'),
      });
      router.replace({ name: 'gateway-sync-data' });
    },
  });
};

const handleBack = () => {
  currentStep.value = 1;
};

const handleCancel = () => {
  router.replace({ name: 'import-export' });
};
</script>

<style lang="scss" scoped>
.upload-page-wrapper {
  position: relative;
  overflow: hidden;
  height: 100%;

  .page-header {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 52px;
    border-bottom: 1px solid #dcdee5;
    background-color: #ffffff;

    .header-time-line {
      width: 400px;
    }
  }

  .page-content {
    overflow: hidden;
    max-height: calc(100vh - 155px);
    padding: 20px 24px 0 24px;

    .drop-zone-wrapper {
      padding-inline: 56px;

      :deep(.bk-upload-trigger.bk-upload-trigger--draggable) {
        height: 120px;
      }

      :deep(.bk-upload-trigger.bk-upload-trigger--draggable > div > span > svg) {
        width: 56px !important;
        height: 50px !important;
      }
    }
  }

  .page-actions-wrap {
    position: sticky;
    z-index: 100;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    align-items: center;
    height: 52px;
    padding-left: 24px;
    border-top: 1px solid #dcdee5;
    background: #ffffff;

    .page-actions {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }
}
</style>
