<script setup lang="ts">
import { computed } from 'vue';
import { NButton, NInput, NInputNumber, NPopconfirm, NSelect, NSwitch, NTag } from 'naive-ui';
import { $t } from '@/locales';

defineOptions({ name: 'JsonValueEditor' });

/**
 * 递归 JSON 表单编辑器：根据值的运行时类型自动渲染。
 * - object：每个字段一行；key 可改名、可加/删字段
 * - array：每项一行；可加/删元素
 * - string / number / boolean：直接输入
 * - null：占位 + 类型切换
 *
 * 任意节点的"类型"都可通过下拉切换，切换后会把值重置为对应类型的零值，
 * 这样无论原始 JSON 里出现什么形状，UI 都能继续编辑（满足"适配所有 JSON 形式"）。
 */

type JsonType = 'object' | 'array' | 'string' | 'number' | 'boolean' | 'null';

interface Props {
  modelValue: any;
  /** 父节点传入的字段名（对象成员）或 [idx]（数组项） */
  fieldKey?: string;
  /** 是否允许修改 key（仅对象成员为 true） */
  keyEditable?: boolean;
  /** 嵌套深度（保留参数，未来用于折叠等） */
  depth?: number;
  /** 是否允许整项删除（外层根节点为 false） */
  removable?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  fieldKey: '',
  keyEditable: false,
  depth: 0,
  removable: false
});

const emit = defineEmits<{
  (e: 'update:modelValue', val: any): void;
  (e: 'update:fieldKey', key: string): void;
  (e: 'remove'): void;
}>();

function detectType(v: any): JsonType {
  if (v === null || v === undefined) return 'null';
  if (Array.isArray(v)) return 'array';
  const t = typeof v;
  if (t === 'object') return 'object';
  if (t === 'string' || t === 'number' || t === 'boolean') return t;
  return 'string';
}

const currentType = computed<JsonType>(() => detectType(props.modelValue));

const TYPE_OPTIONS: Array<{ label: string; value: JsonType }> = [
  { label: 'string', value: 'string' },
  { label: 'number', value: 'number' },
  { label: 'boolean', value: 'boolean' },
  { label: 'null', value: 'null' },
  { label: 'object', value: 'object' },
  { label: 'array', value: 'array' }
];

function setType(t: JsonType) {
  let v: any;
  switch (t) {
    case 'string':
      v = typeof props.modelValue === 'string' ? props.modelValue : '';
      break;
    case 'number':
      v = typeof props.modelValue === 'number' ? props.modelValue : 0;
      break;
    case 'boolean':
      v = typeof props.modelValue === 'boolean' ? props.modelValue : false;
      break;
    case 'null':
      v = null;
      break;
    case 'object':
      v = props.modelValue && typeof props.modelValue === 'object' && !Array.isArray(props.modelValue) ? props.modelValue : {};
      break;
    case 'array':
      v = Array.isArray(props.modelValue) ? props.modelValue : [];
      break;
    default:
      v = props.modelValue;
  }
  emit('update:modelValue', v);
}

function update(v: any) {
  emit('update:modelValue', v);
}

const localKey = computed({
  get: () => props.fieldKey,
  set: (v: string) => emit('update:fieldKey', v)
});

// === object 操作 ===
const objectEntries = computed<Array<[string, any]>>(() => {
  if (currentType.value !== 'object') return [];
  return Object.entries((props.modelValue ?? {}) as Record<string, any>);
});

function setObjectChild(key: string, val: any) {
  emit('update:modelValue', { ...(props.modelValue ?? {}), [key]: val });
}

function renameObjectChild(oldKey: string, newKey: string) {
  if (!newKey || oldKey === newKey) return;
  const obj = (props.modelValue ?? {}) as Record<string, any>;
  if (newKey in obj) return; // 避免重名覆盖
  const next: Record<string, any> = {};
  for (const [k, v] of Object.entries(obj)) {
    next[k === oldKey ? newKey : k] = v;
  }
  emit('update:modelValue', next);
}

function removeObjectChild(key: string) {
  const obj = { ...((props.modelValue ?? {}) as Record<string, any>) };
  delete obj[key];
  emit('update:modelValue', obj);
}

function addObjectChild() {
  const obj = { ...((props.modelValue ?? {}) as Record<string, any>) };
  let i = 1;
  let k = 'newKey';
  while (k in obj) k = `newKey${i++}`;
  obj[k] = '';
  emit('update:modelValue', obj);
}

// === array 操作 ===
const arrayItems = computed<any[]>(() => {
  if (currentType.value !== 'array') return [];
  return (props.modelValue ?? []) as any[];
});

function setArrayItem(idx: number, val: any) {
  const arr = [...((props.modelValue ?? []) as any[])];
  arr[idx] = val;
  emit('update:modelValue', arr);
}

function removeArrayItem(idx: number) {
  const arr = [...((props.modelValue ?? []) as any[])];
  arr.splice(idx, 1);
  emit('update:modelValue', arr);
}

function addArrayItem() {
  const arr = [...((props.modelValue ?? []) as any[]), ''];
  emit('update:modelValue', arr);
}
</script>

<template>
  <div class="jve">
    <div class="jve-row">
      <NInput
        v-if="keyEditable"
        v-model:value="localKey"
        size="small"
        class="jve-key"
        placeholder="key"
      />
      <span v-else-if="fieldKey" class="jve-key-static" :title="fieldKey">{{ fieldKey }}</span>

      <NSelect
        size="small"
        :value="currentType"
        :options="TYPE_OPTIONS"
        class="jve-type"
        :consistent-menu-width="false"
        @update:value="(v: any) => setType(v as JsonType)"
      />

      <div class="jve-value">
        <NInput
          v-if="currentType === 'string'"
          :value="modelValue"
          size="small"
          type="text"
          @update:value="update"
        />
        <NInputNumber
          v-else-if="currentType === 'number'"
          :value="modelValue"
          size="small"
          class="w-full"
          @update:value="(v: number | null) => update(v ?? 0)"
        />
        <NSwitch
          v-else-if="currentType === 'boolean'"
          :value="modelValue"
          size="small"
          @update:value="update"
        />
        <NTag v-else-if="currentType === 'null'" size="small" :bordered="false">null</NTag>
        <NTag v-else-if="currentType === 'object'" size="small" type="info" :bordered="false">
          {{ Object.keys(modelValue ?? {}).length }} fields
        </NTag>
        <NTag v-else-if="currentType === 'array'" size="small" type="info" :bordered="false">
          {{ (modelValue ?? []).length }} items
        </NTag>
      </div>

      <NPopconfirm v-if="removable" @positive-click="emit('remove')">
        <template #trigger>
          <NButton size="tiny" quaternary type="error">{{ $t('common.delete') }}</NButton>
        </template>
        {{ $t('common.confirmDelete') }}
      </NPopconfirm>
    </div>

    <div v-if="currentType === 'object'" class="jve-children">
      <JsonValueEditor
        v-for="[k, v] in objectEntries"
        :key="k"
        :model-value="v"
        :field-key="k"
        key-editable
        removable
        :depth="depth + 1"
        @update:model-value="(val: any) => setObjectChild(k, val)"
        @update:field-key="(newK: string) => renameObjectChild(k, newK)"
        @remove="removeObjectChild(k)"
      />
      <NButton size="tiny" dashed class="jve-add" @click="addObjectChild">
        + {{ $t('page.sysRuntimeConfig.form.addField') }}
      </NButton>
    </div>

    <div v-else-if="currentType === 'array'" class="jve-children">
      <JsonValueEditor
        v-for="(item, idx) in arrayItems"
        :key="idx"
        :model-value="item"
        :field-key="`[${idx}]`"
        removable
        :depth="depth + 1"
        @update:model-value="(val: any) => setArrayItem(idx, val)"
        @remove="removeArrayItem(idx)"
      />
      <NButton size="tiny" dashed class="jve-add" @click="addArrayItem">
        + {{ $t('page.sysRuntimeConfig.form.addItem') }}
      </NButton>
    </div>
  </div>
</template>

<style scoped>
.jve {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}
.jve-row {
  display: flex;
  gap: 8px;
  align-items: center;
  min-width: 0;
}
.jve-key {
  width: 180px;
  flex-shrink: 0;
}
.jve-key-static {
  width: 180px;
  flex-shrink: 0;
  padding: 0 6px;
  font-family: 'JetBrains Mono', 'Fira Code', Menlo, Consolas, monospace;
  font-size: 12px;
  color: var(--n-text-color-2, #888);
  word-break: break-all;
}
.jve-type {
  width: 96px;
  flex-shrink: 0;
}
.jve-value {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
}
.jve-children {
  margin-left: 14px;
  padding-left: 10px;
  border-left: 1px dashed var(--n-divider-color, rgba(128, 128, 128, 0.3));
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 4px;
}
.jve-add {
  align-self: flex-start;
}
</style>
