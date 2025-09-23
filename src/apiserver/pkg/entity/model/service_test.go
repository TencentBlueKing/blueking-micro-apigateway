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
		It("should set id, name, and upstream_id into the Config", func() {
			err := service.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]interface{}
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

			configMap := make(map[string]interface{}, 1)

			err = json.Unmarshal(service.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("name"))
		})

		It("should delete upstream_id from the Config if UpstreamID is empty", func() {
			service.UpstreamID = ""
			service.ResourceCommonModel = model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"upstream_id": "test-upstream-id"
				}`)),
			}
			err := service.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			configMap := make(map[string]interface{}, 1)

			err = json.Unmarshal(service.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("upstream_id"))
		})
	})
})
