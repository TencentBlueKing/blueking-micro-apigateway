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

// Package jsonx ...
package jsonx

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// IsJSONEmpty 判断 json 是否为空
func IsJSONEmpty(raw json.RawMessage) bool {
	// 尝试解析为一个空接口
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return false
	}
	// 根据解析的类型判断是否为空
	switch v := data.(type) {
	case map[string]any:
		return len(v) == 0 // 对象是否为空
	case []any:
		return len(v) == 0 // 数组是否为空
	default:
		return false // 其他类型不被认为是空
	}
}

// MergeJson ...
func MergeJson(doc, patch []byte) ([]byte, error) {
	out, err := jsonpatch.MergePatch(doc, patch)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PatchJson ...
func PatchJson(doc []byte, path, val string) ([]byte, error) {
	patch := []byte(`[ { "op": "replace", "path": "` + path + `", "value": ` + val + `}]`)
	obj, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return nil, err
	}

	out, err := obj.Apply(doc)
	if err != nil {
		// try to add if field not exist
		patch = []byte(`[ { "op": "add", "path": "` + path + `", "value": ` + val + `}]`)
		obj, err = jsonpatch.DecodePatch(patch)
		if err != nil {
			return nil, err
		}
		out, err = obj.Apply(doc)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// MergePatch ...
func MergePatch(obj any, subPath string, reqBody []byte) ([]byte, error) {
	var res []byte
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return res, err
	}

	if subPath != "" {
		res, err = PatchJson(jsonBytes, subPath, string(reqBody))
	} else {
		res, err = MergeJson(jsonBytes, reqBody)
	}

	if err != nil {
		return res, err
	}
	return res, nil
}

// RemoveEmptyObjectsAndArrays 递归地删除 JSON 字符串中值为空对象或空数组的所有键。
func RemoveEmptyObjectsAndArrays(jsonStr string) (string, error) {
	// 使用 gjson 解析 JSON 字符串
	result := gjson.Parse(jsonStr)
	// 如果当前节点是一个对象
	if result.IsObject() {
		// 遍历对象的每个键值对
		for key, value := range result.Map() {
			if key == "plugins" {
				// 跳过 plugins 字段，该字段比较特殊，需要保留
				continue
			}
			if value.IsObject() || value.IsArray() {
				// 递归处理子对象或数组
				updatedValue, err := RemoveEmptyObjectsAndArrays(value.Raw)
				if err != nil {
					return "", err
				}

				// 解析更新后的值
				updatedResult := gjson.Parse(updatedValue)

				// 如果更新后的值是空对象或空数组，删除该键
				if (updatedResult.IsObject() && len(updatedResult.Map()) == 0) ||
					(updatedResult.IsArray() && len(updatedResult.Array()) == 0) {
					jsonStr, err = sjson.Delete(jsonStr, key)
					if err != nil {
						return "", err
					}
				} else {
					// 否则，更新该键的值
					jsonStr, err = sjson.SetRaw(jsonStr, key, updatedValue)
					if err != nil {
						return "", err
					}
				}
			}
		}
	} else if result.IsArray() {
		// 如果当前节点是一个数组，从后向前遍历
		for i := len(result.Array()) - 1; i >= 0; i-- {
			elem := result.Array()[i]
			if elem.IsObject() || elem.IsArray() {
				// 递归处理数组中的对象或数组
				updatedElem, err := RemoveEmptyObjectsAndArrays(elem.Raw)
				if err != nil {
					return "", err
				}
				// 解析更新后的元素
				updatedElemResult := gjson.Parse(updatedElem)

				// 检查更新后的元素是否为空对象或数组
				if (updatedElemResult.IsObject() && len(updatedElemResult.Map()) == 0) ||
					(updatedElemResult.IsArray() && len(updatedElemResult.Array()) == 0) {
					// 删除空元素
					jsonStr, err = sjson.Delete(jsonStr, fmt.Sprintf("%d", i))
					if err != nil {
						return "", err
					}
				} else {
					// 更新非空元素
					jsonStr, err = sjson.SetRaw(jsonStr, fmt.Sprintf("%d", i), updatedElem)
					if err != nil {
						return "", err
					}
				}
			}
		}
	}
	return jsonStr, nil
}

// RemoveJsonKey 删除 JSON 字符串中指定的键
func RemoveJsonKey(jsonStr string, keys []string) string {
	for _, k := range keys {
		res, err := sjson.Delete(jsonStr, k)
		if err == nil {
			jsonStr = res
		}
	}
	return jsonStr
}
