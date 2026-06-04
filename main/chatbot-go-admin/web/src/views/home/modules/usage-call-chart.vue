<script setup lang="ts">
import { fetchBizUsageTrend } from '@/service/api';
import { useEcharts } from '@/hooks/common/echarts';
import { $t } from '@/locales';

defineOptions({ name: 'UsageCallChart' });

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  xAxis: { type: 'category', data: [] as string[] },
  yAxis: { type: 'value', minInterval: 1 },
  series: [
    {
      name: $t('page.homeBiz.callCount'),
      type: 'line',
      smooth: true,
      data: [] as number[],
      areaStyle: {}
    }
  ]
}));

async function load() {
  const trend = await fetchBizUsageTrend(14);
  await updateOptions(opts => {
    opts.xAxis.data = trend.map(t => t.date);
    opts.series[0].data = trend.map(t => t.count);
    return opts;
  });
}

load();

defineExpose({
  reload: load
});
</script>

<template>
  <div ref="domRef" class="h-300px overflow-hidden" />
</template>

<style scoped></style>
