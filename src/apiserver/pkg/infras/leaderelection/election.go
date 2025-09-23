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

// Package election ...
package election

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/infras/logging"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/hostx"
)

// EtcdLeaderElector ...
type EtcdLeaderElector struct {
	ctx context.Context

	client     *clientv3.Client
	session    *concurrency.Session
	election   *concurrency.Election
	closeCh    chan struct{}
	leadingCh  chan struct{}
	prefix     string
	instanceID string
	leading    bool
	running    bool
}

// NewEtcdLeaderElector ...
func NewEtcdLeaderElector(client *clientv3.Client, prefix string) (*EtcdLeaderElector, error) {
	return &EtcdLeaderElector{
		client: client,
		prefix: prefix + "-leader-election",
		instanceID: fmt.Sprintf(
			"%s_%s",
			hostx.GetHostname(),
			hostx.GetLocalIpV4(),
		),
		leading: false,
		running: false,
	}, nil
}

// Run ...
func (ele *EtcdLeaderElector) Run(ctx context.Context) {
	if ele.running {
		return
	}
	ele.ctx = ctx
	ele.running = true
	ele.initElection()
	go ele.run()
}

func (ele *EtcdLeaderElector) initElection() {
	for {
		session, err := concurrency.NewSession(ele.client)
		if err != nil {
			logging.Error(err, "Create election session failed")
			time.Sleep(time.Second * 5)
			continue
		}
		ele.session = session
		break
	}
	ele.election = concurrency.NewElection(ele.session, ele.prefix)
	ele.closeCh = make(chan struct{})
	ele.leadingCh = make(chan struct{})
}

func (ele *EtcdLeaderElector) run() {
	ele.elect()
	ele.checkLeadership()
}

func (ele *EtcdLeaderElector) elect() {
	for {
		logging.Info("try to be leader", "id", ele.instanceID)
		err := ele.election.Campaign(ele.ctx, ele.instanceID)
		if err != nil {
			logging.Error(err, "leader election campaign returns error", "id", ele.instanceID)
			time.Sleep(time.Second * 5)
			continue
		}
		logging.Info("become leader now", "id", ele.instanceID)
		ele.leading = true

		close(ele.leadingCh)
		return
	}
}

// IsLeader ...
func (ele *EtcdLeaderElector) IsLeader() bool {
	return ele.leading
}

func (ele *EtcdLeaderElector) checkLeadership() {
	for {
		select {
		case <-ele.session.Done():
			close(ele.closeCh)
			ele.leading = false
			ele.initElection()
			go ele.run()
			return
		case <-ele.ctx.Done():
			ele.session.Close()
			close(ele.closeCh)
			ele.leading = false
			ele.running = false
			return
		}
	}
}

// Leader ...
func (ele *EtcdLeaderElector) Leader() string {
	if ele.election == nil {
		return ""
	}
	resp, err := ele.election.Leader(ele.ctx)
	if err != nil {
		logging.Error(err, "get Leader info failed")
		return ""
	}
	if resp.Count == 0 {
		return ""
	}
	return string(resp.Kvs[0].Value)
}

// WaitForLeading ...
func (ele *EtcdLeaderElector) WaitForLeading() (closeCh <-chan struct{}) {
	if ele.leading {
		logging.Info("success get leader")
		return ele.closeCh
	}
	<-ele.leadingCh
	return ele.closeCh
}
