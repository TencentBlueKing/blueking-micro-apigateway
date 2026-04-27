# Open API Config 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变 Open API 外部协议的前提下，逐步收敛 `open api` 当前分散在 middleware、serializer、handler 三处的 `config` 整形、draft 组装和 request identity 复算问题。

**Architecture:** 本计划先把 Open 域内部自己能收敛的东西做干净，再为后续共享抽象埋点。顺序固定为：先把 batch create / update 的 draft 组装 helper 化，再把 middleware 的校验 payload 整理成纯函数，接着显式引入 Open 域内的 resolved draft 载体，最后让 middleware 和 serializer 复用同一份 resolved identity，而不是各算各的。

**Tech Stack:** Go, Gin, `gjson` / `sjson`, `testify`, `go test`, `make lint`, `make test`

---

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

- 每个任务的第一组测试，必须先打在“重构前已经存在的 seam”上，不能直接从计划中新引入的 helper 或 context carrier 开始写测试。
- helper / context helper 测试只能作为第二层测试：
  - 第一层：锁定 Open 现有请求路径、middleware 路径、serializer 路径
  - 第二层：在 helper 抽出后补 helper 单测
- Open 计划里的现有 seam 优先级如下：
  - Task 1：优先测现有 `ResourceBatchCreateRequest.ToCommonResource(...)`
  - Task 2：优先测现有 `ResourceUpdateRequest.ToCommonResource(...)`
  - Task 3：优先测现有 `OpenAPIResourceCheck()` middleware 行为
  - Task 4-5：优先测现有 `ResourceBatchCreate` handler 与 middleware 串起来的路径，确认校验和持久化使用的是同一份 request identity
- 执行时，如果任务正文里的示例代码先写了 helper 测试，应按上面的 seam 规则落地：先补现有 seam 的 characterization test，再补 helper test。

---

### Task 1: 抽出 Open batch create 的本地 draft builder

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

在 `resource.go` 里新增：

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

### Task 2: 抽出 Open update 的本地 draft builder

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

在 `resource.go` 增加：

```go
func buildOpenUpdateDraft(
	c *gin.Context,
	resourceID string,
	status constant.ResourceStatus,
	config json.RawMessage,
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

然后把 `ResourceUpdateRequest.ToCommonResource(...)` 改成直接返回这个 helper 的结果。

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

### Task 3: 抽出 Open middleware 的 validation payload helper

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

在 `open_resolved_draft_context.go` 里新增：

```go
type OpenResolvedDraft struct {
	ID               string
	Name             string
	ValidationConfig json.RawMessage
	StorageConfig    json.RawMessage
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

1. 在 middleware 校验循环里为每个请求项生成 `OpenResolvedDraft`

```go
drafts = append(drafts, serializer.OpenResolvedDraft{
	ID:               gjson.GetBytes(configRawForValidation, "id").String(),
	Name:             config.Get("name").String(),
	ValidationConfig: configRawForValidation,
	StorageConfig:    []byte(configRaw),
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

3. 在 `ToCommonResource(...)` 里优先消费 middleware 放进来的 draft：

```go
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
```

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
