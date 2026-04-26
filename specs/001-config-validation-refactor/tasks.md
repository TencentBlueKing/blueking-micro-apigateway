# Tasks: APIServer Config Validation Refactor

**Input**: Design documents from `/specs/001-config-validation-refactor/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are required for this feature. The specification explicitly requires backfilling unit-test coverage for current behavior before refactoring and preserving the same accepted/rejected cases after the refactor.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Each task includes the exact file path to change

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create shared scaffolding for the regression matrix and the new internal codec package without changing runtime behavior.

- [X] T001 Create the shared regression case catalog in `src/apiserver/pkg/utils/testing/validation_cases.go`
- [X] T002 [P] Create the historical and legacy fixture loader helpers in `src/apiserver/pkg/utils/testing/validation_fixtures.go`
- [X] T003 [P] Create the new codec package scaffold in `src/apiserver/pkg/resourcecodec/doc.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define the internal codec abstractions that later stories will implement and wire into request/import/publish flows.

**⚠️ CRITICAL**: No request-path or publish-path refactor should start until this phase is complete.

- [X] T004 Define canonical draft, resolved identity, and materialization interfaces in `src/apiserver/pkg/resourcecodec/types.go`
- [X] T005 [P] Add shared version/profile helper utilities in `src/apiserver/pkg/resourcecodec/common.go` and `src/apiserver/pkg/constant/resource_schema.go`
- [X] T006 [P] Add legacy row comparison and duplicate-field helper functions in `src/apiserver/pkg/resourcecodec/legacy.go`

**Checkpoint**: Internal abstractions exist; current runtime behavior is still unchanged.

---

## Phase 3: User Story 1 - Freeze Current Behavior With A Regression Baseline (Priority: P1) 🎯 MVP

**Goal**: Freeze the current validation and publish behavior in automated regression tests before any behavior-changing refactor lands.

**Independent Test**: Run `go test ./pkg/apis/web/serializer ./pkg/apis/open/serializer ./pkg/middleware ./pkg/biz ./pkg/entity/model ./pkg/publisher` from `src/apiserver/` and verify the baseline covers existing request validation, import validation, model hooks, publish payload shaping, and final publish validation behavior.

### Tests for User Story 1 ⚠️

> **NOTE: Write these tests FIRST, ensure they fail or expose missing coverage before implementation changes begin.**

- [X] T007 [P] [US1] Expand WebAPI config-validation regression coverage in `src/apiserver/pkg/apis/web/serializer/common_test.go`
- [X] T008 [P] [US1] Create OpenAPI request-validation regression coverage in `src/apiserver/pkg/middleware/openapi_resource_check_test.go`
- [X] T009 [P] [US1] Expand shared import and validation regression coverage in `src/apiserver/pkg/biz/common_test.go`
- [X] T010 [P] [US1] Create OpenAPI serializer regression coverage in `src/apiserver/pkg/apis/open/serializer/resource_test.go`
- [X] T011 [P] [US1] Expand existing `HandleConfig` regression coverage in `src/apiserver/pkg/entity/model/route_test.go`, `src/apiserver/pkg/entity/model/service_test.go`, `src/apiserver/pkg/entity/model/upstream_test.go`, `src/apiserver/pkg/entity/model/consumer_test.go`, `src/apiserver/pkg/entity/model/consumer_group_test.go`, `src/apiserver/pkg/entity/model/plugin_config_test.go`, `src/apiserver/pkg/entity/model/global_rule_test.go`, and `src/apiserver/pkg/entity/model/plugin_metadata_test.go`
- [X] T012 [P] [US1] Add missing `HandleConfig` regression coverage in `src/apiserver/pkg/entity/model/proto_test.go`, `src/apiserver/pkg/entity/model/ssl_test.go`, and `src/apiserver/pkg/entity/model/stream_route_test.go`
- [X] T013 [P] [US1] Expand publish payload regression coverage in `src/apiserver/pkg/biz/publish_test.go`
- [X] T014 [P] [US1] Expand final publish JSON Schema validation coverage in `src/apiserver/pkg/publisher/etcd_test.go`

### Implementation for User Story 1

- [X] T015 [US1] Stabilize the focused regression suite and shared assertions using `src/apiserver/Makefile` and the touched `src/apiserver/pkg/**/*_test.go` files

**Checkpoint**: Current behavior is pinned by automated tests and can be compared before and after refactor.

---

## Phase 4: User Story 2 - Validate Web And OpenAPI Inputs With One Consistent Semantics (Priority: P1)

**Goal**: Introduce one shared normalization and DATABASE-validation path for WebAPI, OpenAPI, and import inputs while preserving external contracts.

**Independent Test**: Submit equivalent fixtures through WebAPI, OpenAPI, and import-oriented validation tests and verify the same acceptance result, resolved identity, and target-version DATABASE-schema validation outcome.

### Tests for User Story 2 ⚠️

- [X] T016 [P] [US2] Add codec identity-resolution and conflict-detection tests in `src/apiserver/pkg/resourcecodec/common_test.go`
- [X] T017 [P] [US2] Add WebAPI and OpenAPI parity tests in `src/apiserver/pkg/apis/web/serializer/common_test.go` and `src/apiserver/pkg/middleware/openapi_resource_check_test.go`
- [X] T018 [P] [US2] Add import normalization parity tests in `src/apiserver/pkg/biz/common_test.go`

### Implementation for User Story 2

- [X] T019 [P] [US2] Implement shared normalization and materialization core in `src/apiserver/pkg/resourcecodec/common.go` and `src/apiserver/pkg/constant/resource_schema.go`
- [X] T020 [P] [US2] Implement route, service, upstream, and plugin_config codecs in `src/apiserver/pkg/resourcecodec/route.go`, `src/apiserver/pkg/resourcecodec/service.go`, `src/apiserver/pkg/resourcecodec/upstream.go`, and `src/apiserver/pkg/resourcecodec/plugin_config.go`
- [X] T021 [P] [US2] Implement consumer, consumer_group, global_rule, and plugin_metadata codecs in `src/apiserver/pkg/resourcecodec/consumer.go`, `src/apiserver/pkg/resourcecodec/consumer_group.go`, `src/apiserver/pkg/resourcecodec/global_rule.go`, and `src/apiserver/pkg/resourcecodec/plugin_metadata.go`
- [X] T022 [P] [US2] Implement proto, ssl, and stream_route codecs in `src/apiserver/pkg/resourcecodec/proto.go`, `src/apiserver/pkg/resourcecodec/ssl.go`, and `src/apiserver/pkg/resourcecodec/stream_route.go`
- [X] T023 [US2] Wire WebAPI request validation through codec-based DATABASE materialization in `src/apiserver/pkg/apis/web/serializer/common.go`
- [X] T024 [US2] Wire OpenAPI request validation through codec-based DATABASE materialization in `src/apiserver/pkg/middleware/openapi_resource_check.go`
- [X] T025 [US2] Wire shared import validation through codec-based DATABASE materialization in `src/apiserver/pkg/biz/common.go` and `src/apiserver/pkg/apis/common/resource_slz.go`
- [X] T026 [US2] Preserve the OpenAPI serializer wire contract while delegating normalization helpers in `src/apiserver/pkg/apis/open/serializer/resource.go`

**Checkpoint**: WebAPI, OpenAPI, and import validation share one semantic path without breaking the existing request contracts.

---

## Phase 5: User Story 3 - Keep Stored, Imported, And Historical Data Publishable (Priority: P2)

**Goal**: Make publish-time payload construction use the same codec layer, preserve legacy stored data compatibility, and guarantee target-version ETCD-schema validation before publish.

**Independent Test**: Run legacy stored-data and publish regression tests to verify representative historical rows remain readable and publishable, and that final APISIX payloads pass target-version ETCD validation before etcd writes.

### Tests for User Story 3 ⚠️

- [X] T027 [P] [US3] Add legacy stored-row compatibility tests in `src/apiserver/pkg/resourcecodec/legacy_test.go`
- [X] T028 [P] [US3] Add publish materialization compatibility tests in `src/apiserver/pkg/biz/publish_test.go`
- [X] T029 [P] [US3] Add final ETCD JSON Schema gate tests in `src/apiserver/pkg/publisher/etcd_test.go`
- [X] T030 [P] [US3] Add historical import and legacy fixture compatibility tests in `src/apiserver/pkg/biz/common_test.go`

### Implementation for User Story 3

- [X] T031 [US3] Implement legacy duplicate-field tolerance and draft dematerialization helpers in `src/apiserver/pkg/resourcecodec/legacy.go`
- [X] T032 [US3] Migrate publish payload assembly to codec materialization in `src/apiserver/pkg/biz/publish.go`
- [X] T033 [US3] Route final APISIX publish validation through shared materialized payload handling in `src/apiserver/pkg/publisher/etcd.go`
- [X] T034 [US3] Stop new writes from depending on echoed server-owned config fields in `src/apiserver/pkg/entity/model/common.go` and the resource model files under `src/apiserver/pkg/entity/model/`

**Checkpoint**: Publish and legacy compatibility now use the same internal payload model while keeping old rows operational.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final cleanup, documentation, and verification across all user stories.

- [X] T035 [P] Update backend validation pipeline documentation in `src/apiserver/README.md` and `docs/DEVELOP_GUIDE.md`
- [X] T036 Clean up retired validation helpers and comments in `src/apiserver/pkg/apis/web/serializer/common.go`, `src/apiserver/pkg/middleware/openapi_resource_check.go`, `src/apiserver/pkg/biz/common.go`, and `src/apiserver/pkg/biz/publish.go`
- [X] T037 [P] Update final verification steps in `specs/001-config-validation-refactor/quickstart.md` using the executable commands from `src/apiserver/Makefile`
- [X] T038 Run the focused regression suite and full backend unit suite via `src/apiserver/Makefile`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies; start immediately.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user story work.
- **User Story 1 (Phase 3)**: Depends on Foundational; must complete before behavior-changing refactors begin.
- **User Story 2 (Phase 4)**: Depends on User Story 1 because the regression baseline is the acceptance gate for request-path changes.
- **User Story 3 (Phase 5)**: Depends on User Story 1 and on the shared codec core from User Story 2; publish-path work can start once T019-T022 are complete.
- **Polish (Phase 6)**: Depends on all targeted user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependency on other stories; this is the MVP and the gating acceptance baseline.
- **User Story 2 (P1)**: Requires User Story 1 baseline coverage and the foundational codec abstractions.
- **User Story 3 (P2)**: Requires User Story 1 baseline coverage plus the codec core introduced for User Story 2.

### Within Each User Story

- Test tasks MUST be written before the corresponding implementation tasks.
- Shared codec primitives must land before request-path or publish-path wiring.
- Publish refactors must not begin until target-version DATABASE validation is already consistent across request paths.
- Each story should pass its independent test criteria before moving on.

### Parallel Opportunities

- Setup tasks T002-T003 can run in parallel.
- Foundational tasks T005-T006 can run in parallel after T004 begins.
- User Story 1 test tasks T007-T014 can run in parallel across different packages.
- User Story 2 test tasks T016-T018 can run in parallel, and codec file tasks T020-T022 can run in parallel once T019 establishes the shared core.
- User Story 3 test tasks T027-T030 can run in parallel, and T032-T033 can proceed after T031 prepares legacy helpers.
- Polish tasks T035 and T037 can run in parallel while the final verification task T038 waits for code completion.

---

## Parallel Example: User Story 1

```bash
# Launch package-level regression work in parallel:
Task: "Expand WebAPI config-validation regression coverage in src/apiserver/pkg/apis/web/serializer/common_test.go"
Task: "Create OpenAPI request-validation regression coverage in src/apiserver/pkg/middleware/openapi_resource_check_test.go"
Task: "Create OpenAPI serializer regression coverage in src/apiserver/pkg/apis/open/serializer/resource_test.go"
Task: "Expand publish payload regression coverage in src/apiserver/pkg/biz/publish_test.go"
Task: "Expand final publish JSON Schema validation coverage in src/apiserver/pkg/publisher/etcd_test.go"
```

---

## Parallel Example: User Story 2

```bash
# Launch parity tests together:
Task: "Add codec identity-resolution and conflict-detection tests in src/apiserver/pkg/resourcecodec/common_test.go"
Task: "Add WebAPI and OpenAPI parity tests in src/apiserver/pkg/apis/web/serializer/common_test.go and src/apiserver/pkg/middleware/openapi_resource_check_test.go"
Task: "Add import normalization parity tests in src/apiserver/pkg/biz/common_test.go"

# After the codec core lands, split per-resource codec files:
Task: "Implement route, service, upstream, and plugin_config codecs in src/apiserver/pkg/resourcecodec/route.go, service.go, upstream.go, and plugin_config.go"
Task: "Implement consumer, consumer_group, global_rule, and plugin_metadata codecs in src/apiserver/pkg/resourcecodec/consumer.go, consumer_group.go, global_rule.go, and plugin_metadata.go"
Task: "Implement proto, ssl, and stream_route codecs in src/apiserver/pkg/resourcecodec/proto.go, ssl.go, and stream_route.go"
```

---

## Parallel Example: User Story 3

```bash
# Launch legacy and publish safety tests together:
Task: "Add legacy stored-row compatibility tests in src/apiserver/pkg/resourcecodec/legacy_test.go"
Task: "Add publish materialization compatibility tests in src/apiserver/pkg/biz/publish_test.go"
Task: "Add final ETCD JSON Schema gate tests in src/apiserver/pkg/publisher/etcd_test.go"
Task: "Add historical import and legacy fixture compatibility tests in src/apiserver/pkg/biz/common_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational.
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: Confirm the baseline regression suite is green and captures the current behavior.

### Incremental Delivery

1. Build shared scaffolding and abstractions without changing behavior.
2. Freeze current behavior with User Story 1 tests.
3. Deliver consistent request/import validation with User Story 2.
4. Deliver publish and legacy compatibility with User Story 3.
5. Finish with cross-cutting cleanup and full verification.

### Parallel Team Strategy

1. One engineer owns the shared regression fixtures and codec abstractions (T001-T006).
2. One engineer can drive WebAPI/OpenAPI/import baseline coverage (T007-T010) while another expands model and publish coverage (T011-T014).
3. After T019 lands, per-resource codec files T020-T022 can be split across multiple engineers.
4. After T031 lands, publish migration T032-T033 and persistence cleanup T034 can proceed in parallel with doc updates T035-T037.

## Notes

- `[P]` tasks touch different files and can be parallelized safely.
- User Story 1 is the suggested MVP because it establishes the non-negotiable regression baseline.
- Do not refactor request or publish behavior until the corresponding baseline tests exist.
- Preserve OpenAPI protocol compatibility, keep WebAPI stable by default, and avoid database field changes unless explicitly approved.
