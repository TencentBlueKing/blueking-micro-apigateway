/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
 * Copyright (C) 2025 Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *     http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

// Package resources provides MCP documentation resources
package resources

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterDocumentationResources registers all documentation resources
func RegisterDocumentationResources(server *mcp.Server) {
	// Resource Types
	server.AddResource(&mcp.Resource{
		URI:         "bk-apisix://docs/resource_types",
		Name:        "APISIX Resource Types",
		Description: "Documentation for all supported APISIX resource types",
		MIMEType:    "text/markdown",
	}, resourceTypesHandler)

	// Resource Relations
	server.AddResource(&mcp.Resource{
		URI:         "bk-apisix://docs/resource_relations",
		Name:        "Resource Relations",
		Description: "Documentation for resource dependencies and relationships",
		MIMEType:    "text/markdown",
	}, resourceRelationsHandler)

	// State Machine
	server.AddResource(&mcp.Resource{
		URI:         "bk-apisix://docs/state_machine",
		Name:        "Resource State Machine",
		Description: "Documentation for resource status transitions",
		MIMEType:    "text/markdown",
	}, stateMachineHandler)

	// Three Area Model
	server.AddResource(&mcp.Resource{
		URI:         "bk-apisix://docs/three_area_model",
		Name:        "Three Area Model",
		Description: "Documentation for the Edit/Sync/Publish area model",
		MIMEType:    "text/markdown",
	}, threeAreaModelHandler)

	// Publish Workflow
	server.AddResource(&mcp.Resource{
		URI:         "bk-apisix://docs/publish_workflow",
		Name:        "Publish Workflow",
		Description: "Documentation for the publishing workflow",
		MIMEType:    "text/markdown",
	}, publishWorkflowHandler)

	// API Error Details
	server.AddResource(&mcp.Resource{
		URI:         "bk-apisix://docs/api_error_details",
		Name:        "API Error Details",
		Description: "Documentation for common API errors and solutions",
		MIMEType:    "text/markdown",
	}, apiErrorDetailsHandler)

	// Plugin Precedence
	server.AddResource(&mcp.Resource{
		URI:         "bk-apisix://docs/plugin_precedence",
		Name:        "Plugin Merging Precedence",
		Description: "Documentation for how plugins are merged when configured on multiple objects",
		MIMEType:    "text/markdown",
	}, pluginPrecedenceHandler)
}

func resourceTypesHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/markdown", Text: resourceTypesDoc},
		},
	}, nil
}

func resourceRelationsHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/markdown", Text: resourceRelationsDoc},
		},
	}, nil
}

func stateMachineHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/markdown", Text: stateMachineDoc},
		},
	}, nil
}

func threeAreaModelHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/markdown", Text: threeAreaModelDoc},
		},
	}, nil
}

func publishWorkflowHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/markdown", Text: publishWorkflowDoc},
		},
	}, nil
}

func apiErrorDetailsHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/markdown", Text: apiErrorDetailsDoc},
		},
	}, nil
}

func pluginPrecedenceHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: req.Params.URI, MIMEType: "text/markdown", Text: pluginPrecedenceDoc},
		},
	}, nil
}

// Documentation content
const resourceTypesDoc = `# APISIX Resource Types

BK Micro APIGateway manages the following APISIX resource types:

## Core Concepts

### Plugin
APISIX Plugins extend APISIX's functionalities to meet organization or user-specific
requirements in traffic management, observability, security, request/response
transformation, serverless computing, and more.

A Plugin configuration can be bound directly to a Route, Service, Consumer, or Plugin
Config. When the same plugin is configured on multiple objects, the configurations are
merged with specific precedence rules. See the "Plugin Merging Precedence" documentation
for details.

## HTTP Resources

### Route
Routes match the client's request based on defined rules, load and execute the
corresponding plugins, and forward the request to the specified Upstream.
- **Key Fields**: uri, hosts, methods, plugins, service_id, upstream_id
- **Identifier**: id (auto-generated or user-defined)

### Service
A Service is an abstraction of an API (which can also be understood as a set of Route
abstractions). It usually corresponds to an upstream service abstraction. The relationship
between Routes and a Service is usually N:1; N routes bound to 1 Service.
- **Key Fields**: name, upstream_id, plugins, hosts
- **Identifier**: id

### Upstream
Upstream is a virtual host abstraction that performs load balancing on a given set of
service nodes according to the configured rules. Although Upstream can be directly
configured to the Route or Service, using an Upstream object is recommended when there
is duplication.
- **Key Fields**: nodes, type (roundrobin/chash), timeout, health checks
- **Identifier**: id

### Consumer
Consumers represent API clients/users for authentication and rate limiting.
- **Key Fields**: username, plugins (auth plugins like key-auth, jwt-auth)
- **Identifier**: username (not id)

### Consumer Group
Consumer Groups are used to extract commonly used Plugin configurations and can be bound
directly to a Consumer. With consumer groups, you can define any number of plugins, e.g.
rate limiting and apply them to a set of consumers, instead of managing each consumer
individually.
- **Key Fields**: plugins
- **Identifier**: id

### Plugin Config
Plugin Configs are used to extract commonly used Plugin configurations and can be bound directly to a Route.
- **Key Fields**: plugins
- **Identifier**: id

### Global Rule
If we want a Plugin to work on all requests, this is where we register a global Plugin with Global Rule.
- **Key Fields**: plugins
- **Identifier**: id

### Plugin Metadata
A plugin metadata object is used to configure the common metadata field(s) of all plugin
instances sharing the same plugin name.
- **Key Fields**: plugin-specific metadata
- **Identifier**: plugin name

### SSL
TLS/SSL certificates for HTTPS.
- **Key Fields**: cert, key, snis
- **Identifier**: id

### Proto
Protocol buffer definitions for gRPC transcoding.
- **Key Fields**: content
- **Identifier**: id

## Stream Resources

### Stream Route
Routes for TCP/UDP stream proxying.
- **Key Fields**: server_addr, server_port, upstream_id
- **Identifier**: id
`

const resourceRelationsDoc = `# Resource Relations and Dependencies

## Dependency Graph

` + "```" + `
Route ─────┬──► Service ─────► Upstream
           │
           ├──► Upstream (direct)
           │
           └──► Plugin Config

Consumer ──────► Consumer Group

Stream Route ──► Upstream
               ──► Service
` + "```" + `

## Dependency Rules

### Routes
- Routes can reference: Service, Upstream, Plugin Config
- At least one of service_id or upstream_id should be set (or inline upstream)
- plugin_config_id is optional

### Services
- Services can reference: Upstream
- upstream_id is optional (can use inline upstream)

### Consumers
- Consumers can reference: Consumer Group
- group_id is optional

### Stream Routes
- Stream routes can reference: Service, Upstream
- Similar to HTTP routes

## Import Order

When importing resources from etcd to the edit area, dependencies are automatically resolved:

1. Upstreams (no dependencies)
2. Services (depends on Upstreams)
3. Consumer Groups (no dependencies)
4. Consumers (depends on Consumer Groups)
5. Plugin Configs (no dependencies)
6. Routes (depends on Services, Upstreams, Plugin Configs)
7. Stream Routes (depends on Services, Upstreams)
8. Global Rules (no dependencies)
9. SSL (no dependencies)
10. Proto (no dependencies)
11. Plugin Metadata (no dependencies)

## Publish Order

When publishing, resources are published in dependency order to ensure referenced resources exist before dependents.

## Delete Considerations

Before deleting a resource, check if it's referenced by other resources:
- Deleting an Upstream that's referenced by a Route will cause the Route to fail
- Use diff_resources to check for potential issues
`

const stateMachineDoc = `# Resource State Machine

Resources in BK Micro APIGateway follow a state machine pattern for managing their lifecycle.

## States

| State | Description |
|-------|-------------|
| ` + "`create_draft`" + ` | Newly created, not yet published to APISIX |
| ` + "`update_draft`" + ` | Modified from published version, changes pending |
| ` + "`delete_draft`" + ` | Marked for deletion, waiting for publish to remove |
| ` + "`success`" + ` | Published and in sync with APISIX |

## State Transitions

` + "```" + `
                    Create
                      │
                      ▼
              ┌───────────────┐
              │  create_draft │◄───────┐
              └───────┬───────┘        │
                      │                │ Update
         Publish      │                │
                      ▼                │
              ┌───────────────┐────────┘
      ┌──────►│    success    │◄──────┐
      │       └───────┬───────┘       │
      │               │               │
      │    Update     │     Revert    │
      │               ▼               │
      │       ┌───────────────┐       │
      │       │  update_draft │───────┤
      │       └───────────────┘       │
      │               │               │
      │    Publish    │               │
      │               ▼               │
      │       ┌───────────────┐       │
      │       │  delete_draft │───────┘
      │       └───────┬───────┘
      │               │
      │    Publish    │
      │               ▼
      │         (deleted)
      │
      └─────────────────────────────────
` + "```" + `

## Operations and Transitions

| Current State | Operation | Next State | Notes |
|---------------|-----------|------------|-------|
| (none) | Create | create_draft | New resource |
| create_draft | Publish | success | Written to etcd |
| create_draft | Delete | (deleted) | Hard delete |
| create_draft | Update | create_draft | Stays in draft |
| success | Update | update_draft | Changes staged |
| success | Delete | delete_draft | Deletion staged |
| update_draft | Publish | success | Changes applied |
| update_draft | Revert | success | Discard changes |
| update_draft | Update | update_draft | More changes |
| delete_draft | Publish | (deleted) | Removed from etcd |
| delete_draft | Revert | success | Cancel deletion |
`

const threeAreaModelDoc = `# Three Area Model

BK Micro APIGateway uses a three-area model for managing APISIX configurations:

## Areas

### 1. Etcd/Data Plane (APISIX)
- The actual running configuration in APISIX
- Source of truth for what's deployed
- Changes here take effect immediately

### 2. Sync Area (gateway_sync_data table)
- Read-only mirror of etcd state
- Updated automatically by the scheduler or manually via sync_from_etcd
- Used for:
  - Comparing with edit area (diff)
  - Reverting changes
  - Importing new resources

### 3. Edit Area (resource tables)
- User-editable staging area
- Each resource type has its own table (route, service, upstream, etc.)
- Changes don't affect APISIX until published
- Resources have draft states

## Data Flow

` + "```" + `
┌─────────────────────────────────────────────────────────────┐
│                    APISIX/Etcd (Data Plane)                 │
│                                                             │ <─────┐
│  Current running configuration - changes take effect        │       │
│  immediately                                                │       │
└─────────────────────────┬───────────────────────────────────┘       │
                          │                                           │
                          │                                           │
                  Sync    │                                           │
                          ▼                                           │
┌─────────────────────────────────────────────────────────────┐       │
│                    Sync Area (gateway_sync_data)            │       │
│                                                             │       │
│  Read-only snapshot of etcd - updated periodically or       │       │
│  manually via sync_from_etcd                                │       │
└─────────────────────────┬───────────────────────────────────┘       │
                          │                                           │
                          │ ▲                                         │
                 Import   │ │  Revert                                 │
                          ▼ │                                         │
┌─────────────────────────────────────────────────────────────┐       │
│                    Edit Area (resource tables)              │       │
│                                                             │       │
│  User-editable drafts - changes are staged until published  │───────┘ Publish
│  Resources have status: create_draft, update_draft,         │
│  delete_draft, success                                      │
└─────────────────────────────────────────────────────────────┘
` + "```" + `

## Workflow

1. **Sync**: Fetch current state from etcd to sync area
2. **Import**: Copy unmanaged resources from sync area to edit area
3. **Edit**: Make changes in the edit area (CRUD operations)
4. **Diff**: Compare edit area with sync area to see pending changes
5. **Publish**: Apply changes from edit area to etcd directly

## Benefits

- **Safety**: Changes are staged before applying to production
- **Audit**: Track what changed and who changed it
- **Rollback**: Revert to synced state if needed
- **Collaboration**: Multiple users can prepare changes before publishing
`

const publishWorkflowDoc = `# Publish Workflow

## Standard Workflow

### Step 1: Sync from Etcd
` + "```" + `
sync_from_etcd()
` + "```" + `
- Fetches current state from APISIX/etcd
- Updates the sync area with latest data
- Check sync result to confirm resource counts

### Step 2: Import New Resources (if needed)
` + "```" + `
# List unmanaged resources
list_synced_resource(resource_type="route", status="unmanaged")

# Import to edit area
add_synced_resources_to_edit_area(resource_ids=["route-1", "route-2"])
` + "```" + `
- Check for resources in etcd that aren't being managed
- Import resources you want to manage
- Dependencies are automatically imported

### Step 3: Edit Resources
` + "```" + `
# Create new resource
create_resource(resource_type="route", name="my-route", config={...})

# Update existing resource
update_resource(resource_type="route", resource_id="route-1", config={...})

# Delete resource
delete_resource(resource_type="route", resource_ids=["old-route"])
` + "```" + `
- Resources enter draft status (create_draft, update_draft, delete_draft)
- Changes are NOT applied to APISIX yet

### Step 4: Review Changes
` + "```" + `
# Get change summary
diff_resources()

# Get detailed diff for specific resource
diff_detail(resource_type="route", resource_id="route-1")
` + "```" + `
- Review all pending changes
- Check for potential issues

### Step 5: Publish
` + "```" + `
# Preview what will be published
publish_preview()

` + "```" + `
- Publishing via MCP is disabled for safety. Please use the web UI to publish changes.

## Pre-Publish Checklist

✅ **Data Sync**
- [ ] Executed latest sync (sync_from_etcd)
- [ ] Sync completed within last 5 minutes

✅ **Change Review**
- [ ] Reviewed diff_resources summary
- [ ] Confirmed create count
- [ ] Confirmed update count
- [ ] Confirmed delete count

✅ **Dependency Check**
- [ ] Referenced Services/Upstreams exist or being published together
- [ ] Deleted resources won't break other resources

✅ **Configuration Validation**
- [ ] Configs match target APISIX version schema
- [ ] Plugin configurations are correct

⚠️ **Risk Awareness**
- Changes take effect immediately in production
- Consider testing in a staging environment first
`

const apiErrorDetailsDoc = `# API Error Details

## Authentication Errors

### 401 Unauthorized
- **Missing Authorization header**: MCP requests require Bearer token authentication
- **Invalid token**: The provided token doesn't exist in the database
- **Token format error**: Use format: ` + "`Authorization: Bearer <token>`" + `

### 403 Forbidden
- **Token expired**: The MCP access token has passed its expiration date
- **Insufficient scope**: Token has ` + "`read`" + ` scope but operation requires ` + "`write`" + `
- **Gateway not supported**: Gateway's APISIX version is not 3.13.X (MCP only supports 3.13)

## Resource Errors

### 404 Not Found
- **Resource not found**: The specified resource_id doesn't exist
- **Gateway not found**: The specified gateway_id doesn't exist

### 409 Conflict
- **Name already exists**: Resource name is already used in the gateway
- **Duplicate resource**: Attempting to create a resource that already exists

### 400 Bad Request
- **Invalid resource type**: Use one of: route, service, upstream, consumer, consumer_group,
  plugin_config, global_rule, plugin_metadata, proto, ssl, stream_route
- **Invalid status**: Use one of: create_draft, update_draft, delete_draft, success
- **Missing required field**: Check the tool's required parameters
- **Invalid config**: Configuration doesn't match APISIX schema

## Schema Validation Errors

### Common Schema Errors
- **Additional property not allowed**: Config contains fields not in the APISIX schema
- **Required property missing**: A required field is not provided
- **Type mismatch**: Field value doesn't match expected type

### Version-Specific Errors
- **Name field not supported**: Some resources (consumer_group, stream_route) don't support
  ` + "`name`" + ` field in APISIX 3.11
- Check APISIXVersion-specific field support

## Publish Errors

### Etcd Connection Errors
- **Connection refused**: Cannot connect to etcd
- **Timeout**: Etcd operation timed out
- **Authentication failed**: Etcd credentials are incorrect

### Publish Validation Errors
- **Schema validation failed**: Config doesn't match the strict etcd schema
- **Dependency missing**: Referenced resource (service_id, upstream_id) doesn't exist in etcd

## Best Practices

1. **Always sync before publishing**: Ensure you have the latest state
2. **Use diff before publish**: Review changes before applying
3. **Validate configs**: Use validate_resource_config before create/update
4. **Check dependencies**: Ensure referenced resources exist
5. **Publish in order**: Publish dependencies before dependents
`

const pluginPrecedenceDoc = `# Plugin Merging Precedence

## Overview

When the same plugin is configured both globally in a global rule and locally in an
object (e.g. a route), both plugin instances are executed sequentially.

However, if the same plugin is configured locally on multiple objects, such as on Route,
Service, Consumer, Consumer Group, or Plugin Config, only one copy of configuration is
used as each non-global plugin is only executed once.

## Precedence Order

This is because during execution, plugins configured in these objects are merged with
respect to a specific order of precedence:

` + "```" + `
Consumer > Consumer Group > Route > Plugin Config > Service
` + "```" + `

If the same plugin has different configurations in different objects, the plugin
configuration with the highest order of precedence during merging will be used.

## Examples

### Example 1: Rate Limiting
If you configure the ` + "`limit-count`" + ` plugin on both a Route and a Service:
- The Route's ` + "`limit-count`" + ` configuration will be used (Route has higher precedence than Service)

### Example 2: Consumer Override
If you configure ` + "`key-auth`" + ` on both a Consumer and a Route:
- The Consumer's configuration will be used (Consumer has highest precedence)

### Example 3: Global vs Local
If you configure ` + "`prometheus`" + ` in a Global Rule AND on a Route:
- Both configurations will execute (global plugins don't merge with local)
- The global plugin executes first, then the route-level plugin

## Best Practices

1. **Use Service for defaults**: Configure plugins on Service as baseline defaults
2. **Use Route for overrides**: Override specific plugins at the Route level when needed
3. **Use Consumer for identity-based rules**: Configure auth and rate limits per consumer
4. **Use Plugin Config for reusability**: Share common plugin configurations across routes
5. **Use Global Rules carefully**: Only for plugins that should apply to all traffic
`
