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

// Package idx ...
package idx

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sony/sonyflake"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
)

// 资源ID正则
var resourceIDRegex = regexp.MustCompile(`^bk\.([^.]+)\.`)

var resourceIDResourceTypePrefixMap = map[constant.APISIXResource]string{
	constant.Route:          "r",
	constant.Upstream:       "u",
	constant.Service:        "s",
	constant.Consumer:       "c",
	constant.ConsumerGroup:  "cg",
	constant.GlobalRule:     "gr",
	constant.PluginConfig:   "pc",
	constant.PluginMetadata: "pm",
	constant.Proto:          "pb",
	constant.SSL:            "ss",
	constant.StreamRoute:    "sr",
}

var resourcePrefixResourceTypeMap = map[string]constant.APISIXResource{
	"r":  constant.Route,
	"u":  constant.Upstream,
	"s":  constant.Service,
	"c":  constant.Consumer,
	"cg": constant.ConsumerGroup,
	"gr": constant.GlobalRule,
	"pc": constant.PluginConfig,
	"pm": constant.PluginMetadata,
	"pb": constant.Proto,
	"ss": constant.SSL,
	"sr": constant.StreamRoute,
}

var _sf *sonyflake.Sonyflake

func init() {
	saltStr, ok := os.LookupEnv("FLAKE_SALT")
	var salt uint16
	if ok {
		i, err := strconv.ParseUint(saltStr, 10, 16)
		if err != nil {
			panic(err)
		}
		salt = uint16(i)
	}
	ips, err := getLocalIPs()
	if err != nil {
		panic(err)
	}
	_sf = sonyflake.NewSonyflake(sonyflake.Settings{
		MachineID: func() (u uint16, e error) {
			return sumIPs(ips) + salt, nil
		},
	})
	if _sf == nil {
		panic("sonyflake init failed")
	}
}

func sumIPs(ips []net.IP) uint16 {
	total := 0
	for _, ip := range ips {
		for i := range ip {
			total += int(ip[i])
		}
	}
	return uint16(total)
}

func getLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}
	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ips = append(ips, ipNet.IP)
		}
	}
	return ips, nil
}

const base64Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._"

// uint64ToBase64
func uint64ToBase64(num uint64) string {
	if num == 0 {
		return string(base64Charset[0])
	}
	var sb strings.Builder
	for num > 0 {
		remainder := num % 64
		sb.WriteByte(base64Charset[remainder])
		num /= 64
	}
	result := sb.String()
	runes := []rune(result)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// GenResourceID ...
func GenResourceID(resourceType constant.APISIXResource) string {
	uid, err := _sf.NextID()
	if err != nil {
		panic("get sony flake uid failed:" + err.Error())
	}
	prefix := resourceIDResourceTypePrefixMap[resourceType]
	return fmt.Sprintf("bk.%s.%s", prefix, uint64ToBase64(uid))
}

// GetResourceTypeFromID ...
func GetResourceTypeFromID(id string) constant.APISIXResource {
	// 正则表达式匹配前缀部分
	matches := resourceIDRegex.FindStringSubmatch(id)
	if len(matches) != 2 {
		return ""
	}
	return resourcePrefixResourceTypeMap[matches[1]]
}
