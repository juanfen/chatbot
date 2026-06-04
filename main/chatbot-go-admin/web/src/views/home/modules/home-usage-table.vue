<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import type { DataTableColumns } from 'naive-ui';
import { fetchBizUsageList } from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'HomeUsageTable' });

const loading = ref(false);
const list = ref<Api.Biz.UsageLogItem[]>([]);
const pagination = reactive({ page: 1, pageSize: 8, itemCount: 0 });

const columns = computed<DataTableColumns<Api.Biz.UsageLogItem>>(() => [
  {
    title: () => $t('page.homeBiz.logTime'),
    key: 'createTime',
    width: 180,
    render: row => row.createTime?.replace('T', ' ').slice(0, 19) ?? '—'
  },
  { title: () => $t('page.homeBiz.channel'), key: 'channel', width: 120 },
  { title: () => $t('page.homeBiz.ref'), key: 'refText', ellipsis: { tooltip: true } }
]);

async function load() {
  loading.value = true;
  try {
    const data = await fetchBizUsageList({ page: pagination.page, pageSize: pagination.pageSize });
    list.value = data.list;
    pagination.itemCount = data.total;
  } finally {
    loading.value = false;
  }
}

function onPageChange(p: number) {
  pagination.page = p;
  load();
}

onMounted(load);

defineExpose({ reload: load });
</script>

<template>
  <div>
    <div class="mb-12px text-16px font-500">{{ $t('page.homeBiz.usageDetail') }}</div>
    <NDataTable
      size="small"
      :columns="columns"
      :data="list"
      :loading="loading"
      :pagination="{
        page: pagination.page,
        pageSize: pagination.pageSize,
        itemCount: pagination.itemCount,
        onUpdatePage: onPageChange
      }"
    />
  </div>
</template>

<style scoped></style>
