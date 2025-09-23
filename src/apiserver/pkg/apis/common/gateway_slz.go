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

// Package common ...
package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/base"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/storage"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/validation"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/version"
)

// GatewayInputInfo  网关基本信息
type GatewayInputInfo struct {
	Name string `json:"name" binding:"required" validate:"gatewayName"` // 网关名称
	// 网关control模式：1-direct 2-indirect
	Mode uint8 `json:"mode" binding:"required,gatewayMode" enums:"1,2"`
	// 网关维护者
	Maintainers base.MaintainerList `json:"maintainers"`
	// 网关描述
	Description string `json:"description"`
	// apisix版本
	APISIXVersion string `json:"apisix_version" binding:"required,apisixVersion"`
	// apisix类型: apisix、tapisix、bk-apisix
	APISIXType string `json:"apisix_type" binding:"required,apisixType" enums:"apisix,tapisix,bk-apisix"`

	ReadOnly bool `json:"read_only"` // 是否只读
	// etcd配置
	EtcdConfig
}

// GatewayOutputInfo ...
type GatewayOutputInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"` // 网关名称
	// 网关control模式：1-direct 2-indirect
	Mode        uint8    `json:"mode" binding:"required,gatewayMode" enums:"1,2"`
	ReadOnly    bool     `json:"read_only"`   // 是否只读
	Maintainers []string `json:"maintainers"` // 网关维护者
	Description string   `json:"description"` // 网关描述
	APISIX      APISIX   `json:"apisix"`
	Etcd        EtcdInfo `json:"etcd"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
	Creator     string   `json:"creator"`
	Updater     string   `json:"updater"`
}

// APISIX ...
type APISIX struct {
	Version string `json:"version"` // apisix版本
	Type    string `json:"type"`    // apisix类型: apisix、tapisix、bk-apisix
}

// Etcd : 列表输出使用
type Etcd struct {
	InstanceID string            `json:"instance_id"` // 实例ID
	EndPoints  base.EndpointList `json:"endpoints"`   // etcd集群地址
	Prefix     string            `json:"prefix"`      // etcd前缀
}

// EtcdInfo etcd查看详情配置
type EtcdInfo struct {
	InstanceID string            `json:"instance_id"` // 实例ID
	EndPoints  base.EndpointList `json:"endpoints"`   // etcd集群地址
	// etcd连接类型:http/https
	SchemaType string `json:"schema_type" binding:"required,etcdSchemaType" enums:"http,https"`
	Prefix     string `json:"prefix"`    // etcd前缀
	Username   string `json:"username"`  // etcd用户名
	Password   string `json:"password"`  // etcd密码
	CaCert     string `json:"ca_cert"`   // etcd ca证书
	CertCert   string `json:"cert_cert"` // etcd cert证书
	CertKey    string `json:"cert_key"`  // etcd cert key
}

// EtcdConfig etcd配置(创建、更新)
type EtcdConfig struct {
	EtcdEndPoints base.EndpointList `json:"etcd_endpoints" binding:"required,etcdEndPoints"` // etcd集群地址
	// etcd连接类型:http/https
	EtcdSchemaType string `json:"etcd_schema_type" binding:"required,etcdSchemaType" enums:"http,https"`
	EtcdPrefix     string `json:"etcd_prefix" binding:"required"`             // etcd前缀
	EtcdUsername   string `json:"etcd_username" binding:"omitempty,required"` // etcd用户名
	EtcdPassword   string `json:"etcd_password" binding:"omitempty,required"` // etcd密码
	EtcdCACert     string `json:"etcd_ca_cert,omitempty"`                     // etcd ca证书
	EtcdCertCert   string `json:"etcd_cert_cert,omitempty"`                   // etcd cert
	EtcdCertKey    string `json:"etcd_cert_key,omitempty"`                    // etcd cert key
}

// CheckGatewayMode 校验网关模式
func CheckGatewayMode(fl validator.FieldLevel) bool {
	value := uint8(fl.Field().Uint())
	_, ok := constant.GatewayModeMap[value]
	return ok
}

// CheckAPISIXType 校验apisix类型
func CheckAPISIXType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, ok := constant.APISIXTypeMap[value]
	return ok
}

// CheckEtcdEndPoints 校验 etcd 地址
func CheckEtcdEndPoints(fl validator.FieldLevel) bool {
	endPoints := fl.Field().Interface().(base.EndpointList)
	for _, endpoint := range endPoints {
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			return false
		}
	}
	return true
}

// CheckEtcdSchemaType 校验etcd连接类型
func CheckEtcdSchemaType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, ok := constant.SchemaTypeMap[value]
	return ok
}

// CheckAPISIXVersion 校验apisix版本
func CheckAPISIXVersion(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	ver, err := version.ToXVersion(value)
	if err != nil {
		return false
	}
	// todo:  string类型下沉到SupportAPISIXVersionMap
	_, ok := constant.SupportAPISIXVersionMap[string(ver)]
	return ok
}

// EtcdConfigCheckValidation etcd配置校验
func EtcdConfigCheckValidation(ctx context.Context, sl validator.StructLevel) {
	// 如果是更新操作，跳过
	if ginx.GetGatewayInfoFromContext(ctx) != nil && ginx.GetGatewayInfoFromContext(ctx).ID != 0 {
		return
	}
	etcdConfig := sl.Current().Interface().(EtcdConfig)
	switch etcdConfig.EtcdSchemaType {
	case constant.HTTPS:
		if etcdConfig.EtcdCertCert == "" || etcdConfig.EtcdCertKey == "" || etcdConfig.EtcdCACert == "" {
			sl.ReportError(etcdConfig.EtcdSchemaType, "schema_type",
				"schema_type", "etcd_https_error", etcdConfig.EtcdSchemaType)
			return
		}
	case constant.HTTP:
		if etcdConfig.EtcdUsername == "" && etcdConfig.EtcdPassword == "" {
			sl.ReportError(etcdConfig.EtcdUsername, "etcd_username",
				"etcd_username", "etcd_http_error", etcdConfig.EtcdUsername)
			return
		}
	}
}

// ValidateGatewayName 校验网关名称是否重复
func ValidateGatewayName(ctx context.Context, fl validator.FieldLevel) bool {
	gatewayName := fl.Field().String()
	if gatewayName == "" {
		return false
	}
	var gatewayID int
	gatewayInfo := ginx.GetGatewayInfoFromContext(ctx)
	if gatewayInfo != nil {
		gatewayID = gatewayInfo.ID
	}
	return biz.ExistsGatewayName(ctx, gatewayName, gatewayID)
}

// CheckEtcdConnAndAPISIXInstance 检查etcd连接和apisix实例
func CheckEtcdConnAndAPISIXInstance(gatewayID int, etcdConf EtcdConfig) (string, string, error) {
	etcdStoreConfig := base.EtcdConfig{
		Endpoint: etcdConf.EtcdEndPoints.EndpointJoin(),
		Prefix:   etcdConf.EtcdPrefix,
		Username: etcdConf.EtcdUsername,
		Password: etcdConf.EtcdPassword,
		CACert:   etcdConf.EtcdCACert,
		CertCert: etcdConf.EtcdCertCert,
		CertKey:  etcdConf.EtcdCertKey,
	}

	// 检查etcd连接
	etcdStore, err := storage.NewEtcdStorage(etcdStoreConfig)
	if err != nil {
		return "", "", err
	}
	defer etcdStore.GetClient().Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := etcdStore.List(ctx, fmt.Sprintf("%s/data_plane/server_info", etcdStoreConfig.Prefix))
	if err != nil && !errors.Is(err, storage.KeyNotFoundError) {
		return "", "", err
	}
	var apisixVersion, instanceID string
	if len(res) > 0 {
		resData := gjson.Parse(res[0].Value)
		instanceID = resData.Get("id").String()
		apisixVersion = resData.Get("version").String()
	}
	// 校验实例id是否存在
	if instanceID != "" {
		gateways, err := biz.GetGatewayEtcdConfigList(ctx, "instance_id", instanceID)
		if err != nil {
			return "", "", err
		}
		// 排除自己
		if len(gateways) > 0 && (gatewayID != 0 && gateways[0].ID != gatewayID) {
			return "", "", fmt.Errorf(
				"网关 apisix 实例[%s] 已在另一个网关 [%s] 中注册, 无法重复注册托管, 请联系该网关负责人 [%s] 添加权限", instanceID,
				gateways[0].Name, strings.Join(gateways[0].Maintainers, ","))
		}
	}

	// 校验 prefix相同的网关
	gateways, err := biz.GetGatewayEtcdConfigList(ctx, "prefix", etcdStoreConfig.Prefix)
	if err != nil {
		return "", "", err
	}
	for _, gateway := range gateways {
		// 编辑网关时校验去重需要排除自己
		if gatewayID != 0 && gatewayID == gateway.ID {
			continue
		}
		endpoints := gateway.EtcdConfig.Endpoint.Endpoints()
		for _, storeEndpoint := range etcdStoreConfig.Endpoint.Endpoints() {
			for _, endpoint := range endpoints {
				// 校验 endpoint 是否存在
				cleanStoreEndpoint := strings.TrimPrefix(
					strings.TrimPrefix(storeEndpoint, constant.HTTP),
					constant.HTTPS,
				)
				cleanEndpoint := strings.TrimPrefix(
					strings.TrimPrefix(endpoint, constant.HTTP),
					constant.HTTPS,
				)
				if storeEndpoint != "" && endpoint != "" && cleanStoreEndpoint == cleanEndpoint {
					err = fmt.Errorf(
						"etcd 地址[%s] 已在另一个网关 [%s] 中存在, 请勿重复提交",
						endpoint,
						gateway.Name,
					)
					return "", "", err
				}
			}
		}
	}
	return apisixVersion, instanceID, nil
}

// GatewayToOutputInfo ...
func GatewayToOutputInfo(gatewayInfo *model.Gateway) GatewayOutputInfo {
	output := GatewayOutputInfo{
		ID:          gatewayInfo.ID,
		Name:        gatewayInfo.Name,
		Mode:        gatewayInfo.Mode,
		Maintainers: gatewayInfo.Maintainers,
		Description: gatewayInfo.Desc,
		APISIX: APISIX{
			Version: gatewayInfo.APISIXVersion,
			Type:    gatewayInfo.APISIXType,
		},
		ReadOnly: gatewayInfo.ReadOnly,
		Etcd: EtcdInfo{
			InstanceID: gatewayInfo.EtcdConfig.InstanceID,
			EndPoints:  gatewayInfo.EtcdConfig.Endpoint.Endpoints(),
			Prefix:     gatewayInfo.EtcdConfig.Prefix,
			SchemaType: gatewayInfo.EtcdConfig.GetSchemaType(),
			Username:   gatewayInfo.EtcdConfig.Username,
			Password:   constant.SensitiveInfoFiledDisplay,
			CaCert:     gatewayInfo.EtcdConfig.GetMaskCaCert(),
			CertCert:   gatewayInfo.EtcdConfig.GetMaskCertCert(),
			CertKey:    gatewayInfo.EtcdConfig.GetMaskCertKey(),
		},
		CreatedAt: gatewayInfo.CreatedAt.Unix(),
		UpdatedAt: gatewayInfo.UpdatedAt.Unix(),
		Creator:   gatewayInfo.Creator,
		Updater:   gatewayInfo.Updater,
	}
	return output
}

func init() {
	validation.AddBizFieldTagValidator(
		"gatewayMode",
		CheckGatewayMode,
		validation.GetEnumTransMsgFromUint8KeyMap(constant.GatewayModeMap, false),
	)
	validation.AddBizFieldTagValidator(
		"apisixType",
		CheckAPISIXType,
		validation.GetEnumTransMsgFromStringKeyMap(constant.APISIXTypeMap, true),
	)
	validation.AddBizFieldTagValidator(
		"etcdEndPoints",
		CheckEtcdEndPoints,
		"{0}:{1} etcd 地址必须以 http:// 或 https:// 开头",
	)
	validation.AddBizFieldTagValidator(
		"etcdSchemaType",
		CheckEtcdSchemaType,
		validation.GetEnumTransMsgFromStringKeyMap(constant.SchemaTypeMap, true),
	)
	validation.AddBizFieldTagValidator(
		"apisixVersion",
		CheckAPISIXVersion,
		validation.GetEnumTransMsgFromStringKeyMap(constant.SupportAPISIXVersionMap, true),
	)
	validation.AddBizFieldTagValidatorWithCtx("gatewayName", ValidateGatewayName,
		"{0}:{1} 该网关实例已经被存在的网关注册")
	validation.AddBizStructValidator(EtcdConfig{}, EtcdConfigCheckValidation, map[string]string{
		"etcd_https_error": "{0}={1} 证书或密钥或 ca 不能为空",
		"etcd_http_error":  "{0}={1} 用户名或密码不能为空",
	})
}
