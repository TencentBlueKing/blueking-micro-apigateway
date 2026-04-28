package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("StreamRoute", func() {
	var streamRoute model.StreamRoute

	BeforeEach(func() {
		streamRoute = model.StreamRoute{
			Name:       "stream-route-a",
			ServiceID:  "service-a",
			UpstreamID: "upstream-a",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "stream-route-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should preserve stored config and explicitly restore stream-route read fields", func() {
			err := streamRoute.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(streamRoute.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())

			streamRoute.ResourceCommonModel.NameValue = streamRoute.Name
			streamRoute.ResourceCommonModel.ServiceIDValue = streamRoute.ServiceID
			streamRoute.ResourceCommonModel.UpstreamIDValue = streamRoute.UpstreamID
			err = streamRoute.ResourceCommonModel.RestoreConfigForRead("stream_route")
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(streamRoute.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("stream-route-id"))
			Expect(configMap["name"]).To(Equal("stream-route-a"))
			Expect(configMap["service_id"]).To(Equal("service-a"))
			Expect(configMap["upstream_id"]).To(Equal("upstream-a"))
		})

		It("should leave legacy relation fields untouched when config already carries them", func() {
			streamRoute.ServiceID = ""
			streamRoute.UpstreamID = ""
			streamRoute.ResourceCommonModel = model.ResourceCommonModel{
				ID: "stream-route-id",
				Config: datatypes.JSON([]byte(`{
					"service_id":"service-a",
					"upstream_id":"upstream-a"
				}`)),
			}
			err := streamRoute.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(streamRoute.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["service_id"]).To(Equal("service-a"))
			Expect(configMap["upstream_id"]).To(Equal("upstream-a"))
		})

		It("should preserve typed name and relations when config omits them", func() {
			streamRoute.ResourceCommonModel.NameValue = "typed-stream-route-name"
			streamRoute.ResourceCommonModel.ServiceIDValue = "typed-service-id"
			streamRoute.ResourceCommonModel.UpstreamIDValue = "typed-upstream-id"

			typedStreamRoute := streamRoute.ResourceCommonModel.ToResourceModel("stream_route").(*model.StreamRoute)
			Expect(typedStreamRoute.Name).To(Equal("typed-stream-route-name"))
			Expect(typedStreamRoute.ServiceID).To(Equal("typed-service-id"))
			Expect(typedStreamRoute.UpstreamID).To(Equal("typed-upstream-id"))
		})
	})
})
