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

package syncdata

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

// ComparableConfig removes only known control-plane fields. Plugin payloads are opaque:
// unknown plugins, empty plugin objects, arrays and nested timestamps must survive.
func ComparableConfig(config json.RawMessage, resourceType constant.APISIXResource,
	version constant.APISIXVersion,
) (map[string]any, error) {
	var value map[string]any
	decoder := json.NewDecoder(bytes.NewReader(config))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("resource config must be an object")
	}
	for _, field := range []string{"id", "create_time", "update_time", "created_at", "updated_at"} {
		delete(value, field)
	}
	if !constant.ResourceSupportsNameFieldForVersion(resourceType, version) {
		delete(value, "name")
	}
	if resourceType == constant.SSL {
		delete(value, "validity_start")
		delete(value, "validity_end")
	}
	if resourceType == constant.StreamRoute {
		delete(value, "labels")
	}
	// Model-owned references accept numeric IDs in etcd but are materialized as strings in SQL.
	var references []string
	switch resourceType {
	case constant.Route:
		references = []string{"service_id", "upstream_id", "plugin_config_id"}
	case constant.Service:
		references = []string{"upstream_id"}
	case constant.StreamRoute:
		references = []string{"service_id", "upstream_id"}
	case constant.Consumer:
		references = []string{"group_id"}
	}
	for _, field := range references {
		if _, ok := value[field].(json.Number); ok {
			value[field] = gjson.GetBytes(config, field).String()
		}
	}
	if resourceType == constant.Upstream {
		if tls, ok := value["tls"].(map[string]any); ok {
			if _, numeric := tls["client_cert_id"].(json.Number); numeric {
				tls["client_cert_id"] = gjson.GetBytes(config, "tls.client_cert_id").String()
			}
		}
	}
	return value, nil
}

// ConfigStatus compares a snapshot with the matching edit-area row.
func ConfigStatus(snapshot *model.GatewaySyncData, editor *model.ResourceCommonModel,
	version constant.APISIXVersion,
) (constant.SyncStatus, error) {
	if editor == nil {
		return constant.SyncedResourceStatusMiss, nil
	}
	before, err := ComparableConfig(json.RawMessage(snapshot.Config), snapshot.Type, version)
	if err != nil {
		return "", err
	}
	after, err := ComparableConfig(json.RawMessage(editor.Config), snapshot.Type, version)
	if err != nil {
		return "", err
	}
	if !reflect.DeepEqual(before, after) || editor.Status == constant.ResourceStatusDeleteDraft {
		return constant.SyncedResourceStatusDifferent, nil
	}
	return constant.SyncedResourceStatusSuccess, nil
}

// SyncToken binds an overwrite confirmation to both persisted revisions, including draft state.
func SyncToken(snapshot *model.GatewaySyncData, editor *model.ResourceCommonModel) string {
	payload, err := json.Marshal([]string{
		strconv.Itoa(snapshot.GatewayID),
		string(snapshot.Type),
		snapshot.ID,
		strconv.Itoa(snapshot.ModRevision),
		string(snapshot.Config),
		strconv.Itoa(editor.AutoID),
		string(editor.Config),
		string(editor.Status),
		strconv.FormatInt(editor.UpdatedAt.UnixNano(), 10),
	})
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(payload))
}
