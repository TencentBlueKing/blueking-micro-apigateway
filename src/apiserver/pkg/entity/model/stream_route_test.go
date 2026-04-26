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
		It("should strip echoed id, name, service_id, and upstream_id from stored Config", func() {
			err := streamRoute.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(streamRoute.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("id"))
			Expect(configMap).NotTo(HaveKey("name"))
			Expect(configMap).NotTo(HaveKey("service_id"))
			Expect(configMap).NotTo(HaveKey("upstream_id"))

			err = streamRoute.AfterFind(nil)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(streamRoute.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("stream-route-id"))
			Expect(configMap["name"]).To(Equal("stream-route-a"))
			Expect(configMap["service_id"]).To(Equal("service-a"))
			Expect(configMap["upstream_id"]).To(Equal("upstream-a"))
		})

		It("should delete empty relation fields from the Config", func() {
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
			Expect(configMap).NotTo(HaveKey("service_id"))
			Expect(configMap).NotTo(HaveKey("upstream_id"))
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
