# Open API Config 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **Execution rule:** If a task or step is done, mark it in this `plan.md` before running `git add` and `git commit`.

**Goal:** 在不改变 Open API 外部协议的前提下，先修正 `update` 路径 outer `name` 未回写到 storage config、以及 middleware 对旧版本 schema 的临时 `id` 注入不够精确这两个已确认问题，再逐步收敛 `open api` 当前分散在 middleware、serializer、handler 三处的 `config` 整形、draft 组装和 request identity 复算问题。

**Architecture:** 本计划先把 Open 域内部已经确认的行为问题修正掉，再做本地收敛。执行顺序调整为：先独立补 handler / serializer / middleware characterization tests；然后优先修正 update draft 组装与 outer `name` 注入，再整理 batch create builder，再抽 middleware 的 validation payload helper。`resolved draft` context 只在 Task 1-3 完成后仍然存在 create 校验/落库 identity 不一致时才引入，不作为默认必做项。

**Tech Stack:** Go, Gin, `gjson` / `sjson`, `testify`, `go test`, `make lint`, `make test`

---

## 代码复核结论

- 重构目的判断：整体方向正确，但这里不只是“降低复杂度”。代码复核已经确认两个现状问题：`ResourceUpdateRequest.ToCommonResource(...)` 不消费 outer `Name`，以及 `OpenAPIResourceCheck()` 目前使用了非 version-aware 的 `ResourceRequiresIDInSchema(...)`。
- 复杂度评估：整体偏高；不是因为 helper 很难抽，而是因为当前仓库里几乎没有 Open handler / middleware 的直接测试，前置 characterization test 成本不低。
- 本次修正：把测试前置阶段独立出来；执行顺序改为先修正 update 和 middleware 的正确性问题，再决定是否需要 `resolved draft` carrier。

## 执行顺序（修订）

1. Task 0：独立补 Open characterization tests（包括显式标注 bug vs 锁现状）。
2. Task 2a：先用 failing-then-green 方式修正 update 路径 outer `name` 不回写 `config` 的 correctness bug。
3. Task 2b：再抽 Open update draft builder（纯 refactor，不改行为）。
4. Task 1：整理 batch create builder。
5. Task 3a：修正 middleware 当前 `ResourceRequiresIDInSchema` 为 version-aware `ResourceRequiresIDInSchemaForVersion`（correctness fix）。
6. Task 3b：抽 middleware validation payload helper（纯 refactor）。
7. Task 4-5：**仅当 Task 0 或 Task 1-3 的 characterization test 既实 “校验 id ≠ 落库 id” 这一不一致时**才执行。

## 范围

- 只处理 `src/apiserver/pkg/apis/open/...` 和 `src/apiserver/pkg/middleware/openapi_resource_check.go`
- 可以新增 Open 域内本地 helper 文件与上下文传递结构
- 允许调整 `ResourceBatchCreateRequest.ToCommonResource(...)` 的本地实现和签名

## 非目标

- 不改 Open API HTTP 协议字段
- 不抽跨领域共享 builder
- 不改 import overlay 逻辑
- 不改 `HandleConfig()` 边界

## 文件结构

- `src/apiserver/pkg/apis/open/serializer/resource.go`
  - Open create / update 的 `ResourceCommonModel` 组装逻辑
- `src/apiserver/pkg/apis/open/handler/resource_test.go`
  - Open handler characterization tests，先覆盖 batch create / update 的现状
- `src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers_test.go`
  - create / update builder 的行为测试
- `src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers.go`
  - Open create / update draft builder 与本地 helper
- `src/apiserver/pkg/apis/open/serializer/open_resolved_draft_context.go`
  - Open 域内 resolved draft 结构和 Gin context helper
- `src/apiserver/pkg/apis/open/serializer/open_resolved_draft_context_test.go`
  - resolved draft context helper 测试
- `src/apiserver/pkg/apis/open/handler/resource.go`
  - Open handler 调用 serializer 的位置
- `src/apiserver/pkg/middleware/openapi_resource_check.go`
  - Open middleware 的 request-side validation payload 整形
- `src/apiserver/pkg/middleware/openapi_resource_check_test.go`
  - middleware payload helper 的矩阵测试

## PR 出口要求

- 每个任务里的 `go test` 是最小验收命令
- 每个任务准备合并前，再补跑一次：

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && make lint && make test
```

## 测试策略（必须）

- 新增 `Task 0` 作为独立步骤或独立 PR；在 `Task 0` 合并前，不开始 Task 1-5。
- 每个任务的第一组测试，必须先打在“重构前已经存在的 seam”上，不能直接从计划中新引入的 helper 或 context carrier 开始写测试。
- helper / context helper 测试只能作为第二层测试：
  - 第一层：锁定 Open 现有请求路径、middleware 路径、serializer 路径
  - 第二层：在 helper 抽出后补 helper 单测
- 当前仓库里没有 Open handler / middleware 的测试文件；Task 0 需要先把 `resource_test.go` 和 `openapi_resource_check_test.go` 补出来，而不是直接从 helper test 开始。
- Open 计划里的现有 seam 优先级如下：
  - Task 0：优先测现有 `ResourceBatchCreate(...)`、`ResourceUpdate(...)`、`OpenAPIResourceCheck()`，先锁真实请求路径
  - Task 1：优先测现有 `ResourceBatchCreateRequest.ToCommonResource(...)`
  - Task 2：优先测现有 `ResourceUpdateRequest.ToCommonResource(...)`
  - Task 3：优先测现有 `OpenAPIResourceCheck()` middleware 行为
  - Task 4-5：只有在决定继续做 context carrier 时，才优先测现有 `ResourceBatchCreate` handler 与 middleware 串起来的路径，确认校验和持久化使用的是同一份 request identity
- 执行时，如果任务正文里的示例代码先写了 helper 测试，应按上面的 seam 规则落地：先补现有 seam 的 characterization test，再补 helper test。

## 重构前测试前置阶段（独立）

- Task 0 至少覆盖 5 类现状，**每条显式标注性质**：
  - 【锁现状】batch create 会按资源类型补 `name/username` 并生成 `id`
  - 【**failing-then-green 条款 ／ 表达期望行为**】update 传入 `name=foo` 时最终落库 config 应该含 `name=foo`（用来捕捉当前 `ResourceUpdateRequest.ToCommonResource(...)` 不回写的 bug，Task 2a 负责修绿）
  - 【锁现状】middleware 在 3.11 / 3.13 下构造出的 validation payload 形态
  - 【**failing-then-green 条款**】middleware 针对旧版本 schema （3.2.15 等）的临时 `id` 注入行为需要满足“ConsumerGroup on 3.2.15 → id 不应出现”（Task 3a 负责修绿）
  - 【锁现状】handler batch create 的 id 与 middleware 校验所用 id 是否一致：直接拿两者的实际值做断言；如果 Task 0 这条 test 证明两者一致，Task 4-5 可以直接 defer / drop；如果不一致，Task 4-5 有明确的落地价值

**明确区分 “锁现状” vs “固定期望行为”（review 必改）：** 每一条 Task 0 assertion 在 PR 描述中都要记一笔它属于哪一类，避免将未来 Task 2a / Task 3a 需要修正的 bug “锁”成现状。

- Task 0 完成前，不要引入 `OpenResolvedDraft` 这类 carrier；否则很难分清是在修真实问题，还是在给未锁住的行为加结构。
- 如果 Task 0 跑完后发现仅通过 update normalization（Task 2a）和 version-aware validation（Task 3a）就能消掉主要重复，Task 4-5 可以直接 defer。

### Task 0: 补 Open handler / middleware characterization tests

- [ ] Task 0: 补 Open handler / middleware characterization tests

**要解决的缺口：** 当前文档把 Task 0 写进了执行顺序和策略，但正文还没有单独任务去锁 `ResourceBatchCreate(...)`、`ResourceUpdate(...)` 和 `OpenAPIResourceCheck()` 的现状。先把真实请求路径测起来，后面再抽 builder / middleware helper。

**为什么这个任务适合单独提 PR：** 只新增 Open handler 与 middleware 的 characterization tests，不改 serializer、middleware 和 handler 生产逻辑。

**Files:**
- Create: `src/apiserver/pkg/apis/open/handler/resource_test.go`
- Create: `src/apiserver/pkg/middleware/openapi_resource_check_test.go`

- [ ] **Step 1: 在真实请求路径上补 Open characterization tests**

至少覆盖下面 5 类现状，全部走现有 handler / middleware seam，**每条标注性质（锁现状 ／ failing-then-green）**：

- 【锁现状】batch create 会按资源类型补 `name/username`，并在缺失时生成 `id`
- 【**failing-then-green**】update 传入 `name=foo` + `config` 无 `name` 时，落库 `config` **应该**含 `name=foo`；此 test 当前会失败（bug），Task 2a make it green
- 【锁现状】middleware 在 3.11 / 3.13 下构造出的 validation payload 形态
- 【**failing-then-green**】`resource=ConsumerGroup` + `version=3.2.15` 时，`validationRaw` **不应**出现 `id` 字段；此 test 当前会因为不看 version 而失败，Task 3a make it green
- 【锁现状，但同时作为 Task 4-5 是否落地的判定依据】batch create 完整走一遍 handler + middleware 后，分别拿 middleware 计算出的 `id` 与 serializer 最终落库的 `id` 做对比断言；若两者一致则 Task 4-5 可 defer，若不一致则 Task 4-5 需落地

- [ ] **Step 2: 运行 Open seam tests，确认入口行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/handler ./pkg/middleware -count=1
```

Expected:
- PASS

- [ ] **Step 3: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/open/handler/resource_test.go src/apiserver/pkg/middleware/openapi_resource_check_test.go
git commit -m "test: lock open handler and middleware seams"
```

---

### Task 1: 抽出 Open batch create 的本地 draft builder

- [ ] Task 1: 抽出 Open batch create 的本地 draft builder

**要解决的复杂度：** `ResourceBatchCreateRequest.ToCommonResource(...)` 把“补 name”“生成 id”“组装 `ResourceCommonModel`”揉在一个循环里，后续只要多一种 create 变体就容易继续复制这段逻辑。

**为什么这个任务适合单独提 PR：** 只改 `serializer/resource.go` 和测试文件，不动 middleware 和 handler。

**Files:**
- Create: `src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers.go`
- Create: `src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/open/serializer/resource.go:57-82`

- [ ] **Step 1: 先补 batch create 当前组装逻辑的失败测试**

在 `open_resource_draft_helpers_test.go` 里新增：

```go
func TestBuildOpenCreateDraft(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		req          ResourceCreateRequest
		assertDraft  func(t *testing.T, got *model.ResourceCommonModel)
	}{
		{
			name:         "plugin config injects name and generates id",
			resourceType: constant.PluginConfig,
			req: ResourceCreateRequest{
				Name:   "pc-demo",
				Config: json.RawMessage(`{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","rejected_code":503}}}`),
			},
			assertDraft: func(t *testing.T, got *model.ResourceCommonModel) {
				assert.NotEmpty(t, got.ID)
				assert.Equal(t, "pc-demo", gjson.GetBytes(got.Config, "name").String())
				assert.Equal(t, constant.ResourceStatusCreateDraft, got.Status)
			},
		},
		{
			name:         "consumer writes username instead of name",
			resourceType: constant.Consumer,
			req: ResourceCreateRequest{
				Name:   "consumer-demo",
				Config: json.RawMessage(`{"plugins":{}}`),
			},
			assertDraft: func(t *testing.T, got *model.ResourceCommonModel) {
				assert.Equal(t, "consumer-demo", gjson.GetBytes(got.Config, "username").String())
				assert.Empty(t, gjson.GetBytes(got.Config, "name").String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildOpenCreateDraft(10, tt.resourceType, tt.req)
			assert.Equal(t, 10, got.GatewayID)
			tt.assertDraft(t, got)
		})
	}
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer -run TestBuildOpenCreateDraft -count=1
```

Expected:
- FAIL，报 `undefined: buildOpenCreateDraft`

- [ ] **Step 3: 实现本地 builder，并让 `ToCommonResource(...)` 复用它**

在 `resource.go` 里新增（**review 重点要求：验证 `model.GetResourceNameKey(Consumer)` 的返回值**）：原代码判断的是 `gjson.GetBytes(r.Config, "name").String()` （**固定字面量 `name`**），helper 用的是 `GetResourceNameKey(resourceType)`。若 `GetResourceNameKey(Consumer) == "username"`，则对 Consumer 的判断从“name 是否为空”变成“username 是否为空”——这是一个行为差异，必须在 Task 0 characterization test 里显式记一笔并接受这种对齐（或者严格按原行为将 helper 改成固定判 `"name"`）。推荐前者：代码更一致。

```go
func buildOpenCreateDraft(
	gatewayID int,
	resourceType constant.APISIXResource,
	req ResourceCreateRequest,
) *model.ResourceCommonModel {
	config := req.Config
	if gjson.GetBytes(config, model.GetResourceNameKey(resourceType)).String() == "" {
		config, _ = sjson.SetBytes(config, model.GetResourceNameKey(resourceType), req.Name)
	}

	id := gjson.GetBytes(config, "id").String()
	if id == "" {
		id = idx.GenResourceID(resourceType)
	}

	return &model.ResourceCommonModel{
		ID:        id,
		GatewayID: gatewayID,
		Config:    datatypes.JSON(config),
		Status:    constant.ResourceStatusCreateDraft,
	}
}
```

然后把 `ToCommonResource(...)` 的循环体改成：

```go
for _, r := range rs {
	resources = append(resources, buildOpenCreateDraft(gatewayID, resourceType, r))
}
```

- [ ] **Step 4: 运行 serializer 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers.go src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers_test.go src/apiserver/pkg/apis/open/serializer/resource.go
git commit -m "refactor: extract open batch create draft builder"
```

### Task 2: 修正 Open update outer name 回写 + 抽出 update draft builder

> **按 review 要求拆为 2a / 2b 两个 PR：**
> - **Task 2a（correctness fix，failing-then-green）：** 在 Task 0 的“update 传 `name=foo` 时落库 config 应含 `name=foo`”质问下，让 `ResourceUpdateRequest.ToCommonResource(...)` 在执行前对 `r.Config` 做 `sjson.SetBytes(config, model.GetResourceNameKey(resourceType), r.Name)` 注入；make previously-failing characterization test green；不引入新 helper。
> - **Task 2b（纯 refactor）：** 再抽出 `buildOpenUpdateDraft(...)` helper，绑定 Task 2a 的行为。
>
> 理由：bugfix 和 refactor 不混在同一个 PR，便于 review 和 cherry-pick。

### Task 2: 抽出 Open update 的本地 draft builder

- [ ] Task 2: 抽出 Open update 的本地 draft builder

> **代码复核修正：** 这里不是纯 helper 提取。现状代码里 `ResourceUpdateRequest.ToCommonResource(...)` 不会把 outer `Name` 写回 `Config`，所以 Step 1 必须先锁这件事，再谈 builder 抽取。

**要解决的复杂度：** update 路径虽然比 create 简单，但同样把 `GatewayID`、`Status`、`Updater`、`Config` 组装埋在 `ToCommonResource(...)` 里，不利于之后统一 Open 域的 builder 形态。

**为什么这个任务适合单独提 PR：** 只扩展 Task 1 引入的测试文件和 `resource.go`，仍然不触碰 middleware。

**Files:**
- Modify: `src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers.go`
- Modify: `src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/open/serializer/resource.go:128-144`

- [ ] **Step 1: 先补 update draft 组装测试**

在 `open_resource_draft_helpers_test.go` 里新增：

```go
func TestBuildOpenUpdateDraft(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ginx.SetGatewayInfo(c, &model.Gateway{ID: 7})
	ginx.SetUserID(c, "openapi-user")

	got := buildOpenUpdateDraft(
		c,
		"route-id",
		constant.ResourceStatusUpdateDraft,
		json.RawMessage(`{"uri":"/demo","name":"route-demo"}`),
	)

	assert.Equal(t, "route-id", got.ID)
	assert.Equal(t, 7, got.GatewayID)
	assert.Equal(t, constant.ResourceStatusUpdateDraft, got.Status)
	assert.Equal(t, "openapi-user", got.Updater)
	assert.JSONEq(t, `{"uri":"/demo","name":"route-demo"}`, string(got.Config))
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer -run TestBuildOpenUpdateDraft -count=1
```

Expected:
- FAIL，报 `undefined: buildOpenUpdateDraft`

- [ ] **Step 3: 实现 helper，并让 update path 复用**

**实现前提（review 必改）：** Task 2a 已经在 `ResourceUpdateRequest.ToCommonResource(...)` 的开头做了 `config = sjson.SetBytes(config, model.GetResourceNameKey(resourceType), r.Name)` 注入；Task 2b（本步）只做纯 refactor，把“note 注入后的 `config` 传入 helper”这步预先完成。在 `resource.go` 增加：

```go
func buildOpenUpdateDraft(
	c *gin.Context,
	resourceID string,
	status constant.ResourceStatus,
	config json.RawMessage, // caller MUST already have injected outer name back into config (see Task 2a)
) *model.ResourceCommonModel {
	return &model.ResourceCommonModel{
		ID:        resourceID,
		GatewayID: ginx.GetGatewayInfo(c).ID,
		Config:    datatypes.JSON(config),
		Status:    status,
		BaseModel: model.BaseModel{
			Updater: ginx.GetUserID(c),
		},
	}
}
```

然后把 `ResourceUpdateRequest.ToCommonResource(...)` 改成“先注入 outer name（Task 2a 已实现），再调 helper”：

```go
func (r ResourceUpdateRequest) ToCommonResource(c *gin.Context, resourceType constant.APISIXResource, status constant.ResourceStatus) *model.ResourceCommonModel {
	config := r.Config
	if r.Name != "" {
		config, _ = sjson.SetBytes(config, model.GetResourceNameKey(resourceType), r.Name)
	}
	return buildOpenUpdateDraft(c, r.ID, status, config)
}
```

**显式接口约定：** helper 的 doc comment 写明“caller must inject outer name before calling”，避免未来添加调用者时忘记注入。

- [ ] **Step 4: 运行 serializer 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers.go src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers_test.go src/apiserver/pkg/apis/open/serializer/resource.go
git commit -m "refactor: extract open update draft builder"
```

### Task 3: 修正 middleware version-aware 校验 + 抽出 validation payload helper

> **按 review 要求拆为 3a / 3b 两个 PR：**
> - **Task 3a（correctness fix，failing-then-green）：** 将 `OpenAPIResourceCheck()` 中 `constant.ResourceRequiresIDInSchema(resourceType)` 替换为 `constant.ResourceRequiresIDInSchemaForVersion(resourceType, version)`（后者已在 `constant/resource_schema.go` 中存在，不需新增）；make Task 0 中“ConsumerGroup on 3.2.15 不应包含 id”的 failing test 转绿；不引入新 helper。
> - **Task 3b（纯 refactor）：** 再抽 `prepareOpenValidationPayload(...)` helper。

### Task 3: 抽出 Open middleware 的 validation payload helper

- [ ] Task 3: 抽出 Open middleware 的 validation payload helper

> **代码复核修正：** 这里同时承担一个 correctness fix：把当前 `ResourceRequiresIDInSchema(...)` 的判断改为 version-aware 变体，避免 3.2 / 3.3 等旧 schema 被多注入临时 `id`。

**要解决的复杂度：** middleware 里“补临时 id -> 再走 `BuildConfigRawForValidation()`”这一段属于纯整形逻辑，但现在夹在 schema 校验循环中间，改动成本高。

**为什么这个任务适合单独提 PR：** 只影响 middleware 自己的本地逻辑，不需要同时修改 serializer。

**Files:**
- Create: `src/apiserver/pkg/middleware/openapi_resource_check_test.go`
- Modify: `src/apiserver/pkg/middleware/openapi_resource_check.go:134-151`

- [ ] **Step 1: 先补 validation payload 整形矩阵测试**

在 `openapi_resource_check_test.go` 里新增：

```go
func TestPrepareOpenValidationPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType constant.APISIXResource
		version      constant.APISIXVersion
		configRaw    string
		assertPayload func(t *testing.T, payload string)
	}{
		{
			name:         "consumer group injects temporary id on 3.13",
			resourceType: constant.ConsumerGroup,
			version:      constant.APISIXVersion313,
			configRaw:    `{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","rejected_code":503}}}`,
			assertPayload: func(t *testing.T, payload string) {
				assert.NotEmpty(t, gjson.Get(payload, "id").String())
			},
		},
		{
			name:         "proto on 3.11 strips unsupported name before validation",
			resourceType: constant.Proto,
			version:      constant.APISIXVersion311,
			configRaw:    `{"name":"proto-demo","content":"syntax = \"proto3\";"}`,
			assertPayload: func(t *testing.T, payload string) {
				assert.False(t, gjson.Get(payload, "name").Exists())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := prepareOpenValidationPayload(tt.resourceType, tt.version, tt.configRaw)
			tt.assertPayload(t, string(payload))
		})
	}
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/middleware -run TestPrepareOpenValidationPayload -count=1
```

Expected:
- FAIL，报 `undefined: prepareOpenValidationPayload`

- [ ] **Step 3: 提取 helper，并让 middleware 校验循环只消费整形结果**

在 `openapi_resource_check.go` 里新增：

```go
func prepareOpenValidationPayload(
	resourceType constant.APISIXResource,
	version constant.APISIXVersion,
	configRaw string,
) json.RawMessage {
	validationRaw := configRaw
	if constant.ResourceRequiresIDInSchemaForVersion(resourceType, version) &&
		gjson.Get(validationRaw, "id").String() == "" {
		validationRaw, _ = sjson.Set(validationRaw, "id", idx.GenResourceID(resourceType))
	}
	return biz.BuildConfigRawForValidation(validationRaw, resourceType, version)
}
```

然后把 middleware 中这段：

```go
if constant.ResourceRequiresIDInSchema(resourceType) {
	...
}
configRawForValidation := biz.BuildConfigRawForValidation(...)
```

收敛成：

```go
configRawForValidation := prepareOpenValidationPayload(
	resourceType,
	ginx.GetGatewayInfo(c).GetAPISIXVersionX(),
	configRaw,
)
```

- [ ] **Step 4: 运行 middleware 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/middleware -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/middleware/openapi_resource_check.go src/apiserver/pkg/middleware/openapi_resource_check_test.go
git commit -m "refactor: extract open validation payload helper"
```

### Task 4: 引入 Open 域内的 resolved draft 上下文载体

> **review YAGNI修正：** 原计划的 `OpenResolvedDraft` 同时包含 `ValidationConfig` 和 `StorageConfig`。Task 5 Step 3 的示例代码实际只消费了 `StorageConfig`，`ValidationConfig` 从诞生就是 dead code。**本次落地时 Task 4 的结构体只留 `ID` / `Name` / `StorageConfig` 3 个字段**，一旦将来有明确的“校验 id ≠ 落库 id”测试 demonstrate 证据，再补 `ValidationConfig` 字段。

- [ ] Task 4: 引入 Open 域内的 resolved draft 上下文载体

**要解决的复杂度：** middleware 现在即使将来算出了更完整的 request identity，也没有一个 Open 域内明确的传递载体；后续很容易继续靠重复计算把逻辑摊回 serializer。

**为什么这个任务适合单独提 PR：** 这是纯粹的 Open 本地埋点，只引入结构和 context helper，不立即改最终业务行为。

**Files:**
- Create: `src/apiserver/pkg/apis/open/serializer/open_resolved_draft_context.go`
- Create: `src/apiserver/pkg/apis/open/serializer/open_resolved_draft_context_test.go`

- [ ] **Step 1: 先补 context helper 的失败测试**

在 `open_resolved_draft_context_test.go` 里新增：

```go
func TestOpenResolvedDraftContextHelpers(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	drafts := []OpenResolvedDraft{
		{
			ID:            "pc-fixed-id",
			Name:          "pc-demo",
			StorageConfig: json.RawMessage(`{"id":"pc-fixed-id","name":"pc-demo","plugins":{}}`),
		},
	}

	SetOpenResolvedDrafts(c, drafts)

	got, ok := GetOpenResolvedDrafts(c)
	assert.True(t, ok)
	assert.Len(t, got, 1)
	assert.Equal(t, "pc-fixed-id", got[0].ID)
	assert.Equal(t, "pc-demo", got[0].Name)
}
```

- [ ] **Step 2: 运行测试，确认结构和 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer -run TestOpenResolvedDraftContextHelpers -count=1
```

Expected:
- FAIL，报 `undefined: OpenResolvedDraft` / `undefined: SetOpenResolvedDrafts`

- [ ] **Step 3: 实现 resolved draft 结构和 context helper**

在 `open_resolved_draft_context.go` 里新增（**YAGNI：暂不加 ValidationConfig 字段**）：

```go
// OpenResolvedDraft carries middleware-computed identity + normalized storage payload
// across gin.Context into the serializer layer.
//
// Intentionally omits ValidationConfig for now: Task 5 only consumes StorageConfig.
// Add a ValidationConfig field here only after a characterization test demonstrates
// that validation-time id differs from storage-time id.
type OpenResolvedDraft struct {
	ID            string
	Name          string
	StorageConfig json.RawMessage
}

const openResolvedDraftsContextKey = "openapi_resolved_drafts"

func SetOpenResolvedDrafts(c *gin.Context, drafts []OpenResolvedDraft) {
	c.Set(openResolvedDraftsContextKey, drafts)
}

func GetOpenResolvedDrafts(c *gin.Context) ([]OpenResolvedDraft, bool) {
	value, ok := c.Get(openResolvedDraftsContextKey)
	if !ok {
		return nil, false
	}
	drafts, ok := value.([]OpenResolvedDraft)
	return drafts, ok
}
```

**补充测试（review 要求）：** 在 `open_resolved_draft_context_test.go` 里额外加一条测试，锁住“context key 被占用但类型不对”的行为（`c.Set(openResolvedDraftsContextKey, "not-a-draft-slice")` 后 `GetOpenResolvedDrafts` 应返回 `(nil, false)`）：

```go
t.Run("context key occupied with wrong type returns ok=false", func(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Set(openResolvedDraftsContextKey, "not-a-draft-slice")
    got, ok := GetOpenResolvedDrafts(c)
    assert.False(t, ok)
    assert.Nil(t, got)
})
```

- [ ] **Step 4: 运行 serializer 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/open/serializer/open_resolved_draft_context.go src/apiserver/pkg/apis/open/serializer/open_resolved_draft_context_test.go
git commit -m "refactor: add open resolved draft context helpers"
```

### Task 5: 让 Open middleware 和 serializer 复用同一份 resolved identity

> **Review 必改：范围声明补充**——本 Task 需要修改 `ResourceBatchCreateRequest.ToCommonResource(...)` 的签名（从 `(gatewayID int, resourceType)` 改为 `(c *gin.Context, resourceType)`），这属于 public API change，需同步修改 `handler/resource.go` 中的所有调用点（目前只有 batch create handler 调用了它）。

- [ ] Task 5: 让 Open middleware 和 serializer 复用同一份 resolved identity

**要解决的复杂度：** 当前 middleware 和 serializer 各自推导一次 request identity，create/batch create 很容易出现“校验时是一个 id，落库时是另一个 id”。

**为什么这个任务适合单独提 PR：** 这是 Open 域内最关键的一步，但边界仍然局限在 middleware、open handler、open serializer 三个文件。

**Files:**
- Modify: `src/apiserver/pkg/middleware/openapi_resource_check.go`
- Modify: `src/apiserver/pkg/apis/open/serializer/resource.go`
- Modify: `src/apiserver/pkg/apis/open/handler/resource.go`
- Modify: `src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers_test.go`

- [ ] **Step 1: 先补“复用 middleware draft”的失败测试**

在 `open_resource_draft_helpers_test.go` 中新增：

```go
func TestResourceBatchCreateUsesOpenResolvedDrafts(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ginx.SetGatewayInfo(c, &model.Gateway{ID: 9, APISIXVersion: string(constant.APISIXVersion313)})

		SetOpenResolvedDrafts(c, []OpenResolvedDraft{
		{
			ID:            "pc-from-middleware",
			Name:          "pc-demo",
			StorageConfig: json.RawMessage(`{"id":"pc-from-middleware","name":"pc-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","rejected_code":503}}}`),
		},
	})

	req := ResourceBatchCreateRequest{
		{
			Name:   "pc-demo",
			Config: json.RawMessage(`{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","rejected_code":503}}}`),
		},
	}

	got := req.ToCommonResource(c, constant.PluginConfig)
	assert.Len(t, got, 1)
	assert.Equal(t, "pc-from-middleware", got[0].ID)
	assert.JSONEq(
		t,
		`{"id":"pc-from-middleware","name":"pc-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","rejected_code":503}}}`,
		string(got[0].Config),
	)
}
```

- [ ] **Step 2: 运行测试，确认现有实现还拿不到 middleware 的 draft**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer -run TestResourceBatchCreateUsesOpenResolvedDrafts -count=1
```

Expected:
- FAIL，表现为 `ToCommonResource` 仍然重新生成 id，断言不通过

- [ ] **Step 3: 改造 middleware / handler / serializer，让 resolved draft 真正贯通**

按下面顺序改：

1. 在 middleware 校验循环里为每个请求项生成 `OpenResolvedDraft`（**不再写入 `ValidationConfig`**）：

```go
drafts = append(drafts, serializer.OpenResolvedDraft{
	ID:            gjson.GetBytes(configRawForValidation, "id").String(),
	Name:          config.Get("name").String(),
	StorageConfig: []byte(configRaw),
})
serializer.SetOpenResolvedDrafts(c, drafts)
```

2. 把 batch create 的 serializer 方法签名从：

```go
func (rs ResourceBatchCreateRequest) ToCommonResource(gatewayID int, resourceType constant.APISIXResource) []*model.ResourceCommonModel
```

改成：

```go
func (rs ResourceBatchCreateRequest) ToCommonResource(c *gin.Context, resourceType constant.APISIXResource) []*model.ResourceCommonModel
```

3. 在 `ToCommonResource(...)` 里优先消费 middleware 放进来的 draft（**review 补充测试**：`len(drafts) != len(rs)` 的 fallback 路径必须有显式 test 覆盖）：

```go
resources := make([]*model.ResourceCommonModel, 0, len(rs))
if drafts, ok := GetOpenResolvedDrafts(c); ok && len(drafts) == len(rs) {
	for idx, req := range rs {
		draft := drafts[idx]
		resources = append(resources, &model.ResourceCommonModel{
			ID:        draft.ID,
			GatewayID: ginx.GetGatewayInfo(c).ID,
			Config:    datatypes.JSON(draft.StorageConfig),
			Status:    constant.ResourceStatusCreateDraft,
		})
		_ = req
	}
	return resources
}
// Fallback path: when drafts are missing or count mismatches, fall back to local builder.
for _, r := range rs {
	resources = append(resources, buildOpenCreateDraft(ginx.GetGatewayInfo(c).ID, resourceType, r))
}
return resources
```

**必补测试（review）：** 在 `TestResourceBatchCreateUsesOpenResolvedDrafts` 基础上再加一条 `len(drafts) != len(rs)` 的子用例，断言 fallback 路径产出的 resources 数量 == `len(rs)`，而非空 slice。

4. `handler/resource.go` 里改成：

```go
resources := req.ToCommonResource(c, ginx.GetResourceType(c))
```

- [ ] **Step 4: 运行 Open 相关测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/open/serializer ./pkg/middleware ./pkg/apis/open/handler -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/middleware/openapi_resource_check.go src/apiserver/pkg/apis/open/serializer/open_resource_draft_helpers_test.go src/apiserver/pkg/apis/open/serializer/open_resolved_draft_context.go src/apiserver/pkg/apis/open/serializer/resource.go src/apiserver/pkg/apis/open/handler/resource.go
git commit -m "refactor: reuse open resolved drafts across validation and persistence"
```

## 完成定义

- Open create / update 的 draft 组装都有各自清晰的本地 builder
- Open middleware 的校验 payload 整形是一个纯 helper，而不是散在校验循环中
- Open 域内存在显式的 resolved draft 载体
- middleware 与 serializer 在 create/batch create 上复用同一份 resolved identity
