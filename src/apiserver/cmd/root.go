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

// Package cmd ...
package cmd

import (
	"github.com/spf13/cobra"

	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

var rootCmd = &cobra.Command{
	Use:   "apiserver",
	Short: "Golang Gin Template V2",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("welcome to use apiserver, use `apiserver -h` for help")
	},
}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}
