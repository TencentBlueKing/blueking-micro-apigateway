package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("PluginConfig", func() {
	var pluginConfig model.PluginConfig

	BeforeEach(func() {
		pluginConfig = model.PluginConfig{
			Name: "test-plugin-config",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should strip echoed id and name from stored Config", func() {
			err := pluginConfig.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(pluginConfig.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("id"))
			Expect(configMap).NotTo(HaveKey("name"))

			err = pluginConfig.AfterFind(nil)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(pluginConfig.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("test-id"))
			Expect(configMap["name"]).To(Equal("test-plugin-config"))
		})
	})
})
