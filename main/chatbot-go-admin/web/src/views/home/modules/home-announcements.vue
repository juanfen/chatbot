<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { fetchBizAnnouncements } from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'HomeAnnouncements' });

const list = ref<Api.Biz.Announcement[]>([]);
const loading = ref(false);

async function load() {
  loading.value = true;
  try {
    list.value = await fetchBizAnnouncements();
  } finally {
    loading.value = false;
  }
}

onMounted(load);

defineExpose({ reload: load });
</script>

<template>
  <div>
    <div class="mb-12px text-16px font-500">{{ $t('page.homeBiz.announcements') }}</div>
    <NSpin :show="loading">
      <NList v-if="list.length" bordered>
        <NListItem v-for="item in list" :key="item.id">
          <NThing :title="item.title" :description="item.content" />
        </NListItem>
      </NList>
      <NEmpty v-else :description="$t('common.noData')" />
    </NSpin>
  </div>
</template>

<style scoped></style>
