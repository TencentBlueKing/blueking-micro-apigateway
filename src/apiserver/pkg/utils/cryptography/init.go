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

package cryptography

import (
	"fmt"
)

var commonCrypto *AESGcm

// Init ...
func Init(encryptKey string, nonce string) (err error) {
	commonCrypto, err = NewAESGcm([]byte(encryptKey), []byte(nonce))
	if err != nil {
		return fmt.Errorf("cryptos[id=app_secret_key] key error: %w", err)
	}
	return nil
}

// DecryptSecret ...
func DecryptSecret(encryptedSecret string) (string, error) {
	return commonCrypto.DecryptFromBase64(encryptedSecret)
}

// EncryptSecret ...
func EncryptSecret(plainSecret string) string {
	return commonCrypto.EncryptToBase64(plainSecret)
}
