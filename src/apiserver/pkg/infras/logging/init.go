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
	"fmt"
	"log/slog"
	"strings"

	sentry "github.com/getsentry/sentry-go"
	slogsentry "github.com/samber/slog-sentry/v2"
)

var loggers map[string]*slog.Logger

// GetLogger 获取指定 Logger
func GetLogger(name string) *slog.Logger {
	if logger, ok := loggers[name]; ok {
		return logger
	}

	// 不存在则返回默认的
	return slog.Default()
}

// InitLogger ...
func InitLogger(name string, opts *Options) (err error) {
	if loggers == nil {
		loggers = make(map[string]*slog.Logger)
	}

	// 已存在，则忽略，不需要再初始化
	if _, ok := loggers[name]; ok {
		return nil
	}

	if loggers[name], err = newLogger(opts); err != nil {
		return err
	}
	if name == "default" {
		// SetDefault 会改变 Golang slog 的 默认 logging
		// 同时会改变 Golang log 包使用的默认 log.Logger
		slog.SetDefault(loggers[name])
	}

	return nil
}

// 根据配置生成 Logger
func newLogger(opts *Options) (*slog.Logger, error) {
	w, err := newWriter(opts.WriterName, opts.WriterConfig)
	if err != nil {
		return nil, err
	}

	level, err := toSlogLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	sentryLevel, err := toSlogLevel(opts.SentryReportLevel)
	if err != nil {
		return nil, err
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: replaceAttr,
	}
	// 组合两个 Handler
	multiHandler := &MultiHandler{
		sentryHandler: slogsentry.Option{
			Hub:   sentry.CurrentHub(),
			Level: sentryLevel,
		}.NewSentryHandler(),
	}
	switch opts.HandlerName {
	case "text":
		multiHandler.logFormatHandler = slog.NewTextHandler(w, handlerOpts)
		return slog.New(multiHandler), nil
	case "json":
		multiHandler.logFormatHandler = slog.NewJSONHandler(w, handlerOpts)
		return slog.New(multiHandler), nil
	}
	return nil, fmt.Errorf("[%s] handler not supported", opts.HandlerName)
}

// toSlogLevel 将配置输入的日志级别转换为 slog Level 对象
func toSlogLevel(level string) (slog.Level, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	}

	return slog.LevelInfo, fmt.Errorf("[%s] level not supported", level)
}
