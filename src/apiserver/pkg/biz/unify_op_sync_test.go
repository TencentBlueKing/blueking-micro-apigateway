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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/data"
)

// TestSyncWithPrefix_UpsertLogic tests the UPSERT logic in SyncWithPrefix
// This test verifies that the sync process correctly:
// 1. Creates new resources
// 2. Updates existing resources ONLY if ModRevision changed
// 3. Skips update if ModRevision is the same (optimization)
// 4. Deletes obsolete resources (in DB but not in etcd)
func TestSyncWithPrefix_UpsertLogic(t *testing.T) {
	// Setup: Create test gateway context
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)

	// Clean up gateway_sync_data before test
	u := repo.GatewaySyncData
	_, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Delete()
	assert.NoError(t, err)

	// Phase 1: Setup initial state in database
	// Create 4 resources in gateway_sync_data:
	// - resource1: will be updated (ModRevision changed)
	// - resource2: will be deleted (exists in DB but not in "etcd")
	// - resource3 (missing): will be created (exists in "etcd" but not in DB)
	// - resource4: will be skipped (ModRevision unchanged)

	resource1ID := idx.GenResourceID(constant.Route)
	resource2ID := idx.GenResourceID(constant.Route)
	resource3ID := idx.GenResourceID(constant.Route)
	resource4ID := idx.GenResourceID(constant.Route)

	initialResource1 := &model.GatewaySyncData{
		ID:          resource1ID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + resource1ID + `","name":"route-1-old","uris":["/old"]}`),
		ModRevision: 1,
	}
	initialResource2 := &model.GatewaySyncData{
		ID:        resource2ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"id":"` + resource2ID + `","name":"route-2-to-delete","uris":["/delete"]}`,
		),
		ModRevision: 1,
	}
	initialResource4 := &model.GatewaySyncData{
		ID:        resource4ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"id":"` + resource4ID + `","name":"route-4-unchanged","uris":["/unchanged"]}`,
		),
		ModRevision: 5,
	}

	err = repo.Q.GatewaySyncData.WithContext(ctx).Create(initialResource1)
	assert.NoError(t, err)
	err = repo.Q.GatewaySyncData.WithContext(ctx).Create(initialResource2)
	assert.NoError(t, err)
	err = repo.Q.GatewaySyncData.WithContext(ctx).Create(initialResource4)
	assert.NoError(t, err)

	// Phase 2: Create UnifyOp and prepare "etcd" resources
	// Simulate etcd having: resource1 (updated), resource3 (new), resource4 (unchanged)
	// resource2 is missing from etcd (should be deleted)

	updatedResource1 := &model.GatewaySyncData{
		ID:        resource1ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"id":"` + resource1ID + `","name":"route-1-updated","uris":["/updated"]}`,
		),
		ModRevision: 2, // ModRevision changed
	}
	newResource3 := &model.GatewaySyncData{
		ID:          resource3ID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + resource3ID + `","name":"route-3-new","uris":["/new"]}`),
		ModRevision: 1,
	}
	unchangedResource4 := &model.GatewaySyncData{
		ID:        resource4ID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"id":"` + resource4ID + `","name":"route-4-unchanged","uris":["/unchanged"]}`,
		),
		ModRevision: 5, // Same ModRevision - should NOT be updated
	}

	// Prepare resourceList (simulating what kvToResource would return)
	resourceList := []*model.GatewaySyncData{
		updatedResource1,
		newResource3,
		unchangedResource4,
	}

	// Phase 3: Execute the UPSERT logic (extracted from SyncWithPrefix)
	var actualUpdates int
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		// Build map of etcd resources for quick lookup
		etcdResourceMap := make(map[string]*model.GatewaySyncData)
		for _, resource := range resourceList {
			key := resource.Type.String() + ":" + resource.ID
			etcdResourceMap[key] = resource
		}

		// Phase 1: Get existing resources from database
		existingResources, err := tx.GatewaySyncData.WithContext(ctx).
			Where(u.GatewayID.Eq(gatewayInfo.ID)).
			Find()
		if err != nil {
			return err
		}

		// Build map of existing resources
		existingResourceMap := make(map[string]*model.GatewaySyncData)
		var resourcesToDelete []int // auto_id list
		for _, existing := range existingResources {
			key := existing.Type.String() + ":" + existing.ID
			existingResourceMap[key] = existing

			// If resource exists in DB but not in etcd, mark for deletion
			if _, existsInEtcd := etcdResourceMap[key]; !existsInEtcd {
				resourcesToDelete = append(resourcesToDelete, existing.AutoID)
			}
		}

		// Phase 2: UPSERT resources from etcd
		var resourcesToCreate []*model.GatewaySyncData
		var resourcesToUpdate []*model.GatewaySyncData

		for _, resource := range resourceList {
			key := resource.Type.String() + ":" + resource.ID
			if existing, exists := existingResourceMap[key]; exists {
				// Only update if ModRevision changed
				if existing.ModRevision != resource.ModRevision {
					existing.Config = resource.Config
					existing.ModRevision = resource.ModRevision
					resourcesToUpdate = append(resourcesToUpdate, existing)
				}
			} else {
				// Create new record
				resourcesToCreate = append(resourcesToCreate, resource)
			}
		}

		actualUpdates = len(resourcesToUpdate)

		// Execute updates
		for _, resource := range resourcesToUpdate {
			_, err := tx.GatewaySyncData.WithContext(ctx).
				Where(u.AutoID.Eq(resource.AutoID)).
				Updates(map[string]any{
					"config":       resource.Config,
					"mod_revision": resource.ModRevision,
				})
			if err != nil {
				return err
			}
		}

		// Execute creates
		if len(resourcesToCreate) > 0 {
			err = tx.GatewaySyncData.WithContext(ctx).CreateInBatches(resourcesToCreate, 500)
			if err != nil {
				return err
			}
		}

		// Phase 3: Delete obsolete resources
		if len(resourcesToDelete) > 0 {
			_, err = tx.GatewaySyncData.WithContext(ctx).
				Where(u.AutoID.In(resourcesToDelete...)).
				Delete()
			if err != nil {
				return err
			}
		}

		return nil
	})
	assert.NoError(t, err)

	// Phase 4: Verify results

	// 4.1: Verify resource1 was updated
	updated1, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID), u.ID.Eq(resource1ID)).
		Take()
	assert.NoError(t, err)
	assert.Equal(t, resource1ID, updated1.ID)
	assert.Equal(t, 2, updated1.ModRevision)
	var config1 map[string]any
	err = json.Unmarshal(updated1.Config, &config1)
	assert.NoError(t, err)
	assert.Equal(t, "route-1-updated", config1["name"])

	// 4.2: Verify resource2 was deleted
	_, err = repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID), u.ID.Eq(resource2ID)).
		Take()
	assert.Error(t, err) // Should not exist

	// 4.3: Verify resource3 was created
	created3, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID), u.ID.Eq(resource3ID)).
		Take()
	assert.NoError(t, err)
	assert.Equal(t, resource3ID, created3.ID)
	assert.Equal(t, 1, created3.ModRevision)
	var config3 map[string]any
	err = json.Unmarshal(created3.Config, &config3)
	assert.NoError(t, err)
	assert.Equal(t, "route-3-new", config3["name"])

	// 4.4: Verify resource4 was NOT updated (ModRevision unchanged)
	unchanged4, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID), u.ID.Eq(resource4ID)).
		Take()
	assert.NoError(t, err)
	assert.Equal(t, resource4ID, unchanged4.ID)
	assert.Equal(t, 5, unchanged4.ModRevision)
	var config4 map[string]any
	err = json.Unmarshal(unchanged4.Config, &config4)
	assert.NoError(t, err)
	assert.Equal(t, "route-4-unchanged", config4["name"])

	// 4.5: Verify total count
	allResources, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID)).
		Find()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(allResources)) // resource1, resource3, resource4

	// 4.6: Verify optimization - only 1 update should have been executed (resource1)
	// resource4 should have been skipped
	assert.Equal(t, 1, actualUpdates, "Should only update resources with changed ModRevision")
}

// TestSyncWithPrefix_NoRaceCondition tests that the UPSERT logic eliminates the race condition
// This test verifies that resources are never completely absent during sync
func TestSyncWithPrefix_NoRaceCondition(t *testing.T) {
	// Setup: Create test gateway context
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)

	// Clean up gateway_sync_data before test
	u := repo.GatewaySyncData
	_, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Delete()
	assert.NoError(t, err)

	// Create initial resource
	resourceID := idx.GenResourceID(constant.Route)
	initialResource := &model.GatewaySyncData{
		ID:          resourceID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + resourceID + `","name":"route-1","uris":["/test"]}`),
		ModRevision: 1,
	}
	err = repo.Q.GatewaySyncData.WithContext(ctx).Create(initialResource)
	assert.NoError(t, err)

	// Verify resource exists before sync
	resource, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID), u.ID.Eq(resourceID)).
		Take()
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	// Simulate sync with updated resource
	updatedResource := &model.GatewaySyncData{
		ID:        resourceID,
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config: datatypes.JSON(
			`{"id":"` + resourceID + `","name":"route-1-updated","uris":["/test-updated"]}`,
		),
		ModRevision: 2,
	}
	resourceList := []*model.GatewaySyncData{updatedResource}

	// Execute the UPSERT logic
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		etcdResourceMap := make(map[string]*model.GatewaySyncData)
		for _, resource := range resourceList {
			key := resource.Type.String() + ":" + resource.ID
			etcdResourceMap[key] = resource
		}

		existingResources, err := tx.GatewaySyncData.WithContext(ctx).
			Where(u.GatewayID.Eq(gatewayInfo.ID)).
			Find()
		if err != nil {
			return err
		}

		existingResourceMap := make(map[string]*model.GatewaySyncData)
		var resourcesToDelete []int
		for _, existing := range existingResources {
			key := existing.Type.String() + ":" + existing.ID
			existingResourceMap[key] = existing

			if _, existsInEtcd := etcdResourceMap[key]; !existsInEtcd {
				resourcesToDelete = append(resourcesToDelete, existing.AutoID)
			}
		}

		var resourcesToUpdate []*model.GatewaySyncData
		for _, resource := range resourceList {
			key := resource.Type.String() + ":" + resource.ID
			if existing, exists := existingResourceMap[key]; exists {
				existing.Config = resource.Config
				existing.ModRevision = resource.ModRevision
				resourcesToUpdate = append(resourcesToUpdate, existing)
			}
		}

		// The key test: resource should still exist at this point
		// (not deleted before update)
		existsDuringSync, err := tx.GatewaySyncData.WithContext(ctx).
			Where(u.GatewayID.Eq(gatewayInfo.ID), u.ID.Eq(resourceID)).
			Take()
		assert.NoError(t, err)
		assert.NotNil(t, existsDuringSync)

		for _, resource := range resourcesToUpdate {
			_, err := tx.GatewaySyncData.WithContext(ctx).
				Where(u.AutoID.Eq(resource.AutoID)).
				Updates(map[string]any{
					"config":       resource.Config,
					"mod_revision": resource.ModRevision,
				})
			if err != nil {
				return err
			}
		}

		return nil
	})
	assert.NoError(t, err)

	// Verify resource still exists after sync with updated values
	finalResource, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID), u.ID.Eq(resourceID)).
		Take()
	assert.NoError(t, err)
	assert.Equal(t, resourceID, finalResource.ID)
	assert.Equal(t, 2, finalResource.ModRevision)
	var config map[string]any
	err = json.Unmarshal(finalResource.Config, &config)
	assert.NoError(t, err)
	assert.Equal(t, "route-1-updated", config["name"])
}

// TestSyncWithPrefix_BatchProcessing tests that batch processing works correctly
func TestSyncWithPrefix_BatchProcessing(t *testing.T) {
	// Setup: Create test gateway context
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)

	// Clean up gateway_sync_data before test
	u := repo.GatewaySyncData
	_, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Delete()
	assert.NoError(t, err)

	// Create multiple resources to test batch processing
	const resourceCount = 10
	var resourceList []*model.GatewaySyncData

	for i := 0; i < resourceCount; i++ {
		resourceID := idx.GenResourceID(constant.Route)
		resource := &model.GatewaySyncData{
			ID:        resourceID,
			GatewayID: gatewayInfo.ID,
			Type:      constant.Route,
			Config: datatypes.JSON(
				`{"id":"` + resourceID + `","name":"route-` + string(
					rune(i),
				) + `","uris":["/test-` + string(
					rune(i),
				) + `"]}`,
			),
			ModRevision: 1,
		}
		resourceList = append(resourceList, resource)
	}

	// Execute batch create
	err = repo.Q.Transaction(func(tx *repo.Query) error {
		return tx.GatewaySyncData.WithContext(ctx).CreateInBatches(resourceList, 500)
	})
	assert.NoError(t, err)

	// Verify all resources were created
	allResources, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID)).
		Find()
	assert.NoError(t, err)
	assert.Equal(t, resourceCount, len(allResources))
}

func TestSyncWithPrefix_SnapshotConfigShaping_CurrentSeam(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	u := repo.GatewaySyncData
	_, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Delete()
	assert.NoError(t, err)

	routeID := idx.GenResourceID(constant.Route)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	existingMetadata := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusSuccess)
	existingMetadata.Name = "limit-count-" + suffix
	assert.NoError(t, CreatePluginMetadata(ctx, *existingMetadata))

	existingGroup := data.ConsumerGroup1WithNoRelation(gatewayInfo, constant.ResourceStatusSuccess)
	existingGroup.Name = "cg-from-db-" + suffix
	assert.NoError(t, CreateConsumerGroup(ctx, *existingGroup))

	existingStreamRoute := data.StreamRoute1WithNoRelationResource(
		gatewayInfo, constant.ResourceStatusSuccess,
	)
	existingStreamRoute.Name = "sr-from-db-" + suffix
	existingStreamRoute.Config, _ = sjson.SetBytes(
		existingStreamRoute.Config, "labels", map[string]string{"env": "test"},
	)
	assert.NoError(t, CreateStreamRoute(ctx, *existingStreamRoute))

	prefix := gatewayInfo.GetEtcdPrefixForList()
	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			data: map[string]string{
				prefix + "routes/" + routeID:                        `{"uri":"/from-etcd","create_time":111,"update_time":222}`,
				prefix + "plugin_metadata/" + existingMetadata.Name: `{"value":{"disable":false}}`,
				prefix + "consumer_groups/" + existingGroup.ID:      `{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
				prefix + "stream_routes/" + existingStreamRoute.ID:  `{"server_addr":"127.0.0.1","server_port":8080}`,
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	_, err = syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)

	routeSnapshot, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Route, routeID)
	assert.NoError(t, err)
	assert.Equal(t, "routes_"+routeID, gjson.GetBytes(routeSnapshot.Config, "name").String())
	assert.False(t, gjson.GetBytes(routeSnapshot.Config, "create_time").Exists())
	assert.False(t, gjson.GetBytes(routeSnapshot.Config, "update_time").Exists())

	metadataSnapshot, err := GetSyncedItemByResourceTypeAndID(
		ctx, constant.PluginMetadata, existingMetadata.ID,
	)
	assert.NoError(t, err)
	assert.Equal(t, existingMetadata.Name, metadataSnapshot.GetName())

	groupSnapshot, err := GetSyncedItemByResourceTypeAndID(ctx, constant.ConsumerGroup, existingGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, existingGroup.ID, gjson.GetBytes(groupSnapshot.Config, "id").String())
	assert.Equal(t, existingGroup.Name, gjson.GetBytes(groupSnapshot.Config, "name").String())

	streamRouteSnapshot, err := GetSyncedItemByResourceTypeAndID(
		ctx, constant.StreamRoute, existingStreamRoute.ID,
	)
	assert.NoError(t, err)
	assert.Equal(t, existingStreamRoute.Name, gjson.GetBytes(streamRouteSnapshot.Config, "name").String())
	assert.Equal(t, "test", gjson.GetBytes(streamRouteSnapshot.Config, "labels.env").String())
}

func TestSyncWithPrefix_ReturnsOnlyNewSnapshotCounts(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	prefix := gatewayInfo.GetEtcdPrefixForList()
	u := repo.GatewaySyncData
	_, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Delete()
	assert.NoError(t, err)

	existingID := idx.GenResourceID(constant.Route)
	newID := idx.GenResourceID(constant.Route)

	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:          existingID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + existingID + `","name":"existing-route","uri":"/existing"}`),
		ModRevision: 1,
	}))

	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			data: map[string]string{
				prefix + "routes/" + existingID: `{"name":"existing-route-updated","uri":"/updated"}`,
				prefix + "routes/" + newID:      `{"name":"new-route","uri":"/new"}`,
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	counts, err := syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)
	assert.Equal(t, 1, counts[constant.Route])
}

func TestSyncWithPrefix_ReturnsErrorOnHelperLookupFailure(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	prefix := gatewayInfo.GetEtcdPrefixForList()
	u := repo.GatewaySyncData
	_, err := repo.Q.GatewaySyncData.WithContext(ctx).Where(u.GatewayID.Eq(gatewayInfo.ID)).Delete()
	assert.NoError(t, err)

	existingID := idx.GenResourceID(constant.Route)
	assert.NoError(t, repo.Q.GatewaySyncData.WithContext(ctx).Create(&model.GatewaySyncData{
		ID:          existingID,
		GatewayID:   gatewayInfo.ID,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"id":"` + existingID + `","name":"existing-route","uri":"/existing"}`),
		ModRevision: 1,
	}))

	db := database.Client()
	assert.NoError(t, db.Migrator().DropTable(&model.PluginMetadata{}))
	t.Cleanup(func() {
		restoreErr := db.Exec(
			"CREATE TABLE `plugin_metadata` (`name` varchar(255),`creator` varchar(32),`updater` varchar(32),`created_at` datetime,`updated_at` datetime,`auto_id` integer PRIMARY KEY AUTOINCREMENT,`id` varchar(255),`gateway_id` integer,`config` JSON,`status` varchar(32))",
		).Error
		if restoreErr != nil {
			t.Fatalf("restore plugin_metadata table: %v", restoreErr)
		}
	})

	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			data: map[string]string{
				prefix + "plugin_metadata/failing-plugin": `{"value":{"disable":false}}`,
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	_, err = syncer.SyncWithPrefix(ctx, prefix)
	assert.Error(t, err)

	storedSnapshot, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Route, existingID)
	assert.NoError(t, err)
	assert.Equal(t, existingID, storedSnapshot.ID)
	assert.Equal(t, 1, storedSnapshot.ModRevision)

	allSnapshots, err := repo.Q.GatewaySyncData.WithContext(ctx).
		Where(u.GatewayID.Eq(gatewayInfo.ID)).
		Find()
	assert.NoError(t, err)
	assert.Len(t, allSnapshots, 1)
}

// mockEtcdStore is a mock implementation of storage.StorageInterface for testing
type mockEtcdStore struct {
	storage.StorageInterface
	data map[string]string
}

func (m *mockEtcdStore) List(ctx context.Context, prefix string) ([]storage.KeyValuePair, error) {
	var kvList []storage.KeyValuePair
	for key, value := range m.data {
		kvList = append(kvList, storage.KeyValuePair{
			Key:         key,
			Value:       value,
			ModRevision: 1,
		})
	}
	return kvList, nil
}

func (m *mockEtcdStore) Get(ctx context.Context, key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", nil
}
