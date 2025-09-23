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

// Package validation ...
package validation

import (
	"context"

	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"
)

var bizValidate *validator.Validate

// BindAndValidate ...
func BindAndValidate(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	return bizValidate.StructCtx(c.Request.Context(), obj)
}

// ValidateStruct ...
func ValidateStruct(ctx context.Context, obj interface{}) error {
	return bizValidate.StructCtx(ctx, obj)
}

// RegisterValidator ...
func RegisterValidator() {
	bizValidate = validator.New()
	_ = InitTrans("en")
	registerBizStructValidator()
	registerBizFieldTagValidator()
}
