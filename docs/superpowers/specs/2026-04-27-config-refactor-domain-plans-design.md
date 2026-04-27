# APIServer Config 重构分领域计划设计

## 背景

本设计基于 [to_refactor.md](/root/workspace/tx/wklken/blueking-micro-apigateway/to_refactor.md)。
目标不是立刻开始跨领域统一抽象，
而是先把当前重构思路拆成 4 份相互独立的 plan 文档，分别对应：

- `web api`
- `open api`
- `import`
- `mcp`

每一份 plan 都必须能够拆成一串小步、PR 级、可单独验收的任务。
每一个任务都必须能独立 review、独立合并、独立验证。

## 已确认约束

### 1. 全局约束

- 以最小化修改为原则。
- 目标是降低复杂度、提升维护一致性。
- 保持现有 API 协议、数据库形态、存量数据兼容性。
- `pkg/entity/model/*.go` 中 `HandleConfig()` 的行为是稳定边界，不作为这些 plan 的重构目标。
- `import overlay` 和 `import.ignore_fields` 继续保留为 import 本地能力。
- `mcp` 当前不是主线重构目标，除非未来某个调整对一致性有非常高的杠杆作用。

### 2. 计划约束

- 最终会有 4 份相互独立的 plan 文件。
- 4 份 plan 之间不设依赖关系。
- 每一份 plan 内部有严格顺序，任务之间有依赖。
- 实际执行方式是：在单个 plan 内部一步步执行。
- 每一个任务都预期成为一个单独的 PR。
- 跨领域/共享抽象工作，明确不纳入这 4 份 plan。
- 只有 4 份 plan 全部做完之后，才重新讨论共享抽象。

### 3. 交付约束

- 每一个任务都必须遵循同一个节奏：
  - 先新增或补齐现有行为的测试
  - 再使用 TDD 完成重构
- 一个任务不能捆绑多个无关复杂度问题。
- 每一个任务都必须解决一个明确的复杂度或可维护性问题。
- spec 和 plan 都使用中文编写。

## 为什么要这样拆

当前代码里同时存在两类问题：

- 领域内部的本地复杂度
- 不同领域之间的重复和不一致

如果过早开始抽共享 helper，那么共享层很可能建立在仍然混乱的本地实现之上。
这样会把局部问题提前固化进公共抽象，反而增加后续清理成本。

更稳妥的策略是：

1. 先清理每个领域内部的复杂度
2. 用测试把本地行为显式化
3. 改善命名、helper 边界、职责分配和调用顺序
4. 最后再比较这些已经稳定的领域实现，判断哪些真的值得抽公共

因此，本设计明确停在“分领域计划设计”这一层，
不提前进入跨领域抽象阶段。

## 每份 Plan 的结构要求

每一份 plan 都必须包含下面这些内容：

1. 说明该领域的范围和非目标
2. 列出该领域预计会修改的文件
3. 拆成多个按顺序执行的小任务
4. 对每个任务都明确写清：
   - 该任务解决的具体复杂度问题是什么
   - 为什么这个任务适合独立提 PR
   - 会改哪些文件
   - 先补什么现状测试
   - 后续用什么 TDD 重构动作收敛复杂度
   - 用什么最小验收命令证明该任务完成

## 四份 Plan 的设计

### Plan A：`web api`

#### 目标

降低 web 侧在请求校验顺序、校验期 `config` 改写、以及 handler 侧重复组装逻辑上的复杂度。

#### 非目标

- 不做跨领域共享 helper 抽取
- 不修改 `HandleConfig()` 行为
- 不处理 publish 侧问题

#### 任务顺序

1. 固化 `CheckAPISIXConfig()` 当前行为矩阵。
2. 从 `CheckAPISIXConfig()` 中拆出 web 本地 validation payload helper，但不改行为。
3. 固化 `plugin_config`、`consumer_group`、`global_rule` 三个 create handler 的当前流程。
4. 为这三个 handler 提取 web 本地 create-flow helper。
5. 固化其余仍然“先校验、后生成最终 ID”的 web create handler 当前行为。
6. 为 web handler 中重复的 `ResourceCommonModel` 组装提取 web 本地 helper。

#### 原因

`web api` 当前同时混着处理请求校验整形和 handler 组装流程，
而且不同资源之间顺序不一致，是 4 个领域里局部顺序复杂度最重的一块。

### Plan B：`open api`

#### 目标

降低 open 侧在 middleware 校验整形、serializer 组装、以及 identity 重复解析上的复杂度。

#### 非目标

- 暂不抽跨领域共享 builder
- 不改协议
- 不碰 import overlay 行为

#### 任务顺序

1. 固化 middleware 当前校验前 payload 行为，覆盖资源类型和版本差异。
2. 提取 open 本地 middleware validation helper，并理清阶段职责。
3. 固化 serializer 的 create / batch create / update 当前组装行为。
4. 提取 open 域内的 serializer builder，对齐这三条组装路径。
5. 消除 open 域内的双重 identity 生成，让校验和持久化复用同一份 resolved identity。

#### 原因

`open api` 的核心问题非常集中：
同一个领域里，middleware 和 serializer 两层都在改写 `config` 和处理 identity，
但又没有共享同一份结果。

### Plan C：`import`

#### 目标

降低 import 侧复杂度，把 import 本地 overlay 逻辑、通用校验准备、以及 sync-data 组装边界拆清楚。

#### 非目标

- 不把 `ignore_fields` 移进共享代码
- 暂不抽跨领域 helper
- 不改 `HandleConfig()` 行为

#### 任务顺序

1. 固化当前 `ignore_fields` overlay 行为。
2. 提取 import 本地 overlay helper。
3. 固化 `handleResources(...)` 当前分类和 `GatewaySyncData` 组装行为。
4. 重命名并拆分 `handleResources(...)`，拆成更小的 import 本地函数。
5. 增加 import 本地 validation payload helper，让 import 也有显式的校验前 seam。

#### 原因

`import` 有一条非常明确的本地规则必须保留，
即 overlay / `ignore_fields`。
在未来和 web/open 做比较之前，必须先把这个边界从大函数里明确拆出来。

### Plan D：`mcp`

#### 目标

当前不对 MCP 做深度重构，只固化现状行为，并让它的局部复杂度变得可测、可评估。

#### 非目标

- 默认不接入共享抽象
- 不改 MCP 协议
- 不做大范围结构调整

#### 任务顺序

1. 固化当前 create/update 在 `name` 注入、ID 生成、`ResourceCommonModel` 组装上的行为。
2. 只提取极小的 MCP 本地 helper，处理重复的命名/config 注入逻辑，不改变外部行为。
3. 在测试或文档中增加清晰的评估边界，说明未来什么情况下 MCP 才值得接入共享抽象。

#### 原因

当前共识已经明确：
MCP 不是本轮主线重构目标。
所以 MCP plan 必须保持小、稳、防御性强。

## 后续应产出的 Plan 文件

在这份 spec 被确认之后，下一步应生成 4 份中文 plan 文档：

- `docs/superpowers/plans/2026-04-27-web-api-config-refactor-plan.md`
- `docs/superpowers/plans/2026-04-27-open-api-config-refactor-plan.md`
- `docs/superpowers/plans/2026-04-27-import-config-refactor-plan.md`
- `docs/superpowers/plans/2026-04-27-mcp-config-refactor-plan.md`

## 对这 4 份 Plan 的验收标准

只有同时满足下面条件，plan 才算合格：

- 每一份 plan 都是领域内的、本地的、独立的
- 每一个任务都是 PR 级
- 每一个任务只解决一个明确复杂度问题
- 每一个任务都先补现状测试
- 每一个任务都再用 TDD 做重构
- 每一份 plan 都明确写出内部任务依赖顺序
- 没有任何 plan 提前引入跨领域共享抽象

## 明确不在本设计范围内的内容

- 共享的 `DATABASE` validation payload builder
- 跨领域共享的 resolved identity helper
- 跨领域共享的 pre-`HandleConfig()` builder
- 共享的 publish payload builder
- read-time restore 相关工作

这些内容都明确延后到 4 份领域计划全部完成之后再讨论。
