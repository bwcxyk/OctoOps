<template>
  <div ref="containerRef" class="code-editor" :style="{ height: `${height}px` }" />
</template>

<script setup lang="ts">
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api';
import 'monaco-editor/esm/vs/language/json/monaco.contribution';
import 'monaco-editor/min/vs/editor/editor.main.css';
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue';

import { useSettingStore } from '@/store/modules/setting';

type EditorLanguage = 'json' | 'hocon' | 'plaintext';

const props = withDefaults(
  defineProps<{
    modelValue: string;
    language?: EditorLanguage;
    height?: number;
    readonly?: boolean;
    active?: boolean;
  }>(),
  {
    language: 'plaintext',
    height: 360,
    readonly: false,
    active: true,
  },
);

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void;
}>();

const HOCON_LANGUAGE_ID = 'hocon';
let monacoConfigured = false;
let hoconRegistered = false;

function ensureMonacoEnvironment() {
  if (monacoConfigured)
    return;

  (globalThis as typeof globalThis & {
    MonacoEnvironment?: {
      getWorker: (_moduleId: string, label: string) => Worker;
    };
  }).MonacoEnvironment = {
    getWorker(_moduleId: string, label: string) {
      if (label === 'json')
        return new jsonWorker();
      return new editorWorker();
    },
  };

  monacoConfigured = true;
}

function ensureHoconLanguage() {
  if (hoconRegistered)
    return;

  monaco.languages.register({ id: HOCON_LANGUAGE_ID });
  monaco.languages.setLanguageConfiguration(HOCON_LANGUAGE_ID, {
    comments: {
      lineComment: '#',
      blockComment: ['/*', '*/'],
    },
    brackets: [
      ['{', '}'],
      ['[', ']'],
      ['(', ')'],
    ],
    autoClosingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
    ],
    surroundingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
    ],
  });
  monaco.languages.setMonarchTokensProvider(HOCON_LANGUAGE_ID, {
    tokenizer: {
      root: [
        [/[{}[\]()]/, '@brackets'],
        [/#.*$/, 'comment'],
        [/\/\/.*$/, 'comment'],
        [/\/\*/, 'comment', '@comment'],
        [/"([^"\\]|\\.)*$/, 'string.invalid'],
        [/"/, 'string', '@string'],
        [/\b(true|false|null|include|required)\b/, 'keyword'],
        [/[=:]/, 'delimiter'],
        [/-?\d+(\.\d+)?([eE][\-+]?\d+)?/, 'number'],
        [/[A-Za-z0-9_.\-/$]+/, 'identifier'],
      ],
      comment: [
        [/[^/*]+/, 'comment'],
        [/\*\//, 'comment', '@pop'],
        [/[/*]/, 'comment'],
      ],
      string: [
        [/[^\\"]+/, 'string'],
        [/\\./, 'string.escape'],
        [/"/, 'string', '@pop'],
      ],
    },
  });

  hoconRegistered = true;
}

ensureMonacoEnvironment();
ensureHoconLanguage();

const settingStore = useSettingStore();
const containerRef = ref<HTMLElement>();
const editorRef = shallowRef<monaco.editor.IStandaloneCodeEditor>();
const modelRef = shallowRef<monaco.editor.ITextModel>();
const editorTheme = computed(() => (settingStore.displayMode === 'dark' ? 'vs-dark' : 'vs'));
const normalizedLanguage = computed(() => (props.language === 'hocon' ? HOCON_LANGUAGE_ID : props.language));

function createModel() {
  return monaco.editor.createModel(props.modelValue ?? '', normalizedLanguage.value);
}

function syncValueToEditor(value: string) {
  const editor = editorRef.value;
  const model = modelRef.value;
  if (!editor || !model)
    return;

  if (model.getValue() === value)
    return;

  editor.executeEdits('external-update', [
    {
      range: model.getFullModelRange(),
      text: value,
      forceMoveMarkers: true,
    },
  ]);
}

function recreateModel() {
  const editor = editorRef.value;
  if (!editor)
    return;

  const previousModel = modelRef.value;
  const nextModel = createModel();

  modelRef.value = nextModel;
  editor.setModel(nextModel);
  previousModel?.dispose();
}

function layoutEditor() {
  nextTick(() => {
    editorRef.value?.layout();
  });
}

onMounted(() => {
  monaco.editor.setTheme(editorTheme.value);
  modelRef.value = createModel();
  editorRef.value = monaco.editor.create(containerRef.value as HTMLElement, {
    model: modelRef.value,
    language: normalizedLanguage.value,
    theme: editorTheme.value,
    automaticLayout: true,
    readOnly: props.readonly,
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    wordWrap: 'on',
    lineNumbers: 'on',
    occurrencesHighlight: 'off',
    tabSize: 2,
    fontSize: 13,
    fontFamily: "Consolas, 'Courier New', monospace",
    padding: { top: 12, bottom: 12 },
  });

  editorRef.value.onDidChangeModelContent(() => {
    const value = editorRef.value?.getValue() ?? '';
    if (value !== props.modelValue)
      emit('update:modelValue', value);
  });

  layoutEditor();
});

watch(
  () => props.modelValue,
  (value) => {
    syncValueToEditor(value ?? '');
  },
);

watch(normalizedLanguage, () => {
  recreateModel();
  layoutEditor();
});

watch(
  () => props.readonly,
  (readonly) => {
    editorRef.value?.updateOptions({ readOnly: readonly });
  },
);

watch(editorTheme, (theme) => {
  monaco.editor.setTheme(theme);
});

watch(
  () => props.active,
  (active) => {
    if (active)
      layoutEditor();
  },
);

onBeforeUnmount(() => {
  editorRef.value?.setModel(null);
  editorRef.value?.dispose();
  modelRef.value?.dispose();
});
</script>

<style lang="less" scoped>
.code-editor {
  width: 100%;
  overflow: visible;
  border: 1px solid var(--td-component-border);
  border-radius: var(--td-radius-medium);
}
</style>
