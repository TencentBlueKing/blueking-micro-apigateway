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

package entity

import "encoding/json"

// BaseInfo ...
type BaseInfo struct {
	ID         any               `json:"id"`
	CreateTime int64             `json:"create_time,omitempty"`
	UpdateTime int64             `json:"update_time,omitempty"`
	Name       string            `json:"name,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// Status ...
type Status uint8

// Route ...
type Route struct {
	BaseInfo
	URI             string         `json:"uri,omitempty"`
	Uris            []string       `json:"uris,omitempty"`
	Desc            string         `json:"desc,omitempty"`
	Priority        int            `json:"priority,omitempty"`
	Methods         []string       `json:"methods,omitempty"`
	Host            string         `json:"host,omitempty"`
	Hosts           []string       `json:"hosts,omitempty"`
	RemoteAddr      string         `json:"remote_addr,omitempty"`
	RemoteAddrs     []string       `json:"remote_addrs,omitempty"`
	Vars            []any          `json:"vars,omitempty"`
	FilterFunc      string         `json:"filter_func,omitempty"`
	Script          any            `json:"script,omitempty"`
	ScriptID        any            `json:"script_id,omitempty"`
	Plugins         map[string]any `json:"plugins,omitempty"`
	PluginConfigID  any            `json:"plugin_config_id,omitempty"`
	Upstream        *UpstreamDef   `json:"upstream,omitempty"`
	ServiceID       any            `json:"service_id,omitempty"`
	UpstreamID      any            `json:"upstream_id,omitempty"`
	ServiceProtocol string         `json:"service_protocol,omitempty"`
	EnableWebsocket bool           `json:"enable_websocket,omitempty"`
	Status          Status         `json:"status"`
}

// TimeoutValue ...
type (
	TimeoutValue float32
	Timeout      struct {
		Connect TimeoutValue `json:"connect,omitempty"`
		Send    TimeoutValue `json:"send,omitempty"`
		Read    TimeoutValue `json:"read,omitempty"`
	}
)

// Node ...
type Node struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Weight   int    `json:"weight"`
	Metadata any    `json:"metadata,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// Healthy ...
type Healthy struct {
	Interval     int   `json:"interval,omitempty"`
	HttpStatuses []int `json:"http_statuses,omitempty"`
	Successes    int   `json:"successes,omitempty"`
}

// UnHealthy ...
type UnHealthy struct {
	Interval     int   `json:"interval,omitempty"`
	HTTPStatuses []int `json:"http_statuses,omitempty"`
	TCPFailures  int   `json:"tcp_failures,omitempty"`
	Timeouts     int   `json:"timeouts,omitempty"`
	HTTPFailures int   `json:"http_failures,omitempty"`
}

// Active ...
type Active struct {
	Type                   string       `json:"type,omitempty"`
	Timeout                TimeoutValue `json:"timeout,omitempty"`
	Concurrency            int          `json:"concurrency,omitempty"`
	Host                   string       `json:"host,omitempty"`
	Port                   int          `json:"port,omitempty"`
	HTTPPath               string       `json:"http_path,omitempty"`
	HTTPSVerifyCertificate bool         `json:"https_verify_certificate,omitempty"`
	Healthy                Healthy      `json:"healthy,omitempty"`
	UnHealthy              UnHealthy    `json:"unhealthy,omitempty"`
	ReqHeaders             []string     `json:"req_headers,omitempty"`
}

// Passive ...
type Passive struct {
	Type      string    `json:"type,omitempty"`
	Healthy   Healthy   `json:"healthy,omitempty"`
	UnHealthy UnHealthy `json:"unhealthy,omitempty"`
}

// HealthChecker ...
type HealthChecker struct {
	Active  Active  `json:"active,omitempty"`
	Passive Passive `json:"passive,omitempty"`
}

// UpstreamTLS ...
type UpstreamTLS struct {
	ClientCert   string `json:"client_cert,omitempty"`
	ClientKey    string `json:"client_key,omitempty"`
	ClientCertId string `json:"client_cert_id,omitempty"`
}

// UpstreamKeepalivePool ...
type UpstreamKeepalivePool struct {
	IdleTimeout *TimeoutValue `json:"idle_timeout,omitempty"`
	Requests    int           `json:"requests,omitempty"`
	Size        int           `json:"size"`
}

// UpstreamDef ...
type UpstreamDef struct {
	BaseInfo
	Nodes         any                    `json:"nodes,omitempty"`
	Retries       *int                   `json:"retries,omitempty"`
	Timeout       *Timeout               `json:"timeout,omitempty"`
	Type          string                 `json:"type,omitempty"`
	Checks        any                    `json:"checks,omitempty"`
	HashOn        string                 `json:"hash_on,omitempty"`
	Key           string                 `json:"key,omitempty"`
	Scheme        string                 `json:"scheme,omitempty"`
	DiscoveryType string                 `json:"discovery_type,omitempty"`
	DiscoveryArgs map[string]any         `json:"discovery_args,omitempty"`
	PassHost      string                 `json:"pass_host,omitempty"`
	UpstreamHost  string                 `json:"upstream_host,omitempty"`
	Desc          string                 `json:"desc,omitempty"`
	ServiceName   string                 `json:"service_name,omitempty"`
	TLS           *UpstreamTLS           `json:"tls,omitempty"`
	KeepalivePool *UpstreamKeepalivePool `json:"keepalive_pool,omitempty"`
	RetryTimeout  TimeoutValue           `json:"retry_timeout,omitempty"`
}

// Upstream ...
type Upstream struct {
	UpstreamDef
}

// Consumer ...
type Consumer struct {
	Username   string            `json:"username"`
	Desc       string            `json:"desc,omitempty"`
	Plugins    map[string]any    `json:"plugins,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	CreateTime int64             `json:"create_time,omitempty"`
	UpdateTime int64             `json:"update_time,omitempty"`
	GroupID    string            `json:"group_id,omitempty"`
}

// ConsumerGroup ...
type ConsumerGroup struct {
	Desc       string            `json:"desc,omitempty"`
	Plugins    map[string]any    `json:"plugins,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	CreateTime int64             `json:"create_time,omitempty"`
	UpdateTime int64             `json:"update_time,omitempty"`
}

// Service ...
type Service struct {
	BaseInfo
	Desc            string         `json:"desc,omitempty"`
	Upstream        *UpstreamDef   `json:"upstream,omitempty"`
	UpstreamID      any            `json:"upstream_id,omitempty"`
	Plugins         map[string]any `json:"plugins,omitempty"`
	Script          string         `json:"script,omitempty"`
	EnableWebsocket bool           `json:"enable_websocket,omitempty"`
	Hosts           []string       `json:"hosts,omitempty"`
}

// GlobalRule ...
type GlobalRule struct {
	BaseInfo
	Plugins map[string]any `json:"plugins"`
}

// PluginMetadataConf ...
type PluginMetadataConf map[string]any

// PluginMetaData ...
type PluginMetaData struct {
	PluginMetadataConf
}

// UnmarshalJSON 解析PluginMetadataConf
func (c *PluginMetadataConf) UnmarshalJSON(dAtA []byte) error {
	temp := make(map[string]any)
	if err := json.Unmarshal(dAtA, &temp); err != nil {
		return err
	}
	*c = temp
	return nil
}

// MarshalJSON 将PluginMetadataConf转换为json
func (c *PluginMetadataConf) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any(*c))
}

// ServerInfo ...
type ServerInfo struct {
	BaseInfo
	LastReportTime int64  `json:"last_report_time,omitempty"`
	UpTime         int64  `json:"up_time,omitempty"`
	BootTime       int64  `json:"boot_time,omitempty"`
	EtcdVersion    string `json:"etcd_version,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	Version        string `json:"version,omitempty"`
}

// PluginConfig ...
type PluginConfig struct {
	BaseInfo
	Desc    string         `json:"desc,omitempty"`
	Plugins map[string]any `json:"plugins"`
}

// SSLClient ...
type SSLClient struct {
	CA               string   `json:"ca,omitempty"`
	Depth            int      `json:"depth,omitempty"`
	SkipMtlsUriRegex []string `json:"skip_mtls_uri_regex,omitempty"`
}

// SSL ...
type SSL struct {
	BaseInfo
	Cert          string            `json:"cert,omitempty"`
	Key           string            `json:"key,omitempty"`
	Sni           string            `json:"sni,omitempty"`
	Snis          []string          `json:"snis,omitempty"`
	Certs         []string          `json:"certs,omitempty"`
	Type          string            `json:"type,omitempty"`
	Keys          []string          `json:"keys,omitempty"`
	ExpTime       int64             `json:"exptime,omitempty"`
	Status        int               `json:"status"`
	ValidityStart int64             `json:"validity_start,omitempty"`
	ValidityEnd   int64             `json:"validity_end,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Client        *SSLClient        `json:"client,omitempty"`
	SSLProtocols  []string          `json:"ssl_protocols,omitempty"`
}

// Proto ...
type Proto struct {
	BaseInfo
	Desc    string `json:"desc,omitempty"`
	Content string `json:"content"`
}

// StreamRouteProtocol ...
type StreamRouteProtocol struct {
	Name string         `json:"name,omitempty"`
	Conf map[string]any `json:"conf,omitempty"`
}

// StreamRoute ...
type StreamRoute struct {
	BaseInfo
	Desc       string               `json:"desc,omitempty"`
	RemoteAddr string               `json:"remote_addr,omitempty"`
	ServerAddr string               `json:"server_addr,omitempty"`
	ServerPort int                  `json:"server_port,omitempty"`
	SNI        string               `json:"sni,omitempty"`
	UpstreamID any                  `json:"upstream_id,omitempty"`
	Upstream   *UpstreamDef         `json:"upstream,omitempty"`
	ServiceID  any                  `json:"service_id,omitempty"`
	Plugins    map[string]any       `json:"plugins,omitempty"`
	Protocol   *StreamRouteProtocol `json:"protocol,omitempty"`
	Labels     map[string]string    `json:"labels,omitempty"`
}
