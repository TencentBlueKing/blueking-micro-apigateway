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

// Package logging 实现 resty.Logger 接口
package logging

import (
	"log/slog"

	resty "github.com/go-resty/resty/v2"
)

// Logger 用于实现 resty.Logger
type Logger struct{}

// New 实例化 Logger
func New() *Logger {
	return &Logger{}
}

// Errorf ...
func (l *Logger) Errorf(format string, v ...any) {
	logf(slog.LevelError, format, v...)
}

// Warnf ...
func (l *Logger) Warnf(format string, v ...any) {
	logf(slog.LevelWarn, format, v...)
}

// Debugf ...
func (l *Logger) Debugf(format string, v ...any) {
	logf(slog.LevelDebug, format, v...)
}

var _ resty.Logger = (*Logger)(nil)
