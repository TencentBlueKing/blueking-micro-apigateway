package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("ConsumerGroup", func() {
	var consumerGroup model.ConsumerGroup

	BeforeEach(func() {
		consumerGroup = model.ConsumerGroup{
			Name: "test-consumer-group",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should set id and name into the Config", func() {
			err := consumerGroup.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]interface{}
			err = json.Unmarshal(consumerGroup.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("test-id"))
			Expect(configMap["name"]).To(Equal("test-consumer-group"))
		})

		It("should do nothing if Name is empty", func() {
			consumerGroup.Name = ""
			consumerGroup.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
				}`)),
			}
			err := consumerGroup.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]interface{}, 1)

			err = json.Unmarshal(consumerGroup.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("name"))
		})
	})
})
