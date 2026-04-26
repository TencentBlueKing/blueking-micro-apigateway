# Contract: Validation Compatibility

## Purpose

Define the non-breaking contracts this refactor must preserve across external request surfaces, draft persistence, and APISIX publish behavior.

## External Surface Contract

| Surface | Compatibility rule | Validation requirement before persistence |
| ------- | ------------------ | ----------------------------------------- |
| WebAPI form | Existing request payload shape remains unchanged by default. If a shape change is unavoidable, the matching frontend Vue form change must ship together. | Prepare a resource draft, build the target-version payload, and pass target-version JSON Schema validation before database write. |
| OpenAPI serializer | Existing serializer contract remains unchanged for online callers. | Prepare a resource draft without changing wire format, build the target-version payload, and pass target-version JSON Schema validation before database write. |
| Import payload | Existing supported import format remains accepted. | Prepare a resource draft, build the target-version payload, and pass target-version JSON Schema validation before database write. |

## Persistence Contract

- Existing database fields are fixed by default for this feature.
- Model columns remain authoritative for identity and association data.
- New writes should not depend on duplicated server-owned fields inside config to preserve correctness.
- Legacy stored rows that still contain duplicated `id`, `name`, or association fields inside config must remain readable and publishable.
- For new external inputs, conflicting values between structured fields and config are validation errors.

## Publish Contract

- Final APISIX publish payloads are built from the resource draft plus authoritative model columns.
- Every publish payload must pass target-version JSON Schema validation before etcd write.
- Publish ordering for dependent resources remains intact: upstream/service/plugin_config or consumer_group dependencies still publish before the referencing resource.
- Resource/version special cases remain preserved for approved regression cases.

## Regression Contract

The following entrypoints require regression coverage before their behavior changes:

- `pkg/apis/web/serializer/common.go:CheckAPISIXConfig`
- `pkg/middleware/openapi_resource_check.go:OpenAPIResourceCheck`
- `pkg/biz/common.go:BuildConfigRawForValidation`
- `pkg/biz/common.go:ValidateResource`
- `pkg/apis/open/serializer/resource.go:ToCommonResource`
- `pkg/entity/model/*:HandleConfig`
- `pkg/biz/publish.go` resource-specific `putXxx` / `PutXxx` functions
- `pkg/publisher/etcd.go:EtcdPublisher.Validate`

## Approval Boundary

Any proposal that requires one of the following must be called out explicitly before implementation proceeds:

- OpenAPI serializer contract change
- WebAPI request contract change
- Frontend Vue form coordination
- Database field or schema change
- Intentional regression-case behavior change
