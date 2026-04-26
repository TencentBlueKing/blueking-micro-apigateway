# Implementation Plan: APIServer Config Validation Refactor

**Branch**: `[001-config-validation-refactor]` | **Date**: 2026-04-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-config-validation-refactor/spec.md`

## Summary

Implement a unified apiserver config-validation pipeline that preserves the existing OpenAPI serializer contract, keeps the WebAPI request contract stable by default, avoids database schema changes, and guarantees target-version JSON Schema validation both before persistence and before publish. Delivery is phased: first freeze current behavior with missing unit-test baselines across existing entrypoints, then introduce a canonical draft plus shared materialization layer, migrate publish and request/import validation onto that layer, and finally stop new writes from depending on scattered config echo rules while remaining compatible with legacy stored data.

## Technical Context

**Language/Version**: Go 1.25.5  
**Primary Dependencies**: Gin, go-playground/validator/v10, GORM, gjson/sjson, xeipuuv/gojsonschema, testify, gomega/ginkgo, APISIX versioned schema assets under `pkg/utils/schema/`  
**Storage**: MySQL resource tables for edit-state persistence plus etcd/APISIX publish storage  
**Testing**: `go test`, package-local `*_test.go`, testify, gomega/ginkgo, targeted integration tests under `src/apiserver/tests/`  
**Target Platform**: Linux server running the Go apiserver against MySQL and APISIX/etcd  
**Project Type**: Single backend web-service / control plane  
**Performance Goals**: Preserve current synchronous request-validation and publish-path latency characteristics; avoid adding extra network round trips in validation hot paths; keep validation outcomes deterministic for all covered regression cases  
**Constraints**: Keep OpenAPI serializer protocol unchanged; keep WebAPI form protocol unchanged by default; keep database fields unchanged by default; every accepted persistence path must pass target-version JSON Schema validation before write; every publish path must pass target-version JSON Schema validation before etcd write; support legacy stored configs without pre-migration  
**Scale/Scope**: 11 APISIX resource types; request/import/publish/model-hook validation touchpoints across `pkg/apis/web`, `pkg/apis/open`, `pkg/middleware`, `pkg/biz`, `pkg/entity/model`, `pkg/publisher`, and `pkg/utils/schema`; active target versions centered on APISIX `3.11.X` and `3.13.X` while preserving existing helper behavior for `3.2.X` and `3.3.X`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- `.specify/memory/constitution.md` is still an unfilled template, so there are no enforceable project-specific constitutional rules to fail against.
- Effective gates for this feature are taken from the clarified specification and repository instructions:
  - Regression tests must be added before behavior-changing refactors.
  - OpenAPI serializer protocol must remain backward compatible.
  - WebAPI request protocol must remain stable unless frontend Vue changes ship together.
  - Database fields are fixed by default.
  - Target-version JSON Schema validation is mandatory before persistence and before publish.
  - Legacy stored resources must stay readable, updatable where allowed, and publishable.
- Phase 0 gate result: `PASS`.
- Phase 1 design re-check: `PASS`. The design keeps representation changes internal to the backend and does not require API or database contract changes.

## Project Structure

### Documentation (this feature)

```text
specs/001-config-validation-refactor/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
│   └── validation-compatibility.md
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
src/apiserver/
├── pkg/apis/web/serializer/common.go
├── pkg/apis/web/handler/{consumer_group.go,global_rule.go,plugin_config.go}
├── pkg/apis/open/serializer/resource.go
├── pkg/apis/common/resource_slz.go
├── pkg/middleware/openapi_resource_check.go
├── pkg/biz/common.go
├── pkg/biz/publish.go
├── pkg/publisher/etcd.go
├── pkg/entity/model/*.go
├── pkg/constant/resource_schema.go
├── pkg/utils/schema/
├── pkg/resourcecodec/                 # new package planned for canonicalize/materialize logic
└── tests/
    ├── integration/
    └── util/
```

**Structure Decision**: Keep the existing single-service backend structure. Introduce a new `pkg/resourcecodec/` package for canonical identity resolution, draft normalization, and APISIX payload materialization, while keeping request handlers, serializers, middleware, biz orchestration, publisher code, and model hooks as thin adapters around that shared layer. Add regression tests close to current entrypoints in their existing package-local test files.

## Implementation Strategy

### Phase 0 - Freeze Current Behavior

1. Add or extend unit tests around all currently active config-shaping entrypoints before refactoring behavior:
   - `pkg/apis/web/serializer/common.go:CheckAPISIXConfig`
   - `pkg/middleware/openapi_resource_check.go:OpenAPIResourceCheck`
   - `pkg/biz/common.go:BuildConfigRawForValidation` and `ValidateResource`
   - `pkg/apis/open/serializer/resource.go:ToCommonResource`
   - `pkg/entity/model/*:HandleConfig`
   - `pkg/biz/publish.go` resource-specific `putXxx` / `PutXxx` functions
   - `pkg/publisher/etcd.go:EtcdPublisher.Validate`
2. Build a regression matrix that covers:
   - 11 resource types
   - web input, OpenAPI input, import input, stored draft, publish payload
   - supported version differences and resource-specific special cases
   - representative legacy stored-data fixtures

### Phase 1 - Introduce Canonical Draft And Materialization

1. Add `pkg/resourcecodec/` with per-resource adapters behind shared interfaces:
   - `ResolveIdentity(...)`
   - `NormalizeDraft(...)`
   - `Materialize(..., profile)`
   - legacy-tolerant helpers for dematerializing existing stored configs when needed
2. Define one canonical draft representation used internally for:
   - request normalization
   - import normalization
   - draft persistence projections
   - publish payload generation
3. Keep transport boundaries unchanged:
   - OpenAPI serializer structs remain wire-compatible
   - WebAPI form payloads remain wire-compatible by default
   - database schema remains unchanged

### Phase 2 - Migrate Publish And Validation Paths

1. Migrate publish assembly in `pkg/biz/publish.go` to use `resourcecodec.Materialize(..., ETCD)` rather than per-resource JSON surgery.
2. Migrate request/import validation paths to use resolved identity + normalized draft + `Materialize(..., DATABASE)` before persistence.
3. Preserve `EtcdPublisher.Validate` as the final publish gate, but ensure it consumes the same materialized payload family used earlier in the flow.

### Phase 3 - Reduce Persistence Echo Logic Without Breaking Legacy Data

1. Stop new writes from relying on scattered `HandleConfig()` field echo behavior for server-owned fields.
2. Keep legacy tolerance in the new codec layer so old rows that still contain duplicated `id`, `name`, association fields, or resource-specific echo fields remain publishable.
3. Retain model columns as the authoritative source for identity and associations whenever legacy stored config duplicates are present.

## Test Baseline Scope

- Resource coverage: `route`, `service`, `upstream`, `consumer`, `consumer_group`, `plugin_config`, `global_rule`, `plugin_metadata`, `proto`, `ssl`, `stream_route`
- Validation surfaces: WebAPI request, OpenAPI request, import validation, stored draft compatibility, publish payload validation
- Version coverage: at minimum all resource/version combinations that differ between `3.11.X` and `3.13.X`; preserve existing helper behavior for `3.2.X` and `3.3.X` where schema helpers still expose those paths
- Behavior assertions:
  - acceptance vs rejection
  - resolved identity and associations
  - stored draft shape for new writes
  - final publish payload shape
  - target-version JSON Schema validation result before persistence and before publish

## Complexity Tracking

No constitution violations are currently identified. No exception table is required at planning time.
