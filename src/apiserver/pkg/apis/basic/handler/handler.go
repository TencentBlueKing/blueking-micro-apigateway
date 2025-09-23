/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关(BlueKing - Micro APIGateway) available.
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

/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - APIGateway) available.
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

// Package handler ...
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/apis/basic/serializer"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/version"
)

// Ping ...
//
//	@Summary	服务探活
//	@Tags		basic
//	@Produce	text/plain
//	@Success	200	{string}	string	pong
//	@Router		/ping [get]
func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

// Healthz ...
//
//	@Summary	提供服务健康状态
//	@Tags		basic
//	@Param		token	query		string	true	"healthz api token"
//	@Success	200		{object}	serializer.HealthResponse
//	@Router		/healthz [get]
//
// FIXME 细化健康检查逻辑（分探针）
func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, serializer.HealthResponse{Healthy: true})
}

// Version ...
//
//	@Summary	服务版本信息
//	@Tags		basic
//	@Success	200	{object}	serializer.VersionResponse
//	@Router		/version [get]
func Version(c *gin.Context) {
	respData := serializer.VersionResponse{
		Version:   version.Version,
		GitCommit: version.GitCommit,
		BuildTime: version.BuildTime,
		GoVersion: version.GoVersion,
	}
	c.JSON(http.StatusOK, respData)
}

// Metrics ...
//
//	@Summary	Prometheus 指标
//	@Tags		basic
//	@Produce	text/plain
//	@Param		token	query		string	true	"metrics api token"
//	@Success	200		{string}	string	metrics
//	@Router		/metrics [get]
func Metrics() {} // nolint
