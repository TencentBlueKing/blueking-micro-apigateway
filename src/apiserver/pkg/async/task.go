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

// Package async 提供一个简单的异步 / 定时任务封装：
// 1. 使用 cron 支持定时任务（cmd: scheduler）
// 2. 简单封装 goroutine 以支持异步任务
package async

import (
	"reflect"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/async/task"
	log "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
)

// RegisteredTasks 已注册的任务
var RegisteredTasks = map[string]any{
	"CalcFib": task.CalcFib,
	// TODO: SaaS 开发者可根据需求添加自定义任务
}

// ApplyTask 下发异步任务
func ApplyTask(name string, args []any) {
	go func() {
		taskFunc, ok := RegisteredTasks[name]
		if !ok {
			log.Errorf("task func %s not found", name)
			return
		}

		taskArgs := []reflect.Value{}
		for _, arg := range args {
			taskArgs = append(taskArgs, reflect.ValueOf(arg))
		}
		reflect.ValueOf(taskFunc).Call(taskArgs)
	}()
}
