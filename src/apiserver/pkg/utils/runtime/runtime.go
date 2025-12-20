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

// Package runtime ...
package runtime

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	sentry "github.com/getsentry/sentry-go"
)

// PanicHandlers ...
var PanicHandlers = []func(any){handlePanic}

// HandlePanic handle panic
func HandlePanic(additionalHandlers ...func(any)) {
	if err := recover(); err != nil {
		for _, fn := range PanicHandlers {
			fn(err)
		}
		for _, fn := range additionalHandlers {
			fn(err)
		}
		panic(err)
	}
}

func handlePanic(r any) {
	if r == http.ErrAbortHandler {
		return
	}
	const size = 32 << 10
	stacktrace := make([]byte, size)
	stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
	var msg string
	if _, ok := r.(string); ok {
		msg = fmt.Sprintf("observed a panic: %s\n%s", r, stacktrace)
	} else {
		msg = fmt.Sprintf("observed a panic: %#v (%v)\n%s", r, r, stacktrace)
	}
	log.Println(msg)
	sentry.CurrentHub().Client().CaptureMessage(msg, nil, sentry.NewScope())
}
