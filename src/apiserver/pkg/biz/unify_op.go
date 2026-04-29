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
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
	SyncerRun(ctx context.Context)
	// SyncWithPrefix 同步指定前缀的资源
	SyncWithPrefix(ctx context.Context, prefix string) (map[constant.APISIXResource]int, error)
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
func SyncAll(ctx context.Context) {
	gateways, err := ListGateways(ctx, 0)
	if err != nil {
		logging.Errorf("list gateways error: %s", err.Error())
		return
	}
	for _, gateway := range gateways {
		sy, err := NewUnifyOp(gateway, true)
		if err != nil {
			logging.Errorf("new syncer error: %s", err.Error())
			continue
		}
		goroutinex.GoroutineWithRecovery(ctx, func() {
			sy.SyncerRun(ctx)
		})
	}
}

// RemoveDuplicatedResource 去重重复资源：id 重复或者 name 重复
func RemoveDuplicatedResource(
	ctx context.Context,
	resourceType constant.APISIXResource,
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
			// 如果 name 存在，且 id 不一致，则说明存在冲突
			if id != r.ID {
				return syncedResources,
					fmt.Errorf("existed %s [id:%s name:%s] conflict", r.Type, id, r.GetName())
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
	queryParam := map[string]any{
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
		ctx := ginx.SetTx(ctx, tx)
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
		// type is guaranteed by SyncedResourceToAPISIXResource per resourceType
		resourceList := SyncedResourceToAPISIXResource(resourceType, itemList, status)
		switch resourceType {
		case constant.Route:
			err = BatchCreateRoutes(ctx, resourceList.([]*model.Route)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		case constant.Service:
			err = BatchCreateServices(ctx, resourceList.([]*model.Service)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		case constant.Upstream:
			err = BatchCreateUpstreams(ctx, resourceList.([]*model.Upstream)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		case constant.PluginConfig:
			//nolint:forcetypeassert
			err = BatchCreatePluginConfigs(ctx, resourceList.([]*model.PluginConfig))
			if err != nil {
				return err
			}
		case constant.PluginMetadata:
			//nolint:forcetypeassert
			err = batchCreatePluginMetadatas(ctx, resourceList.([]*model.PluginMetadata))
			if err != nil {
				return err
			}
		case constant.Consumer:
			err = BatchCreateConsumers(ctx, resourceList.([]*model.Consumer)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		case constant.ConsumerGroup:
			//nolint:forcetypeassert
			err = BatchCreateConsumerGroups(ctx, resourceList.([]*model.ConsumerGroup))
			if err != nil {
				return err
			}
		case constant.GlobalRule:
			err = BatchCreateGlobalRules(ctx, resourceList.([]*model.GlobalRule)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		case constant.SSL:
			err = BatchCreateSSL(ctx, resourceList.([]*model.SSL)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		case constant.Proto:
			err = BatchCreateProtos(ctx, resourceList.([]*model.Proto)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		case constant.StreamRoute:
			err = BatchCreateStreamRoutes(ctx, resourceList.([]*model.StreamRoute)) //nolint:forcetypeassert
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
		ctx := ginx.SetTx(ctx, tx)
		// 先删除再插入
		for resourceType, itemList := range updateTypeResourcesTypeMap {
			var ids []string
			for _, item := range itemList {
				if resourceType == constant.PluginMetadata {
					ids = append(ids, item.GetName())
					continue
				}
				ids = append(ids, item.ID)
			}
			err = DeleteResourceByIDs(ctx, resourceType, ids)
			if err != nil {
				return err
			}
		}
		err = insertSyncedResourcesModel(
			ctx,
			updateTypeResourcesTypeMap,
			constant.ResourceStatusUpdateDraft,
			false,
		)
		if err != nil {
			return err
		}
		err = insertSyncedResourcesModel(
			ctx, addResourcesTypeMap, constant.ResourceStatusCreateDraft, false)
		if err != nil {
			return err
		}
		// 处理自定义插件 schema，直接更新
		var updateSchemaNames []string
		updateSchemaMap := make(map[string]*model.GatewayCustomPluginSchema)
		for _, schema := range updatedSchemas {
			updateSchemaNames = append(updateSchemaNames, schema.Name)
			updateSchemaMap[schema.Name] = schema
		}
		if len(updateSchemaNames) > 0 {
			existingSchemas, err := BatchGetSchemaByName(ctx, updateSchemaNames)
			if err != nil {
				return err
			}
			for _, schema := range existingSchemas {
				if updateSchema, ok := updateSchemaMap[schema.Name]; ok {
					schema.Schema = updateSchema.Schema
					schema.Example = updateSchema.Example
					schema.OperationType = updateSchema.OperationType
					schema.Updater = updateSchema.Updater
				}
			}
			err = BatchUpdateSchema(ctx, existingSchemas)
			if err != nil {
				return err
			}
		}
		var schemaInfoList []*model.GatewayCustomPluginSchema
		for _, schema := range addSchemas {
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
	idMap := make(map[string]bool) // type:id 不同类型资源的 id 可能重复
	for _, item := range items {
		idMap[item.GetResourceKey()] = true
	}
	// 遍历处理依赖相关资源
	var associatedIDs []string
	for _, item := range items {
		serviceID := item.GetServiceID()
		if serviceID != "" && !idMap[item.GetResourceKey()] {
			associatedIDs = append(associatedIDs, serviceID)
			idMap[item.GetResourceKey()] = true
		}
		upstreamID := item.GetUpstreamID()
		if upstreamID != "" && !idMap[item.GetResourceKey()] {
			associatedIDs = append(associatedIDs, upstreamID)
			idMap[item.GetResourceKey()] = true
		}
		pluginConfigID := item.GetPluginConfigID()
		if pluginConfigID != "" && !idMap[item.GetResourceKey()] {
			associatedIDs = append(associatedIDs, pluginConfigID)
			idMap[item.GetResourceKey()] = true
		}
		groupID := item.GetGroupID()
		if groupID != "" && !idMap[item.GetResourceKey()] {
			associatedIDs = append(associatedIDs, groupID)
			idMap[item.GetResourceKey()] = true
		}
		sslID := item.GetSSLID()
		if sslID != "" && !idMap[item.GetResourceKey()] {
			associatedIDs = append(associatedIDs, sslID)
			idMap[item.GetResourceKey()] = true
		}
	}
	if len(associatedIDs) > 0 {
		associatedItems, err := QuerySyncedItems(ctx, map[string]any{
			"id": associatedIDs,
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
	var prefix string
	// 如果资源类型为空，则同步所有资源
	if resourceType != "" {
		// 使用标准化的资源类型 prefix
		prefix = gatewayInfo.GetEtcdResourcePrefix(resourceType)
	} else {
		// 使用标准化的网关 prefix（带 "/" 结尾，避免前缀匹配冲突）
		prefix = gatewayInfo.GetEtcdPrefixForList()
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
func (s *UnifyOp) SyncerRun(ctx context.Context) {
	s.elector.Run(ctx)
	s.elector.WaitForLeading()
	s.isLeader = s.elector.IsLeader()
	// 随机数种子
	rand.NewSource(time.Now().UnixNano())
	minDelay := 1
	maxDelay := 300
	ticker := time.NewTicker(config.G.Biz.SyncInterval +
		time.Second*time.Duration(
			rand.Intn(maxDelay-minDelay+1)+minDelay, //nolint:gosec // G404: non-security random for jitter
		))
	for range ticker.C {
		// prefix 可能会更新，再查一次
		gatewayInfo, err := GetGateway(ctx, s.gatewayInfo.ID)
		if err != nil {
			logging.Errorf("get gateway error: %s", err.Error())
			continue
		}
		s.gatewayInfo = gatewayInfo
		gatewayCtx := context.Background()
		gatewayCtx = ginx.SetGatewayInfoToContext(gatewayCtx, s.gatewayInfo)
		// 使用标准化的 prefix（带 "/" 结尾，避免前缀匹配冲突）
		_, err = s.SyncWithPrefix(gatewayCtx, s.gatewayInfo.GetEtcdPrefixForList())
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
	resourceList, err := s.kvToResource(ctx, kvList)
	if err != nil {
		return nil, err
	}

	// 获取已同步资源
	syncedItems, err := QuerySyncedItems(ctx, map[string]any{})
	if err != nil {
		return nil, err
	}
	databaseResourceMap := make(map[string]*model.GatewaySyncData)
	for _, item := range syncedItems {
		databaseResourceMap[item.GetResourceKey()] = item
	}
	changeSet := buildSyncChangeSet(resourceList, syncedItems)

	u := repo.GatewaySyncData
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		// Execute updates in batches using ON CONFLICT (upsert)
		if len(changeSet.ToUpdate) > 0 {
			err = tx.GatewaySyncData.WithContext(ctx).
				Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "auto_id"}},
					DoUpdates: clause.AssignmentColumns(
						[]string{"config", "mod_revision", "updated_at"},
					),
				}).
				CreateInBatches(changeSet.ToUpdate, 500)
			if err != nil {
				return err
			}
		}

		// Execute creates in batches
		if len(changeSet.ToCreate) > 0 {
			err = tx.GatewaySyncData.WithContext(ctx).CreateInBatches(changeSet.ToCreate, 500)
			if err != nil {
				return err
			}
		}

		// Execute deletes in batches
		if len(changeSet.ToDeleteAutoIDs) > 0 {
			for i := 0; i < len(changeSet.ToDeleteAutoIDs); i += 500 {
				end := min(i+500, len(changeSet.ToDeleteAutoIDs))
				batch := changeSet.ToDeleteAutoIDs[i:end]
				_, err = tx.GatewaySyncData.WithContext(ctx).
					Where(u.AutoID.In(batch...)).
					Delete()
				if err != nil {
					return err
				}
			}
		}

		// always update the sync time
		g := tx.Gateway
		s.gatewayInfo.LastSyncedAt = time.Now()
		_, err = g.WithContext(
			ctx,
		).Where(
			g.ID.Eq(s.gatewayInfo.ID),
		).Select(
			g.LastSyncedAt,
		).Updates(
			s.gatewayInfo,
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logging.Errorf("sync gateway:%s resource error: %s", s.gatewayInfo.Name, err.Error())
		return nil, err
	}
	logging.Infof("syncer[gateway:%s] end", s.gatewayInfo.Name)

	// 统计最新同步的资源类型及数量
	syncedResourceTypeStats := make(map[constant.APISIXResource]int)
	for _, resource := range resourceList {
		// 过滤掉已同步的资源
		if _, ok := databaseResourceMap[resource.GetResourceKey()]; !ok {
			syncedResourceTypeStats[resource.Type]++
		}
	}
	return syncedResourceTypeStats, nil
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
	// 使用标准化的资源类型 prefix（带 "/" 结尾，避免前缀匹配冲突）
	prefix := gatewayInfo.GetEtcdResourcePrefix(resourceType)
	kvList, err := s.etcdStore.List(ctx, prefix)
	if err != nil {
		return err
	}
	etcdResourceList, err := s.kvToResource(ctx, kvList)
	if err != nil {
		return err
	}
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
//
//nolint:gocyclo
func (s *UnifyOp) kvToResource(
	ctx context.Context,
	kvList []storage.KeyValuePair,
) ([]*model.GatewaySyncData, error) {
	resources, err := buildSyncSnapshotResources(ctx, s.gatewayInfo, kvList)
	if err != nil {
		logging.Errorf("build sync snapshot resources error: %s", err.Error())
		return nil, err
	}
	return resources, nil
}

// SyncedResourceToAPISIXResource 将同步的资源转换为 apisix 的资源
func SyncedResourceToAPISIXResource(
	resourceType constant.APISIXResource,
	syncedResources []*model.GatewaySyncData,
	status constant.ResourceStatus,
) any {
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
				Status:    status,
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

// GetResourceConfigDiffDetail 获取资源差异详情
func GetResourceConfigDiffDetail(
	ctx context.Context,
	resourceType constant.APISIXResource,
	id string,
) (*dto.ResourceDiffDetailResponse, error) {
	// 获取同步资源配置
	syncedResourceConfig := json.RawMessage("{}")
	syncedResource, err := GetSyncedItemByResourceTypeAndID(ctx, resourceType, id)
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
	// 使用标准化的 prefix（带 "/" 结尾，避免前缀匹配冲突）
	prefix := s.gatewayInfo.GetEtcdPrefixForList()
	kvList, err := s.etcdStore.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	logging.Infof("export [gateway:%s] end ", s.gatewayInfo.Name)
	return s.kvToResource(ctx, kvList)
}
