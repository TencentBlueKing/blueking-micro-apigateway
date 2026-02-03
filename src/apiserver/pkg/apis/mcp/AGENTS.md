# AGENTS.md

## Scope

This package implements the MCP (Model Context Protocol) server that exposes APISIX
gateway management as MCP tools, resources, and prompts.

## Entry Points

- `router.go`: registers MCP routes under `/api/v1/mcp/gateways/:gateway_id`.
- `server.go`: constructs the MCP server instance.
- `register.go`: registers tools, resources, and prompts.

## Routing and Transport

- StreamableHTTP is the primary transport; SSE is kept for backward compatibility.
- The MCP server is mounted at `/api/v1/mcp/gateways/:gateway_id/` with an `/sse` fallback.
- All MCP requests use the `MCPAuth` middleware (bearer token + gateway ID validation).

## Authentication and Access Control

- MCP access tokens are issued and managed via Web APIs in `pkg/apis/web/handler/mcp_access_token.go`.
- Tokens are stored as SHA-256 hashes and are shown once on creation.
- Gateway support is restricted to APISIX `3.13.X` (`biz.MCPSupportedAPISIXVersions`).
- Tool handlers enforce write access via `tools.CheckWriteScope()` in `tools/common.go`.

## MCP Tools

Tools are grouped and registered in `register.go`:

- **Resource CRUD**: `tools/resource_crud.go` (list/get/create/update/delete/revert).
- **Sync**: `tools/sync.go` (sync from etcd, list synced, import to edit area).
- **Diff**: `tools/diff.go` (summary + detailed diffs).
- **Publish**: `tools/publish.go` (preview only; publish tools are intentionally disabled).
- **Schema**: `tools/schema.go` (resource schema, plugin schema, config validation, list plugins).

## MCP Resources

Documentation resources live in `resources/docs.go` and are exposed under `bk-apisix://docs/*`
URIs (resource types, relations, state machine, three-area model, publish workflow, errors,
plugin precedence).

## MCP Prompts

Workflow prompts live in `prompts/workflows.go` and provide usage guidance such as
`standard_workflow`, `troubleshoot_publish_error`, and `resource_dependency_check`.

## Key Dependencies

- Biz layer for resource operations, diffing, publish preview, and schema helpers.
- Middleware `pkg/middleware/mcp_auth.go` for auth and context setup.
- Schema utilities in `pkg/utils/schema`.

## Tests

- MCP tool helpers: `tools/common_test.go`
- MCP auth middleware: `pkg/middleware/mcp_auth_test.go`
