# Publish Flow 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **Execution rule:** If a task or step is done, mark it in this `plan.md` before running `git add` and `git commit`.

**Goal:** 在不改变发布协议、不触碰 4 个输入源实现边界、不改 `HandleConfig()` 行为的前提下，逐步降低 `publish` 当前在 payload 改写、版本差异清理、依赖发布编排、以及最终 `ETCD` 校验上的复杂度。

**Architecture:** 本计划只处理 `publish` 领域自己的复杂度，不要求和 `web/open api/import/mcp` 做强一致，也不提前抽跨领域共享 builder。执行顺序调整为：先独立扩充现有 publish seam tests，优先收拢 payload cleanup、simple payload builder、dependency helper 和 `EtcdPublisher` validator 组装；`persist helper` 作为最后的低优先级收口，只在前 4 步完成后仍然明显改善可读性时再做。

**Tech Stack:** Go, GORM, `testify`, `gomonkey`, `ginkgo`, `gomock`, `go test`, `make lint`, `make test`

---

## 代码复核结论

- 重构目的判断：基本正确。当前 publish 的主要复杂度确实集中在 payload 清理、依赖 fan-out 和最终 validator 组装，而不是 CRUD 输入层。
- 复杂度评估：整体偏高，主要因为要补的是集成型 seam tests，而不是 helper 本身。`publish_test.go` 当前 mostly 只锁 happy path；`etcd_test.go` 也大量通过 patch `Validate()` 跳过了 validator 组装细节。
- 本次修正：把 seam-first 测试明确提成独立阶段；保留 Task 1-3 和 Task 5 为高优先级；Task 4 `persist helper` 降为最后的低优先级收口，因为当前已经有 `batchCreateEtcdResource(...)` 这层抽象。

## 执行顺序（修订）

1. Task 0：独立扩充 `publish_test.go` / `etcd_test.go` 的 characterization tests。
2. Task 1：先锁并抽字段清理 helper。
3. Task 2：再收 simple payload builder。
4. Task 3：之后拆依赖发布 helper。
5. Task 5：单独收拢 `EtcdPublisher` 的 validator 组装与 batch 校验。
6. Task 4：最后再评估是否值得加 `persist helper`。

## 范围

- 只处理 `src/apiserver/pkg/biz/publish.go`
- 只处理 `src/apiserver/pkg/publisher/etcd.go`
- 允许新增 publish 域内 helper 文件
- 允许补齐 `pkg/biz/publish_test.go` 与 `pkg/publisher/etcd_test.go`
- 允许在 `pkg/biz` 内增加少量本地 helper test 文件

## 非目标

- 不抽跨 `web/open api/import/mcp/publish` 的共享 helper
- 不改 `pkg/entity/model/*.go` 中各资源 `HandleConfig()`
- 不改 4 个输入源的 request / DATABASE 校验逻辑
- 不改数据库 schema
- 不把 delete 链路作为本轮主要目标，除非为了 publish helper 落地只做很小的伴随调整
- 不追求一次性把 11 类资源全部改造成声明式大表

## 当前测试缺口

- `src/apiserver/pkg/biz/publish_test.go` 目前主要覆盖“发布成功、同步成功、状态变更成功”，但对 `GatewaySyncData.Config` 里的最终 payload 字段断言不够细。
- `src/apiserver/pkg/biz/publish_test.go` 目前很少锁定“发布一个资源时，相关依赖资源是否也被一起发布”的现状。
- `src/apiserver/pkg/publisher/etcd_test.go` 目前主要覆盖 storage 调用和 `Validate()` 被调用这件事，但没有直接锁住：
  - `Validate()` 是否按当前网关版本构造 `ETCD` profile validator
  - `BatchCreate()` / `BatchUpdate()` 是否严格逐条做最终发布校验

## 文件结构

- `src/apiserver/pkg/biz/publish.go`
  - 当前发布主编排和各资源 `putXxx()/PutXxx()` 实现
- `src/apiserver/pkg/biz/publish_test.go`
  - 当前 publish 集成测试；后续继续作为 seam-first characterization test 主入口
- `src/apiserver/pkg/biz/publish_payload_helpers.go`
  - publish 域内 payload 清理、`BaseInfo` 合并、`ResourceOperation` 构造 helper
- `src/apiserver/pkg/biz/publish_payload_helpers_test.go`
  - payload helper 第二层单测
- `src/apiserver/pkg/biz/publish_dependency_helpers.go`
  - route/service/upstream/consumer/stream_route 的依赖收集 helper
- `src/apiserver/pkg/biz/publish_dependency_helpers_test.go`
  - dependency helper 第二层单测
- `src/apiserver/pkg/biz/publish_persist_helpers.go`
  - 发布后批量写 etcd + 更新资源状态的本地 wrapper
- `src/apiserver/pkg/biz/publish_persist_helpers_test.go`
  - persist helper 第二层单测
- `src/apiserver/pkg/publisher/etcd.go`
  - 最终 `ETCD` JSON Schema 校验和 etcd 写入
- `src/apiserver/pkg/publisher/etcd_test.go`
  - 当前 publisher 单测；后续补 `Validate()` 真正的 validator 组装断言

## PR 出口要求

- 每个任务里的 `go test` 是最小验收命令
- 每个任务准备合并前，再补跑一次：

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && make lint && make test
```

## 测试策略（必须）

- 新增 `Task 0` 作为独立步骤或独立 PR；在 `Task 0` 合并前，不开始 Task 1-5。
- 每个任务的第一组测试，必须先打在“重构前已经存在的 seam”上，不能直接从计划中新引入的 helper 开始写测试。
- characterization test 的目标不是证明“新 helper 没问题”，而是锁住当前 `master` 已有行为，保证重构前后同一组黑盒断言继续成立。
- helper 测试只能作为第二层测试：
  - 第一层：现有 publish seam
  - 第二层：抽出的 helper 单测
- `publish_test.go` 和 `etcd_test.go` 已经有可用基座，但当前覆盖 mostly 是 happy path；Task 0 先扩它们，不要新起 helper test 代替现有 seam。
- publish 计划里的现有 seam 优先级如下：
  - Task 0：优先扩 `publish_test.go` 的 payload / dependency 断言，以及 `etcd_test.go` 的 validator 组装断言
  - Task 1-4：优先直接测现有 `PublishXxx()` / `putXxx()` 路径，并通过 `GetSyncedItemByResourceTypeAndID(...)`、`GetXxx(...)`、`DiffResources(...)` 断言最终结果
  - Task 5：优先直接测现有 `EtcdPublisher.Validate()`、`BatchCreate()`、`BatchUpdate()`
- 执行时，每个任务固定使用下面的节奏：
  1. 先补现有 seam 的 characterization test
  2. 运行一次，确认当前代码已经被锁住
  3. 再补将要抽出的 helper test，让它先失败
  4. 用最小实现完成重构
  5. 重新运行 seam test + helper test

## 重构前测试前置阶段（独立）

- Task 0 至少覆盖 6 类现状：
  - 最终同步后的 payload 字段形态
  - 版本差异字段是否保留/删除
  - **每个资源实际上“clean 了哪些字段”的当前行为**——特别包括 `putRoutes/putServices/putUpstreams` **不会对 `id` 字段调 `ShouldRemoveFieldBeforeValidationOrPublish`**，`Consumer/Routes` **不会对 `name` 调**；只有 `ConsumerGroup/GlobalRule/Proto/SSL/PluginConfig/Consumer/ConsumerGroup` 特定子集会调。这一差异必须在 Task 0 里以 seam test 锁住，防止 Task 1 helper 统一运用时把规则错扩到其他资源。
  - 当前 `publish + sync` 黑盒结果里，`stream_route.labels` 仍会出现在 synced payload 中
  - 依赖资源是否跟随主资源一起发布（**包括 `service→upstream`、`stream_route→service/upstream` 两条真实存在的 fan-out，以及 `upstream` 当前不会自动带出 `ssl` 这条现状**，不只是 `route→deps` 和 `consumer→consumer_group`）
  - `EtcdPublisher.Validate()` 是否按网关版本组装正确的 `ETCD` profile validator
- Task 0 完成后，Task 1-3 和 Task 5 才有足够稳定的黑盒约束；否则 helper 抽完后很容易只验证“代码变短了”，却没验证最终 payload 还对。
- Task 4 只有在前面的任务都落完后，`putXxx()` 结尾的持久化收口仍然明显拖累可读性时再执行；它不是本轮最高杠杆点。

**行为不变量清单（review 必改）：** 以下现有行为被认为是不变的，helper 抽出后必须保持，由 Task 0 的 seam test 锁住：
1. `putRoutes/putServices/putUpstreams` **不**对 payload 的 `id` 字段调 `ShouldRemoveFieldBeforeValidationOrPublish`（这 3 种资源的 `id` 必须保留）。
2. `Routes/Consumer` **不**对 payload 的 `name` 字段调 `ShouldRemoveFieldBeforeValidationOrPublish`。
3. 当前 `publish + sync` seam 下，`consumer_group.name` 在 `3.11` / `3.13` 的 synced payload 中都会保留。
4. 当前 `publish + sync` seam 下，`stream_route.labels` 仍会保留在 synced payload 中。
5. `ssl.validity_start/validity_end` 无论在任何版本下都会被删除。
6. `BatchCreate/BatchUpdate` 在有任何一条校验失败时短路返回，不再继续校验后续资源。
7. 当前 `putUpstreams()` 不会自动发布关联 `ssl`；这条现状需要先锁住，后续不要在“重构”里顺手改行为。

### Task 0: 补 publish / validator characterization tests

- [x] Task 0: 补 publish / validator characterization tests

**要解决的缺口：** 当前文档已经把 seam-first 原则写清了，但正文还没有一个独立任务专门扩 `publish_test.go` 和 `etcd_test.go`。先把最终 payload 和 validator 组装锁住，后面的 helper 才有黑盒护栏。

**为什么这个任务适合单独提 PR：** 只扩现有测试文件，不改 `publish.go`、`etcd.go` 或发布主流程。

**Files:**
- Modify: `src/apiserver/pkg/biz/publish_test.go`
- Modify: `src/apiserver/pkg/publisher/etcd_test.go`

- [x] **Step 1: 扩充现有 seam tests，先锁发布结果和 validator 组装**

至少覆盖下面 6 类现状，全部断言现有黑盒结果：

- 最终同步后的 payload 字段形态（按代码真实黑盒结果锁住，包括 `stream_route.labels` 当前仍会出现在 synced payload）
- 不同网关版本下字段的保留/删除差异（按真实黑盒结果锁住，例如 `consumer_group.name` 当前在 `3.11` / `3.13` 的 synced payload 中都保留）
- 依赖资源是否跟随主资源一起发布：**至少覆盖 `route→upstream/service/plugin_config`、`consumer→consumer_group`、`service→upstream`、`stream_route→service/upstream`，以及 `upstream` 当前不会自动带出 `ssl` 这条现状**（Task 3 后续要按代码真实行为抽 helper，而不是按假设补行为）
- **每个资源“clean 了哪些字段”的当前行为**：`putRoutes/putServices/putUpstreams` 不清 `id`；`Routes/Consumer` 不清 `name`（锁住当前差异，防止 Task 1 误扩）
- `EtcdPublisher.Validate()` 是否按网关版本组装正确的 `ETCD` profile validator
- `BatchCreate/BatchUpdate` 在任一条资源校验失败时短路返回

- [x] **Step 2: 运行 publish seam tests，确认现状已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz ./pkg/publisher -count=1
```

Expected:
- PASS

- [x] **Step 3: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/publish_test.go src/apiserver/pkg/publisher/etcd_test.go
git commit -m "test: lock publish and validator seams"
```

---

### Task 1: 抽 publish 字段清理 helper

- [x] Task 1: 抽 publish 字段清理 helper

**要解决的复杂度：** 现在 `publish.go` 内部真正发生的字段清理散落在多个 `putXxx()/PutXxx()` 里，包括 `consumer.id`、若干资源的 `name` 版本差异、`ssl.validity_*`、以及 `stream_route.labels`。修改某个发布字段规则时，最容易出现“改了一个资源，漏了另一个资源”。

**代码真实情况：** `publish.go` 内部 cleanup 行为和 `publish + sync` 的黑盒结果并不完全一致。当前 seam 下，`consumer_group.name` 在 `3.11` / `3.13` 的 synced payload 中都会保留，`stream_route.labels` 也仍会出现在 synced payload 中。因此 Task 1 同时保留 seam test 和 helper test，分别锁黑盒结果与内部 cleanup 分支。

**为什么这个任务适合单独提 PR：** 只处理 payload 清理，不动依赖发布顺序，不动 `EtcdPublisher`，不会同时引入新的 orchestration 抽象。

**Files:**
- Create: `src/apiserver/pkg/biz/publish_payload_helpers.go`
- Create: `src/apiserver/pkg/biz/publish_payload_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/publish.go`
- Modify: `src/apiserver/pkg/biz/publish_test.go`

- [x] **Step 1: 先补现有 publish seam 的 characterization tests**

在 `publish_test.go` 增加下面这组测试，直接通过现有发布路径锁定最终同步后的 `GatewaySyncData.Config`：

```go
func TestPublishPayloadFieldCleanup_CurrentSeams(t *testing.T) {
	t.Run("consumer removes id before final publish payload", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-consumer-cleanup"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		consumer := data.Consumer1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		consumer.Config = datatypes.JSON(`{
			"id":"should-disappear",
			"username":"consumer1",
			"plugins":{"key-auth":{"key":"auth-one"}}
		}`)
		if err := CreateConsumer(ctx, *consumer); err != nil {
			t.Fatal(err)
		}

		if err := PublishConsumers(ctx, []string{consumer.ID}); err != nil {
			t.Fatal(err)
		}

		synced, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Consumer, consumer.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, gjson.GetBytes(synced.Config, "id").String())
		assert.Equal(t, "consumer1", gjson.GetBytes(synced.Config, "username").String())
	})

	t.Run("consumer group synced payload keeps name across versions", func(t *testing.T) {
		makeGateway := func(version string, name string) *model.Gateway {
			gw := data.Gateway1WithBkAPISIX()
			gw.Name = name
			gw.APISIXVersion = version
			return gw
		}

		runCase := func(t *testing.T, version string, gatewayName string, wantName bool) {
			gateway := makeGateway(version, gatewayName)
			if err := CreateGateway(context.Background(), gateway); err != nil {
				t.Fatal(err)
			}
			ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

			group := data.ConsumerGroup1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
			group.Config = datatypes.JSON(`{"plugins":{},"name":"cg-demo"}`)
			if err := CreateConsumerGroup(ctx, *group); err != nil {
				t.Fatal(err)
			}

			if err := PublishConsumerGroups(ctx, []string{group.ID}); err != nil {
				t.Fatal(err)
			}

			synced, err := GetSyncedItemByResourceTypeAndID(ctx, constant.ConsumerGroup, group.ID)
			if err != nil {
				t.Fatal(err)
			}
			gotName := gjson.GetBytes(synced.Config, "name").Exists()
			assert.Equal(t, wantName, gotName)
		}

		runCase(t, "3.11.0", "gateway-publish-cg-311", true)
		runCase(t, "3.13.0", "gateway-publish-cg-313", true)
	})

	t.Run("ssl synced payload removes internal validity fields", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-ssl"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		ssl := data.SSL1(gateway, constant.ResourceStatusCreateDraft)
		ssl.Config, _ = sjson.SetBytes(ssl.Config, "validity_start", 1710000000)
		ssl.Config, _ = sjson.SetBytes(ssl.Config, "validity_end", 1810000000)
		if err := CreateSSL(ctx, ssl); err != nil {
			t.Fatal(err)
		}

		if err := PublishSSLs(ctx, []string{ssl.ID}); err != nil {
			t.Fatal(err)
		}

		synced, err := GetSyncedItemByResourceTypeAndID(ctx, constant.SSL, ssl.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, gjson.GetBytes(synced.Config, "validity_start").Exists())
		assert.False(t, gjson.GetBytes(synced.Config, "validity_end").Exists())
	})
}
```

- [x] **Step 2: 运行 seam tests，确认当前行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestPublishPayloadFieldCleanup_CurrentSeams -count=1
```

Expected:
- PASS

- [x] **Step 3: 再补 helper test，让它先失败**

**执行备注（按代码真实情况）：** 最终落地的 helper test 额外覆盖了 `global_rule`、`proto`、`stream_route.name` 的版本差异，以及 `route/service/upstream` “整体保持原样”的断言；它锁的是 `publish.go` 内部 cleanup 分支，而不是最终 synced payload。

在 `publish_payload_helpers_test.go` 增加：

```go
func TestCleanupPublishPayloadFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		rawConfig    string
		wantConfig   string
	}{
		{
			name:         "consumer drops id",
			resourceType: constant.Consumer,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"consumer-id","username":"demo","plugins":{}}`,
			wantConfig:   `{"username":"demo","plugins":{}}`,
		},
		{
			name:         "consumer group drops name in 3.11",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"cg-id","name":"cg-demo","plugins":{}}`,
			wantConfig:   `{"id":"cg-id","plugins":{}}`,
		},
		{
			name:         "consumer group keeps name in 3.13",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			rawConfig:    `{"id":"cg-id","name":"cg-demo","plugins":{}}`,
			wantConfig:   `{"id":"cg-id","name":"cg-demo","plugins":{}}`,
		},
		{
			name:         "ssl removes internal validity fields",
			resourceType: constant.SSL,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"ssl-id","validity_start":1,"validity_end":2,"cert":"x","key":"y","snis":["demo.com"]}`,
			wantConfig:   `{"id":"ssl-id","cert":"x","key":"y","snis":["demo.com"]}`,
		},
		{
			name:         "stream route removes labels",
			resourceType: constant.StreamRoute,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"sr-id","labels":{"env":"prod"},"server_addr":"0.0.0.0","server_port":9100,"upstream":{"nodes":[{"host":"127.0.0.1","port":80,"weight":1}],"type":"roundrobin"}}`,
			wantConfig:   `{"id":"sr-id","server_addr":"0.0.0.0","server_port":9100,"upstream":{"nodes":[{"host":"127.0.0.1","port":80,"weight":1}],"type":"roundrobin"}}`,
		},
		// review 补充：锁住“id 不应在 route/service/upstream 上被清除”
		{
			name:         "route keeps id regardless of version",
			resourceType: constant.Route,
			version:      constant.APISIXVersion311,
			rawConfig:    `{"id":"route-id","uri":"/demo"}`,
			wantConfig:   `{"id":"route-id","uri":"/demo"}`,
		},
		{
			name:         "service keeps id regardless of version",
			resourceType: constant.Service,
			version:      constant.APISIXVersion313,
			rawConfig:    `{"id":"svc-id","upstream_id":"u-id"}`,
			wantConfig:   `{"id":"svc-id","upstream_id":"u-id"}`,
		},
		{
			name:         "upstream keeps id regardless of version",
			resourceType: constant.Upstream,
			version:      constant.APISIXVersion313,
			rawConfig:    `{"id":"u-id","nodes":[]}`,
			wantConfig:   `{"id":"u-id","nodes":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanupPublishPayloadFields(publishPayloadCleanupInput{
				ResourceType: tt.resourceType,
				Version:      tt.version,
				RawConfig:    json.RawMessage(tt.rawConfig),
			})
			assert.JSONEq(t, tt.wantConfig, string(got))
		})
	}
}
```

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestCleanupPublishPayloadFields -count=1
```

Expected:
- FAIL，报 `undefined: cleanupPublishPayloadFields` 或 `undefined: publishPayloadCleanupInput`

- [x] **Step 4: 用最小实现抽出 publish 字段清理 helper，并接回现有 `putXxx()`**

**执行备注（按代码真实情况）：** 最终 helper 规则表只保留原始 `publish.go` 里真实发生过的字段删除：`consumer.id`、`consumer_group.name`、`global_rule.name`、`proto.name`、`ssl.name/validity_*`、`stream_route.name/labels`。`plugin_config` 当前接入 helper 但规则表为空，因此不会改动 payload；`putRoutes/putServices/putUpstreams` 继续不接入 helper。

**【Blocker修正】（review）：** 原计划第一版 helper 用 `ShouldRemoveFieldBeforeValidationOrPublish(resourceType, "id"/"name", version)` 做统一判断，但原代码并非所有资源都调用这个方法：`putRoutes/putServices/putUpstreams` 不对 `id` 调，`Routes/Consumer` 不对 `name` 调。统一调用会把字段清理规则扩散到原本不清理的资源上，产生**静默字段丢失**（上线不易察觉）。改用 **resource-specific policy table** 表达清理规则：

在 `publish_payload_helpers.go` 新增：

```go
package biz

import (
	"encoding/json"

	"github.com/tidwall/sjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

type publishPayloadCleanupInput struct {
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	RawConfig    json.RawMessage
}

// cleanupRule describes a single field cleanup decision for a resource type.
// Field          : JSON path to delete.
// VersionGated   : if true, only delete when
//                  constant.ShouldRemoveFieldBeforeValidationOrPublish(resourceType, field, version) == true.
//                  Otherwise always delete.
type cleanupRule struct {
	Field        string
	VersionGated bool
}

// cleanupRules is the resource-specific policy table. Add a row here ONLY for
// resources whose original putXxx()/PutXxx() actually removed the given field.
// Do NOT add id/name rows for Route/Service/Upstream: their original behavior
// does NOT strip id/name, so adding them here would be a silent regression.
var cleanupRules = map[constant.APISIXResource][]cleanupRule{
	constant.PluginConfig:   {{Field: "id", VersionGated: true}},
	constant.Consumer:       {{Field: "id", VersionGated: true}},
	constant.ConsumerGroup:  {{Field: "id", VersionGated: true}, {Field: "name", VersionGated: true}},
	constant.GlobalRule:     {{Field: "id", VersionGated: true}, {Field: "name", VersionGated: true}},
	constant.Proto:          {{Field: "id", VersionGated: true}, {Field: "name", VersionGated: true}},
	constant.SSL:            {{Field: "id", VersionGated: true}, {Field: "name", VersionGated: true}, {Field: "validity_start"}, {Field: "validity_end"}},
	constant.StreamRoute:    {{Field: "labels"}},
	// Route/Service/Upstream intentionally absent: they do not strip id/name in master.
}

func cleanupPublishPayloadFields(input publishPayloadCleanupInput) json.RawMessage {
	cleaned := append(json.RawMessage(nil), input.RawConfig...)
	for _, rule := range cleanupRules[input.ResourceType] {
		if rule.VersionGated &&
			!constant.ShouldRemoveFieldBeforeValidationOrPublish(input.ResourceType, rule.Field, input.Version) {
			continue
		}
		cleaned, _ = sjson.DeleteBytes(cleaned, rule.Field)
	}
	return cleaned
}
```

然后把下面这些函数里重复的删字段逻辑改成统一调用（**仅对表里有规则的资源**）：

- `putPluginConfigs(...)`
- `putConsumers(...)`
- `putConsumerGroups(...)`
- `putGlobalRules(...)`
- `PutProtos(...)`
- `PutSSLs(...)`
- `PutStreamRoutes(...)`

`putRoutes/putServices/putUpstreams` **不要**接入（它们本来就不做字段清理）。

接入方式统一改成：

```go
resource.Config = cleanupPublishPayloadFields(publishPayloadCleanupInput{
	ResourceType: constant.PluginConfig,
	Version:      apisixVersion,
	RawConfig:    resource.Config,
})
```

- [x] **Step 5: 运行任务相关测试，确认 seam 和 helper 都通过**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestPublishPayloadFieldCleanup_CurrentSeams|TestCleanupPublishPayloadFields' -count=1
```

Expected:
- PASS

- [x] **Step 6: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/publish.go src/apiserver/pkg/biz/publish_test.go src/apiserver/pkg/biz/publish_payload_helpers.go src/apiserver/pkg/biz/publish_payload_helpers_test.go
git commit -m "refactor: extract publish payload cleanup helper"
```

### Task 2: 抽 simple publish payload builder helper

- [x] Task 2: 抽 simple publish payload builder helper

**要解决的复杂度：** `plugin_config`、`plugin_metadata`、`consumer_group`、`global_rule`、`proto`、`ssl` 这些资源都在重复做 `BaseInfo` 序列化、`jsonx.MergeJson(...)`、`ResourceOperation` 组装。重复逻辑多，资源特例又掺在里面，`putXxx()` 很难一眼看出真正的资源特有部分。

**为什么这个任务适合单独提 PR：** 只收敛“无依赖资源”的 payload 组装，不碰 route/service/upstream/consumer/stream_route 的依赖发布顺序。

**Files:**
- Modify: `src/apiserver/pkg/biz/publish.go`
- Modify: `src/apiserver/pkg/biz/publish_test.go`
- Modify: `src/apiserver/pkg/biz/publish_payload_helpers.go`
- Modify: `src/apiserver/pkg/biz/publish_payload_helpers_test.go`

- [x] **Step 1: 先补现有 simple publish seam 的 characterization tests**

在 `publish_test.go` 增加：

```go
func TestSimplePublishPayload_CurrentSeams(t *testing.T) {
	t.Run("plugin metadata keeps plugin name in final payload id", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-plugin-metadata"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		pm := data.PluginMetadata1(gateway, constant.ResourceStatusCreateDraft)
		if err := CreatePluginMetadata(ctx, *pm); err != nil {
			t.Fatal(err)
		}

		if err := PublishPluginMetadatas(ctx, []string{pm.ID}); err != nil {
			t.Fatal(err)
		}

		synced, err := GetSyncedItemByResourceTypeAndID(ctx, constant.PluginMetadata, pm.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, pm.Name, gjson.GetBytes(synced.Config, "id").String())
		assert.Equal(t, pm.Name, gjson.GetBytes(synced.Config, "name").String())
	})

	t.Run("proto keeps name on 3.13", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-proto-313"
		gateway.APISIXVersion = "3.13.0"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		pb := data.Proto1(gateway, constant.ResourceStatusCreateDraft)
		pb.Config, _ = sjson.SetBytes(pb.Config, "name", pb.Name)
		if err := CreateProto(ctx, *pb); err != nil {
			t.Fatal(err)
		}

		if err := PublishProtos(ctx, []string{pb.ID}); err != nil {
			t.Fatal(err)
		}

		synced, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Proto, pb.ID)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, pb.Name, gjson.GetBytes(synced.Config, "name").String())
		assert.Equal(t, pb.ID, gjson.GetBytes(synced.Config, "id").String())
	})
}
```

- [x] **Step 2: 运行 seam tests，确认当前外部行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestSimplePublishPayload_CurrentSeams -count=1
```

Expected:
- PASS

- [x] **Step 3: 再补 helper test，让它先失败**

在 `publish_payload_helpers_test.go` 追加：

```go
func TestBuildPublishResourceOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input publishResourceOperationInput
		wantKey string
		wantConfig string
	}{
		{
			name: "plugin metadata uses plugin name as key and id",
			input: publishResourceOperationInput{
				ResourceType: constant.PluginMetadata,
				ResourceKey:  "limit-count",
				BaseInfo: entity.BaseInfo{
					ID:         "limit-count",
					CreateTime: 1700000000,
					UpdateTime: 1700000001,
				},
				Version:   constant.APISIXVersion311,
				RawConfig: json.RawMessage(`{"log_format":{"client_ip":"$remote_addr"}}`),
			},
			wantKey:    "limit-count",
			wantConfig: `{"id":"limit-count","create_time":1700000000,"update_time":1700000001,"log_format":{"client_ip":"$remote_addr"}}`,
		},
		{
			name: "consumer group keeps id and removes name in 3.11",
			input: publishResourceOperationInput{
				ResourceType: constant.ConsumerGroup,
				ResourceKey:  "cg-id",
				BaseInfo: entity.BaseInfo{
					ID:         "cg-id",
					CreateTime: 1700000000,
					UpdateTime: 1700000001,
				},
				Version:   constant.APISIXVersion311,
				RawConfig: json.RawMessage(`{"id":"cg-id","name":"cg-demo","plugins":{}}`),
			},
			wantKey:    "cg-id",
			wantConfig: `{"id":"cg-id","create_time":1700000000,"update_time":1700000001,"plugins":{}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPublishResourceOperation(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantKey, got.Key)
			assert.Equal(t, tt.input.ResourceType, got.Type)
			assert.JSONEq(t, tt.wantConfig, string(got.Config))
		})
	}
}
```

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestBuildPublishResourceOperation -count=1
```

Expected:
- FAIL，报 `undefined: publishResourceOperationInput` 或 `undefined: buildPublishResourceOperation`

- [x] **Step 4: 用最小实现抽出 simple publish payload builder，并迁移 simple 资源**

**执行备注（按代码真实情况）：** 最终 helper 统一承接 `BaseInfo` 序列化、`jsonx.MergeJson(...)`、Task 1 的 cleanup helper，以及 `publisher.ResourceOperation` 组装。迁移范围保持在 simple publish loops：`putPluginConfigs(...)`、`putPluginMetadatas(...)`、`putConsumerGroups(...)`、`putGlobalRules(...)`、`PutProtos(...)`、`PutSSLs(...)`。

**设计注释（review）：** helper 返回值使用值类型 `publisher.ResourceOperation`，让 `MergeJson` 失败时零值 op 不会被误用（调用者遇到 err 必须立即返回，不能跟着 append）。主要用意图通过 Task 2 seam test 和各调用处 `if err != nil { return err }` 模式保证。

在 `publish_payload_helpers.go` 追加：

```go
type publishResourceOperationInput struct {
	ResourceType constant.APISIXResource
	ResourceKey  string
	BaseInfo     entity.BaseInfo
	Version      constant.APISIXVersion
	RawConfig    json.RawMessage
}

func buildPublishResourceOperation(input publishResourceOperationInput) (publisher.ResourceOperation, error) {
	baseConfig, err := json.Marshal(input.BaseInfo)
	if err != nil {
		return publisher.ResourceOperation{}, err
	}
	merged, err := jsonx.MergeJson(input.RawConfig, baseConfig)
	if err != nil {
		return publisher.ResourceOperation{}, err
	}
	cleaned := cleanupPublishPayloadFields(publishPayloadCleanupInput{
		ResourceType: input.ResourceType,
		Version:      input.Version,
		RawConfig:    merged,
	})
	return publisher.ResourceOperation{
		Key:    input.ResourceKey,
		Config: cleaned,
		Type:   input.ResourceType,
	}, nil
}
```

然后把下面这些函数里重复的 `BaseInfo` + `MergeJson` + `ResourceOperation` 组装替换成 `buildPublishResourceOperation(...)`：

- `putPluginConfigs(...)`
- `putPluginMetadatas(...)`
- `putConsumerGroups(...)`
- `putGlobalRules(...)`
- `PutProtos(...)`
- `PutSSLs(...)`

迁移后，每个循环里只保留资源特有输入：

```go
op, err := buildPublishResourceOperation(publishResourceOperationInput{
	ResourceType: constant.PluginMetadata,
	ResourceKey:  pluginMetadata.Name,
	BaseInfo: entity.BaseInfo{
		ID:         pluginMetadata.Name,
		CreateTime: pluginMetadata.CreatedAt.Unix(),
		UpdateTime: pluginMetadata.UpdatedAt.Unix(),
	},
	Version:   apisixVersion,
	RawConfig: pluginMetadata.Config,
})
if err != nil {
	return fmt.Errorf("marshal plugin metadata base info failed: %w", err)
}
pluginMetadataOps = append(pluginMetadataOps, op)
```

- [x] **Step 5: 运行任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestSimplePublishPayload_CurrentSeams|TestBuildPublishResourceOperation|TestCleanupPublishPayloadFields' -count=1
```

Expected:
- PASS

- [x] **Step 6: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/publish.go src/apiserver/pkg/biz/publish_test.go src/apiserver/pkg/biz/publish_payload_helpers.go src/apiserver/pkg/biz/publish_payload_helpers_test.go
git commit -m "refactor: extract simple publish payload builder"
```

### Task 3: 拆依赖资源发布 helper

- [ ] Task 3: 拆依赖资源发布 helper

**要解决的复杂度：** `putRoutes()`、`putServices()`、`putUpstreams()`、`putConsumers()`、`PutStreamRoutes()` 现在把“读取资源”“收集依赖 ID”“递归发布依赖”“构造自身 payload”混在一起，读起来像 5 个相似但不完全一致的大函数。

**为什么这个任务适合单独提 PR：** 这一任务只处理“有依赖资源的发布编排”，不碰 simple 资源，不碰最终 `EtcdPublisher`。

**Files:**
- Create: `src/apiserver/pkg/biz/publish_dependency_helpers.go`
- Create: `src/apiserver/pkg/biz/publish_dependency_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/publish.go`
- Modify: `src/apiserver/pkg/biz/publish_test.go`

- [ ] **Step 0: 前置检查 `model.Upstream.GetSSLID()` 存在（review 补充）**

Task 3 中 `collectUpstreamPublishDependencies` 使用了 `upstream.GetSSLID()`，在开始实现前先确认此 method 在 `pkg/entity/model/upstream.go` 中已存在：

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && grep -n 'func.*Upstream.*GetSSLID' pkg/entity/model/upstream.go || echo 'MISSING'
```

如果上面输出 `MISSING`，需在本 Task 的第一个 commit 里先补齐 getter（只加方法、不改资源字段定义）。

- [ ] **Step 1: 先补现有依赖发布 seam 的 characterization tests**

在 `publish_test.go` 增加：

```go
func TestPublishDependencies_CurrentSeams(t *testing.T) {
	t.Run("publishing route also publishes related upstream service and plugin config", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-route-deps"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}

		service := data.Service1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		service.UpstreamID = upstream.ID
		if err := CreateService(ctx, *service); err != nil {
			t.Fatal(err)
		}

		pc := data.PluginConfig1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := CreatePluginConfig(ctx, *pc); err != nil {
			t.Fatal(err)
		}

		route := data.Route1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
		route.ServiceID = service.ID
		route.UpstreamID = upstream.ID
		route.PluginConfigID = pc.ID
		if err := CreateRoute(ctx, *route); err != nil {
			t.Fatal(err)
		}

		if err := PublishRoutes(ctx, []string{route.ID}); err != nil {
			t.Fatal(err)
		}

		syncedRoute, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Route, route.ID)
		assert.NoError(t, err)
		assert.Equal(t, service.ID, gjson.GetBytes(syncedRoute.Config, "service_id").String())
		assert.Equal(t, upstream.ID, gjson.GetBytes(syncedRoute.Config, "upstream_id").String())
		assert.Equal(t, pc.ID, gjson.GetBytes(syncedRoute.Config, "plugin_config_id").String())

		_, err = GetSyncedItemByResourceTypeAndID(ctx, constant.Service, service.ID)
		assert.NoError(t, err)
		_, err = GetSyncedItemByResourceTypeAndID(ctx, constant.Upstream, upstream.ID)
		assert.NoError(t, err)
		_, err = GetSyncedItemByResourceTypeAndID(ctx, constant.PluginConfig, pc.ID)
		assert.NoError(t, err)
	})

	t.Run("publishing consumer also publishes related consumer group", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-consumer-deps"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		group := data.ConsumerGroup1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := CreateConsumerGroup(ctx, *group); err != nil {
			t.Fatal(err)
		}

		consumer := data.Consumer1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		consumer.GroupID = group.ID
		if err := CreateConsumer(ctx, *consumer); err != nil {
			t.Fatal(err)
		}

		if err := PublishConsumers(ctx, []string{consumer.ID}); err != nil {
			t.Fatal(err)
		}

		_, err = GetSyncedItemByResourceTypeAndID(ctx, constant.ConsumerGroup, group.ID)
		assert.NoError(t, err)
	})

	// review 补充 3 条原先漏覆盖的 fan-out，给 Task 3 的 5 个 collect* helper 1:1 的 seam 护栏
	t.Run("publishing service also publishes related upstream", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-service-deps"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}

		service := data.Service1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		service.UpstreamID = upstream.ID
		if err := CreateService(ctx, *service); err != nil {
			t.Fatal(err)
		}

		if err := PublishServices(ctx, []string{service.ID}); err != nil {
			t.Fatal(err)
		}

		_, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Upstream, upstream.ID)
		assert.NoError(t, err)
	})

	t.Run("publishing upstream also publishes related ssl", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-upstream-ssl"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		ssl := data.SSL1(gateway, constant.ResourceStatusCreateDraft)
		if err := CreateSSL(ctx, ssl); err != nil {
			t.Fatal(err)
		}

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		upstream.Config, _ = sjson.SetBytes(upstream.Config, "tls.client_cert_id", ssl.ID)
		if err := CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}

		if err := PublishUpstreams(ctx, []string{upstream.ID}); err != nil {
			t.Fatal(err)
		}

		_, err := GetSyncedItemByResourceTypeAndID(ctx, constant.SSL, ssl.ID)
		assert.NoError(t, err)
	})

	t.Run("publishing stream route also publishes related service and upstream", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-stream-deps"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		upstream := data.Upstream1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := CreateUpstream(ctx, *upstream); err != nil {
			t.Fatal(err)
		}
		service := data.Service1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		service.UpstreamID = upstream.ID
		if err := CreateService(ctx, *service); err != nil {
			t.Fatal(err)
		}

		sr := data.StreamRoute1WithNoRelationResource(gateway, constant.ResourceStatusCreateDraft)
		sr.ServiceID = service.ID
		sr.UpstreamID = upstream.ID
		if err := CreateStreamRoute(ctx, *sr); err != nil {
			t.Fatal(err)
		}

		if err := PublishStreamRoutes(ctx, []string{sr.ID}); err != nil {
			t.Fatal(err)
		}

		_, err := GetSyncedItemByResourceTypeAndID(ctx, constant.Service, service.ID)
		assert.NoError(t, err)
		_, err = GetSyncedItemByResourceTypeAndID(ctx, constant.Upstream, upstream.ID)
		assert.NoError(t, err)
	})
}
```

- [ ] **Step 2: 运行 seam tests，确认依赖 fan-out 行为被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestPublishDependencies_CurrentSeams -count=1
```

Expected:
- PASS

- [ ] **Step 3: 再补 dependency helper tests，让它先失败**

在 `publish_dependency_helpers_test.go` 增加：

```go
func TestCollectRoutePublishDependencies(t *testing.T) {
	t.Parallel()

	routes := []*model.Route{
		{
			ServiceID:      "service-1",
			UpstreamID:     "upstream-1",
			PluginConfigID: "plugin-config-1",
		},
		{
			ServiceID:      "service-2",
			UpstreamID:     "",
			PluginConfigID: "plugin-config-2",
		},
	}

	deps := collectRoutePublishDependencies(routes)
	assert.Equal(t, []string{"service-1", "service-2"}, deps.ServiceIDs)
	assert.Equal(t, []string{"upstream-1"}, deps.UpstreamIDs)
	assert.Equal(t, []string{"plugin-config-1", "plugin-config-2"}, deps.PluginConfigIDs)
}

func TestCollectConsumerPublishDependencies(t *testing.T) {
	t.Parallel()

	consumers := []*model.Consumer{
		{GroupID: "group-1"},
		{GroupID: ""},
		{GroupID: "group-2"},
	}

	deps := collectConsumerPublishDependencies(consumers)
	assert.Equal(t, []string{"group-1", "group-2"}, deps.ConsumerGroupIDs)
}
```

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestCollectRoutePublishDependencies|TestCollectConsumerPublishDependencies' -count=1
```

Expected:
- FAIL，报 `undefined: collectRoutePublishDependencies` 等

- [ ] **Step 4: 用最小实现抽出 dependency helper，并迁移 dependent 资源**

在 `publish_dependency_helpers.go` 新增：

```go
package biz

import "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"

type routePublishDependencies struct {
	ServiceIDs      []string
	UpstreamIDs     []string
	PluginConfigIDs []string
}

type servicePublishDependencies struct {
	UpstreamIDs []string
}

type upstreamPublishDependencies struct {
	SSLIDs []string
}

type consumerPublishDependencies struct {
	ConsumerGroupIDs []string
}

type streamRoutePublishDependencies struct {
	ServiceIDs  []string
	UpstreamIDs []string
}

func collectRoutePublishDependencies(routes []*model.Route) routePublishDependencies {
	deps := routePublishDependencies{}
	for _, route := range routes {
		if route.ServiceID != "" {
			deps.ServiceIDs = append(deps.ServiceIDs, route.ServiceID)
		}
		if route.UpstreamID != "" {
			deps.UpstreamIDs = append(deps.UpstreamIDs, route.UpstreamID)
		}
		if route.PluginConfigID != "" {
			deps.PluginConfigIDs = append(deps.PluginConfigIDs, route.PluginConfigID)
		}
	}
	return deps
}

func collectServicePublishDependencies(services []*model.Service) servicePublishDependencies {
	deps := servicePublishDependencies{}
	for _, service := range services {
		if service.UpstreamID != "" {
			deps.UpstreamIDs = append(deps.UpstreamIDs, service.UpstreamID)
		}
	}
	return deps
}

func collectUpstreamPublishDependencies(upstreams []*model.Upstream) upstreamPublishDependencies {
	deps := upstreamPublishDependencies{}
	for _, upstream := range upstreams {
		if upstream.GetSSLID() != "" {
			deps.SSLIDs = append(deps.SSLIDs, upstream.GetSSLID())
		}
	}
	return deps
}

func collectConsumerPublishDependencies(consumers []*model.Consumer) consumerPublishDependencies {
	deps := consumerPublishDependencies{}
	for _, consumer := range consumers {
		if consumer.GroupID != "" {
			deps.ConsumerGroupIDs = append(deps.ConsumerGroupIDs, consumer.GroupID)
		}
	}
	return deps
}

func collectStreamRoutePublishDependencies(streamRoutes []*model.StreamRoute) streamRoutePublishDependencies {
	deps := streamRoutePublishDependencies{}
	for _, sr := range streamRoutes {
		if sr.ServiceID != "" {
			deps.ServiceIDs = append(deps.ServiceIDs, sr.ServiceID)
		}
		if sr.UpstreamID != "" {
			deps.UpstreamIDs = append(deps.UpstreamIDs, sr.UpstreamID)
		}
	}
	return deps
}
```

然后迁移：

- `putRoutes(...)`
- `putServices(...)`
- `putUpstreams(...)`
- `putConsumers(...)`
- `PutStreamRoutes(...)`

迁移目标不是发明统一 mega-helper，而是把每个函数中的“依赖收集”和“payload 构造”显式分开，例如：

```go
deps := collectRoutePublishDependencies(routes)
if len(deps.UpstreamIDs) > 0 {
	if err := putUpstreams(ctx, deps.UpstreamIDs); err != nil {
		return err
	}
}
if len(deps.ServiceIDs) > 0 {
	if err := putServices(ctx, deps.ServiceIDs); err != nil {
		return err
	}
}
if len(deps.PluginConfigIDs) > 0 {
	if err := putPluginConfigs(ctx, deps.PluginConfigIDs); err != nil {
		return err
	}
}
```

同时让这些函数内部 payload 构造统一走 Task 2 的 `buildPublishResourceOperation(...)`。

- [ ] **Step 5: 运行任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestPublishDependencies_CurrentSeams|TestCollectRoutePublishDependencies|TestCollectConsumerPublishDependencies|TestBuildPublishResourceOperation' -count=1
```

Expected:
- PASS

- [ ] **Step 6: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/publish.go src/apiserver/pkg/biz/publish_test.go src/apiserver/pkg/biz/publish_dependency_helpers.go src/apiserver/pkg/biz/publish_dependency_helpers_test.go src/apiserver/pkg/biz/publish_payload_helpers.go src/apiserver/pkg/biz/publish_payload_helpers_test.go
git commit -m "refactor: split publish dependency helpers"
```

### Task 4: 抽 publish persist helper

- [ ] Task 4: 抽 publish persist helper

**要解决的复杂度：** 每个 `putXxx()` 结尾都在重复 `batchCreateEtcdResource(...)` + `BatchUpdateResourceStatus(...)` + 中文错误包装。重复多，`putXxx()` 长度被进一步拉大，而且将来改发布成功后的统一动作时必须逐个资源修改。

**为什么这个任务适合单独提 PR：** 这是 publish 领域自己的收口动作，不改变 payload 形态，也不改变依赖发布顺序。

**Files:**
- Create: `src/apiserver/pkg/biz/publish_persist_helpers.go`
- Create: `src/apiserver/pkg/biz/publish_persist_helpers_test.go`
- Modify: `src/apiserver/pkg/biz/publish.go`
- Modify: `src/apiserver/pkg/biz/publish_test.go`

- [ ] **Step 1: 先补现有 persist seam 的 characterization tests**

在 `publish_test.go` 增加一个成功路径 characterization test，继续锁定“写 etcd + 更新状态”这个现有出口：

```go
func TestPublishPersist_CurrentSeams(t *testing.T) {
	t.Run("plugin config publish still writes synced config and updates status", func(t *testing.T) {
		gateway := data.Gateway1WithBkAPISIX()
		gateway.Name = "gateway-publish-persist"
		if err := CreateGateway(context.Background(), gateway); err != nil {
			t.Fatal(err)
		}
		ctx := ginx.SetGatewayInfoToContext(context.Background(), gateway)

		pc := data.PluginConfig1WithNoRelation(gateway, constant.ResourceStatusCreateDraft)
		if err := CreatePluginConfig(ctx, *pc); err != nil {
			t.Fatal(err)
		}

		if err := PublishPluginConfigs(ctx, []string{pc.ID}); err != nil {
			t.Fatal(err)
		}

		synced, err := GetSyncedItemByResourceTypeAndID(ctx, constant.PluginConfig, pc.ID)
		assert.NoError(t, err)
		assert.Equal(t, pc.ID, synced.ID)

		stored, err := GetPluginConfig(ctx, pc.ID)
		assert.NoError(t, err)
		assert.Equal(t, constant.ResourceStatusSuccess, stored.Status)
	})
}
```

- [ ] **Step 2: 运行 seam test，确认当前持久化出口已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestPublishPersist_CurrentSeams -count=1
```

Expected:
- PASS

- [ ] **Step 3: 再补 persist helper test，让它先失败**

在 `publish_persist_helpers_test.go` 增加：

```go
func TestPersistPublishedOperations(t *testing.T) {
	t.Parallel()

	var (
		calledCreate bool
		calledStatus bool
	)
	patches := gomonkey.ApplyFunc(
		batchCreateEtcdResource,
		func(ctx context.Context, ops []publisher.ResourceOperation) error {
			calledCreate = true
			return nil
		},
	)
	defer patches.Reset()
	patches.ApplyFunc(
		BatchUpdateResourceStatus,
		func(
			ctx context.Context,
			resourceType constant.APISIXResource,
			resourceIDs []string,
			status constant.ResourceStatus,
		) error {
			calledStatus = true
			return nil
		},
	)

	ctx := context.Background()
	ops := []publisher.ResourceOperation{
		{
			Type:   constant.PluginConfig,
			Key:    "pc-id",
			Config: json.RawMessage(`{"id":"pc-id","plugins":{}}`),
		},
	}

	err := persistPublishedOperations(
		ctx,
		constant.PluginConfig,
		[]string{"pc-id"},
		ops,
		"插件组发布错误",
	)
	assert.NoError(t, err)
	assert.True(t, calledCreate)
	assert.True(t, calledStatus)
}
```

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run TestPersistPublishedOperations -count=1
```

Expected:
- FAIL，报 `undefined: persistPublishedOperations`

- [ ] **Step 4: 用最小实现抽出 persist helper，并迁移所有 put 路径**

在 `publish_persist_helpers.go` 新增：

```go
package biz

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/publisher"
)

func persistPublishedOperations(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceIDs []string,
	ops []publisher.ResourceOperation,
	errMessage string,
) error {
	if err := batchCreateEtcdResource(ctx, ops); err != nil {
		return err
	}
	if err := BatchUpdateResourceStatus(ctx, resourceType, resourceIDs, constant.ResourceStatusSuccess); err != nil {
		logging.ErrorFWithContext(ctx, "%s status change err: %s", resourceType, err.Error())
		return fmt.Errorf("%s：%w", errMessage, err)
	}
	return nil
}
```

然后把所有 `putXxx()/PutXxx()` 里重复的：

```go
err = batchCreateEtcdResource(ctx, ops)
if err != nil {
	return err
}
if err = BatchUpdateResourceStatus(...); err != nil {
	...
}
```

替换成：

```go
return persistPublishedOperations(ctx, constant.PluginConfig, pluginConfigIDs, pluginConfigOps, "插件组发布错误")
```

- [ ] **Step 5: 运行任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/biz -run 'TestPublishPersist_CurrentSeams|TestPublishDependencies_CurrentSeams|TestSimplePublishPayload_CurrentSeams|TestPersistPublishedOperations' -count=1
```

Expected:
- PASS

- [ ] **Step 6: 提交这个 PR**

```bash
git add src/apiserver/pkg/biz/publish.go src/apiserver/pkg/biz/publish_test.go src/apiserver/pkg/biz/publish_persist_helpers.go src/apiserver/pkg/biz/publish_persist_helpers_test.go
git commit -m "refactor: extract publish persist helper"
```

### Task 5: 抽 `EtcdPublisher` 最终校验 helper

- [ ] Task 5: 抽 `EtcdPublisher` 最终校验 helper

**要解决的复杂度：** `EtcdPublisher.Validate()` 现在把“版本解析”“custom plugin schema 获取”“构造 `ETCD` validator”“执行最终校验”都压在一个方法里，而 `Create/Update/BatchCreate/BatchUpdate` 又各自直接循环调用它。当前测试主要 patch 掉 `Validate()`，没有真正锁住最终发布校验的组装契约。

**为什么这个任务适合单独提 PR：** 只处理 publish 的最终校验门，不碰 `pkg/biz/publish.go` 的 payload 成形。

**Files:**
- Modify: `src/apiserver/pkg/publisher/etcd.go`
- Modify: `src/apiserver/pkg/publisher/etcd_test.go`

- [ ] **Step 1: 先补现有 `EtcdPublisher` seam 的 characterization tests**

在 `etcd_test.go` 增加一个直接锁 `Validate()` 组装行为的 `Describe`：

```go
Describe("Validate", func() {
	It("uses ETCD profile validator with gateway version and customize schema map", func() {
		mockEtcdStore := mock.NewMockStorageInterface(ctrl)
		p := &EtcdPublisher{
			ctx:       context.Background(),
			etcdStore: mockEtcdStore,
			gatewayInfo: &model.Gateway{
				ID:            12,
				APISIXVersion: "3.13.0",
			},
		}

		validatorCalled := false
		patches = gomonkey.ApplyFunc(
			GetCustomizePluginSchemaMap,
			func(ctx context.Context, gatewayID int) map[string]any {
				assert.Equal(GinkgoT(), 12, gatewayID)
				return map[string]any{"demo-plugin": map[string]any{}}
			},
		)
		patches.ApplyFunc(
			schema.NewAPISIXJsonSchemaValidator,
			func(
				version constant.APISIXVersion,
				resourceType constant.APISIXResource,
				jsonPath string,
				customizePluginSchemaMap map[string]any,
				dataType constant.DataType,
			) (schema.Validator, error) {
				validatorCalled = true
				assert.Equal(GinkgoT(), constant.APISIXVersion313, version)
				assert.Equal(GinkgoT(), constant.PluginConfig, resourceType)
				assert.Equal(GinkgoT(), "main.plugin_config", jsonPath)
				assert.Equal(GinkgoT(), constant.ETCD, dataType)
				return validatorStub{err: nil}, nil
			},
		)

		err := p.Validate(constant.PluginConfig, json.RawMessage(`{"id":"pc-id","plugins":{}}`))
		assert.NoError(GinkgoT(), err)
		assert.True(GinkgoT(), validatorCalled)
	})
})
```

并在同文件增加最小 stub：

```go
type validatorStub struct {
	err error
}

func (s validatorStub) Validate(_ json.RawMessage) error {
	return s.err
}
```

- [ ] **Step 2: 运行 seam test，确认当前最终校验契约被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/publisher -count=1
```

Expected:
- PASS

- [ ] **Step 3: 再补 helper-level tests，让它先失败**

继续在 `etcd_test.go` 增加：

```go
Describe("validatePublishOperations", func() {
	It("stops on the first invalid operation", func() {
		mockEtcdStore := mock.NewMockStorageInterface(ctrl)
		p := &EtcdPublisher{
			ctx:       context.Background(),
			etcdStore: mockEtcdStore,
			gatewayInfo: &model.Gateway{
				ID:            12,
				APISIXVersion: "3.11.0",
			},
		}

		patches = gomonkey.ApplyMethod(
			reflect.TypeOf(p),
			"Validate",
			func(_ *EtcdPublisher, resourceType constant.APISIXResource, config json.RawMessage) error {
				if resourceType == constant.PluginConfig {
					return errors.New("invalid plugin config")
				}
				return nil
			},
		)

		err := p.validatePublishOperations([]ResourceOperation{
			{Type: constant.PluginConfig, Key: "pc-id", Config: json.RawMessage(`{"plugins":{}}`)},
			{Type: constant.GlobalRule, Key: "gr-id", Config: json.RawMessage(`{"plugins":{}}`)},
		})
		assert.Error(GinkgoT(), err)
		assert.Equal(GinkgoT(), "invalid plugin config", err.Error())
	})
})
```

Expected:
- FAIL，报 `p.validatePublishOperations undefined`

- [ ] **Step 4: 用最小实现抽出最终校验 helper，并接回 `Create/Update/BatchCreate/BatchUpdate`**

**短路语义变化注释（review）：** `BatchCreate/BatchUpdate` 原行为是“逐条校验+组装 resourcesMap，某条校验失败时前面的 resourcesMap 已部分组装完成（但没写入 etcd）”；抽出 helper 后变为“先全部校验，校验失败直接 return，resourcesMap 根本不会开始组装”。**对外部 side effect 是零差异**（两种情况下失败时都未写入 etcd），但在 Step 4 的改动注释 / commit message 里要标注这一 short-circuit 顺序变化。

在 `etcd.go` 中补两个本地 helper：

```go
func (s *EtcdPublisher) buildETCDValidator(resourceType constant.APISIXResource) (schema.Validator, error) {
	apisixVersion, _ := version.ToXVersion(s.gatewayInfo.APISIXVersion)
	customizePluginSchemaMap := GetCustomizePluginSchemaMap(s.ctx, s.gatewayInfo.ID)
	return schema.NewAPISIXJsonSchemaValidator(
		apisixVersion,
		resourceType,
		"main."+string(resourceType),
		customizePluginSchemaMap,
		constant.ETCD,
	)
}

func (s *EtcdPublisher) validatePublishOperations(resources []ResourceOperation) error {
	for _, resource := range resources {
		if err := s.Validate(resource.Type, resource.Config); err != nil {
			return err
		}
	}
	return nil
}
```

并让：

- `Validate(...)` 只负责 `buildETCDValidator(...) + validator.Validate(...)`
- `BatchCreate(...)` 和 `BatchUpdate(...)` 改成先调 `validatePublishOperations(...)`，再组装 `resourcesMap`

这样 `Create(...)` / `Update(...)` 仍保留原来的单资源出口，但 batch 路径不再内联一份重复循环。

- [ ] **Step 5: 运行 publisher 任务相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/publisher -count=1
```

Expected:
- PASS

- [ ] **Step 6: 提交这个 PR**

```bash
git add src/apiserver/pkg/publisher/etcd.go src/apiserver/pkg/publisher/etcd_test.go
git commit -m "refactor: extract publish validation helpers"
```

## 建议执行顺序

1. Task 1：先收拢字段清理，锁住最容易改漏的版本差异和内部字段清理
2. Task 2：再收拢 simple 资源 payload builder，让一批无依赖资源先变短
3. Task 3：之后拆依赖发布 helper，专心处理 route/service/upstream/consumer/stream_route
4. Task 4：在 payload 和依赖 helper 都到位后，再收口统一 persist 逻辑
5. Task 5：最后处理 `EtcdPublisher` 最终校验 helper，给发布链路末端收边

## 完成标准

- `pkg/biz/publish_test.go` 不再只验证“发布成功”，而是明确覆盖：
  - publish 后最终 `GatewaySyncData.Config` 的关键字段形态
  - 版本差异字段是否保留/删除
  - 依赖资源是否随主资源一起发布
- `pkg/publisher/etcd_test.go` 明确覆盖：
  - `Validate()` 使用 `ETCD` profile
  - 最终校验按网关版本构造 validator
  - batch 路径逐条做最终发布校验
- `pkg/biz/publish.go` 中：
  - 字段清理逻辑从 `putXxx()` 中退出
  - simple payload builder 从 `putXxx()` 中退出
  - 依赖收集逻辑从 `putXxx()` 中退出
  - persist 逻辑从 `putXxx()` 中退出
- `pkg/publisher/etcd.go` 中：
  - validator 构造和 batch 校验循环各自有清晰 helper

## 为什么这份计划符合当前共识

- 它只处理 publish 自己的复杂度，不要求与 `web/open api/import/mcp` 抽象对齐
- 它把 `HandleConfig()` 当成稳定边界，不改各资源本地持久化回写规则
- 它每一步都是小 PR，可以逐步验收、逐步合并
- 它优先用现有 publish seam 锁住行为，再做 helper 拆分，符合“先补现状测试，再在测试保护下重构”的原则
