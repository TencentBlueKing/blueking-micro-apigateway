package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("GlobalRule", func() {
	var globalRule model.GlobalRule

	BeforeEach(func() {
		globalRule = model.GlobalRule{
			Name: "test-global-rule",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should set id and name into the Config", func() {
			err := globalRule.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(globalRule.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("test-id"))
			Expect(configMap["name"]).To(Equal("test-global-rule"))
		})

		It("should do nothing if Name is empty", func() {
			globalRule.Name = ""
			globalRule.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
				}`)),
			}
			err := globalRule.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]any, 1)

			err = json.Unmarshal(globalRule.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("name"))
		})
	})
})
