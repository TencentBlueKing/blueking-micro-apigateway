# Import Config 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在保留 `import.ignore_fields` 本地语义不变的前提下，把 import 链路里当前混在 `handleResources(...)` 和 `HandleUploadResources(...)` 里的 overlay、旧资源装载、sync-data 组装、校验前准备几个步骤拆开，使 import 的本地复杂度降下来。

**Architecture:** 本计划完全承认 import 是一条独立链路，不把 overlay 硬塞进共享逻辑。顺序固定为：先把 overlay 抽成 import 本地 pure helper，再把“装载旧资源”“组装 `GatewaySyncData`”拆开，随后重写 `handleResources(...)` 为更小的 orchestration，最后给 `HandleUploadResources(...)` 引入显式的 import validation seam。

**Tech Stack:** Go, Gin context helper, `gjson` / `sjson`, `testify`, `go test`, `make lint`, `make test`

---

## 代码复核结论

- 重构目的判断：基本正确，真正复杂度集中在 `handleResources(...)` 同时承担旧资源装载、overlay、`GatewaySyncData` 组装和 `allResourceIdMap` 回填。
- 复杂度评估：整体中等；代码拆分本身不重，主要成本在前置测试，因为当前仓库里几乎没有 import 链路的直接测试。
- 本次修正：把“重构前测试”提升为独立阶段；seam 优先级改为先锁 `HandleUploadResources(...)`，再在同包测试里补 `handleResources(...)`，helper 单测一律后置。

## 执行顺序（修订）

1. Task 0：独立补 import characterization tests。
2. Task 1-3：拆纯 helper（overlay / 旧资源装载 / sync-data 组装）。
3. Task 4：仅做 orchestration 重排。
4. Task 5：最后显式化 validation seam。

## 范围

- 只处理 `src/apiserver/pkg/apis/common/resource_slz.go`
- 允许把 import 本地 helper 拆到新文件
- 保持 `ignore_fields` 仍然是 import 本地能力

## 非目标

- 不把 import overlay 抽成跨领域公共代码
- 不改 `biz.BuildConfigRawForValidation(...)`
- 不改 `HandleConfig()` 行为
- 不改 open / web / mcp

## 人工 Review 补充（边界保护）

- **文件上传导入链路的 ID 语义必须单独保留：** import 不是普通 Web create。当前链路以导入数据里的 `resource_id` / `ResourceInfo.ResourceID` 为准，`HandlerResourceIndexMap(...)` 和 `handleResources(...)` 都是“为空直接报错”，不是“本地生成 ID 再继续”。Task 1-5 抽 helper 时，不允许引入 ID fallback、自动补 ID、或把上传值重写成别的 ID；如果某类资源同时在 `config.id` 内也带身份字段，测试必须先锁住“上传 ID 原样沿用”的现状。
- **批量导入的事务边界必须单独保留：** import prepare/validate 在 `pkg/apis/common/resource_slz.go`，真正落库在 `biz.UploadResources(...)`，后者已经自己包事务并串联 delete / insert / schema update。这个计划不重排、不下沉这段事务编排；如果执行中发现需要改 import/batch 的事务边界或引入新事务层，视为超出本计划，单独开任务。

## 文件结构

- `src/apiserver/pkg/apis/common/resource_slz.go`
  - 保留 import 主 orchestration
- `src/apiserver/pkg/apis/common/resource_slz_import_test.go`
  - import seam-first characterization tests，优先覆盖 `HandleUploadResources(...)`
- `src/apiserver/pkg/apis/common/import_resource_helpers.go`
  - import 本地 helper：overlay、旧资源装载、sync-data 组装、validation input 准备
- `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
  - import 本地 helper 的 TDD 测试

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
  - 第一层：先锁定 `HandleUploadResources(...)` 这条真实 import 入口的行为
  - 第二层：helper 抽出后再补 helper 单测
- Import 计划里的现有 seam 优先级如下：
  - Task 0：优先测现有 `HandleUploadResources(...)`，覆盖 add/update 分流、`ignore_fields` overlay、空 `resource_id` 失败、关联资源校验失败
  - Task 1-4：在 Task 0 已覆盖入口行为后，再在同包测试里补 `handleResources(...)`，把重构面缩小到本地 orchestration
  - Task 5：继续以 `HandleUploadResources(...)` 为主，补“prepare 完成后进入 `biz.ValidateResource(...)` 之前”的边界断言
- `HandlerResourceIndexMap(...)` 可以作为辅助 characterization seam，用来锁 `existsResourceIdList` / `allResourceIdList` 组装，但不能代替 `HandleUploadResources(...)`。
- 只有当第一层 seam 测试已经锁住行为时，才允许在同一个 PR 里为 `apply...` / `load...` / `build...` / `prepare...` helper 增加第二层单测。
- 执行时，如果任务正文里的示例代码先写了 helper 测试，应按上面的 seam 规则落地：先补现有 seam 的 characterization test，再补 helper test。

## 重构前测试前置阶段（独立）

- Task 0 的目标不是引入 helper，而是先把现状锁住；建议单独落一组 characterization tests 到 `resource_slz_import_test.go`。
- Task 0 至少覆盖 6 类现状：`ignore_fields` 会从旧资源覆盖导入配置；**旧资源上没有该字段时，导入配置原样保留（overlay 仅在 `gjson.GetBytes(existing, skipRule).Exists()==true` 时生效）**；**上传数据里已有 `resource_id` 的资源会原样沿用该 ID，不会在 import prepare 阶段被改写或重生成**；空 `resource_id` 会直接失败；缺失关联资源会在 `HandleUploadResources(...)` 阶段报错；add/update map 的资源数与输入一致。
- Task 0 完成后，后续 Task 1-5 才允许把断言下沉到 `handleResources(...)` 或新 helper。

**命名一致性约束（review 补充）：**

- 落地时统一使用 `allResourceIDs` 作为 helper 导出/入参命名（与 Go 合语一致）；当前 `resource_slz.go` 内部仍叫 `allResourceIdMap` 时作为局部变量保留、不零散伸到 helper 签名。
- 所有 helper 参数顺序遵循 `(ctx, resourceType, ...)` 或 `(ctx, ...)` 的 context-first 风格；orchestration helper `prepareImportResources(ctx, resourcesImport, allResourceIDs, ignoreFields)` 作为唯一例外（多 map 输入），在 doc comment 里用注释解释这一点。
- 所有会对入参 map 做 in-place 写入的 helper（Task 2 `loadExistingImportResources` / Task 4 `prepareImportResources` / Task 5 `prepareImportValidationInput`）都必须在头部 doc comment 写明 `// mutates allResourceIDs: appends every encountered resource key`，避免后来维护者误以为是纯函数。

### Task 0: 补 import 入口 characterization tests

- [ ] Task 0: 补 import 入口 characterization tests

**要解决的缺口：** 当前文档已经要求 seam-first，但正文还没有独立任务把 `HandleUploadResources(...)` 的现状锁住。先把入口行为补成单独任务，后面的 helper 提取才不是“边改边猜”。

**为什么这个任务适合单独提 PR：** 只新增 import 入口测试，不修改 `resource_slz.go` 的生产逻辑。

**Files:**
- Create: `src/apiserver/pkg/apis/common/resource_slz_import_test.go`

- [ ] **Step 1: 在 `HandleUploadResources(...)` 上补一组入口 characterization tests**

至少覆盖下面 6 类现状，断言都落在现有入口返回值和错误上，不提前引入新 helper：

- `ignore_fields` 会用旧资源上的字段覆盖导入配置
- **旧资源上没有 `ignore_fields` 指定的字段时，导入配置原样保留（锁住 `gjson.Exists==false 时静默跳过` 的现状，避免 Task 1 不小心改成 always-overwrite）**
- **上传数据里已有 `resource_id` 的资源在 `HandleUploadResources(...)` 之后仍沿用原 ID，不会被本地生成或替换**
- 空 `resource_id` 会在进入后续入库前直接失败
- 缺失关联资源会在 `HandleUploadResources(...)` 阶段报错
- add/update map 的资源数与输入一致；必要时可用 `HandlerResourceIndexMap(...)` 辅助准备断言输入

- [ ] **Step 2: 运行 import seam tests，确认当前入口行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -run 'TestHandleUploadResources|TestHandlerResourceIndexMap' -count=1
```

Expected:
- PASS

- [ ] **Step 3: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/common/resource_slz_import_test.go
git commit -m "test: lock import upload characterization seams"
```

---

### Task 1: 抽出 import 本地 `ignore_fields` overlay helper

- [ ] Task 1: 抽出 import 本地 `ignore_fields` overlay helper

**要解决的复杂度：** overlay 逻辑现在埋在 `handleResources(...)` 的双层循环里，后面想看“导入为什么被旧字段覆盖了”必须先通读整个 import 主流程。

**为什么这个任务适合单独提 PR：** 只处理 import 特有能力，不涉及 `biz.ValidateResource(...)` 和后续入库。

**Files:**
- Create: `src/apiserver/pkg/apis/common/import_resource_helpers.go`
- Create: `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/common/resource_slz.go:281-297`

- [ ] **Step 1: 先补 overlay 当前行为的失败测试**

在 `import_resource_helpers_test.go` 里新增：

```go
func TestApplyImportIgnoreFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		imported     string
		existing     string
		ignoreFields []string
		want         string
	}{
		{
			name:         "overlay top level field from existing config",
			imported:     `{"name":"route-a","desc":"new-desc","plugins":{}}`,
			existing:     `{"name":"route-a","desc":"old-desc","plugins":{"limit-count":{"count":1}}}`,
			ignoreFields: []string{"desc"},
			want:         `{"name":"route-a","desc":"old-desc","plugins":{}}`,
		},
		{
			name:         "overlay nested field from existing config",
			imported:     `{"plugins":{"limit-count":{"count":10,"time_window":60}}}`,
			existing:     `{"plugins":{"limit-count":{"count":1,"time_window":120}}}`,
			ignoreFields: []string{"plugins.limit-count.count"},
			want:         `{"plugins":{"limit-count":{"count":1,"time_window":60}}}`,
		},
		{
			name:         "ignore missing field keeps imported config",
			imported:     `{"plugins":{}}`,
			existing:     `{"name":"route-a"}`,
			ignoreFields: []string{"plugins.limit-count"},
			want:         `{"plugins":{}}`,
		},
		// review 补充：锁住“多个 ignoreFields 混合，其中部分字段在旧资源不存在”的行为
		{
			name:         "partial missing fields - only present fields are overlaid",
			imported:     `{"desc":"new-desc","plugins":{"limit-count":{"count":10}}}`,
			existing:     `{"desc":"old-desc"}`,
			ignoreFields: []string{"desc", "plugins.limit-count.count"},
			want:         `{"desc":"old-desc","plugins":{"limit-count":{"count":10}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyImportIgnoreFields(
				json.RawMessage(tt.imported),
				datatypes.JSON([]byte(tt.existing)),
				tt.ignoreFields,
			)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -run TestApplyImportIgnoreFields -count=1
```

Expected:
- FAIL，报 `undefined: applyImportIgnoreFields`

- [ ] **Step 3: 实现 overlay helper，并替换 `handleResources(...)` 内联逻辑**

在 `import_resource_helpers.go` 里新增：

```go
func applyImportIgnoreFields(
	importedConfig json.RawMessage,
	existingConfig datatypes.JSON,
	ignoreFields []string,
) (json.RawMessage, error) {
	merged := append(json.RawMessage(nil), importedConfig...)
	for _, field := range ignoreFields {
		result := gjson.GetBytes(existingConfig, field)
		if !result.Exists() {
			continue
		}
		var err error
		merged, err = sjson.SetBytes(merged, field, json.RawMessage(result.Raw))
		if err != nil {
			return nil, err
		}
	}
	return merged, nil
}
```

然后把 `handleResources(...)` 里原来的内联 overlay 替换成：

```go
if len(ignoreFields[resourceType]) > 0 && ok {
	imp.Config, err = applyImportIgnoreFields(
		imp.Config,
		oldResource.Config,
		ignoreFields[resourceType],
	)
	if err != nil {
		return nil, fmt.Errorf("set config failed, err: %w", err)
	}
}
```

- [ ] **Step 4: 运行 common 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/common/import_resource_helpers.go src/apiserver/pkg/apis/common/import_resource_helpers_test.go src/apiserver/pkg/apis/common/resource_slz.go
git commit -m "refactor: extract import ignore-fields overlay helper"
```

### Task 2: 抽出 import 本地“装载旧资源” helper

- [ ] Task 2: 抽出 import 本地“装载旧资源” helper

**要解决的复杂度：** `handleResources(...)` 每轮循环都要自己取 DB 资源、组 map、回填 `allResourceIDs`，这块和 overlay / sync-data 组装混在一起，不利于单测。

**为什么这个任务适合单独提 PR：** 这一步只把 DB 读取和 map 组装从大函数里抽出来，不调整 overlay 语义。

**Files:**
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers.go`
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/common/resource_slz.go:267-275`

- [ ] **Step 1: 先补旧资源装载测试**

在 `import_resource_helpers_test.go` 里新增（除 happy path 外，review 要求补一条 empty-DB 边界断言）：

```go
func TestLoadExistingImportResources(t *testing.T) {
	util.InitEmbedDb()

	ctx := context.Background()
	gateway := &model.Gateway{Name: "import-test-gateway", APISIXVersion: string(constant.APISIXVersion313)}
	assert.NoError(t, biz.CreateGateway(ctx, gateway))

	gatewayCtx := ginx.SetGatewayInfoToContext(ctx, gateway)
	assert.NoError(t, biz.CreatePluginConfig(gatewayCtx, &model.PluginConfig{
		Name: "pc-demo",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "pc-1",
			GatewayID: gateway.ID,
			Config:    datatypes.JSON([]byte(`{"id":"pc-1","name":"pc-demo","plugins":{}}`)),
			Status:    constant.ResourceStatusSuccess,
		},
	}))

	allResourceIDs := map[string]struct{}{}
	got, err := loadExistingImportResources(gatewayCtx, constant.PluginConfig, allResourceIDs)
	assert.NoError(t, err)
	assert.Contains(t, got, fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-1"))
	assert.Contains(t, allResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-1"))

	// review 补充：DB 中没有该类型资源时 helper 返回空 map + nil error
	t.Run("empty DB returns empty map", func(t *testing.T) {
		empty := map[string]struct{}{}
		got, err := loadExistingImportResources(gatewayCtx, constant.Upstream, empty)
		assert.NoError(t, err)
		assert.Empty(t, got)
		assert.Empty(t, empty)
	})
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -run TestLoadExistingImportResources -count=1
```

Expected:
- FAIL，报 `undefined: loadExistingImportResources`

- [ ] **Step 3: 把“取 DB 资源 + 组 map + 回填 allResourceIDs”抽成 helper**

在 `import_resource_helpers.go` 里新增（**注意 doc comment 显式标注 side-effect**）：

```go
// loadExistingImportResources fetches all stored resources of the given type for
// the current gateway and returns them keyed by GetResourceKey(resourceType, id).
//
// mutates allResourceIDs: every loaded resource key is appended to the passed-in
// set so that the import orchestration can build a global id set across all
// resource types without re-querying the DB.
func loadExistingImportResources(
	ctx context.Context,
	resourceType constant.APISIXResource,
	allResourceIDs map[string]struct{},
) (map[string]model.ResourceCommonModel, error) {
	allResourceList, err := biz.GetResourceByIDs(ctx, resourceType, []string{})
	if err != nil {
		return nil, fmt.Errorf("get exist resources failed, err: %w", err)
	}

	allResourceMap := make(map[string]model.ResourceCommonModel, len(allResourceList))
	for _, resource := range allResourceList {
		key := resource.GetResourceKey(resourceType)
		allResourceMap[key] = resource
		allResourceIDs[key] = struct{}{}
	}
	return allResourceMap, nil
}
```

然后把 `handleResources(...)` 里原来的 9 行 DB 装载逻辑替换为这个 helper 调用。本地变量仍用 `allResourceIdMap` 名字以保持结果最小改动，helper 签名则统一走 `allResourceIDs`。

- [ ] **Step 4: 运行 common 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/common/import_resource_helpers.go src/apiserver/pkg/apis/common/import_resource_helpers_test.go src/apiserver/pkg/apis/common/resource_slz.go
git commit -m "refactor: extract import existing-resource loader"
```

### Task 3: 抽出 import 本地 `GatewaySyncData` 组装 helper

- [ ] Task 3: 抽出 import 本地 `GatewaySyncData` 组装 helper

**要解决的复杂度：** `GatewaySyncData` 组装现在直接夹在 `handleResources(...)` 末尾，和 resource_id 校验、overlay、map append 混在一个循环里。

**为什么这个任务适合单独提 PR：** 这是 import 本地纯组装 helper，不会改变 validate 或 upload 流程边界。

**Files:**
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers.go`
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/common/resource_slz.go:298-309`

- [ ] **Step 1: 先补 sync-data 组装测试**

在 `import_resource_helpers_test.go` 里新增：

```go
func TestBuildImportSyncData(t *testing.T) {
	t.Parallel()

	ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ID: 23})
	info := &ResourceInfo{
		ResourceType: constant.Route,
		ResourceID:   "route-1",
		Name:         "route-demo",
		Config:       json.RawMessage(`{"id":"route-1","name":"route-demo","uri":"/demo"}`),
	}

	got := buildImportSyncData(ctx, constant.Route, info)
	assert.Equal(t, constant.Route, got.Type)
	assert.Equal(t, "route-1", got.ID)
	assert.Equal(t, 23, got.GatewayID)
	assert.JSONEq(t, `{"id":"route-1","name":"route-demo","uri":"/demo"}`, string(got.Config))
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -run TestBuildImportSyncData -count=1
```

Expected:
- FAIL，报 `undefined: buildImportSyncData`

- [ ] **Step 3: 实现 sync-data helper，并替换 `handleResources(...)` 里的内联组装**

在 `import_resource_helpers.go` 里新增：

```go
func buildImportSyncData(
	ctx context.Context,
	resourceType constant.APISIXResource,
	imp *ResourceInfo,
) *model.GatewaySyncData {
	return &model.GatewaySyncData{
		Type:      resourceType,
		ID:        imp.ResourceID,
		Config:    datatypes.JSON(imp.Config),
		GatewayID: ginx.GetGatewayInfoFromContext(ctx).ID,
	}
}
```

然后把 `handleResources(...)` 里的：

```go
resourceImp := &model.GatewaySyncData{...}
```

替换成：

```go
resourceImp := buildImportSyncData(ctx, resourceType, imp)
```

**边界提醒（人工 review 补充）：** `buildImportSyncData(...)` 只消费上传进来的 `imp.ResourceID`。不要在这个 helper 里生成、修正、回填资源 ID；一旦这里开始碰 ID 语义，就不再是“sync-data 组装”，而是在偷偷改导入协议。

- [ ] **Step 4: 运行 common 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/common/import_resource_helpers.go src/apiserver/pkg/apis/common/import_resource_helpers_test.go src/apiserver/pkg/apis/common/resource_slz.go
git commit -m "refactor: extract import sync-data builder"
```

### Task 4: 重写 `handleResources(...)` 为 import 本地 orchestration

> **Review 补充价值陈述：** Task 1-3 已经把 `handleResources(...)` 内部 3 个纯 helper 抽出；Task 4 的价值**不是继续降复杂度**，而是让 `HandleUploadResources(...)` 的调用层拿到一个显式的“import 整套准备完成” seam（`prepareImportResources(...)`）：
> - 可以把 DB helper / overlay helper / sync-data helper 的组合用法用一个公共入口保存下来；
> - Task 5 在引入 validation seam 时可以直接复用它而不是再写一套；
> - 将来想做内存级 mock / import dry-run 时，有明确替换点。
>
> 如果不做 Task 4，`handleResources(...)` 仍然是“将 3 个 helper 在 handler 层拼回去”的写法，并非 blocker，但 Task 5 的 validation seam 会因为没有此层 helper 而需要在外层手写两次类似的拼装。
>
> **review 补充测试**：在 `import_resource_helpers_test.go` 的 `TestPrepareImportResources` 中补一条锁定 Schema 跳过语义的 case——`prepareImportResources` 入参同时包含 `constant.Schema` 和 `constant.Route` 时，返回 map 不含 `constant.Schema`。
>
> **review 补充边界**：`prepareImportResources(...)` 虽然汇总 add/update 资源，但仍停在 prepare 阶段；不要把 `biz.UploadResources(...)` 的事务、审计、schema update 顺序搬进来，否则会把“orchestration 重排”变成“跨层事务重构”。

- [ ] Task 4: 重写 `handleResources(...)` 为 import 本地 orchestration

**要解决的复杂度：** 现在 `handleResources(...)` 同时做资源遍历、旧资源装载、overlay、resource_id 校验、sync-data append，是典型的大函数混合职责。

**为什么这个任务适合单独提 PR：** 前三步 helper 都到位后，这一步只做“重排 orchestration”，不会引入新的业务语义。

**Files:**
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers.go`
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/common/resource_slz.go:256-313`

- [ ] **Step 1: 先补大函数当前行为保护测试**

在 `import_resource_helpers_test.go` 里新增：

```go
func TestPrepareImportResources(t *testing.T) {
	util.InitEmbedDb()

	ctx := context.Background()
	gateway := &model.Gateway{Name: "prepare-import-gateway", APISIXVersion: string(constant.APISIXVersion313)}
	assert.NoError(t, biz.CreateGateway(ctx, gateway))
	gatewayCtx := ginx.SetGatewayInfoToContext(ctx, gateway)

	existing := &model.PluginConfig{
		Name: "pc-demo",
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        "pc-1",
			GatewayID: gateway.ID,
			Config:    datatypes.JSON([]byte(`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{}}`)),
			Status:    constant.ResourceStatusSuccess,
		},
	}
	assert.NoError(t, biz.CreatePluginConfig(gatewayCtx, existing))

	resources, err := prepareImportResources(
		gatewayCtx,
		map[constant.APISIXResource][]*ResourceInfo{
			constant.PluginConfig: {
				{
					ResourceType: constant.PluginConfig,
					ResourceID:   "pc-1",
					Name:         "pc-demo",
					Config:       json.RawMessage(`{"id":"pc-1","name":"pc-demo","desc":"new-desc","plugins":{}}`),
				},
			},
		},
		map[string]struct{}{},
		map[constant.APISIXResource][]string{
			constant.PluginConfig: {"desc"},
		},
	)
	assert.NoError(t, err)
	assert.Len(t, resources[constant.PluginConfig], 1)
	assert.JSONEq(t, `{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{}}`, string(resources[constant.PluginConfig][0].Config))
}
```

- [ ] **Step 2: 运行测试，确认新 orchestration helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -run TestPrepareImportResources -count=1
```

Expected:
- FAIL，报 `undefined: prepareImportResources`

- [ ] **Step 3: 实现 orchestration helper，并让 `handleResources(...)` 只做代理**

在 `import_resource_helpers.go` 里新增：

```go
func prepareImportResources(
	ctx context.Context,
	resourcesImport map[constant.APISIXResource][]*ResourceInfo,
	allResourceIDs map[string]struct{},
	ignoreFields map[constant.APISIXResource][]string,
) (map[constant.APISIXResource][]*model.GatewaySyncData, error) {
	resourceTypeMap := make(map[constant.APISIXResource][]*model.GatewaySyncData)
	for resourceType, resourceInfoList := range resourcesImport {
		if resourceType == constant.Schema {
			continue
		}

		existingMap, err := loadExistingImportResources(ctx, resourceType, allResourceIDs)
		if err != nil {
			return nil, err
		}

		for _, imp := range resourceInfoList {
			if imp.ResourceID == "" {
				return nil, fmt.Errorf("%s: resource id is empty: %s", resourceType, imp.Name)
			}
			if oldResource, ok := existingMap[imp.GetResourceKey()]; ok && len(ignoreFields[resourceType]) > 0 {
				imp.Config, err = applyImportIgnoreFields(imp.Config, oldResource.Config, ignoreFields[resourceType])
				if err != nil {
					return nil, fmt.Errorf("set config failed, err: %w", err)
				}
			}

			allResourceIDs[imp.GetResourceKey()] = struct{}{}
			resourceTypeMap[resourceType] = append(
				resourceTypeMap[resourceType],
				buildImportSyncData(ctx, resourceType, imp),
			)
		}
	}
	return resourceTypeMap, nil
}
```

然后把 `resource_slz.go` 中的 `handleResources(...)` 缩成：

```go
func handleResources(...) (map[constant.APISIXResource][]*model.GatewaySyncData, error) {
	return prepareImportResources(ctx, resourcesImport, allResourceIDs, ignoreFields)
}
```

- [ ] **Step 4: 运行 common 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/common/import_resource_helpers.go src/apiserver/pkg/apis/common/import_resource_helpers_test.go src/apiserver/pkg/apis/common/resource_slz.go
git commit -m "refactor: split import resource preparation orchestration"
```

### Task 5: 给 `HandleUploadResources(...)` 引入显式的 import validation seam

- [ ] Task 5: 给 `HandleUploadResources(...)` 引入显式的 import validation seam

**要解决的复杂度：** 现在 `HandleUploadResources(...)` 一边准备 add/update map，一边直接调用 `biz.ValidateResource(...)`，没有一个明确的“import 进入 DATABASE 校验前”的本地边界。

**为什么这个任务适合单独提 PR：** 这一步仍然只在 import 域内新增 seam，不会把逻辑抽到共享层。

**Files:**
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers.go`
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/common/resource_slz.go:121-149`

- [ ] **Step 1: 先补 validation input 组装测试**

在 `import_resource_helpers_test.go` 里新增（**review 补充**：Add 和 Update 同时非空，锁住 `allResourceIDs` 跨两次调用累加语义）：

```go
func TestPrepareImportValidationInput(t *testing.T) {
	t.Parallel()

	ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ID: 31})

	t.Run("add only", func(t *testing.T) {
		input, err := prepareImportValidationInput(
			ctx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-1",
							Name:         "route-demo",
							Config:       json.RawMessage(`{"id":"route-1","name":"route-demo","uri":"/demo"}`),
						},
					},
				},
				Update: map[constant.APISIXResource][]*ResourceInfo{},
			},
			nil,
		)
		assert.NoError(t, err)
		assert.Contains(t, input.AllResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-1"))
		assert.Len(t, input.Add, 1)
		assert.Len(t, input.Add[constant.Route], 1)
		assert.Empty(t, input.Update)
	})

	// review 补充：Add 和 Update 同时非空时，allResourceIDs 上会同时出现 add 和 update 的 key
	t.Run("add and update accumulate all resource ids", func(t *testing.T) {
		input, err := prepareImportValidationInput(
			ctx,
			&ResourceUploadInfo{
				Add: map[constant.APISIXResource][]*ResourceInfo{
					constant.Route: {
						{ResourceType: constant.Route, ResourceID: "route-new", Name: "route-new",
							Config: json.RawMessage(`{"id":"route-new","uri":"/a"}`)},
					},
				},
				Update: map[constant.APISIXResource][]*ResourceInfo{
					constant.Route: {
						{ResourceType: constant.Route, ResourceID: "route-upd", Name: "route-upd",
							Config: json.RawMessage(`{"id":"route-upd","uri":"/b"}`)},
					},
				},
			},
			nil,
		)
		assert.NoError(t, err)
		assert.Contains(t, input.AllResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-new"))
		assert.Contains(t, input.AllResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-upd"))
		assert.Len(t, input.Add[constant.Route], 1)
		assert.Len(t, input.Update[constant.Route], 1)
	})
}
```

- [ ] **Step 2: 运行测试，确认 seam 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -run TestPrepareImportValidationInput -count=1
```

Expected:
- FAIL，报 `undefined: prepareImportValidationInput`

- [ ] **Step 3: 实现 import validation seam，并让 `HandleUploadResources(...)` 改成两段式**

在 `import_resource_helpers.go` 里新增：

```go
type importValidationInput struct {
	Add            map[constant.APISIXResource][]*model.GatewaySyncData
	Update         map[constant.APISIXResource][]*model.GatewaySyncData
	AllResourceIDs map[string]struct{}
}

func prepareImportValidationInput(
	ctx context.Context,
	resourcesImport *ResourceUploadInfo,
	ignoreFields map[constant.APISIXResource][]string,
) (*importValidationInput, error) {
	allResourceIDs := make(map[string]struct{})
	addMap, err := prepareImportResources(ctx, resourcesImport.Add, allResourceIDs, ignoreFields)
	if err != nil {
		return nil, err
	}
	updateMap, err := prepareImportResources(ctx, resourcesImport.Update, allResourceIDs, ignoreFields)
	if err != nil {
		return nil, err
	}
	return &importValidationInput{
		Add:            addMap,
		Update:         updateMap,
		AllResourceIDs: allResourceIDs,
	}, nil
}
```

然后把 `HandleUploadResources(...)` 改成：

```go
validationInput, err := prepareImportValidationInput(ctx, resourcesImport, ignoreFields)
if err != nil {
	return nil, err
}
if err = biz.ValidateResource(ctx, validationInput.Add, validationInput.AllResourceIDs, allSchemaMap); err != nil {
	return nil, fmt.Errorf("add resources validate failed, err: %w", err)
}
if err = biz.ValidateResource(ctx, validationInput.Update, validationInput.AllResourceIDs, allSchemaMap); err != nil {
	return nil, fmt.Errorf("updated resources validate failed, err: %w", err)
}
```

- [ ] **Step 4: 运行 common 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/common/import_resource_helpers.go src/apiserver/pkg/apis/common/import_resource_helpers_test.go src/apiserver/pkg/apis/common/resource_slz.go
git commit -m "refactor: add explicit import validation seam"
```

## 完成定义

- `import.ignore_fields` 保持 import 本地能力，不被抽到共享层
- overlay、旧资源装载、sync-data 组装、validation input 都有明确本地 helper
- `handleResources(...)` 不再承担所有职责
- `HandleUploadResources(...)` 具备显式的 import validation seam
- 上传导入链路保留“用户提供 `resource_id` / ID，空则报错”的现状，不引入本地自动生成 ID
- 没有任何一步改动 `biz.UploadResources(...)` 现有的事务 / delete-insert / schema update 编排
