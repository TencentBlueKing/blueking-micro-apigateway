package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("Proto", func() {
	var protoModel model.Proto

	BeforeEach(func() {
		protoModel = model.Proto{
			Name: "test.proto",
			ResourceCommonModel: model.ResourceCommonModel{
				ID: "proto-id",
				Config: datatypes.JSON([]byte(`{
					"content":"syntax = \"proto3\"; package demo; service Greeter { rpc SayHello (HelloRequest) returns (HelloReply) {} } message HelloRequest { string name = 1; } message HelloReply { string message = 1; }"
				}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should preserve stored config and explicitly restore proto read fields", func() {
			err := protoModel.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(protoModel.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())

			protoModel.ResourceCommonModel.NameValue = protoModel.Name
			err = protoModel.ResourceCommonModel.RestoreConfigForRead("proto")
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(protoModel.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("proto-id"))
			Expect(configMap["name"]).To(Equal("test.proto"))
		})
	})
})
