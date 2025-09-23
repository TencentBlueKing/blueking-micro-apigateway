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

package util

import (
	"fmt"
	"log"
	"os"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/database"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
)

var (
	genTestDb     *gorm.DB
	genTestOnce   = &sync.Once{}
	genTestDbName = "file::memory:?cache=shared&_mutex=no&_journal=WAL"
	mu            sync.Mutex
)

// InitEmbedDb 初始化内存数据库
func InitEmbedDb() {
	mu.Lock()
	defer mu.Unlock()
	genTestOnce.Do(func() {
		var collectedSQL []string
		// 创建带 SQL 收集功能的 DB 实例
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			DryRun: true, // 启用 dry run 模式
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					Colorful: true,
					LogLevel: logger.Info,
				},
			),
		})
		if err != nil {
			panic(err)
		}
		// 注册到所有操作后的回调
		db.Callback().Raw().After("*").Register("collect_sql", func(d *gorm.DB) {
			if sql := d.Statement.SQL.String(); sql != "" {
				fmt.Printf("[DEBUG] Captured SQL: %s\n", sql)
				collectedSQL = append(collectedSQL, sql)
			}
		})

		models := []interface{}{
			model.Gateway{},
			model.Route{},
			model.Service{},
			model.Upstream{},
			model.PluginConfig{},
			model.PluginMetadata{},
			model.Consumer{},
			model.ConsumerGroup{},
			model.GlobalRule{},
			model.GatewaySyncData{},
			model.GatewayReleaseVersion{},
			model.OperationAuditLog{},
			model.Proto{},
			model.SSL{},
			model.SystemConfig{},
			model.GatewayCustomPluginSchema{},
			model.GatewayResourceSchemaAssociation{},
			model.StreamRoute{},
		}
		for _, m := range models {
			// 执行迁移
			err = db.AutoMigrate(m)
			if err != nil {
				panic(err)
			}
		}

		// 创建真实数据库连接
		genTestDb, err = gorm.Open(sqlite.Open(genTestDbName), &gorm.Config{})
		if err != nil {
			panic(fmt.Errorf("open real db failed: %w", err))
		}
		_ = genTestDb.Exec("PRAGMA journal_mode=WAL;")
		// 执行 SQL 语句
		for _, sql := range collectedSQL {
			if err := genTestDb.Exec(sql).Error; err != nil {
				// sqlite3 在执行迁移时，对于索引不太兼容
				continue
			}
			log.Println("exec sql: ", sql)
		}
		repo.SetDefault(genTestDb)
		database.SetClient(genTestDb)
	})
}
