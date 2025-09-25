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

// filex ...
package filex

import (
	"bytes"
	"encoding/json"
	"mime/multipart"

	"github.com/pkg/errors"
)

// ReadFileToObject 读取文件内容到对象中
func ReadFileToObject(fileHeader *multipart.FileHeader, obj interface{}) error {
	file, err := fileHeader.Open()
	if err != nil {
		return errors.Wrap(err, "open file failed")
	}
	defer file.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return errors.Wrap(err, "read file failed")
	}
	rawData := buf.Bytes()
	if err := json.Unmarshal(rawData, obj); err != nil {
		return errors.Wrap(err, "unmarshal file failed")
	}
	return nil
}
