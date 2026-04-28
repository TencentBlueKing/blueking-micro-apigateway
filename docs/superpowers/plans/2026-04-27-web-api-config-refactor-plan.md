# Web API Config 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

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

- Task 0 至少覆盖 4 类现状：`CheckAPISIXConfig()` 的 identity fallback；`CheckAPISIXConfig()` 的 validation payload 整形；特殊 3 条 create handler 的“先生成 ID 再校验”；默认 create handler 的“先校验再生成 ID，但 draft 组装一致”。
- `create_handlers_test.go` 负责锁 handler 入口行为；`web_create_helpers_test.go` 只在 helper 抽出后再补第二层单测。
- 如果 Task 0 还没建好，不要直接推进 Task 3-5；否则后面很难分清是 handler 入口行为变了，还是 helper 自己写错了。

### Task 0: 补 Web serializer / create handler characterization tests

- [ ] Task 0: 补 Web serializer / create handler characterization tests

**要解决的缺口：** 现在文档只在说明里提了 Task 0，但正文还没有一个真正独立的前置补测任务。先把 `CheckAPISIXConfig()` 和 create handler 入口行为锁住，后面的 helper 提取才不会把“重构”和“补洞”混在一起。

**为什么这个任务适合单独提 PR：** 只新增 serializer / handler 的 characterization tests，不改 `pkg/apis/web` 的生产逻辑。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/serializer/common_test.go`
- Create: `src/apiserver/pkg/apis/web/handler/create_handlers_test.go`

- [ ] **Step 1: 在现有 serializer 和 create handler seam 上补 characterization tests**

至少覆盖下面 4 类现状，避免直接从 `web_create_helpers_test.go` 或新 helper 起手：

- `CheckAPISIXConfig()` 的 identity fallback
- `CheckAPISIXConfig()` 的 validation payload 整形
- `PluginConfigCreate`、`ConsumerGroupCreate`、`GlobalRuleCreate` 这 3 条特殊 create 路径的“先生成 ID 再校验”
- 至少一条默认 create handler（如 `RouteCreate`）的“先校验再生成 ID，但 draft 组装一致”

- [ ] **Step 2: 运行 Web seam tests，确认入口行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer ./pkg/apis/web/handler -count=1
```

Expected:
- PASS

- [ ] **Step 3: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/serializer/common_test.go src/apiserver/pkg/apis/web/handler/create_handlers_test.go
git commit -m "test: lock web serializer and create handler seams"
```

---

### Task 1: 抽出 `CheckAPISIXConfig()` 的 identity 决策 helper

- [ ] Task 1: 抽出 `CheckAPISIXConfig()` 的 identity 决策 helper

**要解决的复杂度：** `CheckAPISIXConfig()` 同时负责“读 config 识别 identity”和“继续做 schema 校验”，identity 决策散在函数中间，后续规则变化容易改漏。

**为什么这个任务适合单独提 PR：** 只碰 `serializer/common.go` 和对应测试，不改变 handler 行为，也不涉及跨文件迁移。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/serializer/common.go:49-127`
- Modify: `src/apiserver/pkg/apis/web/serializer/common_test.go`

- [ ] **Step 1: 先补当前 identity 决策的测试**

在 `common_test.go` 增加下面这组失败测试：

```go
func TestResolveWebValidationIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input webValidationInput
		want  string
	}{
		{
			name: "consumer falls back to username",
			input: webValidationInput{
				ResourceType: constant.Consumer,
				RawConfig:    json.RawMessage(`{"plugins":{}}`),
				Username:     "demo-user",
				Name:         "ignored",
			},
			want: "demo-user",
		},
		{
			name: "route falls back to outer name",
			input: webValidationInput{
				ResourceType: constant.Route,
				RawConfig:    json.RawMessage(`{"plugins":{}}`),
				Name:         "route-a",
			},
			want: "route-a",
		},
		{
			name: "existing config id wins",
			input: webValidationInput{
				ResourceType: constant.Route,
				RawConfig:    json.RawMessage(`{"id":"route-fixed","plugins":{}}`),
				Name:         "route-a",
			},
			want: "route-fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveWebValidationIdentity(tt.input)
			if got != tt.want {
				t.Fatalf("unexpected identity: got %q want %q", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试，确认它先失败**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -run TestResolveWebValidationIdentity -count=1
```

Expected:
- FAIL，报 `undefined: webValidationInput` 或 `undefined: resolveWebValidationIdentity`

- [ ] **Step 3: 用最小实现抽出 identity helper，并接回 `CheckAPISIXConfig()`**

在 `common.go` 里新增本地输入结构和 helper：

```go
type webValidationInput struct {
	ResourceType constant.APISIXResource
	Version      constant.APISIXVersion
	ResourceID   string
	Name         string
	Username     string
	RawConfig    json.RawMessage
}

func resolveWebValidationIdentity(input webValidationInput) string {
	if identity := schema.GetResourceIdentification(input.RawConfig); identity != "" {
		return identity
	}
	if input.ResourceType == constant.Consumer {
		return input.Username
	}
	return input.Name
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
input := webValidationInput{
	ResourceType: resourceType,
	Version:      gatewayInfo.GetAPISIXVersionX(),
	ResourceID:   fl.Parent().FieldByName("ID").String(),
	Name:         fl.Parent().FieldByName("Name").String(),
	Username:     fl.Parent().FieldByName("Username").String(),
	RawConfig:    rawConfig,
}
resourceIdentification := resolveWebValidationIdentity(input)
if resourceIdentification == "" {
	resourceIdentification = getResourceNameByResourceType(resourceTypeName, fl)
}
```

- [ ] **Step 4: 运行 serializer 包测试，确认行为不变**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/web/serializer/common.go src/apiserver/pkg/apis/web/serializer/common_test.go
git commit -m "refactor: extract web validation identity helper"
```

### Task 2: 抽出 Web 本地 validation payload helper

- [ ] Task 2: 抽出 Web 本地 validation payload helper

**要解决的复杂度：** 现在 `id` 注入、`name/username` 注入、`plugin_metadata.id = name` 都埋在 `CheckAPISIXConfig()` 主流程里，修改时必须通读整段 validator。

**为什么这个任务适合单独提 PR：** 仍然只碰 `serializer/common.go` 和测试；这是在 Task 1 的基础上继续把剩余 payload 逻辑从 validator 主流程中拿出来。

**Files:**
- Modify: `src/apiserver/pkg/apis/web/serializer/common.go:49-127`
- Modify: `src/apiserver/pkg/apis/web/serializer/common_test.go`

- [ ] **Step 1: 先补当前 payload 整形矩阵测试**

在 `common_test.go` 里新增：

```go
func TestPrepareWebValidationPayload(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		input webValidationInput
		want string
	}{
		{
			name: "consumer group injects generated id on 3.13",
			input: webValidationInput{
				ResourceType: constant.ConsumerGroup,
				Version:      constant.APISIXVersion313,
				ResourceID:   "cg-generated-id",
				Name:         "cg-demo",
				RawConfig:    json.RawMessage(`{"plugins":{}}`),
			},
			want: `{"plugins":{},"id":"cg-generated-id","name":"cg-demo"}`,
		},
		{
			name: "proto on 3.11 keeps name out of payload",
			input: webValidationInput{
				ResourceType: constant.Proto,
				Version:      constant.APISIXVersion311,
				Name:         "proto-demo",
				RawConfig:    json.RawMessage(`{"content":"syntax = \\\"proto3\\\";"}`),
			},
			want: `{"content":"syntax = \"proto3\";"}`,
		},
		{
			name: "plugin metadata uses outer name as id",
			input: webValidationInput{
				ResourceType: constant.PluginMetadata,
				Version:      constant.APISIXVersion313,
				Name:         "cors",
				RawConfig:    json.RawMessage(`{"log_format":{"client_ip":"$remote_addr"}}`),
			},
			want: `{"log_format":{"client_ip":"$remote_addr"},"id":"cors"}`,
		},
		{
			name: "ssl never injects name",
			input: webValidationInput{
				ResourceType: constant.SSL,
				Version:      constant.APISIXVersion313,
				Name:         "ssl-demo",
				RawConfig:    json.RawMessage(`{"cert":"demo","key":"demo","snis":["demo.com"]}`),
			},
			want: `{"cert":"demo","key":"demo","snis":["demo.com"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareWebValidationPayload(tt.input)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -run TestPrepareWebValidationPayload -count=1
```

Expected:
- FAIL，报 `undefined: prepareWebValidationPayload`

- [ ] **Step 3: 提取 payload builder，并让 `CheckAPISIXConfig()` 只做 orchestration**

在 `common.go` 增加：

```go
func prepareWebValidationPayload(input webValidationInput) json.RawMessage {
	rawConfig := injectGeneratedIDForValidation(
		input.RawConfig,
		input.ResourceType,
		input.Version,
		input.ResourceID,
	)

	identity := resolveWebValidationIdentity(webValidationInput{
		ResourceType: input.ResourceType,
		Version:      input.Version,
		ResourceID:   input.ResourceID,
		Name:         input.Name,
		Username:     input.Username,
		RawConfig:    rawConfig,
	})

	if identity != "" && shouldInjectResourceNameForValidation(input.ResourceType, input.Version) {
		rawConfig, _ = sjson.SetBytes(rawConfig, model.GetResourceNameKey(input.ResourceType), identity)
	}
	if input.ResourceType == constant.PluginMetadata {
		rawConfig, _ = sjson.SetBytes(rawConfig, "id", input.Name)
	}
	return rawConfig
}
```

然后把 `CheckAPISIXConfig()` 前半段收敛成：

```go
input := webValidationInput{
	ResourceType: resourceType,
	Version:      gatewayInfo.GetAPISIXVersionX(),
	ResourceID:   fl.Parent().FieldByName("ID").String(),
	Name:         fl.Parent().FieldByName("Name").String(),
	Username:     fl.Parent().FieldByName("Username").String(),
	RawConfig:    rawConfig,
}
rawConfig = prepareWebValidationPayload(input)
resourceIdentification := resolveWebValidationIdentity(webValidationInput{
	ResourceType: input.ResourceType,
	Version:      input.Version,
	ResourceID:   input.ResourceID,
	Name:         input.Name,
	Username:     input.Username,
	RawConfig:    rawConfig,
})
```

- [ ] **Step 4: 运行 serializer 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/web/serializer -count=1
```

Expected:
- PASS

- [ ] **Step 5: 提交这个 PR**

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

在 `web_create_helpers_test.go` 里新增：

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

同文件再补 `consumer_group`、`global_rule` 两个子测试，断言点保持一致：`err == nil` 且 `req.ID` 已经被填充。

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

在 `web_create_helpers.go` 增加：

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
