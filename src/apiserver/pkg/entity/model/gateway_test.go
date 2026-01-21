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

package model_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("Gateway", func() {
	Describe("NormalizeEtcdPrefix", func() {
		DescribeTable("应该正确标准化 prefix",
			func(input, expected string) {
				result := model.NormalizeEtcdPrefix(input)
				Expect(result).To(Equal(expected))
			},
			Entry("空字符串", "", "/"),
			Entry("无斜杠前缀", "a-b", "a-b/"),
			Entry("已有斜杠结尾", "a-b/", "a-b/"),
			Entry("以斜杠开头", "/apisix", "/apisix/"),
			Entry("以斜杠开头且结尾", "/apisix/", "/apisix/"),
			Entry("多层路径", "/apisix/gateway1", "/apisix/gateway1/"),
		)
	})

	Describe("CheckEtcdPrefixConflict", func() {
		DescribeTable("应该正确检测 prefix 层级冲突",
			func(prefix1, prefix2 string, shouldConflict bool) {
				result := model.CheckEtcdPrefixConflict(prefix1, prefix2)
				Expect(result).To(Equal(shouldConflict))
			},
			// 完全相同的情况 - 冲突
			Entry("完全相同 - 无斜杠", "a-b", "a-b", true),
			Entry("完全相同 - 有斜杠", "a-b/", "a-b/", true),
			Entry("完全相同 - 一个有斜杠一个没有", "a-b", "a-b/", true),
			Entry("完全相同 - 多层路径", "/apisix/gateway1", "/apisix/gateway1/", true),

			// 层级前缀冲突 - 冲突（a/b 和 a/b/c 的情况）
			Entry("层级冲突 - p1 是 p2 的父路径", "/apisix", "/apisix/gateway1", true),
			Entry("层级冲突 - p2 是 p1 的父路径", "/apisix/gateway1", "/apisix", true),
			Entry("层级冲突 - a/b 和 a/b/c", "a/b", "a/b/c", true),
			Entry("层级冲突 - a/b/c 和 a/b", "a/b/c", "a/b", true),
			Entry("层级冲突 - 根路径与子路径", "/", "/apisix", true),

			// 名称相似但不冲突 - 这是允许的！
			Entry("允许 - a-b 和 a-b-test（不同名称）", "a-b", "a-b-test", false),
			Entry("允许 - a-b-test 和 a-b（不同名称）", "a-b-test", "a-b", false),
			Entry("允许 - gateway1 和 gateway10", "gateway1", "gateway10", false),
			Entry("允许 - /apisix-prod 和 /apisix-prod-backup", "/apisix-prod", "/apisix-prod-backup", false),

			// 完全不同的前缀 - 不冲突
			Entry("不同前缀 - gateway1 和 gateway2", "/gateway1", "/gateway2", false),
			Entry("不同前缀 - 同层级不同名", "/apisix/gw1", "/apisix/gw2", false),
			Entry("不同前缀 - 不同根路径", "/prod/gateway", "/test/gateway", false),
		)
	})

	Describe("Gateway.GetEtcdPrefixForList", func() {
		It("应该返回带斜杠结尾的 prefix", func() {
			gateway := &model.Gateway{
				EtcdConfig: model.EtcdConfig{
					EtcdConfig: base.EtcdConfig{
						Prefix: "a-b",
					},
				},
			}
			Expect(gateway.GetEtcdPrefixForList()).To(Equal("a-b/"))
		})

		It("已有斜杠结尾时应保持不变", func() {
			gateway := &model.Gateway{
				EtcdConfig: model.EtcdConfig{
					EtcdConfig: base.EtcdConfig{
						Prefix: "/apisix/",
					},
				},
			}
			Expect(gateway.GetEtcdPrefixForList()).To(Equal("/apisix/"))
		})
	})

	Describe("Gateway.GetEtcdResourcePrefix", func() {
		var gateway *model.Gateway

		BeforeEach(func() {
			gateway = &model.Gateway{
				EtcdConfig: model.EtcdConfig{
					EtcdConfig: base.EtcdConfig{
						Prefix: "/apisix",
					},
				},
			}
		})

		It("应该返回 routes 资源的正确 prefix", func() {
			prefix := gateway.GetEtcdResourcePrefix(constant.Route)
			Expect(prefix).To(Equal("/apisix/routes/"))
		})

		It("应该返回 services 资源的正确 prefix", func() {
			prefix := gateway.GetEtcdResourcePrefix(constant.Service)
			Expect(prefix).To(Equal("/apisix/services/"))
		})

		It("应该返回 upstreams 资源的正确 prefix", func() {
			prefix := gateway.GetEtcdResourcePrefix(constant.Upstream)
			Expect(prefix).To(Equal("/apisix/upstreams/"))
		})

		It("对于无效的资源类型应该返回基础 prefix", func() {
			prefix := gateway.GetEtcdResourcePrefix(constant.APISIXResource("invalid"))
			Expect(prefix).To(Equal("/apisix/"))
		})
	})
})
