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

package biz

import (
	"fmt"
	"strings"

	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
)

// buildSyncedResourceFromKV normalizes one etcd KV into a GatewaySyncData snapshot.
//
// Returns (nil, false) when the KV key cannot be parsed. Caller MUST emit
// logging.Errorf("key is not validate: %s", kv.Key) on the false branch so that
// the existing operational observability is preserved.
//
// The internal branch layout is slightly different from the original inline
// code in kvToResource(...) (if-else-if -> explicit switch on resource type),
// but behavior is equivalent: PluginMetadata always has SetName(id); other
// types set fallback name only when GetName() == "".
func buildSyncedResourceFromKV(
	normalizedPrefix string,
	gatewayID int,
	kv storage.KeyValuePair,
) (*model.GatewaySyncData, bool) {
	resourceKeyWithoutPrefix := strings.TrimPrefix(kv.Key, normalizedPrefix)
	resourceKeyList := strings.Split(resourceKeyWithoutPrefix, "/")
	if len(resourceKeyList) != 2 {
		return nil, false
	}

	resourceTypeValue := resourceKeyList[0]
	id := resourceKeyList[1]
	resourceType := constant.ResourcePrefixTypeMap[resourceTypeValue]
	if resourceType == "" {
		return nil, false
	}

	resourceInfo := &model.GatewaySyncData{
		ID:          id,
		GatewayID:   gatewayID,
		Type:        resourceType,
		Config:      datatypes.JSON(kv.Value),
		ModRevision: int(kv.ModRevision),
	}
	resourceInfo.Config, _ = sjson.DeleteBytes(resourceInfo.Config, "update_time")
	resourceInfo.Config, _ = sjson.DeleteBytes(resourceInfo.Config, "create_time")

	if resourceType == constant.PluginMetadata {
		resourceInfo.SetName(id)
	} else if resourceInfo.GetName() == "" {
		resourceInfo.SetName(fmt.Sprintf("%s_%s", resourceTypeValue, id))
	}

	return resourceInfo, true
}
