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

import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import { runInNewContext } from 'node:vm';
import ts from 'typescript';

const source = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8');
const compile = text => ts.transpileModule(text, {
  compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS },
}).outputText;

const loadInterceptor = (events, options = {}) => {
  const exports = {};
  const modules = {
    'bkui-vue': { Message: value => events.push(value) },
    '@blueking/login-modal': { showLoginModal: () => events.push('login') },
    '@/store': { useCommon: () => ({ noGlobalError: options.noGlobalError }) },
  };
  runInNewContext(compile(source('http/fetch/error-interceptor.ts')), {
    exports, require: name => modules[name], window: { BK_LOGIN_URL: '/login/' }, console: { log() {}, error() {} },
  });
  return exports.default;
};

for (const [name, error, expected] of [
  ['HTML timeout', { status: 504, error: { message: '<html>504 Gateway Time-out</html>' } }, '504'],
  ['network disconnect', new TypeError('Failed to fetch'), 'Failed to fetch'],
  ['JSON API error', { status: 500, error: { message: '{"error":{"code":"SystemError","message":"database failed"}}' } }, 'database failed'],
  ['plain text error', { status: 502, error: { message: 'Bad Gateway' } }, '502'],
]) {
  test(`interceptor reports ${name} and preserves rejection`, async () => {
    const events = [];
    const handle = loadInterceptor(events);
    await assert.rejects(async () => handle(error, { globalError: true }), value => value === (error.error || error));
    assert.equal(events.length, 1);
    assert.ok(events[0].message.includes(expected));
    assert.ok(!events[0].message.includes('<html>'));
  });
}

test('global-error suppression still rejects without a toast', async () => {
  for (const options of [{ globalError: false }, { globalError: true, noGlobalError: true }]) {
    const events = [];
    const handle = loadInterceptor(events, options);
    const error = new TypeError('Failed to fetch');
    await assert.rejects(async () => handle(error, options), value => value === error);
    assert.deepEqual(events, []);
  }
});

test('Unauthorized responses retain login handling without an error toast', async () => {
  const events = [];
  const error = { status: 401, error: { message: '{"error":{"code":"Unauthorized"}}' } };
  const handle = loadInterceptor(events);
  await assert.rejects(async () => handle(error, { globalError: true }), value => value === error.error);
  assert.deepEqual(events, ['login']);
});

// Execute the production Vue event handler with UI/network mocks.
const page = source('views/gateway-sync-data/gateway-sync-data.vue');
const start = page.indexOf('const confirmBatchSync = async');
const end = page.indexOf('const handleAdd =', start);
assert.ok(start >= 0 && end > start);
const confirmSource = compile(page.slice(start, end));
const runConfirmation = async (request, includeDifferences = true) => {
  const dialogs = [];
  const refreshed = [];
  const batchSync = { value: { visible: true, gatewayId: 6, includeDifferences, preview: { resources: [] } } };
  const isAddLoading = { value: false };
  const dependencies = {
    canConfirmBatchSync: { value: true }, batchSync, isAddLoading, batchDraftCount: { value: 0 },
    batchSyncReadonlyEditor: request, addResourceToEditArea: request,
    // bkui-vue 2.0.1-beta.93 maps subTitle to content in InfoBox.update.
    InfoBox: value => dialogs.push({ ...value, content: value.subTitle || value.content }), t: text => text,
    h: (tag, ...content) => ({ tag, content }), common: { gatewayId: 6 },
    handleResetSelection() {}, getList: () => refreshed.push('list'), fetchSyncedSummary: () => refreshed.push('summary'),
  };
  const confirm = runInNewContext(`${confirmSource}\nconfirmBatchSync;`, dependencies);
  await confirm();
  assert.equal(batchSync.value.visible, false);
  assert.equal(isAddLoading.value, false);
  assert.deepEqual(refreshed, ['list', 'summary']);
  assert.equal(dialogs.length, 1);
  return dialogs[0];
};

for (const includeDifferences of [true, false]) {
  test(`batch rejection closes stale confirmation and explains uncertainty (overwrite=${includeDifferences})`, async () => {
    let calls = 0;
    const dialog = await runConfirmation(async () => {
      calls += 1;
      throw new TypeError('Failed to fetch');
    }, includeDifferences);
    assert.equal(calls, 1);
    assert.equal(dialog.type, 'warning');
    assert.match(JSON.stringify(dialog), /部分操作可能已完成/);
    assert.match(JSON.stringify(dialog), /重新预览/);
  });
}

test('normal and partial results still close the confirmation and show their results', async () => {
  for (const failed of [[], [{ resource_type: 'route', id: 'r1', error: 'stale token' }]]) {
    const dialog = await runConfirmation(async () => ({ added: { route: 1 }, synced: 2, failed }));
    assert.equal(dialog.type, failed.length ? 'warning' : 'success');
    assert.equal(dialog.title, failed.length ? '部分资源同步失败' : '已同步到编辑区');
  }
});

test('batch clients suppress global toast, while existing add callers retain it', async () => {
  const calls = [];
  const exports = {};
  const modules = {
    './fetch': { default: { post: (...args) => calls.push(args) } },
    '@/store': { useCommon: () => ({ gatewayId: 6 }) },
    '@/common/util': {},
  };
  runInNewContext(compile(source('http/gateway-sync-data.ts')), {
    exports, require: name => modules[name], window: { BK_DASHBOARD_URL: '/api' },
  });
  exports.batchSyncReadonlyEditor(6, { resources: [], overwrite_drafts: false });
  exports.addResourceToEditArea({ gatewayId: 6, data: {}, globalError: false });
  exports.addResourceToEditArea({ gatewayId: 6, data: {} });
  assert.deepEqual(calls.map(call => call[2]?.globalError), [false, false, true]);
});


test('batch warning retains actionable API errors without displaying HTML', async () => {
  for (const [message, expected] of [
    ['{"error":{"message":"resource name conflict"}}', 'resource name conflict'],
    ['<html>504 Gateway Time-out</html>', null],
  ]) {
    const dialog = await runConfirmation(async () => {
      throw { message };
    });
    assert.equal(dialog.type, 'warning');
    if (expected) assert.ok(JSON.stringify(dialog.content).includes(expected));
    assert.match(JSON.stringify(dialog.content), /部分操作可能已完成/);
    assert.ok(!JSON.stringify(dialog).includes('<html>'));
  }
});
