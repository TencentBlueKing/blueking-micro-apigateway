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
	"context"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
)

// NewPermissionCmd ...
func NewPermissionCmd() *cobra.Command {
	var cfgFile string
	var user string
	var action string

	migrateCmd := cobra.Command{
		Use:   "permission",
		Short: "add/delete user permission.",
		Run: func(cmd *cobra.Command, args []string) {
			if action == "" {
				log.Fatalf("action is required: add/delete")
			}
			if user == "" {
				log.Fatalf("user id is required")
			}
			baseCtx := context.Background()
			// 加载配置
			cfg, err := config.Load(cfgFile)
			if err != nil {
				log.Fatalf("failed to load config: %s", err)
			}

			if cfg.MysqlConfig == nil {
				log.Fatalf("mysql config not found, skip migrate...")
			}

			database.InitDBClient(cfg.MysqlConfig, slog.Default())
			// 设置repo db
			repo.SetDefault(database.Client())
			switch action {
			case "add":
				err = biz.AddAllowUsers(baseCtx, []string{user})
				if err != nil {
					log.Fatalf("add user %s failed: %s", user, err)
					return
				}
			case "delete":
				err = biz.RemoveUsers(baseCtx, []string{user})
				if err != nil {
					log.Fatalf("delete user %s failed: %s", user, err)
					return
				}
			default:
				log.Fatalf("action %s not support", action)
			}
			var users dto.UserWhiteList
			err = biz.GetSystemConfigWithEntity(baseCtx, constant.SystemConfigUserWhitest, &users)
			if err != nil {
				log.Fatalf("get user whitelist failed: %s", err)
				return
			}
			log.Infof("%s user success. all allowerd users: %v success", action, users)
		},
	}

	// 配置文件路径，如果未指定，会从环境变量读取各项配置
	// 注意：目前平台未默认提供配置文件，需通过 `模块配置 - 挂载卷` 添加
	migrateCmd.Flags().StringVar(&cfgFile, "conf", "", "config file")
	migrateCmd.Flags().StringVar(&action, "action", "", "add/delete")
	migrateCmd.Flags().StringVar(&user, "user", "", "user id")
	return &migrateCmd
}

func init() {
	rootCmd.AddCommand(NewPermissionCmd())
}
