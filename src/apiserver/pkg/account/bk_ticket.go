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

package account

import (
	"fmt"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/config"
	slogresty "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

// BkTicketAuthBackend 用于上云版本的用户登录 & 信息获取
type BkTicketAuthBackend struct{}

// GetLoginUrl 获取登录地址
func (b *BkTicketAuthBackend) GetLoginUrl() string {
	return fmt.Sprintf("%s/plain/", config.G.BkPlatUrlConfig.BkLogin)
}

// GetUserInfo 获取用户信息
func (b *BkTicketAuthBackend) GetUserInfo(token string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/user/get_info/", config.G.BkPlatUrlConfig.BkLogin)

	client := resty.New().SetLogger(slogresty.New()).SetTimeout(10 * time.Second)

	respData := map[string]any{}
	_, err := client.R().
		SetQueryParams(map[string]string{"bk_ticket": token}).
		ForceContentType("application/json").
		SetResult(&respData).
		Get(url)
	if err != nil {
		return nil, err
	}

	if retCode, cErr := cast.ToIntE(respData["ret"]); cErr != nil {
		return nil, errors.Errorf("get user info api %s return code isn't integer", url)
	} else if retCode != 0 {
		return nil, errors.Errorf("failed to get user info from %s, message: %s", url, respData["msg"])
	}

	data, ok := respData["data"].(map[string]any)
	if !ok {
		return nil, errors.Errorf("failed to get user info from %s, response data not json format", url)
	}
	return &UserInfo{ID: data["username"].(string)}, nil
}

var _ AuthBackend = (*BkTicketAuthBackend)(nil)

// NewBkTicketAuthBackend ...
func NewBkTicketAuthBackend() AuthBackend {
	return &BkTicketAuthBackend{}
}
