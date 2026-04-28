# Sync Data 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变 `gateway_sync_data` 快照语义、不触碰 `HandleConfig()`、不改 import/publish/web/open/mcp 行为的前提下，把 etcd -> 数据库快照同步链路里当前混在 `SyncWithPrefix(...)` 和 `kvToResource(...)` 里的 config 规范化、DB 反查回填、plugin metadata ID 对齐、以及同步 diff 规划几个步骤拆开，使 sync-data 这条链路的本地复杂度降下来。

**Architecture:** 本计划完全承认 sync-data 是一条独立的 read-side / snapshot 链路，不把它硬并进 import 或 publish 的 helper。执行顺序调整为：先独立补 `SyncWithPrefix(...)` 的 characterization tests；然后优先拆 `kvToResource(...)` 里的基础 KV 规范化、DB 持久化字段回填、plugin metadata ID 对齐；在这些 config 处理 helper 稳定后，再把 `kvToResource(...)` 收敛成薄 orchestration。只有在前 4 步做完后，`SyncWithPrefix(...)` 的 diff/apply 尾巴仍然显著影响可读性时，才继续抽 change-set helper。

**Tech Stack:** Go, GORM, etcd storage interface, Gin context helper, `gjson` / `sjson`, `testify`, `go test`, `make lint`, `make test`

---

## 代码复核结论

- 重构目的判断：正确。当前 etcd -> `gateway_sync_data` 的复杂度核心不在 UPSERT 语句本身，而在 `kvToResource(...)` 同时承担 key 解析、时间字段清理、缺省 name 注入、DB 反查回填、plugin metadata ID 协调这几类 config 处理。
- 复杂度评估：整体中到偏高。因为现有 `unify_op_sync_test.go` 主要锁的是“抽出来的 UPSERT 逻辑”和竞态回归，没有真正锁住 `SyncWithPrefix(...)` 产生的快照 config 形态。
- 本次修正：把同步快照的 seam-first 测试提成独立前置阶段；优先处理 snapshot config shaping，再决定是否继续收拢 `SyncWithPrefix(...)` 的 diff/apply 尾巴。

## 执行顺序（修订）

1. Task 0：独立补 sync snapshot characterization tests。
2. Task 1：先抽 KV -> `GatewaySyncData` 的基础规范化 helper。
3. Task 2：再抽“按数据库已有资源回填 snapshot config”的 helper。
4. Task 3：单独收拢 plugin metadata 的 snapshot ID 对齐 helper。
5. Task 4：把 `kvToResource(...)` 改成 sync-data 本地 orchestration。
6. Task 5：仅当 Task 1-4 做完后，`SyncWithPrefix(...)` 的 diff/apply 仍然明显拖累可读性时，再抽 change-set planner helper。

## 范围

- 只处理 `src/apiserver/pkg/biz/unify_op.go`
- 允许新增 sync-data 域内 helper 文件
- 允许补齐 `src/apiserver/pkg/biz/unify_op_sync_test.go`
- 允许新增 `src/apiserver/pkg/biz/unify_op_sync_helpers_test.go`

## 非目标

- 不改 `gateway_sync_data` 表结构
- 不改 `AddSyncedResources(...)`、`UploadResources(...)`、`DiffResources(...)`、`BatchRevertXxx(...)` 的行为
- 不改 `pkg/entity/model/*.go` 中各资源 `HandleConfig()` 的行为
- 不改 prefix 规范化规则、leader election、定时同步节奏
- 不抽跨 `sync/import/publish/web/open/mcp` 的共享 helper
- 不把 read-side snapshot 和 publish-side ETCD payload 统一成同一套 builder

## 当前测试缺口

- `src/apiserver/pkg/biz/unify_op_sync_test.go` 当前主要覆盖：
  - 从 `SyncWithPrefix(...)` 中抽出来的 UPSERT 逻辑
  - “更新时不先删再插”的竞态保护
  - 批量创建的基本行为
- 但它没有直接锁住下面这些真实同步现状：
  - route / upstream / service 等资源在 snapshot 中会删除 `create_time` / `update_time`
  - 缺省 `name` 的资源在 snapshot 中会被补成 `<resource_prefix>_<id>`
  - plugin metadata 会按 etcd key 名反查数据库已有记录，并尽量复用已有 DB `id`
  - global_rule / plugin_config / consumer_group / proto / stream_route 会通过数据库现有记录回填 snapshot config 中的 `name`、`id`、`labels`
  - `SyncWithPrefix(...)` 返回的统计值只计算新建 snapshot，而不是 update / unchanged

## 文件结构

- `src/apiserver/pkg/biz/unify_op.go`
  - 当前 sync snapshot 主链路：`SyncWithPrefix(...)`、`kvToResource(...)`
- `src/apiserver/pkg/biz/unify_op_sync_test.go`
  - 当前 sync seam 测试；后续继续作为真实同步入口的 characterization test 主入口
- `src/apiserver/pkg/biz/unify_op_sync_helpers.go`
  - sync-data 域内 helper：KV 规范化、snapshot field backfill、plugin metadata ID 协调、change-set planner
- `src/apiserver/pkg/biz/unify_op_sync_helpers_test.go`
  - sync-data helper 第二层单测

## PR 出口要求

- 每个任务里的 `go test` 是最小验收命令
- 每个任务准备合并前，再补跑一次：

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && make lint && make test
```

## 测试策略（必须）

- 新增 `Task 0` 作为独立步骤或独立 PR；在 `Task 0` 合并前，不开始 Task 1-5。
- 每个任务的第一组测试，必须先打在“重构前已经存在的 seam”上，不能直接从计划中新引入的 helper 开始写测试。
- helper 测试只能作为第二层测试：
  - 第一层：优先锁 `SyncWithPrefix(...)` 这条真实同步链路
  - 第二层：helper 抽出后再补 `unify_op_sync_helpers_test.go`
- sync-data 计划里的现有 seam 优先级如下：
  - Task 0：优先测现有 `SyncWithPrefix(...)`，覆盖 snapshot config shaping 和统计值
  - Task 1-4：在 Task 0 已锁住实际同步行为后，再在同包测试里补 helper 单测，把重构面缩小到本地 config processing
  - Task 5：继续以 `SyncWithPrefix(...)` 为主，补 diff/apply 规划边界断言
- `ExportEtcdResources(...)` 可以作为辅助 seam，用来锁“只读导出也走同一份 snapshot shaping”，但不能代替 `SyncWithPrefix(...)`。
- 执行时，如果任务正文里的示例代码先写了 helper 测试，应按上面的 seam 规则落地：先补现有 seam 的 characterization test，再补 helper test。

## 重构前测试前置阶段（独立）

- Task 0 的目标不是引入 helper，而是先把同步快照的现状锁住；建议直接扩展 `unify_op_sync_test.go`。
- Task 0 至少覆盖 4 类现状：
  - `create_time` / `update_time` 会从 snapshot config 被删除
  - 缺省 `name` 会被补成 `<resource_prefix>_<id>`
  - plugin metadata 会优先复用数据库已有 `id`
  - consumer_group / stream_route 这类资源会通过数据库现有记录回填 `name` / `id` / `labels`
- Task 0 完成后，Task 1-4 才允许把断言下沉到 helper；否则后续很难判断是 helper 行为变了，还是同步入口行为本来就没锁住。

### Task 0: 补 sync snapshot characterization tests

- [ ] Task 0: 补 sync snapshot characterization tests

**要解决的缺口：** 当前测试已经锁住了“怎么 upsert”，但还没有锁住“最终 snapshot config 长什么样”。先把 `SyncWithPrefix(...)` 的真实输出锁住，后面的 helper 提取才有黑盒护栏。

**为什么这个任务适合单独提 PR：** 只扩现有 `unify_op_sync_test.go`，不改同步生产逻辑。

**Files:**
- Modify: `src/apiserver/pkg/biz/unify_op_sync_test.go`

- [ ] **Step 1: 在真实 `SyncWithPrefix(...)` seam 上补 characterization tests**

在 `unify_op_sync_test.go` 增加下面这组测试，直接通过真实同步路径锁定 snapshot config shaping：

```go
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

	existingStreamRoute := data.StreamRoute1WithNoRelationResource(gatewayInfo, constant.ResourceStatusSuccess)
	existingStreamRoute.Name = "sr-from-db-" + suffix
	existingStreamRoute.Config, _ = sjson.SetBytes(existingStreamRoute.Config, "labels", map[string]string{"env": "test"})
	assert.NoError(t, CreateStreamRoute(ctx, *existingStreamRoute))

	prefix := gatewayInfo.GetEtcdPrefixForList()
	syncer := &UnifyOp{
		etcdStore: &mockEtcdStore{
			data: map[string]string{
				prefix + "routes/" + routeID: `{"uri":"/from-etcd","create_time":111,"update_time":222}`,
				prefix + "plugin_metadata/" + existingMetadata.Name: `{"value":{"disable":false}}`,
				prefix + "consumer_groups/" + existingGroup.ID: `{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
				prefix + "stream_routes/" + existingStreamRoute.ID: `{"server_addr":"127.0.0.1","server_port":8080}`,
			},
		},
		gatewayInfo: gatewayInfo,
		isLeader:    true,
	}

	_, err := syncer.SyncWithPrefix(ctx, prefix)
	assert.NoError(t, err)

	routeSnapshot, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Route, routeID)
	assert.NoError(t, err)
	assert.Equal(t, "routes_"+routeID, gjson.GetBytes(routeSnapshot.Config, "name").String())
	assert.False(t, gjson.GetBytes(routeSnapshot.Config, "create_time").Exists())
	assert.False(t, gjson.GetBytes(routeSnapshot.Config, "update_time").Exists())

	metadataSnapshot, err := GetSyncedItemByResourceTypeAndID(ctx, constant.PluginMetadata, existingMetadata.ID)
	assert.NoError(t, err)
	assert.Equal(t, existingMetadata.Name, metadataSnapshot.GetName())

	groupSnapshot, err := GetSyncedItemByResourceTypeAndID(ctx, constant.ConsumerGroup, existingGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, existingGroup.ID, gjson.GetBytes(groupSnapshot.Config, "id").String())
	assert.Equal(t, existingGroup.Name, gjson.GetBytes(groupSnapshot.Config, "name").String())

	streamRouteSnapshot, err := GetSyncedItemByResourceTypeAndID(ctx, constant.StreamRoute, existingStreamRoute.ID)
	assert.NoError(t, err)
	assert.Equal(t, existingStreamRoute.Name, gjson.GetBytes(streamRouteSnapshot.Config, "name").String())
	assert.Equal(t, "test", gjson.GetBytes(streamRouteSnapshot.Config, "labels.env").String())
}
```

同文件再补一组统计值 characterization test，锁住“只统计新建 snapshot”这一现状：

```go
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
```

- [ ] **Step 2: 运行 sync seam tests，确认当前 snapshot 行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestSyncWithPrefix_(SnapshotConfigShaping_CurrentSeam|ReturnsOnlyNewSnapshotCounts)' -count=1
```

Expected:
- PASS

- [ ] **Step 3: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/unify_op_sync_test.go
git commit -m "test: lock sync snapshot characterization seams"
```

---

### Task 1: 抽 KV -> `GatewaySyncData` 的基础规范化 helper

- [ ] Task 1: 抽 KV -> `GatewaySyncData` 的基础规范化 helper

**要解决的复杂度：** `kvToResource(...)` 现在在主循环里同时做 key 解析、资源类型识别、快照对象创建、删除 `create_time/update_time`、以及缺省 `name` 注入。后面再加任何一个 snapshot 字段规则，都要先通读这一长段分支。

**为什么这个任务适合单独提 PR：** 只抽“从单个 etcd KV 构出基础 snapshot 资源”这一步，不碰数据库反查回填，也不碰 `SyncWithPrefix(...)` 的 diff/apply。

**Files:**
- Create: `src/apiserver/pkg/biz/unify_op_sync_helpers.go`
- Create: `src/apiserver/pkg/biz/unify_op_sync_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/unify_op.go:708-739`

- [ ] **Step 1: 先补基础规范化 helper 的失败测试**

在 `unify_op_sync_helpers_test.go` 增加：

```go
func TestBuildSyncedResourceFromKV(t *testing.T) {
	t.Parallel()

	normalizedPrefix := model.NormalizeEtcdPrefix("/apisix")

	t.Run("route strips timestamps and injects fallback name", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/routes/route-id",
			Value:       `{"uri":"/demo","create_time":111,"update_time":222}`,
			ModRevision: 9,
		})
		assert.True(t, ok)
		assert.Equal(t, "route-id", got.ID)
		assert.Equal(t, 17, got.GatewayID)
		assert.Equal(t, constant.Route, got.Type)
		assert.Equal(t, 9, got.ModRevision)
		assert.Equal(t, "routes_route-id", gjson.GetBytes(got.Config, "name").String())
		assert.False(t, gjson.GetBytes(got.Config, "create_time").Exists())
		assert.False(t, gjson.GetBytes(got.Config, "update_time").Exists())
	})

	t.Run("plugin metadata uses etcd key as snapshot name", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/plugin_metadata/clickhouse-logger",
			Value:       `{"value":{"disable":false}}`,
			ModRevision: 3,
		})
		assert.True(t, ok)
		assert.Equal(t, constant.PluginMetadata, got.Type)
		assert.Equal(t, "clickhouse-logger", got.ID)
		assert.Equal(t, "clickhouse-logger", got.GetName())
	})

	t.Run("invalid key is ignored", func(t *testing.T) {
		got, ok := buildSyncedResourceFromKV(normalizedPrefix, 17, storage.KeyValuePair{
			Key:         "/apisix/routes/too/many/parts",
			Value:       `{}`,
			ModRevision: 1,
		})
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestBuildSyncedResourceFromKV -count=1
```

Expected:
- FAIL，报 `undefined: buildSyncedResourceFromKV`

- [ ] **Step 3: 实现 helper，并替换 `kvToResource(...)` 的内联规范化逻辑**

在 `unify_op_sync_helpers.go` 里新增：

```go
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
```

然后把 `kvToResource(...)` 里原来这段：

```go
resourceKeyWithoutPrefix := strings.TrimPrefix(kv.Key, normalizedPrefix)
resourceKeyList := strings.Split(resourceKeyWithoutPrefix, "/")
...
resourceInfo := &model.GatewaySyncData{...}
...
if resourceType != constant.PluginMetadata && resourceInfo.GetName() == "" {
	resourceInfo.SetName(fmt.Sprintf("%s_%s", resourceTypeValue, id))
} else if resourceType == constant.PluginMetadata {
	resourceInfo.SetName(id)
}
```

替换成：

```go
resourceInfo, ok := buildSyncedResourceFromKV(normalizedPrefix, s.gatewayInfo.ID, kv)
if !ok {
	logging.Errorf("key is not validate: %s", kv.Key)
	continue
}
resourceType := resourceInfo.Type
id := resourceInfo.ID
resources = append(resources, resourceInfo)
```

- [ ] **Step 4: 运行 biz 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestBuildSyncedResourceFromKV|TestSyncWithPrefix_(SnapshotConfigShaping_CurrentSeam|ReturnsOnlyNewSnapshotCounts)' -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/unify_op.go src/apiserver/pkg/biz/unify_op_sync_helpers.go src/apiserver/pkg/biz/unify_op_sync_helpers_test.go
git commit -m "refactor: extract base sync snapshot kv normalizer"
```

---

### Task 2: 抽“按数据库已有资源回填 snapshot config”的 helper

- [ ] Task 2: 抽“按数据库已有资源回填 snapshot config”的 helper

**要解决的复杂度：** `kvToResource(...)` 里现在有 5 组“先收集 ID，再查数据库，再把 `name/id/labels` 写回 snapshot config”的特殊处理：`global_rule`、`plugin_config`、`consumer_group`、`proto`、`stream_route`。这些回填规则都是真实业务边界，但不该继续散在主函数里。

**为什么这个任务适合单独提 PR：** 只处理数据库已有资源对 snapshot config 的回填，不碰 plugin metadata ID 规则，也不改 `SyncWithPrefix(...)`。

**Files:**
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers.go`
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/unify_op.go:694-818`

- [ ] **Step 1: 先补 DB 回填 helper 的失败测试**

在 `unify_op_sync_helpers_test.go` 增加：

```go
func TestBackfillStoredSnapshotFields(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	pluginConfig := data.PluginConfig1WithNoRelation(gatewayInfo, constant.ResourceStatusSuccess)
	pluginConfig.Name = "pc-from-db-" + suffix
	assert.NoError(t, CreatePluginConfig(ctx, *pluginConfig))

	consumerGroup := data.ConsumerGroup1WithNoRelation(gatewayInfo, constant.ResourceStatusSuccess)
	consumerGroup.Name = "cg-from-db-" + suffix
	assert.NoError(t, CreateConsumerGroup(ctx, *consumerGroup))

	proto := data.Proto1(gatewayInfo, constant.ResourceStatusSuccess)
	proto.Name = "proto-from-db-" + suffix
	assert.NoError(t, CreateProto(ctx, *proto))

	streamRoute := data.StreamRoute1WithNoRelationResource(gatewayInfo, constant.ResourceStatusSuccess)
	streamRoute.Name = "sr-from-db-" + suffix
	streamRoute.Config, _ = sjson.SetBytes(streamRoute.Config, "labels", map[string]string{"env": "test"})
	assert.NoError(t, CreateStreamRoute(ctx, *streamRoute))

	resources := []*model.GatewaySyncData{
		{ID: pluginConfig.ID, GatewayID: gatewayInfo.ID, Type: constant.PluginConfig, Config: datatypes.JSON(`{"plugins":{}}`)},
		{ID: consumerGroup.ID, GatewayID: gatewayInfo.ID, Type: constant.ConsumerGroup, Config: datatypes.JSON(`{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`)},
		{ID: proto.ID, GatewayID: gatewayInfo.ID, Type: constant.Proto, Config: datatypes.JSON(`{"content":"syntax = \"proto3\";"}`)},
		{ID: streamRoute.ID, GatewayID: gatewayInfo.ID, Type: constant.StreamRoute, Config: datatypes.JSON(`{"server_addr":"127.0.0.1","server_port":8080}`)},
	}

	err := backfillStoredSnapshotFields(ctx, resources)
	assert.NoError(t, err)
	assert.Equal(t, pluginConfig.Name, gjson.GetBytes(resources[0].Config, "name").String())
	assert.Equal(t, consumerGroup.ID, gjson.GetBytes(resources[1].Config, "id").String())
	assert.Equal(t, consumerGroup.Name, gjson.GetBytes(resources[1].Config, "name").String())
	assert.Equal(t, proto.Name, gjson.GetBytes(resources[2].Config, "name").String())
	assert.Equal(t, streamRoute.Name, gjson.GetBytes(resources[3].Config, "name").String())
	assert.Equal(t, "test", gjson.GetBytes(resources[3].Config, "labels.env").String())
}
```

同文件再补一个 `global_rule` 子测试，断言它和 `plugin_config` 一样会把数据库列上的 `Name` 回填进 snapshot `config.name`。

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestBackfillStoredSnapshotFields -count=1
```

Expected:
- FAIL，报 `undefined: backfillStoredSnapshotFields`

- [ ] **Step 3: 实现 helper，并替换 `kvToResource(...)` 里的 5 段数据库回填逻辑**

在 `unify_op_sync_helpers.go` 里新增：

```go
func backfillStoredSnapshotFields(ctx context.Context, resources []*model.GatewaySyncData) error {
	globalRuleMap := make(map[string]*model.GatewaySyncData)
	pluginConfigMap := make(map[string]*model.GatewaySyncData)
	consumerGroupMap := make(map[string]*model.GatewaySyncData)
	protoMap := make(map[string]*model.GatewaySyncData)
	streamRouteMap := make(map[string]*model.GatewaySyncData)

	var globalRuleIDs []string
	var pluginConfigIDs []string
	var consumerGroupIDs []string
	var protoIDs []string
	var streamRouteIDs []string

	for _, resource := range resources {
		switch resource.Type {
		case constant.GlobalRule:
			globalRuleMap[resource.ID] = resource
			globalRuleIDs = append(globalRuleIDs, resource.ID)
		case constant.PluginConfig:
			pluginConfigMap[resource.ID] = resource
			pluginConfigIDs = append(pluginConfigIDs, resource.ID)
		case constant.ConsumerGroup:
			consumerGroupMap[resource.ID] = resource
			consumerGroupIDs = append(consumerGroupIDs, resource.ID)
		case constant.Proto:
			protoMap[resource.ID] = resource
			protoIDs = append(protoIDs, resource.ID)
		case constant.StreamRoute:
			streamRouteMap[resource.ID] = resource
			streamRouteIDs = append(streamRouteIDs, resource.ID)
		}
	}

	if len(globalRuleIDs) > 0 {
		globalRules, err := QueryGlobalRules(ctx, map[string]any{"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID, "id": globalRuleIDs})
		if err != nil {
			return err
		}
		for _, globalRule := range globalRules {
			if resource, ok := globalRuleMap[globalRule.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", globalRule.Name)
			}
		}
	}

	if len(pluginConfigIDs) > 0 {
		pluginConfigs, err := QueryPluginConfigs(ctx, map[string]any{"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID, "id": pluginConfigIDs})
		if err != nil {
			return err
		}
		for _, pluginConfig := range pluginConfigs {
			if resource, ok := pluginConfigMap[pluginConfig.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", pluginConfig.Name)
			}
		}
	}

	if len(consumerGroupIDs) > 0 {
		consumerGroups, err := QueryConsumerGroups(ctx, map[string]any{"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID, "id": consumerGroupIDs})
		if err != nil {
			return err
		}
		for _, consumerGroup := range consumerGroups {
			if resource, ok := consumerGroupMap[consumerGroup.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "id", consumerGroup.ID)
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", consumerGroup.Name)
			}
		}
	}

	if len(protoIDs) > 0 {
		protos, err := QueryProtos(ctx, map[string]any{"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID, "id": protoIDs})
		if err != nil {
			return err
		}
		for _, proto := range protos {
			if resource, ok := protoMap[proto.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", proto.Name)
			}
		}
	}

	if len(streamRouteIDs) > 0 {
		streamRoutes, err := QueryStreamRoutes(ctx, map[string]any{"id": streamRouteIDs})
		if err != nil {
			return err
		}
		for _, streamRoute := range streamRoutes {
			if resource, ok := streamRouteMap[streamRoute.ID]; ok {
				resource.Config, _ = sjson.SetBytes(resource.Config, "name", streamRoute.Name)
				if labels := streamRoute.GetLabels(); labels != nil {
					resource.Config, _ = sjson.SetBytes(resource.Config, "labels", labels)
				}
			}
		}
	}
	return nil
}
```

然后把 `kvToResource(...)` 里原来的 `globalRuleIdMap` / `pluginConfigIdMap` / `consumerGroupIdMap` / `protoIdMap` / `streamRouteIdMap` 这 5 段逻辑删掉，改成在主循环完成后统一调用：

```go
if err := backfillStoredSnapshotFields(ctx, resources); err != nil {
	logging.Errorf("backfill stored snapshot fields error: %s", err.Error())
	return nil
}
```

- [ ] **Step 4: 运行任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestBackfillStoredSnapshotFields|TestSyncWithPrefix_SnapshotConfigShaping_CurrentSeam' -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/unify_op.go src/apiserver/pkg/biz/unify_op_sync_helpers.go src/apiserver/pkg/biz/unify_op_sync_helpers_test.go
git commit -m "refactor: extract sync snapshot field backfill helper"
```

---

### Task 3: 抽 plugin metadata snapshot ID 对齐 helper

- [ ] Task 3: 抽 plugin metadata snapshot ID 对齐 helper

**要解决的复杂度：** plugin metadata 的 snapshot 规则和其他资源不一样：etcd key 里的名字要先变成 snapshot `name`，再反查数据库现有记录，尽量复用已有 DB `id`，找不到才生成新 `id`。这段逻辑现在夹在 `kvToResource(...)` 里，和普通 name backfill 混在一起，很容易被误改。

**为什么这个任务适合单独提 PR：** 它是独立的特殊规则，和 Task 2 的“数据库列回填到 config”不是一类问题，单独拆能把风险隔离。

**Files:**
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers.go`
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/unify_op.go:695-779`

- [ ] **Step 1: 先补 plugin metadata ID 对齐 helper 的失败测试**

在 `unify_op_sync_helpers_test.go` 增加：

```go
func TestReconcilePluginMetadataSyncIDs(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	existing := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusSuccess)
	existing.Name = "limit-count-" + suffix
	existing.ID = idx.GenResourceID(constant.PluginMetadata)
	assert.NoError(t, CreatePluginMetadata(ctx, *existing))

	resources := []*model.GatewaySyncData{
		{
			ID:        existing.Name,
			GatewayID: gatewayInfo.ID,
			Type:      constant.PluginMetadata,
			Config:    datatypes.JSON(`{"value":{"disable":false}}`),
		},
		{
			ID:        "new-plugin",
			GatewayID: gatewayInfo.ID,
			Type:      constant.PluginMetadata,
			Config:    datatypes.JSON(`{"value":{"disable":true}}`),
		},
	}
	resources[0].SetName(existing.Name)
	resources[1].SetName("new-plugin")

	err := reconcilePluginMetadataSyncIDs(ctx, resources)
	assert.NoError(t, err)
	assert.Equal(t, existing.ID, resources[0].ID)
	assert.NotEmpty(t, resources[1].ID)
	assert.NotEqual(t, "new-plugin", resources[1].ID)
	assert.Equal(t, "new-plugin", resources[1].GetName())
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestReconcilePluginMetadataSyncIDs -count=1
```

Expected:
- FAIL，报 `undefined: reconcilePluginMetadataSyncIDs`

- [ ] **Step 3: 实现 helper，并替换 `kvToResource(...)` 里的 plugin metadata 分支**

在 `unify_op_sync_helpers.go` 里新增：

```go
func reconcilePluginMetadataSyncIDs(ctx context.Context, resources []*model.GatewaySyncData) error {
	metadataByName := make(map[string]*model.GatewaySyncData)
	var names []string

	for _, resource := range resources {
		if resource.Type != constant.PluginMetadata {
			continue
		}
		metadataByName[resource.ID] = resource
		names = append(names, resource.ID)
	}
	if len(names) == 0 {
		return nil
	}

	metadatas, err := QueryPluginMetadatas(ctx, map[string]any{
		"gateway_id": ginx.GetGatewayInfoFromContext(ctx).ID,
		"name":       names,
	})
	if err != nil {
		return err
	}

	existingIDByName := make(map[string]string)
	for _, metadata := range metadatas {
		existingIDByName[metadata.Name] = metadata.ID
	}

	for name, resource := range metadataByName {
		if existingID, ok := existingIDByName[name]; ok {
			resource.ID = existingID
			continue
		}
		resource.ID = idx.GenResourceID(constant.PluginMetadata)
	}
	return nil
}
```

然后把 `kvToResource(...)` 中原来的 `metadataNames` / `metadataNameMap` 以及后面的 plugin metadata 反查逻辑删掉，改成：

```go
if err := reconcilePluginMetadataSyncIDs(ctx, resources); err != nil {
	logging.Errorf("reconcile plugin metadata sync ids error: %s", err.Error())
	return nil
}
```

- [ ] **Step 4: 运行任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestReconcilePluginMetadataSyncIDs|TestSyncWithPrefix_SnapshotConfigShaping_CurrentSeam' -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/unify_op.go src/apiserver/pkg/biz/unify_op_sync_helpers.go src/apiserver/pkg/biz/unify_op_sync_helpers_test.go
git commit -m "refactor: extract plugin metadata sync id resolver"
```

---

### Task 4: 把 `kvToResource(...)` 收拢成 sync-data 本地 orchestration

- [ ] Task 4: 把 `kvToResource(...)` 收拢成 sync-data 本地 orchestration

**要解决的复杂度：** Task 1-3 抽完后，`kvToResource(...)` 里仍然会留下“循环遍历 KV -> 调基础 helper -> 调 DB 回填 -> 调 plugin metadata ID 对齐 -> 返回 resources”的 orchestration。这个函数本身已经不该继续手写细节，应该只表达同步快照构建的阶段顺序。

**为什么这个任务适合单独提 PR：** 这是纯 orchestration 重排，不新增新规则；前面 3 个任务已经把真实行为锁在 seam tests 和 helper tests 上了。

**Files:**
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers.go`
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/unify_op.go:690-818`

- [ ] **Step 1: 先补 orchestration helper 的失败测试**

在 `unify_op_sync_helpers_test.go` 增加：

```go
func TestBuildSyncSnapshotResources(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), gatewayInfo)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	pluginConfig := data.PluginConfig1WithNoRelation(gatewayInfo, constant.ResourceStatusSuccess)
	pluginConfig.Name = "pc-from-db-" + suffix
	assert.NoError(t, CreatePluginConfig(ctx, *pluginConfig))

	existingMetadata := data.PluginMetadata1(gatewayInfo, constant.ResourceStatusSuccess)
	existingMetadata.Name = "limit-count-" + suffix
	existingMetadata.ID = idx.GenResourceID(constant.PluginMetadata)
	assert.NoError(t, CreatePluginMetadata(ctx, *existingMetadata))

	kvList := []storage.KeyValuePair{
		{
			Key:         gatewayInfo.GetEtcdPrefixForList() + "routes/route-id",
			Value:       `{"uri":"/demo","create_time":1,"update_time":2}`,
			ModRevision: 1,
		},
		{
			Key:         gatewayInfo.GetEtcdPrefixForList() + "plugin_configs/" + pluginConfig.ID,
			Value:       `{"plugins":{}}`,
			ModRevision: 2,
		},
		{
			Key:         gatewayInfo.GetEtcdPrefixForList() + "plugin_metadata/" + existingMetadata.Name,
			Value:       `{"value":{"disable":false}}`,
			ModRevision: 3,
		},
	}

	resources, err := buildSyncSnapshotResources(ctx, gatewayInfo, kvList)
	assert.NoError(t, err)
	assert.Len(t, resources, 3)
	assert.Equal(t, "routes_route-id", gjson.GetBytes(resources[0].Config, "name").String())
	assert.Equal(t, "pc-from-db", gjson.GetBytes(resources[1].Config, "name").String())
	assert.Equal(t, existingMetadata.ID, resources[2].ID)
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestBuildSyncSnapshotResources -count=1
```

Expected:
- FAIL，报 `undefined: buildSyncSnapshotResources`

- [ ] **Step 3: 实现 orchestration helper，并让 `kvToResource(...)` 只做包装**

在 `unify_op_sync_helpers.go` 里新增：

```go
func buildSyncSnapshotResources(
	ctx context.Context,
	gatewayInfo *model.Gateway,
	kvList []storage.KeyValuePair,
) ([]*model.GatewaySyncData, error) {
	normalizedPrefix := model.NormalizeEtcdPrefix(gatewayInfo.EtcdConfig.Prefix)
	resources := make([]*model.GatewaySyncData, 0, len(kvList))

	for _, kv := range kvList {
		resource, ok := buildSyncedResourceFromKV(normalizedPrefix, gatewayInfo.ID, kv)
		if !ok {
			continue
		}
		resources = append(resources, resource)
	}

	if err := backfillStoredSnapshotFields(ctx, resources); err != nil {
		return nil, err
	}
	if err := reconcilePluginMetadataSyncIDs(ctx, resources); err != nil {
		return nil, err
	}
	return resources, nil
}
```

然后把 `kvToResource(...)` 改成：

```go
func (s *UnifyOp) kvToResource(
	ctx context.Context,
	kvList []storage.KeyValuePair,
) []*model.GatewaySyncData {
	resources, err := buildSyncSnapshotResources(ctx, s.gatewayInfo, kvList)
	if err != nil {
		logging.Errorf("build sync snapshot resources error: %s", err.Error())
		return nil
	}
	return resources
}
```

- [ ] **Step 4: 运行任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestBuildSyncSnapshotResources|TestSyncWithPrefix_SnapshotConfigShaping_CurrentSeam' -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/unify_op.go src/apiserver/pkg/biz/unify_op_sync_helpers.go src/apiserver/pkg/biz/unify_op_sync_helpers_test.go
git commit -m "refactor: thin kvToResource orchestration"
```

---

### Task 5: 给 `SyncWithPrefix(...)` 引入显式的 sync change-set planner

- [ ] Task 5: 给 `SyncWithPrefix(...)` 引入显式的 sync change-set planner

**要解决的复杂度：** 前 4 步做完后，`SyncWithPrefix(...)` 剩下的主要复杂度会集中在“根据 etcd snapshot 和数据库 snapshot 计算 create/update/delete 三组变更”。这部分已经有测试基座，但逻辑仍然直接展开在主函数里，不容易单测也不容易讨论后续优化。

**为什么这个任务适合单独提 PR：** 它只显式化 diff 规划，不改事务落库协议；属于同步尾部的本地收口动作。

**Files:**
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers.go`
- Modify: `src/apiserver/pkg/biz/unify_op_sync_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/unify_op.go:514-558`

- [ ] **Step 1: 先补 change-set planner 的失败测试**

在 `unify_op_sync_helpers_test.go` 增加：

```go
func TestBuildSyncChangeSet(t *testing.T) {
	route1ID := idx.GenResourceID(constant.Route)
	route2ID := idx.GenResourceID(constant.Route)
	route3ID := idx.GenResourceID(constant.Route)
	route4ID := idx.GenResourceID(constant.Route)

	databaseResources := []*model.GatewaySyncData{
		{
			AutoID:      11,
			ID:          route1ID,
			GatewayID:   gatewayInfo.ID,
			Type:        constant.Route,
			Config:      datatypes.JSON(`{"name":"route-1-old"}`),
			ModRevision: 1,
		},
		{
			AutoID:      22,
			ID:          route2ID,
			GatewayID:   gatewayInfo.ID,
			Type:        constant.Route,
			Config:      datatypes.JSON(`{"name":"route-2-delete"}`),
			ModRevision: 1,
		},
		{
			AutoID:      44,
			ID:          route4ID,
			GatewayID:   gatewayInfo.ID,
			Type:        constant.Route,
			Config:      datatypes.JSON(`{"name":"route-4-keep"}`),
			ModRevision: 5,
		},
	}

	etcdResources := []*model.GatewaySyncData{
		{
			ID:          route1ID,
			GatewayID:   gatewayInfo.ID,
			Type:        constant.Route,
			Config:      datatypes.JSON(`{"name":"route-1-new"}`),
			ModRevision: 2,
		},
		{
			ID:          route3ID,
			GatewayID:   gatewayInfo.ID,
			Type:        constant.Route,
			Config:      datatypes.JSON(`{"name":"route-3-new"}`),
			ModRevision: 1,
		},
		{
			ID:          route4ID,
			GatewayID:   gatewayInfo.ID,
			Type:        constant.Route,
			Config:      datatypes.JSON(`{"name":"route-4-keep"}`),
			ModRevision: 5,
		},
	}

	changeSet := buildSyncChangeSet(etcdResources, databaseResources)
	assert.Len(t, changeSet.ToCreate, 1)
	assert.Len(t, changeSet.ToUpdate, 1)
	assert.Len(t, changeSet.ToDeleteAutoIDs, 1)
	assert.Equal(t, route3ID, changeSet.ToCreate[0].ID)
	assert.Equal(t, 11, changeSet.ToUpdate[0].AutoID)
	assert.Equal(t, 2, changeSet.ToUpdate[0].ModRevision)
	assert.Equal(t, 22, changeSet.ToDeleteAutoIDs[0])
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestBuildSyncChangeSet -count=1
```

Expected:
- FAIL，报 `undefined: buildSyncChangeSet`

- [ ] **Step 3: 实现 change-set planner，并让 `SyncWithPrefix(...)` 复用它**

在 `unify_op_sync_helpers.go` 里新增：

```go
type syncChangeSet struct {
	ToCreate        []*model.GatewaySyncData
	ToUpdate        []*model.GatewaySyncData
	ToDeleteAutoIDs []int
}

func buildSyncChangeSet(
	etcdResources []*model.GatewaySyncData,
	databaseResources []*model.GatewaySyncData,
) syncChangeSet {
	etcdResourceMap := make(map[string]*model.GatewaySyncData, len(etcdResources))
	for _, resource := range etcdResources {
		etcdResourceMap[resource.GetResourceKey()] = resource
	}

	databaseResourceMap := make(map[string]*model.GatewaySyncData, len(databaseResources))
	var changeSet syncChangeSet
	for _, item := range databaseResources {
		databaseResourceMap[item.GetResourceKey()] = item
		if _, exists := etcdResourceMap[item.GetResourceKey()]; !exists {
			changeSet.ToDeleteAutoIDs = append(changeSet.ToDeleteAutoIDs, item.AutoID)
		}
	}

	for key, etcdResource := range etcdResourceMap {
		if dbResource, exists := databaseResourceMap[key]; exists {
			if dbResource.ModRevision != etcdResource.ModRevision {
				dbResource.Config = etcdResource.Config
				dbResource.ModRevision = etcdResource.ModRevision
				changeSet.ToUpdate = append(changeSet.ToUpdate, dbResource)
			}
			continue
		}
		changeSet.ToCreate = append(changeSet.ToCreate, etcdResource)
	}
	return changeSet
}
```

然后把 `SyncWithPrefix(...)` 中原来的 `resourcesToCreate` / `resourcesToUpdate` / `resourcesToDelete` 计算逻辑替换成：

```go
	changeSet := buildSyncChangeSet(resourceList, syncedItems)
```

并把事务里的更新、创建、删除分支改成使用：

```go
changeSet.ToUpdate
changeSet.ToCreate
changeSet.ToDeleteAutoIDs
```

最后把统计值收口成：

```go
syncedResourceTypeStats := make(map[constant.APISIXResource]int)
for _, resource := range changeSet.ToCreate {
	syncedResourceTypeStats[resource.Type]++
}
```

- [ ] **Step 4: 运行任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestBuildSyncChangeSet|TestSyncWithPrefix_(UpsertLogic|NoRaceCondition|BatchProcessing|ReturnsOnlyNewSnapshotCounts)' -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/unify_op.go src/apiserver/pkg/biz/unify_op_sync_helpers.go src/apiserver/pkg/biz/unify_op_sync_helpers_test.go
git commit -m "refactor: extract sync snapshot change-set planner"
```

## 完成标准

- `SyncWithPrefix(...)` 的真实 snapshot config shaping 有独立 characterization tests 锁住
- `kvToResource(...)` 不再直接承载：
  - 单 KV 规范化细节
  - DB 已有资源字段回填
  - plugin metadata ID 协调
- `SyncWithPrefix(...)` 中 create/update/delete 的计算逻辑变成显式 change-set，而不是继续散在主函数里
- 新 helper 都是 sync-data 域内本地 helper，没有提前抽成跨领域共享抽象
