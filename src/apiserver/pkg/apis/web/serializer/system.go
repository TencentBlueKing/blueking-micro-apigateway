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

import (
	"sort"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/schema"
)

// PluginListResponse ...
type PluginListResponse []*TypePluginInfo

// TypePluginInfo ...
type TypePluginInfo struct {
	Plugins []*schema.Plugin `json:"plugins"`
	Type    string           `json:"type"`
}

// SortPlugins 对 PluginListResponse 进行排序
func SortPlugins(plugins PluginListResponse) {
	// 首先根据 Type 对 PluginListResponse 进行排序
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Type < plugins[j].Type
	})

	// 然后对每个 TypePluginInfo 中的 Plugins 根据 Name 进行排序
	for _, typePluginInfo := range plugins {
		sort.Slice(typePluginInfo.Plugins, func(i, j int) bool {
			return typePluginInfo.Plugins[i].Name < typePluginInfo.Plugins[j].Name
		})
	}
}
