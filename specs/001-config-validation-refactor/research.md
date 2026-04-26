# Research: APIServer Config Validation Refactor

## Decision 1: Introduce a canonical draft plus a dedicated `pkg/resourcecodec/` package

- **Decision**: Implement a new internal codec layer that resolves identity once, normalizes external input into a canonical draft representation, and materializes APISIX payloads for validation and publish.
- **Rationale**: Current behavior is split across request validators, middleware, serializers, model hooks, and publish functions. A shared codec layer reduces repeated JSON surgery and lets request-time validation and publish-time validation reason about the same payload family.
- **Alternatives considered**:
  - Keep patching the current per-layer functions: rejected because the same field conflicts would continue to be solved multiple times with inconsistent precedence.
  - Build only a table-driven rule map inside existing functions: rejected because it centralizes some field rules but does not solve identity resolution, payload ownership, or legacy-data compatibility.

## Decision 2: Preserve OpenAPI, WebAPI, and database contracts as outer boundaries

- **Decision**: Keep the existing OpenAPI serializer protocol unchanged, keep WebAPI request payloads unchanged by default, and avoid database field changes.
- **Rationale**: All three boundaries are already in production use. The feature goal is to make validation behavior coherent without forcing online callers, frontend forms, or database migrations to move first.
- **Alternatives considered**:
  - Change OpenAPI serializer payloads to match the new canonical draft: rejected because there are already online callers.
  - Change WebAPI request shapes immediately: rejected because it would couple backend refactor risk with frontend rollout risk.
  - Add or reshape database fields for canonical draft storage now: rejected because the current requirement explicitly avoids schema change unless proven unavoidable.

## Decision 3: Build the regression baseline before refactoring behavior

- **Decision**: Missing unit tests must be added first for every current config-shaping entrypoint that materially affects acceptance, stored draft shape, or publish payload shape.
- **Rationale**: This is a production refactor with legacy data. The fastest way to lose compatibility is to refactor first and discover implicit behavior later. Baseline tests turn existing behavior into executable acceptance criteria.
- **Alternatives considered**:
  - Rely only on integration tests: rejected because most compatibility bugs in this area come from small per-layer payload mutations that are easier to localize with unit tests.
  - Add tests only for the new codec package: rejected because that would leave current behavior unpinned and make it impossible to prove parity.

## Decision 4: Use one materialization path with different validation profiles

- **Decision**: Materialize one authoritative payload family per resource and target version, then validate it with `DATABASE` profile before persistence and `ETCD` profile before publish.
- **Rationale**: The user requires target-version JSON Schema validation before persistence and before publish. Reusing one materialization path keeps request-time and publish-time validation aligned while still respecting the stricter top-level constraints of the `ETCD` profile.
- **Alternatives considered**:
  - Keep separate request-time and publish-time payload builders: rejected because this is the current source of divergence.
  - Validate only on publish: rejected because invalid data would still enter draft storage.
  - Validate request-time payloads directly with `ETCD` profile: rejected as the default because draft storage may still need to preserve edit-state distinctions better handled by `DATABASE` profile, while the materialized payload itself remains shared.

## Decision 5: New writes should stop depending on server-owned echo fields, but legacy rows must still work

- **Decision**: Treat model columns and structured request fields as authoritative for identity and associations; stop depending on echoed server-owned fields inside new draft config writes, but tolerate and correctly interpret those legacy duplicates when reading old rows.
- **Rationale**: This reduces future complexity without breaking existing data. It also avoids making stored draft config the accidental source of truth for server-owned metadata.
- **Alternatives considered**:
  - Continue writing server-owned fields back into every draft config forever: rejected because it keeps publish cleanup and validation divergence alive.
  - Require a one-shot data migration before rollout: rejected because the feature explicitly must remain compatible with existing online data.

## Decision 6: Migrate in the order publish -> request/import -> persistence cleanup

- **Decision**: First move publish assembly to the new codec layer, then reuse the same layer for request/import validation, and finally reduce persistence echo behavior for new writes.
- **Rationale**: Publish is currently the most fragmented area and also the final correctness gate before APISIX. Moving it first gives the highest risk reduction and proves the materializer design before request paths depend on it.
- **Alternatives considered**:
  - Refactor request validation first: rejected because publish would still have a second divergent payload builder.
  - Refactor model hooks first: rejected because it would change stored data shape before the shared materializer exists to absorb legacy compatibility.

## Decision 7: Treat target versions `3.11.X` and `3.13.X` as the primary acceptance matrix

- **Decision**: Plan the regression suite and codec rules around the resource/version differences that matter in the currently active schemas, especially `3.11.X` and `3.13.X`, while preserving existing helper behavior for `3.2.X` and `3.3.X` code paths that still exist in schema utilities.
- **Rationale**: The current production-oriented solution and compatibility notes are centered on `3.11` and `3.13`, but the codebase still exposes helper and schema assets for older versions. The new design must not accidentally regress those helpers even if they are not the main rollout target.
- **Alternatives considered**:
  - Ignore older schema helpers entirely: rejected because that could break currently compiled validation paths or tests.
  - Treat all historical versions as first-class rollout targets: rejected because it would dilute effort away from the explicitly clarified acceptance focus.
