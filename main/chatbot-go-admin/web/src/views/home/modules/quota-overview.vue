<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { fetchBizQuotaSummary, fetchBizSimulateCall } from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';
import { $t } from '@/locales';

defineOptions({ name: 'QuotaOverview' });

const authStore = useAuthStore();
const summary = ref<Api.Biz.QuotaSummary | null>(null);
const loading = ref(false);
const simLoading = ref(false);

async function load() {
  loading.value = true;
  try {
    summary.value = await fetchBizQuotaSummary();
  } finally {
    loading.value = false;
  }
}

async function simulateCall() {
  simLoading.value = true;
  try {
    await fetchBizSimulateCall();
    window.$message?.success($t('page.homeBiz.simulateOk'));
    await load();
  } finally {
    simLoading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <NSpin :show="loading">
    <NSpace v-if="summary" vertical :size="12">
      <div class="text-16px font-500">{{ $t('page.homeBiz.quotaTitle') }}</div>
      <div v-if="summary.isUnlimited" class="text-20px text-primary">{{ $t('page.homeBiz.unlimited') }}</div>
      <NGrid v-else cols="1 s:2 m:4" :x-gap="12" :y-gap="12">
        <NGi>
          <NStatistic :label="$t('page.homeBiz.packageName')" :value="summary.packageTitle || '—'" />
        </NGi>
        <NGi>
          <NStatistic :label="$t('page.homeBiz.totalQuota')" :value="summary.totalQuota" />
        </NGi>
        <NGi>
          <NStatistic :label="$t('page.homeBiz.usedQuota')" :value="summary.usedQuota" />
        </NGi>
        <NGi>
          <NStatistic :label="$t('page.homeBiz.remainingQuota')" :value="summary.remainingQuota" />
        </NGi>
      </NGrid>
      <NButton
        v-if="!summary.isUnlimited"
        type="primary"
        secondary
        :loading="simLoading"
        :disabled="!authStore.isLogin"
        @click="simulateCall"
      >
        {{ $t('page.homeBiz.simulateExperience') }}
      </NButton>
    </NSpace>
  </NSpin>
</template>

<style scoped></style>
