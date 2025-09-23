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

export interface ICreatePayload {
  apisix_type?: string
  apisix_version?: string
  description?: string
  etcd_endpoints?: string[]
  etcd_password?: string
  etcd_prefix?: string
  etcd_username?: string
  maintainers?: string[]
  mode?: number
  name?: string
  id?: number
  etcd_cert_key?: string
  etcd_cert_cert?: string
  etcd_ca_cert?: string
  etcd_schema_type?: string
  read_only?: boolean
}

interface IApisixType {
  type: string
  version: string
}

interface IEtcdType {
  endpoints: string[]
  prefix: string
  username?: string
  password?: string
  schema_type: string
  cert_key?: string
  cert_cert?: string
  ca_cert?: string
}

interface ICountType {
  route: number
  service: number
  upstream: number
}

export interface IGatewayItem {
  apisix_type?: string
  apisix_version?: string
  created_at?: number
  creator?: string
  description?: string
  etcd_endpoints?: string[]
  etcd_password?: string
  etcd_prefix?: string
  etcd_username?: string
  id?: number
  maintainers?: string[]
  mode?: number
  name?: string
  updated_at?: number
  updater?: string
  is24HoursAgo?: boolean
  apisix?: IApisixType
  etcd?: IEtcdType
  count?: ICountType
  read_only?: boolean
}

export interface IConnectTest {
  etcd_endpoints: string[]
  etcd_prefix: string
  etcd_username?: string
  etcd_password?: string
  gateway_id?: number
  etcd_cert_key?: string
  etcd_cert_cert?: string
  etcd_ca_cert?: string
  etcd_schema_type?: string
}

export interface ICheckName {
  name: string
  id?: number
}
