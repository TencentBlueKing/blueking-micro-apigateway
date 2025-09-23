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

// Package base ...
package base

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// LabelMap 存储 label 的 map
type LabelMap map[string][]string

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (l *LabelMap) UnmarshalJSON(value []byte) error {
	var labelData map[string][]string
	err := json.Unmarshal(value, &labelData)
	if err == nil {
		labels := labelData["label"]
		if len(labels) > 0 {
			labelList := make(LabelMap)
			for _, label := range labels {
				splitLabel := strings.Split(fmt.Sprintf("%v", label), ",")
				for _, l := range splitLabel {
					if l == "" {
						continue
					}
					labelData := strings.Split(fmt.Sprintf("%v", l), ":")
					if len(labelData) != 2 {
						return fmt.Errorf("[%s] 标签无效", l)
					}
					key := labelData[0]
					val := labelData[1]
					if _, ok := labelList[key]; !ok {
						labelList[key] = []string{}
					}
					labelList[key] = append(labelList[key], fmt.Sprint(val))
				}
			}
			*l = labelList
		}
	}
	return nil
}

// Endpoint 用来表示一个集群实例地址，通过;分割
type Endpoint string

// Endpoints 用来表示一个集群实例地址的列表
func (e Endpoint) Endpoints() []string {
	return strings.Split(string(e), ";")
}

// String ...
func (e Endpoint) String() string {
	return string(e)
}

// EndpointList 用来表示一个集群实例地址的列表
type EndpointList []string

// EndpointJoin 将 EndpointList 转换成 Endpoint
func (e EndpointList) EndpointJoin() Endpoint {
	return Endpoint(strings.Join(e, ";"))
}

// EtcdConfig ...
type EtcdConfig struct {
	Endpoint Endpoint `json:"endpoint,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Prefix   string   `json:"prefix,omitempty"`
	CACert   string   `json:"ca_cert,omitempty"`
	CertCert string   `json:"cert_cert,omitempty"`
	CertKey  string   `json:"cert_key,omitempty"`
}

// GetSchemaType 获取 schema 类型
func (e EtcdConfig) GetSchemaType() string {
	if e.Username != "" && e.Password != "" {
		return constant.HTTP
	}
	if e.CertCert != "" && e.CertKey != "" && e.CACert != "" {
		return constant.HTTPS
	}
	return ""
}

// MaintainerList 用来表示一个网关管理员的列表
type MaintainerList []string

// Strip 去除空字符串
func (m MaintainerList) Strip() []string {
	var maintainerList []string
	for _, v := range m {
		v = strings.TrimSpace(v)
		if v != "" && v != `""` {
			maintainerList = append(maintainerList, v)
		}
	}
	return maintainerList
}

var (
	certKeyReg  = regexp.MustCompile(`(?s)(-----BEGIN RSA PRIVATE KEY-----.*?-----END RSA PRIVATE KEY-----)`)
	certificate = regexp.MustCompile(`(?s)(-----BEGIN CERTIFICATE-----.*?-----END CERTIFICATE-----)`)
)

// GetMaskCertCert 获取隐藏的证书内容
func (e EtcdConfig) GetMaskCertCert() string {
	return maskContent(e.CertCert, certificate)
}

// GetMaskCaCert 获取隐藏的证书内容
func (e EtcdConfig) GetMaskCaCert() string {
	return maskContent(e.CACert, certificate)
}

// GetMaskCertKey 获取隐藏的 key 内容
func (e EtcdConfig) GetMaskCertKey() string {
	return maskContent(e.CertKey, certKeyReg)
}

// maskContent 获取隐藏的证书内容
func maskContent(content string, re *regexp.Regexp) string {
	if strings.TrimSpace(content) == "" {
		return content
	}
	matches := re.FindStringSubmatch(content)
	if len(matches) < 1 {
		return content // 如果没有找到，返回原始字符串
	}
	// 获取证书内容并掩盖
	certContent := matches[1]
	contentLines := strings.Split(certContent, "\n")
	// 仅掩盖中间的内容
	if len(contentLines) > 2 {
		startLine := contentLines[0]
		endLine := contentLines[len(contentLines)-1]

		// 获取中间的内容
		middleContent := strings.Join(contentLines[1:len(contentLines)-1], "\n")
		middleContentRunes := []rune(middleContent)
		// 保留前后 6 位
		var maskedContent string
		if len(middleContentRunes) > 12 {
			maskedContent = string(middleContentRunes[:6]) +
				strings.Repeat("*", len(middleContentRunes)-12) +
				string(middleContentRunes[len(middleContentRunes)-6:])
		} else {
			maskedContent = middleContent // 如果内容长度小于等于 12，直接返回原内容
		}
		return fmt.Sprintf("%s\n%s\n%s", startLine, maskedContent, endLine)
	}
	return certContent // 如果没有足够的内容行，则返回原始内容
}
