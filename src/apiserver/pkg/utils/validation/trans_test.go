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
package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
)

var _ = Describe("Translation", func() {
	BeforeEach(func() {
		validation.Trans = nil
	})

	Describe("InitTrans", func() {
		It("should initialize translator for en", func() {
			err := validation.InitTrans("en")
			Expect(err).To(BeNil())
			Expect(validation.Trans).NotTo(BeNil())
		})

		It("should initialize translator for zh", func() {
			err := validation.InitTrans("zh")
			Expect(err).To(BeNil())
			Expect(validation.Trans).NotTo(BeNil())
		})

		It("should return error for unknown locale", func() {
			err := validation.InitTrans("fr")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("uni.GetTranslator(fr) failed"))
		})
	})

	Describe("Enum Translation Messages", func() {
		Context("GetEnumTransMsgFromUint8KeyMap", func() {
			It("should generate translation message for uint8 enum", func() {
				enum := map[uint8]string{
					1: "Active",
					2: "Inactive",
				}
				msg := validation.GetEnumTransMsgFromUint8KeyMap(enum, false)
				Expect(msg).To(ContainSubstring("1-(Active)"))
				Expect(msg).To(ContainSubstring("2-(Inactive)"))
				Expect(msg).To(HavePrefix("{0}:{1} must be: "))
			})

			It("should generate translation message with only values", func() {
				enum := map[uint8]string{
					1: "Active",
					2: "Inactive",
				}
				msg := validation.GetEnumTransMsgFromUint8KeyMap(enum, true)
				Expect(msg).To(ContainSubstring("Active"))
				Expect(msg).To(ContainSubstring("Inactive"))
				Expect(msg).To(HavePrefix("{0}:{1} must be: "))
			})
		})

		Context("GetEnumTransMsgFromStringKeyMap", func() {
			It("should generate translation message for string enum", func() {
				enum := map[string]string{
					"active":   "Active",
					"inactive": "Inactive",
				}
				msg := validation.GetEnumTransMsgFromStringKeyMap(enum, false)
				Expect(msg).To(ContainSubstring("active-(Active)"))
				Expect(msg).To(ContainSubstring("inactive-(Inactive)"))
				Expect(msg).To(HavePrefix("{0}:{1} must be: "))
			})

			It("should generate translation message with only values", func() {
				enum := map[string]string{
					"active":   "Active",
					"inactive": "Inactive",
				}
				msg := validation.GetEnumTransMsgFromStringKeyMap(enum, true)
				Expect(msg).To(ContainSubstring("Active"))
				Expect(msg).To(ContainSubstring("Inactive"))
				Expect(msg).To(HavePrefix("{0}:{1} must be: "))
			})
		})
	})
})
