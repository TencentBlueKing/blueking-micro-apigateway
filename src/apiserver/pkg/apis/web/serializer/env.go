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

package serializer

// EnvVarsResponse ...
type EnvVarsResponse struct {
	Edition string   `json:"edition"`
	Links   LinkInfo `json:"links"`
}

type LinkInfo struct {
	BKGuideLink      string `json:"bk_guide_link"`      // 产品使用指南地址
	BKFeedBackLink   string `json:"bk_feed_back_link"`  // 产品反馈地址
	BKApigatewayLink string `json:"bk_apigateway_link"` // 蓝鲸共享网关地址
}
