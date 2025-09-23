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
package tls_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/tls"
)

func TestTLS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TLS Suite")
}

func generateTestCert() (caPEM, certPEM, keyPEM string) {
	// 生成CA证书
	caPrivKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caBytes, _ := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivKey.PublicKey, caPrivKey)
	caPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}))

	// 生成服务端证书
	certPrivKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test Server"},
		},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 5},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certBytes, _ := x509.CreateCertificate(rand.Reader, &certTemplate, &caTemplate, &certPrivKey.PublicKey, caPrivKey)
	certPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}))

	// 生成私钥
	keyPEM = string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}))

	return caPEM, certPEM, keyPEM
}

var _ = Describe("NewClientTLSConfig", func() {
	Context("when given valid CA, cert, and key", func() {
		It("should return a valid tls.Config", func() {
			ca, cert, key := generateTestCert()

			config, err := tls.NewClientTLSConfig(ca, cert, key)
			Expect(err).To(BeNil())
			Expect(config).NotTo(BeNil())
			Expect(config.Certificates).To(HaveLen(1))
			Expect(config.RootCAs).NotTo(BeNil())
			Expect(config.InsecureSkipVerify).To(BeTrue())
		})
	})

	Context("when given empty CA", func() {
		It("should return an error", func() {
			_, cert, key := generateTestCert()

			config, err := tls.NewClientTLSConfig("", cert, key)
			Expect(err).NotTo(BeNil())
			Expect(config).To(BeNil())
		})
	})

	Context("when given empty cert and key", func() {
		It("should return an error", func() {
			ca, _, _ := generateTestCert()

			config, err := tls.NewClientTLSConfig(ca, "", "")
			Expect(err).NotTo(BeNil())
			Expect(config).To(BeNil())
		})
	})

	Context("when given mismatched cert and key", func() {
		It("should return an error", func() {
			ca, cert, _ := generateTestCert()
			_, wrongKeyPEM := func() (string, string) {
				wrongKey, _ := rsa.GenerateKey(rand.Reader, 2048)
				return "", string(pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(wrongKey),
				}))
			}()

			config, err := tls.NewClientTLSConfig(ca, cert, wrongKeyPEM)
			Expect(err).NotTo(BeNil())
			Expect(config).To(BeNil())
		})
	})

	Context("when given invalid CA", func() {
		It("should return an error", func() {
			_, cert, key := generateTestCert()

			config, err := tls.NewClientTLSConfig("invalid-ca", cert, key)
			Expect(err).NotTo(BeNil())
			Expect(config).To(BeNil())
		})
	})

	Context("when given invalid cert and key", func() {
		It("should return an error", func() {
			ca, _, _ := generateTestCert()

			config, err := tls.NewClientTLSConfig(ca, "invalid-cert", "invalid-key")
			Expect(err).NotTo(BeNil())
			Expect(config).To(BeNil())
		})
	})
})
