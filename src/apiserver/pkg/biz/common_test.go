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

package biz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

func TestParseOrderByExprList(t *testing.T) {
	ascFieldMap := map[string]field.Expr{
		"name": field.NewField("", "name").Asc(),
		"age":  field.NewField("", "age").Asc(),
	}
	descFieldMap := map[string]field.Expr{
		"name": field.NewField("", "name").Desc(),
		"age":  field.NewField("", "age").Desc(),
	}

	tests := []struct {
		name     string
		orderBy  string
		expected int
	}{
		{
			name:     "test_empty",
			orderBy:  "",
			expected: 0,
		},
		{
			name:     "test_asc",
			orderBy:  "name:asc",
			expected: 1,
		},
		{
			name:     "test_desc",
			orderBy:  "age:desc",
			expected: 1,
		},
		{
			name:     "test_asc_and_desc",
			orderBy:  "name:asc,age:desc",
			expected: 2,
		},
		{
			name:     "test_invalid_key",
			orderBy:  "invalid:asc",
			expected: 0,
		},
		{
			name:     "test_invalid_value",
			orderBy:  "name:invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseOrderByExprList(ascFieldMap, descFieldMap, tt.orderBy)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestGetResourcesLabels(t *testing.T) {
	route := data.Route2WithNoRelationResource(gatewayInfo, constant.ResourceStatusCreateDraft)
	// 创建资源
	if err := CreateRoute(gatewayCtx, *route); err != nil {
		t.Errorf("CreateRoute error = %v", err)
		return
	}
	labels, err := GetResourcesLabels(gatewayCtx, constant.Route)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(labels))
}
