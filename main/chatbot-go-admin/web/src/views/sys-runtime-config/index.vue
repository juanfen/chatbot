<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { NAlert, NBadge, NButton, NCard, NInput, NPopconfirm, NRadioButton, NRadioGroup, NScrollbar, NSpace, NSpin, NTag } from 'naive-ui';
import { fetchSysRuntimeConfig, fetchSysRuntimeConfigPatch } from '@/service/api';
import { $t } from '@/locales';
import JsonValueEditor from './modules/json-value-editor.vue';

defineOptions({ name: 'SysRuntimeConfig' });

/**
 * 算法运行时配置（与 RUNTIME_CONFIG_API.md 对齐）
 *
 * 视图模式：
 * - JSON 编辑器（drafts[key]：字符串）
 * - 表单视图（formData：对象，仅在表单模式下作为数据源）
 *
 * 双向同步策略：
 * - 进入表单：parse drafts[activeKey] → formData；JSON 不合法时阻止进入并提示
 * - 表单变更：deep watch formData，序列化为 pretty JSON 写回 drafts[activeKey]
 * - 离开表单（切回 JSON 或切换分区）：先把 formData 序列化回 drafts[oldKey]，
 *   再 parse drafts[newKey] 到 formData
 *
 * 这样无论用户在哪个视图改了什么，"已生效"基线 (live) 与"草稿" (drafts)
 * 始终保持一致；切回 JSON 视图能立即看到表单的所有修改，反之亦然。
 */
type SectionKey = 'gateway' | 'ota' | 'llm' | 'tts' | 'vad' | 'asr' | 'voiceprint' | 'chatbot_admin' | 'audio_pipeline';

const SECTION_KEYS: SectionKey[] = ['gateway', 'ota', 'llm', 'tts', 'vad', 'asr', 'voiceprint', 'chatbot_admin', 'audio_pipeline'];

const loading = ref(false);
const submitting = ref(false);
const activeKey = ref<SectionKey>('gateway');

const viewMode = ref<'json' | 'form'>('json');

/** 服务器最新拉取到的"已生效"配置（作为 baseline） */
const live = ref<Record<SectionKey, unknown>>(emptyConfig());
/** 用户当前正在编辑的 JSON 文本（按分区） */
const drafts = ref<Record<SectionKey, string>>({
  gateway: '',
  ota: '',
  llm: '',
  tts: '',
  vad: '',
  asr: '',
  voiceprint: '',
  chatbot_admin: '',
  audio_pipeline: ''
});
/** JSON 解析错误信息（按分区） */
const parseErrors = ref<Record<SectionKey, string>>({
  gateway: '',
  ota: '',
  llm: '',
  tts: '',
  vad: '',
  asr: '',
  voiceprint: '',
  chatbot_admin: '',
  audio_pipeline: ''
});

/** 表单视图的活动数据（仅 activeKey 当前分区） */
const formData = ref<any>({});
/** 用于在表单 deep watch 同步时避免反馈循环 */
let suppressFormSync = false;

function emptyConfig(): Record<SectionKey, unknown> {
  return {
    gateway: {},
    ota: {},
    llm: {},
    tts: {},
    vad: {},
    asr: {},
    voiceprint: {},
    chatbot_admin: {},
    audio_pipeline: {}
  };
}

function pretty(v: unknown): string {
  return JSON.stringify(v ?? {}, null, 2);
}

function deepClone<T>(v: T): T {
  return v == null ? v : (JSON.parse(JSON.stringify(v)) as T);
}

function syncDraftsFromLive() {
  for (const k of SECTION_KEYS) {
    drafts.value[k] = pretty(live.value[k]);
    parseErrors.value[k] = '';
  }
}

async function loadConfig() {
  loading.value = true;
  try {
    const data = await fetchSysRuntimeConfig();
    const next = emptyConfig();
    for (const k of SECTION_KEYS) {
      if (data && Object.prototype.hasOwnProperty.call(data, k)) {
        next[k] = (data as any)[k] ?? {};
      }
    }
    live.value = next;
    syncDraftsFromLive();
    // 拉取后让表单视图也指向最新值
    if (viewMode.value === 'form') {
      suppressFormSync = true;
      formData.value = deepClone(live.value[activeKey.value]);
      suppressFormSync = false;
    }
  } finally {
    loading.value = false;
  }
}

function parseSection(key: SectionKey): { ok: boolean; value?: unknown } {
  const text = drafts.value[key]?.trim();
  if (!text) {
    parseErrors.value[key] = '';
    return { ok: true, value: {} };
  }
  try {
    const v = JSON.parse(text);
    parseErrors.value[key] = '';
    return { ok: true, value: v };
  } catch (e: any) {
    parseErrors.value[key] = String(e?.message || e);
    return { ok: false };
  }
}

watch(
  drafts,
  () => {
    for (const k of SECTION_KEYS) parseSection(k);
  },
  { deep: true }
);

function isSectionDirty(key: SectionKey): boolean {
  const liveText = pretty(live.value[key]);
  const draftText = (drafts.value[key] ?? '').trim();
  let normalized = draftText;
  try {
    if (draftText) normalized = pretty(JSON.parse(draftText));
  } catch {
    return true;
  }
  return liveText.trim() !== normalized.trim();
}

const dirtySections = computed<SectionKey[]>(() => SECTION_KEYS.filter(isSectionDirty));
const dirtyCount = computed(() => dirtySections.value.length);

const sectionLabels = computed(() =>
  SECTION_KEYS.reduce<Record<SectionKey, string>>(
    (acc, k) => {
      acc[k] = $t(`page.sysRuntimeConfig.sections.${k}` as any);
      return acc;
    },
    {} as Record<SectionKey, string>
  )
);

function resetSection(key: SectionKey) {
  drafts.value[key] = pretty(live.value[key]);
  parseErrors.value[key] = '';
  if (viewMode.value === 'form' && key === activeKey.value) {
    suppressFormSync = true;
    formData.value = deepClone(live.value[key]);
    suppressFormSync = false;
  }
}

async function submitOne(key: SectionKey) {
  // 表单模式下先把 formData 同步到 drafts，确保提交的是最新内容
  if (viewMode.value === 'form' && key === activeKey.value) {
    drafts.value[key] = pretty(formData.value);
  }
  const r = parseSection(key);
  if (!r.ok) {
    window.$message?.error(
      $t('page.sysRuntimeConfig.jsonInvalidSection', { section: sectionLabels.value[key], msg: parseErrors.value[key] })
    );
    return;
  }
  const patch = { [key]: r.value };
  submitting.value = true;
  try {
    await fetchSysRuntimeConfigPatch(patch);
    await loadConfig();
    window.$message?.success($t('page.sysRuntimeConfig.saveOkSection', { section: sectionLabels.value[key] }));
  } finally {
    submitting.value = false;
  }
}

async function submitAll() {
  // 表单模式下先把 formData 同步到 drafts
  if (viewMode.value === 'form') {
    drafts.value[activeKey.value] = pretty(formData.value);
  }
  const keys = dirtySections.value;
  if (keys.length === 0) {
    window.$message?.info($t('page.sysRuntimeConfig.noChanges'));
    return;
  }
  const patch: Record<string, unknown> = {};
  for (const k of keys) {
    const r = parseSection(k);
    if (!r.ok) {
      window.$message?.error(
        $t('page.sysRuntimeConfig.jsonInvalidSection', { section: sectionLabels.value[k], msg: parseErrors.value[k] })
      );
      return;
    }
    patch[k] = r.value;
  }
  submitting.value = true;
  try {
    await fetchSysRuntimeConfigPatch(patch);
    await loadConfig();
    window.$message?.success($t('page.sysRuntimeConfig.saveOkAll'));
  } finally {
    submitting.value = false;
  }
}

function formatActive() {
  const r = parseSection(activeKey.value);
  if (r.ok) drafts.value[activeKey.value] = pretty(r.value);
}

/** 切换视图模式：JSON ↔ 表单。表单只在 JSON 合法时才能进入。 */
function changeView(mode: 'json' | 'form') {
  if (mode === viewMode.value) return;
  if (mode === 'form') {
    const r = parseSection(activeKey.value);
    if (!r.ok) {
      window.$message?.error(
        $t('page.sysRuntimeConfig.jsonInvalidSection', {
          section: sectionLabels.value[activeKey.value],
          msg: parseErrors.value[activeKey.value]
        })
      );
      return;
    }
    suppressFormSync = true;
    formData.value = deepClone(r.value);
    suppressFormSync = false;
  } else {
    // form -> json：把表单值同步到 drafts，让用户在 JSON 里直接看到改动
    drafts.value[activeKey.value] = pretty(formData.value);
  }
  viewMode.value = mode;
}

/** 切换分区：在表单模式下保留前一个分区的修改，再 parse 新分区。 */
function changeActive(newKey: SectionKey) {
  if (newKey === activeKey.value) return;
  if (viewMode.value === 'form') {
    drafts.value[activeKey.value] = pretty(formData.value);
    activeKey.value = newKey;
    const r = parseSection(newKey);
    if (r.ok) {
      suppressFormSync = true;
      formData.value = deepClone(r.value);
      suppressFormSync = false;
    } else {
      // 新分区 JSON 不合法，自动退回 JSON 模式让用户修复
      viewMode.value = 'json';
      window.$message?.warning(
        $t('page.sysRuntimeConfig.jsonInvalidSection', {
          section: sectionLabels.value[newKey],
          msg: parseErrors.value[newKey]
        })
      );
    }
  } else {
    activeKey.value = newKey;
  }
}

/** 表单变更 → 同步到 drafts[activeKey]（只有在表单模式且非外部触发时） */
watch(
  formData,
  () => {
    if (suppressFormSync) return;
    if (viewMode.value !== 'form') return;
    drafts.value[activeKey.value] = pretty(formData.value);
  },
  { deep: true }
);

onMounted(loadConfig);
</script>

<template>
  <div class="box-border min-h-0 flex flex-1 flex-col">
    <NCard
      :bordered="false"
      class="card-wrapper min-h-0 flex flex-1 flex-col overflow-hidden"
      :content-style="{ padding: 0, display: 'flex', flexDirection: 'column', flex: '1 1 0%', minHeight: 0 }"
    >
      <div class="shrink-0 border-b border-divider/60 px-16px pb-12px pt-16px">
        <NSpace justify="space-between" align="center" :wrap="false">
          <div class="min-w-0 flex flex-col gap-4px">
            <span class="text-16px font-semibold">{{ $t('page.sysRuntimeConfig.pageTitle') }}</span>
            <span class="truncate text-12px text-text-3">{{ $t('page.sysRuntimeConfig.pageHint') }}</span>
          </div>
          <NSpace :wrap="false">
            <NTag v-if="dirtyCount > 0" type="warning" :bordered="false" round>
              {{ $t('page.sysRuntimeConfig.diffOnlyHint', { n: dirtyCount }) }}
            </NTag>
            <NButton :loading="loading" @click="loadConfig">
              {{ $t('page.sysRuntimeConfig.reload') }}
            </NButton>
            <NPopconfirm v-if="dirtyCount > 0" @positive-click="submitAll">
              <template #trigger>
                <NButton type="primary" :loading="submitting" :disabled="loading">
                  {{ $t('page.sysRuntimeConfig.submitAll') }}
                </NButton>
              </template>
              {{ $t('common.confirm') }}?
            </NPopconfirm>
            <NButton v-else type="primary" disabled>
              {{ $t('page.sysRuntimeConfig.submitAll') }}
            </NButton>
          </NSpace>
        </NSpace>
      </div>

      <div class="min-h-0 flex flex-1 overflow-hidden">
        <!-- 左侧分区导航 -->
        <div class="w-200px shrink-0 overflow-y-auto border-r border-divider/60 py-8px">
          <div
            v-for="k in SECTION_KEYS"
            :key="k"
            class="mx-8px my-2px box-border flex cursor-pointer items-center justify-between gap-8px rounded-md px-12px py-8px transition-colors"
            :class="
              activeKey === k
                ? 'bg-primary/10 text-primary'
                : 'text-text-2 hover:bg-default-200/40 hover:text-text-1'
            "
            @click="changeActive(k)"
          >
            <span class="truncate text-13px">{{ sectionLabels[k] }}</span>
            <NBadge v-if="isSectionDirty(k)" dot type="warning" />
          </div>
        </div>

        <!-- 右侧编辑区域 -->
        <div class="right-pane min-h-0 min-w-0 flex flex-1 flex-col">
          <NSpin :show="loading" class="min-h-0 flex flex-1 flex-col">
            <div class="min-h-0 flex flex-1 flex-col">
              <!-- 工具条 -->
              <NSpace align="center" justify="space-between" class="shrink-0 border-b border-divider/60 px-16px py-10px">
                <NSpace align="center">
                  <span class="text-14px font-medium">{{ sectionLabels[activeKey] }}</span>
                  <NTag size="small" :bordered="false">{{ activeKey }}</NTag>
                  <NTag v-if="isSectionDirty(activeKey)" size="small" type="warning" :bordered="false">
                    {{ $t('page.sysRuntimeConfig.diffOnly') }}
                  </NTag>
                </NSpace>
                <NSpace align="center">
                  <NRadioGroup
                    :value="viewMode"
                    size="small"
                    @update:value="(v: any) => changeView(v as 'json' | 'form')"
                  >
                    <NRadioButton value="json">{{ $t('page.sysRuntimeConfig.viewJson') }}</NRadioButton>
                    <NRadioButton value="form">{{ $t('page.sysRuntimeConfig.viewForm') }}</NRadioButton>
                  </NRadioGroup>
                  <NButton v-if="viewMode === 'json'" size="small" tertiary @click="formatActive">
                    {{ $t('page.sysRuntimeConfig.formatJson') }}
                  </NButton>
                  <NButton size="small" :disabled="!isSectionDirty(activeKey)" @click="resetSection(activeKey)">
                    {{ $t('page.sysRuntimeConfig.resetSection') }}
                  </NButton>
                  <NButton
                    size="small"
                    type="primary"
                    :loading="submitting"
                    :disabled="!isSectionDirty(activeKey) || !!parseErrors[activeKey]"
                    @click="submitOne(activeKey)"
                  >
                    {{ $t('page.sysRuntimeConfig.submitSection') }}
                  </NButton>
                </NSpace>
              </NSpace>

              <!-- JSON 视图：textarea 占满剩余高度 -->
              <div v-if="viewMode === 'json'" class="json-area min-h-0 flex flex-1 flex-col">
                <NInput
                  v-model:value="drafts[activeKey]"
                  type="textarea"
                  :placeholder="'{}'"
                  class="json-editor"
                  :input-props="{ spellcheck: false, autocapitalize: 'off', autocomplete: 'off' }"
                />
                <NAlert
                  v-if="parseErrors[activeKey]"
                  type="error"
                  :show-icon="false"
                  class="shrink-0 rounded-none border-x-0 border-b-0"
                  :title="$t('page.sysRuntimeConfig.jsonInvalid')"
                >
                  {{ parseErrors[activeKey] }}
                </NAlert>
              </div>

              <!-- 表单视图：递归渲染 -->
              <div v-else class="min-h-0 flex flex-1 flex-col">
                <NScrollbar class="form-scroll min-h-0 flex-1">
                  <div class="px-16px py-12px">
                    <JsonValueEditor v-model="formData" :field-key="activeKey" />
                  </div>
                </NScrollbar>
              </div>
            </div>
          </NSpin>
        </div>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
/* NSpin 默认 .n-spin-container / .n-spin-content 是 display:block，
   会"截断"父级 flex 高度链路，导致 textarea 的 height:100% 解析为 0。
   这里强制让它的两层内部包装继续做列向 flex 容器。 */
.right-pane :deep(.n-spin-container),
.right-pane :deep(.n-spin-content) {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  min-height: 0;
  height: 100%;
}

.json-area {
  position: relative;
  overflow: hidden;
}

/* 让 NInput textarea 撑满整个剩余高度 */
.json-editor {
  flex: 1 1 0;
  min-height: 0;
  display: flex;
}
.json-editor :deep(.n-input),
.json-editor :deep(.n-input-wrapper),
.json-editor :deep(.n-input__textarea) {
  flex: 1 1 0;
  min-height: 0;
  height: 100%;
}
.json-editor :deep(.n-input__textarea-el),
.json-editor :deep(.n-input__textarea-mirror) {
  height: 100% !important;
  min-height: 0;
}
.json-editor :deep(.n-input) {
  border-radius: 0;
}
.json-editor :deep(.n-input__textarea-el) {
  font-family: 'JetBrains Mono', 'Fira Code', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre;
  tab-size: 2;
  resize: none;
}

.form-scroll :deep(.n-scrollbar-content) {
  min-height: 100%;
}
</style>
