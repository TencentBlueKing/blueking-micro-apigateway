/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
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

package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/web/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

func TestFillEtcdTestConnectionSensitiveFieldsDoesNothingWithoutGateway(t *testing.T) {
	req := serializer.EtcdTestConnectionRequest{}
	req.EtcdPassword = "user-input"

	fillEtcdTestConnectionSensitiveFields(&req, nil)

	assert.Equal(t, "user-input", req.EtcdPassword)
}

func TestFillEtcdTestConnectionSensitiveFieldsReplacesMaskedHTTPPassword(t *testing.T) {
	gateway := &model.Gateway{
		ID: 12,
		EtcdConfig: model.EtcdConfig{
			EtcdConfig: base.EtcdConfig{
				Password: "saved-password",
			},
		},
	}
	req := serializer.EtcdTestConnectionRequest{}
	req.EtcdSchemaType = constant.HTTP
	req.EtcdPassword = constant.SensitiveInfoFiledDisplay

	fillEtcdTestConnectionSensitiveFields(&req, gateway)

	assert.Equal(t, gateway.EtcdConfig.Password, req.EtcdPassword)
}

func TestFillEtcdTestConnectionSensitiveFieldsPreservesNonMaskedHTTPPassword(t *testing.T) {
	gateway := &model.Gateway{
		ID: 12,
		EtcdConfig: model.EtcdConfig{
			EtcdConfig: base.EtcdConfig{
				Password: "saved-password",
			},
		},
	}
	req := serializer.EtcdTestConnectionRequest{}
	req.EtcdSchemaType = constant.HTTP
	req.EtcdPassword = "user-provided-password"

	fillEtcdTestConnectionSensitiveFields(&req, gateway)

	assert.Equal(t, "user-provided-password", req.EtcdPassword)
}

func TestFillEtcdTestConnectionSensitiveFieldsReplacesMaskedHTTPSCerts(t *testing.T) {
	gateway := &model.Gateway{
		ID: 34,
		EtcdConfig: model.EtcdConfig{
			EtcdConfig: base.EtcdConfig{
				CACert:   "saved-ca",
				CertCert: "saved-cert",
				CertKey:  "saved-key",
			},
		},
	}
	req := serializer.EtcdTestConnectionRequest{}
	req.EtcdSchemaType = constant.HTTPS
	req.EtcdCACert = gateway.EtcdConfig.GetMaskCaCert()
	req.EtcdCertCert = gateway.EtcdConfig.GetMaskCertCert()
	req.EtcdCertKey = gateway.EtcdConfig.GetMaskCertKey()

	fillEtcdTestConnectionSensitiveFields(&req, gateway)

	assert.Equal(t, gateway.EtcdConfig.CACert, req.EtcdCACert)
	assert.Equal(t, gateway.EtcdConfig.CertCert, req.EtcdCertCert)
	assert.Equal(t, gateway.EtcdConfig.CertKey, req.EtcdCertKey)
}

func TestFillEtcdTestConnectionSensitiveFieldsPreservesNonMaskedHTTPSCerts(t *testing.T) {
	gateway := &model.Gateway{
		ID: 34,
		EtcdConfig: model.EtcdConfig{
			EtcdConfig: base.EtcdConfig{
				CACert:   "saved-ca",
				CertCert: "saved-cert",
				CertKey:  "saved-key",
			},
		},
	}
	req := serializer.EtcdTestConnectionRequest{}
	req.EtcdSchemaType = constant.HTTPS
	req.EtcdCACert = "user-provided-ca"
	req.EtcdCertCert = "user-provided-cert"
	req.EtcdCertKey = "user-provided-key"

	fillEtcdTestConnectionSensitiveFields(&req, gateway)

	assert.Equal(t, "user-provided-ca", req.EtcdCACert)
	assert.Equal(t, "user-provided-cert", req.EtcdCertCert)
	assert.Equal(t, "user-provided-key", req.EtcdCertKey)
}
