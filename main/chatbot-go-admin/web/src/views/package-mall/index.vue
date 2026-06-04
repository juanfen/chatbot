<script setup lang="ts">
import { onMounted, ref } from 'vue';
import {
  fetchBizPackageCatalog,
  fetchBizPurchaseSimulate,
  fetchBizQuotaSummary
} from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'PackageMall' });

const loading = ref(false);
const buyingId = ref<string | null>(null);
const catalog = ref<Api.Biz.PackageCatalogItem[]>([]);
const summary = ref<Api.Biz.QuotaSummary | null>(null);

async function load() {
  loading.value = true;
  try {
    const [c, s] = await Promise.all([fetchBizPackageCatalog(), fetchBizQuotaSummary()]);
    catalog.value = c;
    summary.value = s;
  } finally {
    loading.value = false;
  }
}

async function buy(pkg: Api.Biz.PackageCatalogItem) {
  buyingId.value = pkg.id;
  try {
    const res = await fetchBizPurchaseSimulate({ packageId: pkg.id });
    window.$message?.success(res.message || $t('page.packageMall.purchaseOk'));
    await load();
  } finally {
    buyingId.value = null;
  }
}

onMounted(load);
</script>

<template>
  <div class="box-border min-h-0 flex flex-1 flex-col p-16px">
    <NCard v-if="summary" :bordered="false" class="card-wrapper mb-16px">
      <NSpace vertical :size="8">
        <div class="text-16px font-500">{{ $t('page.packageMall.currentPlan') }}</div>
        <div v-if="summary.isUnlimited" class="text-primary">{{ $t('page.packageMall.unlimited') }}</div>
        <NSpace v-else :size="24">
          <span>{{ summary.packageTitle || $t('common.noData') }}</span>
          <span>{{ $t('page.packageMall.remaining') }}: {{ summary.remainingQuota }} / {{ summary.totalQuota }}</span>
        </NSpace>
      </NSpace>
    </NCard>

    <NSpin :show="loading">
      <NGrid :x-gap="16" :y-gap="16" cols="1 s:2 m:3" responsive="screen">
        <NGi v-for="pkg in catalog" :key="pkg.id">
          <NCard :title="pkg.name" class="card-wrapper h-full">
            <div class="text-14px text-gray-600 dark:text-gray-400 min-h-48px">{{ pkg.description }}</div>
            <div class="mt-12px text-18px font-600 text-primary">
              {{ $t('page.packageMall.quotaCalls', { n: pkg.quotaCalls }) }}
            </div>
            <NButton
              class="mt-16px"
              type="primary"
              block
              :loading="buyingId === pkg.id"
              @click="buy(pkg)"
            >
              {{ $t('page.packageMall.mockBuy') }}
            </NButton>
          </NCard>
        </NGi>
      </NGrid>
    </NSpin>
  </div>
</template>

<style scoped></style>
