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

// Package cmd ...
package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/sentry"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/trace"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/router"
)

// NewWebServerCmd ...
func NewWebServerCmd() *cobra.Command {
	var cfgFile string

	wsCmd := cobra.Command{
		Use:   "webserver",
		Short: "webserver start http server.",
		Run: func(cmd *cobra.Command, args []string) {
			// 加载配置
			cfg, err := config.Load(cfgFile)
			if err != nil {
				logging.Fatalf("failed to load config: %s", err)
			}

			// 初始化 Logger
			if err = initLogger(&cfg.Service.Log); err != nil {
				logging.Fatalf("failed to init logging: %s", err)
			}

			// init cryptography
			if err = initCryptos(cfg.Crypto.Key, cfg.Crypto.Nonce); err != nil {
				logging.Fatalf("failed to init cryptography: %s", err)
			}

			// 初始化 DB Client
			database.InitDBClient(cfg.MysqlConfig, logging.GetLogger("gorm"))

			// 设置repo db
			repo.SetDefault(database.Client())

			// 初始化 sentry
			if err = sentry.Init(cfg.Sentry); err != nil {
				logging.Warnf("failed to init sentry: %s", err)
			}

			// 初始化trace
			if err = trace.InitTrace(cfg.Tracing); err != nil {
				logging.Warnf("failed to init trace: %s", err)
			}

			// 启动 Web 服务
			logging.Infof("Starting server at http://0.0.0.0:%d", config.G.Service.Server.Port)
			srv := &http.Server{
				Addr:    ":" + strconv.Itoa(cfg.Service.Server.Port),
				Handler: router.New(logging.GetLogger("gin")),
			}
			go func() {
				if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logging.Fatalf("Start server failed: %s", err)
				}
			}()
			baseCtx := context.Background()
			// 启动同步
			biz.SyncAll(baseCtx)
			ctx, cancel := context.WithTimeout(
				baseCtx, time.Duration(cfg.Service.Server.GraceTimeout)*time.Second,
			)
			defer cancel()
			// 等待中断信号以优雅地关闭服务器
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt)
			<-quit

			logging.Infof("Shutdown server ...")
			if err = srv.Shutdown(ctx); err != nil {
				logging.Fatalf("Shutdown server failed: %s", err)
			}
			logging.Infof("Server exiting")
		},
	}

	// 配置文件路径，如果未指定，会从环境变量读取各项配置
	// 注意：目前平台未默认提供配置文件，需通过 `模块配置 - 挂载卷` 添加
	wsCmd.Flags().StringVar(&cfgFile, "conf", "", "config file")

	return &wsCmd
}

func init() {
	rootCmd.AddCommand(NewWebServerCmd())
}
