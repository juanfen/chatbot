<script setup lang="ts">
import { computed } from 'vue';
import { useAppStore } from '@/store/modules/app';
import HeaderBanner from './modules/header-banner.vue';
import QuotaOverview from './modules/quota-overview.vue';
import UsageCallChart from './modules/usage-call-chart.vue';
import HomeUsageTable from './modules/home-usage-table.vue';
import HomeAnnouncements from './modules/home-announcements.vue';

const appStore = useAppStore();

const gap = computed(() => (appStore.isMobile ? 0 : 16));
</script>

<template>
  <NSpace vertical :size="16">
    <HeaderBanner />
    <NCard :bordered="false" class="card-wrapper">
      <QuotaOverview />
    </NCard>
    <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
      <NGi span="24 s:24 m:14">
        <NCard :bordered="false" class="card-wrapper">
          <div class="mb-12px text-16px font-500">{{ $t('page.homeBiz.callTrend') }}</div>
          <UsageCallChart />
        </NCard>
      </NGi>
      <NGi span="24 s:24 m:10">
        <NCard :bordered="false" class="card-wrapper">
          <HomeAnnouncements />
        </NCard>
      </NGi>
    </NGrid>
    <NCard :bordered="false" class="card-wrapper">
      <HomeUsageTable />
    </NCard>
  </NSpace>
</template>

<style scoped></style>
