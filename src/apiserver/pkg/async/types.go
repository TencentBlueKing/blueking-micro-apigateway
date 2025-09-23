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

package async

import (
	"sync"

	cron "github.com/robfig/cron/v3"
)

// 任务 ID 与 cron.EntryID 的映射表
type taskEntryMap struct {
	mapping map[int64]cron.EntryID
	sync.RWMutex
}

func (m *taskEntryMap) get(taskId int64) (cron.EntryID, bool) {
	m.RLock()
	defer m.RUnlock()
	entryID, ok := m.mapping[taskId]
	return entryID, ok
}

func (m *taskEntryMap) set(taskId int64, entryID cron.EntryID) {
	m.Lock()
	defer m.Unlock()
	m.mapping[taskId] = entryID
}

func (m *taskEntryMap) delete(taskId int64) {
	m.Lock()
	defer m.Unlock()
	delete(m.mapping, taskId)
}
