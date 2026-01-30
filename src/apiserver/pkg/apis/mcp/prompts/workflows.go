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
		Description: "Complete workflow for syncing, editing, and publishing APISIX resources. " +
			"Follow this workflow for safe and organized configuration management.",
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
		Name:        "troubleshoot_publish_error",
		Description: "Guide for diagnosing and fixing publish failures. Use this when a publish operation fails.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "error_message",
				Description: "The error message from the failed publish",
				Required:    false,
			},
		},
	}, troubleshootPublishErrorHandler)

	// Resource Dependency Check
	server.AddPrompt(&mcp.Prompt{
		Name: "resource_dependency_check",
		Description: "Check resource dependencies before performing operations. " +
			"Use this before deleting or modifying resources.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "resource_type",
				Description: "The resource type to check",
				Required:    true,
			},
			{
				Name:        "resource_id",
				Description: "The resource ID to check",
				Required:    true,
			},
		},
	}, resourceDependencyCheckHandler)
}

func standardWorkflowHandler(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	content := `# Standard APISIX Configuration Workflow

Follow this workflow for safe and organized configuration management.

Note: The gateway is determined by the MCP endpoint URL you connected to.

---

## Phase 1: Synchronization

First, sync the latest state from etcd to ensure you're working with current data.

**Action:**
` + "```" + `
sync_from_etcd()
` + "```" + `

**Verify:**
- Check sync completed successfully
- Note the resource counts returned

---

## Phase 2: Import (If Managing New Resources)

If you need to manage resources that exist in etcd but aren't in the edit area:

**List unmanaged resources:**
` + "```" + `
list_synced_resource(resource_type="route", status="unmanaged")
` + "```" + `

**Import selected resources:**
` + "```" + `
add_synced_resources_to_edit_area(resource_ids=["id1", "id2"])
` + "```" + `

**Note:** Dependencies are automatically imported.

---

## Phase 3: Edit Resources

Make your configuration changes:

**Create new resource:**
When creating related resources (upstream -> service -> route), capture and use the returned IDs:
` + "```" + `
# Step 1: Create upstream first
result = create_resource(resource_type="upstream", name="my-upstream", config={...})
# Result contains: {"resource_id": "upstream-xxx", ...}

# Step 2: Create service with upstream_id from step 1
result = create_resource(resource_type="service", name="my-service", config={"upstream_id": "upstream-xxx", ...})
# Result contains: {"resource_id": "service-yyy", ...}

# Step 3: Create route with service_id from step 2
create_resource(resource_type="route", name="my-route", config={"service_id": "service-yyy", ...})
` + "```" + `

**Update existing resource:**
` + "```" + `
update_resource(resource_type="route", resource_id="route-1", config={...})
` + "```" + `

**Delete resource:**
` + "```" + `
delete_resource(resource_type="route", resource_ids=["old-route"])
` + "```" + `

**Tip:**
- Use validate_resource_config to check configs before creating/updating.
- Before update, should get the resource first, update the fields in ` + "`config`" + ` and put the whole ` + "`config`" + ` into the request!

---

## Phase 4: Review Changes

Before publishing, review all pending changes:

**Get change summary:**
` + "```" + `
diff_resources()
` + "```" + `

**Get detailed diff for specific resource:**
` + "```" + `
diff_detail(resource_type="route", resource_id="route-1")
` + "```" + `

**Preview pending changes:**
` + "```" + `
publish_preview()
` + "```" + `

---

## Phase 5: Publish

Note: Publishing via MCP is currently disabled for safety. Please use the web UI to publish changes.

---

## Best Practices

1. Always sync before making changes
2. Review diffs before publishing
3. When creating related resources, follow this order: Upstreams -> Services -> Routes
4. Capture returned resource IDs and use them in subsequent create calls
5. Test in staging environment first if possible
6. Changes take effect immediately after publish
7. Currently the publish operation can only be done on the website, not from the MCP Server. This is for safety.
`

	return &mcp.GetPromptResult{
		Description: "Standard workflow for APISIX configuration management",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}

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

	content := `# Troubleshoot Publish Error
` + errorSection + `
---

## Common Error Categories

### 1. Schema Validation Errors

**Symptoms:**
- "Additional property X is not allowed"
- "Required property Y is missing"
- "Type mismatch"

**Solutions:**
1. Check the resource schema:
` + "```" + `
get_resource_schema(apisix_version="3.13.X", resource_type="route")
` + "```" + `

2. Validate your config:
` + "```" + `
validate_resource_config(apisix_version="3.13.X", resource_type="route", config={...})
` + "```" + `

3. Remove unsupported fields for the APISIX version

---

### 2. Dependency Errors

**Symptoms:**
- "Referenced service not found"
- "Upstream does not exist"
- "Plugin config not found"

**Solutions:**
1. Check if referenced resources exist in etcd
2. When creating related resources, follow this order and capture IDs:
   - Create upstream first, note the returned resource_id
   - Create service with the upstream_id from step 1
   - Create route with the service_id from step 2
3. Publish dependencies before dependents (use web UI)

---

### 3. Etcd Connection Errors

**Symptoms:**
- "Connection refused"
- "Timeout"
- "Authentication failed"

**Solutions:**
1. Check etcd connectivity (contact infrastructure team)
2. Verify gateway etcd configuration
3. Retry the publish operation

---

### 4. Conflict Errors

**Symptoms:**
- "Resource already exists"
- "Duplicate key"

**Solutions:**
1. Sync to get latest state:
` + "```" + `
sync_from_etcd()
` + "```" + `

2. Check for conflicts in diff:
` + "```" + `
diff_resources()
` + "```" + `

3. Resolve conflicts and retry

---

## Diagnostic Steps

1. **Get current state:**
` + "```" + `
sync_from_etcd()
` + "```" + `

2. **Check pending changes:**
` + "```" + `
diff_resources()
` + "```" + `

3. **Review specific resource:**
` + "```" + `
get_resource(resource_type="route", resource_id="...")
` + "```" + `

4. **Validate configuration:**
` + "```" + `
validate_resource_config(apisix_version="3.13.X", resource_type="route", config={...})
` + "```" + `

---

## Recovery Options

### Option 1: Fix and Retry
1. Identify the issue from error message
2. Update the resource to fix the issue
3. Retry publish (via web UI)

### Option 2: Revert and Retry
1. Revert the problematic resource:
` + "```" + `
revert_resource(resource_type="route", resource_ids=["..."])
` + "```" + `
2. Re-apply changes correctly
3. Retry publish (via web UI)

### Option 3: Skip Problematic Resource
1. Publish other resources first
2. Debug the problematic resource separately
`

	return &mcp.GetPromptResult{
		Description: "Troubleshoot publish errors",
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

	content := `# Resource Dependency Check

## Resource Details
- **Resource Type:** ` + resourceType + `
- **Resource ID:** ` + resourceID + `

---

## Step 1: Get Resource Details

First, retrieve the resource configuration:

` + "```" + `
get_resource(resource_type="` + resourceType + `", resource_id="` + resourceID + `")
` + "```" + `

---

## Step 2: Check Dependencies (What This Resource Depends On)

Based on resource type, check these references:

### For Routes:
- ` + "`service_id`" + `: Does the referenced service exist?
- ` + "`upstream_id`" + `: Does the referenced upstream exist?
- ` + "`plugin_config_id`" + `: Does the referenced plugin config exist?

### For Services:
- ` + "`upstream_id`" + `: Does the referenced upstream exist?

### For Consumers:
- ` + "`group_id`" + `: Does the referenced consumer group exist?

---

## Step 3: Check Dependents (What Depends On This Resource)

Find resources that reference this resource:

### If Deleting an Upstream:
` + "```" + `
# Check if any routes reference this upstream
list_resource(resource_type="route")
# Look for upstream_id="` + resourceID + `" in results

# Check if any services reference this upstream
list_resource(resource_type="service")
# Look for upstream_id="` + resourceID + `" in results
` + "```" + `

### If Deleting a Service:
` + "```" + `
# Check if any routes reference this service
list_resource(resource_type="route")
# Look for service_id="` + resourceID + `" in results
` + "```" + `

### If Deleting a Consumer Group:
` + "```" + `
# Check if any consumers reference this group
list_resource(resource_type="consumer")
# Look for group_id="` + resourceID + `" in results
` + "```" + `

---

## Step 4: Impact Analysis

### Before Deleting:
1. List all dependents found above
2. Decide how to handle each dependent:
   - Update dependent to use different reference
   - Delete dependent as well
   - Keep dependent (will cause errors)

### Before Updating:
1. Check if the update breaks dependent resources
2. Especially if changing:
   - ID (rare, usually not allowed)
   - Structure/format that dependents rely on

---

## Resource Relationship Creation Pattern

When creating related resources, follow this order and capture IDs:

` + "```" + `
# Step 1: Create upstream first
result = create_resource(resource_type="upstream", name="my-upstream", config={...})
# Result contains: {"resource_id": "upstream-xxx", ...}

# Step 2: Create service with upstream_id from step 1
result = create_resource(resource_type="service", name="my-service", config={"upstream_id": "upstream-xxx", ...})
# Result contains: {"resource_id": "service-yyy", ...}

# Step 3: Create route with service_id from step 2
create_resource(resource_type="route", name="my-route", config={"service_id": "service-yyy", ...})
` + "```" + `

---

## Best Practices

1. **Create order**: Create dependencies before dependents
   - Upstreams before Services
   - Services before Routes
   - Capture and use returned resource IDs

2. **Delete order**: Delete dependents before dependencies
   - Routes before Services
   - Services before Upstreams
   - Consumers before Consumer Groups

3. **Update order**: Update dependencies before dependents
   - Upstreams before Services
   - Services before Routes

4. **Publish order**: Publish dependencies first (via web UI)
   - Upstreams -> Services -> Routes
`

	return &mcp.GetPromptResult{
		Description: "Check resource dependencies before operations",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}
