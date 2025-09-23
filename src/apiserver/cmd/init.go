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

package cmd

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
)

func initLogger(cfg *config.LogConfig) error {
	// 自动创建日志目录
	if err := os.MkdirAll(cfg.Dir, os.ModePerm); err != nil {
		// 只有当错误不是 “目录已存在” 时，需要抛出错误
		if !os.IsExist(err) {
			return errors.Wrapf(err, "creating log dir %s", cfg.Dir)
		}
	}

	// 输出位置
	writerName := "file"
	if cfg.ForceToStdout {
		writerName = "stdout"
	}

	// 初始化默认 Logger
	loggerName := "default"
	if err := logging.InitLogger(loggerName, &logging.Options{
		Level:             cfg.Level,
		HandlerName:       lo.Ternary(writerName == "stdout", "json", "text"),
		WriterName:        writerName,
		WriterConfig:      map[string]string{"filename": filepath.Join(cfg.Dir, loggerName+".log")},
		SentryReportLevel: cfg.SentryReportLevel,
	}); err != nil {
		return errors.Wrapf(err, "creating logger %s", loggerName)
	}

	// 初始化 Gorm Logger
	loggerName = "gorm"
	if err := logging.InitLogger(loggerName, &logging.Options{
		Level:             logging.GormLogLevel,
		HandlerName:       "json",
		WriterName:        writerName,
		WriterConfig:      map[string]string{"filename": filepath.Join(cfg.Dir, loggerName+".log")},
		SentryReportLevel: cfg.SentryReportLevel,
	}); err != nil {
		return errors.Wrapf(err, "creating logger %s", loggerName)
	}

	// 初始化 Gin Logger
	loggerName = "gin"
	if err := logging.InitLogger(loggerName, &logging.Options{
		Level:             logging.GinLogLevel,
		HandlerName:       "json",
		WriterName:        writerName,
		WriterConfig:      map[string]string{"filename": filepath.Join(cfg.Dir, loggerName+".log")},
		SentryReportLevel: cfg.SentryReportLevel,
	}); err != nil {
		return errors.Wrapf(err, "creating logger %s", loggerName)
	}

	return nil
}

func initCryptos(key, nonce string) error {
	if key == "" {
		return errors.New("cryptoKey should be configured")
	}

	if nonce == "" {
		return errors.New("cryptoNonce should be configured")
	}

	validEncryptKeyRegex := regexp.MustCompile("^[a-zA-Z0-9]{32}$")
	errInvalidEncryptKey := "invalid encrypt_key: encrypt_key should " +
		"contains letters(a-z, A-Z), numbers(0-9), length should be 32 bit"
	if !validEncryptKeyRegex.MatchString(key) {
		return errors.New(errInvalidEncryptKey)
	}

	err := cryptography.Init(key, nonce)
	if err != nil {
		return err
	}
	return nil
}
