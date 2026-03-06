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

// Package prompts provides MCP prompts for workflow guidance
package prompts

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterWorkflowPrompts registers all workflow prompts
// Note: gateway_id is now in the URL path, so prompts don't need it as argument
func RegisterWorkflowPrompts(server *mcp.Server) {
	// Standard Workflow
	server.AddPrompt(&mcp.Prompt{
		Name: "standard_workflow",
		Description: "Default safe workflow for APISIX change management: " +
			"sync, import, edit, diff, preview, then publish via web UI/API.",
	}, standardWorkflowHandler)

	// NOTE: publish_checklist is commented out for safety.
	// Publishing directly via MCP is not currently enabled.
	// // Publish Checklist
	// server.AddPrompt(&mcp.Prompt{
	// 	Name: "publish_checklist",
	// 	Description: "Pre-publish verification checklist to ensure safe deployment. " +
	// 		"Use this before publishing changes to production.",
	// }, publishChecklistHandler)

	// Troubleshoot Publish Error
	server.AddPrompt(&mcp.Prompt{
		Name: "troubleshoot_publish_error",
		Description: "Step-by-step diagnosis workflow for publish failures, " +
			"including schema, dependency, and environment checks.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "error_message",
				Description: "Optional raw error message from failed publish operation",
				Required:    false,
			},
		},
	}, troubleshootPublishErrorHandler)

	// Resource Dependency Check
	server.AddPrompt(&mcp.Prompt{
		Name:        "resource_dependency_check",
		Description: "Dependency safety checklist for update/delete operations on a specific resource.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "resource_type",
				Description: "Required APISIX resource type for the target resource",
				Required:    true,
			},
			{
				Name:        "resource_id",
				Description: "Required resource ID for the target resource",
				Required:    true,
			},
		},
	}, resourceDependencyCheckHandler)
}

func standardWorkflowHandler(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	content := `# Standard APISIX Change Workflow (MCP-Safe)

Use this workflow by default when operating through MCP.

## Hard Constraints

1. Gateway context comes from the MCP endpoint URL
   (` + "`/mcp/gateways/:gateway_id/`" + `). Do not pass ` + "`gateway_id`" + ` to tools.
2. MCP publish tools are disabled. Use MCP for preparation/review, then publish in Web UI/OpenAPI.
3. ` + "`update_resource`" + ` requires a full replacement ` + "`config`" + ` (no partial patch).
4. Always provide both ` + "`resource_type`" + ` and ` + "`resource_id`" + ` for single-resource tools.

---

## Step 1: Sync Latest Runtime State

` + "```" + `
sync_from_etcd()
` + "```" + `

Confirm the call succeeded and record returned counts.

---

## Step 2: Import Unmanaged Runtime Resources (Optional)

Use this when resources already exist in APISIX but are not managed in edit area.

` + "```" + `
list_synced_resource(resource_type="route", status="unmanaged")
add_synced_resources_to_edit_area(resource_ids=["id1", "id2"])
` + "```" + `

Dependencies are imported automatically.

---

## Step 3: Edit in Draft Area

### Create (dependency order)

` + "```" + `
# 1) Upstream
u = create_resource(resource_type="upstream", name="my-upstream", config={...})

# 2) Service references upstream
s = create_resource(resource_type="service", name="my-service", config={"upstream_id": u.resource_id, ...})

# 3) Route references service
create_resource(resource_type="route", name="my-route", config={"service_id": s.resource_id, ...})
` + "```" + `

### Update (full config replacement)

` + "```" + `
origin = get_resource(resource_type="route", resource_id="route-1")
# Edit origin.config locally, then send full config back:
update_resource(resource_type="route", resource_id="route-1", config={...full config...})
` + "```" + `

### Delete

` + "```" + `
delete_resource(resource_type="route", resource_ids=["old-route"])
` + "```" + `

` + "`delete_resource`" + ` blocks unsafe deletes when dependencies exist.

---

## Step 4: Validate and Review

` + "```" + `
validate_resource_config(apisix_version="3.13", resource_type="route", config={...})
diff_resources()
diff_detail(resource_type="route", resource_id="route-1")
publish_preview()
` + "```" + `

Review summary counts and detailed diffs before publish.

---

## Step 5: Publish Outside MCP

Publish from Web UI/OpenAPI only. Then run ` + "`sync_from_etcd()`" + ` again to refresh MCP sync snapshot.

---

## Quick Safety Rules

1. Sync before edit.
2. Validate before create/update.
3. Diff and preview before publish.
4. Create dependencies before dependents (Upstream -> Service -> Route).
5. Delete in reverse order (Route -> Service -> Upstream).
6. Keep returned ` + "`resource_id`" + ` values and reuse them explicitly.
`

	return &mcp.GetPromptResult{
		Description: "Default safe workflow for APISIX changes through MCP",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}

//nolint:unused // Kept for future use when MCP publishing is enabled
func publishChecklistHandler(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	content := `# Pre-Publish Verification Checklist

Complete this checklist before publishing changes to production.

---

## ✅ Data Synchronization

- [ ] **Sync Executed**: Run sync_from_etcd() to get latest state
- [ ] **Sync Recent**: Sync completed within the last 5 minutes
- [ ] **Sync Successful**: No errors during sync

**Verify with:**
` + "```" + `
sync_from_etcd()
` + "```" + `

---

## ✅ Change Review

- [ ] **Diff Reviewed**: Examined diff_resources output
- [ ] **Create Count Confirmed**: Verified number of new resources
- [ ] **Update Count Confirmed**: Verified number of modified resources
- [ ] **Delete Count Confirmed**: Verified number of resources to be removed

**Verify with:**
` + "```" + `
diff_resources()
publish_preview()
` + "```" + `

---

## ✅ Dependency Verification

- [ ] **Services Exist**: All referenced service_ids exist or are being published
- [ ] **Upstreams Exist**: All referenced upstream_ids exist or are being published
- [ ] **Plugin Configs Exist**: All referenced plugin_config_ids exist or are being published
- [ ] **Delete Impact Checked**: Deleted resources won't break other resources

---

## ✅ Configuration Validation

- [ ] **Schema Valid**: Configs match target APISIX version schema
- [ ] **Plugin Configs Valid**: All plugin configurations are correct
- [ ] **No Additional Properties**: No unsupported fields in configs

**Verify with:**
` + "```" + `
validate_resource_config(apisix_version="3.13.X", resource_type="route", config={...})
` + "```" + `

---

## ⚠️ Risk Awareness

- [ ] **Production Impact**: Understand changes take effect immediately
- [ ] **Rollback Plan**: Know how to revert if issues occur (use revert_resource)
- [ ] **Monitoring Ready**: Have monitoring/alerts in place

---

## 🚀 Ready to Publish?

If all checks pass, publish changes using the web UI.

---

## 🔙 Rollback Instructions

If issues occur after publish:

1. **Sync latest state:**
` + "```" + `
sync_from_etcd()
` + "```" + `

2. **Revert problematic resources:**
` + "```" + `
revert_resource(resource_type="route", resource_ids=["..."])
` + "```" + `

3. **Publish the reverted state using the web UI**
`

	return &mcp.GetPromptResult{
		Description: "Pre-publish verification checklist",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}

func troubleshootPublishErrorHandler(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	errorMessage := ""
	if req.Params.Arguments != nil {
		if err, ok := req.Params.Arguments["error_message"]; ok {
			errorMessage = err
		}
	}

	errorSection := ""
	if errorMessage != "" {
		errorSection = "\n## Error: " + errorMessage + "\n"
	}

	content := `# Troubleshoot Publish Failure
` + errorSection + `
Use this checklist when publish fails in Web UI/OpenAPI.

## Step 1: Reproduce and Capture Context

Record:
1. Full error message
2. Resource type + resource ID (if available)
3. APISIX gateway version
4. Approximate failure time

---

## Step 2: Refresh Local Snapshot

` + "```" + `
sync_from_etcd()
diff_resources()
publish_preview()
` + "```" + `

This confirms your draft state and current runtime snapshot.

---

## Step 3: Classify Error Quickly

### A) Schema Validation
Typical error text:
- "additional property ... is not allowed"
- "required property ... is missing"
- "type mismatch"

Checks:
` + "```" + `
get_resource(resource_type="route", resource_id="...")
get_resource_schema(apisix_version="3.13", resource_type="route")
validate_resource_config(apisix_version="3.13", resource_type="route", config={...})
` + "```" + `

### B) Dependency Missing
Typical error text:
- "service/upstream/plugin_config not found"

Checks:
1. Verify referenced IDs exist.
2. Ensure dependency creation order was followed: Upstream -> Service -> Route.
3. Publish dependencies first (outside MCP).

### C) Data Conflict
Typical error text:
- "duplicate key"
- "resource already exists"

Checks:
1. Re-sync (` + "`sync_from_etcd`" + `).
2. Re-open resource with ` + "`get_resource`" + `.
3. Re-apply change on top of latest state.

### D) Infrastructure/Connectivity
Typical error text:
- etcd timeout/connection/auth failure

Checks:
1. Retry once.
2. If persistent, escalate to platform/infrastructure owner.

---

## Step 4: Remediate Safely

Option 1: Fix config and retry publish.

Option 2: Revert problematic draft, then re-apply:
` + "```" + `
revert_resource(resource_type="route", resource_ids=["..."])
` + "```" + `

Option 3: Publish unaffected resources first, isolate failing resource for separate fix.

---

## Step 5: Post-Fix Validation

Before retrying publish:
1. ` + "`diff_resources()`" + ` shows expected changes only.
2. ` + "`publish_preview()`" + ` contains no unintended resources.
3. All changed resources pass ` + "`validate_resource_config`" + `.
`

	return &mcp.GetPromptResult{
		Description: "Publish failure diagnosis and recovery runbook",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}

func resourceDependencyCheckHandler(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	resourceType := ""
	resourceID := ""
	if req.Params.Arguments != nil {
		if rt, ok := req.Params.Arguments["resource_type"]; ok {
			resourceType = rt
		}
		if rid, ok := req.Params.Arguments["resource_id"]; ok {
			resourceID = rid
		}
	}

	content := `# Resource Dependency Safety Check

## Target
- **resource_type**: ` + resourceType + `
- **resource_id**: ` + resourceID + `

Use this checklist before ` + "`update_resource`" + ` or ` + "`delete_resource`" + `.

---

## Step 1: Load Current Resource

` + "```" + `
get_resource(resource_type="` + resourceType + `", resource_id="` + resourceID + `")
` + "```" + `

Capture key references from config:
` + "`service_id`" + `, ` + "`upstream_id`" + `, ` + "`plugin_config_id`" + `, ` + "`group_id`" + `.

---

## Step 2: Validate Outbound Dependencies

Check that referenced resources exist and are in expected state.

Common references:
1. Route -> service/upstream/plugin_config
2. Service -> upstream
3. Consumer -> consumer_group
4. Stream Route -> service/upstream

---

## Step 3: Validate Inbound Dependents

Identify what may break if this resource is changed or deleted.

Quick scans:
` + "```" + `
list_resource(resource_type="route")
list_resource(resource_type="service")
list_resource(resource_type="consumer")
list_resource(resource_type="stream_route")
` + "```" + `

Then filter locally for references to ` + "`resource_id`" + `.

---

## Step 4: Choose Safe Action

### If deleting
1. Remove or update dependents first.
2. Then call:
` + "```" + `
delete_resource(resource_type="` + resourceType + `", resource_ids=["` + resourceID + `"])
` + "```" + `
3. If deletion is blocked, follow returned dependency error details.

### If updating
1. Update dependents if contract changes.
2. Send full replacement config (not partial patch).

---

## Step 5: Re-check Publish Impact

` + "```" + `
diff_resources(resource_type="` + resourceType + `")
publish_preview(resource_type="` + resourceType + `")
` + "```" + `

Proceed to publish in Web UI/OpenAPI after verification.

---

## Recommended Ordering

1. Create: Upstream -> Service -> Route
2. Delete: Route -> Service -> Upstream
3. Publish: dependencies before dependents
`

	return &mcp.GetPromptResult{
		Description: "Dependency safety checklist for update/delete operations",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}
