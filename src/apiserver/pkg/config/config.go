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

// Package config 管理蓝鲸 SaaS 配置项，支持从配置文件 / 环境变量中读取配置
package config

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"

	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

// G ...
var G *Config

// Load 加载配置
func Load(cfgFile string) (*Config, error) {
	var cfg *Config
	var err error

	if cfgFile != "" {
		// 若已经指定配置文件，则从配置文件中加载
		log.Infof("load config from file: %s", cfgFile)
		cfg, err = loadConfigFromFile(cfgFile)
	} else {
		// 若没有指定配置文件，则环境变量构建配置
		log.Infof("config file not specified, load config from env vars")
		cfg, err = loadConfigFromEnv()
	}

	if err != nil {
		cfgFrom := lo.Ternary(cfgFile != "", "file: "+cfgFile, "env vars")
		return nil, errors.Wrapf(err, "failed to load config from: %+v", cfgFrom)
	}
	// 设置全局环境变量
	G = cfg
	return cfg, nil
}

// IsDemoMode 是否为演示模式
func IsDemoMode() bool {
	if G == nil {
		return false
	}
	return G.Service.DemoMode
}
