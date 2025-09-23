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

package jsonx

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJSONEmpty(t *testing.T) {
	tests := []struct {
		name     string
		raw      json.RawMessage
		expected bool
	}{
		{"Empty Array", json.RawMessage(`[]`), true},
		{"Empty Object", json.RawMessage(`{}`), true},
		{"Non-empty Array", json.RawMessage(`[1, 2]`), false},
		{"Non-empty Object", json.RawMessage(`{"key": "value"}`), false},
		{"Null Value", json.RawMessage(`null`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := IsJSONEmpty(tt.raw)
			if isEmpty != tt.expected {
				t.Errorf("isJSONEmpty() = %v, expected %v", isEmpty, tt.expected)
			}
		})
	}
}

func TestMergeJson(t *testing.T) {
	doc := []byte(`{"key1": "value1", "key2": "value2"}`)
	patch := []byte(`{"key2": "newvalue2", "key3": "value3"}`)

	expected := []byte(`{"key1": "value1", "key2": "newvalue2", "key3": "value3"}`)

	out, err := MergeJson(doc, patch)
	assert.NoError(t, err)
	assert.JSONEq(t, string(expected), string(out))
}

func TestMergeJson_ErrorCases(t *testing.T) {
	tests := []struct {
		name  string
		doc   []byte
		patch []byte
	}{
		{"Invalid Doc", []byte(`{`), []byte(`{}`)},
		{"Invalid Patch", []byte(`{}`), []byte(`{`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := MergeJson(tt.doc, tt.patch)
			assert.Error(t, err)
		})
	}
}

func TestPatchJson(t *testing.T) {
	doc := []byte(`{"key1": "value1", "key2": "value2"}`)

	out, err := PatchJson(doc, "/key2", `"newvalue2"`)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"key1": "value1", "key2": "newvalue2"}`, string(out))

	// Test add operation
	out, err = PatchJson(doc, "/key3", `"value3"`)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"key1": "value1", "key2": "value2", "key3": "value3"}`, string(out))
}

func TestPatchJson_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		doc  []byte
		path string
		val  string
	}{
		{"Invalid Doc", []byte(`{`), "/key", `"value"`},
		{"Invalid Path", []byte(`{}`), "invalidpath", `"value"`},
		{"Invalid Value", []byte(`{}`), "/key", `invalid`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PatchJson(tt.doc, tt.path, tt.val)
			assert.Error(t, err)
		})
	}
}

func TestMergePatch(t *testing.T) {
	obj := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	reqBody := []byte(`{"key2": "newvalue2", "key3": "value3"}`)

	expected := []byte(`{"key1": "value1", "key2": "newvalue2", "key3": "value3"}`)

	// Test without subPath
	res, err := MergePatch(obj, "", reqBody)
	assert.NoError(t, err)
	assert.JSONEq(t, string(expected), string(res))

	// Test with subPath
	reqBody = []byte(`{"key4": "value4"}`)
	expected = []byte(`{"key1": "value1", "key2": "value2", "key3": {"key4": "value4"}}`)
	res, err = MergePatch(obj, "/key3", reqBody)
	assert.NoError(t, err)
	assert.JSONEq(t, string(expected), string(res))
}

func TestRemoveEmptyObjects(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input: `{
				"name": "John",
				"age": 30,
				"address": {},
				"email": "john.doe@example.com",
				"details": {
					"height": 175,
					"weight": {},
					"plugins": {}
				}
			}`,
			expected: `{
				"name": "John",
				"age": 30,
				"email": "john.doe@example.com",
				"details": {
					"height": 175,
					"plugins": {}
				}
			}`,
		},
		{
			input: `{
				"empty": {},
				"notEmpty": {"key": "value"}
			}`,
			expected: `{
				"notEmpty": {"key": "value"}
			}`,
		},
		{
			input:    `{}`,
			expected: `{}`,
		},
		{
			input: `{
				"level1": {
					"level2": {}
				}
			}`,
			expected: `{
			}`,
		},
		{
			input: `{
				"level1": {
					"level2":[]
				}
			}`,
			expected: `{
			}`,
		},
		{
			input: `{
				"level1": {
					"level2":[
                       {}
                     ]
				}
			}`,
			expected: `{
			}`,
		},
		{
			input: `{
				"level1": {
					"level2":[
                       {},
                       {
                          "level3":123
                        }
                     ]
				}
			}`,
			expected: `{
                "level1": {
					"level2":[
                       {
                          "level3":123
                        }
                     ]
				}
			}`,
		},
		{
			input: `{
				"level1": {
					"level2":[
                       {
                         []
                        },
                       {
                          "level3":123
                        }
                     ]
				}
			}`,
			expected: `{
                "level1": {
					"level2":[
                       {
                          "level3":123
                        }
                     ]
				}
			}`,
		},
	}

	for _, test := range tests {
		result, err := RemoveEmptyObjectsAndArrays(test.input)
		if err != nil {
			t.Errorf("Error removing empty objects: %v", err)
		}

		// 为了比较两个JSON字符串，去掉所有空格和换行符
		if compactJSON(result) != compactJSON(test.expected) {
			t.Errorf("Unexpected result. Got: %s, Expected: %s", result, test.expected)
		}
	}
}

// compactJSON removes all whitespace and newlines for a compact comparison
func compactJSON(jsonStr string) string {
	res := ""
	for _, char := range jsonStr {
		if char != ' ' && char != '\n' && char != '\t' {
			res += string(char)
		}
	}
	return res
}

func TestRemoveJsonKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input: `{
				"name": "John",
				"age": 30,
				"email": "john.doe@example.com"
			}`,
			expected: `{
				"name": "John"
			}`,
		},
	}

	for _, test := range tests {
		result := RemoveJsonKey(test.input, []string{"age", "email"})

		// 为了比较两个JSON字符串，去掉所有空格和换行符
		if compactJSON(result) != compactJSON(test.expected) {
			t.Errorf("Unexpected result. Got: %s, Expected: %s", result, test.expected)
		}
	}
}
