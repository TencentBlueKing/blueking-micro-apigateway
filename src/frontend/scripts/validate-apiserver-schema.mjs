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

import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

import Ajv from 'ajv';
import addFormats from 'ajv-formats';

const failures = [];
let compiledSchemaCount = 0;
let validatedExampleCount = 0;

const validateVersion = async (version) => {
  const schemaDirectory = fileURLToPath(
    new URL(`../../apiserver/pkg/utils/schema/${version}/`, import.meta.url),
  );
  const loadJSON = async filename => JSON.parse(await readFile(`${schemaDirectory}${filename}`, 'utf8'));
  const [officialSchema, bkSchema, tapisixSchema, officialPlugins, bkPlugins, tapisixPlugins] = await Promise.all([
    loadJSON('schema.json'),
    loadJSON('bk_apisix_plugin_schema.json'),
    loadJSON('tapisix_plugin_schema.json'),
    loadJSON('plugin.json'),
    loadJSON('bk_apisix_plugin.json'),
    loadJSON('tapisix_plugin.json'),
  ]);

  const schemaSources = [officialSchema, bkSchema, tapisixSchema];
  const lookup = (path) => {
    for (const source of schemaSources) {
      const value = path.split('.').reduce((node, key) => node?.[key], source);
      if (value !== undefined) return value.plugin_schema ?? value;
    }
    return undefined;
  };

  const ajv = new Ajv();
  addFormats(ajv);
  const compileSchema = (source, plugin, scope, schema, count = true) => {
    try {
      const validate = ajv.compile(schema?.plugin_schema ?? schema);
      if (count) compiledSchemaCount += 1;
      return validate;
    } catch (error) {
      failures.push({ source: `${version}/${source}`, plugin, scope, errors: error.message });
      return undefined;
    }
  };

  const compileSource = (sourceName, source) => {
    for (const [name, schema] of Object.entries(source.main ?? {})) {
      compileSchema(sourceName, name, 'resource', schema);
    }
    for (const [name, plugin] of Object.entries(source.plugins ?? {})) {
      for (const scope of ['schema', 'consumer_schema', 'metadata_schema']) {
        if (plugin[scope] !== undefined) {
          compileSchema(sourceName, name, scope, plugin[scope]);
        }
      }
    }
    for (const [name, plugin] of Object.entries(source.stream_plugins ?? {})) {
      if (plugin.schema !== undefined) {
        compileSchema(sourceName, name, 'stream_schema', plugin.schema);
      }
    }
  };

  compileSource('official-schema', officialSchema);
  compileSource('bk-schema', bkSchema);
  compileSource('tapisix-schema', tapisixSchema);

  const validateExample = (source, plugin, scope, schemaPath, example) => {
    const schema = lookup(schemaPath);
    if (schema === undefined) {
      failures.push({
        source: `${version}/${source}`,
        plugin,
        scope,
        errors: `schema not found: ${schemaPath}`,
      });
      return;
    }

    const validate = compileSchema(source, plugin, `${scope}-example`, schema, false);
    if (!validate) return;

    validatedExampleCount += 1;
    if (!validate(example)) {
      failures.push({
        source: `${version}/${source}`,
        plugin,
        scope,
        errors: ajv.errorsText(validate.errors, { separator: '; ' }),
      });
    }
  };

  for (const [source, catalog] of [
    ['official-catalog', officialPlugins],
    ['bk-catalog', bkPlugins],
    ['tapisix-catalog', tapisixPlugins],
  ]) {
    for (const plugin of catalog) {
      const mainPath = plugin.proxy_type === 'stream'
        ? `stream_plugins.${plugin.name}.schema`
        : `plugins.${plugin.name}.schema`;
      validateExample(source, plugin.name, 'main', mainPath, plugin.example);
      if (plugin.consumer_example !== undefined) {
        validateExample(
          source,
          plugin.name,
          'consumer',
          `plugins.${plugin.name}.consumer_schema`,
          plugin.consumer_example,
        );
      }
      if (plugin.metadata_example !== undefined) {
        validateExample(
          source,
          plugin.name,
          'metadata',
          `plugins.${plugin.name}.metadata_schema`,
          plugin.metadata_example,
        );
      }
    }
  }
};

for (const version of ['3.17', '3.18']) {
  await validateVersion(version);
}

if (failures.length > 0) {
  for (const failure of failures) {
    console.error(`${failure.source}/${failure.plugin}/${failure.scope}: ${failure.errors}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Compiled ${compiledSchemaCount} schema nodes and validated ${validatedExampleCount} examples across 2 APISIX versions.`,
  );
}
