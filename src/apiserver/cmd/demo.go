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
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"

	cron "github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/envx"
)

// NewDemoCmd demo站点需要用到  ...
func NewDemoCmd() *cobra.Command {
	var cfgFile string
	migrateCmd := cobra.Command{
		Use:   "demo",
		Short: "demo site data init",
		Run: func(cmd *cobra.Command, args []string) {
			// 加载配置
			cfg, err := config.Load(cfgFile)
			if err != nil {
				log.Fatalf("failed to load config: %s", err)
			}

			if cfg.MysqlConfig == nil {
				log.Fatalf("mysql config not found, skip migrate...")
			}

			// init cryptography
			if err = initCryptos(cfg.Crypto.Key, cfg.Crypto.Nonce); err != nil {
				log.Fatalf("failed to init cryptography: %s", err)
			}

			database.InitDBClient(cfg.MysqlConfig, slog.Default())
			// 设置repo db
			repo.SetDefault(database.Client())
			// 设置定时任务
			c := cron.New()
			_, _ = c.AddFunc(envx.Get("DEMO_INIT_DATA_CRON", "0 6 * * *"), func() {
				err := truncateAndInitialize(database.Client())
				if err != nil {
					log.Errorf("Error during truncate and initialize: %v", err)
				}
			})
			c.Start()
			// 保持主程序运行
			select {}
		},
	}
	return &migrateCmd
}

func init() {
	rootCmd.AddCommand(NewDemoCmd())
}

func truncateAndInitialize(db *gorm.DB) error {
	log.Info("Start to truncate and initialize database...")

	// 开启事务
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}
	// defer 事务回滚
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 64<<10)
			n := runtime.Stack(buf, false)
			buf = buf[:n]
			msg := fmt.Sprintf("painic err:%s", buf)
			log.Error(msg)
			tx.Rollback()
		}
	}()

	// 获取所有表名
	var tables []string
	if err := tx.Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get table names: %v", err)
	}

	// 执行 TRUNCATE
	for _, table := range tables {
		if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE `%s`", table)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to truncate table %s: %v", table, err)
		}
	}

	// 执行 SQL 文件
	sqlFileContent, err := os.ReadFile(envx.Get("DEMO_SQL_FILE_PATH", "./bk-apisix-init.sql"))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to read sql file: %v", err)
	}

	// 使用 bufio.Scanner 按分号分隔 SQL 语句
	scanner := bufio.NewScanner(strings.NewReader(string(sqlFileContent)))
	scanner.Split(splitSQLStatements)
	for scanner.Scan() {
		sqlStatement := scanner.Text()
		if strings.TrimSpace(sqlStatement) != "" {
			if err := tx.Exec(sqlStatement).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute sql statement: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to scan sql statements: %v", err)
	}

	gateway, err := biz.GetGateway(context.Background(), 1)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get gateway: %v", err)
	}
	ctx := context.WithValue(context.Background(), constant.GatewayInfoKey, gateway)
	// 进行发布
	if err := biz.PublishAllResource(ctx, 1); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to publish all resources: %v", err)
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	log.Info("Database initialization completed successfully.")
	return nil
}

func splitSQLStatements(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 查找分号，并返回语句（包括分号）
	if i := strings.Index(string(data), ";"); i >= 0 {
		return i + 1, data[:i+1], nil
	}
	// 如果在 EOF 时仍有数据，返回剩余数据
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}
