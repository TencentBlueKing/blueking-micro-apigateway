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

import { Message } from 'bkui-vue';
import { ErrorObject } from 'ajv';
import { useI18n } from 'vue-i18n';

export default function useSchemaErrorMessage() {
  const { t } = useI18n();
  const showSchemaErrorMessages = (errors: ErrorObject[]) => {
    let schemaErrorMsg = '';
    errors.forEach((err) => {
      schemaErrorMsg += `${err.instancePath}: ${err.message}
`;
    });

    Message({
      theme: 'error',
      actions: [
        {
          id: 'assistant',
          disabled: true,
        },
      ],
      message: {
        code: t('Schema 校验失败'),
        overview: t('请正确设置以下字段'),
        suggestion: '',
        details: schemaErrorMsg,
      },
    });
  };

  return {
    showSchemaErrorMessages,
  };
}
