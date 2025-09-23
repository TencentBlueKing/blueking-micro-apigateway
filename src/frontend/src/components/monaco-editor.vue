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

<template>
  <div :id="id" :style="style" class="monaco-editor"></div>
</template>

<script lang="ts" setup>
import monaco from 'monaco-editor';
import yaml from 'js-yaml';
import { Message } from 'bkui-vue';
import { computed, nextTick, onBeforeUnmount, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import useJsonTransformer from '@/hooks/use-json-transformer';

interface IProps {
  id?: string;
  source?: string;
  language?: string;
  readOnly?: boolean;
  width?: string | number;
  height?: string | number;
  theme?: string;
  minimap?: boolean;
  options?: monaco.editor.IEditorOptions;
}

const {
  id = 'monaco-editor',
  source = '{}',
  language = 'json',
  readOnly = false,
  width = '100%',
  height = 600,
  theme = 'vs-dark',
  // theme = 'vs',
  minimap = false,
  options = {},
} = defineProps<IProps>();

const emit = defineEmits<{
  'change': [{ source: string }]
  'created': [editor: monaco.editor.IStandaloneCodeEditor],
}>();

const { t } = useI18n();
const { formatJSON } = useJsonTransformer();
let editor: monaco.editor.IStandaloneCodeEditor; // 编辑器实例

const style = computed(() => ({
  width: typeof width === 'number' ? `${width}px` : width,
  height: typeof height === 'number' ? `${height}px` : height,
}));

watch(() => source, (newSource, oldSource) => {
  if (newSource === oldSource) {
    return;
  }
  nextTick(() => {
    editor?.setValue(formatJSON({ source }));
  });
}, { immediate: true });

watch(() => readOnly, () => {
  nextTick(() => {
    editor?.updateOptions({ readOnly });
  });
}, { immediate: true });

watch(() => height, () => {
  nextTick(() => {
    editor?.layout();
  });
});

// 设置值
const setValue = (value: string) => {
  editor?.setValue(value);
};

// 获取编辑器中的值
const getValue = () => {
  return editor ? editor.getValue() : '';
};

// 初始化编辑器
const initEditor = () => {
  editor = monaco.editor.create(document.querySelector(`#${id}`), {
    theme, // 主题
    language,
    readOnly, // 是否只读  取值 true | false
    value: source,
    folding: true, // 是否折叠
    foldingHighlight: true, // 折叠等高线
    foldingStrategy: 'indentation', // 折叠方式  auto | indentation
    showFoldingControls: 'always', // 是否一直显示折叠 always | mouseover
    disableLayerHinting: true, // 等宽优化
    emptySelectionClipboard: false, // 空选择剪切板
    selectionClipboard: false, // 选择剪切板
    automaticLayout: true, // 自动布局
    codeLens: false, // 代码镜头
    scrollBeyondLastLine: true, // 滚动完最后一行后再滚动一屏幕
    colorDecorators: true, // 颜色装饰器
    accessibilitySupport: 'off', // 辅助功能支持  "auto" | "off" | "on"
    lineNumbers: 'on', // 行号 取值： "on" | "off" | "relative" | "interval" | function
    lineNumbersMinChars: 5, // 行号最小字符   number
    lineHeight: 24,
    minimap: {
      enabled: minimap, // 小地图
    },
    wordWrap: 'on', // 启用 soft-wraps
    contextmenu: false, // 禁用右键菜单
    stickyScroll: { // 隐藏紧贴在编辑器顶部的上级对象预览，避免当某行内容过长时显示不正常
      enabled: false,
    },
    ...options,
  });

  emit('created', editor);

  // 编辑器初始化后
  editor.onDidChangeModelContent(() => {
    emit('change', { source: getValue() });
  });

  // 定义一个资源导入导出页要用的主题
  monaco.editor.defineTheme('import-theme', {
    base: 'vs-dark',
    inherit: true,
    rules: [],
    colors: {
      'editor.background': '#1a1a1a',
    },
  });
};

const getModel = () => editor.getModel();

const setTheme = (theme: string) => {
  monaco.editor.setTheme(theme);
};

const setLanguage = (language: string) => {
  try {
    monaco.editor.setModelLanguage(editor?.getModel(), language);

    let code = '';

    if (language.toLowerCase() === 'yaml') {
      code = yaml.dump(yaml.load(source));
    } else if (language.toLowerCase() === 'json') {
      code = JSON.stringify(yaml.load(getValue()));
    }

    setValue(code);
    format();
  } catch {
    Message({
      theme: 'error',
      message: t('转换格式失败'),
    });
  }
};

const format = () => {
  editor?.getAction('editor.action.formatDocument')
    .run();
  emit('change', { source: getValue() });
};

const updateOptions = (options: monaco.editor.IEditorOptions) => {
  editor.updateOptions(options);
};

// 挂载
onMounted(() => {
  initEditor();
});

// 卸载
onBeforeUnmount(() => {
  editor?.dispose();
  editor = null;
});

defineExpose({
  getModel,
  getValue,
  setValue,
  setTheme,
  setLanguage,
  format,
  updateOptions,
});

</script>
