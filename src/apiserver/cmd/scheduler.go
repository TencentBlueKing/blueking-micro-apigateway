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
	"github.com/spf13/cobra"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/async"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

// NewSchedulerCmd 用于创建定时任务调度器启动命令
// 需要注意的是：为避免重复执行定时任务，需要确保同时只有一个 scheduler 正在运行
// 如果希望同时启动多个 scheduler 启动，则需要添加诸如 redis / zk 这样的分布式锁
func NewSchedulerCmd() *cobra.Command {
	var cfgFile string

	schedulerCmd := cobra.Command{
		Use:   "scheduler",
		Short: "scheduler apply tasks based on cron expression, please ensure only one running scheduler.",
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

			// 初始化 DB Client
			database.InitDBClient(cfg.MysqlConfig, logging.GetLogger("gorm"))
			// 初始化 task server
			async.InitTaskScheduler()

			srv := async.Scheduler()
			// 加载周期任务
			if err = srv.LoadTasks(); err != nil {
				logging.Fatal(err.Error())
			}
			// 启用调度服务器
			srv.Run()
		},
	}

	// 配置文件路径，如果未指定，会从环境变量读取各项配置
	// 注意：目前平台未默认提供配置文件，需通过 `模块配置 - 挂载卷` 添加
	schedulerCmd.Flags().StringVar(&cfgFile, "conf", "", "config file")

	return &schedulerCmd
}

func init() {
	rootCmd.AddCommand(NewSchedulerCmd())
}
