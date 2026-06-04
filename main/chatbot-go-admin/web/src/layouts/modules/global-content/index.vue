<script setup lang="ts">
import { computed } from 'vue';
import { LAYOUT_SCROLL_EL_ID } from '@sa/materials';
import { useAppStore } from '@/store/modules/app';
import { useThemeStore } from '@/store/modules/theme';
import { useRouteStore } from '@/store/modules/route';
import { useTabStore } from '@/store/modules/tab';

defineOptions({
  name: 'GlobalContent'
});

interface Props {
  /** Show padding for content */
  showPadding?: boolean;
}

withDefaults(defineProps<Props>(), {
  showPadding: true
});

const appStore = useAppStore();
const themeStore = useThemeStore();
const routeStore = useRouteStore();
const tabStore = useTabStore();

const transitionName = computed(() => (themeStore.page.animate ? themeStore.page.animateMode : ''));

function resetScroll() {
  const el = document.querySelector(`#${LAYOUT_SCROLL_EL_ID}`);

  el?.scrollTo({ left: 0, top: 0 });
}
</script>

<template>
  <!-- flex-1 + min-h-0：让子页面（如系统管理表格）能拿到主内容区剩余高度 -->
  <div class="min-h-0 flex flex-1 flex-col">
    <RouterView v-slot="{ Component, route }">
      <Transition
        :name="transitionName"
        mode="out-in"
        @before-leave="appStore.setContentXScrollable(true)"
        @after-leave="resetScroll"
        @after-enter="appStore.setContentXScrollable(false)"
      >
        <KeepAlive :include="routeStore.cacheRoutes" :exclude="routeStore.excludeCacheRoutes">
          <component
            :is="Component"
            v-if="appStore.reloadFlag"
            :key="tabStore.getTabIdByRoute(route)"
            :class="{ 'p-16px': showPadding }"
            class="flex min-h-0 flex-grow flex-col bg-layout transition-300"
          />
        </KeepAlive>
      </Transition>
    </RouterView>
  </div>
</template>

<style></style>
