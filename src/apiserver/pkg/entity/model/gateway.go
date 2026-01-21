/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
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

package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/goutil/arrutil"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/version"
)

// Gateway 网关基础信息表
type Gateway struct {
	ID            int            `gorm:"column:id;primaryKey;autoIncrement"`               // 自增主键
	Name          string         `gorm:"column:name;type:varchar(255)"`                    // 网关名称
	Mode          uint8          `gorm:"column:mode;type:tinyint"`                         // 纳管/直营
	Maintainers   pq.StringArray `gorm:"column:maintainers;type:text"`                     // 网关负责人
	Desc          string         `gorm:"column:desc;type:text"`                            // 网关描述
	APISIXType    string         `gorm:"column:apisix_type;type:varchar(255)"`             // apisix/bk-apisix/tapisix
	APISIXVersion string         `gorm:"column:apisix_version;type:varchar(255)"`          // apisix 实例版本
	EtcdConfig    EtcdConfig     `gorm:"column:etcd_config;type:json"`                     // etcd 组件配置，JSON 存储
	Token         string         `gorm:"column:token;type:varchar(255)"`                   // 网关 token
	ReadOnly      bool           `gorm:"column:read_only;type:tinyint"`                    // 是否只读
	LastSyncedAt  time.Time      `json:"last_synced_at" gorm:"type:datetime;default:null"` // 上次同步时间
	auditSnapshot datatypes.JSON `gorm:"-"`                                                // 用于审计日志传递网关信息，不持久化到数据库
	BaseModel
}

// EtcdConfig etcd 配置
type EtcdConfig struct {
	InstanceID string `json:"instance_id,omitempty"`
	base.EtcdConfig
}

// Value 实现 driver.Valuer 接口
func (e EtcdConfig) Value() (driver.Value, error) {
	// 将结构体转换为 JSON 字符串
	return json.Marshal(e)
}

// Scan 实现 sql.Scanner 接口
func (e *EtcdConfig) Scan(value any) error {
	// 将数据库中的 JSON 字符串转换回结构体
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, e)
}

// TableName ...
func (Gateway) TableName() string {
	return "gateway"
}

// GetAPISIXVersionX 获取 apisix 版本号
func (g Gateway) GetAPISIXVersionX() constant.APISIXVersion {
	apisVersion, _ := version.ToXVersion(g.APISIXVersion)
	return apisVersion
}

// HasPermission 是否有权限
func (g *Gateway) HasPermission(userID string) bool {
	// demo 模式有所有网关的权限
	if config.IsDemoMode() {
		return true
	}
	return arrutil.HasValue(g.Maintainers, userID)
}

// BeforeCreate 创建前钩子
func (g *Gateway) BeforeCreate(tx *gorm.DB) (err error) {
	if err := g.HandleEtcdConfig(false); err != nil {
		return err
	}
	return nil
}

// AfterCreate 创建之后钩子
func (g *Gateway) AfterCreate(tx *gorm.DB) (err error) {
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeCreate)
}

// BeforeUpdate 更新前钩子
func (g *Gateway) BeforeUpdate(tx *gorm.DB) (err error) {
	if err := g.HandleEtcdConfig(false); err != nil {
		return err
	}
	g.auditSnapshot, err = g.Snapshot(tx)
	return err
}

// AfterUpdate 更新后钩子
func (g *Gateway) AfterUpdate(tx *gorm.DB) (err error) {
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeUpdate)
}

// BeforeDelete 删除前钩子
func (g *Gateway) BeforeDelete(tx *gorm.DB) (err error) {
	// 添加审计
	return g.AddAuditLog(tx, constant.OperationTypeDelete)
}

// AfterFind 查询后钩子
func (g *Gateway) AfterFind(tx *gorm.DB) (err error) {
	return g.HandleEtcdConfig(true)
}

// CopyAndMaskPassword 复制同时隐私密码
func (g *Gateway) CopyAndMaskPassword() Gateway {
	gateway := Gateway{
		ID:            g.ID,
		Name:          g.Name,
		Mode:          g.Mode,
		Maintainers:   g.Maintainers,
		Desc:          g.Desc,
		APISIXType:    g.APISIXType,
		APISIXVersion: g.APISIXVersion,
		EtcdConfig:    g.EtcdConfig,
		Token:         g.Token,
		ReadOnly:      g.ReadOnly,
		LastSyncedAt:  g.LastSyncedAt,
		BaseModel:     g.BaseModel,
	}
	if gateway.EtcdConfig.GetSchemaType() == constant.HTTP {
		pwd := gateway.EtcdConfig.Password
		if len(pwd) >= 6 {
			gateway.EtcdConfig.Password = fmt.Sprintf("%s****%s", pwd[:3], pwd[len(pwd)-3:])
		} else {
			gateway.EtcdConfig.Password = fmt.Sprintf("%s****", pwd[:3])
		}
	}
	return gateway
}

// Snapshot 获取网关当前数据
func (g *Gateway) Snapshot(tx *gorm.DB) (datatypes.JSON, error) {
	var origin Gateway
	if err := tx.First(&origin, "id = ?", g.ID).Error; err != nil {
		return nil, err
	}
	if err := origin.HandleEtcdConfig(false); err != nil {
		return nil, err
	}
	return json.Marshal(origin.CopyAndMaskPassword())
}

// AddAuditLog 添加审计
func (g *Gateway) AddAuditLog(tx *gorm.DB, operation constant.OperationType) (err error) {
	if g.ID == 0 {
		return nil
	}
	// auditCallback 方法中会过滤 delete 操作的 dataAfter
	dataAfter, err := g.Snapshot(tx)
	if err != nil {
		return err
	}
	dataBefore := datatypes.JSON{}
	// 只有 update/delete 操作有 dataBefore
	switch operation {
	case constant.OperationTypeUpdate:
		dataBefore = g.auditSnapshot
	case constant.OperationTypeDelete:
		dataBefore, err = g.Snapshot(tx)
		if err != nil {
			return err
		}
	}
	return auditCallback(
		tx,
		g.ID,
		strconv.Itoa(g.ID),
		g.Updater,
		"",
		operation,
		constant.Gateway,
		dataBefore,
		dataAfter,
	)
}

// HandleEtcdConfig 处理 etcd 配置
func (g *Gateway) HandleEtcdConfig(read bool) (err error) {
	g.EtcdConfig.Password, err = getSecret(g.EtcdConfig.Password, read)
	if err != nil {
		return err
	}

	g.EtcdConfig.CertCert, err = getSecret(g.EtcdConfig.CertCert, read)
	if err != nil {
		return err
	}

	g.EtcdConfig.CertKey, err = getSecret(g.EtcdConfig.CertKey, read)
	if err != nil {
		return err
	}

	g.EtcdConfig.CACert, err = getSecret(g.EtcdConfig.CACert, read)
	if err != nil {
		return err
	}
	g.Token, err = getSecret(g.Token, read)
	if err != nil {
		return err
	}
	return nil
}

func getSecret(secret string, read bool) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", nil
	}
	if read {
		decryptSecret, err := cryptography.DecryptSecret(secret)
		if err != nil {
			return "", err
		}
		return decryptSecret, nil
	}
	return cryptography.EncryptSecret(secret), nil
}

// RemoveSensitive ... 去除敏感信息
func (g *Gateway) RemoveSensitive() {
	g.EtcdConfig.Password = constant.SensitiveInfoFiledDisplay
	g.Token = constant.SensitiveInfoFiledDisplay
}

// GetEtcdPrefixForList 获取用于 etcd list 操作的 prefix（带 "/" 结尾）
// 确保前缀匹配时不会匹配到同名前缀的其他网关资源
// 例如：prefix "a-b" 会变成 "a-b/"，避免匹配到 "a-b-test" 的资源
func (g *Gateway) GetEtcdPrefixForList() string {
	return NormalizeEtcdPrefix(g.EtcdConfig.Prefix)
}

// GetEtcdResourcePrefix 获取指定资源类型的 etcd prefix（用于 list 操作）
func (g *Gateway) GetEtcdResourcePrefix(resourceType constant.APISIXResource) string {
	basePrefix := g.GetEtcdPrefixForList()
	resourcePrefix := constant.ResourceTypePrefixMap[resourceType]
	if resourcePrefix == "" {
		return basePrefix
	}
	return fmt.Sprintf("%s%s/", basePrefix, resourcePrefix)
}

// NormalizeEtcdPrefix 标准化 etcd prefix，确保以 "/" 结尾
// 用于 etcd list 操作，避免前缀匹配冲突
func NormalizeEtcdPrefix(prefix string) string {
	if prefix == "" {
		return "/"
	}
	// 确保以 "/" 结尾
	if !strings.HasSuffix(prefix, "/") {
		return prefix + "/"
	}
	return prefix
}

// CheckEtcdPrefixConflict 检查两个 etcd prefix 是否存在层级冲突
// 冲突情况：
// 1. 两个 prefix 完全相同（标准化后）
// 2. 一个 prefix 是另一个的父路径（如 a/b 和 a/b/c）
// 注意：a-b 和 a-b-test 不算冲突，因为加上 "/" 后不会互相匹配
func CheckEtcdPrefixConflict(prefix1, prefix2 string) bool {
	// 标准化处理：确保以 "/" 结尾，便于比较层级关系
	p1 := NormalizeEtcdPrefix(prefix1)
	p2 := NormalizeEtcdPrefix(prefix2)

	// 完全相同
	if p1 == p2 {
		return true
	}

	// 检查是否互为父子路径
	// p1="/a/b/" 和 p2="/a/b/c/" 冲突，因为 p1 是 p2 的前缀
	// p1="/a-b/" 和 p2="/a-b-test/" 不冲突，因为互不为前缀
	return strings.HasPrefix(p1, p2) || strings.HasPrefix(p2, p1)
}
