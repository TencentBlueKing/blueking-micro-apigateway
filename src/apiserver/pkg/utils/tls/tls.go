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

// Package tls ...
package tls

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/rotisserie/eris"
)

// NewClientTLSConfig ...
func NewClientTLSConfig(caConfig, certConfig, key string) (*tls.Config, error) {
	caPool, err := loadCa(caConfig)
	if err != nil {
		return nil, err
	}
	cert, err := loadCertificates(certConfig, key)
	if err != nil {
		return nil, err
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caPool,
		Certificates:       []tls.Certificate{*cert},
	}

	return conf, nil
}

func loadCa(caConfig string) (*x509.CertPool, error) {
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM([]byte(caConfig)); !ok {
		return nil, eris.Errorf("append ca cert failed")
	}
	return caPool, nil
}

func loadCertificates(cert, key string) (*tls.Certificate, error) {
	tlsCert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}
