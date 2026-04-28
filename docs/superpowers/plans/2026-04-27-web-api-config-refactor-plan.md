# Web API Config 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **Execution rule:** If a task or step is done, mark it in this `plan.md` before running `git add` and `git commit`.

**Goal:** 在不改变 Web API 协议、不触碰 `HandleConfig()` 边界的前提下，逐步收敛 `web api` 当前分散在 serializer 和 handler 中的 `config` 校验整形、生成 ID 时机、以及 create draft 组装重复逻辑。

**Architecture:** 本计划只处理 `pkg/apis/web` 域内复杂度，不提前抽跨领域共享 helper。执行顺序调整为：先独立补 `CheckAPISIXConfig()` 和 create handler 的 characterization tests，再把 `CheckAPISIXConfig()` 的 identity / payload 逻辑拆清，最后收拢 create handler 里“先生成 ID 再校验”和“组装 `ResourceCommonModel`”两类重复动作。

**Tech Stack:** Go, Gin, `validator`, `testify`, `go test`, `make lint`, `make test`

---

## 代码复核结论

- 重构目的判断：基本正确。`CheckAPISIXConfig()` 的前置整形逻辑和 Web create handler 的 draft 组装重复，都是当前代码里的真实复杂度来源。
- 复杂度评估：Task 1-2 为低到中等；Task 3-5 为中等，因为当前仓库里还没有 `pkg/apis/web/handler` 的直接测试，前置测试基座需要先补。
- 本次修正：把 create handler characterization tests 提升为独立前置阶段；先锁特殊 3 条 create 路径，再迁移其余 handler，避免 helper 先行。

## 执行顺序（修订）

1. Task 0：独立补 serializer / create handler characterization tests。
2. Task 1-2：先清理 `CheckAPISIXConfig()` 的 identity / payload 逻辑。
3. Task 3-4：再处理 `plugin_config` / `consumer_group` / `global_rule` 这 3 条“先生成 ID 再校验”的特殊 create 路径。
4. Task 5：最后把其余 create handler 迁移到本地 draft helper。

## 范围

- 只处理 `src/apiserver/pkg/apis/web/...`
- 允许新增 Web 域内 helper 文件
- 允许调整 Web create handler 的命名、函数拆分、调用顺序表达

## 非目标

- 不抽跨 `web/open/import` 的共享 helper
- 不改 `pkg/entity/model/*.go` 中各资源 `HandleConfig()`
- 不改 publish 链路
- 不把 `mcp` 拉进本计划

## 文件结构

- `src/apiserver/pkg/apis/web/serializer/common.go`
  - `CheckAPISIXConfig()` 的校验前整形逻辑
- `src/apiserver/pkg/apis/web/serializer/common_test.go`
  - `CheckAPISIXConfig()` 本地 helper 与校验 payload 行为矩阵
- `src/apiserver/pkg/apis/web/handler/create_handlers_test.go`
  - Web create handler 的 seam-first characterization tests
- `src/apiserver/pkg/apis/web/handler/web_create_helpers.go`
  - Web create handler 本地 helper，只服务 `pkg/apis/web/handler`
- `src/apiserver/pkg/apis/web/handler/web_create_helpers_test.go`
  - create helper 的 TDD 用例
- `src/apiserver/pkg/apis/web/handler/plugin_config.go`
- `src/apiserver/pkg/apis/web/handler/consumer_group.go`
- `src/apiserver/pkg/apis/web/handler/global_rule.go`
- `src/apiserver/pkg/apis/web/handler/route.go`
- `src/apiserver/pkg/apis/web/handler/service.go`
- `src/apiserver/pkg/apis/web/handler/upstream.go`
- `src/apiserver/pkg/apis/web/handler/consumer.go`
- `src/apiserver/pkg/apis/web/handler/stream_route.go`
- `src/apiserver/pkg/apis/web/handler/proto.go`
- `src/apiserver/pkg/apis/web/handler/plugin_metadata.go`
- `src/apiserver/pkg/apis/web/handler/ssl.go`
  - 这些文件只做 create 路径上的局部收敛，不顺手改 update/list 等其他逻辑

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
  - 第一层：锁定现有行为，确保重构前能跑、重构后继续通过
  - 第二层：在 helper 抽出后补 helper 单测，避免 helper 自己再退化
- 当前仓库里没有 `pkg/apis/web/handler` 的测试文件；Task 0 需要先补 create handler characterization tests，不能直接从 `web_create_helpers_test.go` 起手。
- Web 计划里的现有 seam 优先级如下：
  - Task 0：优先直接测 `PluginConfigCreate`、`ConsumerGroupCreate`、`GlobalRuleCreate` 和至少一个默认 create handler（如 `RouteCreate`）
  - Task 1-2：优先通过 `validation.ValidateStruct(...)` 触发带 `apisixConfig` tag 的现有 serializer 校验路径，从外部行为锁定 `CheckAPISIXConfig()`
  - Task 3-4：优先直接测 `PluginConfigCreate`、`ConsumerGroupCreate`、`GlobalRuleCreate` 这 3 个现有 handler
  - Task 5：优先直接测其余现有 `XxxCreate` handler，而不是先测 `buildWebCreateDraft(...)`
- 执行时，如果任务正文里的示例代码先写了 helper 测试，应按上面的 seam 规则落地：先补现有 seam 的 characterization test，再补 helper test。

## 重构前测试前置阶段（独立）

- Task 0 至少覆盖 5 类现状：`CheckAPISIXConfig()` 的 identity fallback；`CheckAPISIXConfig()` 的 validation payload 整形；特殊 3 条 create handler 的“先生成 ID 再校验”；默认 create handler 的“先校验再生成 ID，但 draft 组装一致”；**`SSLCreate` 的 create 路径**（证书解析 + `validity_start/validity_end` 处理 + 默认组装）的当前行为，避免 Task 5 迁移 SSL 时黑盒改动。
- **Task 0 必须显式断言当前特殊 3 条 create handler（`PluginConfigCreate` / `ConsumerGroupCreate` / `GlobalRuleCreate`）写入 `ResourceCommonModel` 时 `Updater` 字段的当前值**：当前分支原代码已经同时写入 `Creator` 和 `Updater`，并且两者都等于 `userID`。Task 4 `buildWebCreateDraft` 需要保持这个现状，不应引入新的 `Updater` 行为漂移。
- `create_handlers_test.go` 负责锁 handler 入口行为；`web_create_helpers_test.go` 只在 helper 抽出后再补第二层单测。
- 如果 Task 0 还没建好，不要直接推进 Task 3-5；否则后面很难分清是 handler 入口行为变了，还是 helper 自己写错了。

### Task 0: 补 Web serializer / create handler characterization tests

- [x] Task 0: 补 Web serializer / create handler characterization tests

**要解决的缺口：** 现在文档只在说明里提了 Task 0，但正文还没有一个真正独立的前置补测任务。先把 `CheckAPISIXConfig()` 和 create handler 入口行为锁住，后面的 helper 提取才不会把“重构”和“补洞”混在一起。

**为什么这个任务适合单独提 PR：** 只新增 serializer / handler 的 characterization tests，不改 `pkg/apis/web` 的生产逻辑。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/serializer/common_test.go`
- Create: `src/apiserver/pkg/apis/web/handler/create_handlers_test.go`

- [x] **Step 1: 在现有 serializer 和 create handler seam 上补 characterization tests**

至少覆盖下面 5 类现状，避免直接从 `web_create_helpers_test.go` 或新 helper 起手：

- `CheckAPISIXConfig()` 的 identity fallback
- `CheckAPISIXConfig()` 的 validation payload 整形
- `PluginConfigCreate`、`ConsumerGroupCreate`、`GlobalRuleCreate` 这 3 条特殊 create 路径的“先生成 ID 再校验”
- 至少一条默认 create handler（如 `RouteCreate`）的“先校验再生成 ID，但 draft 组装一致”
- **`SSLCreate` 的 create 路径**：证书字段、`validity_start/validity_end`、默认 draft 组装同时被锁住，作为 Task 5 迁移 SSL 时的黑盒护栏

**额外断言（review 补充，当前行为锁定）：**
在上面第 3 点的 3 条特殊 create handler 测试里，除了断言 `req.ID` 已填充、`Creator == userID` 外，**还要显式断言 `Updater == userID`**（当前现状）。Task 4 落地时需要保持这组断言不变，确认 helper 抽取没有带来行为漂移。

- [x] **Step 2: 运行 Web seam tests，确认入口行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer ./pkg/apis/web/handler -count=1
```

Expected:
- PASS

- [x] **Step 3: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/serializer/common_test.go src/apiserver/pkg/apis/web/handler/create_handlers_test.go
git commit -m "test: lock web serializer and create handler seams"
```

---

### Task 1: 抽出 `CheckAPISIXConfig()` 的 identity 决策 helper

- [x] Task 1: 抽出 `CheckAPISIXConfig()` 的 identity 决策 helper

**要解决的复杂度：** `CheckAPISIXConfig()` 同时负责“读 config 识别 identity”和“继续做 schema 校验”，identity 决策散在函数中间，后续规则变化容易改漏。

> **当前代码实况修正：** 当前 `web` consumer serializer 并没有单独的 `Username` 字段，只有 `Name` 字段；而 `CheckAPISIXConfig()` 现状仍通过 `getResourceNameByResourceType(...)` 在调用点决定 fallback identity。Task 1 先抽“config identity vs caller-provided fallback identity”的决策 helper，不在这一 PR 里顺手改 consumer fallback 语义。

**为什么这个任务适合单独提 PR：** 只碰 `serializer/common.go` 和对应测试，不改变 handler 行为，也不涉及跨文件迁移。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/serializer/common.go:49-127`
- Modify: `src/apiserver/pkg/apis/web/serializer/common_test.go`

- [x] **Step 1: 先补当前 identity 决策的测试**

在 `common_test.go` 增加下面这组失败测试：

```go
func TestResolveWebValidationIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            webValidationInput
		wantIdentity     string
		wantUsedFallback bool
	}{
		{
			name: "falls back to provided identity when config id is absent",
			input: webValidationInput{
				RawConfig:         json.RawMessage(`{"plugins":{}}`),
				FallbackIdentity: "route-a",
			},
			wantIdentity:     "route-a",
			wantUsedFallback: true,
		},
		{
			name: "existing config id wins",
			input: webValidationInput{
				RawConfig:         json.RawMessage(`{"id":"route-fixed","plugins":{}}`),
				FallbackIdentity: "route-a",
			},
			wantIdentity:     "route-fixed",
			wantUsedFallback: false,
		},
		{
			name: "empty fallback is preserved when no config id exists",
			input: webValidationInput{
				RawConfig: json.RawMessage(`{"plugins":{}}`),
			},
			wantIdentity:     "",
			wantUsedFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIdentity, gotUsedFallback := resolveWebValidationIdentity(tt.input)
			assert.Equal(t, tt.wantIdentity, gotIdentity)
			assert.Equal(t, tt.wantUsedFallback, gotUsedFallback)
		})
	}
}
```

- [x] **Step 2: 运行测试，确认它先失败**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -run TestResolveWebValidationIdentity -count=1
```

Expected:
- FAIL，报 `undefined: webValidationInput` 或 `undefined: resolveWebValidationIdentity`

- [x] **Step 3: 用最小实现抽出 identity helper，并接回 `CheckAPISIXConfig()`**

在 `common.go` 里新增本地输入结构和 helper（**当前代码实况修正**：Task 1 不假设 web consumer 有单独的 `Username` 字段；helper 只处理“config identity vs 调用点已算好的 fallback identity”）：

```go
type webValidationInput struct {
	RawConfig         json.RawMessage
	FallbackIdentity string
}

func resolveWebValidationIdentity(input webValidationInput) (string, bool) {
	if identity := schema.GetResourceIdentification(input.RawConfig); identity != "" {
		return identity, false
	}
	return input.FallbackIdentity, true
}
```

然后把 `CheckAPISIXConfig()` 中原来这段：

```go
resourceIdentification := schema.GetResourceIdentification(rawConfig)
if resourceIdentification == "" {
	resourceIdentification = getResourceNameByResourceType(resourceTypeName, fl)
	...
}
```

改成：

```go
resourceIdentification, usedFallback := resolveWebValidationIdentity(webValidationInput{
	RawConfig:         rawConfig,
	FallbackIdentity: getResourceNameByResourceType(resourceTypeName, fl),
})
if usedFallback {
	...
}
```

- [x] **Step 4: 运行 serializer 包测试，确认行为不变**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -count=1
```

Expected:
- PASS

- [x] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/serializer/common.go src/apiserver/pkg/apis/web/serializer/common_test.go
git commit -m "refactor: extract web validation identity helper"
```

### Task 2: 抽出 Web 本地 validation payload helper

> **Review 建议（范围控制）：** 本 Task 外表上是抽一个 helper，实际上 `prepareWebValidationPayload` 同时涉及 `injectGeneratedIDForValidation` / `shouldInjectResourceNameForValidation` / `GetResourceNameKey` / `PluginMetadata` 特判 4 件事，对 `CheckAPISIXConfig()` 核心分支是整体搬迁。落地时允许再拆为两个 PR：
> - Task 2a：先抽 `id` 注入部分（`injectGeneratedIDForValidation`），保留原 `CheckAPISIXConfig()` 的 `name` 注入和 `plugin_metadata` 特判在原处
> - Task 2b：再抽 `name` 注入 + `plugin_metadata.id = name` 特判，让 helper 达到下面的最终形态
>
> 如果 PR 体积在可接受范围内（e.g. <200 行改动），可以不拆；但执行者需明确意识到 helper 在一次抽多层行为。

- [x] Task 2: 抽出 Web 本地 validation payload helper

**要解决的复杂度：** 现在 `id` 注入、`name/username` 注入、`plugin_metadata.id = name` 都埋在 `CheckAPISIXConfig()` 主流程里，修改时必须通读整段 validator。

**为什么这个任务适合单独提 PR：** 仍然只碰 `serializer/common.go` 和测试；这是在 Task 1 的基础上继续把剩余 payload 逻辑从 validator 主流程中拿出来。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/serializer/common.go:49-127`
- Modify: `src/apiserver/pkg/apis/web/serializer/common_test.go`

- [x] **Step 1: 先补当前 payload 整形矩阵测试**

在 `common_test.go` 里新增：

```go
func TestPrepareWebValidationPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        webValidationInput
		wantPayload  string
		wantIdentity string
	}{
		{
			name: "consumer group injects generated id and then uses that id as identity on 3.13",
			input: webValidationInput{
				ResourceType:     constant.ConsumerGroup,
				Version:          constant.APISIXVersion313,
				ResourceID:       "cg-generated-id",
				FallbackIdentity: "cg-demo",
				RawConfig:        json.RawMessage(`{"plugins":{}}`),
			},
			wantPayload:  `{"plugins":{},"id":"cg-generated-id"}`,
			wantIdentity: "cg-generated-id",
		},
		{
			name: "proto on 3.11 keeps name out of payload",
			input: webValidationInput{
				ResourceType:     constant.Proto,
				Version:          constant.APISIXVersion311,
				FallbackIdentity: "proto-demo",
				RawConfig:        json.RawMessage(`{"content":"syntax = \\\"proto3\\\";"}`),
			},
			wantPayload:  `{"content":"syntax = \"proto3\";"}`,
			wantIdentity: "proto-demo",
		},
		{
			name: "plugin metadata uses outer name as id on update-like input",
			input: webValidationInput{
				ResourceType:     constant.PluginMetadata,
				Version:          constant.APISIXVersion313,
				ResourceID:       "existing-plugin-metadata-id",
				Name:             "authz-casbin",
				FallbackIdentity: "authz-casbin",
				RawConfig: json.RawMessage(`{
					"model": "rbac_model.conf",
					"policy": "rbac_policy.csv"
				}`),
			},
			wantPayload: `{
				"model": "rbac_model.conf",
				"policy": "rbac_policy.csv",
				"id": "authz-casbin"
			}`,
			wantIdentity: "authz-casbin",
		},
		{
			name: "ssl never injects name",
			input: webValidationInput{
				ResourceType:     constant.SSL,
				Version:          constant.APISIXVersion313,
				FallbackIdentity: "ssl-demo",
				RawConfig:        json.RawMessage(`{"cert":"demo","key":"demo","snis":["demo.com"]}`),
			},
			wantPayload:  `{"cert":"demo","key":"demo","snis":["demo.com"]}`,
			wantIdentity: "ssl-demo",
		},
		{
			name: "existing config id stays authoritative when fallback is empty",
			input: webValidationInput{
				ResourceType: constant.ConsumerGroup,
				Version:      constant.APISIXVersion313,
				ResourceID:   "cg-generated-id",
				RawConfig:    json.RawMessage(`{"id":"cg-fixed","plugins":{}}`),
			},
			wantPayload:  `{"id":"cg-fixed","plugins":{}}`,
			wantIdentity: "cg-fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPayload, gotIdentity := prepareWebValidationPayload(tt.input)
			assert.JSONEq(t, tt.wantPayload, string(gotPayload))
			assert.Equal(t, tt.wantIdentity, gotIdentity)
		})
	}
}
```

- [x] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -run TestPrepareWebValidationPayload -count=1
```

Expected:
- FAIL，报 `undefined: prepareWebValidationPayload` 或 `webValidationInput` 尚未扩展出 Task 2 需要的字段

- [x] **Step 3: 提取 payload builder，并让 `CheckAPISIXConfig()` 只做 orchestration**

**注意（review 修正）：** `prepareWebValidationPayload` **内部会调用一次 `resolveWebValidationIdentity`**；为避免 `CheckAPISIXConfig()` 外部再算一次 identity 导致重复计算，Task 2 的 helper 返回 `(rawConfig, identity)` 二元组，`CheckAPISIXConfig()` 消费后不再自行调用 identity helper。

同时在 Task 1 的 `webValidationInput` 基础上扩充 `ResourceType` / `Version` / `ResourceID` / `Name`；继续保留 `FallbackIdentity` 作为调用点已算好的 current-seam fallback。在 `common.go` 增加：

```go
type webValidationInput struct {
	ResourceType     constant.APISIXResource
	Version          constant.APISIXVersion
	ResourceID       string
	Name             string
	RawConfig        json.RawMessage
	FallbackIdentity string
}

func prepareWebValidationPayload(input webValidationInput) (json.RawMessage, string) {
	rawConfig := injectGeneratedIDForValidation(
		input.RawConfig,
		input.ResourceType,
		input.Version,
		input.ResourceID,
	)

	identity, usedFallback := resolveWebValidationIdentity(webValidationInput{
		RawConfig:         rawConfig,
		FallbackIdentity: input.FallbackIdentity,
	})

	if usedFallback && shouldInjectResourceNameForValidation(input.ResourceType, input.Version) {
		rawConfig, _ = sjson.SetBytes(rawConfig, model.GetResourceNameKey(input.ResourceType), identity)
	}
	if input.ResourceType == constant.PluginMetadata {
		rawConfig, _ = sjson.SetBytes(rawConfig, "id", input.Name)
	}
	return rawConfig, identity
}
```

然后把 `CheckAPISIXConfig()` 前半段收敛成：

```go
input := webValidationInput{
	ResourceType:     resourceType,
	Version:          gatewayInfo.GetAPISIXVersionX(),
	ResourceID:       fl.Parent().FieldByName("ID").String(),
	Name:             fl.Parent().FieldByName("Name").String(),
	RawConfig:        rawConfig,
	FallbackIdentity: getResourceNameByResourceType(resourceTypeName, fl),
}
rawConfig, resourceIdentification := prepareWebValidationPayload(input)
```

**补充测试（review 要求 — identity integration）：** 在 `common_test.go` 里再增一条矩阵 case，锁住“`ConsumerGroup` on 3.13 + outer fallback 为空 + `Config` 已有 `id` → payload prep 后 identity 与 payload.id 一致且稳定”，防止 Task 2 在 helper 内重新算 identity 时跳层带回 regression。

**测试补充（review 要求 — update 路径）：** `PluginMetadata` 的 `id = name` 特规则不仅影响 create 路径。`common_test.go` 里要在原矩阵基础上再补一条 update 路径 case，驱动 `CheckAPISIXConfig()` 通过 `apisixConfig` tag 触发时，`PluginMetadata` 的 update payload 也会带上 `id=name`。

- [x] **Step 4: 运行 serializer 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -count=1
```

Expected:
- PASS

- [x] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/serializer/common.go src/apiserver/pkg/apis/web/serializer/common_test.go
git commit -m "refactor: extract web validation payload helper"
```

### Task 3: 收拢 3 个“先生成 ID 再校验”的 create 路径

- [ ] Task 3: 收拢 3 个“先生成 ID 再校验”的 create 路径

**要解决的复杂度：** `plugin_config`、`consumer_group`、`global_rule` 三个 handler 现在都手写一遍 `ShouldBindJSON -> GenResourceID -> ValidateStruct`，同一类顺序逻辑重复 3 次。

**为什么这个任务适合单独提 PR：** 只动 3 个特殊 create handler 和一个新的 Web 本地 helper 文件，不影响其他 create 路径。

**Files:**
- Create: `src/apiserver/pkg/apis/web/handler/web_create_helpers.go`
- Create: `src/apiserver/pkg/apis/web/handler/web_create_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/web/handler/plugin_config.go:49-62`
- Modify: `src/apiserver/pkg/apis/web/handler/consumer_group.go:50-62`
- Modify: `src/apiserver/pkg/apis/web/handler/global_rule.go:49-62`

- [ ] **Step 1: 先补这 3 条 create 顺序的失败测试**

在 `web_create_helpers_test.go` 里新增（**review 要求**：3 个子测试必须一次性全部写出来，不要留“同文件再补”）：

```go
func TestBindAndValidateWebCreateWithGeneratedID(t *testing.T) {
	util.InitEmbedDb()
	gin.SetMode(gin.TestMode)

	gateway := &model.Gateway{ID: 1, APISIXVersion: string(constant.APISIXVersion313)}

	t.Run("plugin config gets id before validation", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/",
			strings.NewReader(`{"name":"pc-demo","config":{"plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","rejected_code":503}}}}`),
		)
		ginx.SetGatewayInfo(c, gateway)

		var req serializer.PluginConfigInfo
		err := bindAndValidateWebCreateWithGeneratedID(c, &req, constant.PluginConfig, func(resourceID string) {
			req.ID = resourceID
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, req.ID)
	})
}
```

同文件再补 `consumer_group`、`global_rule` 两个子测试（**review**：完整给出，不仅是在注释里记一笔），断言点保持一致：`err == nil`、`req.ID` 已被填充；在 Task 0 已锁定“`Updater==""`”的前提下，本步不改组装逻辑，因此不断言 `Updater`。

**注意 Step 3 的实现签名（review 建议记录设计理由）：** helper 使用 `setResourceID func(string)` 回调而非反射设置 `req.ID`，是为了避免在 hot path 里引入反射。属于有意识的取舍，需在注释里写明原因。

- [ ] **Step 2: 运行测试，确认 helper 尚未实现**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/handler -run TestBindAndValidateWebCreateWithGeneratedID -count=1
```

Expected:
- FAIL，报 `undefined: bindAndValidateWebCreateWithGeneratedID`

- [ ] **Step 3: 新增 Web 本地 bind/validate helper，并迁移 3 个 handler**

在 `web_create_helpers.go` 增加：

```go
func bindAndValidateWebCreateWithGeneratedID(
	c *gin.Context,
	req any,
	resourceType constant.APISIXResource,
	setResourceID func(string),
) error {
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}
	setResourceID(idx.GenResourceID(resourceType))
	return validation.ValidateStruct(c.Request.Context(), req)
}
```

然后把 3 个 handler 的起手逻辑都改成同一模式：

```go
var req serializer.PluginConfigInfo
if err := bindAndValidateWebCreateWithGeneratedID(
	c,
	&req,
	constant.PluginConfig,
	func(resourceID string) { req.ID = resourceID },
); err != nil {
	ginx.BadRequestErrorJSONResponse(c, err)
	return
}
```

`consumer_group` 和 `global_rule` 只替换类型和 `constant.*` 即可，不改后续 `biz.CreateXxx(...)`。

- [ ] **Step 4: 运行 handler 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/handler -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/handler/web_create_helpers.go src/apiserver/pkg/apis/web/handler/web_create_helpers_test.go src/apiserver/pkg/apis/web/handler/plugin_config.go src/apiserver/pkg/apis/web/handler/consumer_group.go src/apiserver/pkg/apis/web/handler/global_rule.go
git commit -m "refactor: unify generated-id web create validation flow"
```

### Task 4: 抽出特殊 create handler 的 draft 组装 helper

- [ ] Task 4: 抽出特殊 create handler 的 draft 组装 helper

**要解决的复杂度：** 上一步收拢了校验顺序，但 3 个特殊 handler 里仍然各自手写一份相同的 `ResourceCommonModel` 组装。

**为什么这个任务适合单独提 PR：** 只影响已经归到同一类的 3 个 handler，风险边界清晰。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/handler/web_create_helpers.go`
- Modify: `src/apiserver/pkg/apis/web/handler/web_create_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/web/handler/plugin_config.go:64-72`
- Modify: `src/apiserver/pkg/apis/web/handler/consumer_group.go:64-71`
- Modify: `src/apiserver/pkg/apis/web/handler/global_rule.go:64-71`

- [ ] **Step 1: 先补 create draft 组装测试**

在 `web_create_helpers_test.go` 里新增：

```go
func TestBuildWebCreateDraft(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ginx.SetGatewayInfo(c, &model.Gateway{ID: 12})
	ginx.SetUserID(c, "tester")

	draft := buildWebCreateDraft(
		c,
		"resource-id",
		json.RawMessage(`{"plugins":{"limit-count":{"count":1}}}`),
	)

	assert.Equal(t, "resource-id", draft.ID)
	assert.Equal(t, 12, draft.GatewayID)
	assert.Equal(t, constant.ResourceStatusCreateDraft, draft.Status)
	assert.Equal(t, "tester", draft.Creator)
	assert.Equal(t, "tester", draft.Updater)
	assert.JSONEq(t, `{"plugins":{"limit-count":{"count":1}}}`, string(draft.Config))
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/handler -run TestBuildWebCreateDraft -count=1
```

Expected:
- FAIL，报 `undefined: buildWebCreateDraft`

- [ ] **Step 3: 实现 helper，并让 3 个 handler 复用**

在 `web_create_helpers.go` 增加（**行为保持要求**：当前特殊 3 条 create handler 原代码已经写入 `Updater = userID`；helper 抽取后必须保持 `Updater == Creator == userID` 这一现状）：

```go
func buildWebCreateDraft(
	c *gin.Context,
	resourceID string,
	config json.RawMessage,
) model.ResourceCommonModel {
	userID := ginx.GetUserID(c)
	return model.ResourceCommonModel{
		ID:        resourceID,
		GatewayID: ginx.GetGatewayInfo(c).ID,
		Config:    datatypes.JSON(config),
		Status:    constant.ResourceStatusCreateDraft,
		BaseModel: model.BaseModel{
			Creator: userID,
			Updater: userID,
		},
	}
}
```

然后把 3 个 handler 里的内联组装改成：

```go
pluginConfig := &model.PluginConfig{
	Name: req.Name,
	ResourceCommonModel: buildWebCreateDraft(c, req.ID, req.Config),
}
```

`consumer_group` / `global_rule` 同样只保留资源本地字段，`ResourceCommonModel` 统一走 helper。

**行为保持同步（review）：** 保持 Task 0 在 `create_handlers_test.go` 里已经锁住的 `Updater == ginx.GetUserID(c)`（== `Creator`）断言不变，确认 helper 提取没有改变 create draft 的写入结果。

- [ ] **Step 4: 运行 handler 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/handler -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/handler/web_create_helpers.go src/apiserver/pkg/apis/web/handler/web_create_helpers_test.go src/apiserver/pkg/apis/web/handler/plugin_config.go src/apiserver/pkg/apis/web/handler/consumer_group.go src/apiserver/pkg/apis/web/handler/global_rule.go
git commit -m "refactor: extract web create draft builder for special handlers"
```

### Task 5: 把其余 Web create handler 统一迁移到本地 draft helper

> **Review 建议（Files 语义描述）：** 下面 `Files:` 里出现的行号（如 `route.go:61-69`）是押当前 master 的满包描述，落地时位置可能漂移。执行者应以“对应 handler 中 `ResourceCommonModel:` 字面量所在行段”为准，而非硬绑行号。SSL 迁移时务必保留证书解析 + `validity_start/validity_end` 处理逻辑（由 Task 0 的 SSLCreate characterization test 兑保）。

- [ ] Task 5: 把其余 Web create handler 统一迁移到本地 draft helper

**要解决的复杂度：** 其余 create handler 仍然各自内联组装 `ResourceCommonModel`，即使生成 ID 的时机不同，draft 组装本身仍是重复代码。

**为什么这个任务适合单独提 PR：** 这一步只做“组装 helper 迁移”，不改变其余 handler 现在“先校验后生成 ID”的既有顺序。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/handler/web_create_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/web/handler/route.go:61-69`
- Modify: `src/apiserver/pkg/apis/web/handler/service.go:59-67`
- Modify: `src/apiserver/pkg/apis/web/handler/upstream.go:60-68`
- Modify: `src/apiserver/pkg/apis/web/handler/consumer.go:60-68`
- Modify: `src/apiserver/pkg/apis/web/handler/stream_route.go:58-66`
- Modify: `src/apiserver/pkg/apis/web/handler/proto.go:56-64`
- Modify: `src/apiserver/pkg/apis/web/handler/plugin_metadata.go:58-66`
- Modify: `src/apiserver/pkg/apis/web/handler/ssl.go:109-117`

- [ ] **Step 1: 先补迁移保护测试**

在 `web_create_helpers_test.go` 增加一个面向默认 create 路径的测试，锁定“helper 只负责组装，不影响 ID 生成时机”：

```go
func TestBuildWebCreateDraftDoesNotGenerateID(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ginx.SetGatewayInfo(c, &model.Gateway{ID: 21})
	ginx.SetUserID(c, "tester")

	draft := buildWebCreateDraft(c, "route-generated-later", json.RawMessage(`{"uri":"/demo"}`))

	assert.Equal(t, "route-generated-later", draft.ID)
	assert.JSONEq(t, `{"uri":"/demo"}`, string(draft.Config))
}
```

- [ ] **Step 2: 运行测试，确认它先保护住 helper 合约**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/handler -run 'TestBuildWebCreateDraft|TestBuildWebCreateDraftDoesNotGenerateID' -count=1
```

Expected:
- PASS

- [ ] **Step 3: 逐个迁移剩余 create handler，只替换 draft 组装，不改顺序**

把类似下面这段：

```go
ResourceCommonModel: model.ResourceCommonModel{
	ID:        idx.GenResourceID(constant.Route),
	GatewayID: ginx.GetGatewayInfo(c).ID,
	Config:    datatypes.JSON(req.Config),
	Status:    constant.ResourceStatusCreateDraft,
	BaseModel: model.BaseModel{
		Creator: ginx.GetUserID(c),
	},
},
```

统一改成：

```go
ResourceCommonModel: buildWebCreateDraft(
	c,
	idx.GenResourceID(constant.Route),
	req.Config,
),
```

注意：
- `route/service/upstream/consumer/stream_route/proto/plugin_metadata/ssl` 都只做这一类替换
- `ssl` 仍保留自己证书相关字段处理
- 不顺手改 update 路径

- [ ] **Step 4: 运行 handler 包测试，再跑一次 Web serializer 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/handler ./pkg/apis/web/serializer -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/handler/web_create_helpers_test.go src/apiserver/pkg/apis/web/handler/route.go src/apiserver/pkg/apis/web/handler/service.go src/apiserver/pkg/apis/web/handler/upstream.go src/apiserver/pkg/apis/web/handler/consumer.go src/apiserver/pkg/apis/web/handler/stream_route.go src/apiserver/pkg/apis/web/handler/proto.go src/apiserver/pkg/apis/web/handler/plugin_metadata.go src/apiserver/pkg/apis/web/handler/ssl.go
git commit -m "refactor: reuse web create draft builder across handlers"
```

## 完成定义

- `CheckAPISIXConfig()` 只负责 orchestration，不再把 identity / payload 细节摊在主流程里
- “先生成 ID 再校验”的 3 条特殊 create 路径共用同一个 Web 本地 helper
- Web create draft 的公共字段组装有且只有一个本地 helper
- 没有任何一步触碰 `HandleConfig()` 或 publish 逻辑
