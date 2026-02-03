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

package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/mcp/prompts"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/mcp/resources"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/mcp/tools"
)

// registerMiddleware registers MCP receiving middleware for common operations
func registerMiddleware(server *mcp.Server) {
	// GatewayContextMiddleware injects gateway info into context for all tool calls
	server.AddReceivingMiddleware(tools.GatewayContextMiddleware)
	// WriteAccessMiddleware enforces write scope for write tools
	server.AddReceivingMiddleware(tools.WriteAccessMiddleware)
}

// registerTools registers all MCP tools
func registerTools(server *mcp.Server) {
	// Resource CRUD tools
	tools.RegisterResourceCRUDTools(server)

	// Sync tools
	tools.RegisterSyncTools(server)

	// Diff tools
	tools.RegisterDiffTools(server)

	// Publish tools
	tools.RegisterPublishTools(server)

	// Schema tools
	tools.RegisterSchemaTools(server)
}

// registerResources registers all MCP resources
func registerResources(server *mcp.Server) {
	resources.RegisterDocumentationResources(server)
}

// registerPrompts registers all MCP prompts
func registerPrompts(server *mcp.Server) {
	prompts.RegisterWorkflowPrompts(server)
}
