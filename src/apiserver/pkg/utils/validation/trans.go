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
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// Trans 定义一个全局翻译器 T
var Trans ut.Translator

// InitTrans 初始化翻译器
func InitTrans(locale string) (err error) {
	ginValidator, _ := binding.Validator.Engine().(*validator.Validate)
	validatorList := []*validator.Validate{ginValidator, bizValidate}
	for _, v := range validatorList {
		// 注册一个获取 json tag 的自定义方法
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器

		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		// uni := ut.New(zhT, zhT) 也是可以的
		uni := ut.New(enT, zhT, enT)

		// locale 通常取决于 http 请求头的 'Accept-LanguageCode'
		var ok bool
		// 也可以使用 uni.FindTranslator(...) 传入多个 locale 进行查找
		Trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 注册翻译器
		switch locale {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, Trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, Trans)
		default:
			err = enTranslations.RegisterDefaultTranslations(v, Trans)
		}
	}
	return err
}

// registerTranslator 为自定义字段添加翻译功能
func registerTranslator(tag, msg string) validator.RegisterTranslationsFunc {
	return func(trans ut.Translator) error {
		if err := trans.Add(tag, msg, false); err != nil {
			return err
		}
		return nil
	}
}

// translate 自定义字段的翻译方法
func translate(trans ut.Translator, fe validator.FieldError) string {
	msg, err := trans.T(fe.Tag(), fe.Field(), fmt.Sprintf("%s", fe.Value()))
	if err != nil {
		panic(fe.(error).Error())
	}
	return msg
}

func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.LastIndex(field, ".")+1:]] = err
	}
	return res
}

// TranslateToString ...
func TranslateToString(validateErrs validator.ValidationErrors) string {
	errorMap := removeTopStruct(validateErrs.Translate(Trans))
	res := ""
	for _, value := range errorMap {
		res += value + ". "
	}
	return res
}

// GetEnumTransMsgFromUint8KeyMap ...
func GetEnumTransMsgFromUint8KeyMap(enum map[uint8]string, onlyValue bool) string {
	msg := "{0}:{1} must be: "
	var enumItem []string
	for key, value := range enum {
		if onlyValue {
			enumItem = append(enumItem, fmt.Sprintf("%v", value))
		} else {
			enumItem = append(enumItem, fmt.Sprintf("%v-(%v)", key, value))
		}
	}
	return msg + strings.Join(enumItem, ";")
}

// GetEnumTransMsgFromStringKeyMap ...
func GetEnumTransMsgFromStringKeyMap(enum map[string]string, onlyValue bool) string {
	msg := "{0}:{1} must be: "
	var enumItem []string
	for key, value := range enum {
		if onlyValue {
			enumItem = append(enumItem, fmt.Sprintf("%v", value))
		} else {
			enumItem = append(enumItem, fmt.Sprintf("%v-(%v)", key, value))
		}
	}
	return msg + strings.Join(enumItem, ";")
}
