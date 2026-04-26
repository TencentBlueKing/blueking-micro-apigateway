# Data Model: APIServer Config Validation Refactor

## Overview

This feature does not introduce new database tables. It introduces new internal representations and compatibility rules that sit on top of the existing resource tables and lifecycle states.

## Entities

### ExternalResourceInput

- **Purpose**: Represents incoming resource data from WebAPI forms, OpenAPI requests, or import payloads.
- **Fields**:
  - `source`: `web` | `openapi` | `import`
  - `resourceType`: one of the 11 APISIX resource types
  - `operation`: `create` | `update` | `import`
  - `gatewayID`
  - `targetVersion`
  - `pathID` or equivalent route identity when present
  - `outerFields`: structured request fields such as `name`, `username`, `service_id`, `upstream_id`, `plugin_config_id`, `group_id`
  - `configRaw`: original JSON config payload
- **Validation rules**:
  - Wire format must remain compatible with current external contracts.
  - Input is not the internal source of truth.
  - Matching duplicated semantic fields are allowed; conflicting duplicated semantic fields are rejected before persistence.

### ResolvedIdentity

- **Purpose**: The single authoritative identity and association resolution result for one request lifecycle.
- **Fields**:
  - `resourceType`
  - `resourceID`
  - `nameValue`
  - `nameKey` such as `name` or `username`
  - `associations`: `service_id`, `upstream_id`, `plugin_config_id`, `group_id`, or resource-specific equivalents
  - `resolvedFrom`: `path`, `structured field`, `generated`, or `legacy row`
- **Validation rules**:
  - Create flows may generate IDs exactly once.
  - Update flows must treat the path or existing row identity as authoritative.
  - Conflicting identity or association values between structured fields and config are rejected for new external inputs.

### ResourceDraft

- **Purpose**: The internal edit-state representation used after request preparation and before publish payload building.
- **Fields**:
  - `gatewayID`
  - `resourceType`
  - `resourceID`
  - `storedName`
  - `storedAssociations`
  - `configSpec`: user-editable spec data without relying on server-owned echo fields for new writes
  - `status`: existing lifecycle status such as `create_draft`, `update_draft`, `delete_draft`, `success`
  - `legacyEchoesPresent`: boolean or equivalent compatibility signal when legacy rows still contain duplicated fields in config
- **Validation rules**:
  - New writes keep database columns authoritative for identity and associations.
  - Legacy rows may still contain duplicated server-owned fields inside config and must remain readable.
  - The resource draft is the internal source of truth used by request validation, import validation, and publish preparation.

### BuiltPayload

- **Purpose**: Version-aware payload generated from the resource draft for schema validation.
- **Fields**:
  - `resourceType`
  - `targetVersion`
  - `profile`: `DATABASE` or `ETCD`
  - `payloadRaw`
  - `derivedDependencies`: referenced resources needed for association or publish ordering
- **Validation rules**:
  - Before persistence, accepted inputs must produce a payload that passes target-version JSON Schema validation with `DATABASE` profile.
  - Before publish, final payloads must pass target-version JSON Schema validation with `ETCD` profile.
  - Resource-specific field removal and version-aware field inclusion happen here rather than in scattered callers.

### HistoricalCompatibilityFixture

- **Purpose**: Captures representative online stored rows or imported payloads that use legacy config shapes.
- **Fields**:
  - `resourceType`
  - `targetVersion`
  - `storedColumns`
  - `storedConfig`
  - `expectedOutcome`: readable, updatable-if-allowed, publishable, or explicitly rejected
- **Validation rules**:
  - Fixtures must cover duplicated identity/name fields, association echoes, resource-specific special cases, and version-specific payload differences.
  - Compatibility decisions must prefer authoritative model columns over stale duplicated legacy config fields when reading stored data.

### RegressionBaselineCase

- **Purpose**: Defines the automated parity contract that must hold before and after refactor.
- **Fields**:
  - `caseID`
  - `surface`: `web`, `openapi`, `import`, `stored-draft`, `publish`
  - `resourceType`
  - `targetVersion`
  - `inputFixture`
  - `expectedAcceptanceResult`
  - `expectedResolvedIdentity`
  - `expectedStoredDraftShape`
  - `expectedPublishPayloadShape`
  - `expectedSchemaValidationResult`
- **Validation rules**:
  - Every approved baseline case must continue to pass or fail the same way after refactor unless a change is explicitly approved and documented.

## Relationships

```text
ExternalResourceInput
  -> ResolvedIdentity
  -> ResourceDraft
  -> BuiltPayload (DATABASE)
  -> Stored draft row / existing resource table row
  -> BuiltPayload (ETCD)
  -> APISIX publish payload

HistoricalCompatibilityFixture
  -> ResourceDraft
  -> BuiltPayload

RegressionBaselineCase
  -> asserts all observable transitions above
```

## State Transitions

Existing resource lifecycle states remain unchanged:

- `create_draft`
- `update_draft`
- `delete_draft`
- `success`

Validation and payload-building refactor changes how payloads are formed, not the lifecycle state machine itself.

## Resource-Specific Notes

- `consumer` uses `username` rather than `id` as the publish-side identifier field.
- `plugin_metadata` derives publish identity from plugin name.
- `consumer_group`, `plugin_config`, and `global_rule` have version-sensitive `id` requirements.
- `proto`, `consumer_group`, and `stream_route` have version-sensitive `name` support.
- `ssl` and `stream_route` require removal of internal-only fields before final publish payload building.
