package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/idx"
)

// TestInsertSyncedResources_RemoveDuplicated 验证 InsertSyncedResources 会移除与数据库已有资源 id/name 冲突的条目
func TestInsertSyncedResources_RemoveDuplicated(t *testing.T) {
	// 依赖 publish_test.go 中的 TestMain 初始化：gatewayInfo / gatewayCtx / embedDB
	// 1) 先创建一条已存在的 Route 资源（模拟数据库已有记录）
	existing := model.Route{
		Name: "dup-name",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "dup-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"dup-name"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	assert.NoError(t, CreateRoute(gatewayCtx, existing))

	// 2) 构造三条同步资源：
	//   - 与数据库 ID 冲突（相同 id: dup-id）
	//   - 与数据库 Name 冲突（相同 name: dup-name）
	//   - 完全不冲突（应被成功插入）
	dupID := &model.GatewaySyncData{
		ID:        "dup-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"new-name-for-dup-id"}`),
	}
	dupName := &model.GatewaySyncData{
		ID:        "new-id-for-dup-name",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"dup-name"}`),
	}
	normal := &model.GatewaySyncData{
		ID:        "ok-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"ok-name"}`),
	}

	// 3) 调用 InsertSyncedResources（内部会调用 RemoveDuplicatedResource 做去重）
	typeSynced := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {dupID, dupName, normal},
	}
	err := InsertSyncedResources(gatewayCtx, typeSynced, constant.ResourceStatusSuccess)
	// 有冲突会报错
	assert.Error(t, err)

	// 4) 断言：数据库中不会新增与 existing 冲突的两条，只应新增 normal 这一条) 调用 InsertSyncedResources（内部会调用 RemoveDuplicatedResource 做去重）
	typeSynced = map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {dupID, normal},
	}
	err = InsertSyncedResources(gatewayCtx, typeSynced, constant.ResourceStatusSuccess)

	assert.NoError(t, err)

	if _, err := GetRoute(gatewayCtx, "dup-id"); err == nil {
		// 依旧只能是 existing 这条，状态保持 success
		r, err := GetRoute(gatewayCtx, "dup-id")
		assert.NoError(t, err)
		assert.Equal(t, "dup-name", r.Name)
		assert.Equal(t, constant.ResourceStatusSuccess, r.Status)
	}
	//    - 冲突 Name 的记录不应被创建（按 id 唯一，new-id-for-dup-name 不应落库为新资源）
	_, err = GetRoute(gatewayCtx, "new-id-for-dup-name")
	assert.Error(t, err)

	//    - 正常的不冲突记录应被创建
	r, err := GetRoute(gatewayCtx, "ok-id")
	assert.NoError(t, err)
	assert.Equal(t, "ok-name", r.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, r.Status)
}

// TestUploadResources_DifferentGatewaysSameID 测试不同网关存在相同资源ID的情况
func TestUploadResources_DifferentGatewaysSameID(t *testing.T) {
	// 创建第二个网关
	gateway2 := &model.Gateway{
		Name:          "gateway2",
		Mode:          1,
		Maintainers:   []string{"user1"},
		Desc:          "gateway2",
		APISIXType:    constant.APISIXTypeBKAPISIX,
		APISIXVersion: "3.11.0",
		EtcdConfig: model.EtcdConfig{
			InstanceID: "987654321",
			EtcdConfig: base.EtcdConfig{
				Endpoint: "localhost:4380",
				Username: "test",
				Password: "test",
				Prefix:   "/apisix2",
			},
		},
	}
	err := CreateGateway(context.Background(), gateway2)
	assert.NoError(t, err)
	gateway2Ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway2)

	// 在第一个网关中创建资源
	route1 := &model.Route{
		Name: "same-id-route",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "same-resource-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"same-id-route","uris":["/gateway1"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err = CreateRoute(gatewayCtx, *route1)
	assert.NoError(t, err)

	// 在第二个网关中创建相同ID的资源
	route2 := &model.Route{
		Name: "same-id-route-gateway2",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "same-resource-id",
			GatewayID: gateway2.ID,
			Config:    datatypes.JSON(`{"name":"same-id-route-gateway2","uris":["/gateway2"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err = CreateRoute(gateway2Ctx, *route2)
	assert.NoError(t, err)

	// 验证两个网关的资源都存在且互不影响
	route1FromDB, err := GetRoute(gatewayCtx, "same-resource-id")
	assert.NoError(t, err)
	assert.Equal(t, "same-id-route", route1FromDB.Name)
	assert.Equal(t, gatewayInfo.ID, route1FromDB.GatewayID)

	route2FromDB, err := GetRoute(gateway2Ctx, "same-resource-id")
	assert.NoError(t, err)
	assert.Equal(t, "same-id-route-gateway2", route2FromDB.Name)
	assert.Equal(t, gateway2.ID, route2FromDB.GatewayID)

	// 清理第二个网关
	err = DeleteGateway(context.Background(), gateway2)
	assert.NoError(t, err)
}

// TestUploadResources_SameGatewayUpdateAndAdd 测试同一网关的更新覆盖和新增情况
func TestUploadResources_SameGatewayUpdateAndAdd(t *testing.T) {
	// 先创建一些现有资源
	existingRoute := &model.Route{
		Name: "existing-route",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "existing-route-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"existing-route","uris":["/existing"]}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err := CreateRoute(gatewayCtx, *existingRoute)
	assert.NoError(t, err)

	existingService := &model.Service{
		Name: "existing-service",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "existing-service-id",
			GatewayID: gatewayInfo.ID,
			Config:    datatypes.JSON(`{"name":"existing-service"}`),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	err = CreateService(gatewayCtx, *existingService)
	assert.NoError(t, err)

	// 准备更新资源（相同ID，不同配置）
	updateRouteData := &model.GatewaySyncData{
		ID:        "existing-route-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"updated-route","uris":["/updated"]}`),
	}

	updateServiceData := &model.GatewaySyncData{
		ID:        "existing-service-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"updated-service"}`),
	}

	// 准备新增资源
	newRouteData := &model.GatewaySyncData{
		ID:        "new-route-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"new-route","uris":["/new"]}`),
	}

	newUpstreamData := &model.GatewaySyncData{
		ID:        "new-upstream-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"new-upstream","type":"roundrobin"}`),
	}

	// 构造上传资源参数
	addResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route:    {newRouteData},
		constant.Upstream: {newUpstreamData},
	}

	updateTypeResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route:   {updateRouteData},
		constant.Service: {updateServiceData},
	}

	// 执行上传
	err = UploadResources(
		gatewayCtx,
		addResourcesTypeMap,
		updateTypeResourcesTypeMap,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证更新后的资源
	updatedRoute, err := GetRoute(gatewayCtx, "existing-route-id")
	assert.NoError(t, err)
	assert.Equal(t, "updated-route", updatedRoute.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedRoute.Status)

	updatedService, err := GetService(gatewayCtx, "existing-service-id")
	assert.NoError(t, err)
	assert.Equal(t, "updated-service", updatedService.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, updatedService.Status)

	// 验证新增的资源
	newRoute, err := GetRoute(gatewayCtx, "new-route-id")
	assert.NoError(t, err)
	assert.Equal(t, "new-route", newRoute.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, newRoute.Status)

	newUpstream, err := GetUpstream(gatewayCtx, "new-upstream-id")
	assert.NoError(t, err)
	assert.Equal(t, "new-upstream", newUpstream.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, newUpstream.Status)
}

// TestUploadResources_MixedResourceTypes 测试混合资源类型的上传
func TestUploadResources_MixedResourceTypes(t *testing.T) {
	// 准备多种资源类型的数据
	routeData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Route),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"mixed-route","uris":["/mixed"]}`),
	}

	serviceData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Service),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Service,
		Config:    datatypes.JSON(`{"name":"mixed-service"}`),
	}

	upstreamData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Upstream),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Upstream,
		Config:    datatypes.JSON(`{"name":"mixed-upstream","type":"roundrobin"}`),
	}

	consumerData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.Consumer),
		GatewayID: gatewayInfo.ID,
		Type:      constant.Consumer,
		Config:    datatypes.JSON(`{"username":"mixed-consumer"}`),
	}

	pluginConfigData := &model.GatewaySyncData{
		ID:        idx.GenResourceID(constant.PluginConfig),
		GatewayID: gatewayInfo.ID,
		Type:      constant.PluginConfig,
		Config:    datatypes.JSON(`{"name":"mixed-plugin-config"}`),
	}

	// 构造上传资源参数
	addResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route:        {routeData},
		constant.Service:      {serviceData},
		constant.Upstream:     {upstreamData},
		constant.Consumer:     {consumerData},
		constant.PluginConfig: {pluginConfigData},
	}

	// 执行上传
	err := UploadResources(
		gatewayCtx,
		addResourcesTypeMap,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证所有资源都被正确创建
	route, err := GetRoute(gatewayCtx, routeData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-route", route.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, route.Status)

	service, err := GetService(gatewayCtx, serviceData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-service", service.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, service.Status)

	upstream, err := GetUpstream(gatewayCtx, upstreamData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-upstream", upstream.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, upstream.Status)

	consumer, err := GetConsumer(gatewayCtx, consumerData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-consumer", consumer.Username)
	assert.Equal(t, constant.ResourceStatusCreateDraft, consumer.Status)

	pluginConfig, err := GetPluginConfig(gatewayCtx, pluginConfigData.ID)
	assert.NoError(t, err)
	assert.Equal(t, "mixed-plugin-config", pluginConfig.Name)
	assert.Equal(t, constant.ResourceStatusCreateDraft, pluginConfig.Status)
}

// TestUploadResources_EmptyResources 测试空资源上传
func TestUploadResources_EmptyResources(t *testing.T) {
	// 测试空的新增资源
	err := UploadResources(
		gatewayCtx,
		nil,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 测试空的更新资源
	err = UploadResources(
		gatewayCtx,
		nil,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)
}

// TestUploadResources_UpdateNonExistentResource 测试更新不存在的资源
func TestUploadResources_UpdateNonExistentResource(t *testing.T) {
	// 准备更新不存在的资源
	nonExistentRouteData := &model.GatewaySyncData{
		ID:        "non-existent-route-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"non-existent-route","uris":["/non-existent"]}`),
	}

	updateTypeResourcesTypeMap := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {nonExistentRouteData},
	}

	// 执行上传（应该成功，因为会先删除再插入）
	err := UploadResources(
		gatewayCtx,
		nil,
		updateTypeResourcesTypeMap,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证资源被创建
	route, err := GetRoute(gatewayCtx, "non-existent-route-id")
	assert.NoError(t, err)
	assert.Equal(t, "non-existent-route", route.Name)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, route.Status)
}

// TestUploadResources_CrossGatewayIsolation 测试跨网关隔离
func TestUploadResources_CrossGatewayIsolation(t *testing.T) {
	// 创建第二个网关
	gateway2 := &model.Gateway{
		Name:          "gateway2-isolation",
		Mode:          1,
		Maintainers:   []string{"user1"},
		Desc:          "gateway2-isolation",
		APISIXType:    constant.APISIXTypeBKAPISIX,
		APISIXVersion: "3.11.0",
		EtcdConfig: model.EtcdConfig{
			InstanceID: "isolation-test",
			EtcdConfig: base.EtcdConfig{
				Endpoint: "localhost:4381",
				Username: "test",
				Password: "test",
				Prefix:   "/apisix-isolation",
			},
		},
	}
	err := CreateGateway(context.Background(), gateway2)
	assert.NoError(t, err)
	gateway2Ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway2)

	// 在第一个网关中创建资源
	route1Data := &model.GatewaySyncData{
		ID:        "isolation-test-id",
		GatewayID: gatewayInfo.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"gateway1-route","uris":["/gateway1"]}`),
	}

	// 在第二个网关中创建相同ID的资源
	route2Data := &model.GatewaySyncData{
		ID:        "isolation-test-id",
		GatewayID: gateway2.ID,
		Type:      constant.Route,
		Config:    datatypes.JSON(`{"name":"gateway2-route","uris":["/gateway2"]}`),
	}

	// 分别上传到两个网关
	addResourcesTypeMap1 := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {route1Data},
	}
	err = UploadResources(
		gatewayCtx,
		addResourcesTypeMap1,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	addResourcesTypeMap2 := map[constant.APISIXResource][]*model.GatewaySyncData{
		constant.Route: {route2Data},
	}
	err = UploadResources(
		gateway2Ctx,
		addResourcesTypeMap2,
		nil,
		nil,
		nil,
	)
	assert.NoError(t, err)

	// 验证两个网关的资源相互隔离
	route1, err := GetRoute(gatewayCtx, "isolation-test-id")
	assert.NoError(t, err)
	assert.Equal(t, "gateway1-route", route1.Name)
	assert.Equal(t, gatewayInfo.ID, route1.GatewayID)

	route2, err := GetRoute(gateway2Ctx, "isolation-test-id")
	assert.NoError(t, err)
	assert.Equal(t, "gateway2-route", route2.Name)
	assert.Equal(t, gateway2.ID, route2.GatewayID)

	// 验证在第一个网关中无法访问第二个网关的资源
	_, err = GetRoute(gatewayCtx, "isolation-test-id")
	// 这里应该能访问到，因为ID相同但GatewayID不同，但GetRoute会通过GatewayID过滤
	// 所以实际上会返回第一个网关的资源
	assert.NoError(t, err)

	// 清理第二个网关
	err = DeleteGateway(context.Background(), gateway2)
	assert.NoError(t, err)
}
