package model_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"strconv"
	"testing"
	"time"

	gomonkey "github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	protoutil "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/proto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/sslx"
)

const validProtoContent = `syntax = "proto3";
package test;

message TestMessage {
  string name = 1;
}`

type setBytesCall struct {
	path  string
	value any
}

func TestHandleConfigStoresCompleteConfig(t *testing.T) {
	t.Run("consumer injects identity and association fields", func(t *testing.T) {
		resource := model.Consumer{
			Username: "test-username",
			GroupID:  "test-group-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"group_id":"test-group-id",
			"username":"test-username"
		}`, string(resource.Config))
	})

	t.Run("consumer removes stale group_id when column is empty", func(t *testing.T) {
		resource := model.Consumer{
			Username: "test-username",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{"group_id":"stale-group-id"}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"username":"test-username"
		}`, string(resource.Config))
		assert.False(t, gjson.GetBytes(resource.Config, "group_id").Exists())
	})

	t.Run("consumer_group injects id and name", func(t *testing.T) {
		resource := model.ConsumerGroup{
			Name: "test-consumer-group",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-consumer-group"
		}`, string(resource.Config))
	})

	t.Run("global_rule injects id and name", func(t *testing.T) {
		resource := model.GlobalRule{
			Name: "test-global-rule",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-global-rule"
		}`, string(resource.Config))
	})

	t.Run("plugin_config injects id and name", func(t *testing.T) {
		resource := model.PluginConfig{
			Name: "test-plugin-config",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-plugin-config"
		}`, string(resource.Config))
	})

	t.Run("plugin_metadata uses name as persisted config id", func(t *testing.T) {
		resource := model.PluginMetadata{
			Name: "test-plugin-metadata",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "ignored-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-plugin-metadata",
			"name":"test-plugin-metadata"
		}`, string(resource.Config))
	})

	t.Run("proto stores full config and keeps parseable content", func(t *testing.T) {
		resource := model.Proto{
			Name: "test.proto",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{"content":` + quoteJSON(t, validProtoContent) + `}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.Equal(t, "test-id", gjson.GetBytes(resource.Config, "id").String())
		assert.Equal(t, "test.proto", gjson.GetBytes(resource.Config, "name").String())
		assert.Equal(t, validProtoContent, gjson.GetBytes(resource.Config, "content").String())
	})

	t.Run("route injects all identity and association fields", func(t *testing.T) {
		resource := model.Route{
			Name:           "test-route",
			ServiceID:      "test-service-id",
			PluginConfigID: "test-plugin-config-id",
			UpstreamID:     "test-upstream-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-route",
			"service_id":"test-service-id",
			"plugin_config_id":"test-plugin-config-id",
			"upstream_id":"test-upstream-id"
		}`, string(resource.Config))
	})

	t.Run("route removes stale empty associations", func(t *testing.T) {
		resource := model.Route{
			Name: "test-route",
			ResourceCommonModel: model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"service_id":"stale-service-id",
					"plugin_config_id":"stale-plugin-config-id",
					"upstream_id":"stale-upstream-id"
				}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-route"
		}`, string(resource.Config))
		assert.False(t, gjson.GetBytes(resource.Config, "service_id").Exists())
		assert.False(t, gjson.GetBytes(resource.Config, "plugin_config_id").Exists())
		assert.False(t, gjson.GetBytes(resource.Config, "upstream_id").Exists())
	})

	t.Run("service injects id name and upstream association", func(t *testing.T) {
		resource := model.Service{
			Name:       "test-service",
			UpstreamID: "test-upstream-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-service",
			"upstream_id":"test-upstream-id"
		}`, string(resource.Config))
	})

	t.Run("service removes stale upstream_id when column is empty", func(t *testing.T) {
		resource := model.Service{
			Name: "test-service",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{"upstream_id":"stale-upstream-id"}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-service"
		}`, string(resource.Config))
		assert.False(t, gjson.GetBytes(resource.Config, "upstream_id").Exists())
	})

	t.Run("ssl preserves explicit snis in stored config", func(t *testing.T) {
		crt, key, _ := mustGenerateCertificate(t)
		resource := model.SSL{
			Name: "test-ssl",
			ResourceCommonModel: model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"cert":` + quoteJSON(t, crt) + `,
					"key":` + quoteJSON(t, key) + `,
					"snis":["preset.example.com"]
				}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.Equal(t, "test-id", gjson.GetBytes(resource.Config, "id").String())
		assert.Equal(t, "test-ssl", gjson.GetBytes(resource.Config, "name").String())
		assert.Equal(t, "preset.example.com", gjson.GetBytes(resource.Config, "snis.0").String())
	})

	t.Run("ssl derives snis from cert when omitted", func(t *testing.T) {
		crt, key, dnsNames := mustGenerateCertificate(t)
		resource := model.SSL{
			Name: "test-ssl",
			ResourceCommonModel: model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"cert":` + quoteJSON(t, crt) + `,
					"key":` + quoteJSON(t, key) + `
				}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.Equal(t, dnsNames[0], gjson.GetBytes(resource.Config, "snis.0").String())
	})

	t.Run("stream_route injects id name and associations", func(t *testing.T) {
		resource := model.StreamRoute{
			Name:       "test-stream-route",
			ServiceID:  "test-service-id",
			UpstreamID: "test-upstream-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-stream-route",
			"service_id":"test-service-id",
			"upstream_id":"test-upstream-id"
		}`, string(resource.Config))
	})

	t.Run("stream_route removes stale associations when columns are empty", func(t *testing.T) {
		resource := model.StreamRoute{
			Name: "test-stream-route",
			ResourceCommonModel: model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"service_id":"stale-service-id",
					"upstream_id":"stale-upstream-id"
				}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-stream-route"
		}`, string(resource.Config))
		assert.False(t, gjson.GetBytes(resource.Config, "service_id").Exists())
		assert.False(t, gjson.GetBytes(resource.Config, "upstream_id").Exists())
	})

	t.Run("upstream injects tls client cert association", func(t *testing.T) {
		resource := model.Upstream{
			Name:  "test-upstream",
			SSLID: "test-ssl-id",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-upstream",
			"tls":{"client_cert_id":"test-ssl-id"}
		}`, string(resource.Config))
	})

	t.Run("upstream removes stale tls client cert association", func(t *testing.T) {
		resource := model.Upstream{
			Name: "test-upstream",
			ResourceCommonModel: model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"tls":{"client_cert_id":"stale-ssl-id"}
				}`)),
			},
		}

		mustNoError(t, resource.HandleConfig())
		assert.JSONEq(t, `{
			"id":"test-id",
			"name":"test-upstream"
		}`, string(resource.Config))
		assert.False(t, gjson.GetBytes(resource.Config, "tls.client_cert_id").Exists())
	})
}

func TestHandleConfigRejectsInvalidStoredJSON(t *testing.T) {
	tests := []struct {
		name   string
		invoke func() error
	}{
		{
			name: "consumer",
			invoke: func() error {
				resource := model.Consumer{
					Username: "test-username",
					GroupID:  "test-group-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "consumer_group",
			invoke: func() error {
				resource := model.ConsumerGroup{
					Name: "test-consumer-group",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "global_rule",
			invoke: func() error {
				resource := model.GlobalRule{
					Name: "test-global-rule",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "plugin_config",
			invoke: func() error {
				resource := model.PluginConfig{
					Name: "test-plugin-config",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "plugin_metadata",
			invoke: func() error {
				resource := model.PluginMetadata{
					Name: "test-plugin-metadata",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "ignored-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "proto",
			invoke: func() error {
				resource := model.Proto{
					Name: "test.proto",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "route",
			invoke: func() error {
				resource := model.Route{
					Name:           "test-route",
					ServiceID:      "test-service-id",
					PluginConfigID: "test-plugin-config-id",
					UpstreamID:     "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "service",
			invoke: func() error {
				resource := model.Service{
					Name:       "test-service",
					UpstreamID: "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "ssl",
			invoke: func() error {
				resource := model.SSL{
					Name: "test-ssl",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "stream_route",
			invoke: func() error {
				resource := model.StreamRoute{
					Name:       "test-stream-route",
					ServiceID:  "test-service-id",
					UpstreamID: "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "upstream",
			invoke: func() error {
				resource := model.Upstream{
					Name:  "test-upstream",
					SSLID: "test-ssl-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`[]`)),
					},
				}
				return resource.HandleConfig()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mustError(t, tc.invoke())
		})
	}
}

func TestHandleConfigPropagatesInjectedFieldErrors(t *testing.T) {
	expectedErr := errors.New("set bytes failed")

	tests := []struct {
		name         string
		successCalls []setBytesCall
		invoke       func() error
	}{
		{
			name: "consumer group_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.Consumer{
					Username: "test-username",
					GroupID:  "test-group-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "consumer username",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "group_id", value: "test-group-id"},
			},
			invoke: func() error {
				resource := model.Consumer{
					Username: "test-username",
					GroupID:  "test-group-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "consumer_group name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.ConsumerGroup{
					Name: "test-consumer-group",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "global_rule name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.GlobalRule{
					Name: "test-global-rule",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "plugin_config name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.PluginConfig{
					Name: "test-plugin-config",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "plugin_metadata name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-plugin-metadata"},
			},
			invoke: func() error {
				resource := model.PluginMetadata{
					Name: "test-plugin-metadata",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "ignored-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "proto name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.Proto{
					Name: "test.proto",
					ResourceCommonModel: model.ResourceCommonModel{
						ID: "test-id",
						Config: datatypes.JSON(
							[]byte(`{"content":` + quoteJSON(t, validProtoContent) + `}`),
						),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "route name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.Route{
					Name:           "test-route",
					ServiceID:      "test-service-id",
					PluginConfigID: "test-plugin-config-id",
					UpstreamID:     "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "route service_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "name", value: "test-route"},
			},
			invoke: func() error {
				resource := model.Route{
					Name:           "test-route",
					ServiceID:      "test-service-id",
					PluginConfigID: "test-plugin-config-id",
					UpstreamID:     "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "route plugin_config_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "name", value: "test-route"},
				{path: "service_id", value: "test-service-id"},
			},
			invoke: func() error {
				resource := model.Route{
					Name:           "test-route",
					ServiceID:      "test-service-id",
					PluginConfigID: "test-plugin-config-id",
					UpstreamID:     "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "route upstream_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "name", value: "test-route"},
				{path: "service_id", value: "test-service-id"},
				{path: "plugin_config_id", value: "test-plugin-config-id"},
			},
			invoke: func() error {
				resource := model.Route{
					Name:           "test-route",
					ServiceID:      "test-service-id",
					PluginConfigID: "test-plugin-config-id",
					UpstreamID:     "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "service name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.Service{
					Name:       "test-service",
					UpstreamID: "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "service upstream_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "name", value: "test-service"},
			},
			invoke: func() error {
				resource := model.Service{
					Name:       "test-service",
					UpstreamID: "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "ssl name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				crt, key, _ := mustGenerateCertificate(t)
				resource := model.SSL{
					Name: "test-ssl",
					ResourceCommonModel: model.ResourceCommonModel{
						ID: "test-id",
						Config: datatypes.JSON([]byte(`{
							"cert":` + quoteJSON(t, crt) + `,
							"key":` + quoteJSON(t, key) + `
						}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "stream_route name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.StreamRoute{
					Name:       "test-stream-route",
					ServiceID:  "test-service-id",
					UpstreamID: "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "stream_route service_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "name", value: "test-stream-route"},
			},
			invoke: func() error {
				resource := model.StreamRoute{
					Name:       "test-stream-route",
					ServiceID:  "test-service-id",
					UpstreamID: "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "stream_route upstream_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "name", value: "test-stream-route"},
				{path: "service_id", value: "test-service-id"},
			},
			invoke: func() error {
				resource := model.StreamRoute{
					Name:       "test-stream-route",
					ServiceID:  "test-service-id",
					UpstreamID: "test-upstream-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "upstream name",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
			},
			invoke: func() error {
				resource := model.Upstream{
					Name:  "test-upstream",
					SSLID: "test-ssl-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
		{
			name: "upstream tls.client_cert_id",
			successCalls: []setBytesCall{
				{path: "id", value: "test-id"},
				{path: "name", value: "test-upstream"},
			},
			invoke: func() error {
				resource := model.Upstream{
					Name:  "test-upstream",
					SSLID: "test-ssl-id",
					ResourceCommonModel: model.ResourceCommonModel{
						ID:     "test-id",
						Config: datatypes.JSON([]byte(`{}`)),
					},
				}
				return resource.HandleConfig()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			patches := patchSetBytesErrorAfterSuccesses(t, []byte(`{}`), tc.successCalls, expectedErr)
			defer patches.Reset()

			err := tc.invoke()
			mustErrorIs(t, err, expectedErr)
		})
	}
}

func TestHandleConfigPropagatesSpecializedValidationErrors(t *testing.T) {
	t.Run("proto parse error", func(t *testing.T) {
		expectedErr := errors.New("parse proto failed")
		patches := gomonkey.ApplyFuncReturn(protoutil.ParseContent, expectedErr)
		defer patches.Reset()

		resource := model.Proto{
			Name: "test.proto",
			ResourceCommonModel: model.ResourceCommonModel{
				ID:     "test-id",
				Config: datatypes.JSON([]byte(`{"content":` + quoteJSON(t, validProtoContent) + `}`)),
			},
		}

		err := resource.HandleConfig()
		mustErrorIs(t, err, expectedErr)
	})

	t.Run("ssl parse cert error", func(t *testing.T) {
		expectedErr := errors.New("parse cert failed")
		patches := gomonkey.ApplyFuncReturn(sslx.ParseCert, []string(nil), expectedErr)
		defer patches.Reset()

		crt, key, _ := mustGenerateCertificate(t)
		resource := model.SSL{
			Name: "test-ssl",
			ResourceCommonModel: model.ResourceCommonModel{
				ID: "test-id",
				Config: datatypes.JSON([]byte(`{
					"cert":` + quoteJSON(t, crt) + `,
					"key":` + quoteJSON(t, key) + `
				}`)),
			},
		}

		err := resource.HandleConfig()
		mustErrorIs(t, err, expectedErr)
	})
}

func patchSetBytesErrorAfterSuccesses(
	t *testing.T,
	initial []byte,
	successCalls []setBytesCall,
	expectedErr error,
) *gomonkey.Patches {
	t.Helper()

	current := initial
	outputs := make([]gomonkey.OutputCell, 0, len(successCalls)+1)
	for _, call := range successCalls {
		current = mustSetBytes(t, current, call.path, call.value)
		outputs = append(outputs, gomonkey.OutputCell{
			Values: gomonkey.Params{current, nil},
			Times:  1,
		})
	}

	outputs = append(outputs, gomonkey.OutputCell{
		Values: gomonkey.Params{[]byte(nil), expectedErr},
		Times:  1,
	})

	return gomonkey.ApplyFuncSeq(sjson.SetBytesOptions, outputs)
}

func mustSetBytes(t *testing.T, raw []byte, path string, value any) []byte {
	t.Helper()

	out, err := sjson.SetBytes(raw, path, value)
	mustNoError(t, err)
	return out
}

func mustGenerateCertificate(t *testing.T) (string, string, []string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	mustNoError(t, err)

	dnsNames := []string{"test.example.com"}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   dnsNames[0],
			Organization: []string{"Test Org"},
		},
		DNSNames:              dnsNames,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	mustNoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	return string(certPEM), string(keyPEM), dnsNames
}

func quoteJSON(t *testing.T, value string) string {
	t.Helper()

	return strconv.Quote(value)
}

func mustNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func mustErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("expected error %v, got %v", target, err)
	}
}
