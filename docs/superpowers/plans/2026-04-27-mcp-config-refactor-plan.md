# MCP Config 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **Execution rule:** If a task or step is done, mark it in this `plan.md` before running `git add` and `git commit`.

**Goal:** 在维持“MCP 暂不纳入主线重构目标”这一共识不变的前提下，只做 MCP 域内的小步整理：固化创建与更新路径的当前行为，提取极小的本地 helper，避免这条链路继续靠手写重复逻辑生长。

**Architecture:** MCP 这份计划刻意保持小。它不试图对齐 Web/Open/Import，也不尝试抽跨域共享 builder。顺序固定为：先收拢 create 的 config 注入，再收拢 update 的 config 注入，最后把 create/update 共用的 `ResourceCommonModel` 组装放进 MCP 本地 helper，并用注释明确 MCP 为什么仍然保持本地实现。

**Tech Stack:** Go, MCP SDK, Gin context helper, `testify`, `go test`, `make lint`, `make test`

---

## 代码复核结论

- 重构目的判断：Task 1-2 是正确的，本质上是在收拢 create/update 中必须存在的 config 注入逻辑；`biz.CreateResource(...)` / `biz.UpdateResourceByTypeAndID(...)` 不会替 handler 补这一步。
- 复杂度评估：Task 1-2 为低到中等；真正缺口在于当前 `resource_crud_test.go` 还没有直接覆盖 `createResourceHandler(...)` / `updateResourceHandler(...)`。Task 3 的收益明显低于 Task 1-2。
- 本次修正：把 handler characterization tests 提升为独立前置阶段；Task 3 降为低优先级可选整理，而不是默认必做项。

## 执行顺序（修订）

1. Task 0：独立补 MCP create/update handler characterization tests。
2. Task 1：收拢 create 的 config 注入。
3. Task 2：收拢 update 的 config 注入。
4. Task 3：仅当 Task 1-2 完成后，create handler 仍然因为 draft 组装显著影响可读性时再执行。

## 范围

- 只处理 `src/apiserver/pkg/apis/mcp/tools/resource_crud.go`
- 允许新增 `pkg/apis/mcp/tools` 域内 helper 文件
- 允许给 MCP 本地 helper 增加注释，明确边界

## 非目标

- 不把 MCP 接入本轮跨领域抽象
- 不改 MCP 协议
- 不改 `biz.CreateResource(...)` / `biz.UpdateResourceByTypeAndID(...)` 的更大契约

## 文件结构

- `src/apiserver/pkg/apis/mcp/tools/resource_crud.go`
  - MCP create/update handler 主流程
- `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers.go`
  - MCP 本地 config 注入与 draft 组装 helper
- `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers_test.go`
  - MCP 本地 helper 的行为测试
- `src/apiserver/pkg/apis/mcp/tools/resource_crud_test.go`
  - create/update handler 的 characterization 测试

## PR 出口要求

- 每个任务里的 `go test` 是最小验收命令
- 每个任务准备合并前，再补跑一次：

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && make lint && make test
```

## 测试策略（必须）

- 新增 `Task 0` 作为独立步骤或独立 PR；在 `Task 0` 合并前，不开始 Task 1-3。
- 每个任务的第一组测试，必须先打在“重构前已经存在的 seam”上，不能直接从计划中新引入的 helper 开始写测试。
- helper 测试只能作为第二层测试：
  - 第一层：先锁定 `createResourceHandler(...)` / `updateResourceHandler(...)` 的现有行为
  - 第二层：helper 抽出后补 helper 单测，确保本地整理不改变 MCP 语义
- 当前 `resource_crud_test.go` 里还没有 handler characterization tests；Task 0 需要先把 create/update 的真实调用路径补进去，而不是直接从 `prepare...` / `build...` helper 开始。
- MCP 计划里的现有 seam 优先级如下：
  - Task 0：优先测现有 `createResourceHandler(...)` / `updateResourceHandler(...)`，覆盖 name/username 注入、create_draft/update_draft 状态、以及 config 与 typed name 的一致性
  - Task 1：优先测现有 `createResourceHandler(...)`
  - Task 2：优先测现有 `updateResourceHandler(...)`
  - Task 3：只有在决定执行时，才继续优先测现有 `createResourceHandler(...)`，验证 draft 组装提取前后一致
- 执行时，如果任务正文里的示例代码先写了 helper 测试，应按上面的 seam 规则落地：先补现有 seam 的 characterization test，再补 helper test。

## 重构前测试前置阶段（独立）

- Task 0 至少覆盖 5 类现状：create 会把 outer `name` 写入 `config`；consumer create 写的是 `username`；update 在提供 `name` 时会同步更新 `config`；不提供 `name` 时保留原有 config 形态；**update 在 `input.Config` 为非法 JSON 时当前静默吞掉 `sjson.SetBytes` 的错误（`_ = err` 模式），这是 helper 抽出后会变成 `return err` 的 side-effect 行为，必须在 Task 0 明确锁住当前静默行为，Task 2 落地时再显式记录这一行为差异。**
- Task 0 建议直接扩展 `resource_crud_test.go`，不要把第一批断言写到 `mcp_resource_crud_helpers_test.go`。
- 如果 Task 0 跑完后发现 Task 3 只是在抽一层结构体字面量，而没有显著降低 handler 复杂度，可以直接把 Task 3 标记为 defer。

**Task 0 需要显式列出的 side-effect 不变量（review 补充）：**
- update 路径原本使用 `config, _ = sjson.SetBytes(config, nameKey, input.Name)` 静默忽略错误；Task 2 的 helper 会把这里改成 `return err`，这是**行为变化**，必须在 Task 0 特征测试里锁住“update 传入非法 JSON config 时当前不会报错”的现状，Task 2 落地后允许更新该断言为“会报错”。
- Task 1 helper 里的 `gjson.Exists` 兜底判断来自原 handler 的防御代码，helper 抽出后保留同样行为。

### Task 0: 补 MCP create/update handler characterization tests

- [x] Task 0: 补 MCP create/update handler characterization tests

**要解决的缺口：** 现在文档里已经明确了 Task 0 的必要性，但正文还没有真正把 `createResourceHandler(...)` / `updateResourceHandler(...)` 的现状测试独立出来。先锁 handler 入口，后面的 helper 才有稳定边界。

**为什么这个任务适合单独提 PR：** 只扩 `resource_crud_test.go`，不改 `pkg/apis/mcp/tools` 的生产代码。

**Files:**
- Modify: `src/apiserver/pkg/apis/mcp/tools/resource_crud_test.go`

- [x] **Step 1: 直接在 handler seam 上补 characterization tests**

至少覆盖下面 5 类现状，断言都走真实 create/update handler，而不是新 helper：

- route create 会把 outer `name` 写回 `config.name`
- consumer create 会把 outer `name` 写到 `config.username`
- update 在提供 `name` 时会同步更新 `config` 中对应 typed name
- update 在不提供 `name` 时保留原有 config 形态不变
- **update 路径传入“非法 JSON config” / 会让 sjson 写入失败的入参时，当前 handler 不会向调用者返回错误（`_ = err` 静默路径）；锁住这条断言后，Task 2 helper 抽出会把它变成 `return err`，届时同步更新此断言（记录行为变化）**

- [x] **Step 2: 运行 MCP seam tests，确认当前 handler 行为已经被锁住**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/mcp/tools -run 'Test.*ResourceHandler' -count=1
```

Expected:
- PASS

- [x] **Step 3: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/mcp/tools/resource_crud_test.go
git commit -m "test: lock mcp resource handler seams"
```

---

### Task 1: 抽出 MCP create 的 config 注入 helper

- [x] Task 1: 抽出 MCP create 的 config 注入 helper

**要解决的复杂度：** `createResourceHandler(...)` 现在把“marshal config”“按资源类型写入 name/username”“校验写入结果”都内联在主流程里，create 逻辑很快就会继续膨胀。

**为什么这个任务适合单独提 PR：** 只影响 create 路径，不碰 update，也不动 biz 层。

**Files:**

- Create: `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers.go`
- Create: `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/resource_crud.go:287-303`

- [x] **Step 1: 先补 MCP create config 注入的失败测试**

在 `mcp_resource_crud_helpers_test.go` 里新增：

```go
func TestPrepareMCPCreateConfig(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name         string
        resourceType constant.APISIXResource
        inputConfig  any
        nameValue    string
        assertConfig func(t *testing.T, config []byte)
    }{
        {
            name:         "route injects name",
            resourceType: constant.Route,
            inputConfig:  map[string]any{"uri": "/demo"},
            nameValue:    "route-demo",
            assertConfig: func(t *testing.T, config []byte) {
                assert.Equal(t, "route-demo", gjson.GetBytes(config, "name").String())
                assert.Equal(t, "/demo", gjson.GetBytes(config, "uri").String())
            },
        },
        {
            name:         "consumer injects username",
            resourceType: constant.Consumer,
            inputConfig:  map[string]any{"plugins": map[string]any{}},
            nameValue:    "consumer-demo",
            assertConfig: func(t *testing.T, config []byte) {
                assert.Equal(t, "consumer-demo", gjson.GetBytes(config, "username").String())
                assert.Empty(t, gjson.GetBytes(config, "name").String())
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config, err := prepareMCPCreateConfig(tt.resourceType, tt.inputConfig, tt.nameValue)
            assert.NoError(t, err)
            tt.assertConfig(t, config)
        })
    }
}
```

- [x] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/mcp/tools -run TestPrepareMCPCreateConfig -count=1
```

Expected:

- FAIL，报 `undefined: prepareMCPCreateConfig`

- [x] **Step 3: 实现 helper，并让 create handler 复用**

在 `mcp_resource_crud_helpers.go` 里新增（注意 doc comment 明确标注 defensive check 的来源）：

```go
// prepareMCPCreateConfig marshals the inbound MCP create payload and injects the
// outer `name` field according to the resource type (route/service/upstream use
// "name", consumer uses "username" — see model.GetResourceNameKey).
//
// The trailing gjson.Exists check is preserved from the original inline handler
// implementation as a defensive guard; sjson.SetBytes should normally not leave
// the key missing, but the behavior is kept to avoid silently introducing a
// regression when the helper is adopted by the handler.
func prepareMCPCreateConfig(
    resourceType constant.APISIXResource,
    inputConfig any,
    name string,
) ([]byte, error) {
    config, err := json.Marshal(inputConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal config: %w", err)
    }

    nameKey := model.GetResourceNameKey(resourceType)
    config, err = sjson.SetBytes(config, nameKey, name)
    if err != nil {
        return nil, fmt.Errorf("failed to inject name into config: %w", err)
    }
    // preserved defensive check from original handler (resource_crud.go prior to refactor)
    if !gjson.GetBytes(config, nameKey).Exists() {
        return nil, fmt.Errorf("name field not found in config after injection")
    }
    return config, nil
}
```

**额外测试补充（review 要求）：** 在 `TestPrepareMCPCreateConfig` 里再加一条子测试，断言当 `nameKey` 被显式清空后 helper 能返回 “name field not found in config after injection” 错误，保证这条 defensive branch 不会随时间沉默成死代码。

然后把 `createResourceHandler(...)` 中的：

```go
config, err := json.Marshal(input.Config)
...
nameKey := model.GetResourceNameKey(resourceType)
config, err = sjson.SetBytes(config, nameKey, input.Name)
...
if !gjson.GetBytes(config, nameKey).Exists() { ... }
```

替换成：

```go
config, err := prepareMCPCreateConfig(resourceType, input.Config, input.Name)
if err != nil {
    return errorResult(err), nil, nil
}
```

- [x] **Step 4: 运行 MCP tools 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/mcp/tools -count=1
```

Expected:

- PASS

- [x] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers.go src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers_test.go src/apiserver/pkg/apis/mcp/tools/resource_crud.go
git commit -m "refactor: extract mcp create config helper"
```

### Task 2: 抽出 MCP update 的 config 注入 helper

- [x] Task 2: 抽出 MCP update 的 config 注入 helper

**要解决的复杂度：** update 路径和 create 路径一样，也在 handler 里手写了 `marshal + optional name inject`，只是分支略有不同；这类重复最容易导致两个入口后续越改越不一致。

**为什么这个任务适合单独提 PR：** 只处理 update 路径，保持 MCP 小步收敛。

**Files:**

- Modify: `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/resource_crud.go:366-376`

- [x] **Step 1: 先补 MCP update config 注入测试**

在 `mcp_resource_crud_helpers_test.go` 里新增（注意：Consumer 使用 `username` 而非 `name`，补子测试锁住差异）：

```go
func TestPrepareMCPUpdateConfig(t *testing.T) {
    t.Parallel()

    t.Run("route inject name when provided", func(t *testing.T) {
        config, err := prepareMCPUpdateConfig(
            constant.Route,
            map[string]any{"uri": "/demo"},
            "route-demo",
        )
        assert.NoError(t, err)
        assert.Equal(t, "route-demo", gjson.GetBytes(config, "name").String())
        assert.Equal(t, "/demo", gjson.GetBytes(config, "uri").String())
    })

    t.Run("route keeps config untouched when name is empty", func(t *testing.T) {
        config, err := prepareMCPUpdateConfig(
            constant.Route,
            map[string]any{"uri": "/demo"},
            "",
        )
        assert.NoError(t, err)
        assert.Equal(t, "/demo", gjson.GetBytes(config, "uri").String())
        assert.False(t, gjson.GetBytes(config, "name").Exists())
    })

    // review 补充：consumer 使用 username，不使用 name
    t.Run("consumer injects username when name is provided", func(t *testing.T) {
        config, err := prepareMCPUpdateConfig(
            constant.Consumer,
            map[string]any{"plugins": map[string]any{}},
            "consumer-demo",
        )
        assert.NoError(t, err)
        assert.Equal(t, "consumer-demo", gjson.GetBytes(config, "username").String())
        assert.False(t, gjson.GetBytes(config, "name").Exists())
    })
}
```

- [x] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/mcp/tools -run TestPrepareMCPUpdateConfig -count=1
```

Expected:

- FAIL，报 `undefined: prepareMCPUpdateConfig`

- [x] **Step 3: 实现 helper，并让 update handler 复用**

在 `mcp_resource_crud_helpers.go` 里新增（**行为变化标记**：helper 把原 handler 里 `_ = err` 的静默错误改为向上 `return err`）：

```go
// prepareMCPUpdateConfig mirrors prepareMCPCreateConfig but with an early return
// when the caller did not supply a new outer name, matching the existing
// updateResourceHandler branching.
//
// Behavior change vs. the previous inline code in updateResourceHandler:
// sjson.SetBytes errors are now returned to the caller rather than silently
// swallowed via `config, _ = sjson.SetBytes(...)`. Task 0 locks the previous
// silent behavior; the Task 2 assertion is updated at the same commit that
// lands this helper.
func prepareMCPUpdateConfig(
    resourceType constant.APISIXResource,
    inputConfig any,
    name string,
) ([]byte, error) {
    config, err := json.Marshal(inputConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal config: %w", err)
    }
    if name == "" {
        return config, nil
    }

    nameKey := model.GetResourceNameKey(resourceType)
    config, err = sjson.SetBytes(config, nameKey, name)
    if err != nil {
        return nil, fmt.Errorf("failed to inject name into config: %w", err)
    }
    return config, nil
}
```

然后把 `updateResourceHandler(...)` 中的：

```go
config, err := json.Marshal(input.Config)
...
if input.Name != "" {
    nameKey := model.GetResourceNameKey(resourceType)
    config, _ = sjson.SetBytes(config, nameKey, input.Name)
}
```

替换成：

```go
config, err := prepareMCPUpdateConfig(resourceType, input.Config, input.Name)
if err != nil {
    return errorResult(err), nil, nil
}
```

**提交前同步更新 Task 0 中 “update 静默忽略 sjson 错误” 的 characterization 断言**：由“不会报错”更新为“会返回错误”，并在 commit message 里标注“behavior change: sjson failure during update now surfaces as an error”。

- [x] **Step 4: 运行 MCP tools 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/mcp/tools -count=1
```

Expected:

- PASS

- [x] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers.go src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers_test.go src/apiserver/pkg/apis/mcp/tools/resource_crud.go
git commit -m "refactor: extract mcp update config helper"
```

### Task 3: 抽出 MCP create draft 组装 helper，并明确 MCP 保持本地的边界

- [x] Task 3: 抽出 MCP create draft 组装 helper，并明确 MCP 保持本地的边界

**要解决的复杂度：** create handler 里仍然手写了 `ResourceCommonModel` 组装；如果不把这层也收掉，MCP 依旧是第四套手工 draft builder。

**为什么这个任务适合单独提 PR：** 这是 MCP 计划里最后一个本地整理动作，改动仍然局限在 `pkg/apis/mcp/tools`。

**Files:**

- Modify: `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/resource_crud.go:305-323`

- [x] **Step 1: 先补 draft 组装测试**

在 `mcp_resource_crud_helpers_test.go` 里新增（**review 补充**：至少覆盖 route 和 consumer 两个资源类型，在多资源类型上同时锁住 `Creator/Updater="mcp"` + `Status=create_draft` 这 3 个不变量）：

```go
func TestBuildMCPCreateDraft(t *testing.T) {
    t.Parallel()

    t.Run("route", func(t *testing.T) {
        config := []byte(`{"name":"route-demo","uri":"/demo"}`)
        got := buildMCPCreateDraft(17, "route-id", config)

        assert.Equal(t, "route-id", got.ID)
        assert.Equal(t, 17, got.GatewayID)
        assert.Equal(t, constant.ResourceStatusCreateDraft, got.Status)
        assert.Equal(t, "mcp", got.Creator)
        assert.Equal(t, "mcp", got.Updater)
        assert.JSONEq(t, `{"name":"route-demo","uri":"/demo"}`, string(got.Config))
    })

    t.Run("consumer", func(t *testing.T) {
        config := []byte(`{"username":"consumer-demo","plugins":{}}`)
        got := buildMCPCreateDraft(17, "consumer-id", config)

        assert.Equal(t, "consumer-id", got.ID)
        assert.Equal(t, 17, got.GatewayID)
        assert.Equal(t, constant.ResourceStatusCreateDraft, got.Status)
        assert.Equal(t, "mcp", got.Creator)
        assert.Equal(t, "mcp", got.Updater)
        assert.JSONEq(t, `{"username":"consumer-demo","plugins":{}}`, string(got.Config))
    })
}
```

- [x] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/mcp/tools -run TestBuildMCPCreateDraft -count=1
```

Expected:

- FAIL，报 `undefined: buildMCPCreateDraft`

- [x] **Step 3: 实现 helper，迁移 create handler，并加上 MCP 本地边界注释**

在 `mcp_resource_crud_helpers.go` 中补充（**review 补充**：注释中明确 MCP helper 签名不要向 web `buildWebCreateDraft(c *gin.Context, ...)` 对齐——MCP 没有 gin.Context，保持 `gatewayID int`）：

```go
// MCP stays local for now: we only deduplicate MCP's own config prep and draft assembly.
// It does not join the cross-domain abstraction track unless a later change proves very high leverage.
//
// Intentional signature drift vs. web's buildWebCreateDraft(c *gin.Context, ...): MCP is
// invoked from an MCP tool handler (no gin.Context in scope), so the helper takes
// gatewayID directly. Do NOT align this signature with the web helper.
func buildMCPCreateDraft(
    gatewayID int,
    resourceID string,
    config []byte,
) model.ResourceCommonModel {
    return model.ResourceCommonModel{
        ID:        resourceID,
        GatewayID: gatewayID,
        Config:    datatypes.JSON(config),
        Status:    constant.ResourceStatusCreateDraft,
        BaseModel: model.BaseModel{
            Creator: "mcp",
            Updater: "mcp",
        },
    }
}
```

然后把 `createResourceHandler(...)` 中的内联组装替换为：

```go
resource := buildMCPCreateDraft(gateway.ID, resourceID, config)
specificResource := resource.ToResourceModel(resourceType)
```

这一步只收敛 draft 组装，不顺手改 `biz.CreateResource(...)` 的更大契约。

- [x] **Step 4: 运行 MCP tools 包测试**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/mcp/tools -count=1
```

Expected:

- PASS

- [x] **Step 5: 提交这个 PR**

```bash
git add src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers.go src/apiserver/pkg/apis/mcp/tools/mcp_resource_crud_helpers_test.go src/apiserver/pkg/apis/mcp/tools/resource_crud.go
git commit -m "refactor: extract mcp draft builder"
```

## 完成定义

- MCP create / update 的 config 注入逻辑各自有清晰的本地 helper
- MCP create draft 的组装不再内联在 handler 里
- 代码里明确写出：MCP 目前只做本地去重，不进入本轮跨领域抽象主线
