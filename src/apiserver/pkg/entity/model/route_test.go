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
		It(
			"should preserve stored config and explicitly restore route read fields",
			func() {
				err := route.HandleConfig()
				Expect(err).NotTo(HaveOccurred())

				var configMap map[string]any
				err = json.Unmarshal(route.Config, &configMap)
				Expect(err).NotTo(HaveOccurred())

				route.ResourceCommonModel.NameValue = route.Name
				route.ResourceCommonModel.ServiceIDValue = route.ServiceID
				route.ResourceCommonModel.UpstreamIDValue = route.UpstreamID
				route.ResourceCommonModel.PluginConfigIDValue = route.PluginConfigID
				err = route.ResourceCommonModel.RestoreConfigForRead("route")
				Expect(err).NotTo(HaveOccurred())
				err = json.Unmarshal(route.Config, &configMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(configMap["id"]).To(Equal("test-id"))
				Expect(configMap["name"]).To(Equal("test-route"))
				Expect(configMap["service_id"]).To(Equal("test-service-id"))
				Expect(configMap["upstream_id"]).To(Equal("test-upstream-id"))
				Expect(configMap["plugin_config_id"]).To(Equal("test-plugin-config-id"))
			},
		)

		It(
			"should leave legacy echoed relation fields untouched in stored Config",
			func() {
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

				configMap := make(map[string]any, 3)

				err = json.Unmarshal(route.Config, &configMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(configMap["service_id"]).To(Equal("test-service-id"))
				Expect(configMap["plugin_config_id"]).To(Equal("test-plugin-config-id"))
				Expect(configMap["upstream_id"]).To(Equal("test-upstream-id"))
			},
		)

		It("should preserve name and relations from typed fields even when config omits them", func() {
			route.ResourceCommonModel.NameValue = "typed-route-name"
			route.ResourceCommonModel.ServiceIDValue = "typed-service-id"
			route.ResourceCommonModel.UpstreamIDValue = "typed-upstream-id"
			route.ResourceCommonModel.PluginConfigIDValue = "typed-plugin-config-id"

			typedRoute := route.ResourceCommonModel.ToResourceModel("route").(*model.Route)
			Expect(typedRoute.Name).To(Equal("typed-route-name"))
			Expect(typedRoute.ServiceID).To(Equal("typed-service-id"))
			Expect(typedRoute.UpstreamID).To(Equal("typed-upstream-id"))
			Expect(typedRoute.PluginConfigID).To(Equal("typed-plugin-config-id"))
		})
	})
})
