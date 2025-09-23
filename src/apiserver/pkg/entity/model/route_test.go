package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("Route", func() {
	var route model.Route

	BeforeEach(func() {
		route = model.Route{
			Name:           "test-route",
			ServiceID:      "test-service-id",
			UpstreamID:     "test-upstream-id",
			PluginConfigID: "test-plugin-config-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should set id, name, service_id, upstream_id, and plugin_config_id into the Config", func() {
			err := route.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]interface{}
			err = json.Unmarshal(route.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("test-id"))
			Expect(configMap["name"]).To(Equal("test-route"))
			Expect(configMap["service_id"]).To(Equal("test-service-id"))
			Expect(configMap["upstream_id"]).To(Equal("test-upstream-id"))
			Expect(configMap["plugin_config_id"]).To(Equal("test-plugin-config-id"))
		})

		It("should delete service_id, plugin_config_id, and upstream_id from the Config if they are empty", func() {
			route.ServiceID = ""
			route.PluginConfigID = ""
			route.UpstreamID = ""
			route.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"service_id": "test-service-id",
					"plugin_config_id": "test-plugin-config-id",
					"upstream_id": "test-upstream-id"
				}`)),
			}
			err := route.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]interface{}, 3)

			err = json.Unmarshal(route.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("service_id"))
			Expect(configMap).NotTo(HaveKey("plugin_config_id"))
			Expect(configMap).NotTo(HaveKey("upstream_id"))
		})
	})
})
