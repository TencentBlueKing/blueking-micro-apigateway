# AGENTS.md

## Scope

This file applies to everything under `pkg/utils/schema/`, especially every new version folder such as `3.14/` or later.

Read this file before adding or refreshing any `<version>/plugin.json` or `<version>/schema.json`.

## Why This File Exists

The branch `fix_schema_validate_bugs` fixed a recurring class of problems caused by copying upstream APISIX JSON files without reconciling them with this repository's validator, plugin catalog, and test expectations.

Future version bumps must preserve the same fixes so the same bugs are not reintroduced.

## Branch Summary To Preserve

- `a795173` fixed the broken `proxy-rewrite` schema in `3.11/schema.json` and `3.13/schema.json`.
- `6d3ead3` refreshed `3.11` and `3.13` from newer upstream schema sources and restored missing core definitions used by our validator.
- `0c6189f` corrected plugin examples so they validate against the shipped schemas and filled in missing definitions.
- `623ff87` fixed `openfunction.authorization` so it is an object whose `service_token` is nested under `properties`.
- `5f09433` removed the non-standard `encrypt_fields` entries from plugin schemas.

These are not one-off fixes. Treat them as authoring rules for every new APISIX version added here.

## Directory Contract

- `plugin.json` is a plugin catalog consumed by `plugin.go`.
- `schema.json` is the versioned APISIX resource and plugin schema consumed by `schema.go` and `validate.go`.
- The code expects stable gjson paths such as:
  - `main.<resource>`
  - `plugins.<name>.schema`
  - `plugins.<name>.consumer_schema`
  - `plugins.<name>.metadata_schema`
  - `stream_plugins.<name>.schema`

If those paths move or disappear, validation and example lookup will silently break.

## When Adding A New Version

- Start from the upstream APISIX files, but normalize them to this repository before committing.
- Add the new embedded files and version maps in `plugin.go` and `schema.go`.
- Update `version.json` so the new version is advertised as supported.
- Update the test version lists in `plugin_test.go` and `validate_test.go` when the new version should be covered.
- Run `go test ./pkg/utils/schema` before considering the update complete.

## Rules For `plugin.json`

### Do

- Keep each entry compatible with the `Plugin` struct in `plugin.go`: `name`, `type`, optional `proxy_type`, `example`, optional `consumer_example`, and optional `metadata_example`.
- Keep every example minimal but valid against the matching schema in the same version.
- Only include `consumer_example` when that plugin has a `consumer_schema`.
- Only include `metadata_example` when that plugin has a `metadata_schema`.
- Keep `openfunction` examples shaped like this:

```json
{
  "function_uri": "https://example.com/function",
  "authorization": {
    "service_token": "token"
  }
}
```

### Remove From `plugin.json`

- Remove plugin catalog entries we do not expose from this repository's plugin list. On this branch, the removed plugins were `server-info` in `3.2` and `3.3`, `log-rotate`, `node-status`, and `server-info` in `3.11`, and `log-rotate`, `node-status`, and `mcp-bridge` in `3.13`.
- Remove examples that fail the schema in the same version.
- Remove copied upstream catalog noise that has no matching schema or no consumer in this repository.

### Do Not In `plugin.json`

- Do not assume every upstream plugin entry belongs in our catalog.
- Do not add schema-definition details to `plugin.json`; it is a catalog file, not the source of truth for validation.
- Do not add example scopes that `GetPluginSchema()` does not know how to resolve.

## Rules For `schema.json`

### Preserve And Fix

- Any schema node that behaves like an object and has `properties` must explicitly declare `"type": "object"`.
- Keep `openfunction.authorization` as an object with `service_token` nested under `properties`. Do not place `service_token` beside `type`.
- Keep `proxy-rewrite` as a valid object schema with correctly typed `headers`, `regex_uri`, `host`, `uri`, `_meta`, and `use_real_request_uri_unsafe` fields.
- Preserve the core resource fields restored by this branch: `plugins`, `upstream`, `service_id`, `remote_addrs`, and `consumer`.
- Keep id-like fields that accept either strings or integers as `anyOf` unions when upstream expects both.
- Keep `remote_addrs` as a non-empty array of IP or CIDR strings.

### Remove From `schema.json`

- Remove non-standard `encrypt_fields` entries from plugin schemas.
- Remove malformed object fragments that have `properties` but no object type.
- Remove accidental schema rewrites that break the expected gjson paths under `main`, `plugins`, or `stream_plugins`.

### Do Not In `schema.json`

- Do not rely on implicit object typing.
- Do not drop shared fields just because upstream diffs are noisy.
- Do not move plugin schemas to new paths without updating `schema.go` and the tests.

## Validation Checklist

Run these checks after every schema refresh or new version addition:

- `go test ./pkg/utils/schema`
- `go test ./pkg/utils/schema -run TestPluginExamplesMatchSchema`
- `go test ./pkg/utils/schema -run TestOpenFunctionAuthorizationSchemaShape`

If any example fails validation, fix the JSON files first. Do not weaken the tests to accept an invalid schema or example.
