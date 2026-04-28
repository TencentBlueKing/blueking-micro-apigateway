/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
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

package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestApplyImportIgnoreFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		imported     string
		existing     string
		ignoreFields []string
		want         string
	}{
		{
			name:         "overlay top level field from existing config",
			imported:     `{"name":"route-a","desc":"new-desc","plugins":{}}`,
			existing:     `{"name":"route-a","desc":"old-desc","plugins":{"limit-count":{"count":1}}}`,
			ignoreFields: []string{"desc"},
			want:         `{"name":"route-a","desc":"old-desc","plugins":{}}`,
		},
		{
			name:         "overlay nested field from existing config",
			imported:     `{"plugins":{"limit-count":{"count":10,"time_window":60}}}`,
			existing:     `{"plugins":{"limit-count":{"count":1,"time_window":120}}}`,
			ignoreFields: []string{"plugins.limit-count.count"},
			want:         `{"plugins":{"limit-count":{"count":1,"time_window":60}}}`,
		},
		{
			name:         "ignore missing field keeps imported config",
			imported:     `{"plugins":{}}`,
			existing:     `{"name":"route-a"}`,
			ignoreFields: []string{"plugins.limit-count"},
			want:         `{"plugins":{}}`,
		},
		{
			name:         "partial missing fields only overlays existing fields",
			imported:     `{"desc":"new-desc","plugins":{"limit-count":{"count":10}}}`,
			existing:     `{"desc":"old-desc"}`,
			ignoreFields: []string{"desc", "plugins.limit-count.count"},
			want:         `{"desc":"old-desc","plugins":{"limit-count":{"count":10}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyImportIgnoreFields(
				json.RawMessage(tt.imported),
				datatypes.JSON([]byte(tt.existing)),
				tt.ignoreFields,
			)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}
