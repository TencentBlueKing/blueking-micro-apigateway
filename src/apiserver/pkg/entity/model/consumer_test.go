package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("Consumer", func() {
	var consumer model.Consumer

	BeforeEach(func() {
		consumer = model.Consumer{
			Username: "test-username",
			GroupID:  "test-group-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should preserve stored config and explicitly restore consumer read fields", func() {
			err := consumer.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(consumer.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())

			consumer.ResourceCommonModel.NameValue = consumer.Username
			consumer.ResourceCommonModel.GroupIDValue = consumer.GroupID
			err = consumer.ResourceCommonModel.RestoreConfigForRead("consumer")
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(consumer.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("test-id"))
			Expect(configMap["username"]).To(Equal("test-username"))
			Expect(configMap["group_id"]).To(Equal("test-group-id"))
		})

		It("should do nothing if Username is empty", func() {
			consumer.Username = ""
			consumer.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
				}`)),
			}
			err := consumer.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]any, 1)

			err = json.Unmarshal(consumer.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("username"))
		})

		It("should leave legacy group_id untouched when config already carries it", func() {
			consumer.GroupID = ""
			consumer.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"group_id": "test-group-id"
				}`)),
			}
			err := consumer.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]any, 1)

			err = json.Unmarshal(consumer.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["group_id"]).To(Equal("test-group-id"))
		})

		It("should preserve username and group from typed fields even when config omits them", func() {
			consumer.ResourceCommonModel.NameValue = "typed-consumer-name"
			consumer.ResourceCommonModel.GroupIDValue = "typed-group-id"

			typedConsumer := consumer.ResourceCommonModel.ToResourceModel("consumer").(*model.Consumer)
			Expect(typedConsumer.Username).To(Equal("typed-consumer-name"))
			Expect(typedConsumer.GroupID).To(Equal("typed-group-id"))
		})
	})
})
