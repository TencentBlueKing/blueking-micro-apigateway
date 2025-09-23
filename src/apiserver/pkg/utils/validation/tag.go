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

import (
	"github.com/gin-gonic/gin/binding"
	validator "github.com/go-playground/validator/v10"

	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

type bizFiledTagValidator struct {
	ValidateFunCtx validator.FuncCtx
	ValidateFunc   validator.Func
	TagName        string
	TransMsg       string
}

var bizFiledTagValidators []bizFiledTagValidator

// AddBizFieldTagValidatorWithCtx 注册业务字段验证器
func AddBizFieldTagValidatorWithCtx(tagName string, validatorFunCtx validator.FuncCtx, transMsg string) {
	bizValidator := bizFiledTagValidator{
		ValidateFunCtx: validatorFunCtx,
		TagName:        tagName,
		TransMsg:       transMsg,
	}
	bizFiledTagValidators = append(bizFiledTagValidators, bizValidator)
}

// AddBizFieldTagValidator 注册业务字段验证器
func AddBizFieldTagValidator(tagName string, validatorFun validator.Func, transMsg string) {
	bizValidator := bizFiledTagValidator{
		ValidateFunc: validatorFun,
		TagName:      tagName,
		TransMsg:     transMsg,
	}
	bizFiledTagValidators = append(bizFiledTagValidators, bizValidator)
}

func registerBizFieldTagValidator() {
	bindValidator, _ := binding.Validator.Engine().(*validator.Validate)
	for _, v := range bizFiledTagValidators {
		if v.ValidateFunCtx != nil && bizValidate != nil {
			err := bizValidate.RegisterValidationCtx(v.TagName, v.ValidateFunCtx)
			if err != nil {
				log.Fatal("init validator error: ", err)
				return
			}
			err = bizValidate.RegisterTranslation(
				v.TagName,
				Trans,
				registerTranslator(v.TagName, v.TransMsg),
				translate,
			)
			if err != nil {
				log.Fatal("init translator error: ", err)
				return
			}
		}

		if v.ValidateFunc != nil && bindValidator != nil {
			err := bindValidator.RegisterValidation(v.TagName, v.ValidateFunc)
			if err != nil {
				log.Fatal("init validator error: ", err)
			}
			err = bindValidator.RegisterTranslation(
				v.TagName,
				Trans,
				registerTranslator(v.TagName, v.TransMsg),
				translate,
			)
			if err != nil {
				log.Fatal("init translator error: ", err)
			}
		}
	}
}
