# Quickstart: APIServer Config Validation Refactor

## 1. Review Scope

1. Read [spec.md](./spec.md), [plan.md](./plan.md), and [research.md](./research.md).
2. Confirm the non-breaking boundaries in [contracts/validation-compatibility.md](./contracts/validation-compatibility.md).
3. Treat the regression suite as the acceptance gate before any refactor behavior lands.

## 2. Build The Regression Baseline First

From `src/apiserver/` add or extend unit tests around the current validation and publish entrypoints, then run the focused suite:

```bash
cd src/apiserver
GOTOOLCHAIN=auto go test ./pkg/apis/web/serializer ./pkg/apis/open/serializer ./pkg/middleware ./pkg/apis/common ./pkg/biz ./pkg/entity/model ./pkg/publisher ./pkg/resourcecodec
```

Recommended coverage priorities:

1. Web request validation: `pkg/apis/web/serializer/common.go`
2. OpenAPI request validation: `pkg/middleware/openapi_resource_check.go`
3. Import validation and shared helpers: `pkg/biz/common.go`, `pkg/apis/common/resource_slz.go`
4. OpenAPI serializer payload building: `pkg/apis/open/serializer/resource.go`
5. Shared draft-preparation/payload-building core: `pkg/resourcecodec/*`
6. Model persistence projection hooks: `pkg/entity/model/*`
7. Publish payload construction and final validation: `pkg/biz/publish.go`, `pkg/publisher/etcd.go`

## 3. Implement The Shared Codec Layer

Create `pkg/resourcecodec/` and introduce:

- identity resolution helpers
- resource-draft preparation helpers
- version-aware payload-building helpers for `DATABASE` and `ETCD` validation profiles
- legacy compatibility helpers for stored rows that still contain duplicated config fields

Keep external request structs and database fields unchanged while wiring new logic behind existing entrypoints.

## 4. Runtime Shape After Refactor

当前已落地的目标链路是：

```text
request/import -> prepare draft -> build payload(DATABASE) -> schema validate -> persist
stored draft   -> legacy-tolerant build payload(ETCD) -> schema validate -> publish
```

兼容边界保持不变：

- 不改 OpenAPI serializer 对外协议
- 默认不改 WebAPI 表单协议
- 不改数据库字段
- 入库前必须通过对应版本 `DATABASE` JSON Schema 校验
- 发布前必须通过对应版本 `ETCD` JSON Schema 校验

## 5. Migrate Publish Before Request Paths

1. Move `pkg/biz/publish.go` resource-specific JSON surgery into codec-based payload building.
2. Preserve `pkg/publisher/etcd.go:Validate` as the final publish gate.
3. After publish parity is proven, route WebAPI, OpenAPI, and import validation through the same codec layer.

## 6. Verify Compatibility

Run the focused suite again, then run the broader unit suite:

```bash
cd src/apiserver
GOTOOLCHAIN=auto go test ./pkg/apis/web/serializer ./pkg/apis/open/serializer ./pkg/middleware ./pkg/apis/common ./pkg/biz ./pkg/entity/model ./pkg/publisher ./pkg/resourcecodec
GOTOOLCHAIN=auto go test ./...
```

If publish-path behavior or legacy compatibility is still uncertain, run the integration environment before merge:

```bash
cd src/apiserver/tests/integration
docker compose down
docker compose up --abort-on-container-exit
```

## 7. Release Checklist

- No OpenAPI serializer contract changes
- No unplanned WebAPI request changes
- No database field changes
- Regression baseline green before and after refactor
- Accepted persistence cases pass target-version JSON Schema validation before write
- Accepted publish cases pass target-version JSON Schema validation before etcd write
- Representative legacy fixtures remain readable and publishable
