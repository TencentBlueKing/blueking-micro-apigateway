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

package idx

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestGetFlakeUid(t *testing.T) {
	assert.True(t, len(GenResourceID(constant.Route)) < 64)
}

func TestGetResourceTypeFromID(t *testing.T) {
	tests := []struct {
		id       string
		expected constant.APISIXResource
	}{
		{"bk.r.someEncodedID", constant.Route},
		{"bk.u.someEncodedID", constant.Upstream},
		{"bk.s.someEncodedID", constant.Service},
		{"bk.c.someEncodedID", constant.Consumer},
		{"bk.cg.someEncodedID", constant.ConsumerGroup},
		{"bk.gr.someEncodedID", constant.GlobalRule},
		{"bk.pc.someEncodedID", constant.PluginConfig},
		{"bk.pm.someEncodedID", constant.PluginMetadata},
		{"bk.pb.someEncodedID", constant.Proto},
		{"bk.ss.someEncodedID", constant.SSL},
		{"bk.sr.someEncodedID", constant.StreamRoute},
		{"bk.sss.sssss", ""}, // 测试无效ID
		{"bk", ""},           // 测试无效ID
	}

	for _, test := range tests {
		result := GetResourceTypeFromID(test.id)
		if result != test.expected {
			t.Errorf("GetPrefixFromID(%s) = %s; want %s", test.id, result, test.expected)
		}
	}
}
