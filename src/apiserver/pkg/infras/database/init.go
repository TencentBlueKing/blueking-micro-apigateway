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

// Package database 提供了数据库相关的封装，目前实现的是主流的 gorm + mysql
// SaaS 开发者可根据需要替换为其他 orm（如 SQLBoiler，Ent）或者其他数据库（如 mongodb）
// 如果对性能要有很高的话，也可以考虑 sqlx，这是一个高性能的标准 sql 库增强 & 扩展包，
// 其缺点是没有提供完整的 ORM 功能（如自动迁移，关系处理等等），开发者用起来不太方便（需要写不少的 SQL）
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

var db *gorm.DB

var initOnce sync.Once

const (
	// string 类型字段的默认长度
	defaultStringSize = 256
	// 默认批量创建数量
	defaultBatchSize = 100
	// 默认最大空闲连接
	defaultMaxIdleConns = 20
	// 默认最大连接数
	defaultMaxOpenConns = 100
)

// Client 获取数据库客户端
func Client() *gorm.DB {
	if db == nil {
		log.Fatalf("database client not init")
	}
	return db
}

// SetClient 设置数据库客户端 (only for test)
func SetClient(client *gorm.DB) {
	db = client
}

// InitDBClient 初始化数据库客户端
func InitDBClient(cfg *config.MysqlConfig, slogger *slog.Logger) {
	if db != nil {
		return
	}
	if cfg == nil {
		log.Fatalf("mysql config is required when init database client")
	}
	initOnce.Do(func() {
		dbInfo := fmt.Sprintf("mysql %s:%d/%s", cfg.Host, cfg.Port, cfg.Name)

		var err error
		if db, err = newClient(cfg, slogger); err != nil {
			log.Fatalf("failed to connect database %s: %s", dbInfo, err)
		} else {
			log.Infof("database: %s connected", dbInfo)
		}
	})
}

// 初始化 DB Client
func newClient(cfg *config.MysqlConfig, slogger *slog.Logger) (*gorm.DB, error) {
	sqlDB, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(defaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(defaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查 DB 是否可用
	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	mysqlCfg := mysql.Config{
		DSN:                       cfg.DSN(),
		DefaultStringSize:         defaultStringSize,
		SkipInitializeWithVersion: false,
	}
	gormCfg := &gorm.Config{
		ConnPool: sqlDB,
		// 禁用默认事务（需要手动管理）
		SkipDefaultTransaction: true,
		// 缓存预编译语句
		PrepareStmt: true,
		// Mysql 本身即不支持嵌套事务
		DisableNestedTransaction: true,
		// 批量操作数量
		CreateBatchSize: defaultBatchSize,
		// 数据库迁移时，忽略外键约束
		DisableForeignKeyConstraintWhenMigrating: true,
		// 日志相关
		Logger: slogGorm.New(
			slogGorm.WithHandler(slogger.Handler()),
			slogGorm.WithSlowThreshold(200*time.Millisecond),
			slogGorm.WithRecordNotFoundError(),
			slogGorm.WithTraceAll(),
			slogGorm.SetLogLevel(slogGorm.DefaultLogType, constant.LOG_NOTICE),
		),
	}
	client, err := gorm.Open(mysql.New(mysqlCfg), gormCfg)
	if err != nil {
		return nil, err
	}

	if config.G.Tracing.DBAPIEnabled() {
		err = client.Use(tracing.NewPlugin())
		if err != nil {
			return client, err
		}
	}

	return client, nil
}
