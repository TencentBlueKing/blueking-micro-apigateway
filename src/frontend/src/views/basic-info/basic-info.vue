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
  <div class="basic-info-wrapper">
    <bk-loading :loading="basicInfoDetailLoading">
      <section class="header-info">
        <div
          class="header-info-left">
          <span class="name">{{ basicInfoData?.name?.[0]?.toUpperCase() }}</span>
        </div>
        <div class="header-info-right">
          <div class="header-info-name">
            <span class="name">{{ basicInfoData.name }}</span>
          </div>
          <div class="header-info-description">
            <!-- <GateWaysEditTextarea
              field="description"
              width="600px"
              :placeholder="t('请输入描述')"
              :content="basicInfoData.description"
              @on-change="(e:Record<string, any>) => handleInfoChange(e)"
            /> -->
            {{ basicInfoData.description }}
          </div>
          <div class="header-info-button">
            <bk-button @click="showEditDialog" class="operate-btn">
              {{ t('编辑') }}
            </bk-button>
            <bk-button @click="handleDelete" class="operate-btn">
              {{ t('删除') }}
            </bk-button>
          </div>
        </div>
      </section>
      <section class="basic-info-detail">
        <div class="basic-info-detail-item">
          <div class="detail-item-title">{{ t('基础信息') }}</div>
          <div class="detail-item-content">
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('创建人')}：` }}</div>
              <div class="value">
                <span>{{ basicInfoData.creator || '--' }}</span>
              </div>
            </div>
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('创建时间')}：` }}</div>
              <div class="value">
                <span class="link">
                  {{ dayjs.unix(basicInfoData.created_at).format('YYYY-MM-DD HH:mm:ss Z') || '--' }}
                </span>
              </div>
            </div>
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('维护人员')}：` }}</div>
              <div class="value">
                <!-- <GateWaysEditMemberSelector
                  mode="edit"
                  width="600px"
                  field="maintainers"
                  :is-required="true"
                  :placeholder="t('请选择维护人员')"
                  :content="basicInfoData.maintainers"
                  :is-error-class="'maintainers-error-tip'"
                  :error-value="t('维护人员不能为空')"
                  @on-change="(e:Record<string, any>) => handleInfoChange(e)"
                /> -->
                {{ basicInfoData.maintainers?.join(', ') }}
              </div>
            </div>
          </div>
        </div>
        <div class="basic-info-detail-item">
          <div class="detail-item-title">{{ t('APISIX 信息') }}</div>
          <div class="detail-item-content">
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('APISIX 类型')}：` }}</div>
              <div class="value">
                <span class="link">{{ basicInfoData?.apisix?.type || '--' }}</span>
              </div>
            </div>
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('APISIX 版本')}：` }}</div>
              <div class="value">
                <span class="link">{{ basicInfoData?.apisix?.version || '--' }}</span>
              </div>
            </div>
            <!-- <div class="detail-item-content-item">
              <div class="label">{{ `${t('模式')}：` }}</div>
              <div class="value">
                <span class="link">{{ basicInfoData.mode === 1 ? t('直管') : t('纳管') }}</span>
              </div>
            </div> -->
          </div>
        </div>
        <div class="basic-info-detail-item">
          <div class="detail-item-title">{{ t('etcd 信息') }}</div>
          <div class="detail-item-content">
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('etcd 地址')}：` }}</div>
              <div class="value">
                <span class="link">{{ basicInfoData?.etcd?.endpoints?.join('; ') || '--' }}</span>
              </div>
            </div>
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('etcd 前缀')}：` }}</div>
              <div class="value">
                <span class="link">{{ basicInfoData?.etcd?.prefix || '--' }}</span>
              </div>
            </div>

            <div v-show="basicInfoData?.etcd?.schema_type === 'http'">
              <div class="detail-item-content-item">
                <div class="label">{{ `${t('etcd 用户名')}：` }}</div>
                <div class="value">
                  <span class="link">{{ basicInfoData?.etcd?.username || '--'}}</span>
                </div>
              </div>
              <div class="detail-item-content-item">
                <div class="label">{{ `${t('etcd 密码')}：` }}</div>
                <div class="value">
                  <span class="link">******</span>
                </div>
              </div>
            </div>

            <div v-show="basicInfoData?.etcd?.schema_type === 'https'">
              <div class="detail-item-content-item">
                <div class="label">{{ `${t('CACert')}：` }}</div>
                <div class="value">
                  <span class="link">******</span>
                </div>
              </div>
              <div class="detail-item-content-item">
                <div class="label">{{ `${t('Cert')}：` }}</div>
                <div class="value">
                  <span class="link">******</span>
                </div>
              </div>
              <div class="detail-item-content-item">
                <div class="label">{{ `${t('私钥')}：` }}</div>
                <div class="value">
                  <span class="link">******</span>
                </div>
              </div>
            </div>

          </div>
        </div>
        <div class="basic-info-detail-item">
          <div class="detail-item-title">{{ t('其他') }}</div>
          <div class="detail-item-content">
            <div class="detail-item-content-item">
              <div class="label">{{ `${t('只读模式')}：` }}</div>
              <div class="value">
                <span>{{ basicInfoData.read_only ? t('是') : t('否') }}</span>
              </div>
            </div>
          </div>
        </div>
      </section>
    </bk-loading>
  </div>

  <bk-dialog
    width="540"
    :is-show="delGatewayDialog.isShow"
    :title="t(`确认删除网关【${basicInfoData.name}】？`)"
    :theme="'primary'"
    :loading="delGatewayDialog.loading"
    @closed="delGatewayDialog.isShow = false">
    <div class="ps-form">
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div class="form-tips">
        {{ t('请完整输入') }} <span class="gateway-del-tips">{{ basicInfoData.name }}</span> {{ t('来确认删除网关！') }}
      </div>
      <div class="mt15">
        <bk-input v-model="confirmGatewayName"></bk-input>
      </div>
    </div>
    <template #footer>
      <bk-button theme="primary" :disabled="!delGatewayDisabled" @click="confirmDelGateway">
        {{ t('确定') }}
      </bk-button>
      <bk-button
        @click="delGatewayDialog.isShow = false">
        {{ t('取消') }}
      </bk-button>
    </template>
  </bk-dialog>

  <create ref="createRef" @done="getBasicInfo()" />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useCommon } from '@/store';
import { IGatewayItem } from '@/types';
import dayjs from 'dayjs';
import { deleteGateway, getGatewaysDetail } from '@/http';
// @ts-ignore
import Create from '@/views/gateway/create.vue';
import { IDialog } from '@/types/common';
// @ts-ignore
// import GateWaysEditTextarea from '@/components/blur-edit-textarea.vue';
// @ts-ignore
// import GateWaysEditMemberSelector from '@/components/blur-edit-member.vue';

const router = useRouter();
const { t } = useI18n();
const common = useCommon();

const gatewayId = ref(common.gatewayId);
const basicInfoDetailLoading = ref(false);
const basicInfoData = ref<IGatewayItem>({});
const createRef = ref<InstanceType<typeof Create>>();

const confirmGatewayName = ref<string>('');
const delGatewayDialog = ref<IDialog>({
  isShow: false,
  loading: false,
});

const getBasicInfo = async () => {
  try {
    const res = await getGatewaysDetail(gatewayId.value);
    basicInfoData.value = Object.assign({}, res);
    common.setCurGatewayData(res);
  } catch (e) {
    console.error(e);
  }
};
getBasicInfo();

const showEditDialog = () => {
  createRef.value?.show({
    ...basicInfoData.value,
    apisix_type: basicInfoData.value.apisix.type,
    apisix_version: basicInfoData.value.apisix.version,
    etcd_endpoints: basicInfoData.value.etcd.endpoints,
    etcd_password: basicInfoData.value.etcd.password,
    etcd_prefix: basicInfoData.value.etcd.prefix,
    etcd_username: basicInfoData.value.etcd.username,
    etcd_schema_type: basicInfoData.value.etcd.schema_type,
    etcd_cert_key: basicInfoData.value.etcd.cert_key,
    etcd_cert_cert: basicInfoData.value.etcd.cert_cert,
    etcd_ca_cert: basicInfoData.value.etcd.ca_cert,
  });
};

// const handleInfoChange = async (payload: Record<string, string>) => {
//   const params = {
//     ...basicInfoData.value,
//     ...payload,
//   };
//   await updateGateways(gatewayId.value, params);
//   basicInfoData.value = Object.assign(basicInfoData.value, params);
//   Message({
//     message: t('编辑成功'),
//     theme: 'success',
//     width: 'auto',
//   });
// };

const delGatewayDisabled = computed(() => {
  return basicInfoData.value.name === confirmGatewayName.value;
});

const handleDelete = () => {
  delGatewayDialog.value.isShow = true;
  confirmGatewayName.value = '';
};

const confirmDelGateway = async () => {
  try {
    await deleteGateway(gatewayId.value);

    Message({
      theme: 'success',
      message: t('删除成功'),
      width: 'auto',
    });

    delGatewayDialog.value.isShow = false;
    confirmGatewayName.value = '';

    setTimeout(() => {
      router.push({
        name: 'root',
      });
    }, 300);
  } catch (e) {
    console.error(e);
  }
};

</script>

<style lang="scss" scoped>
.basic-info-wrapper {
  padding: 24px;
  font-size: 12px;

  .header-info {
    padding: 24px;
    background: #ffffff;
    box-shadow: 0 2px 4px 0 #1919290d;
    display: flex;

    &-left {
      width: 80px;
      height: 80px;
      background: #f0f5ff;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;

      .name {
        font-weight: 700;
        font-size: 40px;
        color: #3a84ff;
      }

      &-disabled {
        background: #F0F1F5;

        .name {
          color: #C4C6CC;
        }
      }
    }

    &-right {
      width: calc(100% - 50px);
      padding: 0 16px;

      .header-info-name {
        display: flex;

        .name {
          font-weight: 700;
          font-size: 16px;
          color: #313238;
        }

        .header-info-tag {
          display: flex;
          margin-left: 8px;
          font-size: 12px;
          .bk-tag {
            margin: 2px 4px 2px 0;
          }

          .website {
            background-color: #EDF4FF;
            color: #3a84ff;
            padding: 8px;
          }

          .vip {
            background-color: #FFF1DB;
            color: #FE9C00;
          }

          .enabling {
            background-color: #E4FAF0;
            color: #14A568;
          }

          .deactivated {
            background-color: #F0F1F5;
            color: #63656E;
          }

          .icon-ag-yiqiyong,
          .icon-ag-minus-circle {
            font-size: 14px;
          }
        }
      }

      .header-info-description {
        margin-top: 8px;
        margin-bottom: 23px;
      }
      .header-info-button {
        display: flex;

        .operate-btn {
          min-width: 88px;
          margin-right: 8px;
        }

        .deactivate-btn {
          &:hover {
            background-color: #ff5656;
            border-color: #ff5656;
            color: #ffffff;
          }
        }
      }
    }
  }

  .basic-info-detail {
    padding: 24px;
    margin-top: 16px;
    background: #ffffff;
    box-shadow: 0 2px 4px 0 #1919290d;

    &-item {
      &:not(&:first-child) {
        padding-top: 40px;
      }

      .detail-item-title {
        font-weight: 700;
        font-size: 14px;
        color: #313238;
      }

      .detail-item-content {
        padding-left: 100px;
        padding-top: 24px;

        &-item {
          display: flex;
          align-items: center;
          line-height: 32px;

          .label {
            color: #63656E;
            min-width: 60px;
            text-align: right;
            &.w0 {
              min-width: 0px;
            }
          }

          .value {
            display: flex;
            align-items: center;
            vertical-align: middle;
            margin-left: 8px;
            flex: 1;
            color: #313238;

            .icon-ag-copy-info {
              margin-left: 3px;
              padding: 3px;
              color: #3A84FF;
              cursor: pointer;
            }

            .link {
              margin-right: 14px;
            }

            .more-detail {
              color: #3A84FF;
              cursor: pointer;
            }

            .apigateway-icon {
              font-size: 16px;
              color: #979BA5;
              &:hover {
                color: #3A84FF;
                cursor: pointer;
              }
              &.icon-ag-lock-fill1 {
                &:hover {
                  color: #979BA5;
                  cursor: default;
                }
              }
            }

            &.more-tip {
              display: flex;
              align-items: center;

              .icon-ag-info {
                color: #63656E;
                margin-right: 5px;
              }
            }

            &.public-key-content {
              min-width: 670px;
              height: 40px;
              line-height: 40px;
              background-color: #F5F7FA;
              color: #63656E;
              margin-left: 0;

              .value-icon-lock {
                width: 40px;
                background-color: #F0F1F5;
                border-radius: 2px 0 0 2px;
                text-align: center;
              }

              .value-public-key {
                width: calc(100% - 40px);
                display: flex;
                justify-content: space-between;
                padding: 0 12px;
              }
            }
          }
        }

      }
    }
  }
}
</style>

<style>
.gateway-del-tips {
  color: #c7254e;
  padding: 3px 4px;
  margin: 0;
  background-color: rgba(0,0,0,.04);
  border-radius: 3px;
}
</style>
