# APISIX 3.17 Data Plane Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为新建的 `apisix 3.17.0` 与 `bk-apisix 3.17.0` 网关提供独立、完整且可验证的资源 Schema、插件 Schema、插件目录、发布校验和 MCP 纳管能力。

**Architecture:** 继续使用当前“完整版本转 `3.17.X`、按版本选择 embed 快照”的机制，不修改接入探测、Web/Open API、导入、同步和发布主流程。3.17 使用固定 APISIX 3.17.0 和 BK-APISIX 3.17 数据面导出的独立资产；官方、TAPISIX、BK 插件继续按 3.13 的组合和 Schema fallback 顺序工作。

**Tech Stack:** Go 1.24、`go:embed`、gjson、gojsonschema、Gin validator、Vue 3、Node.js、AJV 8.17.1、Bruno、Docker Compose。

## Global Constraints

- 官方 APISIX 源固定为 `3.17.0` / `9ef2ecab67f652d38365049613610ef649bb4ad0`。
- BK 数据面源固定为 `blueking-apigateway-apisix` 提交 `c1c44e5dc192120ccfa432e9f54703285a80ca38`。
- 对外仅新增 `apisix 3.17.0` 和 `bk-apisix 3.17.0`；不新增 `tapisix 3.17.0` 入口。
- 内部版本常量必须是 `3.17.X`，不得用版本阈值让未来 minor 自动继承 3.17 能力。
- 3.17 必须是完整独立快照；禁止复用 3.13 `schema.json` 或在运行时叠加差异。
- `tapisix_plugin.json` 必须是 `[]`，`tapisix_plugin_schema.json` 必须是 `{"plugins": {}}`，并继续参与组合与 fallback。
- 插件组合保持：`apisix = official`、`tapisix = official + TAPISIX`、`bk-apisix = official + TAPISIX + BK`。
- 不修改 `{etcd_prefix}/data_plane/server_info` 第一条记录探测、版本回填、缺失容忍或版本不一致处理。
- 不支持存量 3.13 网关原地升级，不新增数据库迁移或配置转换器。
- 不新增依赖；前端生产代码仅在 3.17 标准 Schema 无法用当前 AJV 配置编译时才允许最小修改。
- 不修改或提交当前工作树中既有的未跟踪文件，特别是 `src/apiserver/tests/dev/` 和 `src/apiserver/docs/` 下的用户文件。

---

## File Map

**Create**

- `src/apiserver/pkg/utils/schema/3.17/schema.json`：完整官方 3.17 资源、HTTP 插件、consumer、metadata、stream Schema。
- `src/apiserver/pkg/utils/schema/3.17/plugin.json`：控制面开放的官方 3.17 插件及合法示例。
- `src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin.json`：BK 3.17 插件目录及示例。
- `src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin_schema.json`：BK 3.17 插件 Schema。
- `src/apiserver/pkg/utils/schema/3.17/tapisix_plugin.json`：空 TAPISIX 插件目录。
- `src/apiserver/pkg/utils/schema/3.17/tapisix_plugin_schema.json`：空 TAPISIX Schema。
- `src/apiserver/pkg/utils/schema/version_test.go`：对外版本目录断言。
- `src/apiserver/pkg/utils/schema/schema_317_test.go`：3.17 资产完整性、组合、版本隔离和代表性差异测试。
- `src/frontend/scripts/validate-apiserver-schema.mjs`：用前端实际 AJV 配置编译 3.17 Schema 并校验示例。
- `src/apiserver/tests/integration/openapi/plugin_matrix/00_gateways/03_create_gateway_3_17_0.bru`：3.17 测试网关。
- `src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/07_success_3_17_0_consumer.bru`：consumer 代表成功场景。
- `src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/08_success_3_17_0_http.bru`：新增/变更 HTTP 插件代表成功场景。
- `src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/09_success_3_17_0_metadata.bru`：metadata 代表成功场景。
- `src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/10_success_3_17_0_stream.bru`：`traffic-split` stream 成功场景。
- `src/apiserver/tests/integration/openapi/plugin_matrix/03_invalid_cases/151_invalid_3_17_0_cas-auth_missing-cookie.bru`：3.17 新必填字段失败场景。
- `src/apiserver/tests/integration/openapi/plugin_matrix/03_invalid_cases/152_invalid_3_17_0_hmac-auth_max-body-size.bru`：3.17 新数值约束失败场景。

**Modify**

- `src/apiserver/pkg/constant/apisix.go`：注册 `APISIXVersion317`。
- `src/apiserver/pkg/constant/resource_schema.go`：将 3.17 加入 name/id 能力矩阵。
- `src/apiserver/pkg/constant/resource_schema_test.go`：3.17 字段能力测试。
- `src/apiserver/pkg/utils/version/version_test.go`：`3.17.0 -> 3.17.X` 测试。
- `src/apiserver/pkg/utils/schema/version.json`：公开 `3.17.0`。
- `src/apiserver/pkg/utils/schema/plugin.go`：embed 插件目录、更新组合 map、文档 URL 和 stream 映射。
- `src/apiserver/pkg/utils/schema/schema.go`：embed Schema 并更新三个版本 map。
- `src/apiserver/pkg/utils/schema/plugin_test.go`、`schema_test.go`、`validate_test.go`：将 3.17 纳入通用资产测试。
- `src/apiserver/pkg/biz/resourcevalidation/prepare_payload_test.go`：3.17 name/id 注入和清理。
- `src/apiserver/pkg/biz/resourcevalidation/validator_test.go`：3.17 Schema 选择。
- `src/apiserver/pkg/biz/publish/payload_test.go`：3.17 发布 payload 字段处理。
- `src/apiserver/pkg/biz/mcp/access_token.go`、`access_token_test.go`：3.17 MCP 白名单。
- `src/apiserver/pkg/biz/mcp/resource_crud_test.go`：3.17 插件目录查询。
- `src/apiserver/pkg/apis/mcp/tools/common.go`、`common_test.go`：MCP Schema 工具版本解析。
- `src/apiserver/pkg/apis/mcp/tools/schema.go`：工具输入和说明文案。
- `src/apiserver/pkg/apis/mcp/tools/mcp_resource_validation_test.go`：3.17 Schema 校验。
- `src/apiserver/pkg/apis/mcp/prompts/workflows.go`、`resources/docs.go`：支持版本说明和示例。
- `src/frontend/package.json`：增加 `test:schema` 命令，不增加依赖。
- `src/apiserver/AGENTS.md`、`src/apiserver/pkg/apis/mcp/AGENTS.md`、`src/apiserver/pkg/utils/schema/AGENTS.md`：同步支持矩阵、来源和测试规则。

---

### Task 1: Register the 3.17 Version Family

**Files:**

- Create: `src/apiserver/pkg/utils/schema/version_test.go`
- Modify: `src/apiserver/pkg/constant/apisix.go`
- Modify: `src/apiserver/pkg/utils/schema/version.json`
- Modify: `src/apiserver/pkg/utils/version/version_test.go`

**Interfaces:**

- Produces: `constant.APISIXVersion317` with value `3.17.X`.
- Produces: `SupportAPISIXVersionMap["3.17.X"]`.
- Produces: `GetSupportVersionMap()["apisix"|"bk-apisix"]` containing `3.17.0`.

- [ ] **Step 1: Write the failing version tests**

Add `3.17.0 -> constant.APISIXVersion317` to `TestToXVersion`, and create:

```go
func TestGetSupportVersionMapIncludes317(t *testing.T) {
    versions := GetSupportVersionMap()
    assert.Contains(t, versions[constant.APISIXTypeAPISIX].SupportVersion, "3.17.0")
    assert.Contains(t, versions[constant.APISIXTypeBKAPISIX].SupportVersion, "3.17.0")
    _, hasTAPISIX := versions[constant.APISIXTypeTAPISIX]
    assert.False(t, hasTAPISIX)
}
```

- [ ] **Step 2: Run the focused tests and verify failure**

Run from `src/apiserver`:

```bash
go test ./pkg/utils/version ./pkg/utils/schema -run 'TestToXVersion|TestGetSupportVersionMapIncludes317'
```

Expected: compile failure because `APISIXVersion317` is not defined, followed by missing `3.17.0` until `version.json` is updated.

- [ ] **Step 3: Add the version constant and advertised versions**

Add to `apisix.go`:

```go
APISIXVersion317 APISIXVersion = "3.17.X"
```

Add the explicit map entry:

```go
"3.17.X": string(APISIXVersion317),
```

Append `3.17.0` to both `apisix.support_version` and `bk-apisix.support_version` in `version.json`; do not add a `tapisix` key.

- [ ] **Step 4: Run the focused tests and verify pass**

Run:

```bash
go test ./pkg/utils/version ./pkg/utils/schema -run 'TestToXVersion|TestGetSupportVersionMapIncludes317'
```

Expected: PASS.

- [ ] **Step 5: Commit the version registration**

```bash
git add src/apiserver/pkg/constant/apisix.go \
  src/apiserver/pkg/utils/schema/version.json \
  src/apiserver/pkg/utils/schema/version_test.go \
  src/apiserver/pkg/utils/version/version_test.go
git commit -m "feat: register APISIX 3.17 version family"
```

---

### Task 2: Import the Official 3.17 Schema and Plugin Catalog

**Files:**

- Create: `src/apiserver/pkg/utils/schema/3.17/schema.json`
- Create: `src/apiserver/pkg/utils/schema/3.17/plugin.json`
- Modify: `src/apiserver/pkg/utils/schema/schema.go`
- Modify: `src/apiserver/pkg/utils/schema/plugin.go`
- Modify: `src/apiserver/pkg/utils/schema/schema_test.go`
- Modify: `src/apiserver/pkg/utils/schema/plugin_test.go`
- Modify: `src/apiserver/pkg/utils/schema/validate_test.go`

**Interfaces:**

- Consumes: `constant.APISIXVersion317` from Task 1.
- Produces: `schemaVersionMap[APISIXVersion317]`, `versionPluginMap[APISIXVersion317]`, and `VersionDocUrlMap[APISIXVersion317]`.
- Produces: stable paths `main.*`, `plugins.*.(schema|consumer_schema|metadata_schema)`, and `stream_plugins.*.schema`.

- [ ] **Step 1: Write failing official-asset lookup tests**

Add `APISIXVersion317` to the version lists in `plugin_test.go`, `schema_test.go`, and `validate_test.go`. Add explicit assertions in `schema_317_test.go` later in Task 4; here the existing generic tests must already require:

```go
assert.NotNil(t, GetResourceSchema(constant.APISIXVersion317, constant.Route.String()))
assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "jwt-auth", ""))
assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "traffic-split", "stream"))
```

- [ ] **Step 2: Run tests and verify the missing-map failure**

Run:

```bash
go test ./pkg/utils/schema -run 'TestGetResourceSchema|TestGetPluginSchema|TestPluginExamplesMatchSchema'
```

Expected: FAIL because 3.17 has no embedded official assets or map entries.

- [ ] **Step 3: Export the official 3.17.0 runtime Schema**

Use an isolated checkout of Apache APISIX at exactly `9ef2ecab67f652d38365049613610ef649bb4ad0`. Start its local runtime and capture the Control API result:

```bash
git rev-parse HEAD
make init
make run
curl --fail --silent --show-error http://127.0.0.1:9090/v1/schema \
  -o /tmp/apisix-3.17.0-schema.raw.json
make quit
jq -e '.main and .plugins and .stream_plugins' /tmp/apisix-3.17.0-schema.raw.json
```

The first command must print `9ef2ecab67f652d38365049613610ef649bb4ad0`; stop if it does not.

- [ ] **Step 4: Normalize the exported Schema mechanically**

Run from the micro-gateway repository root:

```bash
jq -S '
  walk(
    if type == "object" then
      (if has("properties") and (has("type") | not) then . + {"type":"object"} else . end)
      | del(.encrypt_fields)
    else . end
  )
  | .main.plugin_metadata = {"type":"object"}
' /tmp/apisix-3.17.0-schema.raw.json \
  > src/apiserver/pkg/utils/schema/3.17/schema.json
```

Then assert the project contract:

```bash
jq -e '.main.route and .main.service and .main.upstream and
       .main.consumer and .main.consumer_group and .main.plugin_config and
       .main.global_rule and .main.plugin_metadata and
       .main.proto and .main.ssl and .main.stream_route' \
  src/apiserver/pkg/utils/schema/3.17/schema.json
! rg -n '"encrypt_fields"' src/apiserver/pkg/utils/schema/3.17/schema.json
```

APISIX Control API 不导出 `main.plugin_metadata`，但本项目的 11 资源通用查询测试和资源 Schema API 依赖该稳定路径，因此显式恢复与 3.13 相同的空 object Schema；具体插件 metadata 配置仍由各插件的 `metadata_schema` 校验。

- [ ] **Step 5: Build the 3.17 official plugin catalog**

Use the exact default-plugin set difference from APISIX `apisix/cli/config.lua` at `3.13.0..3.17.0`. Retain every 3.13 catalog plugin that still exists, add the following default-enabled plugins when their Schema is present, and continue excluding `mcp-bridge`, `server-info`, `node-status`, and `log-rotate`:

```text
proxy-buffering
saml-auth
dingtalk-auth
feishu-auth
acl
data-mask
ai-aliyun-content-moderation
graphql-proxy-cache
graphql-limit-count
traffic-label
oas-validator
```

For every entry, set `doc_url` via `VersionDocUrlMap`, keep `type`/`proxy_type` consistent with its scope, and provide the smallest valid `example`, `consumer_example`, and `metadata_example`. Do not copy a 3.13 example for a Schema path changed by the structural diff without revalidating it.

- [ ] **Step 6: Embed and register the official assets**

Add to `schema.go`:

```go
//go:embed 3.17/schema.json
var rawSchemaV317 []byte
```

and:

```go
constant.APISIXVersion317: gjson.ParseBytes(rawSchemaV317),
```

Add to `plugin.go`:

```go
//go:embed 3.17/plugin.json
var rawPluginV317 []byte
```

and register it in `versionPluginMap`. Add:

```go
constant.APISIXVersion317: "https://apisix.apache.org/zh/docs/apisix/plugins/%s/",
```

- [ ] **Step 7: Run official schema and example tests**

Run:

```bash
go test ./pkg/utils/schema -run 'TestGetResourceSchema|TestGetPluginSchema|TestPluginExamplesMatchSchema|TestConsumerPluginFrontendFallbackExamplesMatchConsumerSchema'
```

Expected: PASS for all 3.17 official resources and catalog examples. Fix the 3.17 JSON asset—not the validator or test—when an example fails.

- [ ] **Step 8: Commit the official snapshot**

```bash
git add src/apiserver/pkg/utils/schema/3.17/schema.json \
  src/apiserver/pkg/utils/schema/3.17/plugin.json \
  src/apiserver/pkg/utils/schema/schema.go \
  src/apiserver/pkg/utils/schema/plugin.go \
  src/apiserver/pkg/utils/schema/schema_test.go \
  src/apiserver/pkg/utils/schema/plugin_test.go \
  src/apiserver/pkg/utils/schema/validate_test.go
git commit -m "feat: add APISIX 3.17 schema snapshot"
```

---

### Task 3: Add BK and TAPISIX 3.17 Composition

**Files:**

- Create: `src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin.json`
- Create: `src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin_schema.json`
- Create: `src/apiserver/pkg/utils/schema/3.17/tapisix_plugin.json`
- Create: `src/apiserver/pkg/utils/schema/3.17/tapisix_plugin_schema.json`
- Modify: `src/apiserver/pkg/utils/schema/plugin.go`
- Modify: `src/apiserver/pkg/utils/schema/schema.go`
- Test: `src/apiserver/pkg/utils/schema/schema_317_test.go`

**Interfaces:**

- Consumes: official 3.17 maps from Task 2.
- Produces: `versionBkAPISIXPluginMap[APISIXVersion317]`, `versionTAPISIXPluginMap[APISIXVersion317]`, `bkAPISIXPluginSchemaVersionMap[APISIXVersion317]`, and `tapisixPluginSchemaVersionMap[APISIXVersion317]`.
- Preserves: `GetPlugins()` order official -> TAPISIX -> BK and `GetPluginSchema()` fallback official -> BK -> TAPISIX.

- [ ] **Step 1: Write failing composition tests**

Create `schema_317_test.go` in package `schema` and assert:

```go
func Test317PluginComposition(t *testing.T) {
    official, err := GetPlugins(constant.APISIXTypeAPISIX, constant.APISIXVersion317)
    require.NoError(t, err)
    tapisix, err := GetPlugins(constant.APISIXTypeTAPISIX, constant.APISIXVersion317)
    require.NoError(t, err)
    bk, err := GetPlugins(constant.APISIXTypeBKAPISIX, constant.APISIXVersion317)
    require.NoError(t, err)

    assert.Equal(t, pluginNames(official), pluginNames(tapisix))
    assert.Greater(t, len(bk), len(official))
    assert.Contains(t, pluginNames(bk), "bk-jwt")
    assert.NotContains(t, pluginNames(official), "bk-jwt")
}
```

Add a local `pluginNames([]*Plugin) []string` test helper. Also assert all four map keys exist and the two TAPISIX JSON values decode to the required empty shapes.

- [ ] **Step 2: Run tests and verify missing BK/TAPISIX registrations**

Run:

```bash
go test ./pkg/utils/schema -run 'Test317PluginComposition|Test317EmptyTAPISIXAssets'
```

Expected: FAIL because the 3.17 BK and TAPISIX assets/maps do not exist.

- [ ] **Step 3: Export BK plugin Schema from the pinned BK data plane**

Use an isolated checkout at exactly `c1c44e5dc192120ccfa432e9f54703285a80ca38`. Build and run the image with Control API enabled and only these control-plane-exposed BK plugins loaded:

```text
bk-break-recursive-call
bk-delete-cookie
bk-echo
bk-header-rewrite
bk-jwt
bk-login-required
bk-traffic-label
```

Capture `/v1/schema`, select only those seven keys, and normalize with the same object-type and `encrypt_fields` rules from Task 2:

```bash
curl --fail --silent --show-error http://127.0.0.1:19090/v1/schema \
  -o /tmp/bk-apisix-3.17-schema.raw.json
jq -S '
  {plugins: (.plugins | with_entries(select(.key as $name | [
    "bk-break-recursive-call", "bk-delete-cookie", "bk-echo",
    "bk-header-rewrite", "bk-jwt", "bk-login-required", "bk-traffic-label"
  ] | index($name))))}
  | walk(
      if type == "object" then
        (if has("properties") and (has("type") | not) then . + {"type":"object"} else . end)
        | del(.encrypt_fields)
      else . end
    )
' /tmp/bk-apisix-3.17-schema.raw.json \
  > src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin_schema.json
```

Assert the output contains exactly seven plugin keys before continuing.

- [ ] **Step 4: Create the BK catalog and empty TAPISIX assets**

Create `bk_apisix_plugin.json` with the same seven public BK plugin names, using their 3.17 runtime Schema and minimal valid examples. Create exactly:

```json
[]
```

in `tapisix_plugin.json`, and:

```json
{
  "plugins": {}
}
```

in `tapisix_plugin_schema.json`.

- [ ] **Step 5: Embed all four assets and add traffic-split to stream mapping**

Add 3.17 `go:embed` variables and map entries to `plugin.go` and `schema.go` following the 3.13 names with `V317` suffixes. Add:

```go
"traffic-split": "traffic-split",
```

to `StreamRoutePluginMap` without removing existing entries.

- [ ] **Step 6: Run composition and BK example tests**

Run:

```bash
go test ./pkg/utils/schema -run 'Test317|TestPluginExamplesMatchSchema|TestPluginCanonicalScopeInventory'
```

Expected: PASS, including official-only equality for APISIX/TAPISIX and official-plus-BK ordering for BK-APISIX.

- [ ] **Step 7: Commit BK/TAPISIX composition**

```bash
git add src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin.json \
  src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin_schema.json \
  src/apiserver/pkg/utils/schema/3.17/tapisix_plugin.json \
  src/apiserver/pkg/utils/schema/3.17/tapisix_plugin_schema.json \
  src/apiserver/pkg/utils/schema/plugin.go \
  src/apiserver/pkg/utils/schema/schema.go \
  src/apiserver/pkg/utils/schema/schema_317_test.go
git commit -m "feat: add BK APISIX 3.17 plugin composition"
```

---

### Task 4: Lock Full Schema Integrity and 3.13/3.17 Differences

**Files:**

- Modify: `src/apiserver/pkg/utils/schema/schema_317_test.go`
- Modify when tests expose asset defects: `src/apiserver/pkg/utils/schema/3.17/schema.json`
- Modify when tests expose example defects: `src/apiserver/pkg/utils/schema/3.17/plugin.json`
- Modify when tests expose BK defects: `src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin.json`
- Modify when tests expose BK defects: `src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin_schema.json`

**Interfaces:**

- Consumes: all 3.17 assets and maps from Tasks 2-3.
- Produces: data-driven proof that catalog entries and every example resolve to the correct version/scope Schema.
- Produces: stable representative assertions for known same-name Schema changes.

- [ ] **Step 1: Add bidirectional catalog/Schema coverage tests**

For every official and BK catalog entry:

1. determine main scope from `proxy_type`;
2. require its main Schema;
3. require consumer/metadata Schema when the matching example is present;
4. compile the Schema with gojsonschema;
5. validate the example;
6. reject duplicate plugin names across official/BK/TAPISIX catalogs.

Also walk every `main` resource in `constant.ResourceTypeList`, including the project-required `main.plugin_metadata` placeholder, and compile it. Walk every `plugins.*.(schema|consumer_schema|metadata_schema)` and `stream_plugins.*.schema` node in the three 3.17 Schema sources and compile each node.

- [ ] **Step 2: Add exact representative version-difference assertions**

Encode these fixed `3.13.0..3.17.0` differences:

```go
// jwt-auth main config: new claims_to_verify and realm.
assert.Nil(t, schemaProperty(t, constant.APISIXVersion313, "jwt-auth", "", "claims_to_verify"))
assert.NotNil(t, schemaProperty(t, constant.APISIXVersion317, "jwt-auth", "", "claims_to_verify"))

// jwt-auth consumer: HS384/RS384/PS256/EdDSA are added in 3.17.
assert.NotContains(t, jwtAlgorithms(t, constant.APISIXVersion313), "EdDSA")
assert.Contains(t, jwtAlgorithms(t, constant.APISIXVersion317), "EdDSA")

// hmac-auth main config: new bounded request body setting and realm.
assert.Nil(t, schemaProperty(t, constant.APISIXVersion313, "hmac-auth", "", "max_req_body_size"))
assert.Equal(t, float64(1), schemaMinimum(t, constant.APISIXVersion317, "hmac-auth", "", "max_req_body_size"))

// cas-auth: cookie becomes required and cookie.secret has minLength 32.
assert.NotContains(t, schemaRequired(t, constant.APISIXVersion313, "cas-auth", ""), "cookie")
assert.Contains(t, schemaRequired(t, constant.APISIXVersion317, "cas-auth", ""), "cookie")

// batch-requests metadata and request item constraints.
assert.Nil(t, schemaProperty(t, constant.APISIXVersion313, "batch-requests", "metadata", "max_pipeline_items"))
assert.NotNil(t, schemaProperty(t, constant.APISIXVersion317, "batch-requests", "metadata", "max_pipeline_items"))

// ai-rate-limiting accepts expression strategy in 3.17.
assert.NotContains(t, schemaEnum(t, constant.APISIXVersion313, "ai-rate-limiting", "", "limit_strategy"), "expression")
assert.Contains(t, schemaEnum(t, constant.APISIXVersion317, "ai-rate-limiting", "", "limit_strategy"), "expression")

// traffic-split has a new stream schema while preserving its HTTP schema.
assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "traffic-split", ""))
assert.NotNil(t, GetPluginSchema(constant.APISIXVersion317, "traffic-split", "stream"))
```

Implement these small test-only accessors; each one must call `GetPluginSchema`, use `require` for every map/slice conversion, and include the full JSON path in its failure message:

```go
func schemaProperty(t *testing.T, version constant.APISIXVersion, pluginName, schemaType, property string) any
func jwtAlgorithms(t *testing.T, version constant.APISIXVersion) []string
func schemaMinimum(t *testing.T, version constant.APISIXVersion, pluginName, schemaType, property string) float64
func schemaRequired(t *testing.T, version constant.APISIXVersion, pluginName, schemaType string) []string
func schemaEnum(t *testing.T, version constant.APISIXVersion, pluginName, schemaType, property string) []string
```

Pass `t` and the explicit schema type at every call site (for example, `"metadata"` for `batch-requests.max_pipeline_items`); do not use unchecked type assertions.

- [ ] **Step 3: Add positive/negative config validation pairs**

Compile the resolved schemas and prove:

- 3.17 `cas-auth` rejects a config without `cookie`, while 3.13 accepts the same legacy shape.
- 3.17 `hmac-auth` rejects `max_req_body_size: 0` and accepts `1`.
- 3.17 `jwt-auth` consumer accepts `algorithm: EdDSA` with a non-empty `public_key`; 3.13 rejects it.
- 3.17 `batch-requests` metadata accepts `max_pipeline_items: 1` and rejects `0`.

- [ ] **Step 4: Run all schema tests and repair only 3.17 assets**

Run:

```bash
go test ./pkg/utils/schema
```

Expected: PASS. Preserve 3.13 JSON and its expectations unchanged. When a 3.17 compile failure names `encrypt_fields`, implicit object typing, malformed regex, or an invalid example, correct only the corresponding 3.17 asset.

- [ ] **Step 5: Commit Schema integrity coverage**

```bash
git add src/apiserver/pkg/utils/schema/schema_317_test.go \
  src/apiserver/pkg/utils/schema/3.17/schema.json \
  src/apiserver/pkg/utils/schema/3.17/plugin.json \
  src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin.json \
  src/apiserver/pkg/utils/schema/3.17/bk_apisix_plugin_schema.json
git commit -m "test: lock APISIX 3.17 schema differences"
```

---

### Task 5: Extend Resource Field and Payload Capabilities to 3.17

**Files:**

- Modify: `src/apiserver/pkg/constant/resource_schema.go`
- Modify: `src/apiserver/pkg/constant/resource_schema_test.go`
- Modify: `src/apiserver/pkg/biz/resourcevalidation/prepare_payload_test.go`
- Modify: `src/apiserver/pkg/biz/resourcevalidation/validator_test.go`
- Modify: `src/apiserver/pkg/biz/publish/payload_test.go`

**Interfaces:**

- Consumes: `constant.APISIXVersion317` and the 3.17 core resource Schema.
- Produces: correct `name` support and `id` requirements for validation and publishing.

- [ ] **Step 1: Add failing capability tests**

Add 3.17 table cases proving:

```text
consumer_group.name  keep
stream_route.name    keep
proto.name           keep
global_rule.name     remove
ssl.name             remove
consumer_group.id    required
plugin_config.id     required
global_rule.id       required
consumer.id          remove
```

For `prepare_payload_test.go`, assert generated IDs are injected for consumer group/plugin config/global rule validation, outer names are injected for consumer group/stream route/proto, and consumer IDs/global-rule names are removed. For `payload_test.go`, assert the same cleanup on the final 3.17 publish payload.

- [ ] **Step 2: Run focused tests and verify 3.17 failures**

Run:

```bash
go test ./pkg/constant ./pkg/biz/resourcevalidation ./pkg/biz/publish \
  -run '317|ResourceSupportsNameFieldForVersion|ResourceRequiresIDInSchemaForVersion|Prepare|CleanupPublishPayloadFields'
```

Expected: FAIL because current equality/switch cases stop at 3.13.

- [ ] **Step 3: Extend only the explicit capability cases**

Change `ResourceSupportsNameFieldForVersion` to:

```go
return version == APISIXVersion313 || version == APISIXVersion317
```

for `ConsumerGroup`, `StreamRoute`, and `Proto`. Add `APISIXVersion317` to the `ResourceRequiresIDInSchemaForVersion` switch for `ConsumerGroup`, `PluginConfig`, and `GlobalRule`. Update comments/tables in the same file to name 3.17; do not change generic consumer ID behavior.

- [ ] **Step 4: Run capability and lifecycle tests**

Run:

```bash
go test ./pkg/constant ./pkg/biz/resourcevalidation ./pkg/biz/publish
```

Expected: PASS for 3.11, 3.13, and 3.17 cases.

- [ ] **Step 5: Commit resource capability support**

```bash
git add src/apiserver/pkg/constant/resource_schema.go \
  src/apiserver/pkg/constant/resource_schema_test.go \
  src/apiserver/pkg/biz/resourcevalidation/prepare_payload_test.go \
  src/apiserver/pkg/biz/resourcevalidation/validator_test.go \
  src/apiserver/pkg/biz/publish/payload_test.go
git commit -m "feat: support APISIX 3.17 resource payloads"
```

---

### Task 6: Enable APISIX 3.17 for MCP

**Files:**

- Modify: `src/apiserver/pkg/biz/mcp/access_token.go`
- Modify: `src/apiserver/pkg/biz/mcp/access_token_test.go`
- Modify: `src/apiserver/pkg/biz/mcp/resource_crud_test.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/common.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/common_test.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/schema.go`
- Modify: `src/apiserver/pkg/apis/mcp/tools/mcp_resource_validation_test.go`
- Modify: `src/apiserver/pkg/apis/mcp/prompts/workflows.go`
- Modify: `src/apiserver/pkg/apis/mcp/resources/docs.go`
- Modify: `src/apiserver/pkg/apis/mcp/AGENTS.md`

**Interfaces:**

- Consumes: 3.17 official/BK plugin maps and resource Schema.
- Produces: MCP token support and Schema tools accepting exactly `3.13.X` and `3.17.X`.

- [ ] **Step 1: Add failing MCP version and Schema tests**

Add table cases proving:

- `CheckGatewayMCPSupport` accepts gateways saved as `3.17.0` and `3.17.X`.
- `parseAPISIXVersion("3.17.X")` returns `APISIXVersion317`.
- `ValidAPISIXVersions` and `APISIXVersionDescription()` contain both versions.
- MCP resource and plugin validation receives `APISIXVersion317`.
- `GetPluginsList(..., APISIXVersion317, "bk-apisix")` contains official and BK plugins.
- 3.11 continues to return `ErrMCPGatewayNotSupported`.

- [ ] **Step 2: Run focused MCP tests and verify failure**

Run:

```bash
go test ./pkg/biz/mcp ./pkg/apis/mcp/tools -run 'MCP|APISIXVersion|PluginsList|Validation'
```

Expected: FAIL because both MCP whitelists only contain 3.13.

- [ ] **Step 3: Extend the two explicit MCP whitelists**

Set:

```go
var MCPSupportedAPISIXVersions = []constant.APISIXVersion{
    constant.APISIXVersion313,
    constant.APISIXVersion317,
}
```

and make `ValidAPISIXVersions` contain the same two constants. Extend `parseAPISIXVersion` with the `APISIXVersion317` case. Change the gateway error to:

```text
gateway does not support MCP (requires APISIX 3.13.X or 3.17.X)
```

- [ ] **Step 4: Update MCP tool descriptions, prompts, resources, and package rules**

Replace statements that say MCP supports only 3.13 with explicit `3.13.X or 3.17.X`. Examples that demonstrate a generic Schema call should use `3.17.X`; keep one 3.13 example where it demonstrates backward compatibility. Do not change tool names, input JSON field names, resource URIs, or publish permissions.

- [ ] **Step 5: Run all MCP tests**

Run:

```bash
go test ./pkg/biz/mcp ./pkg/apis/mcp/...
```

Expected: PASS, with 3.13 and 3.17 accepted and 3.11 rejected.

- [ ] **Step 6: Commit MCP 3.17 support**

```bash
git add src/apiserver/pkg/biz/mcp/access_token.go \
  src/apiserver/pkg/biz/mcp/access_token_test.go \
  src/apiserver/pkg/biz/mcp/resource_crud_test.go \
  src/apiserver/pkg/apis/mcp/tools/common.go \
  src/apiserver/pkg/apis/mcp/tools/common_test.go \
  src/apiserver/pkg/apis/mcp/tools/schema.go \
  src/apiserver/pkg/apis/mcp/tools/mcp_resource_validation_test.go \
  src/apiserver/pkg/apis/mcp/prompts/workflows.go \
  src/apiserver/pkg/apis/mcp/resources/docs.go \
  src/apiserver/pkg/apis/mcp/AGENTS.md
git commit -m "feat: enable MCP for APISIX 3.17"
```

---

### Task 7: Verify 3.17 Assets with the Frontend AJV Runtime

**Files:**

- Create: `src/frontend/scripts/validate-apiserver-schema.mjs`
- Modify: `src/frontend/package.json`

**Interfaces:**

- Consumes: all six 3.17 JSON assets.
- Produces: `npm run test:schema`, using the same `new Ajv()` plus `addFormats(ajv)` behavior as plugin editors.

- [ ] **Step 1: Add the frontend validation command**

Add to `package.json`:

```json
"test:schema": "node scripts/validate-apiserver-schema.mjs"
```

The script must:

1. load `../apiserver/pkg/utils/schema/3.17/schema.json` and the BK/TAPISIX Schema files;
2. load all three plugin catalogs;
3. compile every `main.*`, `plugins.*.(schema|consumer_schema|metadata_schema)`, and `stream_plugins.*.schema` node with `new Ajv()` and `addFormats(ajv)`;
4. unwrap a `plugin_schema` object exactly as Go's `normalizePluginSchema` does;
5. resolve examples in official -> BK -> TAPISIX order;
6. validate `example`, `consumer_example`, and `metadata_example` in their correct scopes;
7. collect all failures and exit non-zero after printing `source/plugin/scope` and AJV errors.

Use this exact lookup shape:

```js
const schemaSources = [officialSchema, bkSchema, tapisixSchema];
const lookup = (path) => {
  for (const source of schemaSources) {
    const value = path.split('.').reduce((node, key) => node?.[key], source);
    if (value !== undefined) return value.plugin_schema ?? value;
  }
  return undefined;
};
```

Do not set `strict: false`; the command must reproduce the production editor's default AJV constructor.

- [ ] **Step 2: Run the AJV command and verify real compatibility**

Run from `src/frontend`:

```bash
npm run test:schema
```

Expected: PASS with a summary containing the number of compiled Schema nodes and validated examples. If it fails on project-normalization issues, repair the 3.17 JSON in Tasks 2-4. Do not alter the three Vue editor components unless a standards-compliant 3.17 Schema still fails with their exact current AJV setup.

- [ ] **Step 3: Run frontend lint only for the new script/package change**

Run:

```bash
npx eslint scripts/validate-apiserver-schema.mjs --fix --color
npm run test:schema
```

Expected: both commands exit 0.

- [ ] **Step 4: Commit the AJV compatibility gate**

```bash
git add src/frontend/package.json src/frontend/scripts/validate-apiserver-schema.mjs \
  src/apiserver/pkg/utils/schema/3.17
git commit -m "test: verify APISIX schemas with frontend AJV"
```

---

### Task 8: Add Compact 3.17 Integration Coverage and Update Support Docs

**Files:**

- Create: the seven Bruno files listed in the File Map.
- Modify: `src/apiserver/AGENTS.md`
- Modify: `src/apiserver/pkg/utils/schema/AGENTS.md`

**Interfaces:**

- Consumes: Open API version registration, 3.17 Schema maps, resource capabilities, and plugin examples.
- Produces: one 3.17 gateway plus representative consumer, HTTP, metadata, stream, and invalid Schema scenarios.

- [ ] **Step 1: Add the 3.17 gateway fixture**

Create `03_create_gateway_3_17_0.bru` by following the existing 3.13 request shape with these exact differences:

```json
{
  "name": "bk-apisix-plugin-matrix-317",
  "description": "Plugin matrix gateway 3.17.0",
  "apisix_type": "bk-apisix",
  "apisix_version": "3.17.0",
  "etcd_prefix": "/plugin-matrix-317"
}
```

Keep mode, endpoint, maintainer, credentials, headers, and success assertions identical to the existing 3.13 gateway fixture.

- [ ] **Step 2: Add four representative success batches**

Use minimal examples already validated by Task 4:

- consumer: `jwt-auth` with an `EdDSA` public-key config and `hmac-auth` with existing consumer credentials;
- HTTP route: `proxy-buffering`, `oas-validator`, and one changed AI plugin (`ai-rate-limiting` with integer limits, not an untrusted Lua expression);
- metadata: `batch-requests.max_pipeline_items: 100` and `kafka-logger.max_pending_entries: 1000`;
- stream route: `traffic-split` using the exact minimal stream example already validated from the 3.17 catalog in Task 4; do not invent or copy HTTP-only fields.

Each request targets `bk-apisix-plugin-matrix-317`, expects HTTP 200, and asserts `res.body.data.length` equals the number of submitted resources.

- [ ] **Step 3: Add two representative invalid cases**

Create:

- a route using `cas-auth` with `idp_uri`, `cas_callback_uri`, and `logout_uri` but no required `cookie`; expect HTTP 400 and `BadRequest`;
- a route using `hmac-auth.max_req_body_size: 0`; expect HTTP 400 and `BadRequest`.

These cases must fail specifically because of 3.17 constraints, not because the plugin value is merely the wrong root type.

- [ ] **Step 4: Update repository support matrices**

In `src/apiserver/AGENTS.md`:

- add 3.17.X to the supported-version paragraph and field matrix;
- state integration coverage now includes compact 3.17 cases;
- add `3.17` to the Schema directory list;
- state MCP supports `3.13.X` and `3.17.X`.

In `pkg/utils/schema/AGENTS.md`:

- record the two fixed source commits;
- record that 3.17 excludes `mcp-bridge`, `server-info`, `node-status`, and `log-rotate` from `plugin.json`;
- add `npm run test:schema` from `src/frontend` to the validation checklist.

- [ ] **Step 5: Run the compact integration suite**

Run from `src/apiserver`:

```bash
make integration-test
```

Expected: Docker Compose exits 0; existing 3.11/3.13 matrix scenarios and the seven 3.17 scenarios pass. If Docker is unavailable, report this gate as blocked with the exact command/output; do not claim integration success.

- [ ] **Step 6: Commit integration and documentation coverage**

```bash
git add src/apiserver/tests/integration/openapi/plugin_matrix/00_gateways/03_create_gateway_3_17_0.bru \
  src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/07_success_3_17_0_consumer.bru \
  src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/08_success_3_17_0_http.bru \
  src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/09_success_3_17_0_metadata.bru \
  src/apiserver/tests/integration/openapi/plugin_matrix/02_success_batches/10_success_3_17_0_stream.bru \
  src/apiserver/tests/integration/openapi/plugin_matrix/03_invalid_cases/151_invalid_3_17_0_cas-auth_missing-cookie.bru \
  src/apiserver/tests/integration/openapi/plugin_matrix/03_invalid_cases/152_invalid_3_17_0_hmac-auth_max-body-size.bru \
  src/apiserver/AGENTS.md src/apiserver/pkg/utils/schema/AGENTS.md
git commit -m "test: cover APISIX 3.17 integration paths"
```

---

## Completion Gate

- [ ] **Step 1: Verify only authorized files changed**

Run:

```bash
git status --short
git diff --check upstream/master...HEAD
git diff --name-status upstream/master...HEAD
```

Expected: no whitespace errors; no pre-existing untracked user file appears in a commit.

- [ ] **Step 2: Run focused backend verification**

From `src/apiserver`:

```bash
go test ./pkg/utils/version ./pkg/utils/schema ./pkg/constant \
  ./pkg/biz/resourcevalidation ./pkg/biz/publish ./pkg/biz/mcp ./pkg/apis/mcp/...
```

Expected: PASS.

- [ ] **Step 3: Run repository-required backend gates**

From `src/apiserver`:

```bash
make lint
make test
```

Expected: both commands exit 0. Because `make lint` applies fixes and `make test` runs `go mod tidy`, inspect `git status` afterward and retain only changes directly caused by this feature.

- [ ] **Step 4: Run frontend verification**

From `src/frontend`:

```bash
npm run test:schema
npx eslint scripts/validate-apiserver-schema.mjs --color
npm run build:ee
```

Expected: all exit 0. Check whether the build rewrote generated/index files and exclude unrelated generated changes.

- [ ] **Step 5: Run integration verification**

From `src/apiserver`:

```bash
make integration-test
```

Expected: exit 0. Record Docker/image/environment blockers verbatim when this gate cannot run.

- [ ] **Step 6: Perform independent final review**

Use `superpowers:requesting-code-review` because this is a cross-module, behavior-facing change. Review against the approved design, with special attention to:

- accidental 3.13 asset mutation;
- incomplete same-name Schema refresh;
- catalog entries without matching scope Schema;
- missing BK/TAPISIX version map entries;
- 3.17 version leakage into future versions;
- MCP docs/whitelists disagreeing;
- frontend AJV and Go validator divergence;
- integration fixtures that fail for unrelated fields.

- [ ] **Step 7: Record final evidence**

The completion report must list each command, exit status, any skipped/blocked gate, final branch/commit range, and the exact file set. Do not claim “tests pass” if integration or frontend validation was skipped.
