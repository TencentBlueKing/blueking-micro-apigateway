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

package constant

import "log/slog"

// RequestIDCtxKey ...
const (
	// RequestIDCtxKey request_id 在 context 中的 key
	RequestIDCtxKey = "request_id"

	// RequestIDHeaderKey request_id 在 HTTP Header 中的 key
	RequestIDHeaderKey = "X-Request-Id"

	// OpenAPITokenHeaderKey openapi token 在 HTTP Header 中的 key
	OpenAPITokenHeaderKey = "X-BK-API-TOKEN"
	// ErrorCtxKey error 在 context 中的 key
	ErrorCtxKey = "error"

	// UserLanguageKey user language 在 cookies / session 中的 key
	UserLanguageKey = "blueking_language"

	// SensitiveInfoFiledDisplay 敏感信息默认值
	SensitiveInfoFiledDisplay = "******"

	// AccessTokenLength access token 长度
	AccessTokenLength = 36
)

// CtxKey ...
type CtxKey string

// GatewayInfoKey gateway info 在 context 中的 key
const GatewayInfoKey CtxKey = "gateway_info"

// APISIXValidateErrKey apisix validate err 在 context 中的 key
const APISIXValidateErrKey CtxKey = "apisix_validate_err"

// UserIDKey user id 在 cookies / session 中的 key
const UserIDKey CtxKey = "bk_uid"

// ResourceTypeKey resource type 在 context 中的 key
const ResourceTypeKey CtxKey = "resource_type"

// DbTxKey transaction 在 context 中的 key
const DbTxKey CtxKey = "db_tx"

// SystemConfigUserWhitest system config key
const (
	// SystemConfigUserWhitest user whitelist
	SystemConfigUserWhitest = "user_whitelist"
)

// LOG_NOTICE ...
const (
	LOG_NOTICE = slog.Level(5)
)

// DBBatchSize ...
const (
	DBBatchSize = 1000
)
