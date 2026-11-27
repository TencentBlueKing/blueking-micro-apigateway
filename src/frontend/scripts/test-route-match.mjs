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
import Ajv from 'ajv';
import addFormats from 'ajv-formats';
import { uniq } from 'lodash-es';
import ts from 'typescript';

const readFrontendSource = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8');

const compileTs = text => ts.transpileModule(text, {
  compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS },
}).outputText;

const loadRouteMatchUtils = () => {
  const exports = {};
  const modules = {
    'lodash-es': { uniq },
  };
  runInNewContext(compileTs(readFrontendSource('views/route/route-match-utils.ts')), {
    exports,
    require: name => modules[name],
  });
  return {
    // 将 vm realm 对象转换为宿主 realm，便于 node:assert/strict 做 deepEqual 断言
    normalizeRouteMatchFields: raw => structuredClone(exports.normalizeRouteMatchFields(raw)),
    applyRouteMatchFieldsForSubmit: (config, values) => {
      return structuredClone(exports.applyRouteMatchFieldsForSubmit(config, values));
    },
  };
};

// 加载 APISIX 路由 Schema，验证生成的配置满足真实后端校验规则
const loadRouteSchemaValidator = (version = '3.11') => {
  const schemaPath = new URL(`../../apiserver/pkg/utils/schema/${version}/schema.json`, import.meta.url);
  const schemaJson = JSON.parse(readFileSync(schemaPath, 'utf8'));
  const ajv = new Ajv({ strict: false });
  addFormats(ajv);
  return ajv.compile(schemaJson.main.route);
};

const {
  normalizeRouteMatchFields,
  applyRouteMatchFieldsForSubmit,
} = loadRouteMatchUtils();

const validateRoute311 = loadRouteSchemaValidator('3.11');
const validateRoute313 = loadRouteSchemaValidator('3.13');

const assertValidRouteSchema = (config) => {
  for (const [version, validator] of [['3.11', validateRoute311], ['3.13', validateRoute313]]) {
    const valid = validator(config);
    assert.equal(
      valid,
      true,
      `APISIX ${version} schema 校验失败: ${JSON.stringify(validator.errors)}，config=${JSON.stringify(config)}`,
    );
  }
};

test('单值路由配置入口归一化为数组并清除旧单值键', () => {
  const rawConfig = {
    uri: '/*',
    host: 'ag.devx.tencent.com',
    remote_addr: '192.168.1.101',
    upstream_id: '474945920484588502',
    methods: ['GET', 'POST'],
  };

  const { config, conflicts } = normalizeRouteMatchFields(rawConfig);

  assert.deepEqual(conflicts, {});
  assert.deepEqual(config.uris, ['/*']);
  assert.deepEqual(config.hosts, ['ag.devx.tencent.com']);
  assert.deepEqual(config.remote_addrs, ['192.168.1.101']);
  assert.equal('uri' in config, false);
  assert.equal('host' in config, false);
  assert.equal('remote_addr' in config, false);

  // 直接保存时每组只保留一种字段，且可通过 3.11 / 3.13 schema 校验
  const submitConfig = applyRouteMatchFieldsForSubmit(config, {
    uris: config.uris,
    hosts: config.hosts,
    remoteAddrs: config.remote_addrs,
  });
  assert.equal('uri' in submitConfig, false);
  assert.equal('host' in submitConfig, false);
  assert.equal('remote_addr' in submitConfig, false);
  assert.deepEqual(submitConfig.uris, ['/*']);
  assert.deepEqual(submitConfig.hosts, ['ag.devx.tencent.com']);
  assert.deepEqual(submitConfig.remote_addrs, ['192.168.1.101']);
  assertValidRouteSchema(submitConfig);
});

test('纯数组配置多条路径、域名和客户端地址完整回显与保存', () => {
  const rawConfig = {
    uris: ['/api/v1/*', '/api/v2/*'],
    hosts: ['a.devx.tencent.com', '*.b.devx.tencent.com'],
    remote_addrs: ['10.0.0.1', '10.0.0.0/24'],
    upstream_id: '474945920484588502',
  };

  const { config, conflicts } = normalizeRouteMatchFields(rawConfig);
  assert.deepEqual(conflicts, {});
  assert.deepEqual(config.uris, ['/api/v1/*', '/api/v2/*']);
  assert.deepEqual(config.hosts, ['a.devx.tencent.com', '*.b.devx.tencent.com']);
  assert.deepEqual(config.remote_addrs, ['10.0.0.1', '10.0.0.0/24']);

  const submitConfig = applyRouteMatchFieldsForSubmit(config, {
    uris: config.uris,
    hosts: config.hosts,
    remoteAddrs: config.remote_addrs,
  });
  assertValidRouteSchema(submitConfig);
});

test('无域名和无客户端地址限制的通配路径路由正常归一化与保存', () => {
  const rawConfig = {
    uri: '/*',
    upstream_id: '474945920484588502',
  };

  const { config, conflicts } = normalizeRouteMatchFields(rawConfig);
  assert.deepEqual(conflicts, {});
  assert.deepEqual(config.uris, ['/*']);
  assert.deepEqual(config.hosts, []);
  assert.deepEqual(config.remote_addrs, []);

  const submitConfig = applyRouteMatchFieldsForSubmit(config, {
    uris: config.uris,
    hosts: [],
    remoteAddrs: [],
  });
  assert.deepEqual(submitConfig.uris, ['/*']);
  assert.equal('host' in submitConfig, false);
  assert.equal('hosts' in submitConfig, false);
  assert.equal('remote_addr' in submitConfig, false);
  assert.equal('remote_addrs' in submitConfig, false);
  assertValidRouteSchema(submitConfig);
});

test('清空域名和客户端地址后，即使历史状态残留单值键，提交体仍彻底省略对应字段', () => {
  const dirtyConfig = {
    uri: '/*',
    uris: ['/*'],
    host: 'ag.devx.tencent.com',
    hosts: ['ag.devx.tencent.com'],
    remote_addr: '127.0.0.1',
    remote_addrs: ['127.0.0.1'],
    upstream_id: '474945920484588502',
  };

  const submitConfig = applyRouteMatchFieldsForSubmit(dirtyConfig, {
    uris: ['/*'],
    hosts: [],
    remoteAddrs: [],
  });

  assert.equal('uri' in submitConfig, false);
  assert.equal('host' in submitConfig, false);
  assert.equal('hosts' in submitConfig, false);
  assert.equal('remote_addr' in submitConfig, false);
  assert.equal('remote_addrs' in submitConfig, false);
  assert.deepEqual(submitConfig.uris, ['/*']);
  assertValidRouteSchema(submitConfig);
});

test('单值与数组同时存在且语义完全一致时不报冲突并归一为数组', () => {
  const { config, conflicts } = normalizeRouteMatchFields({
    uri: '/*',
    uris: ['/*'],
    host: 'ag.devx.tencent.com',
    hosts: ['ag.devx.tencent.com', 'ag.devx.tencent.com'],
    remote_addr: '127.0.0.1',
    remote_addrs: ['127.0.0.1'],
  });

  assert.deepEqual(conflicts, {});
  assert.deepEqual(config.uris, ['/*']);
  assert.deepEqual(config.hosts, ['ag.devx.tencent.com']);
  assert.deepEqual(config.remote_addrs, ['127.0.0.1']);
  assert.equal('uri' in config, false);
  assert.equal('host' in config, false);
  assert.equal('remote_addr' in config, false);
});

test('单值与数组同时存在且内容冲突时，不静默合并而是返回冲突详情', () => {
  const { config, conflicts } = normalizeRouteMatchFields({
    uri: '/only-single',
    uris: ['/from-array-1', '/from-array-2'],
    host: 'a.devx.tencent.com',
    hosts: ['a.devx.tencent.com', 'b.devx.tencent.com'],
    remote_addr: '10.0.0.1',
    remote_addrs: ['10.0.0.2'],
  });

  // 不应静默合并单值到数组中
  assert.deepEqual(config.uris, ['/from-array-1', '/from-array-2']);
  assert.deepEqual(config.hosts, ['a.devx.tencent.com', 'b.devx.tencent.com']);
  assert.deepEqual(config.remote_addrs, ['10.0.0.2']);
  assert.equal('uri' in config, false);
  assert.equal('host' in config, false);
  assert.equal('remote_addr' in config, false);

  assert.deepEqual(conflicts.uris, {
    key: 'uris',
    singleKey: 'uri',
    arrayKey: 'uris',
    singleValue: '/only-single',
    arrayValues: ['/from-array-1', '/from-array-2'],
  });
  assert.deepEqual(conflicts.hosts, {
    key: 'hosts',
    singleKey: 'host',
    arrayKey: 'hosts',
    singleValue: 'a.devx.tencent.com',
    arrayValues: ['a.devx.tencent.com', 'b.devx.tencent.com'],
  });
  assert.deepEqual(conflicts.remote_addrs, {
    key: 'remote_addrs',
    singleKey: 'remote_addr',
    arrayKey: 'remote_addrs',
    singleValue: '10.0.0.1',
    arrayValues: ['10.0.0.2'],
  });
});

test('路由与服务高级配置均使用 v-show 避免折叠卸载丢失最新输入，且 form-uris 校验失败时不再吞异常', () => {
  const routeCreateSource = readFrontendSource('views/route/create.vue');
  const serviceCreateSource = readFrontendSource('views/service/create.vue');
  const formUrisSource = readFrontendSource('components/form/form-uris.vue');

  assert.equal(routeCreateSource.includes('v-if="uiConfig.showAdvanced"'), false);
  assert.equal(serviceCreateSource.includes('v-if="uiConfig.showAdvanced"'), false);
  assert.match(routeCreateSource, /v-show="uiConfig\.showAdvanced"[\s\S]*form-hosts-new/);
  assert.match(routeCreateSource, /v-show="uiConfig\.showAdvanced"[\s\S]*form-remote-addrs-new/);
  assert.match(serviceCreateSource, /v-show="uiConfig\.showAdvanced"[\s\S]*form-hosts-new/);
  assert.equal(formUrisSource.includes('return false;'), false);
});

test('form-uris 校验逻辑在路径为空时抛出异常阻断提交，在路径非空时校验通过', async () => {
  const formUrisSource = readFrontendSource('components/form/form-uris.vue');
  const start = formUrisSource.indexOf('const validate = async () =>');
  const end = formUrisSource.indexOf('defineExpose(', start);
  assert.ok(start >= 0 && end > start);

  const runValidate = (urisValue, formRefsValue) => runInNewContext(
    `${compileTs(formUrisSource.slice(start, end))}\nvalidate();`,
    {
      uris: { value: urisValue },
      formRefs: { value: formRefsValue },
      t: text => text,
    },
  );

  // 空数组阻断
  await assert.rejects(() => runValidate([], []), /必填项/);
  // 单行空值触发子表单校验 reject 时不再吞掉异常
  await assert.rejects(
    () => runValidate([''], [{ validate: () => Promise.reject(new Error('必填项')) }]),
    /必填项/,
  );
  // 有效路径通过
  const ok = await runValidate(['/*'], [{ validate: () => Promise.resolve(true) }]);
  assert.equal(ok, true);
});
