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
import { cloneDeep, isEqual, uniq } from 'lodash-es';
import ts from 'typescript';

const readFrontendSource = path => readFileSync(new URL(`../src/${path}`, import.meta.url), 'utf8');

const compileTs = text => ts.transpileModule(text, {
  compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS },
}).outputText;

const loadRouteMatchUtils = () => {
  const exports = {};
  const modules = {
    'lodash-es': { isEqual, uniq },
  };
  runInNewContext(compileTs(readFrontendSource('views/route/route-match-utils.ts')), {
    exports,
    require: name => modules[name],
  });
  return {
    // 将 vm realm 对象转换为宿主 realm，便于 node:assert/strict 做 deepEqual 断言
    normalizeRouteMatchFields: raw => structuredClone(exports.normalizeRouteMatchFields(raw)),
    shouldClearMatchConflict: (conflict, values) => exports.shouldClearMatchConflict(conflict, values),
    getUnresolvedMatchConflicts: (initial, confirmed, current) => {
      return structuredClone(exports.getUnresolvedMatchConflicts(initial, confirmed, current));
    },
    applyRouteMatchFieldsForSubmit: (config, values) => {
      return structuredClone(exports.applyRouteMatchFieldsForSubmit(config, values));
    },
  };
};

const loadSSLConfigUtils = () => {
  const exports = {};
  const modules = {
    'lodash-es': { cloneDeep, uniq },
  };
  runInNewContext(compileTs(readFrontendSource('views/ssl/ssl-config-utils.ts')), {
    exports,
    require: name => modules[name],
  });
  return {
    extractSSLSnis: config => structuredClone(exports.extractSSLSnis(config)),
    buildSSLSubmitConfig: (rawConfig, params) => structuredClone(exports.buildSSLSubmitConfig(rawConfig, params)),
  };
};

// 加载 APISIX Schema，验证生成的配置满足真实后端校验规则
const loadSchemaValidator = (resource, version = '3.11') => {
  const schemaPath = new URL(`../../apiserver/pkg/utils/schema/${version}/schema.json`, import.meta.url);
  const schemaJson = JSON.parse(readFileSync(schemaPath, 'utf8'));
  const ajv = new Ajv({ strict: false });
  addFormats(ajv);
  return ajv.compile(schemaJson.main[resource]);
};

const {
  normalizeRouteMatchFields,
  shouldClearMatchConflict,
  getUnresolvedMatchConflicts,
  applyRouteMatchFieldsForSubmit,
} = loadRouteMatchUtils();

const {
  extractSSLSnis,
  buildSSLSubmitConfig,
} = loadSSLConfigUtils();

const validateRoute311 = loadSchemaValidator('route', '3.11');
const validateRoute313 = loadSchemaValidator('route', '3.13');
const validateSSL311 = loadSchemaValidator('ssl', '3.11');
const validateSSL313 = loadSchemaValidator('ssl', '3.13');

const assertValidRouteSchema = (config) => {
  for (const [version, validator] of [['3.11', validateRoute311], ['3.13', validateRoute313]]) {
    const valid = validator(config);
    assert.equal(
      valid,
      true,
      `APISIX ${version} route schema 校验失败: ${JSON.stringify(validator.errors)}，config=${JSON.stringify(config)}`,
    );
  }
};

const assertValidSSLSchema = (config) => {
  for (const [version, validator] of [['3.11', validateSSL311], ['3.13', validateSSL313]]) {
    const valid = validator(config);
    assert.equal(
      valid,
      true,
      `APISIX ${version} ssl schema 校验失败: ${JSON.stringify(validator.errors)}，config=${JSON.stringify(config)}`,
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

test('冲突态下新增再删除空行或重复行无法绕过冲突确认，仅有效值实质变化或显式选择才解除冲突', () => {
  const { conflicts } = normalizeRouteMatchFields({
    uri: '/one',
    uris: ['/two'],
    host: 'one.example.com',
    hosts: ['two.example.com'],
    remote_addr: '10.0.0.1',
    remote_addrs: ['10.0.0.2'],
  });

  // 路径冲突：新增空行、删回原值、添加重复行、清空必填路径均不可绕过冲突确认
  assert.equal(shouldClearMatchConflict(conflicts.uris, ['/two', '']), false);
  assert.equal(shouldClearMatchConflict(conflicts.uris, ['/two']), false);
  assert.equal(shouldClearMatchConflict(conflicts.uris, ['/two', '/two']), false);
  assert.equal(shouldClearMatchConflict(conflicts.uris, ['']), false);
  assert.equal(shouldClearMatchConflict(conflicts.uris, []), false);
  // 有效值实质修改（改为单值、改为新值、或新增不同路径）才解除冲突
  assert.equal(shouldClearMatchConflict(conflicts.uris, ['/one']), true);
  assert.equal(shouldClearMatchConflict(conflicts.uris, ['/three']), true);
  assert.equal(shouldClearMatchConflict(conflicts.uris, ['/two', '/three']), true);

  // 域名与客户端地址冲突：增删空行不绕过，清空可选限制或改为不同有效值则解除冲突
  assert.equal(shouldClearMatchConflict(conflicts.hosts, ['two.example.com', '']), false);
  assert.equal(shouldClearMatchConflict(conflicts.hosts, ['two.example.com']), false);
  assert.equal(shouldClearMatchConflict(conflicts.hosts, ['']), true);
  assert.equal(shouldClearMatchConflict(conflicts.hosts, ['one.example.com']), true);

  assert.equal(shouldClearMatchConflict(conflicts.remote_addrs, ['10.0.0.2', '']), false);
  assert.equal(shouldClearMatchConflict(conflicts.remote_addrs, ['10.0.0.2']), false);
  assert.equal(shouldClearMatchConflict(conflicts.remote_addrs, ['']), true);
  assert.equal(shouldClearMatchConflict(conflicts.remote_addrs, ['10.0.0.1']), true);

  // 模拟生产事件流：增删空行或把 /two 改为 /tw 再改回 /two，均无法绕过冲突确认
  const formUrisSource = readFrontendSource('components/form/form-uris.vue');
  const routeCreateSource = readFrontendSource('views/route/create.vue');
  const startUris = formUrisSource.indexOf('const emitChange =');
  const endUris = formUrisSource.indexOf('const validate =', startUris);
  const startRoute = routeCreateSource.indexOf('const handleMatchFieldChange =');
  const endRoute = routeCreateSource.indexOf('const handleHosts =', startRoute);
  assert.ok(startUris >= 0 && endUris > startUris);
  assert.ok(startRoute >= 0 && endRoute > startRoute);

  const initialMatchConflicts = { value: structuredClone(conflicts) };
  const confirmedMatchConflicts = { value: [] };
  const currentMatchValues = {
    value: {
      uris: ['/two'],
      hosts: ['two.example.com'],
      remote_addrs: ['10.0.0.2'],
    },
  };
  const routeConfig = { value: { uris: ['/two'] } };
  const routeHandlers = runInNewContext(
    `${compileTs(routeCreateSource.slice(startRoute, endRoute))}\n({ handleMatchFieldChange, resolveMatchConflict });`,
    {
      initialMatchConflicts,
      confirmedMatchConflicts,
      currentMatchValues,
      routeConfig,
    },
  );
  const urisRef = { value: ['/two'] };
  const urisHandlers = runInNewContext(
    `${compileTs(formUrisSource.slice(startUris, endUris))}\n({ handleAddItem, handleRemoveItem, handleInput });`,
    {
      uris: urisRef,
      emit: (event, values) => {
        if (event === 'change') {
          routeHandlers.handleMatchFieldChange('uris', values);
        }
      },
    },
  );
  const getUrisConflict = () => getUnresolvedMatchConflicts(
    initialMatchConflicts.value,
    confirmedMatchConflicts.value,
    currentMatchValues.value,
  ).find(item => item.key === 'uris');

  // 1) 点击 + 新增空行，再点击 - 删除空行
  urisHandlers.handleAddItem();
  assert.deepEqual(urisRef.value, ['/two', '']);
  assert.ok(getUrisConflict(), '新增空行不应清除冲突');
  urisHandlers.handleRemoveItem(1);
  assert.deepEqual(urisRef.value, ['/two']);
  assert.ok(getUrisConflict(), '删除空行后有效值未变化，不应清除冲突');

  // 2) 把 /two 改为 /tw 再改回 /two，冲突应自动恢复且不可绕过
  urisHandlers.handleInput(0, '/tw');
  assert.equal(getUrisConflict(), undefined, '改为其他有效路径时当前不拦截');
  urisHandlers.handleInput(0, '/two');
  assert.ok(getUrisConflict(), '改回 /two 后冲突应自动恢复，不可绕过确认');

  // 3) 显式点击“采用 uris”后才永久解除冲突，即使保持 /two 也不再拦截
  routeHandlers.resolveMatchConflict('uris', 'array');
  assert.equal(getUrisConflict(), undefined);
  assert.deepEqual(structuredClone(routeConfig.value.uris), ['/two']);

  // 4) 域名/客户端地址同样支持改回原值后恢复冲突
  const hostsOnlyUnresolved = getUnresolvedMatchConflicts(
    { hosts: conflicts.hosts },
    [],
    { hosts: ['two.example.com'] },
  );
  assert.equal(hostsOnlyUnresolved.length, 1);
  assert.equal(getUnresolvedMatchConflicts({ hosts: conflicts.hosts }, [], { hosts: ['two.example.co'] }).length, 0);
  assert.equal(getUnresolvedMatchConflicts({ hosts: conflicts.hosts }, [], { hosts: ['two.example.com'] }).length, 1);
});

test('路由、服务与 Upstream 高级配置均使用 v-show 避免折叠卸载丢失输入，且 form-uris 校验失败时不再吞异常', () => {
  const routeCreateSource = readFrontendSource('views/route/create.vue');
  const serviceCreateSource = readFrontendSource('views/service/create.vue');
  const upstreamFormSource = readFrontendSource('components/form/form-upstream.vue');
  const formUrisSource = readFrontendSource('components/form/form-uris.vue');

  assert.equal(routeCreateSource.includes('v-if="uiConfig.showAdvanced"'), false);
  assert.equal(serviceCreateSource.includes('v-if="uiConfig.showAdvanced"'), false);
  assert.equal(upstreamFormSource.includes('v-if="uiConfig.showAdvanced"'), false);
  assert.match(routeCreateSource, /v-show="uiConfig\.showAdvanced"[\s\S]*form-hosts-new/);
  assert.match(routeCreateSource, /v-show="uiConfig\.showAdvanced"[\s\S]*form-remote-addrs-new/);
  assert.match(serviceCreateSource, /v-show="uiConfig\.showAdvanced"[\s\S]*form-hosts-new/);
  assert.match(upstreamFormSource, /v-show="uiConfig\.showAdvanced"[\s\S]*form-health-checks/);
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

  await assert.rejects(() => runValidate([], []), /必填项/);
  await assert.rejects(
    () => runValidate([''], [{ validate: () => Promise.reject(new Error('必填项')) }]),
    /必填项/,
  );
  const ok = await runValidate(['/*'], [{ validate: () => Promise.resolve(true) }]);
  assert.equal(ok, true);
});

test('Upstream getHealthChecks 正确保留修改并支持关闭全部健康检查清除旧配置', () => {
  const upstreamFormSource = readFrontendSource('components/form/form-upstream.vue');
  const start = upstreamFormSource.indexOf('const getHealthChecks = () =>');
  const end = upstreamFormSource.indexOf('const validate = async () =>', start);
  assert.ok(start >= 0 && end > start);

  const runGetHealthChecks = healthCheckFormRefValue => structuredClone(runInNewContext(
    `${compileTs(upstreamFormSource.slice(start, end))}\ngetHealthChecks();`,
    { healthCheckFormRef: { value: healthCheckFormRefValue } },
  ));

  // 未挂载返回 undefined
  assert.equal(runGetHealthChecks(null), undefined);

  // 修改探测路径后（即使折叠，v-show 仍保持挂载）返回最新值
  const modifiedChecks = { active: { type: 'http', http_path: '/new' } };
  assert.deepEqual(runGetHealthChecks({ getValue: () => modifiedChecks }), modifiedChecks);

  // 关闭全部健康检查时返回空对象 {}（而非 null），允许调用方覆盖并删除旧 checks
  const disabledChecks = runGetHealthChecks({ getValue: () => ({}) });
  assert.deepEqual(disabledChecks, {});

  // 确认 4 个调用方均通过 `checks !== undefined` 写入 upstreamCopy.checks，且空对象会删除 checks 键
  for (const pagePath of [
    'views/upstream/create.vue',
    'views/route/create.vue',
    'views/service/create.vue',
    'views/stream-route/create.vue',
  ]) {
    const pageSource = readFrontendSource(pagePath);
    assert.match(pageSource, /if \(checks !== undefined\)\s*\{\s*upstreamCopy\.checks = checks;/, pagePath);
  }
});

test('SSL 表单提交保留附加证书 certs/keys 及其他未编辑字段，并清理单值 sni 与旧有效期', () => {
  const primaryCert = 'A'.repeat(140);
  const primaryKey = 'B'.repeat(140);
  const extraCert = 'C'.repeat(140);
  const extraKey = 'D'.repeat(140);
  const clientCa = 'E'.repeat(140);

  const rawConfig = {
    id: 'ssl-1',
    name: 'old-ssl-name',
    cert: primaryCert,
    key: primaryKey,
    certs: [extraCert],
    keys: [extraKey],
    sni: 'legacy.example.com',
    type: 'server',
    status: 1,
    ssl_protocols: ['TLSv1.2', 'TLSv1.3'],
    client: { ca: clientCa, depth: 2 },
    labels: { env: 'prod' },
    validity_start: 1700000000,
    validity_end: 1730000000,
  };

  // 单值 sni 也能正确回显为数组
  assert.deepEqual(extractSSLSnis(rawConfig), ['legacy.example.com']);

  // 仅修改名称后提交，附加证书 certs/keys、client、ssl_protocols、labels 等完整保留
  const submitConfig = buildSSLSubmitConfig(rawConfig, {
    cert: primaryCert,
    key: primaryKey,
    snis: ['legacy.example.com'],
    fallbackSnis: ['cert-san.example.com'],
  });

  assert.deepEqual(submitConfig.certs, [extraCert]);
  assert.deepEqual(submitConfig.keys, [extraKey]);
  assert.deepEqual(submitConfig.client, { ca: clientCa, depth: 2 });
  assert.deepEqual(submitConfig.ssl_protocols, ['TLSv1.2', 'TLSv1.3']);
  assert.deepEqual(submitConfig.labels, { env: 'prod' });
  assert.deepEqual(submitConfig.snis, ['legacy.example.com']);
  assert.equal('sni' in submitConfig, false);
  assert.equal('validity_start' in submitConfig, false);
  assert.equal('validity_end' in submitConfig, false);
  assert.equal('name' in submitConfig, false);
  assertValidSSLSchema(submitConfig);

  // 新建 SSL（无 rawConfig）且未手填 snis 时回退到证书解析的 fallbackSnis
  const createConfig = buildSSLSubmitConfig(null, {
    cert: primaryCert,
    key: primaryKey,
    snis: [],
    fallbackSnis: ['auto.example.com'],
  });
  assert.deepEqual(createConfig, {
    cert: primaryCert,
    key: primaryKey,
    snis: ['auto.example.com'],
  });
  assertValidSSLSchema(createConfig);
});
