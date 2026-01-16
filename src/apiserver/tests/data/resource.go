/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
 * Copyright (C) 2025 Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *     http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

// Package data ...
package data

import (
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// Gateway1WithBkAPISIX ...
func Gateway1WithBkAPISIX() *model.Gateway {
	gateway := &model.Gateway{
		Name:          "gateway1",
		Mode:          1,
		Maintainers:   []string{"user1", "user2"},
		Desc:          "gateway1",
		APISIXType:    constant.APISIXTypeBKAPISIX,
		APISIXVersion: "3.11.0",
		EtcdConfig: model.EtcdConfig{
			InstanceID: "123456789",
			EtcdConfig: base.EtcdConfig{
				Endpoint: "localhost:4379",
				Username: "test",
				Password: "test",
				Prefix:   "/apisix",
			},
		},
	}
	return gateway
}

// Route1WithNoRelationResource ...
func Route1WithNoRelationResource(gateway *model.Gateway, status constant.ResourceStatus) *model.Route {
	route := &model.Route{
		Name:           "route1",
		ServiceID:      "",
		UpstreamID:     "",
		PluginConfigID: "",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.Route),
			Config: datatypes.JSON(`{
				  "uris": [
				    "/get"
				  ],
				  "methods": [
				    "GET"
				  ],
                  "labels": {
					"build": "16",
					"env": "4"
				  },
				  "upstream": {
				    "type": "roundrobin",
				    "nodes": [
				      {
				        "host": "httpbin.org",
				        "port": 80,
				        "weight": 1
				      }
				    ],
				    "scheme": "http"
				  }
				}`),
			Status: status,
		},
	}
	return route
}

// Route2WithNoRelationResource ...
func Route2WithNoRelationResource(gateway *model.Gateway, status constant.ResourceStatus) *model.Route {
	route := &model.Route{
		Name:           "route2",
		ServiceID:      "",
		UpstreamID:     "",
		PluginConfigID: "",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.Route),
			Config: datatypes.JSON(`{
				  "uris": [
				    "/get"
				  ],
				  "methods": [
				    "GET"
				  ],
                  "labels": {
					"build": "16",
					"env": "4"
				  },
				  "upstream": {
				    "type": "roundrobin",
				    "nodes": [
				      {
				        "host": "httpbin.org",
				        "port": 80,
				        "weight": 1
				      }
				    ],
				    "scheme": "http"
				  }
				}`),
			Status: status,
		},
	}
	return route
}

// Service1WithNoRelation ...
func Service1WithNoRelation(gateway *model.Gateway, status constant.ResourceStatus) *model.Service {
	service := &model.Service{
		Name:       "service1",
		UpstreamID: "",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.Service),
			Config: datatypes.JSON(`{
				  "upstream": {
				    "type": "roundrobin",
				    "nodes": [
				      {
				        "host": "httpbin.org",
				        "port": 80,
				        "weight": 1
				      }
				    ],
				    "scheme": "http"
				  }
				}`),
			Status: status,
		},
	}
	return service
}

// Upstream1WithNoRelation ...
func Upstream1WithNoRelation(gateway *model.Gateway, status constant.ResourceStatus) *model.Upstream {
	upstream := &model.Upstream{
		Name: "upstream1",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.Upstream),
			Config: datatypes.JSON(`{
				  "type": "roundrobin",
				  "nodes": [
				    {
				      "host": "httpbin.org",
				      "port": 80,
				      "weight": 1
				    }
				  ],
				  "scheme": "http"
				}`),
			Status: status,
		},
	}
	return upstream
}

// Consumer1WithNoRelation ...
func Consumer1WithNoRelation(gateway *model.Gateway, status constant.ResourceStatus) *model.Consumer {
	route := &model.Consumer{
		Username: "consumer1",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.Consumer),
			Config: datatypes.JSON(`{
			    "plugins": {
			        "key-auth": {
			            "key": "auth-one"
			        },
			        "limit-count": {
			            "count": 2,
			            "time_window": 60,
			            "rejected_code": 503,
			            "key": "remote_addr",
			            "policy": "local"
			        }
			    }
			}`),
			Status: status,
		},
	}
	return route
}

// PluginConfig1WithNoRelation ...
func PluginConfig1WithNoRelation(gateway *model.Gateway, status constant.ResourceStatus) *model.PluginConfig {
	config := &model.PluginConfig{
		Name: "plugin-config-1",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.PluginConfig),
			Config: datatypes.JSON(`{
				"plugins": {
					"limit-count": {
						"count": 100,
						"time_window": 60,
						"key": "remote_addr",
						"rejected_code": 503,
					"policy": "local"
					}
				}
			}`),
			Status: status,
		},
	}
	return config
}

// GlobalRule1 ...
func GlobalRule1(gateway *model.Gateway, status constant.ResourceStatus) *model.GlobalRule {
	rule := &model.GlobalRule{
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.GlobalRule),
			Config: datatypes.JSON(`{
				"plugins": {
					"prometheus": {
						"prefer_name": true
					}
				}
			}`),
			Status: status,
		},
	}
	return rule
}

// Proto1 ...
func Proto1(gateway *model.Gateway, status constant.ResourceStatus) *model.Proto {
	return &model.Proto{
		Name: "test.proto",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.Proto),
			// nolint:lll
			Config: datatypes.JSON(`{
				"content": "syntax = \"proto3\";\npackage helloworld;\nservice Greeter {\n  rpc SayHello (HelloRequest) returns (HelloReply) {}\n}\nmessage HelloRequest {\n  string name = 1;\n}\nmessage HelloReply {\n  string message = 1;\n}"
			}`),
			Status: status,
		},
	}
}

// PluginMetadata1 ...
func PluginMetadata1(gateway *model.Gateway, status constant.ResourceStatus) *model.PluginMetadata {
	return &model.PluginMetadata{
		Name: "clickhouse-logger",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.PluginMetadata),
			Config: datatypes.JSON(`{
			    "config": {
			        "log_format": {
			            "@timestamp": "$time_iso86011",
			            "client_ip": "$remote_addr1",
			            "host": "$host"
			        },
			        "name": "clickhouse-logger"
			    }
			}`),
			Status: status,
		},
	}
}

// ConsumerGroup1WithNoRelation ...
func ConsumerGroup1WithNoRelation(gateway *model.Gateway, status constant.ResourceStatus) *model.ConsumerGroup {
	return &model.ConsumerGroup{
		Name: "group1",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.ConsumerGroup),
			Config: datatypes.JSON(`{
				"plugins": {
					"limit-count": {
						"count": 100,
						"time_window": 60,
						"key": "remote_addr",
					"policy": "local"
					}
				}
			}`),
			Status: status,
		},
	}
}

// SSL1 ...
func SSL1(gateway *model.Gateway, status constant.ResourceStatus) *model.SSL {
	return &model.SSL{
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.SSL),
			// nolint:lll
			Config: datatypes.JSON(
				`{"cert":"-----BEGIN CERTIFICATE-----\nMIIDLDCCAhSgAwIBAgIRAKCvmzHJdxMK9dkm66wVTE8wDQYJKoZIhvcNAQELBQAw\ngYoxEjAQBgNVBAMMCWxkZGdvLm5ldDEMMAoGA1UECwwDZGV2MQ4wDAYDVQQKDAVs\nZGRnbzELMAkGA1UEBhMCQ04xIzAhBgkqhkiG9w0BCQEWFGxlY2hlbmdhZG1pbkAx\nMjYuY29tMREwDwYDVQQHDAhzaGFuZ2hhaTERMA8GA1UECAwIc2hhbmdoYWkwHhcN\nMjUwMzE4MDgwNzAyWhcNMjcwMzE4MDgwNzAyWjAYMRYwFAYDVQQDDA13d3cuYmFp\nZHUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAgrg/T1KtECD+\nitnOVeOFAx4W4MHLmyv0iaTcZe/HHMuaLbMrKGie5ENbJbXL83Zy7u69y9/YVm3X\n3pPy69cUYwftzSIw+uy+/lftFrGxHW58RytEtEqPAGg97A/L+94n9lQ4UWcOA18m\nQIRwtUjTiF43hnfdSAtfooDoo3BY5Kl3H3tgJGg5SsR1T225KCNNNax4mkdiIkYa\nUfTv3BBA+ifUa99048SL1cqSUeYSYPmjpBvdUZWAQVOkyTqBCq764fu0ruckfaO2\n8GOUN8yvD8HcOip9+vIEbdaDzSGYpsZe6467yQUlcJ7KnBTC5ST8yCqrt2mdIc3c\nHYfwvO+9AQIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQBH1HWCfX5AECJuOhrEWj4B\nqNho7ZdU+vi2B+W7BB+79O2WqOVlN+PzPKstFOe0qEhGovZeHZZnB6qMKsLfzf07\ngPeRV1+aNMvIeVKNwmQF+TbVgqBf5sdg8KKoHmi7u/1ufucZJEFPNB4sYflexErt\nsRjQ9rCCxHPqkqpzcmFlq9Io5QuOY1WkssgJ+xWXp7F7R6gdED6eYoVvFKpfdKDA\nAbzS+2uhTgr+eDMCDWN8lUIStjE73Z0Do9EMZxwMLEwv6Hqj8piXbgvLNdAEuFMu\nttmao4QflHI4UJqrVcmuGQsxUP8D36QCIz4xVR/kWq64vUa4URQ7ttn2ghvaO96h\n-----END CERTIFICATE-----\n","key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCCuD9PUq0QIP6K\n2c5V44UDHhbgwcubK/SJpNxl78ccy5otsysoaJ7kQ1sltcvzdnLu7r3L39hWbdfe\nk/Lr1xRjB+3NIjD67L7+V+0WsbEdbnxHK0S0So8AaD3sD8v73if2VDhRZw4DXyZA\nhHC1SNOIXjeGd91IC1+igOijcFjkqXcfe2AkaDlKxHVPbbkoI001rHiaR2IiRhpR\n9O/cEED6J9Rr33TjxIvVypJR5hJg+aOkG91RlYBBU6TJOoEKrvrh+7Su5yR9o7bw\nY5Q3zK8Pwdw6Kn368gRt1oPNIZimxl7rjrvJBSVwnsqcFMLlJPzIKqu3aZ0hzdwd\nh/C8770BAgMBAAECggEANZCYaMHBJvXOOEmOEoXb0G45A7qF2z0ExI5oveCmX7dS\no11i1vkf+vta0zYOr+IesKfl4DAGr0vthEht55RHx1jNahyTo471qLWQ8pa3tA69\nIFCro5FVzd3pWd0TQk9DYt4aAclP5bPMse1TXgXMoHmzGQfvBgGbG7TlH2v/ERuG\nRY9W1/NgR98QftxmvmDDIBD1M4hnJy3pbUKVPp1+OXfLBy8QshqclWAf7R8m5+/j\nHuXhcDgpFhuE8otFH59a+SnPuhhF85F8qiolPWF9YQ7n7sAyllP3O3y2g9Eazwam\nkqkcoqKUqUZhAhqKlPBiHVA6aO89vzzcxf+/C5urzQKBgQC43oxqzugSMkceYHMl\nK3OZU+ui9OjUfjoPg0sKmJqfci89kYN4Z67EXdmFn0WBUXPY9XH6HjEz2l9MhyQ6\ngHQo/wbJZp8k/4doO/XJw2OiDu0oPKcGJlGuK5H9SfNuU0ntA00vqNxvKvYLvAJl\nvjBMbpb7J/axmbpXLhgrNsggywKBgQC1BAOJGOPGyOECfmiPcSu6PeUy/DmvE20b\nkGfZLwiRtQ949W5lH47Y3aOTHZyXXGL1cqo6O630WxcNfiZ7+LzNP6xvA+mzSjhU\nzxzV45Bdp2vTHiSE+JtddhQKz7ZX2lmbWjP/pR5lEvuq6zXPMk8/r6ToiN/rcPG0\n7wMgk6fb4wKBgFESJ3nfap4wNkf3/Abc20DuMHOx+zjUchnDdfEboxMxO85ANetj\nbJzomy+h/RUM50TJvkX1X5ZhuVESIq0VD9u6mvtPaZMMDBGF2e+1I8g5y37NumFU\nBJXgvZDaEUrcc5rgy8SOxLxrlqLmvBZqJTwfc06I5AJWbAU3TZoF2BWpAoGBAKFU\npHoKLug6nSCF3VcK/HgPNjnMxvSdEb9hYs0UuER05QdfZzbFe6EZWPKDj87vTluI\nCOPB0PZaQR+LcW1Ica1UtLB1AlMDMVWVChQvr7lowBb3ZIEGuiIAXTiNi+yc9QQa\nzwFn/sECvD7HR7wVEMCoIQgHBdtnXGVwKI9eSlsVAoGAWw8UWBWmufMKHQtKVZsV\nf0ruqlIwP3pHJs1A31TUREO0D/vHkZtcuwxps1xUdyicXbcTnNAsi8DR8mOsFozx\nquPN/IpsmFA/f41eFfRYtFh2+mx2+eGFi6Kx1P9eTB4IINofPT1trLsDWzUZcKWa\n++NvBBLYIHuQcxH9WVJ6h8w=\n-----END RSA PRIVATE KEY-----\n","snis":["www.baidu.com"]}`,
			),
			Status: status,
		},
	}
}

// StreamRoute1WithNoRelationResource 测试数据
func StreamRoute1WithNoRelationResource(gateway *model.Gateway, status constant.ResourceStatus) *model.StreamRoute {
	streamRoute := &model.StreamRoute{
		Name:       "streamRoute1",
		ServiceID:  "",
		UpstreamID: "",
		ResourceCommonModel: model.ResourceCommonModel{
			GatewayID: gateway.ID,
			ID:        idx.GenResourceID(constant.StreamRoute),
			Config: datatypes.JSON(`{
				  "remote_addr": "127.0.0.1",
				  "server_addr": "127.0.0.1",
				  "server_port": 8080,
				  "sni": "test.com",
				  "protocol": {
					  "name": "redis",
					  "conf": {}
				  },
				  "upstream": {
				    "type": "roundrobin",
				    "nodes": [
				      {
				        "host": "httpbin.org",
				        "port": 80,
				        "weight": 1
				      }
				    ],
				    "scheme": "http"
				  }
				}`),
			Status: status,
		},
	}
	return streamRoute
}
