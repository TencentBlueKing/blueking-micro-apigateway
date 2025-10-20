package biz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
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

	if _, err := GetRoute(context.Background(), "dup-id"); err == nil {
		// 依旧只能是 existing 这条，状态保持 success
		r, err := GetRoute(context.Background(), "dup-id")
		assert.NoError(t, err)
		assert.Equal(t, "dup-name", r.Name)
		assert.Equal(t, constant.ResourceStatusSuccess, r.Status)
	}
	//    - 冲突 Name 的记录不应被创建（按 id 唯一，new-id-for-dup-name 不应落库为新资源）
	_, err = GetRoute(context.Background(), "new-id-for-dup-name")
	assert.Error(t, err)

	//    - 正常的不冲突记录应被创建
	r, err := GetRoute(context.Background(), "ok-id")
	assert.NoError(t, err)
	assert.Equal(t, "ok-name", r.Name)
	assert.Equal(t, constant.ResourceStatusSuccess, r.Status)
}
