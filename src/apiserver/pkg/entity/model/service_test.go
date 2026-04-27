package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

var _ = Describe("Service", func() {
	var service model.Service

	BeforeEach(func() {
		service = model.Service{
			Name:       "test-service",
			UpstreamID: "test-upstream-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}
	})

	Describe("HandleConfig", func() {
		It("should preserve stored config and explicitly restore service read fields", func() {
			err := service.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(service.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())

			service.ResourceCommonModel.NameValue = service.Name
			service.ResourceCommonModel.UpstreamIDValue = service.UpstreamID
			err = service.ResourceCommonModel.RestoreConfigForRead("service")
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(service.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal("test-id"))
			Expect(configMap["name"]).To(Equal("test-service"))
			Expect(configMap["upstream_id"]).To(Equal("test-upstream-id"))
		})

		It("should do nothing if Name is empty", func() {
			service.Name = ""
			service.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
				}`)),
			}
			err := service.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]any, 1)

			err = json.Unmarshal(service.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("name"))
		})

		It("should leave legacy upstream_id untouched when config already carries it", func() {
			service.UpstreamID = ""
			service.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"upstream_id": "test-upstream-id"
				}`)),
			}
			err := service.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]any, 1)

			err = json.Unmarshal(service.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["upstream_id"]).To(Equal("test-upstream-id"))
		})

		It("should preserve name and upstream from typed fields even when config omits them", func() {
			service.ResourceCommonModel.NameValue = "typed-service-name"
			service.ResourceCommonModel.UpstreamIDValue = "typed-upstream-id"

			typedService := service.ResourceCommonModel.ToResourceModel("service").(*model.Service)
			Expect(typedService.Name).To(Equal("typed-service-name"))
			Expect(typedService.UpstreamID).To(Equal("typed-upstream-id"))
		})
	})
})
