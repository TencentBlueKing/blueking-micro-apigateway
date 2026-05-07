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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

func TestGenResourceID(t *testing.T) {
	id := GenResourceID(constant.Route)
	assert.True(t, len(id) < 64)
	assert.Regexp(t, `^bk\.r\.[0-9a-z]+$`, id)
	assert.Equal(t, strings.ToLower(id), id)
}

func TestGenResourceIDCaseFoldedUnique(t *testing.T) {
	seen := make(map[string]string)
	for i := 0; i < 1000; i++ {
		id := GenResourceID(constant.Route)
		folded := strings.ToLower(id)
		previous, ok := seen[folded]
		if !assert.Falsef(t, ok, "case-folded duplicate resource ID: previous=%s current=%s", previous, id) {
			return
		}
		seen[folded] = id
	}
}
