package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("PluginMetadata", func() {
	var pluginMetadata model.PluginMetadata

	BeforeEach(func() {
		pluginMetadata = model.PluginMetadata{
			Name: "test-plugin-metadata",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should strip echoed id and name from stored Config", func() {
			err := pluginMetadata.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(pluginMetadata.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("id"))
			Expect(configMap).NotTo(HaveKey("name"))

			err = pluginMetadata.AfterFind(nil)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(pluginMetadata.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("test-plugin-metadata"))
			Expect(configMap["name"]).To(Equal("test-plugin-metadata"))
		})

		It("should remove legacy echoed plugin metadata identity from stored Config", func() {
			pluginMetadata.ResourceCommonModel = model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{"id":"other-id"}`)),
			}
			err := pluginMetadata.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(pluginMetadata.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("id"))
		})
	})
})
