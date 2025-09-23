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

// Package database ...
package database

import (
	"gorm.io/gen"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

// RunMigrate 根据模型对数据库执行迁移
// 注意 (重要)：
// 1. Gorm 的 AutoMigrate 可以自动创建、更新表，字段和索引，但不会删除未使用的列
// 2. AutoMigrate 并非总能感知到字段类型的变化，如 string -> sql.NullString 的变更并不会应用到数据库中
// 3. 针对 2 这种场景，需要手动调用 db.Migrator().AlterColumn(&Model{}, "Field") 强制更新字段
// 4. 如果需要更精细地对数据库进行迁移，或版本管理，可使用：https://github.com/golang-migrate/migrate
// 5. Gorm migrate 更多参考：https://gorm.io/docs/migration.html
func RunMigrate() error {
	return Client().AutoMigrate(
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
	)
}

// RunGenDao 生成 dao 文件
func RunGenDao() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./pkg/repo",
		Mode:    gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
	})
	g.UseDB(Client())
	g.ApplyBasic(
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
	)
	g.Execute()
}
