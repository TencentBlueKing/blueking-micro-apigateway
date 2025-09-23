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
  <div>
    <list-content-article>
      <template #title>{{ t('基本信息') }}</template>
      <list-content-row v-if="upstream.name" :label="t('名称')">{{ upstream.name }}</list-content-row>
      <list-content-row v-if="upstream.id" :label="t('ID')">{{ upstream.id }}</list-content-row>
      <list-content-row v-if="upstream.config.desc" :label="t('描述')">{{ upstream.config.desc }}</list-content-row>
      <list-content-row
        v-if="upstream.config.labels && Object.keys(upstream.config.labels).length" :label="t('标签')"
      >
        <tag-label :labels="upstream.config.labels" />
      </list-content-row>
      <!-- DTO 通用信息行 -->
      <list-content-rows-common :resource="upstream" />
      <list-content-row :label="t('负载均衡算法')">{{ upstream.config.type }}</list-content-row>
      <list-content-row :label="t('Host 请求头')">{{ upstream.config.pass_host }}</list-content-row>
      <list-content-row v-if="upstream.config.retries" :label="t('重试次数')">{{
        upstream.config.retries
      }}
      </list-content-row>
      <list-content-row v-if="upstream.config.retry_timeout" :label="t('重试超时时间')">
        {{ upstream.config.retry_timeout }}
      </list-content-row>
      <list-content-row :label="t('协议')">{{ upstream.config.scheme }}</list-content-row>
      <template v-if="upstream.config.tls">
        <list-content-row
          v-if="upstream.config.tls.client_cert && upstream.config.tls.client_key"
          label="TLS"
        >
          {{ t('已配置证书和私钥') }}
        </list-content-row>
        <list-content-row v-else label="TLS 关联证书">
          <bk-button
            text
            theme="primary"
            @click="handleResourceLinkClick('ssl', upstream.config.tls.client_cert_id)"
          >{{ upstream.config.tls.client_cert_id }}
          </bk-button>
        </list-content-row>
      </template>
      <template v-if="upstream.config.timeout">
        <list-content-row :label="t('连接超时(s)')">{{ upstream.config.timeout.connect }}</list-content-row>
        <list-content-row :label="t('发送超时(s)')">{{ upstream.config.timeout.send }}</list-content-row>
        <list-content-row :label="t('接收超时(s)')">{{ upstream.config.timeout.read }}</list-content-row>
      </template>
    </list-content-article>

    <list-content-article v-if="upstream.config.nodes">
      <template #title>{{ t('目标节点') }}</template>
      <template v-for="(node, index) in upstream.config.nodes" :key="index">
        <list-content-row :label="t('主机名')">{{ node.host }}</list-content-row>
        <list-content-row :label="t('端口')">{{ node.port }}</list-content-row>
        <list-content-row :label="t('权重')">{{ node.weight }}</list-content-row>
      </template>
    </list-content-article>

    <list-content-article v-if="upstream.config.service_name">
      <template #title>{{ t('服务发现') }}</template>
      <list-content-row :label="t('服务发现类型')">{{ upstream.config.discovery_type }}</list-content-row>
      <list-content-row :label="t('服务名称')">{{ upstream.config.service_name }}</list-content-row>
    </list-content-article>

    <list-content-article v-if="upstream.config.keepalive_pool">
      <template #title>{{ t('连接池') }}</template>
      <list-content-row :label="t('容量')">{{ upstream.config.keepalive_pool.size }}</list-content-row>
      <list-content-row :label="t('空闲超时时间')">{{ upstream.config.keepalive_pool.idle_timeout }}</list-content-row>
      <list-content-row :label="t('请求数量')">{{ upstream.config.keepalive_pool.requests }}</list-content-row>
    </list-content-article>

    <list-display-health-checks v-if="upstream.config.checks" :checks="upstream.config.checks" />
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { IUpstream } from '@/types/upstream';
import ListDisplayHealthChecks from '@/components/list-display/list-display-health-checks.vue';
import ListContentRow from '@/components/list-display/components/list-content-row.vue';
import ListContentArticle from '@/components/list-display/components/list-content-article.vue';
import { computed } from 'vue';
import ListContentRowsCommon from '@/components/list-display/components/list-content-rows-common.vue';
import TagLabel from '@/components/tag-label.vue';
import { useRouter } from 'vue-router';

interface IProps {
  resource: IUpstream
}

const { resource: upstreamProp } = defineProps<IProps>();

const { t } = useI18n();
const router = useRouter();

const upstream = computed<IUpstream>(() => {
  if (!upstreamProp.config) {
    return {
      config: upstreamProp,
    };
  }

  return upstreamProp;
});

const handleResourceLinkClick = (routeName: string, id: string) => {
  const to = router.resolve({ name: routeName, query: { id } });
  window.open(to.href);
};

</script>

<style lang="scss" scoped>

</style>
