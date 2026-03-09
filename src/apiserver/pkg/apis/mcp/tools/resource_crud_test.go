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

package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildBatchOperationResultPartialFailure(t *testing.T) {
	t.Parallel()

	failures := []batchOperationFailure{
		{
			ResourceID: "r1",
			Stage:      "delete",
			Error:      "failed to delete",
		},
	}

	result := buildBatchOperationResult(
		"Delete operation completed",
		"route",
		2,
		1,
		map[string]int{
			"hard_deleted_count":  1,
			"marked_delete_count": 0,
		},
		failures,
	)

	assert.Equal(t, 2, result["total_requested"])
	assert.Equal(t, 1, result["success_count"])
	assert.Equal(t, 1, result["failed_count"])
	assert.Equal(t, true, result["partial_success"])
	assert.Equal(t, "route", result["resource_type"])
	assert.Equal(t, 1, result["hard_deleted_count"])
	assert.Equal(t, 0, result["marked_delete_count"])
}

func TestBuildBatchOperationResultAllSuccess(t *testing.T) {
	t.Parallel()

	result := buildBatchOperationResult(
		"Revert operation completed",
		"service",
		3,
		3,
		map[string]int{
			"reverted_count": 3,
		},
		nil,
	)

	assert.Equal(t, 3, result["total_requested"])
	assert.Equal(t, 3, result["success_count"])
	assert.Equal(t, 0, result["failed_count"])
	assert.Equal(t, false, result["partial_success"])
	assert.Equal(t, 3, result["reverted_count"])
}
