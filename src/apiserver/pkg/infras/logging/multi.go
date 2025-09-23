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

package logging

import (
	"context"
	"log/slog"
)

// MultiHandler 自定义 Handler，同时写入 logFormatHandler 和 Sentry
type MultiHandler struct {
	logFormatHandler slog.Handler
	sentryHandler    slog.Handler
}

// Enabled 判断是否启用
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// 任一 Handler 启用即返回 true
	return h.logFormatHandler.Enabled(ctx, level) || h.sentryHandler.Enabled(ctx, level)
}

// Handle ...
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	// 克隆 Record 避免数据竞争
	if h.logFormatHandler.Enabled(ctx, r.Level) {
		if err := h.logFormatHandler.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	if h.sentryHandler.Enabled(ctx, r.Level) {
		if err := h.sentryHandler.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	return nil
}

// WithAttrs ...
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MultiHandler{
		logFormatHandler: h.logFormatHandler.WithAttrs(attrs),
		sentryHandler:    h.sentryHandler.WithAttrs(attrs),
	}
}

// WithGroup ...
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	return &MultiHandler{
		logFormatHandler: h.logFormatHandler.WithGroup(name),
		sentryHandler:    h.sentryHandler.WithGroup(name),
	}
}
