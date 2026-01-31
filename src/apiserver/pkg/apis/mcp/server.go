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

// Package mcp provides the MCP (Model Context Protocol) server implementation
package mcp

import (
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/version"
)

// ServerName is the name of the MCP server
const ServerName = "bk-micro-apigateway"

// NewMCPServer creates a new MCP server with all tools, resources, and prompts registered
func NewMCPServer(logger *slog.Logger) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    ServerName,
			Title:   "BlueKing Micro APIGateway",
			Version: version.Version,
		},
		&mcp.ServerOptions{
			Logger: logger,
		},
	)

	// Register tools
	registerTools(server)

	// Register resources
	registerResources(server)

	// Register prompts
	registerPrompts(server)

	return server
}
