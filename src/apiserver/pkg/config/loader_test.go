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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadServiceConfigTAPISIXVisibility(t *testing.T) {
	previousValue, wasSet := os.LookupEnv("ENABLE_TAPISIX")
	require.NoError(t, os.Unsetenv("ENABLE_TAPISIX"))
	t.Cleanup(func() {
		if wasSet {
			require.NoError(t, os.Setenv("ENABLE_TAPISIX", previousValue))
			return
		}
		require.NoError(t, os.Unsetenv("ENABLE_TAPISIX"))
	})

	defaultConfig, err := loadServiceConfigFromEnv()
	require.NoError(t, err)
	assert.False(t, defaultConfig.EnableTAPISIX)

	require.NoError(t, os.Setenv("ENABLE_TAPISIX", "true"))
	enabledConfig, err := loadServiceConfigFromEnv()
	require.NoError(t, err)
	assert.True(t, enabledConfig.EnableTAPISIX)
}
