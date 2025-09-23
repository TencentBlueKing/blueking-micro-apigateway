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
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Debug 打印 debug 日志
func Debug(format string, vars ...any) {
	logf(slog.LevelDebug, format, vars...)
}

// Infof 打印 info 日志
func Infof(format string, vars ...any) {
	logf(slog.LevelInfo, format, vars...)
}

// Info 打印 info 日志
func Info(vars ...any) {
	logv(slog.LevelInfo, vars...)
}

// Warnf 打印 warn 日志
func Warnf(format string, vars ...any) {
	logf(slog.LevelWarn, format, vars...)
}

// Errorf 打印 error 日志
func Errorf(format string, vars ...any) {
	logf(slog.LevelError, format, vars...)
}

// Error 打印 error 日志
func Error(vars ...any) {
	logv(slog.LevelError, vars...)
}

// ErrorWithCtx 打印 error 上下文日志
func ErrorWithCtx(ctx context.Context, vars ...any) {
	logvCtx(ctx, slog.LevelError, vars...)
}

// DebugWithCtx 打印 debug 上下文日志
func DebugWithCtx(ctx context.Context, format string, vars ...any) {
	logfCtx(ctx, slog.LevelDebug, format, vars...)
}

// InfoFWithCtx  打印 info 上下文日志
func InfoFWithCtx(ctx context.Context, format string, vars ...any) {
	logfCtx(ctx, slog.LevelInfo, format, vars...)
}

// InfoWithCtx 打印 info 上下文日志
func InfoWithCtx(ctx context.Context, vars ...any) {
	logvCtx(ctx, slog.LevelInfo, vars...)
}

// WarnFWithCtx 打印 warn 上下文日志
func WarnFWithCtx(ctx context.Context, format string, vars ...any) {
	logfCtx(ctx, slog.LevelWarn, format, vars...)
}

// ErrorFWithContext 打印 error 上下文 日志
func ErrorFWithContext(ctx context.Context, format string, vars ...any) {
	logfCtx(ctx, slog.LevelError, format, vars...)
}

// Fatalf 打印 fatal 日志到标准输出并退出程序
// Q：为什么 Fatalf 是强制使用 stderr 而非 slog.Default() ？
// A：调用 Fatalf 意味着程序即将退出，此时往标准输出而不是文件打日志是更合理的（避免 Pod 崩溃导致日志无法采集）
func Fatalf(format string, vars ...any) {
	// 由于马上会退出，这里直接 New logger 而不是预先初始化也是可以的
	logger := log.New(os.Stderr, "", log.LstdFlags)
	logger.Fatalf(format, vars...)
}

// Fatal 打印 fatal 日志到标准输出并退出程序
func Fatal(vars ...any) {
	// 由于马上会退出，这里直接 New logger 而不是预先初始化也是可以的
	logger := log.New(os.Stderr, "", log.LstdFlags)
	logger.Fatal(vars...)
}

func logf(level slog.Level, format string, vars ...any) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, vars...), pcs[0])
	_ = slog.Default().Handler().Handle(context.Background(), r)
}

func logfCtx(ctx context.Context, level slog.Level, format string, vars ...any) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, vars...), pcs[0])
	_ = ContextHandler{slog.Default().Handler()}.Handle(ctx, r)
}

func logv(level slog.Level, vars ...any) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, fmt.Sprint(vars...), pcs[0])
	_ = slog.Default().Handler().Handle(context.Background(), r)
}

func logvCtx(ctx context.Context, level slog.Level, vars ...any) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, fmt.Sprint(vars...), pcs[0])
	_ = ContextHandler{slog.Default().Handler()}.Handle(ctx, r)
}
