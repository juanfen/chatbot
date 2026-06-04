<script setup lang="ts">
import { computed, h, onMounted, onUnmounted, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { NButton, NInput, NModal, NRadio, NSelect, NSpace, NTag, NText } from 'naive-ui';
import {
  fetchSysDeviceAudioBind,
  fetchSysDeviceAudioBlob,
  fetchSysDeviceAudioList,
  fetchSysDeviceAudioSpeakerPersonal,
  fetchSysDeviceAudioSpeakerPersonalSave,
  fetchSysDeviceAudioSpeakers
} from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'VoiceRecognition' });

const loading = ref(false);
const list = ref<Api.Biz.DeviceAudioAdminItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);

const selectedId = ref<string | null>(null);
const speakers = ref<Api.Biz.UserSpeaker[]>([]);
const showBindModal = ref(false);
const bindMode = ref<'existing' | 'new'>('existing');
const bindSpeakerId = ref<string | null>(null);
const bindSpeakerName = ref('');
const binding = ref(false);

const showPersonalModal = ref(false);
const personalSpeakerId = ref<string | null>(null);
const personalLines = ref<string[]>([]);
const personalDraft = ref('');
const personalLoading = ref(false);
const personalSaving = ref(false);

const playingId = ref<string | null>(null);
let audioEl: HTMLAudioElement | null = null;
let objectUrl: string | null = null;

async function loadList() {
  loading.value = true;
  try {
    const data = await fetchSysDeviceAudioList({ page: page.value, pageSize: pageSize.value });
    list.value = data.list;
    total.value = data.total;
    if (selectedId.value && !list.value.some(r => r.id === selectedId.value)) {
      selectedId.value = null;
    }
  } finally {
    loading.value = false;
  }
}

async function loadSpeakers() {
  const data = await fetchSysDeviceAudioSpeakers();
  speakers.value = data.list;
}

function openBindModal() {
  if (!selectedId.value) {
    window.$message?.warning($t('page.voiceRecognition.selectRowFirst'));
    return;
  }
  const row = list.value.find(r => r.id === selectedId.value);
  if (row?.bound) {
    window.$message?.warning($t('page.voiceRecognition.alreadyBound'));
    return;
  }
  bindMode.value = speakers.value.length > 0 ? 'existing' : 'new';
  bindSpeakerId.value = speakers.value[0]?.speakerId ?? null;
  bindSpeakerName.value = '';
  showBindModal.value = true;
}

async function submitBind() {
  if (!selectedId.value) return;
  binding.value = true;
  try {
    const payload: { deviceAudioId: string; speakerId?: string; speakerName?: string } = {
      deviceAudioId: selectedId.value
    };
    if (bindMode.value === 'existing') {
      if (!bindSpeakerId.value) {
        window.$message?.warning($t('page.voiceRecognition.pickSpeaker'));
        return;
      }
      payload.speakerId = bindSpeakerId.value;
    } else {
      const name = bindSpeakerName.value.trim();
      if (!name) {
        window.$message?.warning($t('page.voiceRecognition.enterSpeakerName'));
        return;
      }
      payload.speakerName = name;
    }
    await fetchSysDeviceAudioBind(payload);
    window.$message?.success($t('page.voiceRecognition.bindSuccess'));
    showBindModal.value = false;
    await Promise.all([loadList(), loadSpeakers()]);
  } finally {
    binding.value = false;
  }
}

function cleanupAudio() {
  if (audioEl) {
    audioEl.pause();
    audioEl.src = '';
    audioEl = null;
  }
  if (objectUrl) {
    URL.revokeObjectURL(objectUrl);
    objectUrl = null;
  }
  playingId.value = null;
}

async function playAudio(row: Api.Biz.DeviceAudioAdminItem) {
  if (playingId.value === row.id) {
    cleanupAudio();
    return;
  }
  cleanupAudio();
  playingId.value = row.id;
  try {
    const blob = await fetchSysDeviceAudioBlob(row.id);
    objectUrl = URL.createObjectURL(blob);
    audioEl = new Audio(objectUrl);
    audioEl.onended = () => cleanupAudio();
    audioEl.onerror = () => {
      window.$message?.error($t('page.voiceRecognition.playFailed'));
      cleanupAudio();
    };
    await audioEl.play();
  } catch {
    window.$message?.error($t('page.voiceRecognition.playFailed'));
    cleanupAudio();
  }
}

function formatTime(iso: string, perpetual: boolean) {
  if (perpetual) return $t('page.voiceRecognition.perpetual');
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

const speakerOptions = computed(() =>
  speakers.value.map(s => {
    const count = s.personalLines?.length ?? 0;
    const countLabel =
      count > 0 ? ` (${$t('page.voiceRecognition.personalLineCount', { count })})` : '';
    return { label: `${s.speakerName}${countLabel}`, value: s.speakerId };
  })
);

const personalLineRunes = computed(() => [...personalDraft.value]);

function limitPersonalDraft(raw: string) {
  const runes = [...raw];
  personalDraft.value = runes.length <= 20 ? raw : runes.slice(0, 20).join('');
}

function openPersonalModal(speakerId?: string | null) {
  if (speakers.value.length === 0) {
    window.$message?.warning($t('page.voiceRecognition.personalNoSpeaker'));
    return;
  }
  const id = typeof speakerId === 'string' && speakerId.trim() ? speakerId : null;
  personalSpeakerId.value = id ?? speakers.value[0]?.speakerId ?? null;
  personalDraft.value = '';
  personalLines.value = [];
  showPersonalModal.value = true;
  if (personalSpeakerId.value) {
    loadPersonalLines(personalSpeakerId.value);
  }
}

async function loadPersonalLines(speakerId: string) {
  personalLoading.value = true;
  try {
    const data = await fetchSysDeviceAudioSpeakerPersonal(speakerId);
    personalLines.value = [...(data.personalLines ?? [])];
  } catch {
    personalLines.value = [];
  } finally {
    personalLoading.value = false;
  }
}

async function onPersonalSpeakerChange(id: string | null) {
  personalSpeakerId.value = id;
  personalDraft.value = '';
  if (id) {
    await loadPersonalLines(id);
  } else {
    personalLines.value = [];
  }
}

function addPersonalLine() {
  const line = personalDraft.value.trim();
  if (!line) {
    window.$message?.warning($t('page.voiceRecognition.personalLineEmpty'));
    return;
  }
  if ([...line].length > 20) {
    window.$message?.warning($t('page.voiceRecognition.personalLineTooLong'));
    return;
  }
  personalLines.value = [...personalLines.value, line];
  personalDraft.value = '';
}

function removePersonalLine(index: number) {
  personalLines.value = personalLines.value.filter((_, i) => i !== index);
}

async function savePersonalLines() {
  if (!personalSpeakerId.value) {
    window.$message?.warning($t('page.voiceRecognition.pickSpeakerForPersonal'));
    return;
  }
  if (personalDraft.value.trim()) {
    addPersonalLine();
  }
  if (personalLines.value.length === 0) {
    window.$message?.warning($t('page.voiceRecognition.personalSaveEmpty'));
    return;
  }
  personalSaving.value = true;
  try {
    const data = await fetchSysDeviceAudioSpeakerPersonalSave(
      personalSpeakerId.value,
      personalLines.value
    );
    personalLines.value = [...(data.personalLines ?? [])];
    window.$message?.success($t('page.voiceRecognition.personalSaveOk'));
    await loadSpeakers();
  } finally {
    personalSaving.value = false;
  }
}

const columns = computed<DataTableColumns<Api.Biz.DeviceAudioAdminItem>>(() => [
  {
    title: () => $t('page.voiceRecognition.select'),
    key: 'select',
    width: 64,
    render: row =>
      h(NRadio, {
        checked: selectedId.value === row.id,
        disabled: row.bound,
        onUpdateChecked: (checked: boolean) => {
          selectedId.value = checked ? row.id : null;
        }
      })
  },
  { title: () => $t('page.voiceRecognition.macId'), key: 'macId', minWidth: 140, ellipsis: { tooltip: true } },
  { title: () => $t('page.voiceRecognition.audioLen'), key: 'audioLen', width: 88 },
  {
    title: () => $t('page.voiceRecognition.context'),
    key: 'context',
    minWidth: 160,
    ellipsis: { tooltip: true }
  },
  {
    title: () => $t('page.voiceRecognition.speakerName'),
    key: 'speakerName',
    minWidth: 120,
    render: row =>
      row.speakerName
        ? row.speakerName
        : row.bound
          ? h(NTag, { size: 'small', type: 'warning' }, { default: () => $t('page.voiceRecognition.boundOther') })
          : '—'
  },
  {
    title: () => $t('page.voiceRecognition.audio'),
    key: 'play',
    width: 88,
    render: row =>
      h(
        NButton,
        {
          size: 'small',
          type: playingId.value === row.id ? 'warning' : 'primary',
          ghost: true,
          onClick: () => playAudio(row)
        },
        { default: () => (playingId.value === row.id ? $t('page.voiceRecognition.stop') : $t('page.voiceRecognition.play')) }
      )
  },
  {
    title: () => $t('page.voiceRecognition.createTime'),
    key: 'createTime',
    width: 168,
    render: row => formatTime(row.createTime, false)
  },
  {
    title: () => $t('page.voiceRecognition.expireTime'),
    key: 'expireTime',
    width: 168,
    render: row => formatTime(row.expireTime, row.perpetual)
  }
]);

const cardContentStyle = {
  display: 'flex',
  flexDirection: 'column' as const,
  height: '100%',
  minHeight: 0,
  padding: 0
};

onMounted(async () => {
  await Promise.all([loadList(), loadSpeakers()]);
});

onUnmounted(() => {
  cleanupAudio();
});
</script>

<template>
  <div class="sys-mgmt-page box-border min-h-0 flex flex-1 flex-col">
    <NCard :bordered="false" class="card-wrapper min-h-0 flex flex-1 flex-col overflow-hidden" :content-style="cardContentStyle">
      <div class="shrink-0 px-16px pb-12px pt-16px">
        <NSpace justify="space-between" align="center">
          <NText depth="3">{{ $t('page.voiceRecognition.hint') }}</NText>
          <NSpace>
            <NButton :disabled="speakers.length === 0" @click="() => openPersonalModal()">
              {{ $t('page.voiceRecognition.personalContent') }}
            </NButton>
            <NButton type="primary" :disabled="!selectedId" @click="openBindModal">
              {{ $t('page.voiceRecognition.assignSpeaker') }}
            </NButton>
          </NSpace>
        </NSpace>
      </div>
      <div v-if="speakers.length" class="shrink-0 px-16px pb-12px">
        <NText strong>{{ $t('page.voiceRecognition.personalConfiguredTitle') }}</NText>
        <div class="mt-8px flex flex-col gap-8px">
          <div
            v-for="sp in speakers"
            :key="sp.speakerId"
            class="flex flex-wrap items-start justify-between gap-8px rd-8px bg-[var(--n-color-embedded)] px-12px py-8px"
          >
            <div class="min-w-0 flex-1">
              <NText>{{ sp.speakerName }}</NText>
              <NSpace v-if="sp.personalLines?.length" wrap class="mt-4px">
                <NTag v-for="(line, idx) in sp.personalLines" :key="`${sp.speakerId}-${idx}`" size="small">
                  {{ line }}
                </NTag>
              </NSpace>
              <NText v-else depth="3" class="mt-4px block">
                {{ $t('page.voiceRecognition.personalNotConfigured') }}
              </NText>
            </div>
            <NButton size="small" quaternary type="primary" @click="openPersonalModal(sp.speakerId)">
              {{ $t('common.edit') }}
            </NButton>
          </div>
        </div>
      </div>
      <div class="table-fill min-h-0 flex flex-1 flex-col px-16px pb-16px">
        <NDataTable
          :columns="columns"
          :data="list"
          :loading="loading"
          :scroll-x="1100"
          flex-height
          class="h-full"
        />
        <div class="mt-12px flex justify-end">
          <NPagination
            v-model:page="page"
            v-model:page-size="pageSize"
            :item-count="total"
            :page-sizes="[10, 20, 50]"
            show-size-picker
            @update:page="loadList"
            @update:page-size="
              () => {
                page = 1;
                loadList();
              }
            "
          />
        </div>
      </div>
    </NCard>

    <NModal
      v-model:show="showBindModal"
      preset="card"
      :title="$t('page.voiceRecognition.assignSpeaker')"
      class="w-480px"
    >
      <NSpace vertical>
        <NRadioGroup v-model:value="bindMode">
          <NSpace>
            <NRadio value="existing" :disabled="speakers.length === 0">
              {{ $t('page.voiceRecognition.useExisting') }}
            </NRadio>
            <NRadio value="new">{{ $t('page.voiceRecognition.createNew') }}</NRadio>
          </NSpace>
        </NRadioGroup>
        <NSelect
          v-if="bindMode === 'existing'"
          v-model:value="bindSpeakerId"
          :options="speakerOptions"
          :placeholder="$t('page.voiceRecognition.pickSpeaker')"
          clearable
        />
        <NInput
          v-else
          v-model:value="bindSpeakerName"
          :placeholder="$t('page.voiceRecognition.speakerNamePlaceholder')"
        />
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showBindModal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="binding" @click="submitBind">{{ $t('common.confirm') }}</NButton>
        </NSpace>
      </template>
    </NModal>

    <NModal
      v-model:show="showPersonalModal"
      preset="card"
      :title="$t('page.voiceRecognition.personalTitle')"
      class="w-520px"
    >
      <NSpace vertical :size="12">
        <NText depth="3">{{ $t('page.voiceRecognition.personalHint') }}</NText>
        <NSelect
          :value="personalSpeakerId"
          :options="speakerOptions"
          :placeholder="$t('page.voiceRecognition.pickSpeakerForPersonal')"
          :loading="personalLoading"
          @update:value="onPersonalSpeakerChange"
        />
        <div>
          <NText depth="3">{{ $t('page.voiceRecognition.personalConfiguredLines') }}</NText>
          <NSpace v-if="personalLines.length" wrap class="mt-8px">
            <NTag
              v-for="(line, idx) in personalLines"
              :key="`${idx}-${line}`"
              closable
              @close="removePersonalLine(idx)"
            >
              {{ line }}
            </NTag>
          </NSpace>
          <NText v-else depth="3" class="mt-8px block">
            {{ $t('page.voiceRecognition.personalNotConfigured') }}
          </NText>
        </div>
        <NSpace align="center">
          <NInput
            :value="personalDraft"
            class="flex-1"
            :placeholder="$t('page.voiceRecognition.personalLinePlaceholder')"
            :disabled="personalLoading || !personalSpeakerId"
            @update:value="limitPersonalDraft"
            @keydown.enter.prevent="addPersonalLine"
          />
          <NText depth="3">{{ personalLineRunes.length }}/20</NText>
          <NButton :disabled="personalLoading || !personalSpeakerId" @click="addPersonalLine">
            {{ $t('page.voiceRecognition.addPersonalLine') }}
          </NButton>
        </NSpace>
      </NSpace>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showPersonalModal = false">{{ $t('common.cancel') }}</NButton>
          <NButton type="primary" :loading="personalSaving" :disabled="!personalSpeakerId" @click="savePersonalLines">
            {{ $t('common.confirm') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.sys-mgmt-page {
  min-height: 0;
}
.table-fill :deep(.n-data-table) {
  flex: 1;
  min-height: 0;
}
</style>
