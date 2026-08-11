# APISIX 3.17 数据面接入与纳管设计

## 1. 背景

BlueKing Micro API Gateway 当前显式支持 APISIX 3.2、3.3、3.11 和 3.13。APISIX 版本会决定：

- 网关创建时可选择的版本；
- 资源 JSON Schema；
- 官方、TAPISIX 和 BK 插件目录；
- 插件主配置、consumer 配置、metadata 配置和 stream 配置的 Schema；
- Web、Open API、导入、MCP 和发布前的配置校验；
- 发布 payload 中 `id`、`name` 等版本相关字段的保留或清理；
- MCP 功能是否可用。

本次为 APISIX 3.17 新增一套独立的版本快照，使新建的 `apisix` 和 `bk-apisix` 3.17 数据面能够接入并由控制面管理。APISIX 3.17 的来源固定为 Apache APISIX `3.17.0`，不直接追随仍会变化的 `release/3.17` 分支。当前已核对的 3.17.0 源码提交为 `9ef2ecab67f652d38365049613610ef649bb4ad0`；BK 数据面基线为 `blueking-apigateway-apisix` 的 3.17 升级提交 `c1c44e5dc192120ccfa432e9f54703285a80ca38`。

参考：

- <https://github.com/apache/apisix/tree/release/3.17>
- <https://raw.githubusercontent.com/apache/apisix/release/3.17/CHANGELOG.md>

## 2. 目标

1. 网关创建页面和 Open API 接受 APISIX `3.17.x` 版本族，并展示已验证的 `3.17.0`。
2. 新建 `apisix 3.17.0` 和 `bk-apisix 3.17.0` 网关后，资源与插件配置使用独立的 3.17 Schema。
3. 完整刷新同名插件在 3.17 中发生变化的 Schema，不把 3.17 实现成“3.13 Schema 加新增插件”。
4. 保持 3.13 的插件组合、接入探测和发布数据流不变。
5. 允许 3.17 网关使用 MCP 管理能力。
6. 保持所有既有 3.13 行为和测试通过。

## 3. 非目标

- 不支持已纳管的 3.13 网关原地切换到 3.17。
- 不新增数据库迁移。
- 不为 3.13 配置提供到 3.17 的自动转换器。
- 不改变接入时的 APISIX 版本探测机制。
- 不比较用户申报版本和 `server_info` 探测版本。
- 不检查同一 etcd 前缀下所有 APISIX 节点的版本一致性。
- 不新增 `tapisix 3.17` 的前端版本入口。
- 不增加 APISIX `credential` 等新的控制面资源类型，继续管理现有 11 类资源。
- 不在运行时从数据面动态下载 Schema。

## 4. 关键设计决策

### 4.1 独立版本快照

3.17 使用独立目录，不复用或修改 3.13 文件：

```text
src/apiserver/pkg/utils/schema/3.17/
├── schema.json
├── plugin.json
├── bk_apisix_plugin.json
├── bk_apisix_plugin_schema.json
├── tapisix_plugin.json
└── tapisix_plugin_schema.json
```

文件职责：

- `schema.json`：APISIX 核心资源、官方 HTTP 插件和 stream 插件的完整 3.17 Schema。
- `plugin.json`：控制面开放的 3.17 官方插件目录及最小合法示例。
- `bk_apisix_plugin.json`：沿用 3.13 的七插件控制面目录；其中四个由目标 BK 3.17 数据面实证，
  三个为兼容性保留。
- `bk_apisix_plugin_schema.json`：四个运行时导出 Schema 加三个从 3.13 保留的兼容 Schema。
- `tapisix_plugin.json`：与 3.13 一致的空数组 `[]`。
- `tapisix_plugin_schema.json`：与 3.13 一致的空对象 `{"plugins": {}}`。

空 TAPISIX 文件是兼容性占位，不为 3.17 增加实际 TAPISIX 插件能力。

### 4.2 版本表示

- 对外展示和保存经过验证的完整版本 `3.17.0`。
- 内部使用 `3.17.X` 选择版本快照，使同一 minor 下的补丁版本共用 Schema。
- 只自动接受显式注册的 minor 版本族，不因为版本号大于 3.17 就继承 3.17 能力。

### 4.3 插件组合保持 3.13 行为

插件目录的组合逻辑不改变：

- `apisix`：官方 APISIX 插件；
- `tapisix`：官方 APISIX 插件加 TAPISIX 插件；
- `bk-apisix`：官方 APISIX 插件加 TAPISIX 插件再加 BK 插件。

3.17 的 TAPISIX 文件为空，因此 `bk-apisix 3.17` 的实际插件集合是官方插件加 BK 插件。

插件 Schema 的既有查找顺序也不改变：官方 APISIX Schema、BK Schema、TAPISIX Schema。插件名不得在不同来源之间产生未评估的冲突。

### 4.4 接入探测保持 3.13 行为

控制面继续查询 `{etcd_prefix}/data_plane/server_info`，从第一条记录读取 `id` 和 `version`：

- Web 页面测试连接后，用探测结果辅助回填版本；
- 没有 `server_info` 时仍允许人工声明版本；
- 创建网关时保存用户提交的版本；
- 不校验探测版本与申报版本是否一致；
- `instance_id` 存在时继续用于避免同一实例被重复纳管。

这是有意保留的兼容行为，不在本次顺带强化。

## 5. Schema 获取与归一化

### 5.1 官方资源和插件 Schema

从固定的 APISIX 3.17.0 源码或对应运行时 Schema 输出获取完整快照。不能以 3.13 文件为基础只追加新增插件。

必须完整覆盖：

- `main.<resource>`；
- `plugins.<name>.schema`；
- `plugins.<name>.consumer_schema`；
- `plugins.<name>.metadata_schema`；
- `stream_plugins.<name>.schema`。

归一化必须遵守 `src/apiserver/pkg/utils/schema/AGENTS.md`：

- 有 `properties` 的对象显式声明 `type: object`；
- 保留控制面校验依赖的核心资源字段；
- 保持上述 gjson 路径稳定；
- 删除项目校验器不接受的非标准 `encrypt_fields` 元数据，但保留相关真实配置字段及约束；
- 修复不合法的 Schema 片段，不能通过放宽测试绕过；
- 插件示例必须最小且能通过同版本、同作用域 Schema。

### 5.2 BK 插件 Schema

BK 数据面固定到 `c1c44e5dc192120ccfa432e9f54703285a80ca38`。该运行时只导出
`bk-break-recursive-call`、`bk-delete-cookie`、`bk-jwt` 和 `bk-traffic-label` 四个既有 BK 插件。
按已确认的“插件组合保持 3.13 逻辑”，控制面继续暴露
`bk-echo`、`bk-header-rewrite` 和 `bk-login-required`；这三个插件的 Schema 明确使用 3.13 兼容快照，
不得标记为目标 3.17 BK 运行时导出结果。

对四个运行时实证插件执行 3.13/3.17 全量结构比较；对三个兼容保留插件验证目录、Schema 和示例完整性。

### 5.3 全量 Schema 差异审计

对 3.13 和 3.17 进行结构化 JSON diff，至少比较：

- 插件新增和删除；
- 同名插件字段新增、删除及重命名；
- `required` 变化；
- 字段类型、枚举、默认值、范围、正则变化；
- `oneOf`、`anyOf`、`allOf`、条件 Schema 和依赖关系变化；
- consumer、metadata、stream 作用域的新增或删除。

3.17 的同名插件必须保存完整的新 Schema。发生变化的插件示例必须按 3.17 重新生成，不得直接复制 3.13 示例。

重点插件包括但不限于：

- `jwt-auth`；
- `hmac-auth`；
- `cas-auth`；
- `batch-requests`；
- `proxy-cache`；
- `tencent-cloud-cls`；
- `ai-rag`；
- `ai-proxy`；
- `ai-proxy-multi`；
- `ai-rate-limiting`；
- logger、限流和 stream 插件。

最终范围以全量 Schema diff 为准，不能只依据 CHANGELOG 中列出的 breaking changes。

## 6. 插件目录策略

3.17 相对 3.13 新增的默认插件包括：

- `proxy-buffering`；
- `saml-auth`；
- `dingtalk-auth`；
- `feishu-auth`；
- `acl`；
- `data-mask`；
- `ai-aliyun-content-moderation`；
- `graphql-proxy-cache`；
- `graphql-limit-count`；
- `traffic-label`；
- `oas-validator`。

这份列表是候选目录，不等于全部自动开放。一个插件进入 `plugin.json` 必须同时满足：

1. 目标数据面实际提供并启用；
2. 有可用的 3.17 Schema；
3. 适合通过当前控制面配置；
4. 至少有一个合法的最小示例；
5. 不属于本项目明确排除的内部、特殊或废弃入口。

继续排除 `mcp-bridge`、`server-info` 等当前不通过插件管理页面开放的特殊插件。3.13 已开放、3.17 仍有有效 Schema 的兼容插件继续保留。

APISIX 3.17 为 stream 增加 `traffic-split`，因此 `StreamRoutePluginMap` 需要加入该插件，同时保留现有 stream 映射。

## 7. 代码改动面

### 7.1 版本注册

`src/apiserver/pkg/constant/apisix.go`：

- 新增 `APISIXVersion317 = "3.17.X"`；
- 加入 `SupportAPISIXVersionMap`。

`src/apiserver/pkg/utils/schema/version.json`：

- `apisix.support_version` 增加 `3.17.0`；
- `bk-apisix.support_version` 增加 `3.17.0`；
- 不新增 `tapisix` 键。

### 7.2 插件 embed 和映射

`src/apiserver/pkg/utils/schema/plugin.go`：

- embed 3.17 的官方、BK、TAPISIX 插件目录；
- 更新 `versionPluginMap`；
- 更新 `versionBkAPISIXPluginMap`；
- 更新 `versionTAPISIXPluginMap`；
- 增加 3.17 官方插件文档 URL；
- `StreamRoutePluginMap` 加入 `traffic-split`。

### 7.3 Schema embed 和映射

`src/apiserver/pkg/utils/schema/schema.go`：

- embed 3.17 官方资源和插件 Schema；
- embed 3.17 BK 插件 Schema；
- embed 3.17 空 TAPISIX Schema；
- 更新 `schemaVersionMap`；
- 更新 `bkAPISIXPluginSchemaVersionMap`；
- 更新 `tapisixPluginSchemaVersionMap`；
- 不改变现有查找和 fallback 逻辑。

### 7.4 资源字段能力

`src/apiserver/pkg/constant/resource_schema.go` 当前只显式识别 3.11 和 3.13。需要将 3.17 加入实际能力矩阵：

- `consumer_group`、`stream_route`、`proto` 在 3.17 支持 `name`；
- `consumer_group`、`plugin_config`、`global_rule` 在 3.17 Schema 校验时需要 `id`；
- consumer 的 `id` 仍按既有控制面规则处理。

使用显式版本枚举，不引入“所有未来版本自动继承 3.17”的阈值判断。

### 7.5 MCP

`src/apiserver/pkg/biz/mcp/access_token.go`：

- `MCPSupportedAPISIXVersions` 加入 `3.17.X`；
- 更新仅提到 3.13 的错误信息。

同时更新 MCP Schema 工具、prompts、resources、工具说明和项目规则文档中“仅支持 3.13”的陈述，包括：

- `src/apiserver/pkg/apis/mcp/tools/schema.go`；
- `src/apiserver/pkg/apis/mcp/tools/common.go`；
- `src/apiserver/pkg/apis/mcp/prompts/workflows.go`；
- `src/apiserver/pkg/apis/mcp/resources/docs.go`；
- `src/apiserver/AGENTS.md`；
- `src/apiserver/pkg/apis/mcp/AGENTS.md`。

前端 MCP 菜单已经按 `>= 3.13` 展示，不改变判断逻辑。

### 7.6 前端

网关版本列表由后端 `version.json` 驱动，因此不需要为 3.17 新增硬编码选项。

插件配置使用 Monaco 编辑 JSON，并由 AJV 8.17.1 动态编译后端返回的 Schema。标准 JSON Schema 结构原则上无需逐插件修改页面。若 3.17 Schema 无法被 AJV 编译，优先修正快照的项目归一化；只有标准 Schema 本身与当前 AJV 配置不兼容时，才允许做最小 AJV 配置调整。

### 7.7 不需要修改的主流程

Web API、Open API、导入、同步、差异和发布主流程继续通过网关保存的版本选择 Schema，不新增 3.17 专用分支。数据库模型和表结构不变。

## 8. 错误处理

- 不支持的版本继续由 `apisixVersion` 校验器拒绝。
- 3.17 embed 或 map 遗漏必须由测试在交付前发现，不能静默回退到 3.13。
- 插件存在于目录但找不到相同版本、相同作用域的 Schema，视为资产错误；修正 JSON 资产，不放宽验证器。
- 3.17 配置不满足新版 Schema 时，沿用 Web、Open API、MCP 和发布流程现有错误返回。
- 接入探测为空或探测版本与申报版本不一致时，保持 3.13 的现有处理。

## 9. 测试设计

### 9.1 Schema 和插件资产

扩展 `pkg/utils/schema` 测试：

- 3.17 的 11 类资源 Schema 均存在且能编译；
- 官方、BK、TAPISIX 三组 map 均注册 3.17；
- TAPISIX 两个占位文件保持空；
- 每个插件目录项在对应 main、consumer、metadata 或 stream 作用域有 Schema；
- 所有 3.17 示例通过对应 Schema；
- `traffic-split` 的 stream Schema 可查询；
- 3.17 资产不引用 3.13 map。

### 9.2 版本隔离

为发生破坏性变化的代表插件增加 3.13/3.17 对照测试，覆盖 `jwt-auth`、`hmac-auth`、`cas-auth`、`batch-requests`、`proxy-cache`、AI 插件、一个 logger 和 `traffic-split` stream Schema。

测试必须证明：

- 两个版本查询到的 Schema 存在预期差异；
- 各自示例按对应版本通过；
- 至少有一组配置能体现 3.13 与 3.17 的不同校验结论；
- 3.17 资产变更不改变 3.13 行为。

### 9.3 字段能力和 payload

扩展以下测试：

- `pkg/constant/resource_schema_test.go`；
- `pkg/biz/resourcevalidation/prepare_payload_test.go`；
- `pkg/biz/resourcevalidation/validator_test.go`；
- `pkg/biz/publish/payload_test.go`。

覆盖 3.17 的 `name`、`id` 注入、清理、数据库校验和发布 payload。

### 9.4 MCP

扩展 MCP token 和 Schema 工具测试：

- 3.17 网关可创建并使用 MCP token；
- 3.17 能查询资源和插件 Schema；
- 3.17 配置校验使用 3.17 Schema；
- 3.11 等未支持 MCP 的版本继续被拒绝。

### 9.5 前端兼容

验证 AJV 8.17.1 能编译全部 3.17 插件 Schema，且所有作用域示例都能通过。现有前端没有单元测试入口，因此该检查作为可重复的 Node/AJV 验证命令和构建验证执行；若实施中修改前端生产代码，再运行前端 lint 和相应构建。

### 9.6 集成测试

新增紧凑的 3.17 集成场景：

- 创建一个 `bk-apisix 3.17.0` 测试网关；
- HTTP、consumer、metadata、stream 各一个批量成功场景；
- 为新增插件和破坏性变更插件增加代表性失败场景。

不复制约 100 个“每插件一个 `.bru`”的 3.17 文件。全量插件合法性由数据驱动的 Go 测试覆盖。若后续要求完整复制当前 Bruno 矩阵，应拆成独立测试工程任务。

## 10. 验证门禁

从 `src/apiserver` 执行：

```bash
go test ./pkg/utils/schema
go test ./pkg/constant ./pkg/biz/resourcevalidation ./pkg/biz/publish ./pkg/biz/mcp
make lint
make test
```

条件允许时执行：

```bash
make integration-test
```

另外从 `src/frontend` 执行 AJV 全量 Schema 编译/示例验证。若前端代码发生修改，再执行前端 lint 和适用版本构建。

## 11. 验收标准

1. 创建页面能为 `apisix` 和 `bk-apisix` 选择 `3.17.0`。
2. 可新建并查询 `apisix 3.17.0` 和 `bk-apisix 3.17.0` 网关。
3. 3.17 网关的资源 Schema、插件目录和插件 Schema 全部来自 3.17 映射。
4. 新增插件和同名变更插件能按 3.17 Schema 正确接受或拒绝配置。
5. Web、Open API、导入、同步、差异和发布路径能处理 3.17 网关。
6. 3.17 网关可以使用 MCP token 和 MCP Schema/资源工具。
7. 3.17 stream 插件列表包含 `traffic-split`，并使用其 stream Schema。
8. TAPISIX 3.17 占位资产保持空，插件组合行为与 3.13 一致。
9. 3.13 的资源、插件、发布和 MCP 行为无回归。
10. 没有新增存量网关版本切换、严格版本探测或数据库迁移能力。

## 12. 实施边界与风险

- 最大风险是 Schema 快照不完整或归一化时丢失 3.17 约束。通过固定源码、全量结构 diff、目录与 Schema 双向覆盖测试缓解。
- 同名插件的 3.13 示例可能不再适用于 3.17。所有变更插件必须重新生成和验证示例。
- 前后端分别使用 Go JSON Schema 校验器和 AJV，可能出现方言或关键字解释差异。通过对同一批示例执行双端校验缓解。
- `server_info` 探测仍不是严格版本证明。这是明确接受的既有兼容风险，不在本次扩大范围。
- 完整复制 3.17 Bruno 插件矩阵会使改动超过 100 个文件。本设计采用数据驱动单元测试和代表性集成测试控制范围。
