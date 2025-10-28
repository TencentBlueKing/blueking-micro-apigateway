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

package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	election "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/leaderelection"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/status"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/goroutinex"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/jsonx"
)

// UnifyOpInterface ...
type UnifyOpInterface interface {
	// SyncerRun 同步资源
	SyncerRun(ctx context.Context, resourceChan chan []*model.GatewaySyncData)
	// SyncWithPrefix 同步指定前缀的资源
	SyncWithPrefix(ctx context.Context, prefix string) (map[constant.APISIXResource]int, error)
	// SyncWithPrefixWithChannel 同步指定前缀的资源，使用 channel 进行落库
	SyncWithPrefixWithChannel(ctx context.Context, prefix string, resourceChan chan []*model.GatewaySyncData) error
	// RevertConfigByIDList 回滚指定资源
	RevertConfigByIDList(ctx context.Context, resourceType constant.APISIXResource, idList []string) error
}

var _ UnifyOpInterface = &UnifyOp{}

// UnifyOp ...
type UnifyOp struct {
	etcdStore   storage.StorageInterface // etcd client
	gatewayInfo *model.Gateway
	elector     *election.EtcdLeaderElector
	isLeader    bool
}

// SyncAll 同步所有资源
func SyncAll(ctx context.Context, resourceChan chan []*model.GatewaySyncData) {
	gateways, err := ListGateways(ctx, 0)
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, gateway := range gateways {
		sy, err := NewUnifyOp(gateway, true)
		if err != nil {
			logging.Errorf("new syncer error: %s", err.Error())
			continue
		}
		goroutinex.GoroutineWithRecovery(ctx, func() {
			sy.SyncerRun(ctx, resourceChan)
		})
	}
}

// RemoveDuplicatedResource 去重重复资源：id重复或者name重复
func RemoveDuplicatedResource(ctx context.Context, resourceType constant.APISIXResource,
	resources []*model.GatewaySyncData,
) ([]*model.GatewaySyncData, error) {
	var syncedResources []*model.GatewaySyncData
	// 查询数据库所有的资源
	resourceList, err := BatchGetResources(ctx, resourceType, []string{})
	if err != nil {
		return syncedResources, err
	}
	resourceNameMap := make(map[string]string, len(resourceList))
	resourceIDMap := make(map[string]string, len(resourceList))
	for _, r := range resourceList {
		resourceNameMap[r.GetName(resourceType)] = r.ID
		resourceIDMap[r.ID] = r.GetName(resourceType)
	}
	for _, r := range resources {
		if id, ok := resourceNameMap[r.GetName()]; ok {
			// 排除已经添加的资源
			// 如果name存在,且id不一致，则说明存在冲突
			if id != r.ID {
				return syncedResources,
					fmt.Errorf("existed %s [id:%s name:%s]conflict", r.Type, id, r.GetName())
			}
			continue
		}
		if _, ok := resourceIDMap[r.ID]; ok {
			// 去除已经添加的资源
			continue
		}
		syncedResources = append(syncedResources, r)
	}
	return syncedResources, nil
}

// AddSyncedResources 添加同步资源到编辑区
func AddSyncedResources( //nolint:gocyclo
	ctx context.Context,
	idList []string,
) (map[constant.APISIXResource]int, error) {
	// 同步资源统计
	syncedResourceTypeStats := make(map[constant.APISIXResource]int)
	queryParam := map[string]interface{}{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
	}
	if len(idList) != 0 {
		queryParam["id"] = idList
	}
	items, err := QuerySyncedItems(ctx, queryParam)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return syncedResourceTypeStats, nil
	}
	// 获取关联资源
	associatedResources, err := GetSyncItemsAssociatedResources(ctx, items)
	if err != nil {
		return nil, err
	}
	if len(associatedResources) > 0 {
		items = append(items, associatedResources...)
	}
	// 分类
	typeSyncedItemMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	for _, item := range items {
		if _, ok := typeSyncedItemMap[item.Type]; !ok {
			typeSyncedItemMap[item.Type] = []*model.GatewaySyncData{item}
			continue
		}
		typeSyncedItemMap[item.Type] = append(typeSyncedItemMap[item.Type], item)
	}

	// 去重
	for resourceType, itemList := range typeSyncedItemMap {
		itemList, err = RemoveDuplicatedResource(ctx, resourceType, itemList)
		if err != nil {
			return nil, err // Return error if duplicate removal fails
		}
		typeSyncedItemMap[resourceType] = itemList
		for _, item := range itemList {
			syncedResourceTypeStats[item.Type]++
		}
	}

	// 分类同步
	err = InsertSyncedResources(ctx, typeSyncedItemMap, constant.ResourceStatusSuccess)
	if err != nil {
		return nil, err
	}
	return syncedResourceTypeStats, nil
}

// InsertSyncedResources 插入数据
func InsertSyncedResources(
	ctx context.Context,
	typeSyncedItemMap map[constant.APISIXResource][]*model.GatewaySyncData,
	status constant.ResourceStatus,
) error {
	// 分类同步
	var err error
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		err = insertSyncedResourcesModel(ctx, typeSyncedItemMap, status, true)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func insertSyncedResourcesModel(
	ctx context.Context,
	typeSyncedItemMap map[constant.APISIXResource][]*model.GatewaySyncData,
	status constant.ResourceStatus,
	removeDuplicated bool,
) error {
	var err error
	for resourceType, itemList := range typeSyncedItemMap {
		if removeDuplicated {
			itemList, err = RemoveDuplicatedResource(ctx, resourceType, itemList)
			if err != nil {
				return err
			}
		}
		resourceList := SyncedResourceToAPISIXResource(resourceType, itemList, status)
		switch resourceType {
		case constant.Route:
			err = BatchCreateRoutes(ctx, resourceList.([]*model.Route))
			if err != nil {
				return err
			}
		case constant.Service:
			err := BatchCreateServices(ctx, resourceList.([]*model.Service))
			if err != nil {
				return err
			}
		case constant.Upstream:
			err := BatchCreateUpstreams(ctx, resourceList.([]*model.Upstream))
			if err != nil {
				return err
			}
		case constant.PluginConfig:
			err := BatchCreatePluginConfigs(ctx, resourceList.([]*model.PluginConfig))
			if err != nil {
				return err
			}
		case constant.PluginMetadata:
			err := batchCreatePluginMetadatas(ctx, resourceList.([]*model.PluginMetadata))
			if err != nil {
				return err
			}
		case constant.Consumer:
			err := BatchCreateConsumers(ctx, resourceList.([]*model.Consumer))
			if err != nil {
				return err
			}
		case constant.ConsumerGroup:
			err := BatchCreateConsumerGroups(ctx, resourceList.([]*model.ConsumerGroup))
			if err != nil {
				return err
			}
		case constant.GlobalRule:
			err := BatchCreateGlobalRules(ctx, resourceList.([]*model.GlobalRule))
			if err != nil {
				return err
			}
		case constant.SSL:
			err = BatchCreateSSL(ctx, resourceList.([]*model.SSL))
			if err != nil {
				return err
			}
		case constant.Proto:
			err := BatchCreateProtos(ctx, resourceList.([]*model.Proto))
			if err != nil {
				return err
			}
		case constant.StreamRoute:
			err := BatchCreateStreamRoutes(ctx, resourceList.([]*model.StreamRoute))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UploadResources 插入数据
func UploadResources(
	ctx context.Context,
	addResourcesTypeMap map[constant.APISIXResource][]*model.GatewaySyncData,
	updateTypeResourcesTypeMap map[constant.APISIXResource][]*model.GatewaySyncData,
	addSchemas map[string]*model.GatewayCustomPluginSchema,
	updatedSchemas map[string]*model.GatewayCustomPluginSchema,
) error {
	// 分类同步
	var err error
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		ctx = ginx.SetTx(ctx, tx)
		// 先删除再插入
		for resourceType, itemList := range updateTypeResourcesTypeMap {
			var ids []string
			for _, item := range itemList {
				ids = append(ids, item.ID)
			}
			err = DeleteResourceByIDs(ctx, resourceType, ids)
			if err != nil {
				return err
			}
		}
		err = insertSyncedResourcesModel(
			ctx, updateTypeResourcesTypeMap, constant.ResourceStatusUpdateDraft, false)
		if err != nil {
			return err
		}
		err = insertSyncedResourcesModel(
			ctx, addResourcesTypeMap, constant.ResourceStatusCreateDraft, false)
		if err != nil {
			return err
		}
		// 处理自定义插件资源,更新的插件先删除再插入
		var updateSchemaNames []string
		for _, schema := range updatedSchemas {
			updateSchemaNames = append(updateSchemaNames, schema.Name)
		}
		if len(updateSchemaNames) > 0 {
			err = DeleteSchemaByNames(ctx, updateSchemaNames)
			if err != nil {
				return err
			}
		}
		var schemaInfoList []*model.GatewayCustomPluginSchema
		for _, schema := range addSchemas {
			schemaInfoList = append(schemaInfoList, schema)
		}
		for _, schema := range updatedSchemas {
			schemaInfoList = append(schemaInfoList, schema)
		}
		if len(schemaInfoList) > 0 {
			err = BatchCreateSchema(ctx, schemaInfoList)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// GetSyncItemsAssociatedResources 获取同步资源关联的资源
func GetSyncItemsAssociatedResources(
	ctx context.Context,
	items []*model.GatewaySyncData,
) ([]*model.GatewaySyncData, error) {
	idMap := make(map[string]bool)
	for _, item := range items {
		idMap[item.ID] = true
	}
	// 遍历处理依赖相关资源
	var associatedIDs []string
	for _, item := range items {
		serviceID := item.GetServiceID()
		if serviceID != "" && !idMap[serviceID] {
			associatedIDs = append(associatedIDs, serviceID)
			idMap[serviceID] = true
		}
		upstreamID := item.GetUpstreamID()
		if upstreamID != "" && !idMap[upstreamID] {
			associatedIDs = append(associatedIDs, upstreamID)
			idMap[upstreamID] = true
		}
		pluginConfigID := item.GetPluginConfigID()
		if pluginConfigID != "" && !idMap[pluginConfigID] {
			associatedIDs = append(associatedIDs, pluginConfigID)
			idMap[pluginConfigID] = true
		}
		groupID := item.GetGroupID()
		if groupID != "" && !idMap[groupID] {
			associatedIDs = append(associatedIDs, groupID)
			idMap[groupID] = true
		}
		sslID := item.GetGroupID()
		if sslID != "" && !idMap[sslID] {
			associatedIDs = append(associatedIDs, sslID)
			idMap[sslID] = true
		}
	}
	if len(associatedIDs) > 0 {
		associatedItems, err := QuerySyncedItems(ctx, map[string]interface{}{
			"id":         associatedIDs,
			"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
		})
		if err != nil {
			return nil, err
		}
		return associatedItems, nil
	}
	return nil, nil
}

// SyncResources 同步资源
func SyncResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
) (map[constant.APISIXResource]int, error) {
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	prefix := gatewayInfo.EtcdConfig.Prefix
	if !strings.HasPrefix(gatewayInfo.EtcdConfig.Prefix, "/") {
		prefix = fmt.Sprintf("%s/", prefix)
	}
	if resourceType == "" {
		prefix = fmt.Sprintf("%s/%s", prefix, constant.ResourceTypePrefixMap[resourceType])
	}
	syncer, err := NewUnifyOp(gatewayInfo, false)
	if err != nil {
		logging.ErrorFWithContext(ctx, "new syncer error: %s", err.Error())
		return nil, err
	}
	syncedResourceTypeStats, err := syncer.SyncWithPrefix(ctx, prefix)
	if err != nil {
		logging.ErrorFWithContext(ctx, "sync all error: %s", err.Error())
		return nil, err
	}
	return syncedResourceTypeStats, nil
}

// NewUnifyOp 创建 UnifyOp
func NewUnifyOp(gatewayInfo *model.Gateway, needElector bool) (*UnifyOp, error) {
	etcdStore, err := storage.NewEtcdStorage(gatewayInfo.EtcdConfig.EtcdConfig)
	if err != nil {
		return nil, err
	}
	var elector *election.EtcdLeaderElector
	isLeader := true
	if needElector {
		elector, err = election.NewEtcdLeaderElector(etcdStore.GetClient(), gatewayInfo.Name)
		if err != nil {
			return nil, err
		}
		isLeader = false
	}
	return &UnifyOp{
		etcdStore:   etcdStore,
		gatewayInfo: gatewayInfo,
		elector:     elector,
		isLeader:    isLeader,
	}, nil
}

// SyncerRun 定时同步
func (s *UnifyOp) SyncerRun(ctx context.Context, resourceChan chan []*model.GatewaySyncData) {
	s.elector.Run(ctx)
	s.elector.WaitForLeading()
	s.isLeader = s.elector.IsLeader()
	// 随机数种子
	rand.NewSource(time.Now().UnixNano())
	minDelay := 1
	maxDelay := 300
	ticker := time.NewTicker(config.G.Biz.SyncInterval +
		time.Second*time.Duration(rand.Intn(maxDelay-minDelay+1)+minDelay))
	for range ticker.C {
		// prefix可能会更新,再查一次
		gatewayInfo, err := GetGateway(ctx, s.gatewayInfo.ID)
		if err != nil {
			logging.Errorf("get gateway error: %s", err.Error())
			continue
		}
		s.gatewayInfo = gatewayInfo
		_, err = s.SyncWithPrefix(ctx, s.gatewayInfo.EtcdConfig.Prefix)
		if err != nil {
			logging.Errorf("sync all error: %s", err.Error())
		}
	}
}

// SyncWithPrefix 同步 prefix 下面的所有资源
func (s *UnifyOp) SyncWithPrefix(ctx context.Context, prefix string) (map[constant.APISIXResource]int, error) {
	if !s.isLeader {
		return nil, nil
	}
	logging.Infof("syncer[gateway:%s] start", s.gatewayInfo.Name)
	kvList, err := s.etcdStore.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	resourceList := s.kvToResource(kvList)

	// 获取已同步资源
	items, err := QuerySyncedItems(ctx, map[string]interface{}{"gateway_id": s.gatewayInfo.ID})
	if err != nil {
		return nil, err
	}
	syncedResources := make(map[string]struct{})
	for _, item := range items {
		syncedResources[item.ID] = struct{}{}
	}

	// 统计最新同步的资源类型及数量
	syncedResourceTypeStats := make(map[constant.APISIXResource]int)
	for _, resource := range resourceList {
		// 过滤掉已同步的资源
		if _, ok := syncedResources[resource.ID]; !ok {
			syncedResourceTypeStats[resource.Type]++
		}
	}

	u := repo.GatewaySyncData
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		// 先删除后插入
		_, err := tx.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(s.gatewayInfo.ID)).Delete()
		if err != nil {
			return err
		}
		// 更新同步时间
		g := tx.Gateway
		s.gatewayInfo.LastSyncedAt = time.Now()
		_, err = g.WithContext(ctx).Where(g.ID.Eq(s.gatewayInfo.ID)).Select(g.LastSyncedAt).Updates(s.gatewayInfo)
		if err != nil {
			return err
		}
		return tx.GatewaySyncData.WithContext(ctx).CreateInBatches(resourceList, 500)
	})
	if err != nil {
		logging.Errorf("sync gateway:%s resource error: %s", s.gatewayInfo.Name, err.Error())
		return nil, err
	}
	logging.Infof("syncer[gateway:%s] end", s.gatewayInfo.Name)
	return syncedResourceTypeStats, nil
}

// SyncWithPrefixWithChannel 同步 prefix 下面的所有资源，通过 channel 来落库
func (s *UnifyOp) SyncWithPrefixWithChannel(
	ctx context.Context,
	prefix string,
	resourceChannel chan []*model.GatewaySyncData,
) error {
	if !s.isLeader {
		return nil
	}
	logging.Infof("syncer[gateway:%s] start", s.gatewayInfo.Name)
	kvList, err := s.etcdStore.List(ctx, prefix)
	if err != nil {
		return err
	}
	resourceList := s.kvToResource(kvList)
	resourceChannel <- resourceList
	logging.Infof("syncer[gateway:%s] end", s.gatewayInfo.Name)
	return nil
}

var revertConfigByIDListFunc = map[constant.APISIXResource]func(ctx context.Context,
	syncDataList []*model.GatewaySyncData) error{
	constant.Route:          BatchRevertRoutes,
	constant.Service:        BatchRevertServices,
	constant.Upstream:       BatchRevertUpstreams,
	constant.PluginConfig:   BatchRevertPluginConfigs,
	constant.PluginMetadata: BatchRevertPluginMetadatas,
	constant.Consumer:       BatchRevertConsumers,
	constant.ConsumerGroup:  BatchRevertConsumerGroups,
	constant.GlobalRule:     BatchRevertGlobalRules,
	constant.SSL:            BatchRevertSSLs,
	constant.Proto:          BatchRevertProtos,
	constant.StreamRoute:    BatchRevertStreamRoutes,
}

// RevertConfigByIDList 根据 ID 列表，回滚配置
func (s *UnifyOp) RevertConfigByIDList(
	ctx context.Context,
	resourceType constant.APISIXResource,
	idList []string,
) error {
	// 状态机判断
	resources, err := BatchGetResources(ctx, resourceType, idList)
	if err != nil {
		return err
	}
	resourceIDMap := make(map[string]*model.ResourceCommonModel)
	for _, resource := range resources {
		resourceIDMap[resource.ID] = resource
		statusOp := status.NewResourceStatusOp(*resource)
		err = statusOp.CanDo(ctx, constant.OperationTypeRevert)
		if err != nil {
			return fmt.Errorf("resource: %s can not do revert: %s", resource.ID, err.Error())
		}
	}
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	prefix := fmt.Sprintf("%s/%s", gatewayInfo.EtcdConfig.Prefix, constant.ResourceTypePrefixMap[resourceType])
	kvList, err := s.etcdStore.List(ctx, prefix)
	if err != nil {
		return err
	}
	etcdResourceList := s.kvToResource(kvList)
	var needRevertResourceList []*model.GatewaySyncData
	for _, etcdResource := range etcdResourceList {
		// 过滤掉不需要回滚的资源
		if _, ok := resourceIDMap[etcdResource.ID]; !ok {
			continue
		}
		needRevertResourceList = append(needRevertResourceList, etcdResource)
	}
	return revertConfigByIDListFunc[resourceType](ctx, needRevertResourceList)
}

// kvToResource 将 etcd 中的 key-value 转换为资源
func (s *UnifyOp) kvToResource(kvList []storage.KeyValuePair) []*model.GatewaySyncData { //nolint:gocyclo
	var resources []*model.GatewaySyncData
	var metadataNames []string
	metadataNameMap := make(map[string]*model.GatewaySyncData)
	globalRuleIdMap := make(map[string]*model.GatewaySyncData)
	pluginConfigIdMap := make(map[string]*model.GatewaySyncData)
	consumerGroupIdMap := make(map[string]*model.GatewaySyncData)
	protoIdMap := make(map[string]*model.GatewaySyncData)
	streamRouteIdMap := make(map[string]*model.GatewaySyncData)
	var globalRuleIDs []string
	var pluginConfigIDs []string
	var consumerGroupIDs []string
	var protoIDs []string
	var streamRouteIDs []string
	for _, kv := range kvList {
		resourceKeyWithoutPrefix := strings.ReplaceAll(kv.Key, s.gatewayInfo.EtcdConfig.Prefix, "")
		resourceKeyList := strings.Split(resourceKeyWithoutPrefix, "/")
		if len(resourceKeyList) != 3 {
			// key不合法
			logging.Errorf("key is not validate: %s", kv.Key)
			continue
		}
		resourceTypeValue := resourceKeyList[1]
		id := resourceKeyList[2]
		resourceType := constant.ResourcePrefixTypeMap[resourceTypeValue]
		if resourceType == "" {
			logging.Errorf("key is not validate without resource type: %s", kv.Key)
			continue
		}
		resourceInfo := &model.GatewaySyncData{
			ID:          id,
			GatewayID:   s.gatewayInfo.ID,
			Type:        resourceType,
			Config:      datatypes.JSON(kv.Value),
			ModRevision: int(kv.ModRevision),
		}
		// config 中去除 update_time/create_time，避免影响资源的 diff
		resourceInfo.Config, _ = sjson.DeleteBytes(resourceInfo.Config, "update_time")
		resourceInfo.Config, _ = sjson.DeleteBytes(resourceInfo.Config, "create_time")
		if resourceType != constant.PluginMetadata && resourceInfo.GetName() == "" {
			resourceInfo.SetName(fmt.Sprintf("%s_%s", resourceTypeValue, id))
		} else if resourceType == constant.PluginMetadata {
			// 插件元数据的名称就是取id
			resourceInfo.SetName(id)
		}
		resources = append(resources, resourceInfo)
		if resourceType == constant.PluginMetadata {
			metadataNames = append(metadataNames, id)
			metadataNameMap[id] = resourceInfo
		}
		// global rule name 需要特殊处理
		if resourceType == constant.GlobalRule {
			globalRuleIdMap[id] = resourceInfo
			globalRuleIDs = append(globalRuleIDs, id)
		}
		// PluginConfig name 需要特殊处理
		if resourceType == constant.PluginConfig {
			pluginConfigIdMap[id] = resourceInfo
			pluginConfigIDs = append(pluginConfigIDs, id)
		}
		// Consumer id，name 需要特殊处理
		if resourceType == constant.ConsumerGroup {
			consumerGroupIdMap[id] = resourceInfo
			consumerGroupIDs = append(consumerGroupIDs, id)
		}
		// Proto name 需要特殊处理
		if resourceType == constant.Proto {
			protoIdMap[id] = resourceInfo
			protoIDs = append(protoIDs, id)
		}
		// StreamRoute name，labels 需要特殊处理
		if resourceType == constant.StreamRoute {
			streamRouteIdMap[id] = resourceInfo
			streamRouteIDs = append(streamRouteIDs, id)
		}
	}
	if len(metadataNames) > 0 {
		// 反向查找ID
		metadatas, err := QueryPluginMetadatas(
			context.Background(),
			map[string]interface{}{"gateway_id": s.gatewayInfo.ID, "name": metadataNames},
		)
		if err != nil {
			logging.Errorf("SearchPluginMetadata error: %s", err.Error())
			return nil
		}
		idNameMap := make(map[string]string)
		for _, metadata := range metadatas {
			idNameMap[metadata.Name] = metadata.ID
		}
		for _, metadata := range metadataNameMap {
			if _, ok := idNameMap[metadata.ID]; ok {
				metadata.ID = idNameMap[metadata.ID]
			} else {
				metadata.ID = idx.GenResourceID(constant.PluginMetadata)
			}
		}
	}

	// 处理 global rule name
	if len(globalRuleIDs) > 0 {
		globalRules, err := QueryGlobalRules(context.Background(), map[string]interface{}{
			"gateway_id": s.gatewayInfo.ID,
			"id":         globalRuleIDs,
		})
		if err != nil {
			logging.Errorf("SearchGlobalRule error: %s", err.Error())
			return nil
		}
		for _, globalRule := range globalRules {
			if g, ok := globalRuleIdMap[globalRule.ID]; ok {
				g.Config, _ = sjson.SetBytes(g.Config, "name", globalRule.Name)
			}
		}
	}

	// 处理 PluginConfig name
	if len(pluginConfigIDs) > 0 {
		pluginConfigs, err := QueryPluginConfigs(context.Background(), map[string]interface{}{
			"gateway_id": s.gatewayInfo.ID,
			"id":         pluginConfigIDs,
		})
		if err != nil {
			logging.Errorf("SearchPluginConfig error: %s", err.Error())
			return nil
		}
		for _, pluginConfig := range pluginConfigs {
			if g, ok := pluginConfigIdMap[pluginConfig.ID]; ok {
				g.Config, _ = sjson.SetBytes(g.Config, "name", pluginConfig.Name)
			}
		}
	}

	// 处理 ConsumerGroup id，name
	if len(consumerGroupIDs) > 0 {
		consumerGroups, err := QueryConsumerGroups(context.Background(), map[string]interface{}{
			"gateway_id": s.gatewayInfo.ID,
			"id":         consumerGroupIDs,
		})
		if err != nil {
			logging.Errorf("SearchConsumerGroup error: %s", err.Error())
			return nil
		}
		for _, consumerGroup := range consumerGroups {
			if g, ok := consumerGroupIdMap[consumerGroup.ID]; ok {
				g.Config, _ = sjson.SetBytes(g.Config, "id", consumerGroup.ID)
				g.Config, _ = sjson.SetBytes(g.Config, "name", consumerGroup.Name)
			}
		}
	}

	// 处理 Proto name
	if len(protoIDs) > 0 {
		protos, err := QueryProtos(context.Background(), map[string]interface{}{
			"gateway_id": s.gatewayInfo.ID,
			"id":         protoIDs,
		})
		if err != nil {
			logging.Errorf("SearchProto error: %s", err.Error())
			return nil
		}
		for _, proto := range protos {
			if g, ok := protoIdMap[proto.ID]; ok {
				g.Config, _ = sjson.SetBytes(g.Config, "name", proto.Name)
			}
		}
	}

	// 处理 StreamRoute name，labels
	if len(streamRouteIDs) > 0 {
		streamRoutes, err := QueryStreamRoutes(context.Background(), map[string]interface{}{
			"gateway_id": s.gatewayInfo.ID,
			"id":         streamRouteIDs,
		})
		if err != nil {
			logging.Errorf("SearchStreamRoute error: %s", err.Error())
			return nil
		}
		for _, streamRoute := range streamRoutes {
			if g, ok := streamRouteIdMap[streamRoute.ID]; ok {
				g.Config, _ = sjson.SetBytes(g.Config, "name", streamRoute.Name)
				labels := streamRoute.GetLabels()
				if labels != nil {
					g.Config, _ = sjson.SetBytes(g.Config, "labels", labels)
				}
			}
		}
	}
	return resources
}

// SyncedResourceToAPISIXResource 将同步的资源转换为 apisix 的资源
func SyncedResourceToAPISIXResource(
	resourceType constant.APISIXResource,
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) interface{} {
	switch resourceType {
	case constant.Route:
		return syncedResourceToAPISIXRoute(syncedResources, status)
	case constant.Service:
		return syncedServiceToAPISIXRoute(syncedResources, status)
	case constant.Upstream:
		return syncedResourceToAPISIXUpstream(syncedResources, status)
	case constant.PluginConfig:
		return syncedResourceToAPISIXPluginConfig(syncedResources, status)
	case constant.PluginMetadata:
		return syncedResourceToAPISIXPluginMetadata(syncedResources, status)
	case constant.Consumer:
		return syncedResourceToAPISIXConsumer(syncedResources, status)
	case constant.ConsumerGroup:
		return syncedResourceToAPISIXConsumerGroup(syncedResources, status)
	case constant.GlobalRule:
		return syncedResourceToAPISIXGlobalRule(syncedResources, status)
	case constant.SSL:
		return syncedResourceToAPISIXSSL(syncedResources, status)
	case constant.Proto:
		return syncedResourceToAPISIXProto(syncedResources, status)
	case constant.StreamRoute:
		return syncedResourceToAPISIXStreamRoute(syncedResources, status)
	}
	return nil
}

func syncedResourceToAPISIXRoute(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.Route {
	var routes []*model.Route
	var OperationType constant.OperationType
	// 对于同步数量大于 100 的资源，批量操作暂时忽略审计
	if len(syncedResources) > 100 {
		OperationType = constant.OperationOneClickManaged
	}
	for _, syncedResource := range syncedResources {
		routes = append(routes, &model.Route{
			Name:           syncedResource.GetName(),
			ServiceID:      syncedResource.GetServiceID(),
			UpstreamID:     syncedResource.GetUpstreamID(),
			PluginConfigID: syncedResource.GetPluginConfigID(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
			OperationType: OperationType,
		})
	}
	return routes
}

func syncedServiceToAPISIXRoute(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.Service {
	var services []*model.Service
	for _, syncedResource := range syncedResources {
		services = append(services, &model.Service{
			Name:       syncedResource.GetName(),
			UpstreamID: syncedResource.GetUpstreamID(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return services
}

func syncedResourceToAPISIXUpstream(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.Upstream {
	var upstreams []*model.Upstream
	for _, syncedResource := range syncedResources {
		upstreams = append(upstreams, &model.Upstream{
			Name: syncedResource.GetName(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return upstreams
}

func syncedResourceToAPISIXPluginConfig(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.PluginConfig {
	var pluginConfigs []*model.PluginConfig
	for _, syncedResource := range syncedResources {
		pluginConfigs = append(pluginConfigs, &model.PluginConfig{
			Name: syncedResource.GetName(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return pluginConfigs
}

func syncedResourceToAPISIXPluginMetadata(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.PluginMetadata {
	var pluginMetadata []*model.PluginMetadata
	for _, syncedResource := range syncedResources {
		name := syncedResource.GetConfigID()
		if syncedResource.ID == "" {
			syncedResource.ID = idx.GenResourceID(constant.PluginMetadata)
		}
		pluginMetadata = append(pluginMetadata, &model.PluginMetadata{
			Name: name,
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return pluginMetadata
}

func syncedResourceToAPISIXConsumer(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.Consumer {
	var consumers []*model.Consumer
	for _, syncedResource := range syncedResources {
		consumers = append(consumers, &model.Consumer{
			Username: syncedResource.GetName(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    constant.ResourceStatusSuccess,
			},
		})
	}
	return consumers
}

func syncedResourceToAPISIXConsumerGroup(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.ConsumerGroup {
	var consumerGroups []*model.ConsumerGroup
	for _, syncedResource := range syncedResources {
		consumerGroups = append(consumerGroups, &model.ConsumerGroup{
			Name: syncedResource.GetName(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return consumerGroups
}

func syncedResourceToAPISIXGlobalRule(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.GlobalRule {
	var globalRules []*model.GlobalRule
	for _, syncedResource := range syncedResources {
		globalRules = append(globalRules, &model.GlobalRule{
			Name: syncedResource.GetName(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return globalRules
}

func syncedResourceToAPISIXSSL(syncedResources []*model.GatewaySyncData, status constant.ResourceStatus) []*model.SSL {
	var ssls []*model.SSL
	for _, syncedResource := range syncedResources {
		ssls = append(ssls, &model.SSL{
			Name: syncedResource.GetName(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return ssls
}

func syncedResourceToAPISIXProto(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.Proto {
	var protos []*model.Proto
	for _, syncedResource := range syncedResources {
		protos = append(protos, &model.Proto{
			Name: syncedResource.GetName(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return protos
}

func syncedResourceToAPISIXStreamRoute(
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) []*model.StreamRoute {
	var streamRoutes []*model.StreamRoute
	for _, syncedResource := range syncedResources {
		streamRoutes = append(streamRoutes, &model.StreamRoute{
			Name:       syncedResource.GetName(),
			ServiceID:  syncedResource.GetServiceID(),
			UpstreamID: syncedResource.GetUpstreamID(),
			ResourceCommonModel: model.ResourceCommonModel{
				ID:        syncedResource.ID,
				GatewayID: syncedResource.GatewayID,
				Config:    syncedResource.Config,
				Status:    status,
			},
		})
	}
	return streamRoutes
}

// DiffResources 对比资源数据
func DiffResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
	idList []string,
	name string,
	resourceStatus []constant.ResourceStatus,
	isDiffAll bool,
) ([]dto.ResourceChangeInfo, error) {
	diffResourceTypeMap := make(map[constant.APISIXResource][]string) // type:idList
	for _, rt := range constant.ResourceTypeList {
		diffResourceTypeMap[rt] = []string{}
	}
	if resourceType != "" && len(idList) != 0 {
		diffResourceTypeMap[resourceType] = idList
	}
	var result []dto.ResourceChangeInfo
	for _, rT := range constant.ResourceTypeList {
		var resourceName string
		param := map[string]interface{}{
			"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
			"status": []constant.ResourceStatus{
				constant.ResourceStatusCreateDraft,
				constant.ResourceStatusDeleteDraft,
				constant.ResourceStatusUpdateDraft,
			},
		}
		resourcesIDList := diffResourceTypeMap[rT]
		// 单资源对比的时候需要考虑关联的资源 id
		if len(resourcesIDList) != 0 && resourceType != "" {
			param["id"] = resourcesIDList
		}
		// 只对当前请求的资源类型过滤
		if resourceType == rT {
			resourceName = name
			if len(resourceStatus) > 0 {
				param["status"] = resourceStatus
			}
		}
		if resourceType != "" && len(resourcesIDList) == 0 {
			if !isDiffAll {
				// 单类型的资源对比
				continue
			}
		}

		resources, err := QueryResource(ctx, rT, param, resourceName)
		if err != nil {
			return nil, err
		}

		if len(resources) == 0 {
			continue
		}

		resourceTypeDiffResult := dto.ResourceChangeInfo{
			ResourceType: rT,
			AddedCount:   0,
			DeletedCount: 0,
			UpdateCount:  0,
			ChangeDetail: []dto.ResourceChangeDetail{},
		}
		for _, resourceInfo := range resources {
			statusOp := status.NewResourceStatusOp(*resourceInfo)
			afterStatus, err := statusOp.NextStatus(ctx, constant.OperationTypePublish)
			if err != nil {
				return nil, err
			}
			resourceChangeDetail := dto.ResourceChangeDetail{
				ResourceID:   resourceInfo.ID,
				BeforeStatus: resourceInfo.Status,
				Name:         resourceInfo.GetName(rT),
				UpdatedAt:    resourceInfo.UpdatedAt.Unix(),
				AfterStatus:  afterStatus,
			}
			switch resourceInfo.Status {
			case constant.ResourceStatusCreateDraft:
				resourceTypeDiffResult.AddedCount++
				resourceChangeDetail.PublishFrom = constant.OperationTypeCreate
			case constant.ResourceStatusDeleteDraft:
				resourceTypeDiffResult.DeletedCount++
				resourceChangeDetail.PublishFrom = constant.OperationTypeDelete
			case constant.ResourceStatusUpdateDraft:
				resourceTypeDiffResult.UpdateCount++
				resourceChangeDetail.PublishFrom = constant.OperationTypeUpdate
			}
			resourceTypeDiffResult.ChangeDetail = append(resourceTypeDiffResult.ChangeDetail, resourceChangeDetail)
			// 处理关联资源
			serviceID := resourceInfo.GetServiceID()
			if serviceID != "" {
				diffResourceTypeMap[constant.Service] = append(diffResourceTypeMap[constant.Service], serviceID)
			}
			if upstreamID := resourceInfo.GetUpstreamID(); upstreamID != "" {
				diffResourceTypeMap[constant.Upstream] = append(diffResourceTypeMap[constant.Upstream], upstreamID)
			}
			if pluginConfigID := resourceInfo.GetPluginConfigID(); pluginConfigID != "" {
				diffResourceTypeMap[constant.PluginConfig] = append(
					diffResourceTypeMap[constant.PluginConfig], pluginConfigID)
			}
			if consumerGroupID := resourceInfo.GetGroupID(); consumerGroupID != "" {
				diffResourceTypeMap[constant.ConsumerGroup] = append(
					diffResourceTypeMap[constant.ConsumerGroup],
					consumerGroupID,
				)
			}
		}
		result = append(result, resourceTypeDiffResult)
	}

	return result, nil
}

// GetResourceConfigDiffDetail 获取资源差异详情
func GetResourceConfigDiffDetail(ctx context.Context, resourceType constant.APISIXResource, id string) (
	*dto.ResourceDiffDetailResponse, error,
) {
	// todo: 同步资源可能存在延时，基于什么样的策略选择从mysql拿还是从etcd拿
	// 获取同步资源配置
	syncedResourceConfig := json.RawMessage("{}")
	syncedResource, err := GetSyncedItemByID(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if syncedResource != nil {
		syncedResourceConfig = json.RawMessage(syncedResource.Config)
	}
	// 获取编辑区资源配置
	resourceInfo, err := GetResourceByID(ctx, resourceType, id)
	if err != nil {
		return nil, err
	}
	// 删除状态需要特殊处理
	if resourceInfo.Status == constant.ResourceStatusDeleteDraft {
		resourceInfo.Config = datatypes.JSON("{}")
	}
	if resourceType == constant.PluginMetadata {
		resourceInfo.Config = []byte(jsonx.RemoveJsonKey(string(resourceInfo.Config), []string{"name"}))
		syncedResourceConfig = []byte(jsonx.RemoveJsonKey(string(syncedResourceConfig), []string{"name"}))
	}
	return &dto.ResourceDiffDetailResponse{
		EtcdConfig:   syncedResourceConfig,
		EditorConfig: json.RawMessage(resourceInfo.Config),
	}, nil
}

// ExportEtcdResources 导出网关下面的所有资源
func (s *UnifyOp) ExportEtcdResources(ctx context.Context) ([]*model.GatewaySyncData, error) {
	logging.Infof("export [gateway:%s] start", s.gatewayInfo.Name)
	prefix := s.gatewayInfo.EtcdConfig.Prefix
	if !strings.HasPrefix(s.gatewayInfo.EtcdConfig.Prefix, "/") {
		prefix = fmt.Sprintf("%s/", prefix)
	}
	kvList, err := s.etcdStore.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	logging.Infof("export [gateway:%s] end ", s.gatewayInfo.Name)
	return s.kvToResource(kvList), nil
}
