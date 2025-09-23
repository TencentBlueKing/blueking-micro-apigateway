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
		It("should set id, username, and group_id into the Config", func() {
			err := consumer.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]interface{}
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

			configMap := make(map[string]interface{}, 1)

			err = json.Unmarshal(consumer.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("username"))
		})

		It("should delete group_id from the Config if GroupID is empty", func() {
			consumer.GroupID = ""
			consumer.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"group_id": "test-group-id"
				}`)),
			}
			err := consumer.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]interface{}, 1)

			err = json.Unmarshal(consumer.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("group_id"))
		})
	})
})
