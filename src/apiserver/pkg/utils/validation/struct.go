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

package validation

import validator "github.com/go-playground/validator/v10"

type bizStructValidator struct {
	slz            interface{}
	validateFunCtx validator.StructLevelFuncCtx
	tagTrans       map[string]string
}

var bizStructValidators []bizStructValidator

// AddBizStructValidator 新增业务 SLZ 验证器
func AddBizStructValidator(slz interface{}, validatorFunCtx validator.StructLevelFuncCtx, tagTrans map[string]string) {
	structValidator := bizStructValidator{
		validateFunCtx: validatorFunCtx,
		slz:            slz,
		tagTrans:       tagTrans,
	}
	bizStructValidators = append(bizStructValidators, structValidator)
}

// registerBizStructValidator 注册业务 SLZ Struct 验证器
func registerBizStructValidator() {
	for _, v := range bizStructValidators {
		bizValidate.RegisterStructValidationCtx(v.validateFunCtx, v.slz)
		for tag, transMsg := range v.tagTrans {
			err := bizValidate.RegisterTranslation(
				tag,
				Trans,
				registerTranslator(tag, transMsg),
				translate,
			)
			if err != nil {
				panic(err)
			}
		}
	}
}
