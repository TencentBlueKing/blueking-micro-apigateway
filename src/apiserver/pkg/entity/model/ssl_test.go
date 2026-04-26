package model_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

var _ = Describe("SSL", func() {
	var ssl model.SSL

	BeforeEach(func() {
		fixture := data.SSL1(data.Gateway1WithBkAPISIX(), "create_draft")
		ssl = *fixture
		ssl.Name = "ssl-a"
	})

	Describe("HandleConfig", func() {
		It("should strip echoed id and name from stored Config while keeping derived snis", func() {
			err := ssl.HandleConfig()
			Expect(err).NotTo(HaveOccurred())

			var configMap map[string]any
			err = json.Unmarshal(ssl.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap).NotTo(HaveKey("id"))
			Expect(configMap).NotTo(HaveKey("name"))
			Expect(configMap["snis"]).NotTo(BeNil())

			err = ssl.AfterFind(nil)
			Expect(err).NotTo(HaveOccurred())
			err = json.Unmarshal(ssl.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["id"]).To(Equal(ssl.ID))
			Expect(configMap["name"]).To(Equal("ssl-a"))
			Expect(configMap["snis"]).NotTo(BeNil())
		})

		It("should derive snis when missing", func() {
			ssl.ResourceCommonModel = model.ResourceCommonModel{
				ID: ssl.ID,
				Config: datatypes.JSON(
					[]byte(
						`{"cert":"-----BEGIN CERTIFICATE-----\nMIIDLDCCAhSgAwIBAgIRAKCvmzHJdxMK9dkm66wVTE8wDQYJKoZIhvcNAQELBQAw\ngYoxEjAQBgNVBAMMCWxkZGdvLm5ldDEMMAoGA1UECwwDZGV2MQ4wDAYDVQQKDAVs\nZGRnbzELMAkGA1UEBhMCQ04xIzAhBgkqhkiG9w0BCQEWFGxlY2hlbmdhZG1pbkAx\nMjYuY29tMREwDwYDVQQHDAhzaGFuZ2hhaTERMA8GA1UECAwIc2hhbmdoYWkwHhcN\nMjUwMzE4MDgwNzAyWhcNMjcwMzE4MDgwNzAyWjAYMRYwFAYDVQQDDA13d3cuYmFp\nZHUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAgrg/T1KtECD+\nitnOVeOFAx4W4MHLmyv0iaTcZe/HHMuaLbMrKGie5ENbJbXL83Zy7u69y9/YVm3X\n3pPy69cUYwftzSIw+uy+/lftFrGxHW58RytEtEqPAGg97A/L+94n9lQ4UWcOA18m\nQIRwtUjTiF43hnfdSAtfooDoo3BY5Kl3H3tgJGg5SsR1T225KCNNNax4mkdiIkYa\nUfTv3BBA+ifUa99048SL1cqSUeYSYPmjpBvdUZWAQVOkyTqBCq764fu0ruckfaO2\n8GOUN8yvD8HcOip9+vIEbdaDzSGYpsZe6467yQUlcJ7KnBTC5ST8yCqrt2mdIc3c\nHYfwvO+9AQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQBH1HWCfX5AECJuOhrEWj4B\nqNho7ZdU+vi2B+W7BB+79O2WqOVlN+PzPKstFOe0qEhGovZeHZZnB6qMKsLfzf07\ngPeRV1+aNMvIeVKNwmQF+TbVgqBf5sdg8KKoHmi7u/1ufucZJEFPNB4sYflexErt\nsRjQ9rCCxHPqkqpzcmFlq9Io5QuOY1WkssgJ+xWXp7F7R6gdED6eYoVvFKpfdKDA\nAbzS+2uhTgr+eDMCDWN8lUIStjE73Z0Do9EMZxwMLEwv6Hqj8piXbgvLNdAEuFMu\nttmao4QflHI4UJqrVcmuGQsxUP8D36QCIz4xVR/kWq64vUa4URQ7ttn2ghvaO96h\n-----END CERTIFICATE-----\n","key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCCuD9PUq0QIP6K\n2c5V44UDHhbgwcubK/SJpNxl78ccy5otsysoaJ7kQ1sltcvzdnLu7r3L39hWbdfe\nk/Lr1xRjB+3NIjD67L7+V+0WsbEdbnxHK0S0So8AaD3sD8v73if2VDhRZw4DXyZA\nhHC1SNOIXjeGd91IC1+igOijcFjkqXcfe2AkaDlKxHVPbbkoI001rHiaR2IiRhpR\n9O/cEED6J9Rr33TjxIvVypJR5hJg+aOkG91RlYBBU6TJOoEKrvrh+7Su5yR9o7bw\nY5Q3zK8Pwdw6Kn368gRt1oPNIZimxl7rjrvJBSVwnsqcFMLlJPzIKqu3aZ0hzdwd\nh/C8770BAgMBAAECggEANZCYaMHBJvXOOEmOEoXb0G45A7qF2z0ExI5oveCmX7dS\no11i1vkf+vta0zYOr+IesKfl4DAGr0vthEht55RHx1jNahyTo471qLWQ8pa3tA69\nIFCro5FVzd3pWd0TQk9DYt4aAclP5bPMse1TXgXMoHmzGQfvBgGbG7TlH2v/ERuG\nRY9W1/NgR98QftxmvmDDIBD1M4hnJy3pbUKVPp1+OXfLBy8QshqclWAf7R8m5+/j\nHuXhcDgpFhuE8otFH59a+SnPuhhF85F8qiolPWF9YQ7n7sAyllP3O3y2g9Eazwam\nkqkcoqKUqUZhAhqKlPBiHVA6aO89vzzcxf+/C5urzQKBgQC43oxqzugSMkceYHMl\nK3OZU+ui9OjUfjoPg0sKmJqfci89kYN4Z67EXdmFn0WBUXPY9XH6HjEz2l9MhyQ6\ngHQo/wbJZp8k/4doO/XJw2OiDu0oPKcGJlGuK5H9SfNuU0ntA00vqNxvKvYLvAJl\nvjBMbpb7J/axmbpXLhgrNsggywKBgQC1BAOJGOPGyOECfmiPcSu6PeUy/DmvE20b\nkGfZLwiRtQ949W5lH47Y3aOTHZyXXGL1cqo6O630WxcNfiZ7+LzNP6xvA+mzSjhU\nzxzV45Bdp2vTHiSE+JtddhQKz7ZX2lmbWjP/pR5lEvuq6zXPMk8/r6ToiN/rcPG0\n7wMgk6fb4wKBgFESJ3nfap4wNkf3/Abc20DuMHOx+zjUchnDdfEboxMxO85ANetj\nbJzomy+h/RUM50TJvkX1X5ZhuVESIq0VD9u6mvtPaZMMDBGF2e+1I8g5y37NumFU\nBJXgvZDaEUrcc5rgy8SOxLxrlqLmvBZqJTwfc06I5AJWbAU3TZoF2BWpAoGBAKFU\npHoKLug6nSCF3VcK/HgPNjnMxvSdEb9hYs0UuER05QdfZzbFe6EZWPKDj87vTluI\nCOPB0PZaQR+LcW1Ica1UtLB1AlMDMVWVChQvr7lowBb3ZIEGuiIAXTiNi+yc9QQa\nzwFn/sECvD7HR7wVEMCoIQgHBdtnXGVwKI9eSlsVAoGAWw8UWBWmufMKHQtKVZsV\nf0ruqlIwP3pHJs1A31TUREO0D/vHkZtcuwxps1xUdyicXbcTnNAsi8DR8mOsFozx\nquPN/IpsmFA/f41eFfRYtFh2+mx2+eGFi6Kx1P9eTB4IINofPT1trLsDWzUZcKWa\n++NvBBLYIHuQcxH9WVJ6h8w=\n-----END RSA PRIVATE KEY-----\n"}`,
					),
				),
			}
			err := ssl.HandleConfig()
			Expect(err).NotTo(HaveOccurred())
			var configMap map[string]any
			err = json.Unmarshal(ssl.Config, &configMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(configMap["snis"]).NotTo(BeNil())
		})
	})
})
