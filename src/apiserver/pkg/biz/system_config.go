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

package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gookit/goutil"
	"github.com/gookit/goutil/arrutil"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
)

// AddAllowUsers 添加用户白名单
func AddAllowUsers(ctx context.Context, users []string) error {
	// 新增用户
	var userWhitelist dto.UserWhiteList
	err := GetSystemConfigWithEntity(ctx, constant.SystemConfigUserWhitest, &userWhitelist)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { // 没有配置用户白名单，则新增用户
		return err
	}
	for _, user := range users {
		if goutil.Contains(userWhitelist.Users, user) {
			return fmt.Errorf("user %s already exists", user)
		}
	}
	userWhitelist.Users = append(userWhitelist.Users, users...)
	err = SaveSystemConfig(ctx, constant.SystemConfigUserWhitest, userWhitelist.ToSystemConfig())
	if err != nil {
		return err
	}
	return nil
}

// RemoveUsers 删除用户白名单
func RemoveUsers(ctx context.Context, users []string) error {
	// 新增用户
	var userWhitelist dto.UserWhiteList
	err := GetSystemConfigWithEntity(ctx, constant.SystemConfigUserWhitest, &userWhitelist)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for _, user := range users {
		if !goutil.Contains(userWhitelist.Users, user) {
			continue
		}
		userWhitelist.Users = arrutil.Remove(userWhitelist.Users, user)
	}
	err = SaveSystemConfig(ctx, constant.SystemConfigUserWhitest, userWhitelist.ToSystemConfig())
	if err != nil {
		return err
	}
	return nil
}

// GetAllowUsers 获取用户白名单
func GetAllowUsers(ctx context.Context) ([]string, error) {
	var userWhitelist dto.UserWhiteList
	err := GetSystemConfigWithEntity(ctx, constant.SystemConfigUserWhitest, &userWhitelist)
	if err != nil {
		return nil, err
	}
	return userWhitelist.Users, nil
}

// SaveSystemConfig 保存系统配置
func SaveSystemConfig(ctx context.Context, key string, config json.RawMessage) error {
	// 判断是否存在
	systemConfig, err := GetSystemConfig(ctx, key)
	if err != nil {
		return err
	}
	s := repo.SystemConfig
	// 不存在则创建
	if len(systemConfig) == 0 {
		return s.WithContext(ctx).Create(&model.SystemConfig{
			Key:   key,
			Value: datatypes.JSON(config),
		})
	}
	_, err = s.WithContext(ctx).Where(s.Key.Eq(key)).Update(s.Value, datatypes.JSON(config))
	if err != nil { // 更新失败
		return err
	}
	return nil
}

// GetSystemConfig 获取系统配置
func GetSystemConfig(ctx context.Context, key string) (json.RawMessage, error) {
	s := repo.SystemConfig
	keyConfig, err := s.WithContext(ctx).Where(s.Key.Eq(key)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return json.RawMessage{}, nil
	}
	if err != nil {
		return json.RawMessage{}, err
	}
	return json.RawMessage(keyConfig.Value), nil
}

// GetSystemConfigWithEntity 获取系统配置
func GetSystemConfigWithEntity(ctx context.Context, key string, entityModel interface{}) error {
	s := repo.SystemConfig
	keyConfig, err := s.WithContext(ctx).Where(s.Key.Eq(key)).First()
	if err != nil {
		return err
	}
	return json.Unmarshal(keyConfig.Value, entityModel)
}
