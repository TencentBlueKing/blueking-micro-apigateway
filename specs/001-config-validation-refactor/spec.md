# Feature Specification: APIServer Config Validation Refactor

**Feature Branch**: `[001-config-validation-refactor]`  
**Created**: 2026-04-25  
**Status**: Draft  
**Input**: User description: "阅读 current_validation.md and solution.md and AGENTS.md；实现 apiserver config validation new solution，并在实现前补全当前实现涉及模块、函数、入口的相关单元测试，确保重构后相同 case 通过；这是线上项目且已有存量数据，需要兼容 web 表单、openapi 和最终发布的数据正确、合法、有效。"

## Clarifications

### Session 2026-04-25

- Q: OpenAPI serializer 协议在本次重构中是否允许变更？ → A: 不允许，现有 OpenAPI serializer 协议必须保持兼容，因为已经存在线上调用依赖。
- Q: WebAPI 表单入参协议在本次重构中是否允许变更？ → A: 原则上不变；若确认必须调整，必须同步变更前端 Vue 表单并明确协商。
- Q: 数据库字段在本次重构中是否允许变更？ → A: 不允许作为默认方案；若确实需要，必须在实施前显式说明。
- Q: 数据入库前是否必须经过对应版本 JSON Schema 校验？ → A: 必须，入库前的数据必须经过对应版本 JSON Schema 校验。
- Q: 发布到 APISIX 的配置是否必须经过对应版本 JSON Schema 校验？ → A: 必须，发布到 APISIX 的配置必须经过对应版本 JSON Schema 校验。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Freeze Current Behavior With A Regression Baseline (Priority: P1)

作为维护线上网关控制面的后端负责人，我需要先把当前配置校验与发布成形链路的真实行为固化为自动化测试基线，再开始重构，这样任何行为变化都能被识别，而不是在发布后才暴露为兼容性事故。

**Why this priority**: 这是生产系统重构的前置条件。没有完整基线，就无法证明重构只是在收敛实现复杂度，而不是无意改变线上行为。

**Independent Test**: 通过自动化单元测试运行完整回归用例集，验证每个在范围内的入口、模块、资源特例、版本差异都已有当前行为断言，并能区分接受、拒绝、持久化结果和发布结果。

**Acceptance Scenarios**:

1. **Given** 某个当前仍在生产路径中参与配置校验、配置投影、持久化回写或发布 payload 构造的入口或模块，**When** 功能进入重构阶段，**Then** 该入口或模块必须先拥有自动化单元测试，覆盖其当前可观察行为。
2. **Given** 某个当前会被接受或拒绝的资源配置案例，**When** 在重构前后分别执行同一回归用例，**Then** 其接受或拒绝结果、关键字段成形结果以及最终发布结果必须保持一致，除非差异已事先批准并记录。
3. **Given** 某类资源存在已知特例或版本差异，**When** 建立测试基线，**Then** 该特例或差异必须以显式断言存在于回归用例中，而不是依赖人工记忆。

---

### User Story 2 - Validate Web And OpenAPI Inputs With One Consistent Semantics (Priority: P1)

作为管理台使用者或 OpenAPI 集成方，我需要 web 表单与 OpenAPI 在面对同一资源、同一操作、同一目标版本时遵守同一套语义规则，这样合法数据会稳定通过，不合法或冲突数据会在入库前被一致地拦截。

**Why this priority**: 用户最直接感知的是“为什么同样的配置从一个入口能过、另一个入口不能过”。统一入口语义是降低故障率和运维成本的核心收益。

**Independent Test**: 使用等价的 web 表单输入和 OpenAPI 输入分别执行创建、更新和校验相关测试，确认它们对身份字段、名称字段、关联字段、版本差异与冲突处理的结果一致。

**Acceptance Scenarios**:

1. **Given** 一份在结构化字段与配置内容中表达一致语义的合法资源定义，**When** 它分别通过 web 表单和 OpenAPI 提交，**Then** 两条链路都必须得出相同的校验结论与相同的资源身份解析结果，并且在入库前通过对应版本 JSON Schema 校验。
2. **Given** 一份在结构化字段与配置内容中表达冲突语义的资源定义，**When** 用户提交请求，**Then** 系统必须在入库前返回确定性的冲突错误，且不得在不同入口静默选择不同字段作为胜出值。
3. **Given** 某资源依赖自动生成标识、替代命名字段或版本特定字段支持规则，**When** 该资源通过任一入口提交，**Then** 系统必须以同一语义处理这些特例，并保持后续发布行为一致。
4. **Given** 已有线上 OpenAPI 调用方和现有 WebAPI 表单请求，**When** 本次重构上线，**Then** 默认不要求调用方或前端调整协议；若某个 WebAPI 协议调整被批准，相关前端 Vue 表单变更必须同时交付。

---

### User Story 3 - Keep Stored, Imported, And Historical Data Publishable (Priority: P2)

作为发布负责人，我需要当前草稿数据、导入数据和历史存量数据在重构后仍然能够被正确验证并生成合法发布数据，这样上线不会因为旧数据形态或阶段间规则不一致而中断发布。

**Why this priority**: 该系统已经承载线上真实数据。只保证新建数据正确还不够，存量数据和导入数据的兼容性才是决定能否安全上线的关键。

**Independent Test**: 使用代表性的历史数据夹具、导入数据夹具和正常草稿数据夹具，执行入库前校验、草稿持久化、发布前构造和最终发布校验，确认其结果满足兼容性预期。

**Acceptance Scenarios**:

1. **Given** 一条历史存量资源记录仍包含旧形态的重复字段或服务端字段副本，**When** 系统读取、更新或发布该资源，**Then** 资源必须仍可被正确处理，而不要求人工先清洗数据。
2. **Given** 一份能够进入草稿态的合法资源，**When** 系统准备最终发布数据，**Then** 最终发布数据必须在写入发布通道前通过对应版本 JSON Schema 校验，而不是依赖发布阶段临时修补差异。
3. **Given** 一份导入数据在当前系统中属于受支持输入，**When** 它在重构后被校验并进入后续生命周期，**Then** 它的可接受性和最终发布结果必须与既有兼容契约保持一致。

### Edge Cases

- 历史存量数据的草稿配置中仍保留重复的 `id`、`name`、关联字段或其他服务端字段副本，且这些副本与结构化字段可能一致也可能冲突。
- 某些资源在创建时需要生成标识，而更新时必须复用已解析的既有标识，不能在不同阶段生成两次不同结果。
- 某些资源使用替代命名字段或派生标识字段，要求 web、OpenAPI、导入和发布链路都遵守同一解释规则。
- 同一字段在一个受支持版本中合法、在另一个受支持版本中非法，系统必须对目标版本作出一致判断。
- 某请求若在草稿校验阶段被接受，不应仅因为生命周期后续阶段采用了另一套字段注入或清理规则而在最终发布时失败。
- 重构上线后，未重新保存过的旧资源仍可能直接参与发布，需要验证其在兼容路径中的行为。
- 若某项实现尝试依赖 OpenAPI 协议变更、WebAPI 表单协议变更或数据库字段变更来完成重构，系统必须在实施前显式暴露该影响，而不能作为隐含前提推进。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define one canonical edit-state representation for every managed resource and use it as the internal source of truth across request handling, draft persistence, and publish preparation.
- **FR-002**: System MUST create automated unit-test coverage for every in-scope validation, normalization, persistence-projection, and publish-preparation entrypoint before refactoring the behavior of that entrypoint.
- **FR-003**: System MUST cover all 11 managed resource types in the regression baseline, and MUST include every supported version-specific rule that changes acceptance, rejection, or final payload shape.
- **FR-004**: System MUST include both positive and negative baseline cases, and MUST assert observable outcomes that affect compatibility, including acceptance result, resolved identity, stored draft shape, and final publish payload shape where applicable.
- **FR-005**: System MUST preserve the pre-refactor result for every approved regression case unless a behavior change is explicitly documented, approved before release, and accompanied by compatibility guidance.
- **FR-006**: System MUST apply the same semantic rules to management-console submissions and OpenAPI submissions for the same resource type, operation, and target version.
- **FR-007**: When structured fields and configuration content express the same semantic field, the system MUST accept matching values and MUST reject conflicting values before persistence; it MUST NOT silently let different entrypoints pick different winners.
- **FR-008**: System MUST resolve resource identity exactly once per request lifecycle and MUST use that resolved identity consistently for validation, draft persistence, and publish preparation.
- **FR-009**: System MUST preserve compatibility with existing stored resources, including records that still contain legacy duplicated fields or server-owned fields in stored configuration, without requiring manual data migration before rollout.
- **FR-010**: System MUST validate management-console inputs against the corresponding target-version JSON Schema before persistence and MUST return deterministic, user-actionable errors when configuration content, associations, or supported-version rules are violated.
- **FR-011**: System MUST validate OpenAPI and import inputs before persistence using the same canonical resource semantics that govern management-console requests, and those accepted inputs MUST pass the corresponding target-version JSON Schema before persistence while preserving resource-specific acceptance rules already supported in production.
- **FR-012**: System MUST construct one authoritative final publish payload per resource and target version, and that payload MUST pass the corresponding target-version JSON Schema before any publish write occurs.
- **FR-013**: System MUST ensure that a resource accepted into draft state does not later fail final publish solely because different lifecycle stages assembled the payload using divergent field-injection or cleanup rules.
- **FR-014**: System MUST preserve currently supported resource-specific special cases, including alternate naming fields, derived metadata identity, generated identifiers, association fields, and version-specific field removal, unless a change is explicitly approved and documented.
- **FR-015**: System MUST provide acceptance evidence that representative historical resources can still be read, updated when allowed, and published successfully after the refactor.
- **FR-016**: System MUST support phased delivery of the new validation solution, but each delivered phase MUST maintain the regression baseline and MUST NOT introduce inconsistent validation outcomes between covered request flows and publish flows.
- **FR-017**: System MUST preserve the existing OpenAPI serializer contract for this refactor and MUST NOT require protocol changes for existing online callers.
- **FR-018**: System MUST preserve the existing WebAPI form contract by default; if a contract change is unavoidable, that change MUST be explicitly approved and delivered together with the corresponding frontend Vue form updates.
- **FR-019**: System MUST deliver the refactor without changing existing database fields by default; any unavoidable database field change MUST be explicitly identified before implementation begins.

### Key Entities *(include if feature involves data)*

- **Resource Input Request**: 来自 web 表单、OpenAPI 或导入流程的资源输入，包含结构化字段与资源配置内容，是外部输入而不是内部真相。
- **Canonical Draft Resource**: 系统内部用于编辑态流转和持久化的统一资源表示，承载资源身份、关联关系和用户真正编辑的配置语义。
- **Publish Payload**: 面向目标版本最终发布的数据表示，是发布前校验和写入发布通道所使用的唯一权威载荷。
- **Regression Baseline Suite**: 在重构前建立的自动化回归用例集合，用于固化当前接受、拒绝、持久化和发布行为。
- **Historical Compatibility Fixture**: 代表线上存量数据、旧字段形态和导入数据形态的测试夹具，用于证明兼容性不会因重构而破坏。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 在任何重构代码获准合入前，100% 已识别的在范围内校验与 payload 成形入口都已纳入自动化单元测试基线。
- **SC-002**: 对批准纳入回归基线的测试语料，100% 案例在重构前后保持相同的接受或拒绝结果，以及相同的关键字段成形结果，除非该差异已被明确批准并记录。
- **SC-003**: 对 11 类资源和每个存在版本差异的受支持版本，回归测试都至少包含 1 个合法案例和 1 个非法案例，且结果全部符合预期。
- **SC-004**: 代表性历史数据夹具在验收中 100% 能够被读取，并在允许的操作范围内完成更新或发布，且不需要因为本次重构而进行人工数据修复。
- **SC-005**: 所有纳入验收语料的合法 web、OpenAPI 与导入案例，都能在最终发布前生成通过目标版本校验的发布数据；所有非法案例都在最早适用阶段被确定性拒绝。
- **SC-006**: 所有纳入验收语料的合法入库案例，在真正写入前 100% 经过对应版本 JSON Schema 校验；所有合法发布案例，在真正发布到 APISIX 前 100% 经过对应版本 JSON Schema 校验。

## Assumptions

- 当前受支持的资源类型范围保持不变，仍为现有 11 类 APISIX 资源。
- 当前受支持的目标版本范围保持不变，本次功能不包含新增或下线平台版本。
- 现有 OpenAPI serializer 协议已经被线上调用依赖，本次功能默认保持兼容，不以协议调整作为前提。
- 现有 WebAPI 表单入参协议默认保持兼容；若必须调整，需同步交付前端 Vue 表单变更。
- 现有数据库字段视为既定边界，本次功能默认不包含数据库字段变更。
- 线上数据库中已经存在旧形态草稿数据，本次方案必须兼容这些数据在读取、更新、发布上的正常使用。
- 本次规格聚焦于 apiserver 后端的配置校验、草稿表示、发布表示和回归验收标准，不包含前端界面改版。
