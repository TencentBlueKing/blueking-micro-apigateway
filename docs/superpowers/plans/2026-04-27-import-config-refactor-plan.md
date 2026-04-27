# Import Config 小步重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在保留 `import.ignore_fields` 本地语义不变的前提下，把 import 链路里当前混在 `handleResources(...)` 和 `HandleUploadResources(...)` 里的 overlay、旧资源装载、sync-data 组装、校验前准备几个步骤拆开，使 import 的本地复杂度降下来。

**Architecture:** 本计划完全承认 import 是一条独立链路，不把 overlay 硬塞进共享逻辑。顺序固定为：先把 overlay 抽成 import 本地 pure helper，再把“装载旧资源”“组装 `GatewaySyncData`”拆开，随后重写 `handleResources(...)` 为更小的 orchestration，最后给 `HandleUploadResources(...)` 引入显式的 import validation seam。

**Tech Stack:** Go, Gin context helper, `gjson` / `sjson`, `testify`, `go test`, `make lint`, `make test`

---

## 范围

- 只处理 `src/apiserver/pkg/apis/common/resource_slz.go`
- 允许把 import 本地 helper 拆到新文件
- 保持 `ignore_fields` 仍然是 import 本地能力

## 非目标

- 不把 import overlay 抽成跨领域公共代码
- 不改 `biz.BuildConfigRawForValidation(...)`
- 不改 `HandleConfig()` 行为
- 不改 open / web / mcp

## 文件结构

- `src/apiserver/pkg/apis/common/resource_slz.go`
  - 保留 import 主 orchestration
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

- 每个任务的第一组测试，必须先打在“重构前已经存在的 seam”上，不能直接从计划中新引入的 helper 开始写测试。
- helper 测试只能作为第二层测试：
  - 第一层：先锁定 `handleResources(...)` / `HandleUploadResources(...)` 这些现有 import 入口的行为
  - 第二层：helper 抽出后再补 helper 单测
- Import 计划里的现有 seam 优先级如下：
  - Task 1-4：优先测现有 `handleResources(...)`
  - Task 5：优先测现有 `HandleUploadResources(...)`
- 只有当第一层 seam 测试已经锁住行为时，才允许在同一个 PR 里为 `apply...` / `load...` / `build...` / `prepare...` helper 增加第二层单测。
- 执行时，如果任务正文里的示例代码先写了 helper 测试，应按上面的 seam 规则落地：先补现有 seam 的 characterization test，再补 helper test。

---

### Task 1: 抽出 import 本地 `ignore_fields` overlay helper

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

**要解决的复杂度：** `handleResources(...)` 每轮循环都要自己取 DB 资源、组 map、回填 `allResourceIDs`，这块和 overlay / sync-data 组装混在一起，不利于单测。

**为什么这个任务适合单独提 PR：** 这一步只把 DB 读取和 map 组装从大函数里抽出来，不调整 overlay 语义。

**Files:**
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers.go`
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/common/resource_slz.go:267-275`

- [ ] **Step 1: 先补旧资源装载测试**

在 `import_resource_helpers_test.go` 里新增：

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
}
```

- [ ] **Step 2: 运行测试，确认 helper 还不存在**

Run:

```bash
cd /root/workspace/tx/wklken/blueking-micro-apigateway/src/apiserver && source .envrc && go test ./pkg/apis/common -run TestLoadExistingImportResources -count=1
```

Expected:
- FAIL，报 `undefined: loadExistingImportResources`

- [ ] **Step 3: 把“取 DB 资源 + 组 map + 回填 allResourceIds”抽成 helper**

在 `import_resource_helpers.go` 里新增：

```go
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

然后把 `handleResources(...)` 里原来的 9 行 DB 装载逻辑替换为这个 helper 调用。

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

**要解决的复杂度：** 现在 `HandleUploadResources(...)` 一边准备 add/update map，一边直接调用 `biz.ValidateResource(...)`，没有一个明确的“import 进入 DATABASE 校验前”的本地边界。

**为什么这个任务适合单独提 PR：** 这一步仍然只在 import 域内新增 seam，不会把逻辑抽到共享层。

**Files:**
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers.go`
- Modify: `src/apiserver/pkg/apis/common/import_resource_helpers_test.go`
- Modify: `src/apiserver/pkg/apis/common/resource_slz.go:121-149`

- [ ] **Step 1: 先补 validation input 组装测试**

在 `import_resource_helpers_test.go` 里新增：

```go
func TestPrepareImportValidationInput(t *testing.T) {
	t.Parallel()

	ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ID: 31})

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
